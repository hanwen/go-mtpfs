// Copyright 2012 Google Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"context"
	"errors"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/hanwen/go-mtpfs/mtp"
	"github.com/hanwen/go-mtpfs/static"
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

	eg, ctx := errgroup.WithContext(context.Background())

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	eg.Go(func() error {
		for {
			select {
			case <-ctx.Done():
				return nil
			case s := <-sigChan:
				return errors.New(s.String())
			}
		}
	})

	lvs := mtp.NewLVServer(dev, ctx)
	eg.Go(lvs.Run)

	srv := http.Server{Addr: "localhost:42839"}
	eg.Go(func() error {
		http.Handle("/", http.FileServer(static.Root))
		http.HandleFunc("/view", lvs.HandleClient)

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
