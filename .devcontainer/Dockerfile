# Copyright 2022 Wayback Archiver. All rights reserved.
# Use of this source code is governed by the GNU GPL v3
# license that can be found in the LICENSE file.
#
# See here for image contents: https://github.com/microsoft/vscode-dev-containers/tree/v0.195.0/containers/go/.devcontainer/base.Dockerfile
# [Choice] Go version (use -bookworm variants on local arm64/Apple Silicon): 1, 1.16, 1.17, 1-bullseye, 1.16-bullseye, 1.17-bullseye, 1-buster, 1.16-buster, 1.17-buster
ARG VARIANT=1-bookworm
FROM mcr.microsoft.com/devcontainers/go:dev-${VARIANT}

# [Choice] Node.js version: lts/*, 16, 14, 12, 10
# ARG NODE_VERSION="lts/*"
# RUN if [ "${NODE_VERSION}" != "none" ]; then su vscode -c ". /usr/local/share/nvm/nvm.sh && nvm install ${NODE_VERSION} 2>&1"; fi

# [Optional] Uncomment this section to install additional OS packages.
RUN apt-get update && export DEBIAN_FRONTEND=noninteractive \
    && apt-get -y install --no-install-recommends \
        git \
        vim \
        tor \
        ffmpeg \
        chromium \
        youtube-dl \
        webp \
        fonts-freefont-ttf \
        fonts-font-awesome \
        fonts-noto \
        fonts-noto-core \
        fonts-noto-cjk \
        fonts-noto-extra

# [Optional] Uncomment the next line to use go get to install anything else you need
# RUN go get -x <your-dependency-or-tool>

# [Optional] Uncomment this line to install global node packages.
# RUN su vscode -c "source /usr/local/share/nvm/nvm.sh && npm install -g <your-package-here>" 2>&1
RUN su vscode -c "curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.60.3"

# Change directory permission
RUN chown -R vscode:golang /go/pkg
