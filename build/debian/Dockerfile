ARG IMAGE_ARCH="amd64"
ARG PKG_VERSION="0.0.1"

FROM ${IMAGE_ARCH}/golang:buster AS build

ARG PKG_VERSION
ENV DEBIAN_FRONTEND noninteractive

RUN apt update -q && \
    apt install -y -qq build-essential devscripts dh-make dh-systemd

RUN mkdir -p /build/debian /pkg
COPY . /src
WORKDIR /src/build/debian
ENV PKG_VERSION ${PKG_VERSION}

CMD ["/src/build/debian/builder"]
