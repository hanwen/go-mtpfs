package fs

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"syscall"
	"time"

	"github.com/hanwen/go-fuse/fuse"
	"github.com/hanwen/go-fuse/fuse/nodefs"
	"github.com/hanwen/go-mtpfs/mtp"
)

type classicNode struct {
	mtpNodeImpl

	// local file containing the contents.
	backing string

	// If set, the backing file was changed.
	dirty bool

	// If set, there was some error writing to the backing store;
	// don't flush file to device.
	error fuse.Status
}

func (n *classicNode) send() error {
	if !n.dirty {
		return nil
	}

	if n.backing == "" {
		log.Panicf("sending file without backing store: %q", n.obj.Filename)
	}

	f := n.obj
	if !n.error.Ok() {
		n.dirty = false
		os.Remove(n.backing)
		n.backing = ""
		n.error = fuse.OK
		n.obj.CompressedSize = 0
		n.Size = 0
		log.Printf("not sending file %q due to write errors", f.Filename)
		return syscall.EIO // TODO - send back n.error
	}

	fi, err := os.Stat(n.backing)
	if err != nil {
		log.Printf("could not do stat for send: %v", err)
		return err
	}
	if fi.Size() == 0 {
		log.Printf("cannot send 0 byte file %q", f.Filename)
		return syscall.EINVAL
	}

	if n.obj.Filename == "" {
		return nil
	}
	if n.fs.mungeVfat[n.StorageID()] {
		f.Filename = SanitizeDosName(f.Filename)
	}

	backing, err := os.Open(n.backing)
	if err != nil {
		return err
	}
	defer backing.Close()

	log.Printf("sending file %q to device: %d bytes.", f.Filename, fi.Size())
	if n.Handle() != 0 {
		// Apparently, you can't overwrite things in MTP.
		err := n.fs.dev.DeleteObject(n.Handle())
		if err != nil {
			return err
		}
		n.handle = 0
	}

	if fi.Size() > 0xFFFFFFFF {
		f.CompressedSize = 0xFFFFFFFF
	} else {
		f.CompressedSize = uint32(fi.Size())
	}
	n.Size = fi.Size()
	start := time.Now()

	_, _, handle, err := n.fs.dev.SendObjectInfo(n.StorageID(), f.ParentObject, f)
	if err != nil {
		log.Printf("SendObjectInfo failed %v", err)
		return syscall.EINVAL
	}
	err = n.fs.dev.SendObject(backing, fi.Size())
	if err != nil {
		log.Printf("SendObject failed %v", err)
		return syscall.EINVAL
	}
	dt := time.Now().Sub(start)
	log.Printf("sent %d bytes in %d ms. %.1f MB/s", fi.Size(),
		dt.Nanoseconds()/1e6, 1e3*float64(fi.Size())/float64(dt.Nanoseconds()))
	n.dirty = false
	n.handle = handle

	// We could leave the file for future reading, but the
	// management of free space is a hassle when doing large
	// copies.
	if len(n.Inode().Files(0)) == 1 {
		os.Remove(n.backing)
		n.backing = ""
	}
	return err
}

// Drop backing data if unused. Returns freed up space.
func (n *classicNode) trim() int64 {
	if n.dirty || n.backing == "" || n.Inode().AnyFile() != nil {
		return 0
	}

	fi, err := os.Stat(n.backing)
	if err != nil {
		return 0
	}

	log.Printf("removing local cache for %q, %d bytes", n.obj.Filename, fi.Size())
	err = os.Remove(n.backing)
	if err != nil {
		return 0
	}
	n.backing = ""
	return fi.Size()
}

// PTP supports partial fetch (not exposed in libmtp), but we might as
// well get the whole thing.
func (n *classicNode) fetch() error {
	if n.backing != "" {
		return nil
	}
	sz := n.Size
	if err := n.fs.ensureFreeSpace(sz); err != nil {
		return err
	}

	f, err := ioutil.TempFile(n.fs.options.Dir, "")
	if err != nil {
		return err
	}

	defer f.Close()

	start := time.Now()
	err = n.fs.dev.GetObject(n.Handle(), f)
	dt := time.Now().Sub(start)
	if err == nil {
		n.backing = f.Name()
		n.dirty = false
		log.Printf("fetched %q, %d bytes in %d ms. %.1f MB/s", n.obj.Filename, sz,
			dt.Nanoseconds()/1e6, 1e3*float64(sz)/float64(dt.Nanoseconds()))
	} else {
		log.Printf("error fetching: %v", err)
		err = syscall.EIO
	}

	return err
}

