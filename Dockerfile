############################
# STEP 1 build executable binary
############################
FROM golang:1.14-alpine AS builder

RUN apk update && apk add --no-cache build-base ca-certificates git

ARG TARGETPLATFORM
WORKDIR /go/src/github.com/wabarc/wayback

RUN git clone --progress https://github.com/wabarc/wayback.git . \
    && sh ./build/binary.sh $TARGETPLATFORM \
    && mv ./bin/wayback-* /wayback

############################
# STEP 2 build a small image
############################
FROM alpine:3.12

LABEL maintainer "WaybackBot <wabarc@tuta.io>"
COPY --from=builder /wayback /usr/local/bin
RUN apk update && apk add ca-certificates tor
RUN mv /etc/tor/torrc.sample /etc/tor/torrc
RUN echo 'ExcludeNodes {cn},{hk},{mo},{kp},{ir},{sy},{pk},{cu},{vn},{ru}' >> /etc/tor/torrc
RUN echo 'ExcludeExitNodes {cn},{hk},{mo},{sg},{th},{pk},{by},{ru},{ir},{vn},{ph},{my},{cu}' >> /etc/tor/torrc
RUN echo 'StrictNodes 1' >> /etc/tor/torrc

USER tor
WORKDIR /tmp

ENTRYPOINT ["/usr/local/bin/wayback"]

HEALTHCHECK --interval=5s --timeout=3s --retries=12 \
  CMD ps | awk '{print $4}' | grep wayback || exit 1
