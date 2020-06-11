### INTRODUCTION

Go-mtpfs is a simple FUSE filesystem for mounting Android devices as a
MTP device.

It will expose all storage areas of a device in the mount, and only
reads file metadata as needed, making it mount quickly. It uses
Android extensions to read/write partial data, so manipulating large
files requires no extra space in /tmp.

It has been tested on various flagship devices (Galaxy Nexus, Xoom,
Nexus 7).  As of Jan. 2013, it uses a pure Go implementation of MTP,
which is based on libusb.



### COMPILATION

* Install the Go compiler suite; e.g. on Ubuntu:
```
sudo apt-get install golang-go
```

You need a sufficiently recent `go`, 1.13.8 is reported to be working.

* Install libmtp header files
```
sudo apt-get install libusb1-devel
```

* Then check out go-mtpfs
```
git checkout 'https://github.com/hanwen/go-mtpfs'
```

* Make sure that you have an internet connection that can access Google's servers

  The build process will download additional material from Google, so make sure you are 
  connected to the Internet when building.

* Build `go-mtpfs`

  Run
```
go build ./
```
  This will leave a binary `go-mtpfs`

* You may need some tweaking to get libusb to compile.  See the
  comment near the top of https://github.com/hanwen/usb/usb.go.

* 32-bit and 64-bit linux x86 binaries are at

  https://hanwen.home.xs4all.nl/public/software/go-mtpfs/


### USAGE
```
mkdir xoom
go-mtpfs xoom &
cp -a ~/Music/Some-Album xoom/Music/
fusermount -u xoom
```
After a file is closed (eg. if "cp" completes), it is safe to unplug
the device; the filesystem then will continue to function, but
generates I/O errors when it reads from or writes to the device.


### CAVEATS

* It does not implement rename between directories, because the
  Android stack does not implement it.

* It does not implement Event handling, ie. it will not notice changes
  that the phone makes to the media database while connected.

* Some Sony Xperia devices claim to implement Android extension, but
  don't. See [issue
  #104](https://github.com/hanwen/go-mtpfs/issues/104). Symptom:

     AndroidGetPartialObject64 failed: OperationNotSupported

  In this case, disable Android extensions with the flag -android=0


### FEEDBACK

You can send your feedback through the issue tracker at
https://github.com/hanwen/go-mtpfs


### DISCLAIMER

This is not an official Google product.
