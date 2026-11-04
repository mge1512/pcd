#!/bin/sh
# check-tools.sh - verify a PCD skill's files and find the PCD engine
#
# Usage:
#   sh scripts/check-tools.sh [skill-root]
#
# Run by a session before it uses a PCD skill. skill-root is the directory
# that contains SKILL.md; the default is the parent of this script's
# directory, which is where `make skills` places the copy.
#
# Checks, in this order:
#   1. Every file under skill-root matches MANIFEST.sha256, and no file is
#      present that the manifest does not list. A .git directory, as left
#      by a clone, and a .github directory, which holds the mirror's
#      release workflow, are ignored. A failure here means the skill was
#      installed or transferred incompletely; nothing else is checked.
#   2. For every tool the SKILL.md front matter names with a
#      metadata key <tool>-min, the tool is found on PATH, then under
#      skill-root/bin/, its `version` output parses, its version is at
#      least the minimum, and for pcd-lint its schema equals the skill's
#      spec-schema in major.minor.
#
# stdout, one line each:
#   skill <name> <version> spec-schema <schema>
#   <tool> <version> cli <path>        found and acceptable
#   <tool> <version> old <path>        found, below the minimum
#   <tool> <version> schema <path>     pcd-lint found, schema differs
#   <tool> unverified <path>           found, version output unreadable
#   <tool> missing                     not found
#
# Exit status: 0 every tool acceptable; 1 at least one tool missing, old,
# unverified or on a different schema; 2 skill files damaged, SKILL.md
# unreadable or invocation error. What a session does with 1 is the
# skill's decision, stated in its SKILL.md.
#
# Needs: POSIX sh, awk, sed, find, sort, GNU coreutils sha256sum.
#
# License: CC-BY-4.0
# SPDX-FileCopyrightText: 2026 Matthias G. Eckermann <pcd@mailbox.org>

set -u

case "${1:-}" in
    -h|--help)
        sed -n '2,/^$/p' "$0" | sed 's/^# \{0,1\}//'
        exit 0
        ;;
esac

