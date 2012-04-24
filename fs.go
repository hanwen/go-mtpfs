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
	backingData []string
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

func NewDeviceFs(d *Device) *DeviceFs {
	root := RootNode{}
	fs := &DeviceFs{root: &root, dev: d}
	root.fs = fs
	return fs
}

func (fs *DeviceFs) OnUnmount() {
	for _, name :=  range fs.backingData {
		os.Remove(name)
	}
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
	
	backing string
}

type FolderNode struct {
	FileNode
	files   map[string]*File
	folders map[string]uint32
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
	
	n.fs.backingData = append(n.fs.backingData, f.Name())
	
	err = n.fs.dev.GetFileToFileDescriptor(n.id, f.Fd())
	if err == nil {
		n.backing = f.Name()
	}
	return err
}

func (n *FileNode) Open(flags uint32, context *fuse.Context) (file fuse.File, code fuse.Status) {
	if flags & fuse.O_ANYWRITE != 0 {
		return nil, fuse.EROFS
	}

	n.fetch()
	f, err := os.Open(n.backing)
	if err != nil {
		return nil, fuse.ToStatus(err)
	}
	return &fuse.LoopbackFile{File: f}, fuse.OK
}

func (n *FileNode) GetAttr(file fuse.File, context *fuse.Context) (fi *fuse.Attr, code fuse.Status) {
	if file != nil {
		return file.GetAttr()
	}
	
	return &fuse.Attr{
		Mode: fuse.S_IFREG | 0644,
		Size: uint64(n.file.filesize),
	}, fuse.OK
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

func (p *PendingFile) Flush() fuse.Status {
	// Send to device. Release would be better, but we want to report errors.
	a, code := p.LoopbackFile.GetAttr()
	if !code.Ok() {
		log.Println("could not do GetAttr on close.", code)
		return code
	}
	
	p.node.file.SetFilesize(a.Size)
	err := p.node.fs.dev.SendFromFileDescriptor(p.node.file, p.LoopbackFile.File.Fd())
	flushCode := p.LoopbackFile.Flush()
	if err == nil {
		return flushCode
	}
	return fuse.ToStatus(err)
}
