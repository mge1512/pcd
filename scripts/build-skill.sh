#!/bin/sh
# build-skill.sh - rebuild the derived content of one skill directory
#
# Usage:
#   sh scripts/build-skill.sh <skill-dir>
#
# A skill directory contains hand-authored files (SKILL.md, SOURCES, modes/)
# and derived copies of repository content. The SOURCES file lists the
# copies, one per line:
#
#   <pattern> <subdir>
#
# <pattern> is a shell glob relative to the repository root; <subdir> is
# the directory inside the skill that receives the matched files. Lines
# starting with # and blank lines are ignored. A subdirectory named in
# SOURCES is owned by this script: it is removed and rebuilt on every run,
# so nothing hand-authored may live there.
#
# After copying, MANIFEST.sha256 is written at the skill root: one
# sha256sum line per file, paths relative to the skill root, sorted, the
# manifest itself excluded. `sha256sum -c MANIFEST.sha256` run from the
# skill root verifies an installed copy.
#
# Exit status: 0 built, 2 invocation error or a pattern that matched
# nothing.
#
# License: CC-BY-4.0
# SPDX-FileCopyrightText: 2026 Matthias G. Eckermann <pcd@mailbox.org>

set -eu

case "${1:-}" in
    -h|--help|'')
        sed -n '2,/^$/p' "$0" | sed 's/^# \{0,1\}//'
        [ -n "${1:-}" ] && exit 0
        echo "error: skill directory is required" >&2
        exit 2
        ;;
esac

REPO_ROOT=$(cd "$(dirname "$0")/.." && pwd)
cd "$REPO_ROOT"

SKILL=${1%/}
test -d "$SKILL"           || { echo "error: not a directory: $SKILL" >&2; exit 2; }
test -f "$SKILL/SKILL.md"  || { echo "error: no SKILL.md in $SKILL" >&2; exit 2; }
test -f "$SKILL/SOURCES"   || { echo "error: no SOURCES in $SKILL" >&2; exit 2; }

# Pass 1: validate every entry and clear the target subdirectories.
# Globbing is off while a line is split, so the pattern stays one word.
cleared=""
while IFS= read -r line || [ -n "$line" ]; do
    case "$line" in ''|'#'*) continue ;; esac
    set -f; set -- $line; set +f
    [ $# -eq 2 ] || { echo "error: SOURCES line needs <pattern> <subdir>: $line" >&2; exit 2; }
    pattern=$1; sub=$2
    case "$sub" in
        ''|.|..|/*|*/*|*..*)
            echo "error: SOURCES subdir must be one plain name: $sub" >&2; exit 2 ;;
        SKILL.md|SOURCES|modes|MANIFEST.sha256)
            echo "error: SOURCES must not target hand-authored content: $sub" >&2; exit 2 ;;
    esac
    found=0
    for f in $pattern; do
        [ -f "$f" ] && found=1
    done
    [ $found -eq 1 ] || { echo "error: SOURCES pattern matched nothing: $pattern" >&2; exit 2; }
    case " $cleared " in
        *" $sub "*) ;;
        *) rm -rf "$SKILL/$sub"; mkdir -p "$SKILL/$sub"; cleared="$cleared $sub" ;;
    esac
done < "$SKILL/SOURCES"

# Pass 2: copy. -p keeps the mode, so an executable script stays executable.
while IFS= read -r line || [ -n "$line" ]; do
    case "$line" in ''|'#'*) continue ;; esac
    set -f; set -- $line; set +f
    pattern=$1; sub=$2
    for f in $pattern; do
        [ -f "$f" ] || continue
        cp -p "$f" "$SKILL/$sub/$(basename "$f")"
    done
done < "$SKILL/SOURCES"

# Manifest: every file except the manifest itself, relative paths, sorted.
(
    cd "$SKILL"
    find . -type f ! -name MANIFEST.sha256 | sed 's|^\./||' | LC_ALL=C sort \
        | while IFS= read -r f; do sha256sum "$f"; done > MANIFEST.sha256
)

printf '%s: %s files\n' "$SKILL" "$(wc -l < "$SKILL/MANIFEST.sha256" | tr -d ' ')"
