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
# Skills under skills/ carry derived copies of templates, hints, rule
# documents and scripts, listed in each skill's SOURCES file and written by
# `make skills` together with the skill's MANIFEST.sha256. The same target
# rewrites prompts/, whose files are copies of the skills' procedure files
# (prompts/MAP). `make check-skills` fails in CI when any copy is stale.
#
# Usage:
#   make spec          regenerate all derived spec content
#   make check-spec    verify no hand-edits inside auto-blocks (CI)
#   make skills        rebuild the derived content of every skill and prompts/
#   make check-skills  verify the committed skill content and prompts/ are current (CI)
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
TRANSLATION_PROCEDURE := skills/pcd-translate/modes/translation-procedure.md

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

# ── Skills: directories with derived copies and a manifest ────────────────────

SKILLS_DIR := skills
SKILL_DIRS := $(patsubst %/SKILL.md,%,$(wildcard $(SKILLS_DIR)/*/SKILL.md))

# ── Helpers ───────────────────────────────────────────────────────────────────

REPLACE_AUTO := sh scripts/replace-auto-block.sh
GEN_KNOWN_TEMPLATES       := sh scripts/gen-known-templates.sh $(TEMPLATES_DIR)
GEN_KNOWN_TEMPLATES_COUNT := sh scripts/gen-known-templates-count.sh $(TEMPLATES_DIR)
MATERIALISE_PROMPT        := sh scripts/materialise-prompt.sh
BUILD_SKILL               := sh scripts/build-skill.sh
BUILD_PROMPTS             := sh scripts/build-prompts.sh

# ── Targets ───────────────────────────────────────────────────────────────────

.PHONY: help spec spec-doc-blocks spec-bundles \
        spec-pcd-lint spec-mcp-server-pcd check-spec check-versions \
        skills prompts check-skills

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

$(PCDLINT_BUNDLE_PROMPT): $(TRANSLATION_PROCEDURE) scripts/materialise-prompt.sh
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

# ── Skills ────────────────────────────────────────────────────────────────────
#
# Each skills/<name>/SOURCES lists the repository content copied into the
# skill; scripts/build-skill.sh rebuilds those copies and MANIFEST.sha256.
# prompts/MAP lists the copies of the skills' procedure files that keep
# the paths the templates, the MCP server build and the docs refer to.

skills: prompts ## Rebuild the derived content and manifest of every skill
	@for d in $(SKILL_DIRS); do $(BUILD_SKILL) $$d || exit 1; done

prompts: ## Rebuild prompts/ from the skills' procedure files (prompts/MAP)
	@$(BUILD_PROMPTS)

check-skills: skills ## Fail if `make skills` produced unstaged changes
	@if ! git diff --quiet -- $(SKILLS_DIR) $(PROMPTS_DIR) \
	   || [ -n "$$(git ls-files --others --exclude-standard -- $(SKILLS_DIR) $(PROMPTS_DIR))" ]; then \
	    echo "ERROR: derived skill content differs from the staged state."; \
	    echo "       Run 'make skills', then 'git add' the result."; \
	    echo ""; \
	    git status --short -- $(SKILLS_DIR) $(PROMPTS_DIR); \
	    exit 1; \
	fi

# Version labels: a label names one content. VERSION_BASE=origin/master
# checks a whole branch in CI; the default checks the working tree
# against HEAD.

VERSION_BASE ?= HEAD

check-versions: ## Fail if a versioned file changed without a Version: bump
	sh scripts/check-version-bump.sh $(VERSION_BASE)
