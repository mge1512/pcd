# pcd-lint.fixtures.hints.md

Fixture-completeness hints for translation and test-authoring runs of
`pcd-lint.spec.md`.

Version: 2026.06.10.01
Origin:  moved verbatim from `prompts/prompt.md` (universal translator
         prompt), which must stay generic; project-specific fixture
         knowledge lives here (consistency-check task T-27,
         2026-06-10). Read this file before writing any test fixture
         for pcd-lint. It is a translation input and is recorded in the
         report's Translation Inputs provenance (one further labelled
         line, e.g. `Library-Hints-SHA256:` or a dedicated
         `Fixture-Hints-SHA256:` label).

---

## Structurally complete fixtures for pcd-lint

A "structurally complete" fixture for pcd-lint contains every section
the linted spec schema requires, with content sufficient to pass all
structural rules the test does not intend to trigger:

- META with all required fields (Deployment, Version,
  Spec-Schema, Author, License, Verification, Safety-Level)
- TYPES (may be empty body)
- BEHAVIOR sections with non-empty STEPS lists. A `## BEHAVIOR: foo`
  header with nothing under it fires RULE-08; if your test
  expects a specific warning and the binary instead emits the
  warning *plus* a RULE-08 error, the test fails for the wrong
  reason.
- PRECONDITIONS (may be empty body)
- POSTCONDITIONS (may be empty body)
- INVARIANTS with at least one entry carrying an
  `[observable]` or `[implementation]` tag (RULE-09 warns
  otherwise)
- EXAMPLES with at least one negative-path EXAMPLE if any
  BEHAVIOR in the fixture has error exits in its STEPS list
  (RULE-10)

Before writing the test, mentally run the binary against your fixture
and predict which rules fire. The expected diagnostic set is what the
test asserts on. If your fixture would trigger structural rules you
don't intend to test, complete the structure first.

---

## Changelog

- 2026.06.10.01 - Initial file. Content moved verbatim from the
  universal translator prompt (task T-27); the prompt now carries only
  the generic fixture-completeness rules and a pointer to
  project-scoped hints files of this kind.
