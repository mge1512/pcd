#!/bin/sh
# replace-auto-block.sh — replace content between auto-block markers in a
# Markdown file with the stdout of a generator command.
#
# Usage:
#   replace-auto-block.sh <file> <marker-name> <generator-command>
#
# Markers in <file> must be:
#   <!-- BEGIN AUTO: <marker-name> -->
#   <!-- END AUTO: <marker-name> -->
#
# Both markers must be present exactly once.  The markers themselves are
# preserved; only the content strictly between them is replaced.
#
# The generator command is run via `sh -c` so it may include arguments.
# Its stdout becomes the new content of the block.  If the generator
# exits non-zero, the file is not modified.
#
# License: CC-BY-4.0
# SPDX-FileCopyrightText: 2026 Matthias G. Eckermann <pcd@mailbox.org>

set -eu

if [ $# -ne 3 ]; then
    echo "usage: $0 <file> <marker-name> <generator-command>" >&2
    exit 2
fi

FILE="$1"
NAME="$2"
GEN="$3"

BEGIN_MARK="<!-- BEGIN AUTO: $NAME -->"
END_MARK="<!-- END AUTO: $NAME -->"

test -f "$FILE" || { echo "error: not a file: $FILE" >&2; exit 2; }

# Markers must be present and unique.
b_count=$(grep -Fc -- "$BEGIN_MARK" "$FILE" || true)
e_count=$(grep -Fc -- "$END_MARK"   "$FILE" || true)
if [ "$b_count" -ne 1 ] || [ "$e_count" -ne 1 ]; then
    echo "error: $FILE must contain exactly one '$BEGIN_MARK' and one '$END_MARK'" >&2
    echo "       found: begin=$b_count, end=$e_count" >&2
    exit 2
fi

# Run the generator first; if it fails, abort before mutation.
GEN_OUT=$(mktemp)
trap 'rm -f "$GEN_OUT" "$TMP_OUT"' EXIT
TMP_OUT=$(mktemp)

if ! sh -c "$GEN" > "$GEN_OUT"; then
    echo "error: generator failed: $GEN" >&2
    exit 2
fi

awk -v B="$BEGIN_MARK" -v E="$END_MARK" -v G="$GEN_OUT" '
    BEGIN { skip = 0 }
    $0 == B {
        print
        while ((getline line < G) > 0) print line
        close(G)
        skip = 1
        next
    }
    $0 == E { skip = 0; print; next }
    skip == 0 { print }
' "$FILE" > "$TMP_OUT"

# Only overwrite if the new content differs — avoids touching mtime needlessly.
if ! cmp -s "$FILE" "$TMP_OUT"; then
    cp "$TMP_OUT" "$FILE"
    echo "updated: $FILE  [$NAME]"
fi
