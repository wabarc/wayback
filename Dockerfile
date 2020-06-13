############################
# STEP 1 build executable binary
############################
FROM golang:1.14-alpine AS builder
RUN apk update && apk add --no-cache build-base ca-certificates
COPY . /tmp/wayback
RUN cd /tmp/wayback && make build && mv ./bin/wayback /wayback

############################
# STEP 2 build a small image
############################
FROM alpine:3.12

LABEL maintainer "WaybackBot <wabarc@tutanota.com>"
COPY --from=builder /wayback /usr/local/bin
RUN apk update && apk add ca-certificates

USER nobody
WORKDIR /usr/local/bin

ENTRYPOINT ["/usr/local/bin/wayback"]

HEALTHCHECK --interval=5s --timeout=3s --retries=12 \
  CMD ps | awk '{print $4}' | grep wayback || exit 1
