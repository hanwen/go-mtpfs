package fs

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"syscall"
	"time"

	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
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
	error syscall.Errno
}

func (n *classicNode) send() error {
	if !n.dirty {
		return nil
	}

	if n.backing == "" {
		log.Panicf("sending file without backing store: %q", n.obj.Filename)
	}

	f := n.obj
	if n.error != 0 {
		n.dirty = false
		os.Remove(n.backing)
		n.backing = ""
		n.error = 0
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
		if err := n.fs.dev.DeleteObject(n.Handle()); err != nil {
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
	if err = n.fs.dev.SendObject(backing, fi.Size()); err != nil {
		log.Printf("SendObject failed %v", err)
		return syscall.EINVAL
	}
	dt := time.Now().Sub(start)
	log.Printf("sent %d bytes in %d ms. %.1f MB/s", fi.Size(),
		dt.Nanoseconds()/1e6, 1e3*float64(fi.Size())/float64(dt.Nanoseconds()))
	n.dirty = false
	n.handle = handle

	// TODO - we should create a new child with the new handle as
	// the Inode number here, and send a notification so the new
	// file has the handle we expect.

	// XXX We could leave the file for future reading, but the
	// management of free space is a hassle when doing large
	// copies.
	return err
}

// Drop backing data if unused. Returns freed up space.
func (n *classicNode) trim() int64 {
	if n.dirty || n.backing == "" { // XXX || n.Inode().AnyFile() != nil {
		return 0
	}

	fi, err := os.Stat(n.backing)
	if err != nil {
		return 0
	}

	log.Printf("removing local cache for %q, %d bytes", n.obj.Filename, fi.Size())
	if err := os.Remove(n.backing); err != nil {
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

var _ = (fs.NodeOpener)((*classicNode)(nil))

func (n *classicNode) Open(ctx context.Context, flags uint32) (file fs.FileHandle, fuseFlags uint32, code syscall.Errno) {
	return &pendingFile{
		node: n,
	}, 0, 0
}

var _ = (fs.NodeSetattrer)((*classicNode)(nil))

func (n *classicNode) Setattr(ctx context.Context, file fs.FileHandle, in *fuse.SetAttrIn, out *fuse.AttrOut) (code syscall.Errno) {
	if file != nil {
		return file.(fs.FileSetattrer).Setattr(ctx, in, out)
	}

	return n.mtpNodeImpl.Setattr(ctx, file, in, out)
}

////////////////
// writing files.

type pendingFile struct {
	loopback fs.FileHandle
	flags    uint32
	node     *classicNode
}

func (p *pendingFile) rwLoopback() (fs.FileHandle, syscall.Errno) {
	if p.loopback == nil {
		if err := p.node.fetch(); err != nil {
			return nil, fs.ToErrno(err)
		}
		f, err := os.OpenFile(p.node.backing, os.O_RDWR|os.O_CREATE, 0644)
		if err != nil {
			return nil, fs.ToErrno(err)
		}

		p.loopback = fs.NewLoopbackFile(int(f.Fd()))
	}
	return p.loopback, 0
}

var _ = (fs.FileReader)((*pendingFile)(nil))

func (p *pendingFile) Read(ctx context.Context, data []byte, off int64) (fuse.ReadResult, syscall.Errno) {
	if p.loopback == nil {
		if err := p.node.fetch(); err != nil {
			log.Printf("fetch failed: %v", err)
			return nil, syscall.EIO
		}
		f, err := os.OpenFile(p.node.backing, os.O_RDWR|os.O_CREATE, 0644)
		if err != nil {
			return nil, fs.ToErrno(err)
		}
		p.loopback = fs.NewLoopbackFile(int(f.Fd()))
	}
	return p.loopback.(fs.FileReader).Read(ctx, data, off)
}

var _ = (fs.FileWriter)((*pendingFile)(nil))

func (p *pendingFile) Write(ctx context.Context, data []byte, off int64) (uint32, syscall.Errno) {
	p.node.dirty = true
	f, code := p.rwLoopback()
	if code != 0 {
		return 0, code
	}

	n, code := f.(fs.FileWriter).Write(ctx, data, off)
	if code != 0 {
		p.node.error = code
	}
	return n, code
}

var _ = (fs.FileSetattrer)((*pendingFile)(nil))

func (p *pendingFile) Setattr(ctx context.Context, in *fuse.SetAttrIn, out *fuse.AttrOut) syscall.Errno {
	if size, ok := in.GetSize(); ok {
		f, code := p.rwLoopback()
		if code != 0 {
			return code
		}

		code = f.(fs.FileSetattrer).Setattr(ctx, in, out)
		if code != 0 {
			return code
		}
		p.node.dirty = true
		if code == 0 && size == 0 {
			p.node.error = 0
		}
		return code
	}
	return 0
}

var _ = (fs.FileFlusher)((*pendingFile)(nil))

func (p *pendingFile) Flush(ctx context.Context) syscall.Errno {
	if p.loopback == nil {
		return 0
	}
	code := p.loopback.(fs.FileFlusher).Flush(ctx)
	if code != 0 {
		return code
	}

	s := fs.ToErrno(p.node.send())
	if s == syscall.ENOSYS {
		return syscall.EIO
	}
	return s
}

var _ = (fs.FileReleaser)((*pendingFile)(nil))

func (p *pendingFile) Release(ctx context.Context) syscall.Errno {
	if p.loopback != nil {
		return p.loopback.(fs.FileReleaser).Release(ctx)
	}
	return 0
}

////////////////////////////////////////////////////////////////

func (fs *deviceFS) trimUnused(todo int64, node *fs.Inode) (done int64) {
	for _, ch := range node.Children() {
		if done > todo {
			break
		}

		if fn, ok := ch.Operations().(*classicNode); ok {
			done += fn.trim()
		} else if ch.IsDir() {
			done += fs.trimUnused(todo-done, ch)
		}
	}
	return
}

func (fs *deviceFS) freeBacking() (int64, error) {
	t := syscall.Statfs_t{}
	if err := syscall.Statfs(fs.options.Dir, &t); err != nil {
		return 0, err
	}

	return int64(t.Bfree * uint64(t.Bsize)), nil
}

func (fs *deviceFS) ensureFreeSpace(want int64) error {
	free, err := fs.freeBacking()
	if err != nil {
		return err
	}
	if free > want {
		return nil
	}

	todo := want - free + 10*1024
	fs.trimUnused(todo, &fs.root.Inode)

	free, err = fs.freeBacking()
	if err != nil {
		return err
	}
	if free > want {
		return nil
	}

	return fmt.Errorf("not enough space in %s. Have %d, want %d", fs.options.Dir, free, want)
}

func (fs *deviceFS) setupClassic() error {
	if fs.options.Dir == "" {
		var err error
		fs.options.Dir, err = ioutil.TempDir(os.TempDir(), "go-mtpfs")
		if err != nil {
			return err
		}
		fs.delBackingDir = true
	}
	if fi, err := os.Lstat(fs.options.Dir); err != nil || !fi.IsDir() {
		return fmt.Errorf("%s is not a directory", fs.options.Dir)
	}
	return nil
}

func (dfs *deviceFS) createClassicFile(obj mtp.ObjectInfo) (file fs.FileHandle, node fs.InodeEmbedder, err error) {
	backingFile, err := ioutil.TempFile(dfs.options.Dir, "")
	cl := &classicNode{
		mtpNodeImpl: mtpNodeImpl{
			obj: &obj,
			fs:  dfs,
		},
		dirty:   true,
		backing: backingFile.Name(),
	}
	file = &pendingFile{
		loopback: fs.NewLoopbackFile(int(backingFile.Fd())),
		node:     cl,
	}

	node = cl
	return
}
