



# pcd-lint

## META
Deployment:  cli-tool
Version:     0.4.0
Spec-Schema: 0.4.0
Author:      Matthias G. Eckermann <pcd@mailbox.org>
License:     GPL-2.0-only
Verification: none
Safety-Level: QM

---

## TYPES

```
SpecFile := path where file_exists AND readable AND extension = ".md"

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
// MAJOR.MINOR.PATCH — no pre-release suffixes in v1

TemplatePath := path
// The directory where deployment template files are found.
// All four directories are searched; later entries take precedence (last-wins).
// This means project-local templates override user templates, which override
// system templates, which override the vendor default.
//
// Search order (ascending precedence):
//   1. /usr/share/pcd/templates/    vendor default (pcd-templates package)
//   2. /etc/pcd/templates/          system administrator additions
//   3. ~/.config/pcd/templates/     user additions
//   4. ./.pcd/templates/            project-local
//
// Directories that do not exist are silently skipped.
// v1 supports Linux only. macOS and Windows paths deferred to v2.
//
// Implementation: templateSearchDirs() returns the list of existing dirs
// in ascending precedence order. findTemplateFile(name) returns the path
// of the last match found across all dirs (not the first). This ensures
// project-local templates override installed ones.

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
  section:  string,        // which section triggered the diagnostic
  message:  string,
  line:     u32 where line > 0
}

ExitCode := 0 | 1 | 2
// 0 = valid (no errors; no warnings when strict=true)
// 1 = invalid (at least one Error; or strict=true and at least one Warning)
// 2 = invocation error (bad arguments, file not found, unreadable file)

LintResult := {
  file:        SpecFile,
  diagnostics: List<Diagnostic>,
  exit_code:   ExitCode
}
```

Note: Multiple BEHAVIOR and BEHAVIOR/INTERNAL sections are permitted.
Each describes a distinct operation or internal rule of the component.
All BEHAVIOR sections share the TYPES, INVARIANTS, and EXAMPLES
sections of this specification. BEHAVIOR/INTERNAL sections describe
implementation logic not directly exposed to the user; they are
validated with identical structural rules to BEHAVIOR sections.

---

## BEHAVIOR: lint
Constraint: required

The primary operation. Validates a specification file against the
structural rules defined in this specification.

INPUTS:
```
file:   SpecFile
strict: bool     // strict=true treats warnings as errors; default false
check_report: bool // check-report=true evaluates RULE-18; default false
```

OUTPUTS:
```
result: LintResult
```

PRECONDITIONS:
- file exists and is readable
- file has `.md` extension
- strict is a valid boolean (true | false)
- check_report is a valid boolean (true | false)

STEPS:
1. Verify file has `.md` extension; on failure → exit 2 with
   "error: file must have .md extension: {path}".
2. Open and read file; on failure → exit 2 with
   "error: cannot open file: {path}".
3. Apply RULE-01 through RULE-17 in order; collect all diagnostics.
   Rules are not short-circuited — all rules run regardless of earlier errors.
4. Sort diagnostics by line number (monotonically non-decreasing).
5. Write each diagnostic to stderr in the defined format.
6. Compute exit_code: 1 if any Error present, or (strict=true AND any Warning); else 0.
7. Write summary line to stdout in the defined format.
8. Exit with exit_code.

POSTCONDITIONS:
- result.file = file
- result.exit_code = 0 iff result.diagnostics contains no Error
  AND (strict = false OR result.diagnostics contains no Warning)
- result.exit_code = 1 iff result.diagnostics contains at least one Error,
  OR (strict = true AND result.diagnostics contains at least one Warning)
- diagnostics are written to stderr, one line per diagnostic
- summary line is written to stdout
- order of diagnostics follows order of appearance in file
- input file is not modified

SIDE-EFFECTS:
- stderr: diagnostic lines (errors and warnings), if any
- stdout: summary line (always emitted, see DEPLOYMENT for format)
- no network calls
- no environment variable reads for behaviour control

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

## BEHAVIOR: list-templates
Constraint: required

Prints all known deployment templates with their resolved default
target language. Useful for discovering valid Deployment: values.

INPUTS:
```
none
```

OUTPUTS:
```
stdout: list of template names with default language annotations
```

PRECONDITIONS:
- none

