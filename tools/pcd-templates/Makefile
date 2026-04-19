# pcd-templates Makefile
# SPDX-License-Identifier: GPL-2.0-only

NAME     := pcd-templates
VERSION  := 0.3.22
REPO_ROOT := ../..

TEMPLATES := $(REPO_ROOT)/templates
HINTS     := $(REPO_ROOT)/hints

.PHONY: dist clean

# Generate source tarball for OBS/RPM builds — run once per release.
# Packages templates/ and hints/ from the repo root into a versioned directory.
dist:
	tar -czf $(NAME)-$(VERSION).tar.gz \
	    --transform 's|^$(REPO_ROOT)/|$(NAME)-$(VERSION)/|' \
	    $(shell git -C $(REPO_ROOT) ls-files templates/ hints/ LICENSE)

clean:
	rm -f $(NAME)-$(VERSION).tar.gz
