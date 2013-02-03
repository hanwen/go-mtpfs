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
}

// DeviceFS implements a fuse.NodeFileSystem that mounts multiple
// storages.
type DeviceFs struct {
	fuse.DefaultNodeFileSystem
	backingDir   string
	root         *rootNode
	dev          *mtp.Device
	devInfo      mtp.DeviceInfo
	storages     []uint32
	storageInfos []mtp.StorageInfo

	options *DeviceFsOptions
}

// DeviceFs is a simple filesystem interface to an MTP device. It
// should be wrapped in a Locking(Raw)FileSystem to make sure it is
// threadsafe.  The file system assumes the device does not touch the
// storage.  Arguments are the opened mtp device and a directory for the
// backing store.
func NewDeviceFs(d *mtp.Device, storages []uint32, options DeviceFsOptions) (*DeviceFs, error) {
	o := options

	root := rootNode{}
	fs := &DeviceFs{root: &root, dev: d, options: &o}
	root.fs = fs
	fs.storages = storages
	err := d.GetDeviceInfo(&fs.devInfo)
	if err != nil {
		return fs, nil
	}

	if !strings.Contains(fs.devInfo.MTPExtension, "android.com") {
		return nil, fmt.Errorf("this device has no android.com extensions.")
	}

	for _, sid := range fs.storages {
		var info mtp.StorageInfo
		err := d.GetStorageInfo(sid, &info)
		if err != nil {
			return nil, err
		}
		fs.storageInfos = append(fs.storageInfos, info)
	}
	return fs, nil
}

func (fs *DeviceFs) GetStorageInfo(want uint32) *mtp.StorageInfo {
	for i, sid := range fs.storages {
		if sid == want {
			return &fs.storageInfos[i]
		}
	}
	return nil
}

func (fs *DeviceFs) Root() fuse.FsNode {
	return fs.root
}

func (fs *DeviceFs) String() string {
	return fmt.Sprintf("DeviceFs(%s)", fs.devInfo.Model)
}

func (fs *DeviceFs) statFs() *fuse.StatfsOut {
	total := uint64(0)
	free := uint64(0)
	for _, s := range fs.storageInfos {
		total += uint64(s.MaxCapability)
		free += uint64(s.FreeSpaceInBytes)
	}

	bs := uint64(1024)

	return &fuse.StatfsOut{
		Bsize:  uint32(bs),
		Blocks: total / bs,
		Bavail: free / bs,
		Bfree:  free / bs,
	}
}

func (fs *DeviceFs) newFolder(obj mtp.ObjectInfo, id uint32, storage uint32) *folderNode {
	obj.AssociationType = mtp.OFC_Association
	return &folderNode{
		fileNode: fs.newFile(obj, 0, id, storage),
	}
}

