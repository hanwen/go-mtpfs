// Copyright 2012 Google Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package fs

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"syscall"
	"time"

	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
	"github.com/hanwen/go-mtpfs/mtp"
)

const blockSize = 512

type DeviceFsOptions struct {
	// Assume removable volumes are VFAT and munge filenames
	// accordingly.
	RemovableVFat bool

	// Backing directory.
	Dir string

	// Use android extensions if available.
	Android bool
}

// DeviceFS implements a fuse.NodeFileSystem that mounts multiple
// storages.
type deviceFS struct {
	backingDir    string
	delBackingDir bool
	root          *rootNode
	dev           *mtp.Device
	devInfo       mtp.DeviceInfo
	storages      []uint32
	mungeVfat     map[uint32]bool

	options *DeviceFsOptions
}

// DeviceFs is a simple filesystem interface to an MTP device. It must
// be mounted as SingleThread to make sure it is threadsafe.  The file
// system assumes the device does not touch the storage.  Arguments
// are the opened MTP device and a directory for the backing store.
func NewDeviceFSRoot(d *mtp.Device, storages []uint32, options DeviceFsOptions) (*rootNode, error) {
	fs := &deviceFS{
		root:    &rootNode{},
		dev:     d,
		options: &options,
	}
	fs.root.fs = fs
	fs.storages = storages
	if err := d.GetDeviceInfo(&fs.devInfo); err != nil {
		return nil, err
	}

	if !strings.Contains(fs.devInfo.MTPExtension, "android.com") {
		fs.options.Android = false
	}

	if !options.Android {
		if err := fs.setupClassic(); err != nil {
			return nil, err
		}
	}

	fs.mungeVfat = make(map[uint32]bool)
	for _, sid := range fs.storages {
		var info mtp.StorageInfo
		if err := fs.dev.GetStorageInfo(sid, &info); err != nil {
			return nil, err
		}
		fs.mungeVfat[sid] = info.IsRemovable() && fs.options.RemovableVFat
	}

	return fs.Root(), nil
}

func (fs *deviceFS) Root() *rootNode {
	return fs.root
}

func (fs *deviceFS) String() string {
	return fmt.Sprintf("deviceFS(%s)", fs.devInfo.Model)
}

func (dfs *deviceFS) OnAdd(ctx context.Context) {
	for _, sid := range dfs.storages {
		var info mtp.StorageInfo
		if err := dfs.dev.GetStorageInfo(sid, &info); err != nil {
			log.Printf("GetStorageInfo %x: %v", sid, err)
			continue
		}

		obj := mtp.ObjectInfo{
			ParentObject: NOPARENT_ID,
			StorageID:    sid,
			Filename:     info.StorageDescription,
		}
		folder := dfs.newFolder(obj, NOPARENT_ID)
		name := info.StorageDescription
		stable := fs.StableAttr{
			Mode: syscall.S_IFDIR,
			Ino:  uint64(sid) << 33,
		}

		dfs.root.Inode.AddChild(name,
			dfs.root.Inode.NewPersistentInode(
				ctx,
				folder, stable),
			false)
	}
}

// TODO - this should be per storage and return just the free space in
// the storage.

func (fs *deviceFS) newFile(obj mtp.ObjectInfo, size int64, id uint32) (node fs.InodeEmbedder) {
	if obj.CompressedSize != 0xFFFFFFFF {
		size = int64(obj.CompressedSize)
	}

	var m *mtpNodeImpl
	if fs.options.Android {
		n := &androidNode{}
		m = &n.mtpNodeImpl
		node = n
	} else {
		n := &classicNode{}
		m = &n.mtpNodeImpl
		node = n
	}

	m.obj = &obj
	m.handle = id
	m.fs = fs
	m.Size = size

	return node
}

type rootNode struct {
	fs.Inode
	fs *deviceFS
}

var _ = (fs.NodeOnAdder)((*rootNode)(nil))

func (r *rootNode) OnAdd(ctx context.Context) {
	r.fs.OnAdd(ctx)
}

const NOPARENT_ID = 0xFFFFFFFF

// XXX
func (n *rootNode) OnUnmount() {
	if n.fs.delBackingDir {
		os.RemoveAll(n.fs.options.Dir)
		n.fs.delBackingDir = false
	}
}

