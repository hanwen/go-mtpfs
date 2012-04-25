#!/bin/sh

set -eux

mount=$(mktemp -d)
./go-mtpfs $mount &
sleep 1
root="$mount/SD Card"

rm -rf "$root/mtpfs-test"
mkdir "$root/mtpfs-test"
rmdir "$root/mtpfs-test"
mkdir "$root/mtpfs-test"
echo -n hello > "$root/mtpfs-test/test.txt"
ls -l "$root/mtpfs-test/test.txt"
test $(cat "$root/mtpfs-test/test.txt") == "hello"
touch "$root/mtpfs-test/test.txt"
echo something else > "$root/mtpfs-test/test.txt"

fusermount -u $mount
