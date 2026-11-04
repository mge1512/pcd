---
name: pcd-translate
description: Reading discipline and guard for translating a PCD specification into code using pre-sliced per-behavior bundles. Use whenever a session translates, re-translates, or continues translating any *.spec.md in a PCD repository - including single-behavior work and test-author or reviewer roles. Verifies pcd-slice output is present and current before any translation begins, and governs which files the session may read. Do NOT use for writing or reviewing specifications, or for repositories without pcd-slice.
metadata:
  version: "0.1.0"
  spec-schema: "0.4.0"
  pcd-lint-min: "0.4.0"
  pcd-slice-min: "0.1.1"
---

# pcd-translate

Translation reads bundles, never the specification. This skill is the
guard that makes that true and the contract for how bundles are read.
It assumes pcd-slice is installed from a signed package; it never
builds, fetches, or substitutes the tool.

## 1. Guard, before any translation token

Run, from the repository root:

    pcd-slice check spec=<tool>/spec/<name>.spec.md \
                    hints=<tool>/spec/<name>.<lang>.hints.md \
                    out=<tool>/spec/spec.d

Interpret strictly:

- exit 2, command not found, or version output without a `spec:` line:
  STOP. Report that pcd-slice is not installed correctly. Do not
  build it, do not download anything, do not translate from the
  specification instead.
- exit 1 with findings other than stale-output: STOP. The
  specification needs an edit (a tag, a binding, a heading) and a
  version row before translation is worth money. Report the findings
  verbatim and end the session.
- exit 1 with only stale-output findings, or spec.d absent: the
  pre-run was skipped or the sources moved. Recover in-session, once:

      pcd-slice slice spec=... hints=... out=<tool>/spec/spec.d

  then re-run check. If it is not clean now, STOP as above. Never
  translate against bundles check calls stale.
- exit 0: proceed.

## 2. Anchor

Record, before the first translation step, into the session log and
later into TRANSLATION_REPORT.md:

- the `spec:` hash from `pcd-slice version`
- the source and hints hashes from the first line of
  `spec.d/preamble.md`
- the SHA-256 of `spec.d/MANIFEST.tsv`

The report then pins what was actually read, not only what existed.

## 3. Read bundles, and only bundles

For work on behavior X:

    read spec.d/preamble.md      first, always, in full
    read spec.d/X.md             second

and nothing else from the specification set. Concretely:

- Never open `*.spec.md` during translation. Everything a behavior
  needs is in its bundle plus the preamble; if something seems
  missing, that is a pcd-slice or tagging defect - STOP and report
  it, do not route around it by reading the source.
- Keep the preamble as the first read of every model call, unmodified
  and in full, so its bytes stay a cacheable prefix.
- The bundle's Requires-Types line is a checklist, not an invitation:
  the definitions are already in the preamble.
- Cross-behavior context, where a step names another behavior, comes
  from that behavior's bundle - read it whole, or not at all.

## 4. Scope per session

One behavior's bundle per translation step. The test-author role
reads the same pair; the reviewer role may additionally read the
bundle's source-region line numbers from MANIFEST order but still
not the specification. Milestone scaffolding (module layout, build
files) reads preamble.md only.

## 5. On finishing

Re-run the guard from step 1. A clean check proves the sources did
not move under the session; a stale-output finding here means someone
edited the specification mid-translation - report it and mark every
artifact of this session suspect in the report.

## Failure phrases

Use these verbatim so the outer loop can grep for them:

- "pcd-translate: tool missing or unverified, refusing to translate"
- "pcd-translate: findings block translation:" followed by the lines
- "pcd-translate: bundles stale after re-slice, refusing"
- "pcd-translate: sources moved during session, artifacts suspect"