STEPS:
1. Call templateSearchDirs() to obtain the ordered list of existing template
   directories (ascending precedence; later entries override earlier).
   MECHANISM: templateSearchDirs() checks each of the four candidate paths
   (/usr/share/pcd/templates, /etc/pcd/templates, ~/.config/pcd/templates,
   ./.pcd/templates) and includes only those that exist as directories.
2. For each template T in defined order:
   a. Call findTemplateFile(T) to locate the companion `{T}.template.md`.
      MECHANISM: findTemplateFile(name) iterates templateSearchDirs() and
      records the path of each match found; returns the last match (highest
      precedence). Returns "" if no match in any directory.
   b. If found: call readDefaultLanguage(path) to extract the default language.
      MECHANISM: readDefaultLanguage(path) reads the file, locates the
      ## TEMPLATE-TABLE section, and returns the value from the first row
      where key=LANGUAGE and constraint=default. Returns "" if not found.
      If readDefaultLanguage returns "": annotation = "(installed)".
      If found: annotation = returned language string.
      If findTemplateFile returned "": annotation = "(template file not found)".
   c. For special values (enhance-existing, manual, template, project-manifest):
      use the fixed annotation defined in POSTCONDITIONS regardless of
      whether a companion file exists.
3. Write one line per template to stdout in format: "{T}  →  {annotation}".
4. Exit 0.

POSTCONDITIONS:
- exit_code = 0 always
- stdout contains exactly 17 lines, one per known DeploymentTemplate value
- each line format: "<template-name>  →  <default-language>"
- for enhance-existing: "<template-name>  →  (declare Language: in META)"
- for manual:           "<template-name>  →  (declare Target: in META)"
- for template:         "<template-name>  →  (template definition file, not translatable)"
- for project-manifest: "<template-name>  →  (architect artifact, no code generated)"
- nothing written to stderr

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
if V does not match SemanticVersion pattern "^[0-9]+\.[0-9]+\.[0-9]+$":
  emit Error, section="META",
    message="Version '{V}' is not valid semantic versioning. \
             Required format: MAJOR.MINOR.PATCH (e.g. 0.1.0)"

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
               Every translation run must record the SHA256 of the spec
               it was produced from. Run: sha256sum <specname>.md"

  if Spec-SHA256 field present:
    let recorded_hash = value of Spec-SHA256 field
    let current_hash  = sha256(<specname>.md)
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

## PRECONDITIONS

- For lint: file argument must be provided
- For lint: file must exist and be readable by the current process
- For lint: file must have .md extension
- For lint: if file does not have .md extension:
    exit 2, write to stderr: "error: file must have .md extension: {path}"
- For list-templates: no file argument required
- key=value arguments must use recognised keys (see DEPLOYMENT)
- unrecognised key=value pairs: exit 2, message to stderr

---

## POSTCONDITIONS

- pcd-lint does not modify any file on disk
- pcd-lint does not make network calls
- pcd-lint does not read environment variables for behaviour control
- exit code is always 0, 1, or 2; no other values
- on file-not-found or unreadable:
    exit 2, write to stderr: "error: cannot open file: {path}"
- on file without .md extension:
    exit 2, write to stderr: "error: file must have .md extension: {path}"
- on missing file argument (without list-templates):
    exit 2, write to stderr: usage line (see DEPLOYMENT)
- on unrecognised key=value argument:
    exit 2, write to stderr: "error: unrecognised option: {key}"

---

## INVARIANTS

- [observable]      pcd-lint is idempotent — running it twice on the same file
  produces identical output and identical exit code
- [observable]      all Error diagnostics produce exit_code ≥ 1
- [observable]      Warnings alone never produce exit_code = 1 unless strict=true
- [observable]      exit_code = 1 with only Warnings requires strict=true
- [observable]      exit_code = 2 indicates invocation error only,
  never a lint result
- [observable]      diagnostic line numbers are monotonically non-decreasing
  within a result
- [observable]      pcd-lint never produces exit_code = 0 when any Error
  diagnostic is present, regardless of strict value
- [observable]      stderr receives diagnostics; stdout receives summary and
  list-templates output; these streams are never swapped

---

## EXAMPLES

