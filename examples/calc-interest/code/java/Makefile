# Makefile for calc-interest (Java / Maven)
# Spec: calc-interest v0.1.0  License: Apache-2.0
#
# Targets:
#   build    — compile and package fat JAR
#   test     — run unit tests
#   install  — install JAR + wrapper script to $(PREFIX)/bin
#   clean    — remove build artefacts
#   man      — generate man page from Markdown via pandoc
#
# Template compliance: cli-tool.template.md v0.3.20
#   BINARY-TYPE: static (fat JAR — no external runtime deps beyond JRE)
#   STREAM-OUTPUT: stdout  STREAM-DIAGNOSTICS: stderr
#   EXIT-CODE-OK: 0  EXIT-CODE-ERROR: 1  EXIT-CODE-INVOCATION: 2

NAME    := calc-interest
VERSION := 0.1.0
PREFIX  ?= /usr/local
JARDIR  := target
JAR     := $(JARDIR)/$(NAME).jar

MVN     ?= mvn
PANDOC  ?= pandoc

.PHONY: all build test install clean man

all: build

## build — compile sources and assemble fat JAR
build:
	$(MVN) -q package -DskipTests

## test — compile and run all unit tests
test:
	$(MVN) -q test

## man — generate troff man page from Markdown source
man: $(NAME).1

$(NAME).1: $(NAME).1.md
	$(PANDOC) $(NAME).1.md -s -t man -o $(NAME).1

## install — install wrapper script and JAR under $(PREFIX)
install: build man
	install -d $(DESTDIR)$(PREFIX)/bin
	install -d $(DESTDIR)$(PREFIX)/share/java
	install -d $(DESTDIR)$(PREFIX)/share/man/man1
	install -m 644 $(JAR) $(DESTDIR)$(PREFIX)/share/java/$(NAME).jar
	# Write a thin wrapper script so the tool is invocable as `calc-interest`
	printf '#!/bin/sh\nexec java -jar %s/share/java/%s.jar "$$@"\n' \
	       "$(PREFIX)" "$(NAME)" \
	       > $(DESTDIR)$(PREFIX)/bin/$(NAME)
	chmod 755 $(DESTDIR)$(PREFIX)/bin/$(NAME)
	install -m 644 $(NAME).1 $(DESTDIR)$(PREFIX)/share/man/man1/$(NAME).1

## clean — remove all generated artefacts
clean:
	$(MVN) -q clean
	rm -f $(NAME).1