func (n *rootNode) Statfs(ctx context.Context, out *fuse.StatfsOut) syscall.Errno {
	total := uint64(0)
	free := uint64(0)
	for _, ch := range n.Children() {
		var s fuse.StatfsOut
		if errno := ch.Operations().(fs.NodeStatfser).Statfs(ctx, &s); errno == 0 {
			total += s.Blocks
			free += s.Bfree
		}
	}

	*out = fuse.StatfsOut{
		Bsize:  blockSize,
		Blocks: total,
		Bavail: free,
		Bfree:  free,
	}
	return 0
}

const forbidden = ":*?\"<>|"

func SanitizeDosName(name string) string {
	if strings.IndexAny(name, forbidden) == -1 {
		return name
	}
	dest := make([]byte, len(name))
	for i := 0; i < len(name); i++ {
		if strings.Contains(forbidden, string(name[i])) {
			dest[i] = '_'
		} else {
			dest[i] = name[i]
		}
	}
	return string(dest)
}

////////////////
// mtpNode

type mtpNode interface {
	Handle() uint32
	StorageID() uint32
	SetName(string)
}

type mtpNodeImpl struct {
	fs.Inode

	// MTP handle.
	handle uint32

	obj *mtp.ObjectInfo

	fs *deviceFS

	// This is needed because obj.CompressedSize only goes to
	// 0xFFFFFFFF
	Size int64
}

var _ = (fs.NodeStatfser)((*mtpNodeImpl)(nil))

func (n *mtpNodeImpl) Statfs(ctx context.Context, out *fuse.StatfsOut) syscall.Errno {
	total := uint64(0)
	free := uint64(0)

	var info mtp.StorageInfo
	if err := n.fs.dev.GetStorageInfo(n.StorageID(), &info); err != nil {
		log.Printf("GetStorageInfo %x: %v", n.StorageID(), err)
		return 0
	}

	total += uint64(info.MaxCapability)
	free += uint64(info.FreeSpaceInBytes)

	*out = fuse.StatfsOut{
		Bsize:  blockSize,
		Blocks: total / blockSize,
		Bavail: free / blockSize,
		Bfree:  free / blockSize,
	}
	return 0
}

var _ = (fs.NodeGetxattrer)((*mtpNodeImpl)(nil))

func (n *mtpNodeImpl) Getxattr(ctx context.Context, attr string, dest []byte) (uint32, syscall.Errno) {
	return 0, syscall.ENOSYS
}

var _ = (fs.NodeSetxattrer)((*mtpNodeImpl)(nil))

func (n *mtpNodeImpl) Setxattr(ctx context.Context, attr string, dest []byte, flags uint32) syscall.Errno {
	return syscall.ENOSYS
}

var _ = (fs.NodeGetattrer)((*mtpNodeImpl)(nil))

func (n *mtpNodeImpl) Getattr(ctx context.Context, file fs.FileHandle, out *fuse.AttrOut) (code syscall.Errno) {
	if n.IsDir() {
		out.Mode = 0755
	} else {
		out.Mode = 0644
	}

	f := n.obj
	if f != nil {
		out.Size = uint64(n.Size)
		t := f.ModificationDate
		out.SetTimes(&t, &t, &t)

		out.Blocks = (out.Size + blockSize - 1) / blockSize
	}

	return 0
}

func (n *mtpNodeImpl) setTime(mTime *time.Time) {
	// Unfortunately, we can't set the modtime; it's READONLY in
	// the Android MTP implementation. We just change the time in
	// the mount, but this is not persisted.
	if mTime != nil {
		n.obj.ModificationDate = *mTime
	}
}

func (n *mtpNodeImpl) Handle() uint32 {
	return n.handle
}

func (n *mtpNodeImpl) SetName(nm string) {
	n.obj.Filename = nm
}

func (n *mtpNodeImpl) StorageID() uint32 {
	return n.obj.StorageID
}

func (n *mtpNodeImpl) Setattr(ctx context.Context, file fs.FileHandle, in *fuse.SetAttrIn, out *fuse.AttrOut) (code syscall.Errno) {
	if mt, ok := in.GetMTime(); ok {
		n.setTime(&mt)
		var atime = mt
		if a, ok := in.GetATime(); ok {
			atime = a
		}
		out.SetTimes(&atime, &mt, nil)
	}

	return 0
}

var _ = mtpNode((*folderNode)(nil))

////////////////
// files

//////////////////
// folders

type folderNode struct {
	mtpNodeImpl
	fetched bool
}

