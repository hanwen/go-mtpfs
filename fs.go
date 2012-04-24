package main

import (
	"github.com/hanwen/go-fuse/fuse"
	"io/ioutil"
	"log"
	"os"
	"time"
)


var _ = log.Println

type DeviceFs struct {
	fuse.DefaultNodeFileSystem
	root *RootNode
	dev  *Device
}

/*
 DeviceFs is a simple filesystem interface to an MTP device. It
 should be wrapped in a Locking(Raw)FileSystem to make sure it is
 threadsafe.
*/

func NewDeviceFs(d *Device) *DeviceFs {
	root := RootNode{}
	fs := &DeviceFs{root: &root, dev: d}
	root.fs = fs
	return fs
}

func (fs *DeviceFs) Root() fuse.FsNode {
	return fs.root
}

func (fs *DeviceFs) fetchNewFolder(id uint32, storage uint32) *FolderNode {
	f := fs.newFolder(id, storage)
	f.fetch()
	return f
}

func (fs *DeviceFs) newFolder(id uint32, storage uint32) *FolderNode {
	return &FolderNode{
		FileNode: FileNode{
			storageId: storage,
			id: id,
			fs: fs,
		},
		files:   map[string]*File{},
		folders: map[string]uint32{},
	}
}


func (fs *DeviceFs) newFile(file *File) *FileNode {
	n := &FileNode{
		storageId: file.StorageId(),
		id: file.Id(),
		file: file,
		fs: fs,
	}
	
	return n
}

type RootNode struct {
	fuse.DefaultFsNode
	fs  *DeviceFs
}



func (fs *DeviceFs) OnMount(conn *fuse.FileSystemConnector) {
	for _, s := range fs.dev.ListStorage() {
		folder := fs.fetchNewFolder(NOPARENT_ID, s.Id())
		inode := fs.root.Inode().New(true, folder)
		fs.root.Inode().AddChild(s.Description(), inode)
	}
}

const NOPARENT_ID = 0xFFFFFFFF

type FileNode struct {
	fuse.DefaultFsNode
	storageId uint32
	id      uint32
	file    *File
	fs      *DeviceFs
	dirty bool
	
	backing string
}

type FolderNode struct {
	FileNode
	files   map[string]*File
	folders map[string]uint32
}

func (n *FolderNode) OnForget() {
	if n.backing != "" {
		os.Remove(n.backing)
	}
}

func (n *FolderNode) fetch() {
	l := n.fs.dev.FilesAndFolders(n.storageId, n.id)
	for _, f := range l {
		if f.Filetype() == FILETYPE_FOLDER {
			n.folders[f.Name()] = f.Id()
		} else {
			n.files[f.Name()] = f
		}
	}
}

func (n *FileNode) send() error {
	if !n.dirty {
		return nil
	}
	
	if n.backing == "" {
		log.Panicf("sending file without backing store: %q", n.file.Name())
	}
	
	fi, err := os.Stat(n.backing)
	if err != nil {
		log.Println("could not do GetAttr on close.", err)
		return err
	}
	
	log.Printf("Sending file %q to device: %d bytes.", n.file.Name(), fi.Size())
	if n.file.Id() != 0 { 
		n.fs.dev.DeleteObject(n.file.Id())
	}
	
	n.file.SetFilesize(uint64(fi.Size()))

	f, err := os.Open(n.backing)
	if err != nil {
		return err
	}
	defer f.Close()

	start := time.Now()
	err = n.fs.dev.SendFromFileDescriptor(n.file, f.Fd())
	dt := time.Now().Sub(start)
	log.Printf("Sent %d bytes in %d ms. %.1f MB/s", fi.Size(),
		dt.Nanoseconds()/1e6, 1e3 * float64(fi.Size())/float64(dt.Nanoseconds()))
	n.dirty = false
	return err
}

// PTP supports partial fetch (not exposed in libmtp), but we might as
// well get the whole thing.
func (n *FileNode) fetch() error {
	if n.backing != "" {
		return nil
	}
	
	f, err := ioutil.TempFile("", "go-mtpfs")
	if err != nil {
		return err
	}

	defer f.Close()
	log.Println("fetching to", f.Name())
	
	err = n.fs.dev.GetFileToFileDescriptor(n.id, f.Fd())
	if err == nil {
		n.backing = f.Name()
	}
	return err
}

func (n *FileNode) Open(flags uint32, context *fuse.Context) (file fuse.File, code fuse.Status) {
	n.fetch()
	f, err := os.OpenFile(n.backing, int(flags), 0644)
	if err != nil {
		return nil, fuse.ToStatus(err)
	}

	if flags & fuse.O_ANYWRITE != 0 {
		p := &PendingFile{
			LoopbackFile: fuse.LoopbackFile{File: f},
			node: n,
		}
		return p, fuse.OK
	}
	return &fuse.LoopbackFile{File: f}, fuse.OK
}

func (n *FileNode) Truncate(file fuse.File, size uint64, context *fuse.Context) (code fuse.Status) {
	// TODO - setup a flush to device?
	n.file.filesize = 0
	if file != nil {
		return file.Truncate(size)
	} else if n.backing !=  "" {
		return fuse.ToStatus(os.Truncate(n.backing, int64(size)))
	}
	return fuse.OK
}

