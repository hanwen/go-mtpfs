package fs

// This test requires an unlocked android MTP device plugged in.

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/hanwen/go-fuse/fuse"
	"github.com/hanwen/go-fuse/fuse/nodefs"
	"github.com/hanwen/go-mtpfs/mtp"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func startFs(t *testing.T, useAndroid bool) (root string, cleanup func()) {
	dev, err := mtp.SelectDevice("")
	if err != nil {
		t.Fatalf("SelectDevice failed: %v", err)
	}
	defer func() {
		if dev != nil {
			dev.Close()
		}
	}()

	if err = dev.Configure(); err != nil {
		t.Fatalf("Configure failed: %v", err)
	}

	sids, err := SelectStorages(dev, "")
	if err != nil {
		t.Fatalf("selectStorages failed: %v", err)
	}
	if len(sids) == 0 {
		t.Fatal("no storages found. Unlock device?")
	}
	tempdir, err := ioutil.TempDir("", "mtpfs")
	if err != nil {
		t.Fatal(err)
	}
	opts := DeviceFsOptions{
		Android: useAndroid,
	}
	fs, err := NewDeviceFs(dev, sids, opts)
	if err != nil {
		t.Fatal("NewDeviceFs failed:", err)
	}
	conn := nodefs.NewFileSystemConnector(fs, nodefs.NewOptions())
	rawFs := fuse.NewLockingRawFileSystem(conn.RawFS())
	mount, err := fuse.NewServer(rawFs, tempdir, nil)
	if err != nil {
		t.Fatalf("mount failed: %v", err)
	}

	mount.SetDebug(fuse.VerboseTest())
	dev.MTPDebug = fuse.VerboseTest()
	dev.USBDebug = fuse.VerboseTest()
	dev.DataDebug = fuse.VerboseTest()
	go mount.Serve()

	for i := 0; i < 10; i++ {
		fis, err := ioutil.ReadDir(tempdir)
		if err == nil && len(fis) > 0 {
			root = filepath.Join(tempdir, fis[0].Name())
			break
		}
		time.Sleep(1)
	}

	if root == "" {
		mount.Unmount()
		t.Fatal("could not find entries in mount point.")
	}

	d := dev
	dev = nil
	return root, func() {
		mount.Unmount()
		d.Close()
	}
}

// Use this function to simulate improper connection handling of a
// predecessor.
func xTestBoom(t *testing.T) {
	root, clean := startFs(t, true)
	_ = root
	_ = clean
	go func() { panic("boom") }()
}

func testDevice(t *testing.T, useAndroid bool) {
	root, cleanup := startFs(t, useAndroid)
	defer cleanup()

	_, err := os.Lstat(root + "/Music")
	if err != nil {
		t.Logf("Music not found: %v", err)
	}

	name := filepath.Join(root, fmt.Sprintf("mtpfs-test-%x", rand.Int31()))
	golden := "abcpxq134"
	if err := ioutil.WriteFile(name, []byte("abcpxq134"), 0644); err != nil {
		t.Fatal("WriteFile failed", err)
	}
	got, err := ioutil.ReadFile(name)
	if err != nil {
		t.Fatal("ReadFile failed", err)
	}

	if string(got) != golden {
		t.Fatalf("got %q, want %q", got, golden)
	}

	f, err := os.OpenFile(name, os.O_RDWR|os.O_APPEND, 0644)
	if err != nil {
		t.Fatal("OpenFile failed:", err)
	}
	defer f.Close()

	n, _ := f.ReadAt(make([]byte, 4096), 4096)
	if n > 0 {
		t.Fatalf("beyond EOF read should fail: got %d bytes", n)
	}

	golden += "hello"
	_, err = f.Write([]byte("hello"))
	if err != nil {
		t.Fatal("file.Write failed", err)
	}
	err = f.Close()
	if err != nil {
		t.Fatal("Close failed", err)
	}

	got, err = ioutil.ReadFile(name)
	if err != nil {
		t.Fatal("ReadFile failed", err)
	}

	if string(got) != golden {
		t.Fatalf("got %q, want %q", got, golden)
	}

	newName := filepath.Join(root, fmt.Sprintf("mtpfs-test-%x", rand.Int31()))
	err = os.Rename(name, newName)
	if err != nil {
		t.Fatal("Rename failed", err)
	}

	if fi, err := os.Lstat(name); err == nil {
		t.Fatal("should have disappeared after rename", fi)
	}

	if _, err := os.Lstat(newName); err != nil {
		t.Fatal("should be able to stat after rename", err)
	}

	err = os.Remove(newName)
	if err != nil {
		t.Fatal("Remove failed", err)
	}
	if fi, err := os.Lstat(newName); err == nil {
		t.Fatal("should have disappeared after Remove", fi)
	}
}

func TestAndroid(t *testing.T) {
	testDevice(t, true)
}

func TestNormal(t *testing.T) {
	testDevice(t, false)
}
