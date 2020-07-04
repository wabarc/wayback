export GO111MODULE = on
export GOPROXY = https://proxy.golang.org

NAME = wayback
BINDIR ?= ./bin
PACKDIR ?= ./build/package
VERSION := $(shell cat VERSION)
GOBUILD := CGO_ENABLED=0 go build --ldflags="-s -w"
GOFILES := $(wildcard ./cmd/*.go)
PROJECT := github.com/wabarc/wayback
PACKAGES := $(shell go list ./...)

PLATFORM_LIST = \
	darwin-amd64 \
	linux-386 \
	linux-amd64 \
	linux-armv5 \
	linux-armv6 \
	linux-armv7 \
	linux-armv8 \
	linux-mips-softfloat \
	linux-mips-hardfloat \
	linux-mipsle-softfloat \
	linux-mipsle-hardfloat \
	linux-mips64 \
	linux-mips64le \
	freebsd-386 \
	freebsd-amd64

WINDOWS_ARCH_LIST = \
	windows-386 \
	windows-amd64

.PHONY: all
all: linux-amd64 darwin-amd64 windows-amd64

darwin-amd64:
	GOARCH=amd64 GOOS=darwin $(GOBUILD) -o $(BINDIR)/$(NAME)-$@ $(GOFILES)

linux-386:
	GOARCH=386 GOOS=linux $(GOBUILD) -o $(BINDIR)/$(NAME)-$@ $(GOFILES)

linux-amd64:
	GOARCH=amd64 GOOS=linux $(GOBUILD) -o $(BINDIR)/$(NAME)-$@ $(GOFILES)

linux-armv5:
	GOARCH=arm GOOS=linux GOARM=5 $(GOBUILD) -o $(BINDIR)/$(NAME)-$@ $(GOFILES)

linux-armv6:
	GOARCH=arm GOOS=linux GOARM=6 $(GOBUILD) -o $(BINDIR)/$(NAME)-$@ $(GOFILES)

linux-armv7:
	GOARCH=arm GOOS=linux GOARM=7 $(GOBUILD) -o $(BINDIR)/$(NAME)-$@ $(GOFILES)

linux-armv8:
	GOARCH=arm64 GOOS=linux $(GOBUILD) -o $(BINDIR)/$(NAME)-$@ $(GOFILES)

linux-mips-softfloat:
	GOARCH=mips GOMIPS=softfloat GOOS=linux $(GOBUILD) -o $(BINDIR)/$(NAME)-$@ $(GOFILES)

linux-mips-hardfloat:
	GOARCH=mips GOMIPS=hardfloat GOOS=linux $(GOBUILD) -o $(BINDIR)/$(NAME)-$@ $(GOFILES)

linux-mipsle-softfloat:
	GOARCH=mipsle GOMIPS=softfloat GOOS=linux $(GOBUILD) -o $(BINDIR)/$(NAME)-$@ $(GOFILES)

linux-mipsle-hardfloat:
	GOARCH=mipsle GOMIPS=hardfloat GOOS=linux $(GOBUILD) -o $(BINDIR)/$(NAME)-$@ $(GOFILES)

linux-mips64:
	GOARCH=mips64 GOOS=linux $(GOBUILD) -o $(BINDIR)/$(NAME)-$@ $(GOFILES)

linux-mips64le:
	GOARCH=mips64le GOOS=linux $(GOBUILD) -o $(BINDIR)/$(NAME)-$@ $(GOFILES)

freebsd-386:
	GOARCH=386 GOOS=freebsd $(GOBUILD) -o $(BINDIR)/$(NAME)-$@ $(GOFILES)

freebsd-amd64:
	GOARCH=amd64 GOOS=freebsd $(GOBUILD) -o $(BINDIR)/$(NAME)-$@ $(GOFILES)

windows-386:
	GOARCH=386 GOOS=windows $(GOBUILD) -o $(BINDIR)/$(NAME)-$@.exe $(GOFILES)

windows-amd64:
	GOARCH=amd64 GOOS=windows $(GOBUILD) -o $(BINDIR)/$(NAME)-$@.exe $(GOFILES)

fmt:
	@echo "-> Running go fmt"
	@go fmt $(PACKAGES)

tar_releases := $(addsuffix .gz, $(PLATFORM_LIST))
zip_releases := $(addsuffix .zip, $(WINDOWS_ARCH_LIST))

$(tar_releases): %.gz : %
	chmod +x $(BINDIR)/$(NAME)-$(basename $@)
	tar -czf $(PACKDIR)/$(NAME)-$(basename $@)-$(VERSION).tar.gz --transform "s,$(BINDIR)/,$(NAME)/," $(BINDIR)/$(NAME)-$(basename $@)

$(zip_releases): %.zip : %
	zip -m -j $(PACKDIR)/$(NAME)-$(basename $@)-$(VERSION).zip $(BINDIR)/$(NAME)-$(basename $@).exe

all-arch: $(PLATFORM_LIST) $(WINDOWS_ARCH_LIST)

releases: $(tar_releases) $(zip_releases)
clean:
	rm -f $(BINDIR)/*