func (fs *deviceFS) newFolder(obj mtp.ObjectInfo, h uint32) *folderNode {
	obj.AssociationType = mtp.OFC_Association
	return &folderNode{
		mtpNodeImpl: mtpNodeImpl{
			handle: h,
			obj:    &obj,
			fs:     fs,
		},
	}
}

// Keep the root nodes for all device storages alive.
func (n *folderNode) Deletable() bool {
	return n.Handle() != NOPARENT_ID
}

// Fetches data from device returns false on failure.
func (n *folderNode) fetch(ctx context.Context) bool {
	if n.fetched {
		return true
	}

	handles := mtp.Uint32Array{}
	if err := n.fs.dev.GetObjectHandles(n.StorageID(), 0x0, n.Handle(), &handles); err != nil {
		log.Printf("GetObjectHandles failed: %v", err)
		return false
	}

	infos := map[uint32]*mtp.ObjectInfo{}
	sizes := map[uint32]int64{}
	for _, handle := range handles.Values {
		obj := mtp.ObjectInfo{}
		if err := n.fs.dev.GetObjectInfo(handle, &obj); err != nil {
			log.Printf("GetObjectInfo for handle %d failed: %v", handle, err)
			continue
		}
		if obj.Filename == "" {
			log.Printf("ignoring handle 0x%x with empty name in dir 0x%x",
				handle, n.Handle())
			continue
		}

		if obj.CompressedSize == 0xFFFFFFFF {
			var val mtp.Uint64Value
			if err := n.fs.dev.GetObjectPropValue(handle, mtp.OPC_ObjectSize, &val); err != nil {
				log.Printf("GetObjectPropValue handle %d failed: %v", handle, err)
				return false
			}

			sizes[handle] = int64(val.Value)
		}
		infos[handle] = &obj
	}

	for handle, info := range infos {
		var node fs.InodeEmbedder
		info.ParentObject = n.Handle()
		isdir := info.ObjectFormat == mtp.OFC_Association

		stable := fs.StableAttr{
			// Avoid ID 1.
			Ino: uint64(handle) << 1,
		}
		if isdir {
			fNode := n.fs.newFolder(*info, handle)
			node = fNode
			stable.Mode = syscall.S_IFDIR
		} else {
			sz := sizes[handle]
			node = n.fs.newFile(*info, sz, handle)
			stable.Mode = syscall.S_IFREG
		}

		n.AddChild(info.Filename,
			n.NewPersistentInode(ctx, node, stable),
			true)
	}
	n.fetched = true
	return true
}

var _ = (fs.NodeReaddirer)((*folderNode)(nil))

func (n *folderNode) Readdir(ctx context.Context) (stream fs.DirStream, status syscall.Errno) {
	if !n.fetch(ctx) {
		return nil, syscall.EIO
	}

	r := []fuse.DirEntry{}
	for k, ch := range n.Children() {
		r = append(r, fuse.DirEntry{Mode: ch.Mode(),
			Name: k,
			Ino:  ch.StableAttr().Ino})
	}

	return fs.NewListDirStream(r), 0
}

func (n *folderNode) basenameRename(oldName string, newName string) error {
	ch := n.GetChild(oldName)

	mFile := ch.Operations().(mtpNode)

	if mFile.Handle() != 0 {
		// Only rename on device if it was sent already.
		v := mtp.StringValue{Value: newName}
		if err := n.fs.dev.SetObjectPropValue(mFile.Handle(), mtp.OPC_ObjectFileName, &v); err != nil {
			return err
		}
	}
	return nil
}

var _ = (fs.NodeRenamer)((*folderNode)(nil))

func (n *folderNode) Rename(ctx context.Context, oldName string, newParent fs.InodeEmbedder, newName string, flags uint32) (code syscall.Errno) {
	fn, ok := newParent.(*folderNode)
	if !ok {
		return syscall.ENOSYS
	}
	fn.fetch(ctx)
	n.fetch(ctx)

	if f := n.GetChild(newName); f != nil {
		if fn != n {
			// TODO - delete destination?
			log.Printf("old folder already has child %q", newName)
			return syscall.ENOSYS
		}
		// does mtp overwrite the destination?
	}

	if fn != n {
		return syscall.ENOSYS
	}

	if newName != oldName {
		if err := n.basenameRename(oldName, newName); err != nil {
			log.Printf("basenameRename failed: %v", err)
			return syscall.EIO
		}
	}

	return 0
}

var _ = (fs.NodeLookuper)((*folderNode)(nil))

