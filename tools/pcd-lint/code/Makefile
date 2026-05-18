# generated from spec: pcd-lint.md sha256:293541ab62274835c61de50947f6283748831c4681cf3f02c4be2f8e942d28a9
# SPDX-License-Identifier: GPL-2.0-only

SPEC_SHA256 = 293541ab62274835c61de50947f6283748831c4681cf3f02c4be2f8e942d28a9

BINARY     = pcd-lint
VERSION    = 0.4.0
MODULE     = github.com/mge1512/pcd/tools/pcd-lint
MANPAGE    = pcd-lint.1
MANSRC     = pcd-lint.1.md

.PHONY: all build test install clean man

all: build man

build:
	CGO_ENABLED=0 go build -ldflags="-s -w" -o $(BINARY) ./cmd/pcd-lint/

test:
	go test ./...

install:
	install -Dm755 $(BINARY) $(DESTDIR)/usr/bin/$(BINARY)
	install -Dm644 $(MANPAGE) $(DESTDIR)/usr/share/man/man1/$(MANPAGE)

clean:
	rm -f $(BINARY) $(MANPAGE)

man:
	pandoc $(MANSRC) -s -t man -o $(MANPAGE)
