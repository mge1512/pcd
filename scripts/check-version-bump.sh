#!/bin/sh
# check-version-bump.sh - fail when a versioned file changes without moving
# its Version: field
#
# Usage:
#   sh scripts/check-version-bump.sh [BASE]     (default: HEAD)
#
# Compares the working tree with BASE. A file counts as versioned when one
# of its first 40 lines is a Version: field. Checked: *.md under templates/,
# hints/, examples/ and tools/*/spec/, which covers specs, hints and pinned
# copies, and skills/*/SKILL.md, whose label is the version: key under
# metadata: in the front matter. A skill is published as one directory
# under its label, so any change under skills/<name>/, derived copies
# included, requires that label to move. Not checked: doc/, whose headers
# follow D-5 (the framework version at the last content revision), */code/*,
# which is generated, and prompts/, whose files are copies of skill
# procedure files. Restoring bytes that a path has already carried under
# the same label is accepted, because it adds no new content to that label.
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

# Front matter label of a SKILL.md: the first "  version:" line after
# "metadata:" inside the leading --- block, quotes removed.
skillver() {
    awk 'NR == 1 && $0 != "---" { exit }
         NR > 1 && $0 == "---" { exit }
         /^metadata:/ { inside = 1; next }
         inside && /^[^ ]/ { inside = 0 }
         inside && /^[ ]+version:/ {
             sub(/^[ ]+version:[ ]*/, ""); gsub(/^["'\'']|["'\'']$/, ""); print; exit }'
}

# Label of path $1 read from stdin, by file kind.
label() {
    case $1 in
        skills/*/SKILL.md) skillver ;;
        *) ver ;;
    esac
}

# Has path $1 carried blob $3 at a commit where its Version: was $2?
carried() {
    for c in $(git log --format=%H -- "$1"); do
        [ "$(git rev-parse -q --verify "$c:$1" 2> /dev/null)" = "$3" ] || continue
        [ "$(git show "$c:$1" | label "$1")" = "$2" ] && return 0
    done
    return 1
}

git diff --name-only "$BASE" -- templates hints examples tools skills > "$TMP"
status=0
while IFS= read -r f; do
    case $f in
        */code/*) continue ;;
        templates/*.md|hints/*.md|examples/*.md|tools/*/spec/*.md) ;;
        skills/*/SKILL.md) ;;
        *) continue ;;
    esac
    [ -f "$f" ] || continue
    old=$(git show "$BASE:$f" 2> /dev/null | label "$f")
    [ -n "$old" ] || continue
    new=$(label "$f" < "$f")
    [ "$old" = "$new" ] || continue
    if carried "$f" "$new" "$(git hash-object "$f")"; then
        continue
    fi
    printf '%s: bytes changed, Version: still %s\n' "$f" "$new" >&2
    status=1
done < "$TMP"

# A skill directory with any change needs its SKILL.md label moved.
for s in $(sed -n 's|^\(skills/[^/]*\)/.*|\1|p' "$TMP" | sort -u); do
    [ -f "$s/SKILL.md" ] || continue
    old=$(git show "$BASE:$s/SKILL.md" 2> /dev/null | skillver)
    [ -n "$old" ] || continue
    new=$(skillver < "$s/SKILL.md")
    [ "$old" = "$new" ] || continue
    printf '%s: content changed, SKILL.md version still %s\n' "$s" "$new" >&2
    status=1
done
exit $status
