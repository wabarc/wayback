# Copyright 2020 Wayback Archiver. All rights reserved.
# Use of this source code is governed by the GNU GPL v3
# license that can be found in the LICENSE file.
#
FROM golang@sha256:8e0cab3057642c2e4cbbbc6d02dd5dbcd23eec5fdc5f106bce3a2869abdd2a47 AS builder # golang:1.19-alpine

ARG WAYBACK_IPFS_APIKEY

RUN apk update && apk add --no-cache build-base ca-certificates git
# Required by statically linked binary with OpenSSL
# RUN apk add linux-headers

ENV WAYBACK_IPFS_APIKEY ${WAYBACK_IPFS_APIKEY}

WORKDIR /go/src/app
COPY . .
RUN make linux-amd64

FROM fedora@sha256:a166de31bbecce7d3ffe801ccacee0bf3aa8c7c03438c55445d1268c8e26909b # fedora:38
RUN dnf install -y rpm-build systemd
RUN mkdir -p /root/rpmbuild/{BUILD,RPMS,SOURCES,SPECS,SRPMS}
RUN echo "%_topdir /root/rpmbuild" >> .rpmmacros
COPY --from=builder /go/src/app/build/binary/wayback-linux-amd64 /root/rpmbuild/SOURCES/wayback
COPY --from=builder /go/src/app/LICENSE /root/rpmbuild/SOURCES/
COPY --from=builder /go/src/app/CHANGELOG.md /root/rpmbuild/SOURCES/
COPY --from=builder /go/src/app/wayback.1 /root/rpmbuild/SOURCES/
COPY --from=builder /go/src/app/build/systemd/wayback.service /root/rpmbuild/SOURCES/
COPY --from=builder /go/src/app/build/redhat/wayback.spec /root/rpmbuild/SPECS/wayback.spec
