// Copyright 2012 Google Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/google/gousb"

	"golang.org/x/sync/errgroup"

	"github.com/hanwen/go-mtpfs/mtp"
	"github.com/hanwen/go-mtpfs/public"
	"github.com/sirupsen/logrus"
)

func main() {
	host := flag.String("host", "localhost", "hostname: default = localhost, specify 0.0.0.0 for public access")
	port := flag.Int("port", 42839, "port: default = 42839")
	debug := flag.String("debug", "", "comma-separated list of debugging options: usb, data, mtp, server")
	serverOnly := flag.Bool("server-only", false, "serve frontend without opening a DSLR (for devevelopment)")
	/*
		usbTimeout := flag.Int("usb-timeout", 5000, "timeout in milliseconds")
		deviceFilter := flag.String("dev", "",
			"regular expression to filter device IDs, "+
				"which are composed of manufacturer/product/serial.")
	*/
	flag.Parse()

	debugs := map[string]bool{}
	for _, s := range strings.Split(*debug, ",") {
		debugs[s] = true
	}

	if _, ok := debugs["server"]; ok {
		log.Level = logrus.DebugLevel
	}

	var dev *mtp.Device2
	var err error

	if *serverOnly {
		log.WithField("prefix", "mtp").Info("Server-only mode is activated, skipping USB initialization")
	} else {
		ctx2 := gousb.NewContext()
		defer ctx2.Close()

		ctx2.Debug(999)

		dev, err = mtp.FindDevice(ctx2, 0, 0)
		if err != nil {
			log.WithField("prefix", "mtp").Fatalf("Failed to find MTP device: %s", err)
		}
		dev.AttachLogger(log)
		dev.Debug.MTP = debugs["mtp"]
		dev.Debug.Data = debugs["data"]
		dev.Debug.USB = debugs["usb"]
		if err = dev.Configure(); err != nil {
			log.WithField("prefix", "mtp").Fatalf("Configure failed: %v", err)
		}
	}

	/*
		var dev *mtp.Device
		var err error

		if *serverOnly {
			log.WithField("prefix", "mtp").Info("Server-only mode is activated, skipping USB initialization")
		} else {
			dev, err = mtp.SelectDevice(*deviceFilter)
			if err != nil {
				log.WithField("prefix", "mtp").Fatalf("Failed to detect MTP devices: %v", err)
			}
			defer dev.Close()
			dev.MTPDebug = debugs["mtp"]
			dev.DataDebug = debugs["data"]
			dev.USBDebug = debugs["usb"]
			dev.Timeout = *usbTimeout
			if err = dev.Configure(); err != nil {
				log.WithField("prefix", "mtp").Fatalf("Configure failed: %v", err)
			}
		}
	*/

	eg, ctx := errgroup.WithContext(context.Background())

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	eg.Go(func() error {
		for {
			select {
			case <-ctx.Done():
				return nil
			case s := <-sigChan:
				log.WithField("prefix", "signal").Info("Caught signal: %s", s)
				return errors.New(s.String())
			}
		}
	})

	lvs := mtp.NewLVServer(dev, log, ctx)
	eg.Go(lvs.Run)

	router := http.NewServeMux()
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		f, _ := public.Root.Open("/controller.html")
		_, _ = io.Copy(w, f)
	})
	router.HandleFunc("/view", func(w http.ResponseWriter, r *http.Request) {
		f, _ := public.Root.Open("/index.html")
		_, _ = io.Copy(w, f)
	})
	router.HandleFunc("/stream", lvs.HandleStream)
	router.HandleFunc("/control", lvs.HandleControl)
	router.Handle("/assets/", http.FileServer(public.Root))

	srv := http.Server{
		Addr:    fmt.Sprintf("%s:%d", *host, *port),
		Handler: logging(router),
	}

	eg.Go(func() error {
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			log.WithField("prefix", "http").Error(err)
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
		os.Exit(1)
	}
}