func (n *folderNode) Lookup(ctx context.Context, name string, out *fuse.EntryOut) (node *fs.Inode, code syscall.Errno) {
	if !n.fetch(ctx) {
		return nil, syscall.EIO
	}
	ch := n.GetChild(name)
	if ch == nil {
		return nil, syscall.ENOENT
	}

	if ga, ok := ch.Operations().(fs.NodeGetattrer); ok {
		var attr fuse.AttrOut
		code = ga.Getattr(ctx, nil, &attr)
		out.Attr = attr.Attr
	}

	return ch, code
}

var _ = (fs.NodeMkdirer)((*folderNode)(nil))

func (n *folderNode) Mkdir(ctx context.Context, name string, mode uint32, out *fuse.EntryOut) (*fs.Inode, syscall.Errno) {
	if !n.fetch(ctx) {
		return nil, syscall.EIO
	}

	obj := mtp.ObjectInfo{
		Filename:         name,
		ObjectFormat:     mtp.OFC_Association,
		ModificationDate: time.Now(),
		ParentObject:     n.Handle(),
		StorageID:        n.StorageID(),
	}
	if n.fs.mungeVfat[n.StorageID()] {
		obj.Filename = SanitizeDosName(obj.Filename)
	}
	_, _, newId, err := n.fs.dev.SendObjectInfo(n.StorageID(), n.Handle(), &obj)
	if err != nil {
		log.Printf("CreateFolder failed: %v", err)
		return nil, syscall.EIO
	}

	f := n.fs.newFolder(obj, newId)
	stable := fs.StableAttr{Mode: syscall.S_IFDIR}
	ch := n.NewPersistentInode(ctx, f, stable)

	out.Mode = 0755
	return ch, 0
}

var _ = (fs.NodeUnlinker)((*folderNode)(nil))

func (n *folderNode) Unlink(ctx context.Context, name string) syscall.Errno {
	if !n.fetch(ctx) {
		return syscall.EIO
	}

	ch := n.GetChild(name)
	if ch == nil {
		return syscall.ENOENT
	}

	f := ch.Operations().(mtpNode)
	if f.Handle() != 0 {
		if err := n.fs.dev.DeleteObject(f.Handle()); err != nil {
			log.Printf("DeleteObject failed: %v", err)
			return syscall.EIO
		}
	} else {
		f.SetName("")
	}
	n.RmChild(name)
	return 0
}

var _ = (fs.NodeRmdirer)((*folderNode)(nil))

func (n *folderNode) Rmdir(ctx context.Context, name string) syscall.Errno {
	return n.Unlink(ctx, name)
}

var _ = (fs.NodeCreater)((*folderNode)(nil))

func (n *folderNode) Create(ctx context.Context, name string, flags uint32, mode uint32, out *fuse.EntryOut) (ch *fs.Inode, file fs.FileHandle, fuseFlags uint32, errno syscall.Errno) {
	if !n.fetch(ctx) {
		errno = syscall.EIO
		return
	}

	obj := mtp.ObjectInfo{
		StorageID:        n.StorageID(),
		Filename:         name,
		ObjectFormat:     mtp.OFC_Undefined,
		ModificationDate: time.Now(),
		ParentObject:     n.Handle(),
		CompressedSize:   0,
	}

	var fsNode fs.InodeEmbedder

	if n.fs.options.Android {
		_, _, handle, err := n.fs.dev.SendObjectInfo(n.StorageID(), n.Handle(), &obj)
		if err != nil {
			log.Println("SendObjectInfo failed", err)
			errno = syscall.EIO
			return
		}

		if err = n.fs.dev.SendObject(&bytes.Buffer{}, 0); err != nil {
			log.Println("SendObject failed:", err)
			errno = syscall.EIO
			return
		}

		aNode := &androidNode{
			mtpNodeImpl: mtpNodeImpl{
				obj:    &obj,
				fs:     n.fs,
				handle: handle,
			},
		}

		if !aNode.startEdit() {
			errno = syscall.EIO
			return
		}
		file = &androidFile{
			node: aNode,
		}
		fsNode = aNode
	} else {
		var err error
		file, fsNode, err = n.fs.createClassicFile(obj)
		if err != nil {
			errno = fs.ToErrno(err)
			return
		}
	}
	ch = n.NewPersistentInode(ctx, fsNode, fs.StableAttr{})

	var a fuse.AttrOut
	out.Attr = a.Attr
	return ch, file, 0, 0
}
