// Copyright 2012 Google Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package fs

import (
	"bytes"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/hanwen/go-fuse/fuse"
	"github.com/hanwen/go-mtpfs/mtp"
)

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
type DeviceFs struct {
	fuse.DefaultNodeFileSystem
	backingDir    string
	delBackingDir bool
	root          *rootNode
	dev           *mtp.Device
	devInfo       mtp.DeviceInfo
	storages      []uint32
	mungeVfat     map[uint32]bool

	options *DeviceFsOptions
}

// DeviceFs is a simple filesystem interface to an MTP device. It
// should be wrapped in a Locking(Raw)FileSystem to make sure it is
// threadsafe.  The file system assumes the device does not touch the
// storage.  Arguments are the opened mtp device and a directory for the
// backing store.
func NewDeviceFs(d *mtp.Device, storages []uint32, options DeviceFsOptions) (*DeviceFs, error) {
	root := rootNode{}
	fs := &DeviceFs{root: &root, dev: d, options: &options}
	root.fs = fs
	fs.storages = storages
	err := d.GetDeviceInfo(&fs.devInfo)
	if err != nil {
		return fs, nil
	}

	if !strings.Contains(fs.devInfo.MTPExtension, "android.com") {
		fs.options.Android = false
	}

	if !options.Android {
		err = fs.setupClassic()
		if err != nil {
			return nil, err
		}
	}

	fs.mungeVfat = make(map[uint32]bool)
	for _, sid := range fs.storages {
		var info mtp.StorageInfo
		if err != nil {
			return nil, err
		}

		fs.mungeVfat[sid] = info.IsRemovable() && fs.options.RemovableVFat
	}

	return fs, nil
}

func (fs *DeviceFs) Root() fuse.FsNode {
	return fs.root
}

func (fs *DeviceFs) String() string {
	return fmt.Sprintf("DeviceFs(%s)", fs.devInfo.Model)
}

// TODO - this should be per storage and return just the free space in
// the storage.
func (fs *DeviceFs) statFs() *fuse.StatfsOut {
	total := uint64(0)
	free := uint64(0)
	for _, sid := range fs.storages {
		var info mtp.StorageInfo
		if err := fs.dev.GetStorageInfo(sid, &info); err != nil {
			log.Printf("GetStorageInfo %x: %v", sid, err)
			continue
		}

		total += uint64(info.MaxCapability)
		free += uint64(info.FreeSpaceInBytes)
	}

	bs := uint64(1024)

	return &fuse.StatfsOut{
		Bsize:  uint32(bs),
		Blocks: total / bs,
		Bavail: free / bs,
		Bfree:  free / bs,
	}
}

func (fs *DeviceFs) newFile(obj mtp.ObjectInfo, size int64, id uint32) (node fuse.FsNode) {
	if obj.CompressedSize != 0xFFFFFFFF {
		size = int64(obj.CompressedSize)
	}

	mNode := mtpNodeImpl{
		obj:    &obj,
		handle: id,
		fs:     fs,
		Size:   size,
	}
	if fs.options.Android {
		node = &androidNode{
			mtpNodeImpl: mNode,
		}
	} else {
		node = &classicNode{
			mtpNodeImpl: mNode,
		}
	}
	return node
}

type rootNode struct {
	fuse.DefaultFsNode
	fs *DeviceFs
}

func (n *rootNode) StatFs() *fuse.StatfsOut {
	return n.fs.statFs()
}

const NOPARENT_ID = 0xFFFFFFFF

func (fs *DeviceFs) OnMount(conn *fuse.FileSystemConnector) {
	for _, sid := range fs.storages {
		var info mtp.StorageInfo
		if err := fs.dev.GetStorageInfo(sid, &info); err != nil {
			log.Printf("GetStorageInfo %x: %v", sid, err)
			continue
		}

		obj := mtp.ObjectInfo{
			ParentObject: NOPARENT_ID,
			StorageID:    sid,
			Filename:     info.StorageDescription,
		}
		folder := fs.newFolder(obj, NOPARENT_ID)
		inode := fs.root.Inode().New(true, folder)
		fs.root.Inode().AddChild(info.StorageDescription, inode)
	}
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
	fuse.FsNode
	Handle() uint32
	StorageID() uint32
	SetName(string)
}

