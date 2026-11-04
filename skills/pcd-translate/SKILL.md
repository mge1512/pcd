---
name: pcd-translate
description: Turning a PCD specification into code, tests, findings or a verdict, from pre-sliced per-behavior bundles. Use whenever a session translates, re-translates or continues translating any *.spec.md in a PCD repository; writes the independent test suite for one; reviews generated code against its specification, functionally or for security; or decides between two translations of the same specification. Verifies the bundles before any work and governs which files the session may read. Do NOT use for writing, changing or reviewing a specification itself; that is pcd-spec.
metadata:
  version: "0.3.0"
  spec-schema: "0.4.0"
  pcd-lint-min: "0.4.0"
  pcd-slice-min: "0.1.1"
---

# pcd-translate

Translation reads bundles, never the specification. This skill is the
guard that makes that true and the contract for how bundles are read.
It is also the entry point for every role that works from a
specification. It never writes specification bytes: the `*.spec.md`,
the `CHANGELOG.md` beside it and the MILESTONE `Status:` line are out
of its reach, and a pull request from a translation branch that changes
a specification fails CI. Whatever a role learns about the
specification goes to the maintainer as a finding that names spec,
hints or template text.

## 1. Guard, before any token of work

Step 1, from the skill root (the directory of this file):

    sh scripts/check-tools.sh

- exit 2: STOP. "pcd-translate: skill files damaged, refusing". Do not
  repair, download or rebuild anything.
- exit 0: both tools are present and acceptable. Go to step 2.
- exit 1: a tool is missing, old, unverified or on another schema. Do
  not build it, do not download anything, do not translate from the
  specification instead. Two ways forward:
  - The operator states in the session that pcd-lint and pcd-slice were
    run on this specification before the upload (the manual override of
    record D-11). Then verify, with `sha256sum` only:
    1. `<tool>/spec/spec.d/preamble.md` and `spec.d/MANIFEST.tsv` exist.
    2. `source-sha256=` on the first line of `preamble.md` equals the
       `sha256sum` of the specification file, and `hints-sha256=` equals
       the hashes of the hints files named in `hints=`, in order.
    3. Every bundle listed in `MANIFEST.tsv` hashes to its listed value.
    All equal: proceed, and record the override (section 3). Any
    mismatch: STOP, "pcd-translate: bundles stale after override check,
    refusing". The override replaces the guard's tool run, never the
    pull-request gate, which runs both tools again.
  - No such statement: STOP, "pcd-translate: tool missing or
    unverified, refusing to translate".

Step 2, with the tools present, from the repository root:

    pcd-slice check spec=<tool>/spec/<name>.spec.md \
                    hints=<tool>/spec/<name>.<lang>.hints.md \
                    out=<tool>/spec/spec.d

Interpret strictly:

- exit 2, or version output without a `spec:` line: STOP, "pcd-translate:
  tool missing or unverified, refusing to translate".
- exit 1 with findings other than stale-output: STOP. The specification
  needs an edit (a tag, a binding, a heading) and a Version move before
  translation is worth money. Report the findings verbatim, as input for
  the pcd-spec change mode, and end the session.
- exit 1 with only stale-output findings, or `spec.d` absent: the
  pre-run was skipped or the sources moved. Recover in-session, once:

      pcd-slice slice spec=... hints=... out=<tool>/spec/spec.d

  then re-run check. If it is not clean now, STOP as above. Never
  translate against bundles check calls stale.
- exit 0: proceed.

An MCP server with `lint_file`, `lint_content` and `verify_spec_hash`
covers linting and hash currency for the review modes; slicing needs
the `pcd-slice` CLI or the override until the server slices itself.

## 2. Modes

The mode comes from `ROLE.md` in the input directory, else from the
argument the skill was invoked with, else ask. `ROLE.md`:

    mode: translator
    llm-name: <lowercase, hyphen-separated model name>

`mode:` is one of `translator`, `test-author`, `review`,
`security-review`, `tiebreak`. The translation procedure reads the
first two; this skill reads all five. `llm-name` names the output directory
of the tests and the identity in every report.

