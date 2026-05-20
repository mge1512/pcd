# PCD top-level Makefile
#
# Regenerates spec content that is derived from filesystem state — the
# enumerations and bundled assets that previously had to be hand-synchronised
# with the templates/, hints/, and prompts/ trees.
#
# Auto-derived spec sections are delimited by HTML-comment markers:
#   <!-- BEGIN AUTO: <name> -->
#   <!-- END AUTO: <name> -->
# Content between markers is owned by `make spec` and will be overwritten.
# Hand-edits inside these regions are rejected by `make check-spec` in CI.
#
# Usage:
#   make spec          regenerate all derived spec content
#   make check-spec    verify no hand-edits inside auto-blocks (CI)
#   make help          list targets
#
# After `make spec` the spec hash changes, which invalidates the generated
# code under tools/*/code/. Re-translate before committing — or revert the
# spec change if it was unintended.
#
# License: CC-BY-4.0
# SPDX-FileCopyrightText: 2026 Matthias G. Eckermann <pcd@mailbox.org>

SHELL := /bin/sh

# ── Inputs ────────────────────────────────────────────────────────────────────

TEMPLATES_DIR := templates
HINTS_DIR     := hints
PROMPTS_DIR   := prompts

TEMPLATE_FILES := $(filter-out $(TEMPLATES_DIR)/README.md, \
                    $(wildcard $(TEMPLATES_DIR)/*.template.md))

# ── Outputs: spec files with auto-blocks ──────────────────────────────────────

PCDLINT_SPEC := tools/pcd-lint/spec/pcd-lint.spec.md
MCPSRV_SPEC  := tools/mcp-server-pcd/spec/mcp-server-pcd.spec.md

# ── Outputs: bundled translation inputs for pcd-lint ──────────────────────────

PCDLINT_BUNDLE_TEMPLATE := tools/pcd-lint/spec/cli-tool.template.md
PCDLINT_BUNDLE_PROMPT   := tools/pcd-lint/spec/prompt.md

BUNDLE_FILES := $(PCDLINT_BUNDLE_TEMPLATE) $(PCDLINT_BUNDLE_PROMPT)
BUNDLE_HASHES := $(addsuffix .sha256, $(BUNDLE_FILES))

# ── Helpers ───────────────────────────────────────────────────────────────────

REPLACE_AUTO := sh scripts/replace-auto-block.sh
GEN_KNOWN_TEMPLATES       := sh scripts/gen-known-templates.sh $(TEMPLATES_DIR)
GEN_KNOWN_TEMPLATES_COUNT := sh scripts/gen-known-templates-count.sh $(TEMPLATES_DIR)
MATERIALISE_PROMPT        := sh scripts/materialise-prompt.sh

# ── Targets ───────────────────────────────────────────────────────────────────

.PHONY: help spec spec-doc-blocks spec-bundles \
        spec-pcd-lint spec-mcp-server-pcd check-spec

help: ## Show this help
	@awk 'BEGIN{FS=":.*##"} /^[a-zA-Z][a-zA-Z0-9_-]*:.*##/ \
	    {printf "  %-22s %s\n",$$1,$$2}' $(MAKEFILE_LIST)

spec: spec-doc-blocks spec-bundles ## Regenerate all derived spec content

spec-doc-blocks: spec-pcd-lint spec-mcp-server-pcd ## Refresh auto-blocks in spec files
spec-bundles:    $(BUNDLE_FILES) $(BUNDLE_HASHES) ## Refresh bundled assets and hashes

# ── pcd-lint spec auto-blocks ─────────────────────────────────────────────────
#
# Auto-blocks expected in $(PCDLINT_SPEC):
#   known-templates        — bulleted list of templates with default language
#   known-templates-count  — count phrase used in BEHAVIOR: list-templates

spec-pcd-lint: $(PCDLINT_SPEC) $(TEMPLATE_FILES)
	$(REPLACE_AUTO) $(PCDLINT_SPEC) known-templates       "$(GEN_KNOWN_TEMPLATES)"
	$(REPLACE_AUTO) $(PCDLINT_SPEC) known-templates-count "$(GEN_KNOWN_TEMPLATES_COUNT)"

# ── mcp-server-pcd spec auto-blocks ───────────────────────────────────────────
#
# Auto-blocks expected in $(MCPSRV_SPEC):
#   known-templates        — bulleted list shown in the DeploymentTemplate
#                            doc-comment

spec-mcp-server-pcd: $(MCPSRV_SPEC) $(TEMPLATE_FILES)
	$(REPLACE_AUTO) $(MCPSRV_SPEC) known-templates "$(GEN_KNOWN_TEMPLATES)"

# ── Bundled translation inputs for pcd-lint ───────────────────────────────────

$(PCDLINT_BUNDLE_TEMPLATE): $(TEMPLATES_DIR)/cli-tool.template.md
	cp $< $@

$(PCDLINT_BUNDLE_PROMPT): $(PROMPTS_DIR)/prompt.md scripts/materialise-prompt.sh
	$(MATERIALISE_PROMPT) \
	    --deployment-template cli-tool \
	    --spec-name pcd-lint \
	    $< > $@

%.sha256: %
	sha256sum $< | awk '{print $$1}' > $@

# ── CI check ──────────────────────────────────────────────────────────────────

check-spec: spec ## Fail if `make spec` produced uncommitted changes
	@if ! git diff --quiet -- \
	        $(PCDLINT_SPEC) $(MCPSRV_SPEC) \
	        $(BUNDLE_FILES) $(BUNDLE_HASHES); then \
	    echo "ERROR: derived spec content differs from committed state."; \
	    echo "       Run 'make spec', then commit and re-translate."; \
	    echo ""; \
	    git diff --stat -- \
	        $(PCDLINT_SPEC) $(MCPSRV_SPEC) \
	        $(BUNDLE_FILES) $(BUNDLE_HASHES); \
	    exit 1; \
	fi
