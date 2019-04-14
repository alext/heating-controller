.PHONY: build test clean

BINARY := heating-controller
IMPORT_BASE := github.com/alext
IMPORT_PATH := $(IMPORT_BASE)/heating-controller

export GOFLAGS = -mod=vendor

ifdef RELEASE_VERSION
VERSION := $(RELEASE_VERSION)
else
VERSION := $(shell git describe --always --tags | tr -d '\n'; test -z "`git status --porcelain`" || echo '-dirty')
endif

build:
	go build -ldflags=$(IMPORT_PATH)="-X main.version=$(VERSION) -X github.com/prometheus/common/version.Version=$(VERSION) -X github.com/prometheus/common/version.Revision=$(shell git rev-parse --short HEAD) -X github.com/prometheus/common/version.Branch=$(shell git symbolic-ref --short HEAD)" -o $(BINARY)

test: build
	go test -v ./...
	./$(BINARY) -version

clean:
	rm -rf $(BINARY)
