BINARY  := mcp-server-pcd
VERSION := 0.2.0
STORE   := internal/store/assets

# Asset source directories.
# When building inside the PCD repository, assets live two levels up.
# When building standalone (e.g. from an OBS source tarball), they are
# expected to be staged manually or via a custom embed-assets invocation.
REPO_ROOT  := ../..
TMPL_SRC   := $(REPO_ROOT)/templates
HINTS_SRC  := $(REPO_ROOT)/hints
PROMPTS_SRC := $(REPO_ROOT)/prompts

.PHONY: all build embed-assets vendor-tarball test install clean man

all: build

# Stage assets from the repository root into the embed location.
# Must run before go build.
# Prompts: only stage actual prompt files (not README-*.md).
embed-assets:
	mkdir -p $(STORE)/templates $(STORE)/hints $(STORE)/prompts
	@if ls $(TMPL_SRC)/*.template.md 1>/dev/null 2>&1; then \
	  cp $(TMPL_SRC)/*.template.md $(STORE)/templates/; \
	else \
	  echo "WARNING: no template files found in $(TMPL_SRC)/"; \
	  touch $(STORE)/templates/.keep; \
	fi
	@if ls $(HINTS_SRC)/*.hints.md 1>/dev/null 2>&1; then \
	  cp $(HINTS_SRC)/*.hints.md $(STORE)/hints/; \
	else \
	  echo "WARNING: no hints files found in $(HINTS_SRC)/"; \
	  touch $(STORE)/hints/.keep; \
	fi
	@if ls $(PROMPTS_SRC)/*.md 1>/dev/null 2>&1; then \
	  for f in $(PROMPTS_SRC)/*.md; do \
	    base=$$(basename "$$f"); \
	    case "$$base" in README-*) ;; *) cp "$$f" $(STORE)/prompts/;; esac; \
	  done; \
	else \
	  echo "WARNING: no prompt files found in $(PROMPTS_SRC)/"; \
	  touch $(STORE)/prompts/.keep; \
	fi

build: embed-assets
	CGO_ENABLED=0 go build \
	    -mod=vendor \
	    -ldflags="-X main.serverVersion=$(VERSION)" \
	    -o $(BINARY) \
	    .

test:
	go test -mod=vendor ./independent_tests/...

install: build
	install -D -m 0755 $(BINARY) $(DESTDIR)/usr/bin/$(BINARY)
	install -D -m 0644 mcp-server-pcd.service \
	    $(DESTDIR)/usr/lib/systemd/system/mcp-server-pcd.service
	install -D -m 0644 mcp-server-pcd.1 \
	    $(DESTDIR)/usr/share/man/man1/mcp-server-pcd.1

man: mcp-server-pcd.1.md
	pandoc mcp-server-pcd.1.md -s -t man -o mcp-server-pcd.1

# Generate vendor tarball for OBS/RPM builds — run once per release.
vendor-tarball:
	go mod vendor
	tar -cJf $(BINARY)-$(VERSION)-vendor.tar.xz vendor/

clean:
	rm -rf $(STORE) $(BINARY) mcp-server-pcd.1
