#!/bin/sh
# build-prompts.sh - rebuild the derived copies under prompts/
#
# Usage:
#   sh scripts/build-prompts.sh
#
# Reads prompts/MAP, one line per copy:
#
#   <canonical path> <file name under prompts/>
#
# and copies each canonical file to prompts/<file name>, byte for byte.
# Every *.md under prompts/ other than README.md is owned by this script
# and removed before the copies are written, so a copy whose line left
# MAP disappears. Lines starting with # and blank lines are ignored.
#
# Exit status: 0 built, 2 invocation error or a canonical file that does
# not exist.
#
# License: CC-BY-4.0
# SPDX-FileCopyrightText: 2026 Matthias G. Eckermann <pcd@mailbox.org>

set -eu

case "${1:-}" in
    -h|--help)
        sed -n '2,/^$/p' "$0" | sed 's/^# \{0,1\}//'
        exit 0
        ;;
esac

REPO_ROOT=$(cd "$(dirname "$0")/.." && pwd)
cd "$REPO_ROOT"
test -f prompts/MAP || { echo "error: no prompts/MAP" >&2; exit 2; }

# Pass 1: every canonical file must exist before anything is removed.
while IFS= read -r line || [ -n "$line" ]; do
    case "$line" in ''|'#'*) continue ;; esac
    set -f; set -- $line; set +f
    [ $# -eq 2 ] || { echo "error: MAP line needs <canonical> <name>: $line" >&2; exit 2; }
    test -f "$1" || { echo "error: canonical file missing: $1" >&2; exit 2; }
    case "$2" in
        README.md|MAP|*/*|'') echo "error: MAP target not allowed: $2" >&2; exit 2 ;;
    esac
done < prompts/MAP

# Pass 2: clear and copy.
for f in prompts/*.md; do
    [ "$f" = prompts/README.md ] && continue
    [ -e "$f" ] && rm -f "$f"
done
n=0
while IFS= read -r line || [ -n "$line" ]; do
    case "$line" in ''|'#'*) continue ;; esac
    set -f; set -- $line; set +f
    cp "$1" "prompts/$2"
    n=$((n + 1))
done < prompts/MAP
printf 'prompts: %s copies\n' "$n"
