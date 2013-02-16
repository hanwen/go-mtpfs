// Copyright 2012 Google Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"log"
	"os"
	"strings"

	"github.com/hanwen/go-fuse/fuse"
	"github.com/hanwen/go-mtpfs/fs"
	"github.com/hanwen/go-mtpfs/mtp"
)

func main() {
	debug := flag.String("debug", "", "comma-separated list of debugging options: usb, data, mtp, fuse")
	usbTimeout := flag.Int("usb-timeout", 2000, "timeout in milliseconds")
	vfat := flag.Bool("vfat", true, "assume removable RAM media uses VFAT, and rewrite names.")
	other := flag.Bool("allow-other", false, "allow other users to access mounted fuse. Default: false.")
	deviceFilter := flag.String("dev", "", "regular expression to filter devices.")
	storageFilter := flag.String("storage", "", "regular expression to filter storage areas.")
	android := flag.Bool("android", true, "use android extensions if available")
	flag.Parse()

	if len(flag.Args()) != 1 {
		log.Fatalf("Usage: %s [options] MOUNT-POINT\n", os.Args[0])
	}
	mountpoint := flag.Arg(0)

	dev, err := mtp.SelectDevice(*deviceFilter)
	if err != nil {
		log.Fatalf("detect failed: %v", err)
	}
	defer dev.Close()

	dev.Timeout = *usbTimeout
	if err = dev.Configure(); err != nil {
		log.Fatalf("Configure failed: %v", err)
	}

	sids, err := fs.SelectStorages(dev, *storageFilter)
	if err != nil {
		log.Fatalf("selectStorages failed: %v", err)
	}

	debugs := map[string]bool{}
	for _, s := range strings.Split(*debug, ",") {
		debugs[s] = true
	}
	dev.MTPDebug = debugs["mtp"]
	dev.DataDebug = debugs["data"]
	dev.USBDebug = debugs["usb"]
	
	opts := fs.DeviceFsOptions{
		RemovableVFat: *vfat,
		Android: *android,
	}
	fs, err := fs.NewDeviceFs(dev, sids, opts)
	if err != nil {
		log.Fatalf("NewDeviceFs failed: %v", err)
	}
	conn := fuse.NewFileSystemConnector(fs, fuse.NewFileSystemOptions())
	rawFs := fuse.NewLockingRawFileSystem(conn)

	mount := fuse.NewMountState(rawFs)
	mOpts := &fuse.MountOptions{
		AllowOther: *other,
	}
	if err := mount.Mount(mountpoint, mOpts); err != nil {
		log.Fatalf("mount failed: %v", err)
	}

	conn.Debug = debugs["fuse"]
	mount.Debug = debugs["fuse"]
	log.Printf("starting FUSE %v", fuse.Version())
	mount.Loop()
	fs.OnUnmount()
}
