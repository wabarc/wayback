ARG IMAGE_ARCH="amd64"
ARG PKG_VERSION="0.0.1"
ARG PKG_ARCH="amd64"

FROM ${IMAGE_ARCH}/golang:buster AS build

ARG PKG_ARCH
ARG PKG_VERSION
ENV DEBIAN_FRONTEND noninteractive

RUN apt-get update -q && \
    apt-get install -y -qq --no-install-recommends build-essential devscripts dh-make dh-systemd && \
    rm -rf /var/lib/apt/lists/*

RUN mkdir -p /build/debian /pkg
COPY . /src
WORKDIR /src/build/debian
ENV PKG_VERSION ${PKG_VERSION}
ENV PKG_ARCH ${PKG_ARCH}

CMD ["/src/build/debian/builder"]
