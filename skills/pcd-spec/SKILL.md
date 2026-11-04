---
name: pcd-spec
description: Writing and changing PCD specifications (*.spec.md). Use whenever a session writes a new specification from an interview or from existing notes; derives one from existing source code; changes one BEHAVIOR, TYPE, INTERFACE or INVARIANT of an existing specification; designs or revises MILESTONE sections; or reviews a specification before it is translated. Produces specification text, the Version move, the CHANGELOG.md entry beside the specification and the pull-request body. Do NOT use for turning a specification into code, tests, implementation findings or a verdict between implementations; that is pcd-translate.
metadata:
  version: "0.3.0"
  spec-schema: "0.4.0"
  pcd-lint-min: "0.4.0"
  pcd-slice-min: "0.1.1"
---

# pcd-spec

This skill writes specification bytes. It never writes implementation
code, never edits generated code, and never sets a MILESTONE `Status:`
line (the pipeline maintains that line). The specification is the
durable artefact; everything this skill produces is either the
specification, its changelog, or text for the pull request that
delivers them.

## 1. Guard, before anything else

Run, from the skill root (the directory of this file):

    sh scripts/check-tools.sh

- exit 2: STOP. The skill's files do not match its manifest, or the
  front matter is unreadable. Say "pcd-spec: skill files damaged,
  refusing" and end the session. Do not repair, download or rebuild
  anything.
- exit 1: CONTINUE. One or both tools are missing, old or on another
  schema. Note which, for the hand-over (section 5). When the session
  has an MCP server with the tools `lint_content` or `lint_file`, use
  those for linting instead; `assess_change_impact` likewise. When
  neither CLI nor MCP server is reachable, proceed without linting and
  say so in the hand-over.
- exit 0: CONTINUE, and lint with the CLI at the points the modes name.

Never install a tool, never fetch one, never substitute a hand-written
check for pcd-lint. A specification written without a lint run is a
draft; the pull-request gate runs pcd-lint and is the authority.

## 2. Modes

Select the mode from the request, or from the argument the skill was
invoked with. Ask when two modes fit.

| Mode | When | Read | Writes |
|---|---|---|---|
| interview | A new component, from a conversation or from existing notes, mails or design documents | `modes/interview-procedure.md` | a new `<name>.spec.md`, `CHANGELOG.md` |
| reverse | A new specification derived from existing source code | `modes/reverse-procedure.md` | a new `<name>.spec.md`, `CHANGELOG.md` |
| change | One change to an existing specification | `modes/change.md`, and `modes/change-impact-procedure.md` for the impact assessment | the edited `.spec.md`, `CHANGELOG.md`, the impact assessment |
| milestone | MILESTONE sections for a specification that has none, or a revision of them | `modes/milestone.md` | the edited `.spec.md`, `CHANGELOG.md` |
| spec-review | Findings on a specification before translation money is spent | `modes/spec-review.md` | a findings list; the specification is not edited |

Read the mode file in full before the first question or edit. The
`*-procedure.md` files were written to be pasted as system prompts and
still work that way. Read them as the procedure to follow, and apply the
rules in section 3 on top of them where the two differ.

## 3. Rules that apply in every mode

Format and naming:

- The specification format is defined by `references/lint-rules.md`
  (structure, RULE-01 to RULE-21) and `references/types-table-rules.md`
  (tabular TYPE grammar, RULE-22 to RULE-25). pcd-lint decides; the
  rule text explains a finding. Never paraphrase a rule into a
  different check.
- Specification files are named `<name>.spec.md`. Templates are
  `*.template.md`, hints `*.hints.md`; neither is a specification.
- `Spec-Schema:` in META is the `spec-schema` value of this skill's
  front matter, whatever value a procedure file's example shows.
- A specification names no target language, library, file path or
  function of an implementation. The deployment template resolves the
  language; library knowledge belongs in hints files. Language-specific
  content found in a specification is a finding, not a style choice.

