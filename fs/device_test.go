package fs

// This test requires an unlocked android MTP device plugged in.

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/hanwen/go-fuse/fs"
	"github.com/hanwen/go-fuse/fuse"
	"github.com/hanwen/go-mtpfs/mtp"
)

// VerboseTest returns true if the testing framework is run with -v.
func VerboseTest() bool {
	flag := flag.Lookup("test.v")
	return flag != nil && flag.Value.String() == "true"
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

func startFs(t *testing.T, useAndroid bool) (storageRoot string, cleanup func()) {
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
	root, err := NewDeviceFSRoot(dev, sids, opts)
	if err != nil {
		t.Fatal("NewDeviceFs failed:", err)
	}
	server, err := fs.Mount(tempdir, root,
		&fs.Options{
			MountOptions: fuse.MountOptions{
				SingleThreaded: true,
				Debug:          VerboseTest(),
			},
		})
	if err != nil {
		t.Fatalf("mount failed: %v", err)
	}

	dev.MTPDebug = VerboseTest()
	/*	dev.USBDebug = VerboseTest()
		dev.DataDebug = VerboseTest()
	*/
	go server.Serve()
	server.WaitMount()

	for i := 0; i < 10; i++ {
		fis, err := ioutil.ReadDir(tempdir)
		if err == nil && len(fis) > 0 {
			storageRoot = filepath.Join(tempdir, fis[0].Name())
			break
		}
		time.Sleep(1)
	}

	if storageRoot == "" {
		server.Unmount()
		t.Fatal("could not find entries in mount point.")
	}

	d := dev
	dev = nil
	return storageRoot, func() {
		server.Unmount()
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

	dirName := filepath.Join(root, fmt.Sprintf("mtpfs-dir-test.x:y-%x", rand.Int31()))
	if err := os.Mkdir(dirName, 0755); err != nil {
		t.Fatalf("Mkdir: %v", err)
	}

	if err := os.Remove(dirName); err != nil {
		t.Fatalf("Rmdir: %v", err)
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

func testReadBlockBoundary(t *testing.T, android bool) {
	root, cleanup := startFs(t, android)
	defer cleanup()

	name := filepath.Join(root, fmt.Sprintf("mtpfs-test-%x", rand.Int31()))

	page := 4096
	buf := bytes.Repeat([]byte("a"), 32*page)
	if err := ioutil.WriteFile(name, buf, 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	f, err := os.Open(name)
	if err != nil {
		t.Fatalf("Open(%q): %v", name, err)
	}

	total := 0
	for {
		b := make([]byte, page)
		n, err := f.Read(b)
		total += n
		if n == 0 && err == io.EOF {
			break
		}
		if n != 4096 || err != nil {
			t.Fatalf("Read: %v (%d bytes)", err, n)
		}
	}
	f.Close()
}

func TestReadBlockBoundaryAndroid(t *testing.T) {
	testReadBlockBoundary(t, true)
}

func TestReadBlockBoundaryNormal(t *testing.T) {
	testReadBlockBoundary(t, false)
}

func TestAndroid(t *testing.T) {
	testDevice(t, true)
}

func TestNormal(t *testing.T) {
	testDevice(t, false)
}
