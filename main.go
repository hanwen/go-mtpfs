// Copyright 2012 Google Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"strings"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/hanwen/go-mtpfs/mtp"
)

func main() {
	debug := flag.String("debug", "", "comma-separated list of debugging options: usb, data, mtp, fuse")
	usbTimeout := flag.Int("usb-timeout", 5000, "timeout in milliseconds")
	deviceFilter := flag.String("dev", "",
		"regular expression to filter device IDs, "+
			"which are composed of manufacturer/product/serial.")
	flag.Parse()

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
	if err = dev.Configure(); err != nil {
		log.Fatalf("Configure failed: %v", err)
	}

	lvs := mtp.NewLVServer(dev)

	eg, ctx := errgroup.WithContext(context.Background())
	eg.Go(lvs.Run)

	srv := http.Server{Addr: "localhost:42839"}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "index.html")
	})

	http.HandleFunc("/view", lvs.HandleClient)

	eg.Go(func() error {
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			return err
		}
		return nil
	})

	eg.Go(func() error {
		select {
		case <-ctx.Done():
		}
		ctx, _ := context.WithTimeout(context.Background(), 3*time.Second)
		return srv.Shutdown(ctx)
	})

	err = eg.Wait()
	if err != nil {
		log.Fatalf(err.Error())
	}
}
