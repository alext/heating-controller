.PHONY: build test clean

BINARY := heating-controller
IMPORT_BASE := github.com/alext
IMPORT_PATH := $(IMPORT_BASE)/heating-controller

GOPATH := $(CURDIR)/Godeps/_workspace:$(GOPATH)
export GOPATH

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
