FROM golang:1.16-alpine AS builder

RUN apk update && apk add --no-cache build-base ca-certificates git
# Required by statically linked binary with OpenSSL
RUN apk add linux-headers

ADD . /go/src/app
WORKDIR /go/src/app
RUN make linux-amd64

FROM centos:latest
RUN dnf install -y rpm-build
RUN mkdir -p /root/rpmbuild/{BUILD,RPMS,SOURCES,SPECS,SRPMS}
RUN echo "%_topdir /root/rpmbuild" >> .rpmmacros
COPY --from=builder /go/src/app/build/binary/wayback-linux-amd64 /root/rpmbuild/SOURCES/wayback
COPY --from=builder /go/src/app/LICENSE /root/rpmbuild/SOURCES/
COPY --from=builder /go/src/app/CHANGELOG.md /root/rpmbuild/SOURCES/
COPY --from=builder /go/src/app/wayback.1 /root/rpmbuild/SOURCES/
COPY --from=builder /go/src/app/build/systemd/wayback.service /root/rpmbuild/SOURCES/
COPY --from=builder /go/src/app/build/redhat/wayback.spec /root/rpmbuild/SPECS/wayback.spec