| Mode | Produces | Read | Allowed inputs |
|---|---|---|---|
| translator | implementation, packaging, documentation, `TRANSLATION_REPORT.md` | `modes/translator.md`, then `modes/translation-procedure.md` | bundles, template, hints, `ROLE.md`; optionally test-author's output; optionally `UPGRADE.md` |
| test-author | `independent_tests/<llm-name>/`, `TEST_REPORT.md` | `modes/test-author.md`, then `modes/translation-procedure.md` | bundles, template, hints, `ROLE.md`, nothing else |
| review | findings against the specification, hints or template | `modes/review.md`, then `modes/review-procedure.md` | the audit bundle of one completed run |
| security-review | findings and a threat model | `modes/security-review.md`, then `modes/security-review-procedure.md` | the audit bundle plus deterministic gate output |
| tiebreak | a vote on the specification, never a winning implementation | `modes/tiebreak.md`, then `modes/tiebreak-procedure.md` | two audit bundles and the cross-validation matrix |

Read the mode file in full, then the procedure file it names. The
procedure files were written to be pasted as system prompts and still
work that way; read them as the procedure to follow. Where a procedure
says "read the specification", this skill's section 4 applies instead:
read the bundles.

## 3. Anchor and provenance

Record, before the first step of the mode and later in the report:

- the `spec:` hash from `pcd-slice version`, or `none (override)`
- the `source-sha256=` and `hints-sha256=` values from the first line of
  `spec.d/preamble.md`
- the SHA-256 of `spec.d/MANIFEST.tsv`
- `Skill-SHA256: pcd-translate/SKILL.md <sha256 of this file>`
- `Skill-Manifest-SHA256: <sha256 of MANIFEST.sha256 at the skill root>`
- `Toolchain: pcd-lint <version> <cli|mcp>` and
  `Toolchain: pcd-slice <version> cli`, one line per tool used, or
  `Toolchain: none (override)`
- `Guard-Override: <who declared what, date; verified by sha256sum;
  tools not run in session>`, or `Guard-Override: none`

In `TRANSLATION_REPORT.md` and `TEST_REPORT.md` these are further
labelled lines of the Translation Inputs (provenance) block, after the
lines the template requires. In a review or tie-break report they go
under `## Summary`. The report then pins what was read, by which tools,
not only what existed.

## 4. Read bundles, and only bundles

For work on behavior X:

    read spec.d/preamble.md      first, always, in full
    read spec.d/X.md             second

and nothing else from the specification set. Concretely:

- Never open `*.spec.md` in any mode. Everything a behavior needs is in
  its bundle plus the preamble. If something seems missing, that is a
  pcd-slice or tagging defect: STOP and report it. Do not route around
  it by reading the source.
- Keep the preamble as the first read of every model call, unmodified
  and in full, so its bytes stay a cacheable prefix.
- The bundle's Requires-Types line is a checklist, not an invitation:
  the definitions are already in the preamble.
- Cross-behavior context, where a step names another behavior, comes
  from that behavior's bundle: read it whole, or not at all.
- Templates, hints and rule documents are read from this skill's own
  `templates/`, `hints/` and `references/` directories, which the
  manifest covers. Project-local hints (`.pcd/hints/`, decisions hints
  beside the specification) are read from the input directory.

## 5. Scope per session

One behavior's bundle per translation step. The test-author mode reads
the same pair. The review modes read every bundle they need, whole,
and cite a location by bundle name and section heading, but still not
the specification. Milestone scaffolding (module layout, build files)
reads `preamble.md` only.

## 6. Incremental runs

`UPGRADE.md` beside the specification is the upgrade brief written by
the pcd-spec change mode when the change-impact assessment said
"incremental". It is committed with the specification and is not part
of the specification hash. When it is present in the input directory,
the translator mode runs incrementally. It reads the brief after the
preamble, confines its edits to the files and functions the brief
lists, and leaves everything else byte-identical. The report records
`Upgrade-Brief-SHA256: UPGRADE.md <hash>`. When the brief is absent, a
run against existing code is a full regeneration into a clean worktree;
existing code is never exposed to a fresh run.

## 7. On finishing

Re-run the guard from step 2 of section 1 (or the override check). A
clean check proves the sources did not move under the session. A
stale-output finding here means someone edited the specification
mid-run: report it and mark every artifact of this session suspect in
the report. Then write the report of the mode, with the lines of
section 3.

## Failure phrases

Use these verbatim so the outer loop can grep for them:

- "pcd-translate: skill files damaged, refusing"
- "pcd-translate: tool missing or unverified, refusing to translate"
- "pcd-translate: bundles stale after override check, refusing"
- "pcd-translate: findings block translation:" followed by the lines
- "pcd-translate: bundles stale after re-slice, refusing"
- "pcd-translate: sources moved during session, artifacts suspect"
- "pcd-translate: test-author input contaminated, refusing"
- "pcd-translate: artifacts stale against specification, refusing to review"