### EXAMPLE: valid_minimal_spec
GIVEN:
  file contains all required sections: META, TYPES, BEHAVIOR,
    PRECONDITIONS, POSTCONDITIONS, INVARIANTS, EXAMPLES
  META contains:
    Deployment:   cli-tool
    Version:      0.1.0
    Spec-Schema:  0.1.0
    Author:       Jane Example <jane@example.org>
    License:      Apache-2.0
    Verification: none
    Safety-Level: QM
  EXAMPLES contains one complete block with EXAMPLE:, GIVEN:, WHEN:, THEN:
  invocation: pcd-lint spec.md
WHEN:
  result = lint(file, strict=false)
THEN:
  stderr = (empty)
  stdout = "✓ spec.md: valid"
  exit_code = 0

### EXAMPLE: multiple_authors_valid
GIVEN:
  META contains:
    Deployment:   cli-tool
    Version:      0.1.0
    Spec-Schema:  0.1.0
    Author:       Jane Example <jane@example.org>
    Author:       John Example <john@example.org>
    License:      Apache-2.0
    Verification: none
    Safety-Level: QM
  all other sections valid
  invocation: pcd-lint spec.md
WHEN:
  result = lint(file, strict=false)
THEN:
  stderr = (empty)
  stdout = "✓ spec.md: valid"
  exit_code = 0

### EXAMPLE: invalid_spdx_license
GIVEN:
  META contains:
    License: MIT License
  all other META fields valid
  invocation: pcd-lint spec.md
WHEN:
  result = lint(file, strict=false)
THEN:
  stderr contains one diagnostic:
    severity = Error
    message contains "License 'MIT License' is not a valid SPDX identifier"
    message contains "https://spdx.org/licenses/"
  exit_code = 1

### EXAMPLE: invalid_version_format
GIVEN:
  META contains:
    Version: 1.0
  all other META fields valid
  invocation: pcd-lint spec.md
WHEN:
  result = lint(file, strict=false)
THEN:
  stderr contains one diagnostic:
    severity = Error
    message contains "Version '1.0' is not valid semantic versioning"
  exit_code = 1

### EXAMPLE: missing_author
GIVEN:
  META contains all required fields except no Author: line present
  invocation: pcd-lint spec.md
WHEN:
  result = lint(file, strict=false)
THEN:
  stderr contains one diagnostic:
    severity = Error
    message = "Missing required META field: Author (at least one Author: line required)"
  exit_code = 1

### EXAMPLE: missing_section
GIVEN:
  file is missing the ## INVARIANTS section
  all other required sections present and valid
  invocation: pcd-lint spec.md
WHEN:
  result = lint(file, strict=false)
THEN:
  stderr contains one line:
    "ERROR  spec.md:{line}  [structure]  Missing required section: ## INVARIANTS"
  stdout = "✗ spec.md: 1 error(s), 0 warning(s)"
  exit_code = 1

### EXAMPLE: unknown_deployment_template
GIVEN:
  file is valid except META contains: Deployment: serverless
  invocation: pcd-lint spec.md
WHEN:
  result = lint(file, strict=false)
THEN:
  stderr contains one diagnostic:
    severity = Error
    message contains "Unknown deployment template: 'serverless'"
  exit_code = 1

### EXAMPLE: deprecated_target_field_permissive
GIVEN:
  file is valid with META containing:
    Deployment: backend-service
    Target: Go
    Verification: none
    Safety-Level: QM
  invocation: pcd-lint spec.md
WHEN:
  result = lint(file, strict=false)
THEN:
  stderr contains one diagnostic:
    severity = Warning
    message contains "META field 'Target' is deprecated since v0.3.0"
  stdout = "✓ spec.md: valid (1 warning(s))"
  exit_code = 0

### EXAMPLE: deprecated_target_field_strict
GIVEN:
  same file as deprecated_target_field_permissive
  invocation: pcd-lint strict=true spec.md
WHEN:
  result = lint(file, strict=true)
THEN:
  stderr contains one diagnostic:
    severity = Warning
    message contains "META field 'Target' is deprecated since v0.3.0"
  stdout = "✗ spec.md: 0 error(s), 1 warning(s) [strict mode]"
  exit_code = 1

### EXAMPLE: enhance_existing_missing_language
GIVEN:
  file META contains:
    Deployment: enhance-existing
    Verification: none
    Safety-Level: QM
  META does not contain a Language: field
  invocation: pcd-lint spec.md
