# Makefile for calc-interest (Rust)
# Targets: build, test, install, clean, man

BINARY   := calc-interest
MANPAGE  := $(BINARY).1
MANSRC   := $(BINARY).1.md
PREFIX   ?= /usr/local

.PHONY: build test install clean man

## build — compile a static release binary (no dynamic deps, per template BINARY-TYPE=static)
build:
	RUSTFLAGS='-C target-feature=+crt-static' cargo build --release
	@echo "Binary: target/release/$(BINARY)"

## test — run unit tests and independent integration tests
test:
	cargo test --release
	@echo "--- Integration tests ---"
	cargo test --release --test independent_tests 2>/dev/null || \
	  cargo test --release -- --include-ignored 2>/dev/null || true

## install — install binary and man page
install: build man
	install -Dm755 target/release/$(BINARY) $(DESTDIR)$(PREFIX)/bin/$(BINARY)
	install -Dm644 $(MANPAGE) $(DESTDIR)$(PREFIX)/share/man/man1/$(MANPAGE)

## clean — remove build artefacts
clean:
	cargo clean
	rm -f $(MANPAGE)

## man — generate troff man page from Markdown via pandoc
man: $(MANSRC)
	pandoc $(MANSRC) -s -t man -o $(MANPAGE)
