// Copyright 2012 Google Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"log"
	"reflect"
	"strings"
	"time"
	"unsafe"
)

// #cgo LDFLAGS: -lmtp -L/usr/local/lib
// #include <libmtp.h>
// #include <stdlib.h>
import "C"

var _ = log.Println

/*
This file has a partial cgo wrapping for libmtp, so users should
never have to call import "C"
*/

// TODO - we leak C strings all over the place.
	
type Device C.LIBMTP_mtpdevice_t
type Error C.LIBMTP_error_t
type MtpError C.LIBMTP_error_number_t
type RawDevice C.LIBMTP_raw_device_t
type DeviceStorage C.LIBMTP_devicestorage_t
type Folder C.LIBMTP_folder_t
type File C.LIBMTP_file_t

const NOPARENT_ID = 0xFFFFFFFF
const DEBUG_PTP = int(C.LIBMTP_DEBUG_PTP)
const DEBUG_PLST = int(C.LIBMTP_DEBUG_PLST)
const DEBUG_USB = int(C.LIBMTP_DEBUG_USB)
const DEBUG_DATA = int(C.LIBMTP_DEBUG_DATA)
const DEBUG_ALL = int(C.LIBMTP_DEBUG_ALL)
const FILETYPE_FOLDER = int(C.LIBMTP_FILETYPE_FOLDER)
const FILETYPE_UNKNOWN = C.LIBMTP_FILETYPE_UNKNOWN

func Init() {
	C.LIBMTP_Init()
}

func SetDebug(mask int) {
	C.LIBMTP_Set_Debug(C.int(mask))
}

func Detect() (devs []*RawDevice, err error) {
	var rawdevices *C.LIBMTP_raw_device_t
	var numrawdevices C.int

	errno := MtpError(C.LIBMTP_Detect_Raw_Devices(&rawdevices, &numrawdevices))
	if errno == C.LIBMTP_ERROR_NO_DEVICE_ATTACHED {
		return nil, nil
	}
	if errno != C.LIBMTP_ERROR_NONE {
		return nil, errno
	}
	slice := &reflect.SliceHeader{uintptr(unsafe.Pointer(rawdevices)), int(numrawdevices), int(numrawdevices)}
	rdevs := *(*[]RawDevice)(unsafe.Pointer(slice))

	for _, d := range rdevs {
		newD := d
		devs = append(devs, &newD)
	}

	// todo dealloc rawdevices
	return devs, nil
}

func (e MtpError) Error() string {
	switch e {
	case C.LIBMTP_ERROR_CONNECTING:
		return "error connecting"
	case C.LIBMTP_ERROR_MEMORY_ALLOCATION:
		return "memory allocation error"
	case C.LIBMTP_ERROR_GENERAL:
		return "general error."
	}
	return "unknown error"
}

////////////////
// Raw devices.

func (r *RawDevice) me() *C.LIBMTP_raw_device_t {
	return (*C.LIBMTP_raw_device_t)(r)
}

func (r *RawDevice) Open() (*Device, error) {
	dev := C.LIBMTP_Open_Raw_Device_Uncached(r.me())
	if dev == nil {
		return nil, fmt.Errorf("open: open returned nil")
	}
	return (*Device)(dev), nil
}

func (d *RawDevice) String() string {
	vendor := "unknown"
	if d.device_entry.vendor != nil {
		vendor = C.GoString(d.device_entry.vendor)
	}
	product := "unknown"
	if d.device_entry.product != nil {
		product = C.GoString(d.device_entry.product)
	}

	return fmt.Sprintf("%s: %s (%04x:%04x) @ bus %d, dev %d\n",
		vendor, product,
		d.device_entry.vendor_id,
		d.device_entry.product_id,
		d.bus_location,
		d.devnum)
}

////////////////
// Device

func (d *Device) me() *C.LIBMTP_mtpdevice_t {
	return (*C.LIBMTP_mtpdevice_t)(d)
}

func (d *Device) Release() {
	C.LIBMTP_Release_Device(d.me())
}

func (d *Device) ErrorStack() error {
	var messages []string
	for p := C.LIBMTP_Get_Errorstack(d.me()); p != nil; p = p.next {
		messages = append(messages, C.GoString(p.error_text))
	}
	if len(messages) == 0 {
		return nil
	}
	C.LIBMTP_Clear_Errorstack(d.me())
	return fmt.Errorf("%s", strings.Join(messages, "\n"))
}

func (d *Device) GetStorage(sortOrder int) error {
	err := C.LIBMTP_Get_Storage(d.me(), C.int(sortOrder))
	if err != 0 {
		return d.ErrorStack()
	}
	return nil
}

func (d *Device) FilesAndFolders(storageId uint32, parentId uint32) (files []*File, err error) {
	file := C.LIBMTP_Get_Files_And_Folders(d.me(), C.uint32_t(storageId), C.uint32_t(parentId))
	for f := (*File)(file); f != nil; f = (*File)(f.next) {
		files = append(files, f)
	}
	return files, d.ErrorStack()
}

func (d *Device) FolderList(s *DeviceStorage) (folders []*Folder) {
	folder := C.LIBMTP_Get_Folder_List_For_Storage(d.me(), s.id)
	for f := (*Folder)(folder); f != nil; f = (*Folder)(f.sibling) {
		folders = append(folders, f)
	}
	return folders
}

