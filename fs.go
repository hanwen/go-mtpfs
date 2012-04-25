package main

import (
	"github.com/hanwen/go-fuse/fuse"
	"io/ioutil"
	"log"
	"os"
	"time"
)

type DeviceFs struct {
	fuse.DefaultNodeFileSystem
	backingDir string
	root       *rootNode
	dev        *Device
}

/* DeviceFs is a simple filesystem interface to an MTP device. It
should be wrapped in a Locking(Raw)FileSystem to make sure it is
threadsafe.  The file system assumes the device does not touch the
storage.  Arguments are the opened mtp device and a directory for the
backing store. */

func NewDeviceFs(d *Device, dir string) *DeviceFs {
	root := rootNode{}
	fs := &DeviceFs{root: &root, dev: d, backingDir: dir}
	root.fs = fs
	return fs
}

/*
TODO:

- We leak memory given by LIBMTP.
- Renaming
- Something intelligent with playlists/pictures, maybe?
- Statfs?
- expose properties as xattrs?

*/
func (fs *DeviceFs) Root() fuse.FsNode {
	return fs.root
}

func (fs *DeviceFs) newFolder(id uint32, storage uint32) *folderNode {
	return &folderNode{
		fileNode: fileNode{
			storageId: storage,
			id:        id,
			fs:        fs,
		},
	}
}

func (fs *DeviceFs) newFile(file *File) *fileNode {
	n := &fileNode{
		storageId: file.StorageId(),
		id:        file.Id(),
		file:      file,
		fs:        fs,
	}

	return n
}

type rootNode struct {
	fuse.DefaultFsNode
	fs *DeviceFs
}

func (fs *DeviceFs) OnMount(conn *fuse.FileSystemConnector) {
	for _, s := range fs.dev.ListStorage() {
		folder := fs.newFolder(NOPARENT_ID, s.Id())
		inode := fs.root.Inode().New(true, folder)
		fs.root.Inode().AddChild(s.Description(), inode)
	}
}

////////////////
// files

type fileNode struct {
	fuse.DefaultFsNode
	fs *DeviceFs

	storageId uint32
	id        uint32

	// Underlying mtp file. Maybe nil for the root of a storage.
	file *File

	// local file containing the contents.
	backing string
	// If set, the backing file was changed.
	dirty bool
}

func (n *fileNode) OnForget() {
	if n.file != nil {
		n.file.Destroy()
		n.file = nil
	}
}

func (n *fileNode) send() error {
	if !n.dirty {
		return nil
	}

	if n.backing == "" {
		log.Panicf("sending file without backing store: %q", n.file.Name())
	}

	fi, err := os.Stat(n.backing)
	if err != nil {
		log.Printf("could not do stat for send: %v", err)
		return err
	}
	f, err := os.Open(n.backing)
	if err != nil {
		return err
	}
	defer f.Close()

	log.Printf("Sending file %q to device: %d bytes.", n.file.Name(), fi.Size())
	if n.file.Id() != 0 {
		// Apparently, you can't overwrite things in MTP.
		n.fs.dev.DeleteObject(n.file.Id())
	}
	n.file.SetFilesize(uint64(fi.Size()))
	start := time.Now()
	err = n.fs.dev.SendFromFileDescriptor(n.file, f.Fd())
	dt := time.Now().Sub(start)
	log.Printf("Sent %d bytes in %d ms. %.1f MB/s", fi.Size(),
		dt.Nanoseconds()/1e6, 1e3*float64(fi.Size())/float64(dt.Nanoseconds()))
	n.dirty = false
	return err
}

// PTP supports partial fetch (not exposed in libmtp), but we might as
// well get the whole thing.
func (n *fileNode) fetch() error {
	if n.backing != "" {
		return nil
	}

	f, err := ioutil.TempFile(n.fs.backingDir, "")
	if err != nil {
		return err
	}

	defer f.Close()

	err = n.fs.dev.GetFileToFileDescriptor(n.id, f.Fd())
	if err == nil {
		n.backing = f.Name()
	}
	return err
}

func (n *fileNode) Open(flags uint32, context *fuse.Context) (file fuse.File, code fuse.Status) {
	n.fetch()
	f, err := os.OpenFile(n.backing, int(flags), 0644)
	if err != nil {
		return nil, fuse.ToStatus(err)
	}

	if flags&fuse.O_ANYWRITE != 0 {
		p := &pendingFile{
			LoopbackFile: fuse.LoopbackFile{File: f},
			node:         n,
		}
		return p, fuse.OK
	}
	return &fuse.LoopbackFile{File: f}, fuse.OK
}

func (n *fileNode) Truncate(file fuse.File, size uint64, context *fuse.Context) (code fuse.Status) {
	if file != nil {
		return file.Truncate(size)
	} else if n.backing != "" {
		return fuse.ToStatus(os.Truncate(n.backing, int64(size)))
	}
	return fuse.OK
}

func (n *fileNode) GetAttr(file fuse.File, context *fuse.Context) (fi *fuse.Attr, code fuse.Status) {
	if file != nil {
		return file.GetAttr()
	}

	a := &fuse.Attr{Mode: fuse.S_IFREG | 0644}

	if n.backing != "" {
		fi, err := os.Stat(n.backing)
		if err != nil {
			return nil, fuse.ToStatus(err)
		}
		a.Size = uint64(fi.Size())
		t := fi.ModTime()
		a.SetTimes(&t, &t, &t)
	} else if n.file != nil {
		a.Size = uint64(n.file.filesize)

		t := n.file.Mtime()
		a.SetTimes(&t, &t, &t)
	}

	return a, fuse.OK
}