WHEN:
  result = lint(file, strict=false)
THEN:
  stderr contains one diagnostic:
    severity = Error
    message = "Deployment 'enhance-existing' requires META field 'Language'"
  exit_code = 1

### EXAMPLE: empty_given_block_permissive
GIVEN:
  file is structurally valid, but EXAMPLES contains a block with an empty GIVEN section:
  ```markdown
  EXAMPLE: foo
  GIVEN:

  WHEN:
    result = foo()
  THEN:
    result = Ok
  ```
  invocation: pcd-lint spec.md
WHEN:
  result = lint(file, strict=false)
THEN:
  stderr contains one diagnostic:
    severity = Warning
    message contains "Example 'foo' has empty GIVEN block"
  stdout = "✓ spec.md: valid (1 warning(s))"
  exit_code = 0

### EXAMPLE: multiple_errors
GIVEN:
  file is missing ## INVARIANTS and ## EXAMPLES sections
  META is present but Deployment field is absent
  invocation: pcd-lint spec.md
WHEN:
  result = lint(file, strict=false)
THEN:
  stderr contains at least three diagnostics, all severity = Error:
    message = "Missing required section: ## INVARIANTS"
    message = "Missing required section: ## EXAMPLES"
    message = "Missing required META field: Deployment"
  stdout = "✗ spec.md: 3 error(s), 0 warning(s)"
  exit_code = 1

### EXAMPLE: file_not_found
GIVEN:
  invocation: pcd-lint missing.md
  missing.md does not exist
WHEN:
  pcd-lint is invoked
THEN:
  stderr = "error: cannot open file: missing.md"
  stdout = (empty)
  exit_code = 2

### EXAMPLE: unrecognised_option
GIVEN:
  invocation: pcd-lint verbose=yes spec.md
WHEN:
  pcd-lint is invoked
THEN:
  stderr = "error: unrecognised option: verbose"
  stdout = (empty)
  exit_code = 2

### EXAMPLE: behavior_internal_recognised
GIVEN:
  file contains all required sections, including these BEHAVIOR variants:
  ```markdown
  ## BEHAVIOR: lint
  ## BEHAVIOR/INTERNAL: precedence-resolution
  ```
  no plain "## BEHAVIOR" section without a name suffix is present.
  invocation: pcd-lint spec.md
WHEN:
  result = lint(file, strict=false)
THEN:
  stderr = (empty)
  stdout = "✓ spec.md: valid"
  exit_code = 0
  // BEHAVIOR/INTERNAL is treated as satisfying the BEHAVIOR requirement
  // and is not flagged as an unknown section

### EXAMPLE: behavior_internal_unknown_variant
GIVEN:
  file contains:
    ## BEHAVIOR/PRIVATE: foo
  no standard BEHAVIOR or BEHAVIOR/INTERNAL section present
  invocation: pcd-lint spec.md
WHEN:
  result = lint(file, strict=false)
THEN:
  stderr contains one diagnostic:
    severity = Error
    message contains "Missing required section: ## BEHAVIOR"
  exit_code = 1
  // BEHAVIOR/PRIVATE is not a recognised variant; does not satisfy
  // the BEHAVIOR requirement

### EXAMPLE: list_templates
GIVEN:
  invocation: pcd-lint list-templates
WHEN:
  list-templates is invoked
THEN:
  stdout contains exactly 17 lines
  each line contains template name and default language annotation
  for templates without a companion *.template.md file in the
    search path, annotation is "(template file not found)"
  stderr = (empty)
  exit_code = 0

### EXAMPLE: non_md_extension
GIVEN:
  invocation: pcd-lint myspec.txt
  myspec.txt exists and is readable
WHEN:
  pcd-lint is invoked
THEN:
  stderr = "error: file must have .md extension: myspec.txt"
  stdout = (empty)
  exit_code = 2

### EXAMPLE: multi_pass_example_valid
GIVEN:
  file contains a BEHAVIOR: reconcile section with STEPS including "on failure →"
  EXAMPLES contains:
  ```
  EXAMPLE: reconcile_graceful_stop
  GIVEN:
    VM "testvm-01", spec.desiredState = Stopped
    Domain is Running
  WHEN:  reconcile runs (pass 1)
  THEN:
    domain.Shutdown() is called
    result = RequeueAfter(10s)
  WHEN:  reconcile runs (pass 2); domain is Shutoff
  THEN:
    status.phase = Stopped
    result = RequeueAfter(60s)
  ```
  invocation: pcd-lint spec.md
