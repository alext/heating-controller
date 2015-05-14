.PHONY: build test clean deps

IMPORT_BASE := github.com/alext
IMPORT_PATH := $(IMPORT_BASE)/heating-controller
VENDOR_STAMP := _vendor/stamp

build: $(VENDOR_STAMP)
	gom build -o heating-controller

clean:
	rm -f heating-controller
	rm -rf _vendor

test: $(VENDOR_STAMP)
	gom test -v ./...

deps: $(VENDOR_STAMP)

$(VENDOR_STAMP): Gomfile
	rm -f _vendor/src/$(IMPORT_PATH)
	mkdir -p _vendor/src/$(IMPORT_BASE)
	ln -s $(CURDIR) _vendor/src/$(IMPORT_PATH)
	gom install
	touch $(VENDOR_STAMP)
