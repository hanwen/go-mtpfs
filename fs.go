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

- Moving between directories
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

func (n *fileNode) Deletable() bool {
	return false
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
		err := n.fs.dev.DeleteObject(n.file.Id())
		if err != nil {
			return err
		}
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
		t := n.file.Mtime()
		if n.dirty {
			t = fi.ModTime()
		} 
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
	
	// Unfortunately, we can't set the modtime; it's READONLY in
	// the Android MTP implementation. We just change the time in
	// the mount, but this is not persisted.
	n.file.SetMtime(time.Unix(0, MtimeNs))
	return fuse.OK
}

//////////////////
// folders

type folderNode struct {
	fileNode
	files   map[string]*File
	folders map[string]*File
}

// Fetches data from device returns false on failure.
func (n *folderNode) fetch() bool {
	if n.files != nil {
		return true
	}
	l, err := n.fs.dev.FilesAndFolders(n.storageId, n.id)
	if err != nil {
		log.Printf("FilesAndFolders failed: %v", err)
		return false
	}

	n.files = map[string]*File{}
	n.folders = map[string]*File{}
	for _, f := range l {
		if f.Filetype() == FILETYPE_FOLDER {
			n.folders[f.Name()] = f
		} else {
			n.files[f.Name()] = f
		}
	}
	return true
}

func (n *folderNode) OpenDir(context *fuse.Context) (stream chan fuse.DirEntry, status fuse.Status) {
	if !n.fetch() {
		return nil, fuse.EIO
	}

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

func (n *folderNode) getChild(name string) (f *File, isFolder bool) {
	f, isFolder = n.folders[name]
	if isFolder {
		return
	}
	f, _ = n.files[name]
	return f, false
}

func (n *folderNode) basenameRename(oldName string, newName string) error {
	file, isFolder := n.getChild(oldName)
	
	err := n.fs.dev.SetFileName(file, newName)
	if err != nil {
		return err
	}

	if isFolder {
		delete(n.folders, oldName)
		n.folders[newName] = file
	} else {
		delete(n.files,oldName)
		n.files[newName] = file
	}
	ch := n.Inode().RmChild(oldName)
	if ch == nil {
		log.Panicf("child is not there for %q: got %v", oldName, n.Inode().Children())
	}
	n.Inode().AddChild(newName, ch)
	return nil
}

func (n *folderNode) Rename(oldName string, newParent fuse.FsNode, newName string, context *fuse.Context) (code fuse.Status) {
	fn, ok := newParent.(*folderNode);
	if !ok {
		return fuse.ENOSYS
	}
	fn.fetch()
	n.fetch()
	
	if f, _ := n.getChild(newName); f != nil {
		if fn != n { 
			log.Println("old folder already has child %q", newName)
			return fuse.ENOSYS
		} else {
			// does mtp overwrite the destination?
		}
	}

	if fn != n {
		return fuse.ENOSYS
	}
	
	if newName != oldName {
		err := n.basenameRename(oldName, newName)
		if err != nil {
			log.Printf("basenameRename failed: %v", err)
			return fuse.EIO
		}
	}

	return fuse.OK
}

func (n *folderNode) Lookup(name string, context *fuse.Context) (fi *fuse.Attr, node fuse.FsNode, code fuse.Status) {
	if !n.fetch() {
		return nil, nil, fuse.EIO
	}
	f := n.files[name]
	if f != nil {
		node = n.fs.newFile(f)
	} else if folder := n.folders[name]; folder != nil {
		fNode := n.fs.newFolder(folder.Id(), n.storageId)
		fNode.file = folder
		node = fNode
	}

	if node != nil {
		n.Inode().AddChild(name, n.Inode().New(true, node))
		
		a, s := node.GetAttr(nil, context)
		return a, node, s
	}
	return nil, nil, fuse.ENOENT
}

func (n *folderNode) Mkdir(name string, mode uint32, context *fuse.Context) (*fuse.Attr, fuse.FsNode, fuse.Status) {
	if !n.fetch() {
		return nil, nil, fuse.EIO
	}
	newId, err := n.fs.dev.CreateFolder(n.id, name, n.storageId)
	if err != nil {
		log.Printf("CreateFolder failed: %v", err)
		return nil, nil, fuse.EIO
	}

	f := n.fs.newFolder(newId, n.storageId)
	n.Inode().AddChild(name, n.Inode().New(true, f))

	if meta, err := n.fs.dev.Filemetadata(newId); err != nil {
		log.Printf("Filemetadata failed for directory %q: %v", name, err)
		return nil, nil, fuse.EIO
	} else {
		n.folders[name] = meta
	}

	a, code := f.GetAttr(nil, context)
	return a, f, code
}

func (n *folderNode) Unlink(name string, c *fuse.Context) fuse.Status {
	if !n.fetch() {
		return fuse.EIO
	}
	
	f, isFolder := n.getChild(name)
	if f == nil {
		return fuse.ENOENT
	}
	err := n.fs.dev.DeleteObject(f.Id())
	if err != nil {
		log.Printf("DeleteObject failed: %v", err)	
		return fuse.EIO
	}
	n.Inode().RmChild(name)

	if isFolder {
		delete(n.folders, name)
	} else {
		delete(n.files, name)
	}
	return fuse.OK
}

func (n *folderNode) Rmdir(name string, c *fuse.Context) fuse.Status {
	return n.Unlink(name, c)
}

func (n *folderNode) Create(name string, flags uint32, mode uint32, context *fuse.Context) (file fuse.File, fi *fuse.Attr, node fuse.FsNode, code fuse.Status) {
	if !n.fetch() {
		return nil, nil, nil, fuse.EIO
	}
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
	n.Inode().AddChild(name, n.Inode().New(false, fn))

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
