#!/bin/sh

# Script to exercise everything.
set -eux

for x in fs mtp
do
    go build github.com/puhitaku/mtplvcap/$x
    go test -i github.com/puhitaku/mtplvcap/$x
    go test github.com/puhitaku/mtplvcap/$x
done

go build github.com/puhitaku/mtplvcap
