# syntax=docker/dockerfile:1.2
ARG GO_VERSION=1.17

FROM --platform=$BUILDPLATFORM golang:${GO_VERSION}-alpine AS builder
COPY --from=tonistiigi/xx:golang / /

RUN apk add --no-cache -U build-base ca-certificates linux-headers musl-dev git tar

ARG TARGETPLATFORM
WORKDIR /src

COPY . .
RUN --mount=type=bind,target=/src,rw \
    --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    sh ./build/binary.sh $TARGETPLATFORM \
    && rm -rf .build/binary/wayback-* \
    && mv ./build/binary/wayback-* /wayback

# Application layer
FROM alpine:3.15

LABEL org.wabarc.homepage="http://github.com/wabarc" \
      org.wabarc.repository="http://github.com/wabarc/wayback" \
      org.opencontainers.image.title=wayback \
      org.opencontainers.image.description="A toolkit for snapshot webpage to Internet Archive, archive.today, IPFS and beyond" \
      org.opencontainers.image.url=https://wabarc.eu.org \
      org.opencontainers.image.licenses=GPLv3 \
      org.opencontainers.image.source="https://github.com/wabarc/wayback"

ARG TOR_EXCLUDE_NODE="{cn},{hk},{mo},{sg},{th},{pk},{by},{ru},{ir},{sy},{vn},{ph},{my},{cu}"
ARG TOR_EXCLUDE_EXIT_NODE="{cn},{hk},{mo},{sg},{kp},{th},{pk},{by},{ru},{ir},{sy},{vn},{ph},{my},{cu},{au},{ca},{nz},{gb},{us},{fr},{dk},{nl},{no},{be},{de},{it},{es}"

ENV BASE_DIR /wayback
ENV PUSER wayback
ENV PGROUP wayback
ENV LANG en_US.UTF-8
ENV LC_ALL en_US.UTF-8
ENV LANGUAGE en_US.UTF-8

WORKDIR $BASE_DIR

RUN set -o pipefail; \
    addgroup --system "${PGROUP}"; \
    adduser --system --no-create-home --disabled-password \
    --gecos '' --home "${BASE_DIR}" --ingroup "${PGROUP}" "${PUSER}"; \
    chown -R "${PUSER}:${PGROUP}" "${BASE_DIR}"; \
    chmod -R g+w "${BASE_DIR}"

COPY --from=builder /wayback /usr/local/bin
RUN set -o pipefail; \
    apk add --no-cache -U ca-certificates libressl wget tor; \
    rm -rf /var/cache/apk/*; \
    \
    cp /etc/tor/torrc.sample /etc/tor/torrc; \
    #echo "ExcludeNodes ${TOR_EXCLUDE_NODE}" >> /etc/tor/torrc; \
    #echo "ExcludeExitNodes ${TOR_EXCLUDE_EXIT_NODE}" >> /etc/tor/torrc; \
    #echo 'StrictNodes 1' >> /etc/tor/torrc; \
    echo 'SocksPort 9050' >> /etc/tor/torrc; \
    echo 'ExitRelay 0' >> /etc/tor/torrc; \
    echo 'LongLivedPorts 8964' >> /etc/tor/torrc; \
    echo 'User tor' >> /etc/tor/torrc; \
    chown tor:nogroup /var/lib/tor/ /var/log/tor/; \
    chmod 700 /var/lib/tor

EXPOSE 8964

# Trigger on downstream build, only support for docker,
# add flag `--format=docker` if using podman.
# Ref: https://wiki.alpinelinux.org/wiki/Fonts
ONBUILD RUN set -o pipefail; \
    echo @v3.15 https://dl-cdn.alpinelinux.org/alpine/v3.15/community >> /etc/apk/repositories; \
    echo @v3.15 https://dl-cdn.alpinelinux.org/alpine/v3.15/main >> /etc/apk/repositories; \
    apk add --no-cache -U \
    chromium@v3.15 \
    dbus \
    dumb-init \
    ffmpeg@v3.15 \
    freetype@v3.15 \
    libstdc++ \
    harfbuzz@v3.15 \
    nss@v3.15 \
    you-get \
    rtmpdump \
    youtube-dl \
    libwebp-tools \
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
    font-noto-tibetan; \
    rm -rf /var/cache/apk/* /tmp/* /var/tmp/*

