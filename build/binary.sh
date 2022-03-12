#!/bin/sh
# Copyright 2020 Wayback Archiver. All rights reserved.
# Use of this source code is governed by the GNU GPL v3
# license that can be found in the LICENSE file.
#
#
# Perform package builder
#
set -eux

GOOS=linux
GOARCH=amd64

for arg in "$@"; do
case $arg in
    *arm/v6|*arm32v6)
        GOARCH=armv6
        ;;
    *armv7|*arm/v7|*arm32v7)
        GOARCH=armv7
        ;;
    *arm64|*arm64v8)
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

