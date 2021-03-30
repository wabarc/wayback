FROM golang:1.16-alpine AS builder

RUN apk add --no-cache -U build-base ca-certificates git
# Required by statically linked binary with OpenSSL
#RUN apk add linux-headers

ARG TARGETPLATFORM
WORKDIR /go/src/github.com/wabarc/wayback

COPY . .
RUN sh ./build/binary.sh $TARGETPLATFORM \
    && mv ./build/binary/wayback-* /wayback

# Run application layer
FROM alpine:3.13

LABEL org.wabarc.homepage="http://github.com/wabarc" \
      org.wabarc.repository="http://github.com/wabarc/on-github"

ARG TOR_EXCLUDE_NODE="{cn},{hk},{mo},{sg},{th},{pk},{by},{ru},{ir},{sy},{vn},{ph},{my},{cu}"
ARG TOR_EXCLUDE_EXIT_NODE="{cn},{hk},{mo},{sg},{kp},{th},{pk},{by},{ru},{ir},{sy},{vn},{ph},{my},{cu},{au},{ca},{nz},{gb},{us},{fr},{dk},{nl},{no},{be},{de},{it},{es}"

ENV BASE_DIR /wayback
ENV PUSER wayback
ENV PGROUP wayback

WORKDIR $BASE_DIR

RUN set -o pipefail; \
    addgroup --system "${PGROUP}"; \
    adduser --system --no-create-home --disabled-password \
    --gecos '' --home "${BASE_DIR}" --ingroup "${PGROUP}" "${PUSER}"; \
    chown -R "${PUSER}:${PGROUP}" "${BASE_DIR}"; \
    chmod -R g+w "${BASE_DIR}"

COPY --from=builder /wayback /usr/local/bin
RUN set -o pipefail; \
    apk add --no-cache -U ca-certificates tor; \
    rm -rf /var/cache/apk/*; \
    \
    mv /etc/tor/torrc.sample /etc/tor/torrc; \
    echo "ExcludeNodes ${TOR_EXCLUDE_NODE}" >> /etc/tor/torrc; \
    echo "ExcludeExitNodes ${TOR_EXCLUDE_EXIT_NODE}" >> /etc/tor/torrc; \
    echo 'StrictNodes 1' >> /etc/tor/torrc; \
    #echo 'User tor' >> /etc/tor/torrc; \
    chmod a+w /var/log/tor

# Trigger on downstream build, only support for docker,
# add flag `--format=docker` if using podman.
# Ref: https://wiki.alpinelinux.org/wiki/Fonts
ONBUILD RUN set -o pipefail; \
    apk add --no-cache -U \
    chromium \
    dbus \
    dumb-init \
    freetype \
    libstdc++ \
    harfbuzz \
    nss \
    ttf-freefont \
    ttf-font-awesome \
    font-noto \
    font-noto-arabic \
    font-noto-emoji \
    font-noto-cjk \
    font-noto-extra \
    font-noto-lao \
    font-noto-myanmar \
    font-noto-thai \
    font-noto-tibetan \
    rm -rf /var/cache/apk/* /tmp/* /var/tmp/*

