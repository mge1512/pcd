# PCD Lint Rules (Shared Spec)

## META
Deployment:   none
Version:      0.4.0
Spec-Schema:  0.4.0
Author:       Matthias G. Eckermann <pcd@mailbox.org>
License:      CC-BY-4.0
Verification: none
Safety-Level: QM

This spec defines the structural and semantic validation rules applied
to PCD specification files. It is consumed via the v0.4.0 `Includes:`
mechanism by host components that need to perform lint validation —
currently `pcd-lint` (as a standalone CLI tool) and `mcp-server-pcd` (as
an MCP tool callable from any LLM client).

It is not a deployable component on its own: `Deployment: none` signals
that this spec exists solely as an inclusion target. A translator that
encounters a host spec with `Includes: lint-rules.md` merges this file's
contents into the host's effective spec; no implementation of
`lint-rules.md` alone is ever produced.

---

## TYPES

(Move TYPE definitions from `tools/pcd-lint/spec/pcd-lint.md` that
relate to rule processing into this section. The relevant types from
the v0.3.22 pcd-lint spec are:

- `Severity := Error | Warning`
- `Section := structure | META | EXAMPLES | BEHAVIOR | INVARIANTS | TOOLCHAIN-CONSTRAINTS | MILESTONE`
- `RuleId := RULE-01 | RULE-02 | RULE-02b | ... | RULE-21`
- `Diagnostic := { Severity, Section, Line, Rule, Message }`
- `LintResult := { Diagnostics: list of Diagnostic, ExitCode: int }`
- Any helper types referenced inside the rule bodies (FenceState,
  ParsedSpec, etc.) — keep these where they live in the host spec if
  they are host-orchestration types, move them here if they are
  rule-internal.

The principle: rule-specific types live here; CLI / MCP-tool types stay
in the respective host spec.)

---

## BEHAVIOR: lint-validation-rules
Constraint: required

Defines the ordered set of structural rules applied during lint.
All rules are evaluated; lint does not stop at first error.

(Move the existing `BEHAVIOR: lint-validation-rules` content here from
`tools/pcd-lint/spec/pcd-lint.md`, with the STEPS list extended through
RULE-21 per `pcd-lint-rules-addendum.md`.)

STEPS:
1. Apply RULE-01 (required sections present).
2. Apply RULE-02 through RULE-02e (META fields).
3. Apply RULE-03 (deployment template resolves).
4. Apply RULE-04 (deprecated META fields).
5. Apply RULE-05 (Verification field value).
6. Apply RULE-06 (EXAMPLES section structure, including multi-pass).
7. Apply RULE-07 (EXAMPLES minimum content).
8. Apply RULE-08 (BEHAVIOR blocks contain STEPS).
9. Apply RULE-09 (INVARIANTS entries carry observable/implementation tags).
10. Apply RULE-10 (negative-path EXAMPLE required for BEHAVIOR with error exits).
11. Apply RULE-11 (TOOLCHAIN-CONSTRAINTS section structure, if present).
12. Apply RULE-12 (cross-section consistency: identifiers, types, file names).
13. Apply RULE-13 (Constraint: field value on BEHAVIOR headers).
14. Apply RULE-14 (EXECUTION section present in deployment templates).
15. Apply RULE-15 (MILESTONE section structure and single-active constraint, if present).
16. Apply RULE-16 (MILESTONE BEHAVIOR names exist in spec, if present).
17. Apply RULE-17 (scaffold milestone ordering and uniqueness, if present).
18. Apply RULE-18 (spec hash presence in TRANSLATION_REPORT, if check-report=true).
19. Apply RULE-19 (Includes path resolves, if Includes present).
20. Apply RULE-20 (merged spec has no name collisions, if Includes present).
21. Apply RULE-21 (inclusion graph is acyclic and well-formed, if Includes present).

MECHANISM: rules are independent; a failure in one rule does not prevent
subsequent rules from running. All diagnostics are collected before output.

### RULE-01: Required sections present

(Move RULE-01 content here verbatim from pcd-lint.md.)

### RULE-02: META fields present and non-empty

(Move RULE-02 content here verbatim from pcd-lint.md.)

### RULE-02b: Author field
### RULE-02c: Version format
### RULE-02d: Spec-Schema version
### RULE-02e: License SPDX validation

### RULE-03: Deployment template resolves
### RULE-04: Deprecated META fields
### RULE-05: Verification field value
### RULE-06: EXAMPLES section structure
### RULE-07: EXAMPLES minimum content
### RULE-08: BEHAVIOR blocks must contain STEPS
### RULE-09: INVARIANTS entries should carry observable/implementation tags
### RULE-10: Negative-path EXAMPLE required for BEHAVIOR with error exits
### RULE-11: TOOLCHAIN-CONSTRAINTS section structure
### RULE-12: Cross-section consistency
### RULE-13: Constraint: field value on BEHAVIOR headers
### RULE-14: EXECUTION section required in deployment templates
### RULE-15: MILESTONE section structure and single-active constraint
### RULE-16: MILESTONE BEHAVIOR names exist in spec
### RULE-17: Scaffold milestone ordering and uniqueness
### RULE-18: Spec hash presence in TRANSLATION_REPORT

(Move each rule's content here verbatim from pcd-lint.md, retaining the
exact format including the rule-specific INPUTS, STEPS, MECHANISM
annotations.)

### RULE-19: Includes path resolves
### RULE-20: Merged spec has no name collisions
### RULE-21: Inclusion graph is acyclic and well-formed

(Insert per pcd-lint-rules-addendum.md.)

---

## PRECONDITIONS

(Move preconditions that apply to rule processing — input spec is
readable, parser state, etc. — from pcd-lint.md here. Preconditions
relating to CLI invocation, environment variables, etc. stay in the
host spec.)

---

## POSTCONDITIONS

(Move postconditions on the lint result. Exit-code mapping stays in the
host spec since it's CLI- or MCP-tool-specific.)

---

## INVARIANTS

(Move rule-processing invariants:
- Every emitted diagnostic references a valid RuleId.
- Diagnostic severity is one of {Error, Warning}.
- Rule processing is idempotent: running rules twice produces the same
  diagnostics.
- No diagnostic is suppressed for performance reasons.

Host-orchestration invariants like exit-code-on-error stay in the host.)

---

## EXAMPLES

(Move ALL examples that test rule behaviour here:
- All `valid_minimal_spec`, `missing_section`, `invalid_spdx_license`,
  etc. examples that probe what the rules do.
- Each EXAMPLE's GIVEN/WHEN/THEN format unchanged.

Host-specific examples stay in the host:
- CLI-argument-parsing examples
- MCP-tool-invocation examples
- File-not-found error handling
- list-templates output format)

Add the new examples for RULE-19/20/21 per pcd-lint-rules-addendum.md.
