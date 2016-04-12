###INTRODUCTION

Go-mtpfs is a simple FUSE filesystem for mounting Android devices as a
MTP device.

It will expose all storage areas of a device in the mount, and only
reads file metadata as needed, making it mount quickly. It uses
Android extensions to read/write partial data, so manipulating large
files requires no extra space in /tmp.

It has been tested on various flagship devices (Galaxy Nexus, Xoom,
Nexus 7).  As of Jan. 2013, it uses a pure Go implementation of MTP,
which is based on libusb.



###COMPILATION

* Install the Go compiler suite; e.g. on Ubuntu:
```
sudo apt-get install golang-go
```
* Install libmtp header files
```
sudo apt-get install libusb-1.0-0-dev
```
* Then run
```
mkdir /tmp/go
export GOPATH=/tmp/go
go get github.com/hanwen/go-mtpfs
```
  /tmp/go/bin/go-mtpfs will then contain the program binary.

* You may need some tweaking to get libusb to compile.  See the
  comment near the top of usb/usb.go, ie.
```
# edit to suit libusb installation:
vi /tmp/go/src/github.com/hanwen/go-mtpfs/usb/usb.go
go install github.com/hanwen/go-mtpfs
```
* A 32 and 64-bit linux x86 binaries are at

  http://hanwen.home.xs4all.nl/public/software/go-mtpfs/


###USAGE
```
mkdir xoom
go-mtpfs xoom &
cp -a ~/Music/Some-Album xoom/Music/
fusermount -u xoom
```
After a file is closed (eg. if "cp" completes), it is safe to unplug
the device; the filesystem then will continue to function, but
generates I/O errors when it reads from or writes to the device.


###CAVEATS

* It does not implement rename between directories, because the
  Android stack does not implement it.

* It does not implement Event handling, ie. it will not notice changes
  that the phone makes to the media database while connected.


###FEEDBACK

You can send your feedback through the issue tracker at
https://github.com/hanwen/go-mtpfs


###DISCLAIMER

This is not an official Google product.
