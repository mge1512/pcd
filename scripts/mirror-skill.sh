#!/bin/sh
# mirror-skill.sh - publish one built skill directory to its mirror repository
#
# Usage:
#   sh scripts/mirror-skill.sh <skill-name> <remote-url> [branch]
#
# Takes skills/<skill-name>/ as committed in this repository and pushes it
# as the complete content of the mirror repository's branch (default:
# main), with SKILL.md at the repository root. One commit per call, on
# top of the mirror's history, so that a clone of the mirror updates with
# `git pull`. The commit message names the source commit; a tag that
# points at HEAD here is set on the mirror too.
#
# Preconditions: the working tree is clean under skills/<skill-name>/
# (run `make skills` and commit first), `git` can push to <remote-url>.
# Nothing is read from the network except the mirror's current branch.
#
# Exit status: 0 pushed, 1 nothing to push (mirror already current),
# 2 invocation error or a precondition not met.
#
# License: CC-BY-4.0
# SPDX-FileCopyrightText: 2026 Matthias G. Eckermann <pcd@mailbox.org>

set -eu

case "${1:-}" in
    -h|--help|'')
        sed -n '2,/^$/p' "$0" | sed 's/^# \{0,1\}//'
        [ -n "${1:-}" ] && exit 0
        echo "error: skill name and remote URL are required" >&2
        exit 2
        ;;
esac

NAME=$1
REMOTE=${2:-}
BRANCH=${3:-main}
test -n "$REMOTE" || { echo "error: remote URL is required" >&2; exit 2; }

REPO_ROOT=$(cd "$(dirname "$0")/.." && pwd)
cd "$REPO_ROOT"
SKILL=skills/$NAME
test -f "$SKILL/SKILL.md"        || { echo "error: no SKILL.md under $SKILL" >&2; exit 2; }
test -f "$SKILL/MANIFEST.sha256" || { echo "error: no MANIFEST.sha256 under $SKILL; run make skills" >&2; exit 2; }
if [ -n "$(git status --porcelain -- "$SKILL")" ]; then
    echo "error: $SKILL has uncommitted changes; commit first" >&2
    exit 2
fi
(cd "$SKILL" && sha256sum -c --quiet --strict MANIFEST.sha256) \
    || { echo "error: $SKILL does not match its manifest; run make skills" >&2; exit 2; }

SRC=$(git rev-parse HEAD)
SRC_SHORT=$(git rev-parse --short HEAD)
TAG=$(git describe --tags --exact-match HEAD 2> /dev/null || true)
VERSION=$(sed -n '/^metadata:/,/^[^ ]/{s/^[ ]*version:[ ]*//p;}' "$SKILL/SKILL.md" | sed 's/^"\(.*\)"$/\1/' | sed -n '1p')

WORK=$(mktemp -d) || exit 2
trap 'rm -rf "$WORK"' EXIT

# Fetch the mirror's branch if it exists; start empty otherwise.
git init -q -b "$BRANCH" "$WORK/mirror"
if git -C "$WORK/mirror" fetch -q "$REMOTE" "$BRANCH" 2> /dev/null; then
    git -C "$WORK/mirror" reset -q --hard FETCH_HEAD
fi

# Replace the content with the committed skill directory, byte for byte.
find "$WORK/mirror" -mindepth 1 -maxdepth 1 ! -name .git -exec rm -rf {} +
git archive --format=tar HEAD "$SKILL" | tar -x -C "$WORK" -f -
cp -R "$WORK/$SKILL/." "$WORK/mirror/"
rm -rf "$WORK/skills"

git -C "$WORK/mirror" add -A
if git -C "$WORK/mirror" diff --cached --quiet; then
    echo "$NAME: mirror is current at $REMOTE ($BRANCH)"
    exit 1
fi
AUTHOR_NAME=$(git config user.name || true)
AUTHOR_MAIL=$(git config user.email || true)
git -C "$WORK/mirror" -c user.name="${AUTHOR_NAME:-pcd release}" \
    -c user.email="${AUTHOR_MAIL:-pcd@mailbox.org}" \
    commit -q -m "$NAME ${VERSION:-unversioned}, built from mge1512/pcd $SRC_SHORT

Source commit: $SRC
Mirror of skills/$NAME in https://github.com/mge1512/pcd. Do not edit here;
file issues in mge1512/pcd."

git -C "$WORK/mirror" -c push.negotiate=false push -q "$REMOTE" "$BRANCH:$BRANCH"
if [ -n "$TAG" ]; then
    git -C "$WORK/mirror" tag -f "$TAG" >/dev/null
    git -C "$WORK/mirror" push -q -f "$REMOTE" "refs/tags/$TAG"
fi
echo "$NAME ${VERSION:-unversioned}: pushed to $REMOTE ($BRANCH${TAG:+, tag $TAG}) from $SRC_SHORT"
