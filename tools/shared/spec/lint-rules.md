# PCD Lint Rules (Shared)

## META
Deployment:   none
Version:      0.4.2
Spec-Schema:  0.4.0
Author:       Matthias G. Eckermann <pcd@mailbox.org>
License:      CC-BY-4.0
Verification: none
Safety-Level: QM

This specification is a composition target (Deployment: none). It is not
translated into an implementation on its own; it is included by host specs
via the META field

```
Includes: ../../shared/spec/lint-rules.md
```

and its TYPES, BEHAVIORs and INVARIANTS are merged into each host's
effective specification before translation (see doc/spec-composition.md).

It is the single source of truth for the PCD structural lint rules
(RULE-01 through RULE-21). The two current consumers are:

- `pcd-lint` (Deployment: cli-tool) - the command-line linter.
- `mcp-server-pcd` (Deployment: mcp-server) - the `lint_content` and
  `lint_file` MCP tools.

Both consumers previously described these rules independently, producing
two divergent rule implementations with no mechanism to keep them in sync.
Moving the rules here makes a single edit propagate to every consumer on
re-translation: each host's merged-spec hash changes, both regenerate, and
no drift is possible.

Rule-level acceptance EXAMPLES are intentionally not kept in this shared
spec. They are maintained in each consuming host's EXAMPLES section, because
they assert host-specific output (pcd-lint's CLI summary line and exit code;
mcp-server-pcd's structured MCP JSON response) rather than the host-neutral
rule semantics defined here. The rule semantics in this file are the
normative contract; each host's examples are the acceptance tests for that
host's projection of them.

## TYPES

```
Section := string where matches("^## [A-Z][A-Z0-9-]+")

MetaField := {
  key:   string where non-empty,
  value: string where non-empty
}

SPDXIdentifier := string where matches SPDX license identifier list
// Reference: https://spdx.org/licenses/
// Examples: Apache-2.0, MIT, GPL-2.0-only, LGPL-2.1-or-later
// Compound expressions permitted: Apache-2.0 OR MIT
// pcd-lint validates against the current SPDX license list embedded at build time

SemanticVersion := string where matches "^[0-9]+\.[0-9]+\.[0-9]+$"
// MAJOR.MINOR.PATCH - no pre-release suffixes in v1. Governs Spec-Schema
// (RULE-02d); the spec's own Version field uses SpecVersion below.

SpecVersion := string where matches
  "^([0-9]+\.[0-9]+\.[0-9]+|[0-9]{4}\.[0-9]{2}\.[0-9]{2}\.[0-9]{2})$"
// Spec META Version: semantic MAJOR.MINOR.PATCH (e.g. 0.6.10) or dated
// YYYY.MM.DD.VV with a two-digit per-day counter (e.g. 2026.06.09.03).
// Accepted per maintainer decision D-8 (2026-06-10).

DeploymentTemplate := one_of(
  "wasm" | "ebpf" | "kernel-module" | "verified-library" |
  "cli-tool" | "gui-tool" | "cloud-native" | "backend-service" |
  "library-c-abi" | "enterprise-software" | "academic" |
  "python-tool" | "enhance-existing" | "manual" | "template" |
  "mcp-server" | "project-manifest"
)
// "crypto-library" is retired as of 0.3.6. Use "verified-library" instead.
// "verified-library" covers all safety- and security-critical C-ABI libraries.
// "python-tool" is QM safety level only; Verification: none mandatory.
// "template" is used exclusively in deployment template definition files
// (*.template.md). A spec using Deployment: template is a template
// specification, not a translatable component.
// "project-manifest" added in v0.3.8 for multi-component projects.
// "mcp-server" added in v0.3.8 for MCP server components.

BehaviorConstraint := required | supported | forbidden
// Classifies a BEHAVIOR block. Default is `required` when absent.
// A `forbidden` behavior must include a `reason:` annotation.
// Validated by RULE-13.

MilestoneStatus := pending | active | failed | released
// Pipeline state for ## MILESTONE sections.
// Transitions: pending → active → released (pass) or failed (fail).
// Exactly one milestone may be active at any time (RULE-15).
// Status is managed by the agent pipeline, not by the spec author.

Severity := Error | Warning

Diagnostic := {
  severity: Severity,
  section:  string,        // which section triggered the diagnostic; e.g. "META", "BEHAVIOR", "structure"
  line:     u32 where line > 0,   // 1-based; 1 for file-level diagnostics
  message:  string,
  rule:     string         // the rule that produced this diagnostic; e.g. "RULE-01", "RULE-08", "RULE-17"
}
// This Diagnostic is the union of the fields previously declared
// independently by pcd-lint and mcp-server-pcd. The `rule` field carries
// the producing rule identifier; mcp-server-pcd surfaces it in its JSON
// output, and pcd-lint may populate it without rendering it in the CLI
// diagnostic line (the CLI line format is defined in pcd-lint's DEPLOYMENT
// section and is unchanged). Each host wraps a List<Diagnostic> in its own
// result type (pcd-lint: LintResult with exit_code; mcp-server-pcd:
// LintResult with valid/errors/warnings counts).
```

Note: Multiple BEHAVIOR and BEHAVIOR/INTERNAL sections are permitted.
Each describes a distinct operation or internal rule of the component.
All BEHAVIOR sections share the TYPES, INVARIANTS, and EXAMPLES
sections of the merged specification. BEHAVIOR/INTERNAL sections describe
implementation logic not directly exposed to the user; they are
validated with identical structural rules to BEHAVIOR sections.

---

## BEHAVIOR/INTERNAL: code-fence-tracking
Constraint: required

Tracks whether the parser is currently inside a code-fenced block
and suppresses all structural detection while inside one.

A depth counter (not a boolean toggle) is required to handle nested
fences correctly — for example, a GIVEN block that contains a fenced
code example which itself contains a fenced inner block.

STEPS:
1. Initialise fenceDepth = 0.
2. For each line L in the file:
   a. If TrimSpace(L) begins with ``` or ~~~:
      if fenceDepth = 0: set fenceDepth = 1; skip to next line.
      else: set fenceDepth = fenceDepth - 1; skip to next line.
      (the fence marker line itself is always skipped)
   b. If fenceDepth > 0: skip L entirely — no pattern matching.
   c. If fenceDepth = 0: pass L to all structural detection rules.

---

## BEHAVIOR: lint-validation-rules
Constraint: required

Defines the ordered set of structural rules applied during lint.
All rules are evaluated; lint does not stop at first error.

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

REQUIRED_SECTIONS :=
  [ "## META", "## TYPES", "## BEHAVIOR", "## PRECONDITIONS",
    "## POSTCONDITIONS", "## INVARIANTS", "## EXAMPLES" ]

For each section S in REQUIRED_SECTIONS:
  if S not present in file:
    emit Error, section="structure", line=1,
      message="Missing required section: {S}"
// line=1 is the canonical value for missing-section diagnostics.
// The section does not exist, so no line can be identified;
// line=1 signals a file-level structural error to the caller.

Note: "## BEHAVIOR" is satisfied by the presence of one or more
BEHAVIOR sections. The following BEHAVIOR variants are all recognised
and valid:
  - "## BEHAVIOR: <n>"          user-facing operation
  - "## BEHAVIOR/INTERNAL: <n>" internal implementation logic,
                                   not directly user-facing

Multiple BEHAVIOR and BEHAVIOR/INTERNAL sections are permitted and
may be freely mixed. Section headers must appear at the start of a
line. Case is significant (BEHAVIOR uppercase required).
BEHAVIOR/INTERNAL sections are validated with identical structural
rules to BEHAVIOR sections.

### RULE-02: META fields present and non-empty

REQUIRED_META_FIELDS :=
  [ "Deployment", "Verification", "Safety-Level",
    "Version", "Spec-Schema", "License" ]
// Note: Author is required (at least one) but uses repeating-key pattern.
// See RULE-02b below.

For each field F in REQUIRED_META_FIELDS:
  if F not present in META section:
    emit Error, section="META",
      message="Missing required META field: {F}"
  if value of F is empty:
    emit Error, section="META",
      message="META field {F} has empty value"

### RULE-02b: Author field

if no "Author:" line present in META section:
  emit Error, section="META",
    message="Missing required META field: Author (at least one Author: line required)"

// Multiple Author: lines are permitted and collected as a list.
// Each Author: value must be non-empty.
For each Author: line A in META section:
  if value of A is empty:
    emit Error, section="META",
      message="Author: field has empty value"

### RULE-02c: Version format

Let V = value of META field "Version"
if V does not match SpecVersion pattern
  "^([0-9]+\.[0-9]+\.[0-9]+|[0-9]{4}\.[0-9]{2}\.[0-9]{2}\.[0-9]{2})$":
  emit Error, section="META",
    message="Version '{V}' is not a valid version string. \
             Accepted formats: MAJOR.MINOR.PATCH (e.g. 0.1.0) \
             or YYYY.MM.DD.VV (e.g. 2026.06.09.03)"

### RULE-02d: Spec-Schema version

Let S = value of META field "Spec-Schema"
if S does not match SemanticVersion pattern "^[0-9]+\.[0-9]+\.[0-9]+$":
  emit Error, section="META",
    message="Spec-Schema '{S}' is not valid semantic versioning. \
             Required format: MAJOR.MINOR.PATCH (e.g. 0.1.0)"

// v2 note: validate Spec-Schema against list of known schema versions
// and warn if spec was written against an older schema than current.

### RULE-02e: License SPDX validation

Let L = value of META field "License"

if L is not a valid SPDX license identifier or compound expression:
  emit Error, section="META",
    message="License '{L}' is not a valid SPDX identifier. \
             See https://spdx.org/licenses/ for valid identifiers. \
             Compound expressions permitted (e.g. Apache-2.0 OR MIT)."

// pcd-lint embeds the SPDX license list at build time.
// The embedded list version is reported in pcd-lint version output.

### RULE-03: Deployment template resolves

Let D = value of META field "Deployment"

if D = "crypto-library":
  emit Error, section="META", line=1,
    message="Deployment 'crypto-library' was retired in 0.3.6. \
             Use 'verified-library' instead. \
             verified-library covers all safety- and security-critical \
             C-ABI libraries including cryptographic primitives."

if D not in DeploymentTemplate:
  emit Error, section="META",
    message="Unknown deployment template: '{D}'. \
             Run 'pcd-lint list-templates' to see valid values."

if D = "enhance-existing":
  if META field "Language" not present:
    emit Error, section="META",
      message="Deployment 'enhance-existing' requires META field 'Language'"
  if value of "Language" is empty:
    emit Error, section="META",
      message="META field 'Language' has empty value"

if D = "manual":
  if META field "Target" not present:
    emit Error, section="META",
      message="Deployment 'manual' requires META field 'Target' \
               (no template available for language resolution)"

if D = "python-tool":
  let SL = value of META field "Safety-Level"
  if SL ≠ "QM":
    emit Error, section="META",
      message="Deployment 'python-tool' requires Safety-Level: QM. \
               Python is not suitable for safety-critical components."
  let V = value of META field "Verification"
  if V ≠ "none":
    emit Error, section="META",
      message="Deployment 'python-tool' requires Verification: none. \
               No formal verification path exists for Python."

if D = "verified-library":
  let SL = value of META field "Safety-Level"
  if SL = "QM":
    emit Warning, section="META",
      message="Deployment 'verified-library' with Safety-Level: QM is unusual. \
               verified-library is intended for safety- or security-critical \
               components. Consider using library-c-abi for general-purpose libraries."

### RULE-04: Deprecated META fields

if META field "Target" is present AND D ≠ "manual":
  emit Warning, section="META",
    message="META field 'Target' is deprecated since v0.3.0. \
             Target language is derived from the deployment template. \
             Remove 'Target', or switch to Deployment: manual \
             if explicit language control is required."

if META field "Domain" is present:
  emit Warning, section="META",
    message="META field 'Domain' is deprecated since v0.3.0. \
             Use 'Deployment' instead."

### RULE-05: Verification field value

KNOWN_VERIFICATION_VALUES := [ "none", "lean4", "fstar", "dafny", "custom" ]

Let V = value of META field "Verification"

if V not in KNOWN_VERIFICATION_VALUES:
  emit Warning, section="META",
    message="Unknown verification value: '{V}'. \
             Known values: none, lean4, fstar, dafny, custom. \
             Custom verification backends are permitted; \
             verify the value is intentional."

### RULE-06: EXAMPLES section structure

The ## EXAMPLES section must contain at least one example block.

An example block begins with a heading line matching:
  `### EXAMPLE: <name>`
at column 0. The heading marker is exactly three pound signs, one
space, the literal text `EXAMPLE:`, one space, and the example name.

An example block consists of:
  - a heading line matching "^### EXAMPLE: " (example name declaration)
  - a GIVEN: marker
  - at least one WHEN:/THEN: pair appearing after GIVEN:
  - WHEN: and THEN: must alternate: each WHEN: must be followed by
    its matching THEN: before the next WHEN: or end of block
  (Multi-pass examples with multiple WHEN/THEN pairs are valid — v0.3.12+)

if no example block found:
  emit Error, section="EXAMPLES",
    message="EXAMPLES section contains no example blocks. \
             Each example requires ### EXAMPLE: heading and \
             GIVEN:, WHEN:, THEN: markers."

For each example block E:
  if E missing "GIVEN:":
    emit Error, section="EXAMPLES",
      message="Example '{n}' missing GIVEN: marker"
  if E missing at least one "WHEN:":
    emit Error, section="EXAMPLES",
      message="Example '{n}' missing WHEN: marker"
  if E missing at least one "THEN:":
    emit Error, section="EXAMPLES",
      message="Example '{n}' missing THEN: marker"
  for each WHEN: marker W in E (in order):
    if W is not immediately followed (before next WHEN: or end of block)
       by a THEN: marker:
      emit Error, section="EXAMPLES",
        message="Example '{n}' has WHEN: without a matching THEN:"

Header form errors:

if a line matching "^EXAMPLE: " (flat form) appears inside the ## EXAMPLES section:
  emit Error, section="EXAMPLES",
    message="Example header must be '### EXAMPLE: <name>'. \
             Flat 'EXAMPLE: <name>' is no longer accepted (changed in v0.4.0). \
             See: doc/spec-composition.md or RULE-06."

if a line matching "^## EXAMPLE: " or "^#### EXAMPLE: " appears inside the ## EXAMPLES section:
  emit Error, section="EXAMPLES",
    message="Example header must be at heading level 3 (three pound signs). \
             Found heading level {n} for '{name}'."

### RULE-07: EXAMPLES minimum content

Block boundaries are defined as follows:
  GIVEN block  := lines strictly between GIVEN: and first WHEN: marker
  WHEN block   := lines strictly between a WHEN: marker and its matching THEN: marker
  THEN block   := lines after a THEN: marker until one of:
                    - next WHEN: marker at start of line (multi-pass)
                    - next "### EXAMPLE: " heading at start of line
                    - next ## or ### heading at start of line (non-EXAMPLE)
                    - end of file
  A block is empty if it contains zero non-whitespace lines.
  A marker line itself (GIVEN:, WHEN:, THEN:) is not content.

For each example block E:
  if GIVEN block is empty:
    emit Warning, section="EXAMPLES",
      message="Example '{n}' has empty GIVEN block"
  if WHEN block is empty:
    emit Warning, section="EXAMPLES",
      message="Example '{n}' has empty WHEN block"
  if THEN block is empty:
    emit Warning, section="EXAMPLES",
      message="Example '{n}' has empty THEN block"

### RULE-08: BEHAVIOR blocks must contain STEPS (v0.3.12+)

For each BEHAVIOR or BEHAVIOR/INTERNAL section B:
  if B does not contain a line matching "^STEPS:":
    emit Error, section=B,
      message="BEHAVIOR '{n}' is missing required STEPS: block. \
               Every BEHAVIOR must include ordered, imperative STEPS."

### RULE-09: INVARIANTS entries should carry observable/implementation tags (v0.3.12+)

For each entry line L in the ## INVARIANTS section:
  // An entry line is a non-empty, non-heading line that is not a separator.
  if L does not begin with "- [observable]" AND L does not begin with "- [implementation]":
    emit Warning, section="INVARIANTS",
      message="Invariant entry missing tag. \
               Prefix with [observable] or [implementation] for audit utility."

### RULE-10: Negative-path EXAMPLE required for BEHAVIOR with error exits (v0.3.13+)

For each BEHAVIOR section B:
  let error_exits = lines in B's STEPS block matching "→" (error exit notation)
  if error_exits is non-empty:
    // Collect EXAMPLES that reference this BEHAVIOR (by name or by being the
    // sole BEHAVIOR in the spec). A negative-path EXAMPLE is one whose THEN:
    // block contains at least one of: "Err(", "error", "exit_code = 1",
    // "exit_code = 2", "stderr contains", or a declared ERROR code from B.
    let negative_examples = EXAMPLES referencing B whose THEN block matches
                            negative-path pattern
    if negative_examples is empty:
      emit Error, section=B,
        message="BEHAVIOR '{n}' has error exits in STEPS but no negative-path \
                 EXAMPLE. Add at least one EXAMPLE whose THEN: verifies an \
                 error outcome."

// Note: for specs with a single BEHAVIOR, all EXAMPLES are considered
// to reference that BEHAVIOR. For multi-BEHAVIOR specs, association is
// by name matching between EXAMPLE WHEN: text and BEHAVIOR name.

### RULE-11: TOOLCHAIN-CONSTRAINTS section structure (v0.3.13+)

if ## TOOLCHAIN-CONSTRAINTS section is present:
  For each entry line L in the section:
    if L declares a constraint value other than "required" or "forbidden":
      emit Warning, section="TOOLCHAIN-CONSTRAINTS",
        message="TOOLCHAIN-CONSTRAINTS entry uses unknown constraint value. \
                 Valid values: required, forbidden."
// The section is optional. Its absence is not an error.
// Structural validation is minimal in v0.3.13; semantic validation deferred to v0.4.0.

### RULE-12: Cross-section consistency (v0.3.13+, partial)

**12a — Identifier consistency (warning):**
  Collect all method names declared in ## INTERFACES sections
    (lines matching pattern: "  <MethodName>(")
  For each method name M:
    if M appears in any BEHAVIOR STEPS block in a modified form
       (e.g. "transport.Connect" where M is "Connect"):
      emit Warning, section="BEHAVIOR",
        message="Identifier '{M}' declared in INTERFACES but referenced as \
                 '{variant}' in BEHAVIOR STEPS. Use the declared name verbatim."

**12b — Type name consistency (error):**
  Collect all type names declared in ## TYPES section
    (lines matching "^<TypeName> :=")
  For each type name T:
    if T is redefined (assigned with :=) in any BEHAVIOR section:
      emit Error, section="BEHAVIOR",
        message="Type '{T}' declared in TYPES is redefined in BEHAVIOR. \
                 Types must be declared in TYPES only."

**12c — File name consistency (warning):**
  Collect all file names in ## DELIVERABLES COMPONENT entries
  Collect all file names referenced in ## BEHAVIOR/INTERNAL sections
  For each file name F referenced in BEHAVIOR/INTERNAL but absent from DELIVERABLES:
    emit Warning, section="BEHAVIOR/INTERNAL",
      message="File '{F}' referenced in BEHAVIOR/INTERNAL is not declared \
               in DELIVERABLES. Add a COMPONENT entry or remove the reference."

// State-machine and endpoint semantic consistency deferred to v0.4.0.

### RULE-13: Constraint: field value on BEHAVIOR headers (v0.3.13+)

VALID_CONSTRAINTS := [ "required", "supported", "forbidden" ]

For each BEHAVIOR or BEHAVIOR/INTERNAL section B:
  if B has a line matching "^Constraint:":
    let C = value of Constraint: field
    if C not in VALID_CONSTRAINTS:
      emit Error, section=B,
        message="BEHAVIOR '{n}' has invalid Constraint: value '{C}'. \
                 Valid values: required, supported, forbidden."
    if C = "forbidden":
      if B does not contain a line matching "^  reason:":
        emit Warning, section=B,
          message="BEHAVIOR '{n}' is Constraint: forbidden but has no reason: annotation."
  // Absence of Constraint: field is valid; default is `required`.

---

### RULE-14: EXECUTION section required in deployment templates (v0.3.16+)

This rule applies only when the file under validation has
`Deployment: template` in its META section (i.e. it is a deployment
template, not a component spec).

```
if spec.META["Deployment"] = "template":
  if spec does not contain a top-level section "## EXECUTION":
    emit Warning, line=1,
      message="Deployment template is missing ## EXECUTION section.
               Translators cannot determine delivery phases without it.
               Add ## EXECUTION or declare 'EXECUTION: none' in META
               if this template intentionally has no execution recipe."
  else:
    let exec = content of ## EXECUTION section
    if exec does not contain "### Delivery phases":
      emit Warning, section="EXECUTION",
        message="## EXECUTION section has no '### Delivery phases' subsection."
    if exec does not contain "### Compile gate" and
       exec does not contain "COMPILE-GATE: none":
      emit Warning, section="EXECUTION",
        message="## EXECUTION section has no '### Compile gate' subsection
                 and does not declare 'COMPILE-GATE: none'.
                 Translators will not know how to verify compilation."
    if exec does not contain "### Resume logic":
      emit Warning, section="EXECUTION",
        message="## EXECUTION section has no '### Resume logic' subsection."
```

Exception: if the template META contains `EXECUTION: none`, all RULE-14
checks are skipped. Use this for templates that produce no compiled output
(e.g. `project-manifest`, `python-tool`).

---

### RULE-15: MILESTONE section structure and single-active constraint (v0.3.21+)

Applies only when the spec contains one or more `## MILESTONE:` sections.
MILESTONE sections are optional — their absence is not an error.

```
for each ## MILESTONE: section M in the spec:

  // Structure check
  if M does not contain "Included BEHAVIORs:":
    emit Error, section=M,
      message="MILESTONE '{n}' is missing required 'Included BEHAVIORs:' field."

  if M does not contain "Deferred BEHAVIORs:" AND M does not have Scaffold: true:
    emit Error, section=M,
      message="MILESTONE '{n}' is missing required 'Deferred BEHAVIORs:' field.
               Omit this field only when Scaffold: true (scaffold milestones
               have no deferred BEHAVIORs by definition)."

  if M does not contain "Acceptance criteria:":
    emit Warning, section=M,
      message="MILESTONE '{n}' has no 'Acceptance criteria:' field.
               Translators and agents cannot verify completion."

  // Status check
  if M does not contain a line matching "^Status:":
    emit Warning, section=M,
      message="MILESTONE '{n}' has no Status: field.
               Expected: pending | active | failed | released."
  else:
    let S = value of Status: field
    if S not in [ "pending", "active", "failed", "released" ]:
      emit Error, section=M,
        message="MILESTONE '{n}' has invalid Status: value '{S}'.
                 Valid values: pending, active, failed, released."

  // Scaffold field check (if present)
  if M contains a line matching "^Scaffold:":
    let SC = value of Scaffold: field
    if SC not in [ "true", "false" ]:
      emit Error, section=M,
        message="MILESTONE '{n}' has invalid Scaffold: value '{SC}'.
                 Valid values: true, false."

// Single-active constraint
let active_milestones = [ M for M in milestones if M.Status = "active" ]
if len(active_milestones) > 1:
  emit Error, section="structure", line=1,
    message="More than one MILESTONE has Status: active.
             Exactly one milestone may be active at a time."
```

---

### RULE-16: MILESTONE BEHAVIOR names exist in spec (v0.3.21+)

Applies only when the spec contains one or more `## MILESTONE:` sections.

```
let all_behavior_names = set of all BEHAVIOR and BEHAVIOR/INTERNAL names in the spec

for each ## MILESTONE: section M:
  for each name N listed under "Included BEHAVIORs:" in M:
    if N not in all_behavior_names:
      emit Error, section=M,
        message="MILESTONE '{milestone}' lists BEHAVIOR '{N}' under \
                 Included BEHAVIORs but no such BEHAVIOR exists in the spec."

  for each name N listed under "Deferred BEHAVIORs:" in M:
    if N not in all_behavior_names:
      emit Error, section=M,
        message="MILESTONE '{milestone}' lists BEHAVIOR '{N}' under \
                 Deferred BEHAVIORs but no such BEHAVIOR exists in the spec."
```

---

### RULE-17: Scaffold milestone ordering and uniqueness (v0.3.21+)

Applies only when the spec contains one or more `## MILESTONE:` sections
and at least one has `Scaffold: true`.

```
let scaffold_milestones = [ M for M in milestones if M.Scaffold = "true" ]

if len(scaffold_milestones) > 1:
  emit Error, section="structure", line=1,
    message="More than one MILESTONE has Scaffold: true.
             At most one scaffold milestone is permitted per spec."

if len(scaffold_milestones) = 1:
  let SM = scaffold_milestones[0]
  let first_milestone = milestones[0]   // first in document order
  if SM ≠ first_milestone:
    emit Error, section=SM,
      message="Scaffold milestone '{n}' must appear first in the spec
               (lowest version number / earliest in document order).
               Later milestones depend on the scaffold foundation."
```

---

### RULE-18: Spec hash presence in TRANSLATION_REPORT (v0.3.22+)

Applies when a `TRANSLATION_REPORT.md` file is present adjacent to the spec
being linted (i.e. in the same directory or in a `code/` subdirectory).

```
if TRANSLATION_REPORT.md exists adjacent to spec:
  if TRANSLATION_REPORT.md does not contain line matching /^Spec-SHA256:\s+[0-9a-f]{64}/:
    emit Warning, section="report", line=1,
      message="TRANSLATION_REPORT.md is missing Spec-SHA256: field.
               Every translation run must record the SHA256 of the merged
               spec it was produced from (host plus resolved Includes;
               equals the host file hash when no Includes are declared)."

  if Spec-SHA256 field present:
    let recorded_hash = value of Spec-SHA256 field
    let current_hash  = sha256(merged spec text of <specname>; equals
                        sha256 of the host file when no Includes are declared)
    if recorded_hash ≠ current_hash:
      emit Warning, section="report", line=1,
        message="Spec has changed since last translation run.
                 Recorded hash: {recorded_hash}
                 Current hash:  {current_hash}
                 Regeneration may be required. Run: pcd change-impact
                 or use assess_change_impact via mcp-server-pcd."
```

Note: RULE-18 emits Warnings only, never Errors. The spec itself may be
valid; the mismatch indicates a process concern, not a structural defect.
RULE-18 is only evaluated when 
`check-report=true` is set, to avoid false positives in spec-only workflows.

---

### RULE-19: Includes path resolves

Constraint: required (when host spec declares Includes)

For every `Includes:` directive in the host spec's META section, the
referenced path must resolve to a readable file relative to the host
spec's location.

INPUTS:
- HOST_PATH: absolute path of the spec being linted
- INCLUDES_VALUES: list of `Includes:` values from META

STEPS:
1. For each value in INCLUDES_VALUES:
   a. Compute the absolute path by resolving the value relative to
      directory of HOST_PATH.
   b. If the file does not exist or is not readable, emit:
      ERROR  {file}:{line}  [META]  Includes path does not resolve: {value}

MECHANISM: path resolution uses standard filesystem semantics. Symbolic
links are followed. The check is on existence and readability, not on
spec validity of the included file.

---

### RULE-20: Merged spec has no name collisions

Constraint: required (when host spec declares Includes)

After resolving all `Includes:` directives recursively and merging the
included specs into the host, the resulting merged spec must contain no
duplicate names within any of: TYPES, BEHAVIORs, INTERFACES, EXAMPLES.

INPUTS:
- HOST_SPEC: the host spec content (parsed)
- INCLUDED_SPECS: the recursively-resolved included specs (parsed)

STEPS:
1. Construct the merged TYPE set: collect all TYPE definitions from
   each included spec in declaration order, then from the host.
2. For each TYPE name that appears more than once, emit:
   ERROR  {host-file}  [META]  Name collision after merge: TYPE {name}
   appears in both {first-origin} and {second-origin}
3. Repeat for BEHAVIORs, INTERFACES, EXAMPLES.

MECHANISM: there is no implicit precedence. Collisions are spec-author
errors and must be resolved by renaming or restructuring.

---

### RULE-21: Inclusion graph is acyclic and well-formed

Constraint: required (when host spec declares Includes)

The transitive closure of `Includes:` references from a host spec must
form a directed acyclic graph. Additionally, included specs must not
declare orchestration-only sections that belong to host components.

INPUTS:
- HOST_PATH
- INCLUDED_SPECS: full transitive closure with provenance

STEPS:
1. Detect cycles. Construct the directed graph where each node is a spec
   file and an edge A → B exists when A's `Includes:` references B.
   Perform a DFS from HOST_PATH. If any back-edge is found, emit:
   ERROR  {file}  [META]  Inclusion cycle: {cycle-path}
   where {cycle-path} is the cycle in form A → B → C → A.

2. For each included spec, verify it does not contain a `## MILESTONE:`
   section. If found, emit:
   ERROR  {included-file}  [structure]  Included spec must not declare
   MILESTONE section: {included-file}

3. For each included spec, verify it does not contain a `## DEPLOYMENT`
   section. If found, emit:
   ERROR  {included-file}  [structure]  Included spec must not declare
   DEPLOYMENT section: {included-file}

MECHANISM: cycles are detected by standard DFS with a visited set. There
is no depth cap; practical specs are expected to have inclusion depth of
1 or 2.

---

## INVARIANTS

- [observable]      all rules are evaluated; rule processing does not
  short-circuit on the first Error. A failure in one rule does not prevent
  subsequent rules from running (see BEHAVIOR: lint-validation-rules)
- [observable]      every Diagnostic carries a severity of Error or Warning,
  a section, a message, and a 1-based line number greater than zero
- [observable]      rule evaluation is read-only with respect to the file
  under validation: it does not modify the file, make network calls, or read
  environment variables for behaviour control

---

## CHANGELOG

- 2026.06.10.02 - RULE-02c accepts dated versions (maintainer decision D-8,
  consistency-check task T-29): spec META Version may be semantic
  MAJOR.MINOR.PATCH or dated YYYY.MM.DD.VV. New SpecVersion type carries the
  alternation; SemanticVersion stays semver-only and continues to govern
  Spec-Schema (RULE-02d unchanged). The VERSION rows of all ten deployment
  templates were updated in the same batch.
- 2026.06.10.01 - RULE-18 message and hash definition aligned with the
  merged-spec hash semantics (consistency-check task T-24): the recorded
  and recomputed hash is the SHA256 of the merged spec text (host plus
  recursively resolved Includes), equal to the host file hash when the
  spec declares no Includes. No rule logic change beyond the recomputation
  definition; diagnostics text updated accordingly.
- 2026.06.09.01 - Initial extraction. RULE-01 through RULE-21, the
  lint-validation-rules orchestration BEHAVIOR, the code-fence-tracking
  BEHAVIOR/INTERNAL, and the rule-domain TYPES (Section, MetaField,
  SPDXIdentifier, SemanticVersion, DeploymentTemplate, BehaviorConstraint,
  MilestoneStatus, Severity, Diagnostic) moved verbatim from
  pcd-lint.spec.md (which carried the canonical, complete v0.4.0 rule set).
  Diagnostic harmonised to the union of the pcd-lint and mcp-server-pcd
  field sets (adds the `rule` field). Consumed by pcd-lint and
  mcp-server-pcd via Includes.
