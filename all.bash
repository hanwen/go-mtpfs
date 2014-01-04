#!/bin/sh

# Script to exercise everything.
set -eux

for x in fs mtp
do
    go build github.com/hanwen/go-mtpfs/$x
    go test -i github.com/hanwen/go-mtpfs/$x
    go test github.com/hanwen/go-mtpfs/$x
done

go build github.com/hanwen/go-mtpfs
