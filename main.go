// Copyright 2012 Google Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"io/ioutil"
	"log"
	"os"
	"regexp"

	"github.com/hanwen/go-fuse/fuse"
)

func main() {
	fsdebug := flag.Bool("fs-debug", false, "switch on FS debugging")
	mtpDebug := flag.Int("mtp-debug", 0, "switch on MTP debugging. 1=PTP, 2=PLST, 4=USB, 8=DATA")
	backing := flag.String("backing-dir", "", "backing store for locally cached files. Default: use a temporary directory.")
	vfat := flag.Bool("vfat", true, "assume removable RAM media uses VFAT, and rewrite names.")
	other := flag.Bool("allow-other", false, "allow other users to access mounted fuse. Default: false.")
	deviceFilter := flag.String("dev", "", "regular expression to filter devices.")
	storageFilter := flag.String("storage", "", "regular expression to filter storage areas.")
	flag.Parse()

	if len(flag.Args()) != 1 {
		log.Fatalf("Usage: %s [options] MOUNT-POINT\n", os.Args[0])
	}
	mountpoint := flag.Arg(0)

	Init()
	SetDebug(*mtpDebug)
	devs, err := Detect()
	if err != nil {
		log.Fatalf("detect failed: %v", err)
	}

	if *deviceFilter != "" {
		re, err := regexp.Compile(*deviceFilter)
		if err != nil {
			log.Fatalf("invalid regexp %q: %v", *deviceFilter, err)
		}
		filtered := []*RawDevice{}
		for _, d := range devs {
			if re.FindStringIndex(d.String()) != nil { 
				filtered = append(filtered, d)
			} else {
				log.Printf("filtering out device %v: ", d)
			}
		}
		devs = filtered
	} else {
		for _, d := range devs {
			log.Printf("found device %v: ", d)
		}
	}
	
	if len(devs) == 0 {
		log.Fatal("no device found.  Try replugging it.")
	}
	if len(devs) > 1 {
		log.Fatal("must have exactly one device. Try using -dev")
	}

	rdev := devs[0]

	dev, err := rdev.Open()
	if err != nil {
		log.Fatalf("rdev.open failed: %v", err)
	}
	defer dev.Release()
	dev.GetStorage(0)
	
	storages := []*DeviceStorage{}
	for _, s := range dev.ListStorage()  {
		if !s.IsHierarchical() {
			log.Printf("skipping non hierarchical storage %q", s.Description())
			continue
		}
		storages = append(storages, s)
	}
	
	if *storageFilter != "" {
		re, err := regexp.Compile(*storageFilter)
		if err != nil {
			log.Fatalf("invalid regexp %q: %v", *storageFilter, err)
		}
		
		filtered := []*DeviceStorage{}
		for _, s := range storages {
			if re.FindStringIndex(s.Description()) == nil {
				log.Printf("filtering out storage %q", s.Description())
				continue
			}
			filtered = append(filtered, s)
		}

		if len(filtered) == 0 {
			log.Fatalf("No storages found. Try changing the -storage flag.")
		}
		storages = filtered
	} else {
		for _, s := range storages {
			log.Printf("storage ID %d: %s", s.Id(), s.Description())
		}
	}

	if len(storages) == 0 {
		log.Fatalf("No storages found. Try unlocking the device.")		
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
	fs := NewDeviceFs(dev, storages, opts)
	conn := fuse.NewFileSystemConnector(fs, fuse.NewFileSystemOptions())
	rawFs := fuse.NewLockingRawFileSystem(conn)

	mount := fuse.NewMountState(rawFs)
	mOpts := &fuse.MountOptions{
		AllowOther: *other,
	}
	if err := mount.Mount(mountpoint, mOpts); err != nil {
		log.Fatalf("mount failed: %v", err)
	}

	conn.Debug = *fsdebug
	mount.Debug = *fsdebug
	log.Printf("starting FUSE %v", fuse.Version())
	mount.Loop()
}
