# Mode: tiebreak

Decides what to do when two translators produced two implementations of
one specification, as `prompts/tiebreaker.md` prescribes. A tie-break
is a vote on the specification, not on the code. The two translators
disagreed because the specification, the hints, the template or the
prompt allowed them to; the output is the edit that removes that
latitude. The mode never names a winning implementation.

## Currency first

Both `TRANSLATION_REPORT.md` files record the same `Spec-SHA256`, and
it equals `sha256sum` of the specification (or `verify_spec_hash`
reports current with `Includes:`). Otherwise the two translations are
of different specifications and there is nothing to break: STOP with
"pcd-translate: artifacts stale against specification, refusing to
review".

## Inputs

As the prompt lists them: the specification (bundles), the template,
the hints and the translator prompt. Then translation A and B, each
with artefacts, report, tests and translator identity. Then the
cross-validation matrix, each suite run against each implementation. A
missing matrix is a stated limitation; the mode does not run the suites
itself.

## Procedure

Follow `prompts/tiebreaker.md`: hard prohibitions, operating procedure,
finding classes, output format. Bundles, not the specification
(`SKILL.md` section 4).

## Report

The prompt's report format, provenance lines of `SKILL.md` section 3
under `## Summary`, one set per translation. The cumulative
specification edit is the deliverable: specification, hints or template
text, each item tied to the divergence it removes, ordered as the
prompt's decision rules say. The maintainer applies it through the
pcd-spec change mode and re-translates from the corrected inputs; that
translation replaces both contested implementations.
