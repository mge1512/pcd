#!/bin/sh
# gen-known-templates-count.sh — emit a one-line phrase stating how many
# deployment templates the BEHAVIOR: list-templates output must produce.
#
# Usage:
#   gen-known-templates-count.sh <templates-dir>
#
# Used inside the `known-templates-count` auto-block of pcd-lint.md to keep
# the postcondition assertion ("stdout contains exactly N lines, one per
# known DeploymentTemplate value") in sync with the actual template count.
#
# License: CC-BY-4.0
# SPDX-FileCopyrightText: 2026 Matthias G. Eckermann <pcd@mailbox.org>

set -eu

DIR="${1:-templates}"
test -d "$DIR" || { echo "error: not a directory: $DIR" >&2; exit 2; }

n=0
for f in "$DIR"/*.template.md; do
    [ -e "$f" ] || continue
    name=$(basename "$f" .template.md)
    [ "$name" = "README" ] && continue
    n=$((n + 1))
done

printf 'exactly %d lines, one per known DeploymentTemplate value\n' "$n"