WHEN:
  result = lint(file, strict=false)
THEN:
  stderr = (empty)
  stdout = "✓ spec.md: valid"
  exit_code = 0
  // multi-pass WHEN/THEN is valid under RULE-06

### EXAMPLE: behavior_missing_steps
GIVEN:
  file contains all required sections including:
  ```
  ## BEHAVIOR: do-something
  PRECONDITIONS:
    - input is valid
  POSTCONDITIONS:
    - output is produced
  ```
  BEHAVIOR section has no STEPS: block
  invocation: pcd-lint spec.md
WHEN:
  result = lint(file, strict=false)
THEN:
  stderr contains one diagnostic:
    severity = Error
    section = "BEHAVIOR: do-something"
    message contains "missing required STEPS: block"
  exit_code = 1

### EXAMPLE: invariant_missing_tag_warning
GIVEN:
  file is otherwise valid with INVARIANTS section:
  ```
  ## INVARIANTS
  - tool never modifies input files
  - exit_code = 2 on invocation errors
  ```
  no [observable] or [implementation] tags present
  invocation: pcd-lint spec.md
WHEN:
  result = lint(file, strict=false)
THEN:
  stderr contains two diagnostics, both severity = Warning
    messages contain "missing tag"
  stdout = "✓ spec.md: valid (2 warning(s))"
  exit_code = 0

### EXAMPLE: invariant_missing_tag_strict
GIVEN:
  same file as invariant_missing_tag_warning
  invocation: pcd-lint strict=true spec.md
WHEN:
  result = lint(file, strict=true)
THEN:
  exit_code = 1
  stdout contains "[strict mode]"

### EXAMPLE: behavior_error_exits_no_negative_example
GIVEN:
  file contains BEHAVIOR: transfer with STEPS:
    "1. Validate inputs; on failure → return Err(INVALID)"
  EXAMPLES contains only:
  ```
  EXAMPLE: successful_transfer
  GIVEN:  valid inputs
  WHEN:   transfer(a, b, 10)
  THEN:   result = Ok
  ```
  invocation: pcd-lint spec.md
WHEN:
  result = lint(file, strict=false)
THEN:
  stderr contains one diagnostic:
    severity = Error
    message contains "has error exits in STEPS but no negative-path EXAMPLE"
  exit_code = 1

### EXAMPLE: behavior_error_exits_with_negative_example
GIVEN:
  same BEHAVIOR: transfer as above
  EXAMPLES now contains an additional block:
  ```
  EXAMPLE: transfer_invalid_input
  GIVEN:  amount = -1
  WHEN:   transfer(a, b, -1)
  THEN:   result = Err(INVALID)
  ```
  invocation: pcd-lint spec.md
WHEN:
  result = lint(file, strict=false)
THEN:
  stderr = (empty)
  exit_code = 0

### EXAMPLE: behavior_constraint_invalid_value
GIVEN:
  file contains:
  ```
  ## BEHAVIOR: some-op
  Constraint: optional
  ```
  invocation: pcd-lint spec.md
WHEN:
  result = lint(file, strict=false)
THEN:
  stderr contains one diagnostic:
    severity = Error
    message contains "invalid Constraint: value 'optional'"
    message contains "Valid values: required, supported, forbidden"
  exit_code = 1

### EXAMPLE: behavior_constraint_forbidden_no_reason
GIVEN:
  file contains:
  ```
  ## BEHAVIOR: legacy-mode
  Constraint: forbidden
  ```
  no reason: annotation present
  invocation: pcd-lint spec.md
WHEN:
  result = lint(file, strict=false)
THEN:
  stderr contains one diagnostic:
    severity = Warning
    message contains "Constraint: forbidden but has no reason: annotation"
  exit_code = 0

### EXAMPLE: behavior_constraint_absent_defaults_required
GIVEN:
  file is fully valid; BEHAVIOR: transfer has no Constraint: line
  invocation: pcd-lint spec.md
WHEN:
  result = lint(file, strict=false)
THEN:
  stderr = (empty)
  exit_code = 0
  // Absent Constraint: defaults to required; no diagnostic emitted