type mtpNodeImpl struct {
	fuse.DefaultFsNode

	// MTP handle.
	handle uint32

	obj *mtp.ObjectInfo

	fs *DeviceFs

	// This is needed because obj.CompressedSize only goes to
	// 0xFFFFFFFF
	Size int64
}

func (n *mtpNodeImpl) GetAttr(out *fuse.Attr, file fuse.File, context *fuse.Context) (code fuse.Status) {
	out.Mode = fuse.S_IFREG | 0644
	f := n.obj
	if f != nil {
		out.Size = uint64(n.Size)
		t := f.ModificationDate
		out.SetTimes(&t, &t, &t)
	}

	return fuse.OK
}

func (n *mtpNodeImpl) Chown(file fuse.File, uid uint32, gid uint32, context *fuse.Context) (code fuse.Status) {
	// Get rid of pesky messages from cp -a.
	return fuse.OK
}

func (n *mtpNodeImpl) Chmod(file fuse.File, perms uint32, context *fuse.Context) (code fuse.Status) {
	// Get rid of pesky messages from cp -a.
	return fuse.OK
}

func (n *mtpNodeImpl) Utimens(file fuse.File, aTime *time.Time, mTime *time.Time, context *fuse.Context) (code fuse.Status) {
	// Unfortunately, we can't set the modtime; it's READONLY in
	// the Android MTP implementation. We just change the time in
	// the mount, but this is not persisted.
	if mTime != nil {
		n.obj.ModificationDate = *mTime
	}
	return fuse.OK
}

func (n *mtpNodeImpl) Handle() uint32 {
	return n.handle
}

func (n *mtpNodeImpl) SetName(nm string) {
	n.obj.Filename = nm
}

func (n *mtpNodeImpl) StatFs() *fuse.StatfsOut {
	return n.fs.statFs()
}