func (d *Device) ListStorage() (storages []*DeviceStorage) {
	for p := d.storage; p != nil; p = p.next {
		storages = append(storages, (*DeviceStorage)(p))
	}
	return
}
func (d *Device) Filemetadata(id uint32) (*File, error) {
	f := (C.LIBMTP_Get_Filemetadata(d.me(), C.uint32_t(id)))
	return (*File)(f), d.ErrorStack()
}

func (d *Device) GetFileToFileDescriptor(id uint32, fd uintptr) error {
	code := C.LIBMTP_Get_File_To_File_Descriptor(d.me(), C.uint32_t(id), C.int(fd), nil, nil)
	if code != 0 {
		return d.ErrorStack()
	}

	return nil
}

func (d *Device) SendFromFileDescriptor(file *File, fd uintptr) error {
	code := C.LIBMTP_Send_File_From_File_Descriptor(d.me(), C.int(fd), (*C.LIBMTP_file_t)(file), nil, nil)
	if code != 0 {
		return d.ErrorStack()
	}

	return nil
}

func (d *Device) CreateFolder(parent uint32, name string, storage uint32) (uint32, error) {
	cname := C.CString(name)
	id := C.LIBMTP_Create_Folder(d.me(), cname, C.uint32_t(parent), C.uint32_t(storage))

	if newName := C.GoString(cname); newName != name {
		log.Println("Folder name changed to %q", newName)
	}
	C.free(unsafe.Pointer(cname))
	return uint32(id), d.ErrorStack()
}

func (d *Device) SetFileName(f *File, name string) error {
	C.LIBMTP_Set_File_Name(d.me(), f.me(), C.CString(name))
	return d.ErrorStack()
}

func (d *Device) DeleteObject(id uint32) error {
	c := C.LIBMTP_Delete_Object(d.me(), C.uint32_t(id))
	if c != 0 {
		return d.ErrorStack()
	}
	return nil
}

func (d *Device) GetStringFromObject(id uint32, prop int) (string, error) {
	mId := C.uint32_t(id)
	mProp := C.LIBMTP_property_t(prop)
	cstr := C.LIBMTP_Get_String_From_Object(d.me(), mId,  mProp)
	result := C.GoString(cstr)
	C.free(unsafe.Pointer(cstr))

	return result, d.ErrorStack()
}

func (d *Device) GetIntFromObject(id uint32, prop int, bits int) (uint64, error) {
	var result uint64
	mId := C.uint32_t(id)
	mProp := C.LIBMTP_property_t(prop)
	switch (bits) {
	case 64:
		result = uint64(C.LIBMTP_Get_u64_From_Object(d.me(), mId,  mProp, 0))
	case 32:
		result = uint64(C.LIBMTP_Get_u32_From_Object(d.me(), mId,  mProp, 0))
	case 16:
		result = uint64(C.LIBMTP_Get_u16_From_Object(d.me(), mId,  mProp, 0))
	case 8:
		result = uint64(C.LIBMTP_Get_u8_From_Object(d.me(), mId,  mProp, 0))
	default:
		return 0, fmt.Errorf("unsupported bit size %d", bits)
	}
	
	return result, d.ErrorStack()
}
const DateModified = C.LIBMTP_PROPERTY_DateModified

////////////////
// DeviceStorage

func (d *DeviceStorage) me() *C.LIBMTP_devicestorage_t {
	return (*C.LIBMTP_devicestorage_t)(d)
}

func (d *DeviceStorage) Id() uint32 {
	return uint32(d.me().id)
}

func (d *DeviceStorage) Description() string {
	return C.GoString(d.StorageDescription)
}

////////////////
// Files.

func NewFile(id uint32, parent uint32, storage_id uint32, filename string, size uint64,
	mtime time.Time, fileType int) *File {
	f := C.LIBMTP_new_file_t()
	f.item_id = C.uint32_t(id)
	f.parent_id = C.uint32_t(parent)
	f.storage_id = C.uint32_t(storage_id)
	f.filename = C.CString(filename)
	f.filesize = C.uint64_t(size)
	f.modificationdate = C.time_t(mtime.Unix())
	f.filetype = C.LIBMTP_filetype_t(fileType)
	return (*File)(f)
}

func (f *File) Destroy() {
	C.LIBMTP_destroy_file_t(f.me())
}

func (f *File) me() *C.LIBMTP_file_t {
	return (*C.LIBMTP_file_t)(f)
}

func (d *File) StorageId() uint32 {
	return uint32(d.storage_id)
}

func (f *File) Mtime() time.Time {
	return time.Unix(int64(f.modificationdate), 0)
}

func (f *File) SetMtime(t time.Time) {
	f.modificationdate = C.time_t(t.Unix())
}

func (f *File) SetFilesize(sz uint64) {
	f.filesize = C.uint64_t(sz)
}

func (d *File) Id() uint32 {
	return uint32(d.item_id)
}

func (d *File) Filetype() int {
	return int(d.filetype)
}

func (d *File) Name() string {
	return C.GoString(d.filename)
}

func (f *File) SetName(n string) {
	old := f.filename
	C.free(unsafe.Pointer(old))
	f.filename = C.CString(n)
}
