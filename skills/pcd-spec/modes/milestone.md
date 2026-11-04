# Mode: milestone

MILESTONE sections for a specification that has none, or a revision of
the pending ones. The rules of `SKILL.md` section 3 apply; this file
adds what is specific to milestones.

## When milestones are needed

- More than ten BEHAVIORs: recommend milestones, so that each
  translation pass fits a translator's context.
- Ten or fewer: ask whether the expert wants a working subset first;
  otherwise a single pass needs no milestones.
- A specification whose translation keeps failing on size, or on
  consistency between passes. Milestones with a scaffold are the
  remedy, and this mode designs them.

## Procedure

Follow PHASE 9 of `modes/interview-procedure.md` (propose groupings,
confirm, write the sections) and the milestone checks of PHASE 10.
Apply them to the existing specification instead of a freshly written
one. Read the complete specification first: every BEHAVIOR name, every
TYPE, and the DEPENDENCIES.

The structure the rules enforce:

- The scaffold milestone comes first in document order and is the only
  one with `Scaffold: true` (RULE-17). It establishes package layout,
  interfaces and build files; everything after it builds on those
  decisions, so its scope is chosen for stability, not for size.
- Three to eight BEHAVIORs per milestone after the scaffold.
  Group by dependency, so that a milestone's BEHAVIORs only use TYPEs
  and BEHAVIORs of itself or of earlier milestones.
- Every name under `Included BEHAVIORs:` and `Deferred BEHAVIORs:`
  exists in the specification (RULE-16). Every BEHAVIOR is included in
  exactly one milestone of the sequence.
- `Status: pending` on every milestone this skill writes. Exactly one
  milestone may be `active` at a time (RULE-15), and the pipeline sets
  that; this skill never writes `active`, `released` or `failed`.
- `Acceptance criteria:` are shell commands that exit 0 on success, one
  per line, runnable against the built artefact of that milestone.
  "Works correctly" is not a criterion; `./tool version | grep -q
  "^tool "` is.
- `Hints-file:` names the milestones hints file of the deployment
  template and language, `<template>.<language>.milestones.hints.md`,
  where one exists in this skill's `hints/`. Otherwise omit the field;
  never invent a file.

## Revising milestones

- A pending milestone may be regrouped, renamed or split freely; the
  Version moves and the changelog entry says what moved.
- An active milestone's `Included BEHAVIORs:` is left alone; a new
  BEHAVIOR goes to a pending milestone or under `Deferred BEHAVIORs:`
  of the active one.
- A released milestone is frozen. If the expert asks to change one,
  explain that the path is a new milestone sequence under full
  regeneration. Record the request in the hand-over under
  `Open points:` and change nothing.

## Hand over

As `SKILL.md` section 5. `Change impact:` is "full-regeneration, new
milestone sequence" when the specification already has generated code.
It is "not applicable" when the specification has not been translated
yet.