func (n *mtpNodeImpl) StorageID() uint32 {
	return n.obj.StorageID
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

func (fs *DeviceFs) newFolder(obj mtp.ObjectInfo, h uint32) *folderNode {
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
func (n *folderNode) fetch() bool {
	if n.fetched {
		return true
	}

	handles := mtp.Uint32Array{}
	err := n.fs.dev.GetObjectHandles(n.StorageID(), 0x0,
		n.Handle(), &handles)
	if err != nil {
		log.Printf("GetObjectHandles failed: %v", err)
		return false
	}

	infos := []*mtp.ObjectInfo{}
	sizes := map[uint32]int64{}
	for _, handle := range handles.Values {
		obj := mtp.ObjectInfo{}
		err := n.fs.dev.GetObjectInfo(handle, &obj)
		if err != nil {
			log.Printf("GetObjectInfo failed: %v", err)
			return false
		}
		if obj.Filename == "" {
			log.Printf("ignoring handle 0x%x with empty name in dir 0x%x",
				handle, n.Handle())
			infos = append(infos, nil)
			continue
		}

		if obj.CompressedSize == 0xFFFFFFFF {
			var val mtp.Uint64Value
			err := n.fs.dev.GetObjectPropValue(handle, mtp.OPC_ObjectSize, &val)
			if err != nil {
				log.Printf("GetObjectPropValue handle %d failed: %v", handle, err)
				return false
			}

			sizes[handle] = int64(val.Value)
		}
		infos = append(infos, &obj)
	}

	for i, handle := range handles.Values {
		var node fuse.FsNode
		if infos[i] == nil {
			continue
		}
		obj := *infos[i]
		obj.ParentObject = n.Handle()
		isdir := obj.ObjectFormat == mtp.OFC_Association
		if isdir {
			fNode := n.fs.newFolder(obj, handle)
			node = fNode
		} else {
			sz := sizes[handle]
			node = n.fs.newFile(obj, sz, handle)
		}

		n.Inode().AddChild(obj.Filename, n.Inode().New(isdir, node))
	}
	n.fetched = true
	return true
}

func (n *folderNode) OpenDir(context *fuse.Context) (stream []fuse.DirEntry, status fuse.Status) {
	if !n.fetch() {
		return nil, fuse.EIO
	}
	return n.DefaultFsNode.OpenDir(context)
}

func (n *folderNode) GetAttr(out *fuse.Attr, file fuse.File, context *fuse.Context) (code fuse.Status) {
	out.Mode = fuse.S_IFDIR | 0755
	return fuse.OK
}

func (n *folderNode) basenameRename(oldName string, newName string) error {
	ch := n.Inode().GetChild(oldName)

	mFile := ch.FsNode().(mtpNode)

	if mFile.Handle() != 0 {
		// Only rename on device if it was sent already.
		v := mtp.StringValue{newName}
		err := n.fs.dev.SetObjectPropValue(mFile.Handle(), mtp.OPC_ObjectFileName, &v)
		if err != nil {
			return err
		}
	}
	n.Inode().RmChild(oldName)
	n.Inode().AddChild(newName, ch)
	return nil
}

func (n *folderNode) Rename(oldName string, newParent fuse.FsNode, newName string, context *fuse.Context) (code fuse.Status) {
	fn, ok := newParent.(*folderNode)
	if !ok {
		return fuse.ENOSYS
	}
	fn.fetch()
	n.fetch()

	if f := n.Inode().GetChild(newName); f != nil {
		if fn != n {
			// TODO - delete destination?
			log.Printf("old folder already has child %q", newName)
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
	ch := n.Inode().GetChild(name)
	if ch == nil {
		return nil, fuse.ENOENT
	}

	s := ch.FsNode().GetAttr(out, nil, context)
	return ch.FsNode(), s
}

func (n *folderNode) Mkdir(name string, mode uint32, context *fuse.Context) (fuse.FsNode, fuse.Status) {
	if !n.fetch() {
		return nil, fuse.EIO
	}

	obj := mtp.ObjectInfo{
		Filename:         name,
		ObjectFormat:     mtp.OFC_Association,
		ModificationDate: time.Now(),
		ParentObject:     n.Handle(),
		StorageID:        n.StorageID(),
	}
	_, _, newId, err := n.fs.dev.SendObjectInfo(n.StorageID(), n.Handle(), &obj)
	if err != nil {
		log.Printf("CreateFolder failed: %v", err)
		return nil, fuse.EIO
	}

	f := n.fs.newFolder(obj, newId)
	n.Inode().AddChild(name, n.Inode().New(true, f))
	return f, fuse.OK
}

func (n *folderNode) Unlink(name string, c *fuse.Context) fuse.Status {
	if !n.fetch() {
		return fuse.EIO
	}

	ch := n.Inode().GetChild(name)
	if ch == nil {
		return fuse.ENOENT
	}

	f := ch.FsNode().(mtpNode)
	if f.Handle() != 0 {
		err := n.fs.dev.DeleteObject(f.Handle())
		if err != nil {
			log.Printf("DeleteObject failed: %v", err)
			return fuse.EIO
		}
	} else {
		f.SetName("")
	}
	n.Inode().RmChild(name)
	return fuse.OK
}

func (n *folderNode) Rmdir(name string, c *fuse.Context) fuse.Status {
	return n.Unlink(name, c)
}

func (n *folderNode) Create(name string, flags uint32, mode uint32, context *fuse.Context) (file fuse.File, node fuse.FsNode, code fuse.Status) {
	if !n.fetch() {
		return nil, nil, fuse.EIO
	}

	obj := mtp.ObjectInfo{
		StorageID:        n.StorageID(),
		Filename:         name,
		ObjectFormat:     mtp.OFC_Undefined,
		ModificationDate: time.Now(),
		ParentObject:     n.Handle(),
		CompressedSize:   0,
	}

	if n.fs.options.Android {
		_, _, handle, err := n.fs.dev.SendObjectInfo(n.StorageID(), n.Handle(), &obj)
		if err != nil {
			log.Println("SendObjectInfo failed", err)
			return nil, nil, fuse.EIO
		}

		err = n.fs.dev.SendObject(&bytes.Buffer{}, 0)
		if err != nil {
			log.Println("SendObject failed:", err)
			return nil, nil, fuse.EIO
		}

		if err := n.fs.dev.AndroidBeginEditObject(handle); err != nil {
			log.Println("AndroidBeginEditObject failed:", err)
			return nil, nil, fuse.EIO
		}

		node = &androidNode{
			mtpNodeImpl: mtpNodeImpl{
				obj:    &obj,
				fs:     n.fs,
				handle: handle,
			},
			write: true,
		}
		file = &androidFile{
			node: node.(*androidNode),
		}
	} else {
		var err error
		file, node, err = n.fs.createClassicFile(obj)
		if err != nil {
			return nil, nil, fuse.ToStatus(err)
		}
	}
	n.Inode().AddChild(name, n.Inode().New(false, node))
	return file, node, fuse.OK
}
