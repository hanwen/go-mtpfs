// Copyright 2012 Google Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"github.com/hanwen/go-fuse/fuse"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"syscall"
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
- expose properties as xattrs?

*/
func (fs *DeviceFs) Root() fuse.FsNode {
	return fs.root
}

func (fs *DeviceFs) trimUnused(todo int64, node *fuse.Inode) (done int64) {
	for _, ch := range node.Children() {
		if done > todo {
			break
		}

		if fn, ok := ch.FsNode().(*fileNode); ok {
			done += fn.trim()
		} else if ch.IsDir() {
			done += fs.trimUnused(todo - done, ch)
		}
	}
	return
}

func (fs *DeviceFs) freeBacking() (int64, error) {
	t := syscall.Statfs_t{}
	err := syscall.Statfs(fs.backingDir, &t)
	if err != nil {
		return 0, err
	}
	
	return int64(t.Bfree * uint64(t.Bsize)), nil
}

func (fs *DeviceFs) ensureFreeSpace(want int64) error {
	free, err := fs.freeBacking()
	if err != nil {
		return err
	}
	if free > want {
		return nil
	}

	todo := want - free + 10*1024
	fs.trimUnused(todo, fs.root.Inode())
	
	free, err = fs.freeBacking()
	if err != nil {
		return err
	}
	if free > want {
		return nil
	}

	return fmt.Errorf("not enough space. Have %d, want %d", free, want)
}

func (fs *DeviceFs) String() string {
	return fmt.Sprintf("DeviceFs(%s)", fs.dev.ModelName())
}

