# Mode: review

Reviews the generated artefacts of one completed translation run against
the specification, as `prompts/reviewer.md` prescribes. The output is
findings that name specification, hints or template text; never a code
patch. A review of the specification alone, before translation, is the
spec-review mode of pcd-spec, not this mode.

## Currency first

Before reading any artefact, confirm that the code under review is the
translation of the specification at hand:

- Without `Includes:`, `Spec-SHA256` in `TRANSLATION_REPORT.md` equals
  `sha256sum` of the specification file.
- With `Includes:`, the merged hash from the MCP tool `verify_spec_hash`
  reports `current`.

Otherwise STOP with "pcd-translate: artifacts stale against
specification, refusing to review": findings against a stale
translation would be charged to the wrong specification.

## Inputs

The audit bundle the prompt lists: the specification (for the tools;
the reviewer reads bundles), the template, all hints files in scope,
and the translator prompt as the translator saw it. Further the
generated artefacts, `TRANSLATION_REPORT.md`, and `independent_tests/`
where produced. A missing input is stated under the review
limitations, never inferred.

## Procedure

Follow `prompts/reviewer.md`: its hard prohibitions, operating
procedure, finding classes and output format. Read the specification
through its bundles (`SKILL.md` section 4) and cite a location by
bundle name and section heading; the maintainer maps that to the
specification.

## Report

The prompt's report format, with the provenance lines of `SKILL.md`
section 3 under `## Summary`. Each finding names the bundle, the
artefact and the text the specification, hints or template should say.
The recommendation follows the prompt's decision rules and is addressed
to the maintainer, who applies accepted findings through the pcd-spec
change mode, one unit per run.
