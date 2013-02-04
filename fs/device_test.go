package fs

// This test requires an unlocked android MTP device plugged in.

import (
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/hanwen/go-fuse/fuse"
	"github.com/hanwen/go-mtpfs/mtp"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func TestDevice(t *testing.T) {
	dev, err := mtp.SelectDevice("")
	if err != nil {
		t.Fatalf("SelectDevice failed: %v", err)
	}
	defer dev.Close()

	if err = dev.OpenSession(); err != nil {
		t.Fatalf("OpenSession failed: %v", err)
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
	opts := DeviceFsOptions{}
	fs, err := NewDeviceFs(dev, sids, opts)
	if err != nil {
		t.Fatal("NewDeviceFs failed:", err)
	}
	conn := fuse.NewFileSystemConnector(fs, fuse.NewFileSystemOptions())
	rawFs := fuse.NewLockingRawFileSystem(conn)
	mount := fuse.NewMountState(rawFs)
	if err := mount.Mount(tempdir, nil); err != nil {
		t.Fatalf("mount failed: %v", err)
	}

	mount.Debug = true
	dev.DebugPrint = true
	defer mount.Unmount()
	go mount.Loop()

	var root string
	for i := 0; i < 10; i++ {
		fis, err := ioutil.ReadDir(tempdir)
		if err == nil && len(fis) > 0 {
			root = filepath.Join(tempdir, fis[0].Name())
			break;
		}
		time.Sleep(1)
	}

	_, err = os.Lstat(root + "/Music")
	if err != nil {
		t.Fatal("Music not found", err)
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
	
	log.Println("writing...")
	golden += "hello"
	_, err = f.Write([]byte("hello"))
	log.Println("done...")
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
