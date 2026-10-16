#!/bin/sh
# check-version-bump.sh - fail when a versioned file changes without moving
# its Version: field
#
# Usage:
#   sh scripts/check-version-bump.sh [BASE]     (default: HEAD)
#
# Compares the working tree with BASE. A file counts as versioned when one
# of its first 40 lines is a Version: field. Checked: *.md under templates/,
# hints/, prompts/, examples/ and tools/*/spec/, which covers specs, hints
# and pinned copies. Not checked: doc/, whose headers follow D-5 (the
# framework version at the last content revision), and */code/*, which is
# generated. Restoring bytes that a path has already carried under the same
# label is accepted, because it adds no new content to that label.
#
# Exit status: 0 no label gains a second content, 1 at least one does.
#
# License: CC-BY-4.0
# SPDX-FileCopyrightText: 2026 Matthias G. Eckermann <pcd@mailbox.org>

set -eu

BASE=${1:-HEAD}
TMP=$(mktemp) || exit 1
trap 'rm -f "$TMP"' EXIT

ver() {
    sed -n '1,40s/^[#* ]*Version:[* ]*//p' | sed -n '1{s/[[:space:]]*$//;p;}'
}

# Has path $1 carried blob $3 at a commit where its Version: was $2?
carried() {
    for c in $(git log --format=%H -- "$1"); do
        [ "$(git rev-parse -q --verify "$c:$1" 2> /dev/null)" = "$3" ] || continue
        [ "$(git show "$c:$1" | ver)" = "$2" ] && return 0
    done
    return 1
}

git diff --name-only "$BASE" -- templates hints prompts examples tools > "$TMP"
status=0
while IFS= read -r f; do
    case $f in
        */code/*) continue ;;
        templates/*.md|hints/*.md|prompts/*.md|examples/*.md|tools/*/spec/*.md) ;;
        *) continue ;;
    esac
    [ -f "$f" ] || continue
    old=$(git show "$BASE:$f" 2> /dev/null | ver)
    [ -n "$old" ] || continue
    new=$(ver < "$f")
    [ "$old" = "$new" ] || continue
    if carried "$f" "$new" "$(git hash-object "$f")"; then
        continue
    fi
    printf '%s: bytes changed, Version: still %s\n' "$f" "$new" >&2
    status=1
done < "$TMP"
exit $status
