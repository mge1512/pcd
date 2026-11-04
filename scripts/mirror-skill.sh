#!/bin/sh
# mirror-skill.sh - publish one built skill directory to its mirror repository
#
# Usage:
#   sh scripts/mirror-skill.sh <skill-name> <remote-url> [branch]
#
# Takes skills/<skill-name>/ as committed in this repository and pushes it
# as the complete content of the mirror repository's branch (default:
# main), with SKILL.md at the repository root, plus the release workflow
# from tools/pcd-skills/mirror-release.yml under .github/workflows/. One
# commit per call, on top of the mirror's history, so that a clone of the
# mirror updates with `git pull`. The commit message names the source
# commit.
#
# The skill's version (metadata version in SKILL.md) becomes the tag
# v<version> on the mirror, and the mirror's workflow builds the .skill
# archive and the release from that tag. A version that the mirror has
# already published cannot be published again with different content:
# move the skill's version first.
#
# Preconditions: the working tree is clean under skills/<skill-name>/
# (run `make skills` and commit first), `git` can push to <remote-url>.
#
# Exit status: 0 pushed, 1 nothing to push (mirror already current),
# 2 invocation error, a precondition not met, or the version already
# published with other content.
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
WORKFLOW=tools/pcd-skills/mirror-release.yml
test -f "$SKILL/SKILL.md"        || { echo "error: no SKILL.md under $SKILL" >&2; exit 2; }
test -f "$SKILL/MANIFEST.sha256" || { echo "error: no MANIFEST.sha256 under $SKILL; run make skills" >&2; exit 2; }
test -f "$WORKFLOW"              || { echo "error: no $WORKFLOW" >&2; exit 2; }
if [ -n "$(git status --porcelain -- "$SKILL" "$WORKFLOW")" ]; then
    echo "error: $SKILL or $WORKFLOW has uncommitted changes; commit first" >&2
    exit 2
fi
(cd "$SKILL" && sha256sum -c --quiet --strict MANIFEST.sha256) \
    || { echo "error: $SKILL does not match its manifest; run make skills" >&2; exit 2; }

VERSION=$(sed -n '/^metadata:/,/^[^ ]/{s/^[ ]*version:[ ]*//p;}' "$SKILL/SKILL.md" | sed 's/^"\(.*\)"$/\1/' | sed -n '1p')
test -n "$VERSION" || { echo "error: $SKILL/SKILL.md has no metadata version" >&2; exit 2; }
TAG=v$VERSION
SRC=$(git rev-parse HEAD)
SRC_SHORT=$(git rev-parse --short HEAD)

WORK=$(mktemp -d) || exit 2
trap 'rm -rf "$WORK"' EXIT

# The mirror's branch, if it exists; an empty history otherwise.
git init -q -b "$BRANCH" "$WORK/mirror"
if git -C "$WORK/mirror" fetch -q "$REMOTE" "$BRANCH" 2> /dev/null; then
    git -C "$WORK/mirror" reset -q --hard FETCH_HEAD
fi
published=$(git ls-remote --tags "$REMOTE" "refs/tags/$TAG" | sed -n '1{s/[[:space:]].*//;p;}')

# Replace the content: the committed skill directory, byte for byte, and
# the release workflow.
find "$WORK/mirror" -mindepth 1 -maxdepth 1 ! -name .git -exec rm -rf {} +
git archive --format=tar HEAD "$SKILL" | tar -x -C "$WORK" -f -
cp -R "$WORK/$SKILL/." "$WORK/mirror/"
rm -rf "$WORK/skills"
mkdir -p "$WORK/mirror/.github/workflows"
git show "HEAD:$WORKFLOW" > "$WORK/mirror/.github/workflows/release.yml"

git -C "$WORK/mirror" add -A
if git -C "$WORK/mirror" diff --cached --quiet; then
    changed=0
else
    changed=1
fi

if [ -n "$published" ] && [ "$changed" -eq 1 ]; then
    echo "error: $NAME $VERSION is already published at $REMOTE with other content;" >&2
    echo "       move the skill's version (metadata version in SKILL.md) and rebuild" >&2
    exit 2
fi
if [ -z "$published" ] && [ "$changed" -eq 0 ] && git -C "$WORK/mirror" rev-parse -q --verify HEAD > /dev/null; then
    : # content current, tag missing: tag the existing head below
elif [ "$changed" -eq 0 ]; then
    echo "$NAME $VERSION: mirror is current at $REMOTE ($BRANCH, $TAG)"
    exit 1
fi

AUTHOR_NAME=$(git config user.name || true)
AUTHOR_MAIL=$(git config user.email || true)
if [ "$changed" -eq 1 ]; then
    git -C "$WORK/mirror" -c user.name="${AUTHOR_NAME:-pcd release}" \
        -c user.email="${AUTHOR_MAIL:-pcd@mailbox.org}" \
        commit -q -m "$NAME $VERSION, built from mge1512/pcd $SRC_SHORT

Source commit: $SRC
Mirror of skills/$NAME in https://github.com/mge1512/pcd. Do not edit here;
file issues in mge1512/pcd."
    git -C "$WORK/mirror" -c push.negotiate=false push -q "$REMOTE" "$BRANCH:$BRANCH"
fi
if [ -z "$published" ]; then
    git -C "$WORK/mirror" -c user.name="${AUTHOR_NAME:-pcd release}" \
        -c user.email="${AUTHOR_MAIL:-pcd@mailbox.org}" \
        tag -a "$TAG" -m "$NAME $VERSION, built from mge1512/pcd $SRC_SHORT"
    git -C "$WORK/mirror" -c push.negotiate=false push -q "$REMOTE" "refs/tags/$TAG"
fi
echo "$NAME $VERSION: pushed to $REMOTE ($BRANCH, $TAG) from $SRC_SHORT"
