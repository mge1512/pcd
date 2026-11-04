# Mode: security-review

Reviews the generated artefacts of one completed translation run for
security, as `prompts/security-reviewer.md` prescribes. The output is
an independent threat model, the deterministic gate findings, and
findings that name specification, hints or template text. Never a code
patch.

## Currency first

As in the review mode: `Spec-SHA256` of `TRANSLATION_REPORT.md` against
`sha256sum` of the specification, or `verify_spec_hash` with
`Includes:`. Stale: STOP with "pcd-translate: artifacts stale against
specification, refusing to review".

## Inputs

The audit bundle of the review mode plus the output of every
deterministic security gate the template invoked (static analysis,
dependency audit, SBOM check), as the prompt lists them. The META
fields that bear on the threat surface (`Safety-Level:`, and
`Privilege:`, `Network-Exposure:`, `Trust-Boundary:` where present) are
read from the preamble, which contains META. A missing gate output is a
stated limitation, not a pass.

## Procedure

Follow `prompts/security-reviewer.md`: hard prohibitions, operating
procedure, finding classes, output format. Bundles, not the
specification (`SKILL.md` section 4).

## Report

The prompt's report format, provenance lines of `SKILL.md` section 3
under `## Summary`. A finding that the specification lacks a security
requirement is written as the requirement the specification should
state: an INVARIANT, a PRECONDITION, a TYPE constraint. It is input for
the pcd-spec change mode. A finding against the template or a hints file
names the file and the line.
