// Copyright 2012 Google Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"github.com/hanwen/go-fuse/fuse"
	"io/ioutil"
	"log"
	"os"
)

func main() {
	fsdebug := flag.Bool("fs-debug", false, "switch on FS debugging")
	mtpDebug := flag.Int("mtp-debug", 0, "switch on MTP debugging")
	backing := flag.String("backing-dir", "", "backing store for locally cached files. Default: use a temporary directory.")
	vfat := flag.Bool("vfat", true, "assume removable RAM media uses VFAT, and rewrite names.")
	flag.Parse()

	if len(flag.Args()) != 1 {
		log.Fatal("Usage: mtpfs MOUNT-POINT")
	}
	mountpoint := flag.Arg(0)

	Init()
	SetDebug(*mtpDebug)
	devs, err := Detect()
	if err != nil {
		log.Fatalf("detect: %v", err)
	}
	for _, d := range devs {
		log.Printf("device %v: ", d)
	}
	if len(devs) == 0 {
		log.Fatal("no device found.  Try replugging it.")
	}
	if len(devs) > 1 {
		log.Fatal("must have exactly one device")
	}

	rdev := devs[0]

	dev, err := rdev.Open()
	if err != nil {
		log.Fatalf("rdev.open: %v", err)
	}
	defer dev.Release()
	dev.GetStorage(0)
	for _, s := range dev.ListStorage() {
		log.Printf("storage ID %d: %s", s.Id(), s.Description())
	}

	if len(dev.ListStorage()) == 0 {
		log.Fatalf("No storages found.  Try unlocking the device.")
	}

	if *backing == "" {
		*backing, err = ioutil.TempDir("", "go-mtpfs")
		if err != nil {
			log.Fatalf("TempDir failed: %v", err)
		}
	} else {
		*backing = *backing + "/go-mtpfs"
		err = os.Mkdir(*backing, 0700)
		if err != nil {
			log.Fatalf("Mkdir failed: %v", err)
		}
	}
	log.Println("backing data", *backing)
	defer os.RemoveAll(*backing)

	opts := DeviceFsOptions{
		Dir:           *backing,
		RemovableVFat: *vfat,
	}
	fs := NewDeviceFs(dev, &opts)
	conn := fuse.NewFileSystemConnector(fs, fuse.NewFileSystemOptions())
	rawFs := fuse.NewLockingRawFileSystem(conn)

	mount := fuse.NewMountState(rawFs)
	if err := mount.Mount(mountpoint, nil); err != nil {
		log.Fatalf("mount failed: %v", err)
	}

	conn.Debug = *fsdebug
	mount.Debug = *fsdebug
	log.Println("starting FUSE.")
	mount.Loop()
}
