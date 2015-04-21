// Copyright 2012 Google Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"log"
	"os"
	"strings"
	"sync"

	"github.com/hanwen/go-fuse/fuse"
	"github.com/hanwen/go-fuse/fuse/nodefs"
	"github.com/hanwen/go-mtpfs/fs"
	"github.com/hanwen/go-mtpfs/mtp"
)

func main() {
	debug := flag.String("debug", "", "comma-separated list of debugging options: usb, data, mtp, fuse")
	usbTimeout := flag.Int("usb-timeout", 5000, "timeout in milliseconds")
	vfat := flag.Bool("vfat", true, "assume removable RAM media uses VFAT, and rewrite names.")
	other := flag.Bool("allow-other", false, "allow other users to access mounted fuse. Default: false.")
	deviceFilter := flag.String("dev", "",
		"regular expression to filter device IDs, "+
			"which are composed of manufacturer/product/serial.")
	storageFilter := flag.String("storage", "", "regular expression to filter storage areas.")
	android := flag.Bool("android", true, "use android extensions if available")
	skipNullRead := flag.Bool("skip-null-read", false, "skip null reads. Workaround for Linux xhci bug.")
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
	debugs := map[string]bool{}
	for _, s := range strings.Split(*debug, ",") {
		debugs[s] = true
	}
	dev.MTPDebug = debugs["mtp"]
	dev.DataDebug = debugs["data"]
	dev.USBDebug = debugs["usb"]
	dev.Timeout = *usbTimeout
	dev.SkipNullRead = *skipNullRead
	if err = dev.Configure(); err != nil {
		log.Fatalf("Configure failed: %v", err)
	}

	sids, err := fs.SelectStorages(dev, *storageFilter)
	if err != nil {
		log.Fatalf("selectStorages failed: %v", err)
	}

	opts := fs.DeviceFsOptions{
		RemovableVFat: *vfat,
		Android:       *android,
	}
	root, err := fs.NewDeviceFSRoot(dev, sids, opts)
	if err != nil {
		log.Fatalf("NewDeviceFs failed: %v", err)
	}
	conn := nodefs.NewFileSystemConnector(root, nodefs.NewOptions())
	rawFs := fuse.NewLockingRawFileSystem(conn.RawFS())

	mOpts := &fuse.MountOptions{
		AllowOther: *other,
	}
	mount, err := fuse.NewServer(rawFs, mountpoint, mOpts)
	if err != nil {
		log.Fatalf("mount failed: %v", err)
	}

	conn.SetDebug(debugs["fuse"] || debugs["fs"])
	mount.SetDebug(debugs["fuse"] || debugs["fs"])
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		mount.Serve()
		wg.Done()
	}()
	mount.WaitMount()
	log.Printf("FUSE mounted")
	wg.Wait()
	root.OnUnmount()
}