### EXAMPLE: fenced_block_markers_ignored
GIVEN:
  file contains all required sections and is structurally valid
  the EXAMPLES section contains a block with fenced content:
  ```markdown
  EXAMPLE: outer
  GIVEN:
    some condition
  WHEN:
    ```
    EXAMPLE: fake
    WHEN: something
    THEN: something
    ```
  THEN:
    result = Ok
  ```
  invocation: pcd-lint spec.md
WHEN:
  result = lint(file, strict=false)
THEN:
  stderr = (empty)
  stdout = "✓ spec.md: valid"
  exit_code = 0
  // markers inside fenced blocks are not parsed as real structure

### EXAMPLE: milestone_valid_scaffold_first
GIVEN:
  spec contains:
    ## MILESTONE: 0.0.0
    Status: released
    Scaffold: true
    Included BEHAVIORs: collect, render, main
    Acceptance criteria:
      ./tool version | grep -q "^tool "
    ## MILESTONE: 0.1.0
    Status: active
    Included BEHAVIORs: collect
    Deferred BEHAVIORs: render, main
    Acceptance criteria:
      ./tool collect | jq '.result | length > 0'
  all listed BEHAVIORs exist in spec
  invocation: pcd-lint spec.md
WHEN:
  result = lint(file, strict=false)
THEN:
  stderr = (empty)
  exit_code = 0

### EXAMPLE: milestone_scaffold_not_first
GIVEN:
  spec contains:
    ## MILESTONE: 0.1.0
    Status: released
    Included BEHAVIORs: collect
    Deferred BEHAVIORs: render
    ## MILESTONE: 0.2.0
    Status: active
    Scaffold: true
    Included BEHAVIORs: collect, render
  invocation: pcd-lint spec.md
WHEN:
  result = lint(file, strict=false)
THEN:
  stderr contains one diagnostic:
    severity = Error
    message contains "Scaffold milestone '0.2.0' must appear first"
  exit_code = 1

### EXAMPLE: milestone_two_scaffold_rejected
GIVEN:
  spec contains two MILESTONE sections both with Scaffold: true
  invocation: pcd-lint spec.md
WHEN:
  result = lint(file, strict=false)
THEN:
  stderr contains one diagnostic:
    severity = Error
    message contains "More than one MILESTONE has Scaffold: true"
  exit_code = 1

### EXAMPLE: milestone_two_active_rejected
GIVEN:
  spec contains two MILESTONE sections both with Status: active
  invocation: pcd-lint spec.md
WHEN:
  result = lint(file, strict=false)
THEN:
  stderr contains one diagnostic:
    severity = Error
    message contains "More than one MILESTONE has Status: active"
  exit_code = 1

### EXAMPLE: milestone_unknown_behavior_name
GIVEN:
  spec contains:
    ## MILESTONE: 0.1.0
    Status: active
    Included BEHAVIORs: collect, nonexistent-behavior
    Deferred BEHAVIORs: render
  "nonexistent-behavior" is not declared as a BEHAVIOR in the spec
  invocation: pcd-lint spec.md
WHEN:
  result = lint(file, strict=false)
THEN:
  stderr contains one diagnostic:
    severity = Error
    message contains "lists BEHAVIOR 'nonexistent-behavior' under Included BEHAVIORs
                      but no such BEHAVIOR exists in the spec"
  exit_code = 1

### EXAMPLE: includes_path_resolves

GIVEN: a spec file `host.md` with META containing
`Includes: shared/rules.md`, and the file `shared/rules.md` exists and is
readable relative to host.md's location.
WHEN: `pcd-lint host.md` is run.
THEN: no RULE-19 diagnostic is emitted.

### EXAMPLE: includes_path_unresolvable

GIVEN: a spec file `host.md` with META containing
`Includes: missing/rules.md`, where no such file exists.
WHEN: `pcd-lint host.md` is run.
THEN: an error is emitted:
`ERROR  host.md:{line}  [META]  Includes path does not resolve:
missing/rules.md`

### EXAMPLE: merged_spec_no_collisions

GIVEN: a host spec declaring TYPE `RuleId` and `Includes: shared.md`,
where `shared.md` declares only TYPEs `Severity` and `Diagnostic` (no
overlap).
WHEN: `pcd-lint host.md` is run.
THEN: no RULE-20 diagnostic is emitted.