func (n *FileNode) GetAttr(file fuse.File, context *fuse.Context) (fi *fuse.Attr, code fuse.Status) {
	if file != nil {
		return file.GetAttr()
	}
	// TODO - read n.backing
	return &fuse.Attr{
		Mode: fuse.S_IFREG | 0644,
		Size: uint64(n.file.filesize),
	}, fuse.OK
}

func (n *FileNode) Chown(file fuse.File, uid uint32, gid uint32, context *fuse.Context) (code fuse.Status) {
	// Get rid of pesky messages from cp -a.
	return fuse.OK
}

func (n *FileNode) Chmod(file fuse.File, perms uint32, context *fuse.Context) (code fuse.Status) {
	// Get rid of pesky messages from cp -a.
	return fuse.OK
}

func (n *FileNode) Utimens(file fuse.File, AtimeNs int64, MtimeNs int64, context *fuse.Context) (code fuse.Status) {
	if n.file == nil {
		// TODO - fix mtimes for directories too. 
		return
	}
	
	n.file.SetMtime(time.Unix(0, MtimeNs))
	// TODO - if we have no dirty backing store, should set object property.
	return fuse.OK
}

func (n *FolderNode) OpenDir(context *fuse.Context) (stream chan fuse.DirEntry, status fuse.Status) {
	stream = make(chan fuse.DirEntry, len(n.folders) + len(n.files))
	for n := range n.folders {
		stream <- fuse.DirEntry{Name: n, Mode: fuse.S_IFDIR | 0755}
	}
	for n := range n.files {
		stream <- fuse.DirEntry{Name: n, Mode: fuse.S_IFREG | 0644}
	}
	close(stream)
	return stream, fuse.OK
}

func (n *FolderNode) GetAttr(file fuse.File, context *fuse.Context) (fi *fuse.Attr, code fuse.Status) {
	return &fuse.Attr{Mode: fuse.S_IFDIR | 0755}, fuse.OK
}

func (n *FolderNode) Lookup(name string, context *fuse.Context) (fi *fuse.Attr, node fuse.FsNode, code fuse.Status) {
	f := n.files[name]
	if f != nil {
		node = n.fs.newFile(f)
	} else if folderId := n.folders[name]; folderId != 0 {
		node = n.fs.fetchNewFolder(folderId, n.storageId)
	}

	if node != nil {
		n.Inode().New(true, node)
		a, s := node.GetAttr(nil, context)
		return a, node, s
	}
	return nil, nil, fuse.ENOENT
}

func (n *FolderNode) Mkdir(name string, mode uint32, context *fuse.Context) (*fuse.Attr, fuse.FsNode, fuse.Status) {
	newId := n.fs.dev.CreateFolder(n.id, name, n.storageId)
	if newId == 0 {
		return nil, nil, fuse.EIO
	}

	f := n.fs.newFolder(newId, n.storageId)
	n.Inode().New(true, f)
	n.folders[name] = newId

	a := &fuse.Attr{Mode: fuse.S_IFDIR | 0755}
	return a, f, fuse.OK
}

func (n *FolderNode) Unlink(name string, c *fuse.Context) (fuse.Status) {
	f := n.files[name]
	if f == nil {
		return fuse.ENOENT
	}
	err := n.fs.dev.DeleteObject(f.Id())
	if err != nil {
		return fuse.EIO
	}
	n.Inode().RmChild(name)
	delete(n.files, name)
	return fuse.OK
}

func (n *FolderNode) Rmdir(name string, c *fuse.Context) (fuse.Status) {
	id := n.folders[name]
	if id == 0 {
		return fuse.ENOENT
	}
	err := n.fs.dev.DeleteObject(id)
	if err != nil {
		return fuse.EIO
	}
	n.Inode().RmChild(name)
	delete(n.folders, name)
	return fuse.OK
}

func (n *FolderNode) Create(name string, flags uint32, mode uint32, context *fuse.Context) (file fuse.File, fi *fuse.Attr, node fuse.FsNode, code fuse.Status) {
	f, err := ioutil.TempFile("", "go-mtpfs")
	if err != nil {
		return nil, nil, nil, fuse.ToStatus(err)

	}
	now := time.Now()
	fn := &FileNode{
		storageId: n.storageId,
		file: NewFile(0, n.id, n.storageId, name,
			0, now, FILETYPE_UNKNOWN),
		fs: n.fs,
	}
	fn.backing = f.Name()
	n.files[name] = fn.file
	n.Inode().New(false, fn)
	
	p := &PendingFile{
		LoopbackFile: fuse.LoopbackFile{File: f},
		node: fn,
	}

	a := &fuse.Attr{
		Mode: fuse.S_IFREG | 0644,
	}
	a.SetTimes(&now, &now, &now)
	
	return p, a, fn, fuse.OK
}

type PendingFile struct {
	fuse.LoopbackFile
	node *FileNode
}

func (p *PendingFile) Write(input *fuse.WriteIn, data []byte) (uint32, fuse.Status) {
	p.node.dirty = true
	return p.LoopbackFile.Write(input, data)
}

func (p *PendingFile) Truncate(size uint64) fuse.Status {
	p.node.dirty = true
	return p.LoopbackFile.Truncate(size)
}


func (p *PendingFile) Flush() fuse.Status {
	code := p.LoopbackFile.Flush()
	if !code.Ok() {
		return code
	}
	
	s := fuse.ToStatus(p.node.send())
	if s == fuse.ENOSYS {
		return fuse.EIO
	}
	return s
}

func (p *PendingFile) Release() {
	p.LoopbackFile.Release()
	p.node.send()
}