func (fs *DeviceFs) statFs() *fuse.StatfsOut {
	total := uint64(0)
	free := uint64(0)
	for _, s := range fs.dev.ListStorage() {
		total += uint64(s.MaxCapacity)
		free += uint64(s.FreeSpaceInBytes)
	}
	
	bs := uint64(1024)

	return &fuse.StatfsOut{
		Bsize: uint32(bs), 
		Blocks: total/bs,
		Bavail: free/bs,
		Bfree: free/bs,
	}
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

func (n *rootNode) StatFs() *fuse.StatfsOut {
	return n.fs.statFs()
}

func (fs *DeviceFs) OnMount(conn *fuse.FileSystemConnector) {
	for _, s := range fs.dev.ListStorage() {
		if !s.IsHierarchical() {
			log.Printf("skipping non hierarchical storage %q", s.Description())
			continue
		}
		
		folder := fs.newFolder(NOPARENT_ID, s.Id())
		inode := fs.root.Inode().New(true, folder)
		fs.root.Inode().AddChild(s.Description(), inode)
	}
}


const forbidden = ":*?\"<>|"
func SanitizeDosName(name string) string{
	if strings.IndexAny(name, forbidden) == -1 {
		return name
	}
	dest := make([]byte, len(name))
	for _, c := range name {
		if strings.Contains(forbidden, c) {
			dest = append(dest, '_')
		} else {
			dest = append(dest, c)
		}
	}
	return string(dest)
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

	// If set, there was some error writing to the backing store;
	// don't flush file to device.
	error fuse.Status
}

func (n *fileNode) StatFs() *fuse.StatfsOut {
	return n.fs.statFs()
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

	if !n.error.Ok() {
		n.dirty = false
		os.Remove(n.backing)
		n.backing = ""
		n.error = fuse.OK
		n.file.filesize = 0
		log.Printf("not sending file %q due to write errors", n.file.Name())
		return syscall.EIO // TODO - send back n.error
	}
	
	fi, err := os.Stat(n.backing)
	if err != nil {
		log.Printf("could not do stat for send: %v", err)
		return err
	}
	if fi.Size() == 0 {
		log.Printf("cannot send 0 byte file %q", n.file.Name())
		return syscall.EINVAL
	}
	f, err := os.Open(n.backing)
	if err != nil {
		return err
	}
	defer f.Close()

	if n.file.Name() == "" {
		return nil
	}
	log.Printf("sending file %q to device: %d bytes.", n.file.Name(), fi.Size())
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
	log.Printf("sent %d bytes in %d ms. %.1f MB/s", fi.Size(),
		dt.Nanoseconds()/1e6, 1e3*float64(fi.Size())/float64(dt.Nanoseconds()))
	n.dirty = false

	// This is a heuristic, but if doing a large copy, we want to
	// leave space for the next file.  It would be better if
	// userspace did fallocate and FUSE would support it.
	n.fs.ensureFreeSpace(fi.Size())
	return err
}

// Drop backing data if unused. Returns freed up space.
func (n *fileNode) trim() int64 {
	if n.dirty || n.backing == "" || n.Inode().AnyFile() != nil {
		return 0
	}
	
	fi, err := os.Stat(n.backing)
	if err != nil {
		return 0
	}

	log.Printf("removing local cache for %q, %d bytes", n.file.Name(), fi.Size())
	err = os.Remove(n.backing)
	if err != nil {
		return 0
	}
	n.backing = ""
	return fi.Size()
}


// PTP supports partial fetch (not exposed in libmtp), but we might as
// well get the whole thing.
func (n *fileNode) fetch() error {
	if n.backing != "" {
		return nil
	}
	sz := int64(n.file.filesize)
	if err := n.fs.ensureFreeSpace(sz); err != nil {
		return err
	}
	
	f, err := ioutil.TempFile(n.fs.backingDir, "")
	if err != nil {
		return err
	}

	defer f.Close()

	start := time.Now()
	err = n.fs.dev.GetFileToFileDescriptor(n.id, f.Fd())
	if err == nil {
		n.backing = f.Name()
	}
	dt := time.Now().Sub(start)
	log.Printf("fetched %q, %d bytes in %d ms. %.1f MB/s", n.file.Name(), sz,
		dt.Nanoseconds()/1e6, 1e3*float64(sz)/float64(dt.Nanoseconds()))
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

func (n *fileNode) GetAttr(out *fuse.Attr, file fuse.File, context *fuse.Context) (code fuse.Status) {
	if file != nil {
		return file.GetAttr(out)
	}

	out.Mode = fuse.S_IFREG | 0644

	if n.backing != "" {
		fi, err := os.Stat(n.backing)
		if err != nil {
			return fuse.ToStatus(err)
		}
		out.Size = uint64(fi.Size())
		t := n.file.Mtime()
		if n.dirty {
			t = fi.ModTime()
		}
		out.SetTimes(&t, &t, &t)
	} else if n.file != nil {
		out.Size = uint64(n.file.filesize)

		t := n.file.Mtime()
		out.SetTimes(&t, &t, &t)
	}

	return fuse.OK
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
	children   map[string]*File
}

// Keep the root nodes for all device storages alive.
func (n *folderNode) Deletable() bool {
	return n.id != NOPARENT_ID
}

// Fetches data from device returns false on failure.
func (n *folderNode) fetch() bool {
	if n.children != nil {
		return true
	}
	l, err := n.fs.dev.FilesAndFolders(n.storageId, n.id)
	if err != nil {
		log.Printf("FilesAndFolders failed: %v", err)
		return false
	}

	n.children = map[string]*File{}
	for _, f := range l {
		n.children[f.Name()] = f
	}
	return true
}

func (n *folderNode) OpenDir(context *fuse.Context) (stream []fuse.DirEntry, status fuse.Status) {
	if !n.fetch() {
		return nil, fuse.EIO
	}

	stream = make([]fuse.DirEntry, 0, len(n.children))
	for n, f := range n.children {
		mode := fuse.S_IFREG | 0644
		if f.Filetype() == FILETYPE_FOLDER {
			mode = fuse.S_IFDIR | 0755
		}

		stream = append(stream, fuse.DirEntry{Name: n, Mode: uint32(mode)})
	}
	return stream, fuse.OK
}

func (n *folderNode) GetAttr(out *fuse.Attr, file fuse.File, context *fuse.Context) (code fuse.Status) {
	out.Mode = fuse.S_IFDIR | 0755
	return fuse.OK
}

func (n *folderNode) getChild(name string) (*File) {
	return n.children[name]
}

func (n *folderNode) basenameRename(oldName string, newName string) error {
	file := n.getChild(oldName)

	err := n.fs.dev.SetFileName(file, newName)
	if err != nil {
		return err
	}

	delete(n.children, oldName)
	n.children[newName] = file
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

	if f := n.getChild(newName); f != nil {
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

func (n *folderNode) Lookup(out *fuse.Attr, name string, context *fuse.Context) (node fuse.FsNode, code fuse.Status) {
	if !n.fetch() {
		return nil, fuse.EIO
	}
	f := n.children[name]
	if f == nil {
		return nil, fuse.ENOENT
	}

	if f.Filetype() == FILETYPE_FOLDER {
		fNode := n.fs.newFolder(f.Id(), n.storageId)
		fNode.file = f
		node = fNode
	} else {
		node = n.fs.newFile(f)
	}

	n.Inode().AddChild(name, n.Inode().New(true, node))
	s := node.GetAttr(out, nil, context)
	return node, s
}

func (n *folderNode) Mkdir(name string, mode uint32, context *fuse.Context) (fuse.FsNode, fuse.Status) {
	if !n.fetch() {
		return nil, fuse.EIO
	}
	newId, err := n.fs.dev.CreateFolder(n.id, name, n.storageId)
	if err != nil {
		log.Printf("CreateFolder failed: %v", err)
		return nil, fuse.EIO
	}

	f := n.fs.newFolder(newId, n.storageId)
	n.Inode().AddChild(name, n.Inode().New(true, f))

	if meta, err := n.fs.dev.Filemetadata(newId); err != nil {
		log.Printf("Filemetadata failed for directory %q: %v", name, err)
		return nil, fuse.EIO
	} else {
		n.children[name] = meta
	}
	return f, fuse.OK
}

func (n *folderNode) Unlink(name string, c *fuse.Context) fuse.Status {
	if !n.fetch() {
		return fuse.EIO
	}

	f := n.getChild(name)
	if f == nil {
		return fuse.ENOENT
	}

	if f.Id() != 0 {
		err := n.fs.dev.DeleteObject(f.Id())
		if err != nil {
			log.Printf("DeleteObject failed: %v", err)
			return fuse.EIO
		}
	} else {
		f.SetName("")
	}
	n.Inode().RmChild(name)

	delete(n.children, name)
	return fuse.OK
}

func (n *folderNode) Rmdir(name string, c *fuse.Context) fuse.Status {
	return n.Unlink(name, c)
}

func (n *folderNode) Create(name string, flags uint32, mode uint32, context *fuse.Context) (file fuse.File, node fuse.FsNode, code fuse.Status) {
	if !n.fetch() {
		return nil, nil, fuse.EIO
	}
	f, err := ioutil.TempFile(n.fs.backingDir, "")
	if err != nil {
		return nil, nil, fuse.ToStatus(err)

	}
	now := time.Now()
	fn := &fileNode{
		storageId: n.storageId,
		file: NewFile(0, n.id, n.storageId, name,
			0, now, FILETYPE_UNKNOWN),
		fs: n.fs,
		dirty: true,
	}
	fn.backing = f.Name()
	n.children[name] = fn.file
	n.Inode().AddChild(name, n.Inode().New(false, fn))

	p := &pendingFile{
		LoopbackFile: fuse.LoopbackFile{File: f},
		node:         fn,
	}

	return p, fn, fuse.OK
}

////////////////
// writing files.

type pendingFile struct {
	fuse.LoopbackFile
	node *fileNode
}

func (p *pendingFile) Write(data []byte, off int64) (uint32, fuse.Status) {
	p.node.dirty = true
	n, code := p.LoopbackFile.Write(data, off)
	if !code.Ok() {
		p.node.error = code
	}
	return n, code
}

func (p *pendingFile) Truncate(size uint64) fuse.Status {
	p.node.dirty = true
	code := p.LoopbackFile.Truncate(size)
	if code.Ok() && size == 0 {
		p.node.error = fuse.OK
	}
	return code
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