### EXAMPLE: merged_spec_type_collision

GIVEN: a host spec declaring TYPE `Diagnostic` and `Includes: shared.md`,
where `shared.md` also declares TYPE `Diagnostic`.
WHEN: `pcd-lint host.md` is run.
THEN: an error is emitted:
`ERROR  host.md  [META]  Name collision after merge: TYPE Diagnostic
appears in both shared.md and host.md`

### EXAMPLE: inclusion_acyclic

GIVEN: `a.md` includes `b.md`, `b.md` includes `c.md`, `c.md` includes
nothing.
WHEN: `pcd-lint a.md` is run.
THEN: no RULE-21 cycle diagnostic is emitted.

### EXAMPLE: inclusion_cycle

GIVEN: `a.md` includes `b.md`, `b.md` includes `a.md`.
WHEN: `pcd-lint a.md` is run.
THEN: an error is emitted:
`ERROR  a.md  [META]  Inclusion cycle: a.md → b.md → a.md`

### EXAMPLE: included_spec_with_milestone_rejected

GIVEN: `host.md` includes `shared.md`, where `shared.md` contains a
`## MILESTONE: 0.1.0` section.
WHEN: `pcd-lint host.md` is run.
THEN: an error is emitted:
`ERROR  shared.md  [structure]  Included spec must not declare MILESTONE
section: shared.md`

### EXAMPLE: included_spec_with_deployment_rejected

GIVEN: `host.md` includes `shared.md`, where `shared.md` contains a
`## DEPLOYMENT` section.
WHEN: `pcd-lint host.md` is run.
THEN: an error is emitted:
`ERROR  shared.md  [structure]  Included spec must not declare DEPLOYMENT
section: shared.md`

### EXAMPLE: example_header_heading_form_accepted

GIVEN:
  EXAMPLES section contains: "### EXAMPLE: foo\nGIVEN:\n...\nWHEN:\n...\nTHEN:\n..."
WHEN:
  pcd-lint spec.md
THEN:
  no RULE-06 diagnostic about heading form

### EXAMPLE: example_header_flat_form_rejected

GIVEN:
  EXAMPLES section contains: "EXAMPLE: foo\nGIVEN:\n..."
  (flat form, no heading)
WHEN:
  pcd-lint spec.md
THEN:
  error emitted: "Example header must be '### EXAMPLE: <name>'. Flat 'EXAMPLE: <name>' is no longer accepted (changed in v0.4.0)."

### EXAMPLE: example_header_wrong_heading_level_rejected

GIVEN:
  EXAMPLES section contains: "#### EXAMPLE: foo\nGIVEN:\n..."
  (heading level 4 instead of 3)
WHEN:
  pcd-lint spec.md
THEN:
  error emitted: "Example header must be at heading level 3 (three pound signs). Found heading level 4 for 'foo'."

---

## DEPLOYMENT

Runtime: command-line tool, single static binary, no runtime dependencies

