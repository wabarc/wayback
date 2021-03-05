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
FROM alpine:3.13

LABEL maintainer "WaybackBot <wabarc@tuta.io>"

ARG TOR_EXCLUDE_NODE="{cn},{hk},{mo},{sg},{th},{pk},{by},{ru},{ir},{sy},{vn},{ph},{my},{cu}"
ARG TOR_EXCLUDE_EXIT_NODE="{cn},{hk},{mo},{sg},{kp},{th},{pk},{by},{ru},{ir},{sy},{vn},{ph},{my},{cu},{au},{ca},{nz},{gb},{us},{fr},{dk},{nl},{no},{be},{de},{it},{es}"

ENV BASE_DIR /wayback
ENV PUSER wayback
ENV PGROUP wayback

WORKDIR $BASE_DIR

RUN set -eux; \
    addgroup --system ${PGROUP}; \
    adduser --system --no-create-home --disabled-password \
    --gecos '' --home ${BASE_DIR} --ingroup ${PGROUP} ${PUSER}; \
    chown -R ${PUSER}:${PGROUP} "${BASE_DIR}"; \
    chmod -R g+w "${BASE_DIR}"

COPY --from=builder /wayback /usr/local/bin
RUN apk update; \
    apk add --no-cache ca-certificates tor; \
    rm -rf /var/cache/apk/*; \
    \
    mv /etc/tor/torrc.sample /etc/tor/torrc; \
    echo "ExcludeNodes ${TOR_EXCLUDE_NODE}" >> /etc/tor/torrc; \
    echo "ExcludeExitNodes ${TOR_EXCLUDE_EXIT_NODE}" >> /etc/tor/torrc; \
    echo 'StrictNodes 1' >> /etc/tor/torrc; \
    #echo 'User tor' >> /etc/tor/torrc
