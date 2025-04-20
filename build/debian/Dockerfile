FROM docker.io/golang:1.24-bookworm AS build

ARG WAYBACK_IPFS_APIKEY

ENV DEBIAN_FRONTEND noninteractive

RUN apt-get update -q && \
    apt-get install -y -qq build-essential devscripts dh-make debhelper && \
    mkdir -p /build/debian /pkg && \
    rm -rf /var/lib/apt/lists/*

COPY . /src

ENV WAYBACK_IPFS_TARGET ${WAYBACK_IPFS_TARGET}
ENV WAYBACK_IPFS_APIKEY ${WAYBACK_IPFS_APIKEY}
ENV WAYBACK_IPFS_SECRET ${WAYBACK_IPFS_SECRET}

CMD ["/src/build/debian/builder"]