Content that must not be in a specification:

- No license text. The `License:` field contains an SPDX identifier and
  nothing else.
- No run-time data as a deliverable: no data file, no directory, no
  compiled-in fallback. Fixtures belong in the test tree.
- No copy of a contract. JSON Schema contracts live in `contracts/` and
  are referenced, never reconstructed from prose.
- No MILESTONE `Status:` value other than `pending` on a milestone this
  skill creates. Advancing a milestone is a pipeline operation.

Versions and history:

- Bytes move, Version moves. Every edit to a `.spec.md`, however small,
  moves `Version:` in META. The grammar is MAJOR.MINOR.PATCH or
  YYYY.MM.DD.VV (RULE-02c); keep the grammar the specification already
  uses. With semver: PATCH for a change inside one BEHAVIOR's STEPS,
  ERRORS or EXAMPLES. MINOR for a new or removed BEHAVIOR, for a TYPE,
  INTERFACE or INVARIANT change, and for any other change a translator
  has to act on. MAJOR only when the maintainer says so.
- The specification's history is kept in `CHANGELOG.md` beside it,
  newest entry first, one entry per Version:

      - YYYY-MM-DD: Version X.Y.Z. <what changed and why, two or three sentences>

  The changelog is not part of the specification, is not read by the
  translator and is not covered by the specification hash. Entries are
  never edited after the Version they describe has been committed; a
  correction is a new entry that names the earlier one. Create the file
  when it does not exist.
- DELTA, where a specification has one, is the forward-looking work
  list; the changelog is the backward-looking record. Keep them apart.

Scope of a run:

- One run, one unit: one new specification, or one change to one
  BEHAVIOR, one TYPE, one INTERFACE or one INVARIANT, or one milestone
  design. When a request bundles several units, say so and ask which
  one comes first. Each unit becomes its own pull request.
- Ask one question at a time, wait for the answer, and confirm each
  phase before moving on, as the prompt files prescribe. Never invent a
  value the expert did not give: mark it `[?]` and ask.

## 4. Lint and slice checks

Where a mode says "lint":

    pcd-lint strict=true <path>.spec.md

Zero errors required before hand-over; report warnings as they are. Where
the specification has `Includes:`, lint the host file; pcd-lint resolves
the includes.

Where a mode says "slice check", and pcd-slice is available:

    pcd-slice check spec=<path>.spec.md

A finding here (a missing tag, a binding, a heading) means the
translation skill will refuse the specification later; fix it now.

Without an engine, skip both and record it (section 5). Never claim a
lint result that was not produced by the tool.

## 5. Hand-over

End every run with this block, verbatim labels, one line each:

    Specification: <path> Version <old> -> <new>   (or: new, Version <new>)
    Changelog:     <path to CHANGELOG.md> entry added
    Lint:          pcd-lint <version> <cli|mcp>: clean | <n> errors, <m> warnings | not run (no engine)
    Slice check:   pcd-slice <version>: clean | <findings> | not run
    Change impact: <full-regeneration | incremental | not applicable>, <one-line reason>
    Open points:   <none | list>

The `Lint:` and `Slice check:` lines go into the hand-over message and
into the pull-request body. They never go into the specification bytes.

Pull-request body, for the maintainer to paste:

    <component>: <one line, what changed>

    Version: <old> -> <new>
    Changelog: <the new entry, verbatim>
    Change impact: <recommendation and primary factor>
    Lint: <as above>
    Slice check: <as above>

The title follows the repository convention `component: brief
description`. Reference the issue when one was opened first.

## Failure phrases

Use these verbatim so the outer loop can grep for them:

- "pcd-spec: skill files damaged, refusing"
- "pcd-spec: lint not run, no engine reachable"
- "pcd-spec: lint findings:" followed by the lines
- "pcd-spec: request bundles several units, asking which first"
