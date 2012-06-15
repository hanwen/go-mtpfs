#!/bin/sh

set -eux
storage="Internal Storage"
mount=$(mktemp -d)
localfile=$(mktemp)
root="$mount/$storage"
delay="sleep 2"
./go-mtpfs -fs-debug $mount &
$delay

test -d "$root"
rm -rf "$root/mtpfs-test"
mkdir "$root/mtpfs-test"
mkdir "$root/mtpfs-test/create"
mkdir "$root/mtpfs-test/delete"
rmdir "$root/mtpfs-test/delete"
echo -n hello > "$root/mtpfs-test/test.txt"
ls -l "$root/mtpfs-test/test.txt"
test $(cat "$root/mtpfs-test/test.txt") == "hello"
touch "$root/mtpfs-test/test.txt"
echo something else > "$root/mtpfs-test/test.txt"

# can't send zero bytes file.
! touch "$root/mtpfs-test/zerobytes.txt"
rm "$root/mtpfs-test/zerobytes.txt"

mv "$root/mtpfs-test/test.txt" "$root/mtpfs-test/test2.txt"
! test -f  "$root/mtpfs-test/test.txt"
test -f  "$root/mtpfs-test/test2.txt"

echo hoi > "$root/mtpfs-test/dest.txt"
echo hoi > "$root/mtpfs-test/src.txt"
mv "$root/mtpfs-test/src.txt" "$root/mtpfs-test/dest.txt"
test -f  "$root/mtpfs-test/dest.txt"
! test -f  "$root/mtpfs-test/src.txt"

echo hoi > $localfile
cp $localfile "$root/mtpfs-test/double-copy.txt"
cp $localfile "$root/mtpfs-test/double-copy.txt"

fusermount -u $mount

./go-mtpfs $mount &
$delay

test -d  "$root/mtpfs-test/create"
! test -d  "$root/mtpfs-test/delete"
! test -f  "$root/mtpfs-test/test.txt"
test -f  "$root/mtpfs-test/test2.txt"
test -f  "$root/mtpfs-test/dest.txt"
! test -f  "$root/mtpfs-test/src.txt"
! test -f  "$root/mtpfs-test/zerobytes.txt"
rm "$root/mtpfs-test/double-copy.txt"

fusermount -u $mount

echo PASS