Parsing approach:
  The specification describes validation rule semantics only, not the
  internal parsing implementation. Translators are free to choose any
  parsing strategy — line-by-line state machine, AST, regex, or other.
  The EXAMPLES section is the acceptance test: a correct implementation
  must satisfy all examples regardless of internal parsing approach.
  Common strategies observed in practice:
  - Line-by-line state machine: simple, sufficient for v1 rules
  - Markdown AST parser: more robust for edge cases, higher complexity
  Translators should document their parsing approach in the translation
  report.

  Code-fence exclusion: all content between opening and closing
  code-fence markers (lines where TrimSpace(L) begins with ``` or ~~~)
  is excluded from all structural parsing. No PCD markers, section
  headers, EXAMPLE:, GIVEN:, WHEN:, THEN:, BEHAVIOR patterns, STEPS:,
  Constraint:, or INVARIANTS entries are recognised inside fenced
  blocks. Translators must implement this as a fence-depth counter
  (not a boolean toggle) in the main parsing loop. The depth increments
  on any fence-open marker and decrements on any fence-close marker;
  content is excluded when depth > 0. This correctly handles nested
  fences (e.g. a GIVEN block containing a fenced example that itself
  contains a fenced inner block).

  Column-0 requirement: all structural markers are only recognised
  when they appear at column 0 (no leading whitespace). This applies
  to section headers (## X), EXAMPLE:, GIVEN:, WHEN:, THEN:, STEPS:,
  and Constraint:. A line such as "    ## BEHAVIOR/PRIVATE: foo"
  appearing inside a GIVEN block description is content, not a section
  boundary. Translators must check the original untrimmed line for
  the presence of these markers, not the TrimSpace'd form.
  Exception: fence detection uses TrimSpace(L) (step 2a above) so
  that indented fences inside GIVEN blocks are correctly recognised.

Template search path:
  All four directories are searched; later entries take precedence (last-wins).
  Directories that do not exist are silently skipped.
  Search order (ascending precedence):
    1. /usr/share/pcd/templates/    vendor default (pcd-templates package)
    2. /etc/pcd/templates/          system administrator
    3. ~/.config/pcd/templates/     user
    4. ./.pcd/templates/            project-local
  Platform: Linux only in v1.

Runtime dependency:
  Requires: pcd-templates
  The pcd-templates package installs template and hints files to
  /usr/share/pcd/templates/ and /usr/share/pcd/hints/.
  pcd-lint reads templates from the search path at runtime.
  pcd-lint does not install template or hints files itself.

Invocation:
  pcd-lint <specfile.md>
  pcd-lint strict=true <specfile.md>
  pcd-lint check-report=true <specfile.md>
  pcd-lint strict=true check-report=true <specfile.md>
  pcd-lint list-templates

Key=value options (all optional, precede the file argument):
  strict=true          Treat warnings as errors; exit 1 on warnings
                       Default: strict=false
  check-report=true    Also evaluate RULE-18: look for TRANSLATION_REPORT.md
                       adjacent to the spec, verify Spec-SHA256 field presence
                       and hash currency. Emits warnings only; never errors.
                       Default: check-report=false

Commands (bare words, no file argument):
  list-templates  Print all known deployment templates and exit 0
                  Note: language defaults for templates other than
                  cli-tool require companion *.template.md files
                  to be present in the template search path.
                  If a companion file is absent, the annotation
                  "(template file not found)" is emitted for that entry.
  version         Print pcd-lint version, Spec-Schema version, and
                  embedded SPDX list version, then exit 0.
                  Format: pcd-lint {version} (schema {spec-schema}) spdx/{spdx-version}

Output streams:
  stderr: diagnostic lines (errors and warnings)
  stdout: summary line (lint) or template list (list-templates)

Diagnostic line format (stderr):
  {SEVERITY}  {file}:{line}  [{section}]  {message}

  Examples:
    ERROR    account_transfer.md:1    [structure]  Missing required section: ## INVARIANTS
    ERROR    account_transfer.md:4    [META]       Missing required META field: Deployment
    WARNING  account_transfer.md:6    [META]       META field 'Target' is deprecated since v0.3.0
    ERROR    account_transfer.md:42   [EXAMPLES]   Example 'foo' missing THEN: marker

Summary line format (stdout):
// This format is normative. Implementations must match exactly.
// See also: BEHAVIOR: lint POSTCONDITIONS.
  ✓ {file}: valid                                        exit 0, no warnings
  ✓ {file}: valid ({N} warning(s))                       exit 0, warnings present, strict=false
  ✗ {file}: {N} error(s), {M} warning(s)                 exit 1, strict=false
  ✗ {file}: {N} error(s), {M} warning(s) [strict mode]   exit 1, strict=true

Installation:
  OBS package: pcd-tools
  Available for: openSUSE Leap, SUSE Linux Enterprise, Fedora, Debian/Ubuntu
  No curl-based installation.
  Requires: pcd-templates (provides template and hints files)
    openSUSE/SLES default install path: /usr/share/pcd/templates/

Platform:
  Linux (primary)
  macOS (supported, template path TBD in v2)
  Windows (not supported in v1)

Signal handling note:
  SIGNAL-HANDLING: SIGTERM and SIGINT are declared required in the
  cli-tool template. In practice all tested translators omitted this
  or noted it as a deviation. For v1, clean exit on SIGTERM/SIGINT
  is required but acceptable to implement as the Go/C runtime default
  behaviour (no explicit handler needed for a short-lived CLI tool
  that does not hold open file handles or sockets). Translators must
  document signal handling approach in the translation report.
