// Copyright 2012 Google Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package fs

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"syscall"
	"time"

	"github.com/hanwen/go-fuse/fs"
	"github.com/hanwen/go-fuse/fuse"
)

type androidNode struct {
	mtpNodeImpl

	// If set, the backing file was changed.
	write     bool
	start     time.Time
	byteCount int64
}

func (n *androidNode) startEdit() bool {
	if n.write {
		return true
	}

	n.start = time.Now()
	n.byteCount = 0
	if err := n.fs.dev.AndroidBeginEditObject(n.Handle()); err != nil {
		log.Println("AndroidBeginEditObject failed:", err)
		return false
	}
	n.write = true
	return true
}

func (n *androidNode) endEdit() bool {
	if !n.write {
		return true
	}

	dt := time.Now().Sub(n.start)
	log.Printf("%d bytes in %v: %d mb/s",
		n.byteCount, dt, (1e3*n.byteCount)/(dt.Nanoseconds()))

	if err := n.fs.dev.AndroidEndEditObject(n.Handle()); err != nil {
		log.Println("AndroidEndEditObject failed:", err)
		return false
	}
	n.write = false
	return true
}

var _ = (fs.NodeOpener)((*androidNode)(nil))

func (n *androidNode) Open(ctx context.Context, flags uint32) (file fs.FileHandle, fuseFlags uint32, code syscall.Errno) {
	return &androidFile{
		node: n,
	}, 0, 0
}

var _ = (fs.NodeSetattrer)((*androidNode)(nil))

func (n *androidNode) Setattr(ctx context.Context, file fs.FileHandle, in *fuse.SetAttrIn, out *fuse.AttrOut) (code syscall.Errno) {
	if size, ok := in.GetSize(); ok {
		w := n.write
		if !n.startEdit() {
			return syscall.EIO
		}
		if err := n.fs.dev.AndroidTruncate(n.Handle(), int64(size)); err != nil {
			log.Println("AndroidTruncate failed:", err)
			return syscall.EIO
		}
		n.Size = int64(size)

		if !w {
			if !n.endEdit() {
				return syscall.EIO
			}
		}

		out.Size = size
		out.Mode = syscall.S_IFREG | 0644
	}
	return 0
}

var _ = mtpNode((*androidNode)(nil))

type androidFile struct {
	fs.FileHandle
	node *androidNode
}

var _ = (fs.FileReader)((*androidFile)(nil))

func (f *androidFile) Read(ctx context.Context, dest []byte, off int64) (fuse.ReadResult, syscall.Errno) {
	if off > f.node.Size {
		// ENXIO = no such address.
		return nil, syscall.Errno(int(syscall.ENXIO))
	}

	if off+int64(len(dest)) > f.node.Size {
		dest = dest[:f.node.Size-off]
	}
	b := bytes.NewBuffer(dest[:0])
	err := f.node.fs.dev.AndroidGetPartialObject64(f.node.Handle(), b, off, uint32(len(dest)))
	if err != nil {
		log.Println("AndroidGetPartialObject64 failed:", err)
		return nil, syscall.EIO
	}

	return fuse.ReadResultData(dest[:b.Len()]), 0
}

func (f *androidFile) String() string {
	return fmt.Sprintf("androidFile h=0x%x", f.node.Handle())
}

var _ = (fs.FileWriter)((*androidFile)(nil))

func (f *androidFile) Write(ctx context.Context, dest []byte, off int64) (written uint32, status syscall.Errno) {
	if !f.node.startEdit() {
		return 0, syscall.EIO
	}
	f.node.byteCount += int64(len(dest))
	b := bytes.NewBuffer(dest)
	err := f.node.fs.dev.AndroidSendPartialObject(f.node.Handle(), off, uint32(len(dest)), b)
	if err != nil {
		log.Println("AndroidSendPartialObject failed:", err)
		return 0, syscall.EIO
	}
	written = uint32(len(dest) - b.Len())
	if off+int64(written) > f.node.Size {
		f.node.Size = off + int64(written)
	}
	return written, 0
}

var _ = (fs.FileFlusher)((*androidFile)(nil))

func (f *androidFile) Flush(ctx context.Context) syscall.Errno {
	if !f.node.endEdit() {
		return syscall.EIO
	}
	return 0
}