if [ $# -gt 1 ]; then
    echo "error: at most one argument, the skill root" >&2
    exit 2
fi
ROOT=${1:-$(cd "$(dirname "$0")/.." && pwd)}
ROOT=${ROOT%/}

test -f "$ROOT/SKILL.md"        || { echo "error: no SKILL.md under $ROOT" >&2; exit 2; }
test -f "$ROOT/MANIFEST.sha256" || { echo "error: no MANIFEST.sha256 under $ROOT" >&2; exit 2; }

# --- 1. Integrity ------------------------------------------------------------

if ! (cd "$ROOT" && sha256sum -c --quiet --strict MANIFEST.sha256) 1>&2; then
    echo "error: skill files do not match MANIFEST.sha256; reinstall the skill" >&2
    exit 2
fi

listed=$(sed 's/^[0-9a-f]\{64\} [ *]//' "$ROOT/MANIFEST.sha256" | LC_ALL=C sort)
present=$(cd "$ROOT" && find . \( -path ./.git -o -path ./.github \) -prune -o -type f ! -name MANIFEST.sha256 -print | sed 's|^\./||' | LC_ALL=C sort)
if [ "$listed" != "$present" ]; then
    echo "error: files under $ROOT differ from MANIFEST.sha256:" >&2
    printf '%s\n' "$listed"  > "${TMPDIR:-/tmp}/check-tools.listed.$$"
    printf '%s\n' "$present" > "${TMPDIR:-/tmp}/check-tools.present.$$"
    diff "${TMPDIR:-/tmp}/check-tools.listed.$$" "${TMPDIR:-/tmp}/check-tools.present.$$" \
        | sed -n 's/^</  not present:/p; s/^>/  not listed: /p' >&2
    rm -f "${TMPDIR:-/tmp}/check-tools.listed.$$" "${TMPDIR:-/tmp}/check-tools.present.$$"
    exit 2
fi

# --- 2. Front matter ---------------------------------------------------------

# The front matter is the block between the first two lines that are
# exactly "---". meta KEY prints the value of "  KEY:" inside "metadata:",
# with surrounding quotes removed.
front=$(awk 'NR==1 && $0!="---" {exit} NR>1 && $0=="---" {exit} NR>1 {print}' "$ROOT/SKILL.md")
[ -n "$front" ] || { echo "error: SKILL.md has no front matter" >&2; exit 2; }

top() {
    printf '%s\n' "$front" | sed -n "s/^$1:[[:space:]]*//p" | sed 's/^"\(.*\)"$/\1/; s/^'"'"'\(.*\)'"'"'$/\1/' | sed -n '1p'
}
meta() {
    printf '%s\n' "$front" | awk -v key="$1" '
        /^metadata:/ { inside = 1; next }
        inside && /^[^ ]/ { inside = 0 }
        inside {
            sub(/^[ ]+/, "")
            if (index($0, key ":") == 1) {
                v = substr($0, length(key) + 2)
                sub(/^[ ]+/, "", v); sub(/[ ]+$/, "", v)
                gsub(/^"|"$/, "", v); gsub(/^'\''|'\''$/, "", v)
                print v; exit
            }
        }'
}

name=$(top name)
version=$(meta version)
schema=$(meta spec-schema)
[ -n "$name" ]    || { echo "error: SKILL.md front matter has no name" >&2; exit 2; }
[ -n "$version" ] || { echo "error: SKILL.md metadata has no version" >&2; exit 2; }
[ -n "$schema" ]  || { echo "error: SKILL.md metadata has no spec-schema" >&2; exit 2; }
printf 'skill %s %s spec-schema %s\n' "$name" "$version" "$schema"

# --- 3. Tools ----------------------------------------------------------------

# vge A B: true when version A is at least version B (numeric, dot-separated).
vge() {
    awk -v a="$1" -v b="$2" 'BEGIN {
        na = split(a, x, "."); nb = split(b, y, ".")
        n = (na > nb) ? na : nb
        for (i = 1; i <= n; i++) {
            p = (i <= na) ? x[i] + 0 : 0
            q = (i <= nb) ? y[i] + 0 : 0
            if (p > q) exit 0
            if (p < q) exit 1
        }
        exit 0
    }'
}
mm() { printf '%s\n' "$1" | awk -F. '{ print $1 "." $2 }'; }

tools=$(printf '%s\n' "$front" | awk '
    /^metadata:/ { inside = 1; next }
    inside && /^[^ ]/ { inside = 0 }
    inside && /^[ ]+[a-z0-9-]+-min:/ { sub(/^[ ]+/, ""); sub(/-min:.*/, ""); print }')
[ -n "$tools" ] || { echo "error: SKILL.md metadata names no <tool>-min key" >&2; exit 2; }

status=0
for tool in $tools; do
    min=$(meta "$tool-min")
    path=$(command -v "$tool" 2> /dev/null || true)
    if [ -z "$path" ] && [ -x "$ROOT/bin/$tool" ]; then
        path="$ROOT/bin/$tool"
    fi
    if [ -z "$path" ]; then
        printf '%s missing\n' "$tool"
        status=1
        continue
    fi
    line=$("$path" version 2> /dev/null | sed -n '1p') || true
    have=$(printf '%s\n' "$line" | awk -v t="$tool" '$1 == t && $2 ~ /^[0-9]+(\.[0-9]+)*$/ { print $2 }')
    if [ -z "$have" ]; then
        printf '%s unverified %s\n' "$tool" "$path"
        status=1
        continue
    fi
    if ! vge "$have" "$min"; then
        printf '%s %s old %s\n' "$tool" "$have" "$path"
        status=1
        continue
    fi
    if [ "$tool" = pcd-lint ]; then
        tschema=$(printf '%s\n' "$line" | sed -n 's/.*(schema \([0-9.]*\)).*/\1/p')
        if [ -z "$tschema" ] || [ "$(mm "$tschema")" != "$(mm "$schema")" ]; then
            printf '%s %s schema %s\n' "$tool" "$have" "$path"
            status=1
            continue
        fi
    fi
    printf '%s %s cli %s\n' "$tool" "$have" "$path"
done
exit $status
