// Copyright 2012 Google Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package fs

import (
	"bytes"
	"fmt"
	"log"

	"github.com/hanwen/go-fuse/fuse"
)

type androidFile struct {
	fuse.DefaultFile
	node *fileNode
}

func (f *androidFile) Read(dest []byte, off int64) (fuse.ReadResult, fuse.Status) {
	b := bytes.NewBuffer(dest[:0])
	err := f.node.fs.dev.AndroidGetPartialObject64(f.node.id, b, off, uint32(len(dest)))
	if err != nil {
		log.Println("AndroidGetPartialObject64 failed:", err)
		return nil, fuse.EIO
	}

	return &fuse.ReadResultData{dest[:b.Len()]}, fuse.OK
}

func (f *androidFile) String() string {
	return fmt.Sprintf("androidFile h=0x%x", f.node.id)
}

func (f *androidFile) Write(dest []byte, off int64) (written uint32, status fuse.Status) {
	b := bytes.NewBuffer(dest)
	err := f.node.fs.dev.AndroidSendPartialObject(f.node.id, off, uint32(len(dest)), b)
	if err != nil {
		log.Println("AndroidSendPartialObject failed:", err)
		return 0, fuse.EIO
	}
	written = uint32(len(dest)-b.Len())
	if off + int64(written) > f.node.Size {
		f.node.Size = off + int64(written)
	}
	return written, fuse.OK
}

func (f *androidFile) Release() {
	f.node.endEdit()
}