func (n *fileNode) Chown(file fuse.File, uid uint32, gid uint32, context *fuse.Context) (code fuse.Status) {
	// Get rid of pesky messages from cp -a.
	return fuse.OK
}

func (n *fileNode) Chmod(file fuse.File, perms uint32, context *fuse.Context) (code fuse.Status) {
	// Get rid of pesky messages from cp -a.
	return fuse.OK
}

func (n *fileNode) Utimens(file fuse.File, AtimeNs int64, MtimeNs int64, context *fuse.Context) (code fuse.Status) {
	if n.file == nil {
		return
	}

	n.file.SetMtime(time.Unix(0, MtimeNs))
	// TODO - if we have no dirty backing store, should set object property.
	return fuse.OK
}

//////////////////
// folders

type folderNode struct {
	fileNode
	files   map[string]*File
	folders map[string]*File
}

func (n *folderNode) fetch() {
	if n.files != nil {
		return
	}
	n.files = map[string]*File{}
	n.folders = map[string]*File{}

	l := n.fs.dev.FilesAndFolders(n.storageId, n.id)
	for _, f := range l {
		if f.Filetype() == FILETYPE_FOLDER {
			n.folders[f.Name()] = f
		} else {
			n.files[f.Name()] = f
		}
	}
}

func (n *folderNode) OpenDir(context *fuse.Context) (stream chan fuse.DirEntry, status fuse.Status) {
	n.fetch()

	stream = make(chan fuse.DirEntry, len(n.folders)+len(n.files))
	for n := range n.folders {
		stream <- fuse.DirEntry{Name: n, Mode: fuse.S_IFDIR | 0755}
	}
	for n := range n.files {
		stream <- fuse.DirEntry{Name: n, Mode: fuse.S_IFREG | 0644}
	}
	close(stream)
	return stream, fuse.OK
}

func (n *folderNode) GetAttr(file fuse.File, context *fuse.Context) (fi *fuse.Attr, code fuse.Status) {
	return &fuse.Attr{Mode: fuse.S_IFDIR | 0755}, fuse.OK
}

func (n *folderNode) Lookup(name string, context *fuse.Context) (fi *fuse.Attr, node fuse.FsNode, code fuse.Status) {
	n.fetch()
	f := n.files[name]
	if f != nil {
		node = n.fs.newFile(f)
	} else if folder := n.folders[name]; folder != nil {
		fNode := n.fs.newFolder(folder.Id(), n.storageId)
		fNode.file = folder
		node = fNode
	}

	if node != nil {
		n.Inode().New(true, node)
		a, s := node.GetAttr(nil, context)
		return a, node, s
	}
	return nil, nil, fuse.ENOENT
}

func (n *folderNode) Mkdir(name string, mode uint32, context *fuse.Context) (*fuse.Attr, fuse.FsNode, fuse.Status) {
	n.fetch()
	newId := n.fs.dev.CreateFolder(n.id, name, n.storageId)
	if newId == 0 {
		return nil, nil, fuse.EIO
	}

	f := n.fs.newFolder(newId, n.storageId)
	n.Inode().New(true, f)

	if meta := n.fs.dev.Filemetadata(newId); meta == nil {
		log.Println("could fetch metadata for new directory %q", name)
		return nil, nil, fuse.EIO
	} else {
		n.folders[name] = meta
	}

	a, code := f.GetAttr(nil, context)
	return a, f, code
}

func (n *folderNode) Unlink(name string, c *fuse.Context) fuse.Status {
	n.fetch()
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

func (n *folderNode) Rmdir(name string, c *fuse.Context) fuse.Status {
	n.fetch()

	folderObj := n.folders[name]
	if folderObj == nil {
		return fuse.ENOENT
	}
	err := n.fs.dev.DeleteObject(folderObj.Id())
	if err != nil {
		return fuse.EIO
	}
	n.Inode().RmChild(name)
	delete(n.folders, name)
	return fuse.OK
}

func (n *folderNode) Create(name string, flags uint32, mode uint32, context *fuse.Context) (file fuse.File, fi *fuse.Attr, node fuse.FsNode, code fuse.Status) {
	n.fetch()
	f, err := ioutil.TempFile(n.fs.backingDir, "")
	if err != nil {
		return nil, nil, nil, fuse.ToStatus(err)

	}
	now := time.Now()
	fn := &fileNode{
		storageId: n.storageId,
		file: NewFile(0, n.id, n.storageId, name,
			0, now, FILETYPE_UNKNOWN),
		fs: n.fs,
	}
	fn.backing = f.Name()
	n.files[name] = fn.file
	n.Inode().New(false, fn)

	p := &pendingFile{
		LoopbackFile: fuse.LoopbackFile{File: f},
		node:         fn,
	}

	a := &fuse.Attr{
		Mode: fuse.S_IFREG | 0644,
	}
	a.SetTimes(&now, &now, &now)

	return p, a, fn, fuse.OK
}

////////////////
// writing files.

type pendingFile struct {
	fuse.LoopbackFile
	node *fileNode
}

func (p *pendingFile) Write(input *fuse.WriteIn, data []byte) (uint32, fuse.Status) {
	p.node.dirty = true
	return p.LoopbackFile.Write(input, data)
}

func (p *pendingFile) Truncate(size uint64) fuse.Status {
	p.node.dirty = true
	return p.LoopbackFile.Truncate(size)
}

func (p *pendingFile) Flush() fuse.Status {
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

func (p *pendingFile) Release() {
	p.LoopbackFile.Release()
	p.node.send()
}