func (fs *DeviceFs) newFile(obj mtp.ObjectInfo, size int64, id uint32, storage uint32) *fileNode {
	if obj.StorageID != storage {
		log.Printf("storage mismatch file %s on %d, parent storage %d",
			obj.Filename, obj.StorageID, storage)
	}
	if obj.CompressedSize != 0xFFFFFFFF {
		size = int64(obj.CompressedSize)
	}

	n := &fileNode{
		storageID: storage,
		obj:       &obj,
		Size:      size,
		fs:        fs,
		id:        id,
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

const NOPARENT_ID = 0xFFFFFFFF

func (fs *DeviceFs) OnMount(conn *fuse.FileSystemConnector) {
	for i, s := range fs.storageInfos {
		obj := mtp.ObjectInfo{
			ParentObject: NOPARENT_ID,
			StorageID:    fs.storages[i],
			Filename:     s.StorageDescription,
		}
		folder := fs.newFolder(obj, NOPARENT_ID, fs.storages[i])
		inode := fs.root.Inode().New(true, folder)
		fs.root.Inode().AddChild(s.StorageDescription, inode)
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
// files

type folder struct {
	id uint32
}

type fileNode struct {
	fuse.DefaultFsNode
	fs *DeviceFs

	storageID uint32

	// MTP handle.
	id uint32

	// This is needed because obj.CompressedSize only goes to
	// 0xFFFFFFFF
	Size int64

	// mtp *mtp.ObjectInfo for files
	obj *mtp.ObjectInfo

	// local file containing the contents.
	backing string

	// If set, the backing file was changed.
	write bool

	// If set, there was some error writing to the backing store;
	// don't flush file to device.
	error fuse.Status
}

func (n *fileNode) startEdit() bool {
	if n.write {
		return true
	}

	err := n.fs.dev.AndroidBeginEditObject(n.id)
	if err != nil {
		log.Println("AndroidBeginEditObject failed:", err)
		return false
	}
	n.write = true
	return true
}

func (n *fileNode) endEdit() bool {
	if !n.write {
		return true
	}

	err := n.fs.dev.AndroidEndEditObject(n.id)
	if err != nil {
		log.Println("AndroidEndEditObject failed:", err)
		return false
	}
	n.write = false
	return true
}

func (n *fileNode) Id() uint32 {
	return n.id
}

func (n *fileNode) StatFs() *fuse.StatfsOut {
	return n.fs.statFs()
}

func (n *fileNode) OnForget() {
	if n.obj != nil {
		n.obj = nil
		n.id = 0
	}
}

func (n *fileNode) Open(flags uint32, context *fuse.Context) (file fuse.File, code fuse.Status) {
	write := (flags&fuse.O_ANYWRITE != 0)
	if write {
		if !n.startEdit() {
			return nil, fuse.EIO
		}
	}

	return &androidFile{node: n}, fuse.OK
}

func (n *fileNode) Truncate(file fuse.File, size uint64, context *fuse.Context) (code fuse.Status) {
	w := n.write
	if !n.startEdit() {
		return fuse.EIO
	}
	err := n.fs.dev.AndroidTruncate(n.id, int64(size))
	if err != nil {
		log.Println("AndroidTruncate failed:", err)
		return fuse.EIO
	}
	n.Size = int64(size)

	if !w {
		if !n.endEdit() {
			return fuse.EIO
		}
	}
	return fuse.OK
}

func (n *fileNode) GetAttr(out *fuse.Attr, file fuse.File, context *fuse.Context) (code fuse.Status) {
	out.Mode = fuse.S_IFREG | 0644
	f := n.obj
	if f != nil {
		out.Size = uint64(n.Size)
		t := f.ModificationDate
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

func (n *fileNode) Utimens(file fuse.File, aTime *time.Time, mTime *time.Time, context *fuse.Context) (code fuse.Status) {
	// Unfortunately, we can't set the modtime; it's READONLY in
	// the Android MTP implementation. We just change the time in
	// the mount, but this is not persisted.
	if mTime != nil {
		n.obj.ModificationDate = *mTime
	}
	return fuse.OK
}

//////////////////
// folders

type folderNode struct {
	*fileNode
	fetched bool
}

// Keep the root nodes for all device storages alive.
func (n *folderNode) Deletable() bool {
	return n.Id() != NOPARENT_ID
}

// Fetches data from device returns false on failure.
func (n *folderNode) fetch() bool {
	if n.fetched {
		return true
	}

	handles := mtp.ObjectHandles{}
	err := n.fs.dev.GetObjectHandles(n.storageID, 0x0,
		n.id, &handles)
	if err != nil {
		log.Printf("GetObjectHandles failed: %v", err)
		return false
	}

	infos := []*mtp.ObjectInfo{}
	sizes := map[uint32]int64{}
	for _, handle := range handles.Handles {
		obj := mtp.ObjectInfo{}
		err := n.fs.dev.GetObjectInfo(handle, &obj)
		if err != nil {
			log.Printf("GetObjectInfo failed: %v", err)
			return false
		}
		if obj.Filename == "" {
			log.Printf("ignoring handle 0x%x with empty name in dir 0x%x",
				handle, n.id)
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

	for i, handle := range handles.Handles {
		var node fuse.FsNode
		if infos[i] == nil {
			continue
		}
		obj := *infos[i]
		obj.ParentObject = n.id
		isdir := obj.ObjectFormat == mtp.OFC_Association
		if isdir {
			fNode := n.fs.newFolder(obj, handle, n.storageID)
			node = fNode
		} else {
			sz := sizes[handle]
			node = n.fs.newFile(obj, sz, handle, n.storageID)
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

func toFileNode(n *fuse.Inode) *fileNode {
	switch f := n.FsNode().(type) {
	case *fileNode:
		return f
	case *folderNode:
		return f.fileNode
	}
	return nil
}

func (n *folderNode) basenameRename(oldName string, newName string) error {
	ch := n.Inode().GetChild(oldName)

	mFile := toFileNode(ch)

	if mFile.Id() != 0 {
		// Only rename on device if it was sent already.
		v := mtp.StringValue{newName}
		err := n.fs.dev.SetObjectPropValue(mFile.Id(), mtp.OPC_ObjectFileName, &v)
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
		ParentObject:     n.id,
		StorageID:        n.storageID,
	}
	_, _, newId, err := n.fs.dev.SendObjectInfo(n.storageID, n.id, &obj)
	if err != nil {
		log.Printf("CreateFolder failed: %v", err)
		return nil, fuse.EIO
	}

	f := n.fs.newFolder(obj, newId, n.storageID)
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

	f := toFileNode(ch)
	if f.id != 0 {
		err := n.fs.dev.DeleteObject(f.id)
		if err != nil {
			log.Printf("DeleteObject failed: %v", err)
			return fuse.EIO
		}
	} else {
		f.obj.Filename = ""
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
		StorageID:        n.storageID,
		Filename:         name,
		ObjectFormat:     mtp.OFC_Undefined,
		ModificationDate: time.Now(),
		ParentObject:     n.id,
		CompressedSize:   0,
	}

	_, _, handle, err := n.fs.dev.SendObjectInfo(n.storageID, n.id, &obj)
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

	fn := &fileNode{
		obj:       &obj,
		storageID: n.storageID,
		fs:        n.fs,
		id:        handle,
		write:     true,
	}

	n.Inode().AddChild(name, n.Inode().New(false, fn))
	p := &androidFile{
		node: fn,
	}

	return p, fn, fuse.OK
}
