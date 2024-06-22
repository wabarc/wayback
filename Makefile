export GO111MODULE = on
export CGO_ENABLED = 0
export GOPROXY = https://proxy.golang.org

NAME = wayback
REPO = github.com/wabarc/wayback
BINDIR ?= ./build/binary
PACKDIR ?= ./build/package
LDFLAGS := $(shell echo "-X '${REPO}/version.Version=`git describe --tags --abbrev=0`'")
LDFLAGS := $(shell echo "${LDFLAGS} -X '${REPO}/version.Commit=`git rev-parse --short HEAD`'")
LDFLAGS := $(shell echo "${LDFLAGS} -X '${REPO}/version.BuildDate=`date +%FT%T%z`'")
LDFLAGS := $(shell echo "${LDFLAGS} -X '${REPO}/config.IPFSToken=$(shell echo ${WAYBACK_IPFS_APIKEY})'")
GOBUILD ?= go build -trimpath --ldflags "-s -w ${LDFLAGS} -buildid=" -v
VERSION ?= $(shell git describe --tags `git rev-list --tags --max-count=1` | sed -e 's/v//g')
GOFILES ?= $(wildcard ./cmd/wayback/*.go)
PROJECT := github.com/wabarc/wayback
PACKAGES ?= $(shell go list ./...)
DOCKER ?= $(shell which docker || which podman)
DOCKER_IMAGE := wabarc/wayback
DEB_IMG_ARCH := amd64

.DEFAULT_GOAL := help

.PHONY: help
help: ## show help message
	@awk 'BEGIN {FS = ":.*##"; printf "Usage:\n  make <target>\n\nTargets: \033[36m\033[0m\n"} /^[$$()% 0-9a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

PLATFORM_LIST = \
	darwin-amd64 \
	darwin-arm64 \
	linux-386 \
	linux-amd64 \
	linux-armv5 \
	linux-armv6 \
	linux-armv7 \
	linux-arm64 \
	linux-mips-softfloat \
	linux-mips-hardfloat \
	linux-mipsle-softfloat \
	linux-mipsle-hardfloat \
	linux-mips64 \
	linux-mips64le \
	linux-ppc64 \
	linux-ppc64le \
	linux-s390x \
	freebsd-386 \
	freebsd-amd64 \
	freebsd-arm64 \
	openbsd-386 \
	openbsd-amd64 \
	dragonfly-amd64 \
	android-arm64

WINDOWS_ARCH_LIST = \
	windows-386 \
	windows-amd64 \
	windows-arm \
	windows-arm64

.PHONY: \
	all-arch \
	tar_releases \
	zip_releases \
	releases \
	clean \
	test \
	fmt \
	rpm \
	debian \
	debian-packages \
	docker-image

.SECONDEXPANSION:
%: ## Build binary, format: linux-amd64, darwin-arm64, full list: https://golang.org/doc/install/source#environment
	$(eval OS := $(shell echo $@ | cut -d'-' -f1))
	$(eval ARM := $(shell echo $@ | cut -d'-' -f2 | grep arm | sed -e 's/arm64//' | tr -dc '[0-9]'))
	$(eval ARCH := $(shell echo $@ | cut -d'-' -f2 | sed -e 's/armv.*/arm/' | grep -v $(OS)))
	$(eval MIPS := $(shell echo $@ | cut -d'-' -f3))
	$(if $(strip $(OS)),,$(error missing OS))
	$(if $(strip $(ARCH)),,$(error missing ARCH))
	GOOS="$(OS)" GOARCH="$(ARCH)" GOMIPS="$(MIPS)" GOARM="$(ARM)" $(GOBUILD) -o $(BINDIR)/$(NAME)-$@ $(GOFILES)

.PHONY: build
build: ## Build binary for current OS
	$(GOBUILD) -o $(BINDIR)/$(NAME) $(GOFILES)

.PHONY: linux-armv8
linux-armv8: linux-arm64

ifeq ($(TARGET),)
tar_releases := $(addsuffix .gz, $(PLATFORM_LIST))
zip_releases := $(addsuffix .zip, $(WINDOWS_ARCH_LIST))
else
ifeq ($(findstring windows,$(TARGET)),windows)
zip_releases := $(addsuffix .zip, $(TARGET))
else
tar_releases := $(addsuffix .gz, $(TARGET))
endif
endif

$(tar_releases): %.gz : %
	chmod +x $(BINDIR)/$(NAME)-$(basename $@)
	tar -czf $(PACKDIR)/$(NAME)-$(basename $@)-$(VERSION).tar.gz --transform "s/.*\///g" $(BINDIR)/$(NAME)-$(basename $@) LICENSE CHANGELOG.md README.md

$(zip_releases): %.zip : %
	@mv $(BINDIR)/$(NAME)-$(basename $@) $(BINDIR)/$(NAME)-$(basename $@).exe
	zip -m -j $(PACKDIR)/$(NAME)-$(basename $@)-$(VERSION).zip $(BINDIR)/$(NAME)-$(basename $@).exe LICENSE CHANGELOG.md README.md

all-arch: $(PLATFORM_LIST) $(WINDOWS_ARCH_LIST) ## Build binary for all architecture

releases: $(tar_releases) $(zip_releases) ## Packaging all binaries

clean: ## Clean workspace
	rm -f $(BINDIR)/*
	rm -f $(PACKDIR)/*
	rm -rf data-dir*
	rm -rf coverage*
	rm -rf *.out
	rm -rf wayback.db

fmt: ## Format codebase
	@echo "-> Running go fmt"
	@go fmt $(PACKAGES)

vet: ## Vet codebase
	@echo "-> Running go vet"
	@go vet $(PACKAGES)
	@go vet -tags with_lux $(PACKAGES)

test: ## Run testing
	@echo "-> Running go test"
	@go clean -testcache
	@CGO_ENABLED=1 go test -v -race -cover -coverprofile=coverage.out -covermode=atomic -parallel=1 ./...

test-integration: ## Run integration testing
	@echo 'mode: atomic' > coverage.out
	@CGO_ENABLED=1 go test -race -tags=integration -covermode=atomic -parallel=1 -coverprofile=coverage.out ./...

test-cover: ## Collect code coverage
	@echo "-> Running go tool cover"
	@go tool cover -func=coverage.out
	@go tool cover -html=coverage.out -o coverage.html

bench: ## Benchmark test
	@echo "-> Running benchmark"
	@go test -v -bench ./...

profile: ## Test and profile
	@echo "-> Running profile"
	@go test -cpuprofile cpu.prof -memprofile mem.prof -v -bench ./...

docker-image: ## Build Docker image
	@echo "-> Building docker image..."
	@$(DOCKER) build \
		--build-arg WAYBACK_IPFS_APIKEY=$(shell echo ${WAYBACK_IPFS_APIKEY}) \
		-t $(DOCKER_IMAGE):$(VERSION) \
		-f ./build/docker/Dockerfile.dev .

rpm: ## Build RPM package
	@echo "-> Building rpm package..."
	@$(DOCKER) build \
		--build-arg WAYBACK_IPFS_APIKEY=$(shell echo ${WAYBACK_IPFS_APIKEY}) \
		-t wayback-rpm-builder \
		-f build/redhat/Dockerfile .
	@$(DOCKER) run --rm \
		-e WAYBACK_SIGNING_KEY="$${WAYBACK_SIGNING_KEY}" \
		-e WAYBACK_SIGNING_PASSPHARSE="$${WAYBACK_SIGNING_PASSPHARSE}" \
		-e VERSION="${VERSION}" \
		-v ${PWD}/build/package:/rpmbuild/RPMS/x86_64:Z \
		wayback-rpm-builder

debian: ## Build Debian packages
	@echo "-> Building deb package..."
	@$(DOCKER) buildx build --load \
		--platform linux/$(DOCKER_PLATFORM) \
		--build-arg PKG_VERSION=$(VERSION) \
		--build-arg WAYBACK_IPFS_APIKEY=$(shell echo ${WAYBACK_IPFS_APIKEY}) \
		-t wayback-deb-builder \
		-f build/debian/Dockerfile \
		.
	@$(DOCKER) run --rm --platform linux/$(DOCKER_PLATFORM) \
		-v ${PWD}/build/package:/pkg wayback-deb-builder
	@echo "-> DEB package below:"
	@ls -h ${PWD}/build/package/*.deb

debian-packages: ## Build Debian packages, including amd64, arm32v7, arm64v8
	$(MAKE) debian DOCKER_PLATFORM=amd64
	$(MAKE) debian DOCKER_PLATFORM=arm64
	$(MAKE) debian DOCKER_PLATFORM=arm/v7

submodule: ## Update Git submodule
	@echo "-> Updating Git submodule..."
	@git submodule update --init --recursive --remote

bina: ## Update bina.json
	@echo "-> Updating bina.json"
	$(eval LATEST_TAG := $(shell git describe --tags --abbrev=0 | sed 's/v//'))
	sed "s#0.0.0#${LATEST_TAG}#g" bina.tpl.json > bina.json

scan: ## Scan vulnerabilities
	@echo "-> Scanning vulnerabilities..."
	@go list -json -m all | $(DOCKER) run --rm -i sonatypecommunity/nancy sleuth --skip-update-check
