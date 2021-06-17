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
GOBUILD ?= go build -trimpath --ldflags "-s -w ${LDFLAGS} -buildid=" -v
VERSION ?= $(shell git describe --tags `git rev-list --tags --max-count=1` | sed -e 's/v//g')
GOFILES ?= $(wildcard ./cmd/wayback/*.go)
PROJECT := github.com/wabarc/wayback
PACKAGES ?= $(shell go list ./...)
DOCKER ?= $(shell which docker || which podman)
DOCKER_IMAGE := wabarc/wayback
DEB_IMG_ARCH := amd64

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
	windows-arm

.PHONY: \
	darwin-386 \
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
	windows-386 \
	windows-amd64 \
	windows-arm \
	android-arm64 \
	js-wasm \
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

.PHONY: build
build:
	$(GOBUILD) -o $(BINDIR)/$(NAME) $(GOFILES)

darwin-386:
	GOOS=darwin GOARCH=386 $(GOBUILD) -o $(BINDIR)/$(NAME)-$@ $(GOFILES)

darwin-amd64:
	GOOS=darwin GOARCH=amd64 $(GOBUILD) -o $(BINDIR)/$(NAME)-$@ $(GOFILES)

darwin-arm64:
	GOOS=darwin GOARCH=arm64 $(GOBUILD) -o $(BINDIR)/$(NAME)-$@ $(GOFILES)

linux-386:
	GOOS=linux GOARCH=386 $(GOBUILD) -o $(BINDIR)/$(NAME)-$@ $(GOFILES)

linux-amd64:
	GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BINDIR)/$(NAME)-$@ $(GOFILES)

linux-armv5:
	GOOS=linux GOARCH=arm GOARM=5 $(GOBUILD) -o $(BINDIR)/$(NAME)-$@ $(GOFILES)

linux-armv6:
	GOOS=linux GOARCH=arm GOARM=6 $(GOBUILD) -o $(BINDIR)/$(NAME)-$@ $(GOFILES)

linux-armv7:
	GOOS=linux GOARCH=arm GOARM=7 $(GOBUILD) -o $(BINDIR)/$(NAME)-$@ $(GOFILES)

linux-armv8: linux-arm64
linux-arm64:
	GOOS=linux GOARCH=arm64 $(GOBUILD) -o $(BINDIR)/$(NAME)-$@ $(GOFILES)

linux-mips-softfloat:
	GOOS=linux GOARCH=mips GOMIPS=softfloat $(GOBUILD) -o $(BINDIR)/$(NAME)-$@ $(GOFILES)

linux-mips-hardfloat:
	GOOS=linux GOARCH=mips GOMIPS=hardfloat $(GOBUILD) -o $(BINDIR)/$(NAME)-$@ $(GOFILES)

linux-mipsle-softfloat:
	GOOS=linux GOARCH=mipsle GOMIPS=softfloat $(GOBUILD) -o $(BINDIR)/$(NAME)-$@ $(GOFILES)

linux-mipsle-hardfloat:
	GOOS=linux GOARCH=mipsle GOMIPS=hardfloat $(GOBUILD) -o $(BINDIR)/$(NAME)-$@ $(GOFILES)

linux-mips64:
	GOOS=linux GOARCH=mips64 $(GOBUILD) -o $(BINDIR)/$(NAME)-$@ $(GOFILES)

linux-mips64le:
	GOOS=linux GOARCH=mips64le $(GOBUILD) -o $(BINDIR)/$(NAME)-$@ $(GOFILES)

linux-ppc64:
	GOOS=linux GOARCH=ppc64 $(GOBUILD) -o $(BINDIR)/$(NAME)-$@ $(GOFILES)

linux-ppc64le:
	GOOS=linux GOARCH=ppc64le $(GOBUILD) -o $(BINDIR)/$(NAME)-$@ $(GOFILES)

linux-s390x:
	GOOS=linux GOARCH=s390x $(GOBUILD) -o $(BINDIR)/$(NAME)-$@ $(GOFILES)

freebsd-386:
	GOOS=freebsd GOARCH=386 $(GOBUILD) -o $(BINDIR)/$(NAME)-$@ $(GOFILES)

freebsd-amd64:
	GOOS=freebsd GOARCH=amd64 $(GOBUILD) -o $(BINDIR)/$(NAME)-$@ $(GOFILES)

freebsd-arm64:
	GOOS=freebsd GOARCH=arm64 $(GOBUILD) -o $(BINDIR)/$(NAME)-$@ $(GOFILES)

openbsd-386:
	GOOS=openbsd GOARCH=386 $(GOBUILD) -o $(BINDIR)/$(NAME)-$@ $(GOFILES)

openbsd-amd64:
	GOOS=openbsd GOARCH=amd64 $(GOBUILD) -o $(BINDIR)/$(NAME)-$@ $(GOFILES)

windows-386:
	GOOS=windows GOARCH=386 $(GOBUILD) -o $(BINDIR)/$(NAME)-$@.exe $(GOFILES)

windows-amd64:
	GOOS=windows GOARCH=amd64 $(GOBUILD) -o $(BINDIR)/$(NAME)-$@.exe $(GOFILES)

windows-arm:
	GOOS=windows GOARCH=arm $(GOBUILD) -o $(BINDIR)/$(NAME)-$@.exe $(GOFILES)

dragonfly-amd64:
	GOOS=dragonfly GOARCH=amd64 $(GOBUILD) -o $(BINDIR)/$(NAME)-$@ $(GOFILES)

android-arm64:
	GOOS=android GOARCH=arm64 $(GOBUILD) -o $(BINDIR)/$(NAME)-$@ $(GOFILES)

js-wasm:
	GOOS=js GOARCH=wasm $(GOBUILD) -o $(BINDIR)/$(NAME)-$@ $(GOFILES)

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
	zip -m -j $(PACKDIR)/$(NAME)-$(basename $@)-$(VERSION).zip $(BINDIR)/$(NAME)-$(basename $@).exe LICENSE CHANGELOG.md README.md

all-arch: $(PLATFORM_LIST) $(WINDOWS_ARCH_LIST)

releases: $(tar_releases) $(zip_releases)

clean:
	rm -f $(BINDIR)/*
	rm -f $(PACKDIR)/*
	rm -rf data-dir*
	rm -rf coverage*
	rm -rf *.out
	rm -rf wayback.db

fmt:
	@echo "-> Running go fmt"
	@go fmt $(PACKAGES)

vet:
	@echo "-> Running go vet"
	@go vet $(PACKAGES)

test:
	@echo "-> Running go test"
	@go clean -testcache
	@CGO_ENABLED=1 go test -v -race -cover -coverprofile=coverage.out -covermode=atomic ./...

test-integration:
	@echo 'mode: atomic' > coverage.out
	@go list ./... | xargs -n1 -I{} sh -c 'CGO_ENABLED=1 go test -race -tags=integration -covermode=atomic -coverprofile=coverage.tmp -coverpkg $(go list ./... | tr "\n" ",") {} && tail -n +2 coverage.tmp >> coverage.out || exit 255'
	@rm coverage.tmp

test-cover:
	@echo "-> Running go tool cover"
	@go tool cover -func=coverage.out
	@go tool cover -html=coverage.out -o coverage.html

bench:
	@echo "-> Running benchmark"
	@go test -v -bench .

profile:
	@echo "-> Running profile"
	@go test -cpuprofile cpu.prof -memprofile mem.prof -v -bench .

docker-image:
	@echo "-> Building docker image..."
	@$(DOCKER) build -t $(DOCKER_IMAGE):$(VERSION) -f ./Dockerfile .

rpm:
	@echo "-> Building rpm package..."
	@$(DOCKER) build \
		-t wayback-rpm-builder \
		-f build/redhat/Dockerfile .
	@$(DOCKER) run --rm \
		-v ${PWD}/build/package:/root/rpmbuild/RPMS/x86_64 wayback-rpm-builder \
		rpmbuild -bb --define "_wayback_version $(VERSION)" /root/rpmbuild/SPECS/wayback.spec

debian:
	@echo "-> Building deb package..."
	@$(DOCKER) build \
		--build-arg IMAGE_ARCH=$(DEB_IMG_ARCH) \
		--build-arg PKG_VERSION=$(VERSION) \
		--build-arg PKG_ARCH=$(PKG_ARCH) \
		-t $(DEB_IMG_ARCH)/wayback-deb-builder \
		-f build/debian/Dockerfile .
	@$(DOCKER) run --rm \
		-v ${PWD}/build/package:/pkg \
		$(DEB_IMG_ARCH)/wayback-deb-builder
	@echo "-> DEB package below:"
	@ls -h ${PWD}/build/package/*.deb

debian-packages:
	$(MAKE) debian DEB_IMG_ARCH=amd64
	$(MAKE) debian DEB_IMG_ARCH=arm32v7 PKG_ARCH=armv7
	$(MAKE) debian DEB_IMG_ARCH=arm64v8 PKG_ARCH=arm64

submodule:
	@echo "-> Updating Git submodule..."
	@git submodule update --init --recursive --remote
