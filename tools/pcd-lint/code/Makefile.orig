# pcd-lint Makefile
# SPDX-License-Identifier: GPL-2.0-only

BINARY   := pcd-lint
VERSION  := 0.3.22
PREFIX   := /usr/local
BINDIR   := $(PREFIX)/bin
MANDIR   := $(PREFIX)/share/man/man1

.PHONY: all build test install clean man

all: build man

build:
	CGO_ENABLED=0 go build -ldflags="-s -w" -o $(BINARY) .

test:
	go test ./...
	@if [ -f independent_tests/INDEPENDENT_TESTS.go ]; then \
		cd independent_tests && go test -v .; \
	fi

man: $(BINARY).1

$(BINARY).1: $(BINARY).1.md
	pandoc $(BINARY).1.md -s -t man -o $(BINARY).1

install: build man
	install -Dm755 $(BINARY) $(DESTDIR)$(BINDIR)/$(BINARY)
	install -Dm644 $(BINARY).1 $(DESTDIR)$(MANDIR)/$(BINARY).1

clean:
	rm -f $(BINARY) $(BINARY).1

# Generate vendor tarball for OBS/RPM builds — run once per release.
vendor-tarball:
	go mod vendor
	tar -czf $(BINARY)-$(VERSION)-vendor.tar.gz vendor/
