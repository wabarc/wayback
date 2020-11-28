############################
# STEP 1 build executable binary
############################
FROM golang:1.15-alpine AS builder

RUN apk update && apk add --no-cache build-base ca-certificates git
# Required by statically linked binary with OpenSSL
RUN apk add linux-headers

ARG TARGETPLATFORM
WORKDIR /go/src/github.com/wabarc/wayback

COPY . .
RUN sh ./build/binary.sh $TARGETPLATFORM \
    && mv ./build/binary/wayback-* /wayback

############################
# STEP 2 build a small image
############################
FROM alpine:3.12

LABEL maintainer "WaybackBot <wabarc@tuta.io>"
COPY --from=builder /wayback /usr/local/bin
RUN apk update && apk add ca-certificates tor
RUN mv /etc/tor/torrc.sample /etc/tor/torrc
RUN echo 'ExcludeNodes {cn},{hk},{mo},{sg},{th},{pk},{by},{ru},{ir},{sy},{vn},{ph},{my},{cu},{us},{au},{ca},{nz},{gb},{fr}' >> /etc/tor/torrc
RUN echo 'ExcludeExitNodes {cn},{hk},{mo},{sg},{th},{pk},{by},{ru},{ir},{sy},{vn},{ph},{my},{cu},{us},{au},{ca},{nz},{gb},{fr}' >> /etc/tor/torrc
RUN echo 'StrictNodes 1' >> /etc/tor/torrc

RUN rm -rf /var/cache/apk/*

#USER tor
#WORKDIR /tmp

#ENTRYPOINT ["/usr/local/bin/wayback"]

#HEALTHCHECK --interval=5s --timeout=3s --retries=12 \
#  CMD ps | awk '{print $4}' | grep wayback || exit 1
