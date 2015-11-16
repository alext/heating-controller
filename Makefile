.PHONY: build test clean

BINARY := heating-controller
IMPORT_BASE := github.com/alext
IMPORT_PATH := $(IMPORT_BASE)/heating-controller

GO15VENDOREXPERIMENT := 1
export GO15VENDOREXPERIMENT

ifdef RELEASE_VERSION
VERSION := $(RELEASE_VERSION)
else
VERSION := $(shell git describe --always | tr -d '\n'; test -z "`git status --porcelain`" || echo '-dirty')
endif

build: Godeps/Godeps.json
	go build -ldflags "-X main.version=$(VERSION)" -o $(BINARY)

test: build
	go test -v ./...
	./$(BINARY) -version

clean:
	rm -rf $(BINARY)
