#!/bin/sh
# materialise-prompt.sh — resolve placeholders in prompts/prompt.md for a
# specific translation run.  Replaces the generic placeholder tokens with
# the concrete filenames of the deployment template and spec.
#
# Usage:
#   materialise-prompt.sh --deployment-template <name> --spec-name <name> <input>
#
# Replaces:
#   <deployment-template>  →  <name>
#   <spec-name>            →  <name>
#
# Output is written to stdout.  The input file is not modified.
#
# License: CC-BY-4.0
# SPDX-FileCopyrightText: 2026 Matthias G. Eckermann <pcd@mailbox.org>

set -eu

DT=""
SN=""

while [ $# -gt 0 ]; do
    case "$1" in
        --deployment-template) DT="${2:-}"; shift 2 ;;
        --spec-name)           SN="${2:-}"; shift 2 ;;
        -h|--help)
            sed -n '2,/^$/p' "$0" | sed 's/^# \{0,1\}//'
            exit 0
            ;;
        --) shift; break ;;
        -*) echo "unknown option: $1" >&2; exit 2 ;;
        *)  break ;;
    esac
done

test -n "$DT" || { echo "error: --deployment-template is required" >&2; exit 2; }
test -n "$SN" || { echo "error: --spec-name is required"           >&2; exit 2; }

INPUT="${1:-}"
test -n "$INPUT" || { echo "error: input file is required" >&2; exit 2; }
test -f "$INPUT" || { echo "error: not a file: $INPUT"     >&2; exit 2; }

sed -e "s|<deployment-template>|$DT|g" \
    -e "s|<spec-name>|$SN|g" \
    "$INPUT"
