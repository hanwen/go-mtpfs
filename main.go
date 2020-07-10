// Copyright 2012 Google Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"context"
	"errors"
	"flag"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/hanwen/go-mtpfs/mtp"
	"github.com/hanwen/go-mtpfs/public"
)

func main() {
	debug := flag.String("debug", "", "comma-separated list of debugging options: usb, data, mtp, fuse")
	serverOnly := flag.Bool("server-only", false, "serve frontend without opening a DSLR (for devevelopment)")
	usbTimeout := flag.Int("usb-timeout", 5000, "timeout in milliseconds")
	deviceFilter := flag.String("dev", "",
		"regular expression to filter device IDs, "+
			"which are composed of manufacturer/product/serial.")
	flag.Parse()

	logger := log.New(os.Stdout, "http: ", log.LstdFlags)

	var dev *mtp.Device
	var err error

	if *serverOnly {
		logger.Println("Server-only mode is activated, skipping USB initialization")
	} else {
		dev, err = mtp.SelectDevice(*deviceFilter)
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
		Addr:    "0.0.0.0:42839",
		Handler: logging2(logger, router),
	}

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

func logging2(logger *log.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			logger.Println(r.Method, r.URL.Path, r.RemoteAddr)
		}()
		next.ServeHTTP(w, r)
	})
}
