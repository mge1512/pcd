#!/bin/sh
# gen-known-templates.sh — emit a Markdown bulleted list of all deployment
# templates discovered in <dir>, with each template's default target language
# extracted from its TEMPLATE-TABLE.  Sorted lexicographically for
# determinism.
#
# Usage:
#   gen-known-templates.sh <templates-dir>
#
# License: CC-BY-4.0
# SPDX-FileCopyrightText: 2026 Matthias G. Eckermann <pcd@mailbox.org>

set -eu

DIR="${1:-templates}"
test -d "$DIR" || { echo "error: not a directory: $DIR" >&2; exit 2; }

# Iterate over flat .template.md files only — auto-blocks reflect the
# canonical, discoverable layout, not subdirectory experiments.
for f in "$DIR"/*.template.md; do
    [ -e "$f" ] || continue
    name=$(basename "$f" .template.md)
    [ "$name" = "README" ] && continue

    # Default language: the first row of TEMPLATE-TABLE where the LANGUAGE
    # key is constrained as `default` (most templates) or `required`
    # (single-language templates such as python-tool).  Templates with no
    # LANGUAGE row (e.g. cockpit-module's HTML+JS+CSS triple) fall through.
    lang=$(awk -F'|' '
        /^[[:space:]]*\|[[:space:]]*LANGUAGE[[:space:]]*\|.*\|[[:space:]]*(default|required)[[:space:]]*\|/ {
            gsub(/^[[:space:]]+|[[:space:]]+$/, "", $3)
            print $3
            exit
        }
    ' "$f")

    if [ -z "$lang" ]; then
        # Templates that do not resolve to a single default language
        # (e.g. project-manifest, gui-tool which is OS-dependent) get a dash.
        lang="—"
    fi

    printf '%s\t%s\n' "$name" "$lang"
done | sort | awk -F'\t' '{ printf "- `%s` — default language: %s\n", $1, $2 }'
