# pcd-skills Makefile
# SPDX-License-Identifier: CC-BY-4.0

NAME      := pcd-skills
VERSION   := 0.4.12
REPO_ROOT := ../..

.PHONY: dist clean

# Generate the source tarball for OBS/RPM builds, once per release.
# Packages the committed skills/ tree from the repository root into a
# versioned directory. Run `make skills` and `make check-skills` at the
# repository root first, so the tarball carries current manifests.
dist:
	tar -czf $(CURDIR)/$(NAME)-$(VERSION).tar.gz -C $(REPO_ROOT) \
	    --transform 's|^|$(NAME)-$(VERSION)/|' \
	    $(shell git -C $(REPO_ROOT) ls-files skills/ LICENSE)

clean:
	rm -f $(NAME)-$(VERSION).tar.gz
