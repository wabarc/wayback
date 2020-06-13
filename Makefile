export GO111MODULE = on
export GOPROXY = https://proxy.golang.org

BIN ?= ./bin
BINARY = wayback
PROJECT := github.com/wabarc/wayback
TARGET ?= $(BIN)/$(BINARY)
VERSION = $(strip $(shell cat version))
RELEASE_VERSION = v$(VERSION)
PACKAGES = $(shell go list ./...)

.PHONY: build
build:
	@echo "-> Building package"
	@go build -ldflags="-s -w" -gcflags=all="-l -B" -o $(TARGET) cmd/*.go

format:
	@echo "-> Running go fmt"
	@go fmt $(PACKAGES)

clean:
	rm -f $(TARGET)

