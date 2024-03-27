# Copyright 2020 Wayback Archiver. All rights reserved.
# Use of this source code is governed by the GNU GPL v3
# license that can be found in the LICENSE file.
#
FROM golang:1.20-alpine AS builder

ARG WAYBACK_IPFS_APIKEY

RUN apk update && apk add --no-cache build-base ca-certificates git
# Required by statically linked binary with OpenSSL
# RUN apk add linux-headers

ENV WAYBACK_IPFS_APIKEY ${WAYBACK_IPFS_APIKEY}

WORKDIR /go/src/app

COPY . .

RUN make linux-amd64

FROM fedora:39 AS runtime

WORKDIR /rpmbuild

RUN dnf install -y rpm-build rpm-sign systemd

COPY --from=builder /go/src/app/build/binary/wayback-linux-amd64 /rpmbuild/SOURCES/wayback
COPY --from=builder /go/src/app/LICENSE /rpmbuild/SOURCES/
COPY --from=builder /go/src/app/CHANGELOG.md /rpmbuild/SOURCES/
COPY --from=builder /go/src/app/wayback.1 /rpmbuild/SOURCES/
COPY --from=builder /go/src/app/build/systemd/wayback.service /rpmbuild/SOURCES/
COPY --from=builder /go/src/app/build/redhat/wayback.spec /rpmbuild/SPECS/wayback.spec

COPY build/redhat/entrypoint.sh /entrypoint.sh

ENTRYPOINT "/entrypoint.sh"
