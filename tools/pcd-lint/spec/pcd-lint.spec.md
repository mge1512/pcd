



# pcd-lint

## META
Deployment:  cli-tool
Version:     0.4.1
Spec-Schema: 0.4.0
Author:      Matthias G. Eckermann <pcd@mailbox.org>
License:     GPL-2.0-only
Verification: none
Safety-Level: QM
Module:       github.com/mge1512/pcd/tools/pcd-lint
Includes:     ../../shared/spec/lint-rules.md

---

## TYPES

```
SpecFile := path where file_exists AND readable AND extension = ".md"

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

### Currently-shipped templates

The set of templates known to this build. The list is mechanically
derived from `templates/` by `make spec`; do not hand-edit between the
markers.

<!-- BEGIN AUTO: known-templates -->
- `abap-report` — default language: —
- `backend-service` — default language: Go
- `cli-tool` — default language: Go
- `cloud-native` — default language: Go
- `cockpit-module` — default language: —
- `gui-tool` — default language: CPP
- `kubectl-style-cli` — default language: Go
- `library-c-abi` — default language: C
- `mcp-server` — default language: Go
- `project-manifest` — default language: —
- `python-tool` — default language: Python
- `spack-package` — default language: —
- `verified-library` — default language: C
<!-- END AUTO: known-templates -->

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
3. Apply the lint-validation-rules BEHAVIOR (defined in the included
   lint-rules.md spec), which runs RULE-01 through RULE-21 in order, and
   collect all diagnostics.
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
- stdout line count matches the count phrase below
- each line format: "<template-name>  →  <default-language>"

<!-- BEGIN AUTO: known-templates-count -->
exactly 13 lines, one per known DeploymentTemplate value
<!-- END AUTO: known-templates-count -->

- for enhance-existing: "<template-name>  →  (declare Language: in META)"
- for manual:           "<template-name>  →  (declare Target: in META)"
- for template:         "<template-name>  →  (template definition file, not translatable)"
- for project-manifest: "<template-name>  →  (architect artifact, no code generated)"
- nothing written to stderr

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