func (n *classicNode) Open(flags uint32, context *fuse.Context) (file nodefs.File, code fuse.Status) {
	return &pendingFile{
		node: n,
	}, fuse.OK
}

func (n *classicNode) Truncate(file nodefs.File, size uint64, context *fuse.Context) (code fuse.Status) {
	if file != nil {
		return file.Truncate(size)
	} else if n.backing != "" {
		return fuse.ToStatus(os.Truncate(n.backing, int64(size)))
	}
	return fuse.OK
}

////////////////
// writing files.

type pendingFile struct {
	nodefs.File
	flags    uint32
	loopback nodefs.File
	node     *classicNode
}

func (p *pendingFile) rwLoopback() (nodefs.File, fuse.Status) {
	if p.loopback == nil {
		err := p.node.fetch()
		if err != nil {
			return nil, fuse.ToStatus(err)
		}
		f, err := os.OpenFile(p.node.backing, os.O_RDWR|os.O_CREATE, 0644)
		if err != nil {
			return nil, fuse.ToStatus(err)
		}

		p.loopback = nodefs.NewLoopbackFile(f)
	}
	return p.loopback, fuse.OK
}

func (p *pendingFile) Read(data []byte, off int64) (fuse.ReadResult, fuse.Status) {
	if p.loopback == nil {
		if err := p.node.fetch(); err != nil {
			log.Printf("fetch failed: %v", err)
			return nil, fuse.EIO
		}
		f, err := os.OpenFile(p.node.backing, os.O_RDWR|os.O_CREATE, 0644)
		if err != nil {
			return nil, fuse.ToStatus(err)
		}
		p.loopback = nodefs.NewLoopbackFile(f)
	}
	return p.loopback.Read(data, off)
}

func (p *pendingFile) Write(data []byte, off int64) (uint32, fuse.Status) {
	p.node.dirty = true
	f, code := p.rwLoopback()
	if !code.Ok() {
		return 0, code
	}

	n, code := f.Write(data, off)
	if !code.Ok() {
		p.node.error = code
	}
	return n, code
}

func (p *pendingFile) Truncate(size uint64) fuse.Status {
	f, code := p.rwLoopback()
	if !code.Ok() {
		return code
	}

	code = f.Truncate(size)
	if !code.Ok() {
		return code
	}
	p.node.dirty = true
	if code.Ok() && size == 0 {
		p.node.error = fuse.OK
	}
	return code
}

func (p *pendingFile) Flush() fuse.Status {
	if p.loopback == nil {
		return fuse.OK
	}
	code := p.loopback.Flush()
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
	if p.loopback != nil {
		p.loopback.Release()
	}
}

////////////////////////////////////////////////////////////////

func (fs *DeviceFs) trimUnused(todo int64, node *nodefs.Inode) (done int64) {
	for _, ch := range node.Children() {
		if done > todo {
			break
		}

		if fn, ok := ch.Node().(*classicNode); ok {
			done += fn.trim()
		} else if ch.IsDir() {
			done += fs.trimUnused(todo-done, ch)
		}
	}
	return
}

func (fs *DeviceFs) freeBacking() (int64, error) {
	t := syscall.Statfs_t{}
	err := syscall.Statfs(fs.options.Dir, &t)
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

	return fmt.Errorf("not enough space in %s. Have %d, want %d", fs.options.Dir, free, want)
}

func (fs *DeviceFs) setupClassic() error {
	if fs.options.Dir == "" {
		var err error
		fs.options.Dir, err = ioutil.TempDir(os.TempDir(), "go-mtpfs")
		if err != nil {
			return err
		}
		fs.delBackingDir = true
	}
	if fi, err := os.Lstat(fs.options.Dir); err != nil || !fi.IsDir() {
		return fmt.Errorf("%s is not a directory")
	}
	return nil
}

func (fs *DeviceFs) OnUnmount() {
	if fs.delBackingDir {
		os.RemoveAll(fs.options.Dir)
	}
}

func (fs *DeviceFs) createClassicFile(obj mtp.ObjectInfo) (file nodefs.File, node nodefs.Node, err error) {
	backingFile, err := ioutil.TempFile(fs.options.Dir, "")
	cl := &classicNode{
		mtpNodeImpl: mtpNodeImpl{
			obj: &obj,
			fs:  fs,
		},
		dirty:   true,
		backing: backingFile.Name(),
	}
	file = &pendingFile{
		loopback: nodefs.NewLoopbackFile(backingFile),
		node:     cl,
		File:     nodefs.NewDefaultFile(),
	}

	node = cl
	return
}
