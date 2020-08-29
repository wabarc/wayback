#!/bin/sh
#
# Perform package builder
set -eu
#set -x

GOOS=linux
GOARCH=amd64

for arg in "$@"; do
case $arg in
    *arm/v7)
        GOARCH=armv7
        ;;
    *arm64)
        GOARCH=armv8
        ;;
    *386)
        GOARCH=386
        ;;
    *ppc64le)
        GOARCH=ppc64le
        ;;
    *s390x)
        GOARCH=s390x
        ;;
    windows*)
        GOOS=windows
        ;;
    darwin*)
        GOOS=darwin
        ;;
esac
done

TARGET="${GOOS}-${GOARCH}"

make $TARGET

