# Mode: translator

Produces the implementation of a specification from its bundles: tests
first, then source, packaging, documentation and `TRANSLATION_REPORT.md`,
as `prompts/prompt.md` prescribes for its `translator` role. The
sections of `SKILL.md` apply throughout; this file states what is
specific to the mode.

## Inputs

The input directory contains:

- `spec.d/`, checked by the guard; the specification file itself may be
  present for the tools but is not read (`SKILL.md` section 4)
- the deployment template the specification names, read from this
  skill's `templates/`; the copy in the input directory, where one is
  staged, must hash to the same bytes, otherwise STOP and report
- the hints files in scope: this skill's `hints/` for the shipped
  tiers, the input directory for project-local tiers
- `ROLE.md` with `mode: translator`
- optionally, test-author's `independent_tests/<llm-name>/` and
  `TEST_REPORT.md`: present, the run is dual-LLM; absent, it is
  single-LLM, which the prompt supports
- optionally, `UPGRADE.md` (`SKILL.md` section 6)

Prior translation output is present only in an incremental run, and
then only what `UPGRADE.md` names.

## Procedure

Follow `prompts/prompt.md`, translator role, with the template's
`## EXECUTION` section as the recipe. That means tests before code, the
structural guard that no implementation file exists before the test
suite, the compile gate, and the delivery phases and resume logic of
the template. Where the prompt says to read the specification, read
the preamble and the bundle of the behavior at hand. One behavior per
translation step; milestone scaffolding from the preamble only.

## Report

`TRANSLATION_REPORT.md` as the prompt and the template define it, with
the provenance lines of `SKILL.md` section 3 added to the Translation
Inputs block. Ambiguities met in a bundle are listed under the
specification-ambiguities heading, each with the bundle name and the
text that was ambiguous. That list is input for the pcd-spec change
mode. The report never proposes code; it proposes specification, hints
or template text.
