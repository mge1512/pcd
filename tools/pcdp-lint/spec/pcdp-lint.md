# pcdp-lint

## META
Deployment:  cli-tool
Version:     0.3.7
Spec-Schema: 0.3.7
Author:      Matthias G. Eckermann <pcdp@mailbox.org>
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
// pcdp-lint validates against the current SPDX license list embedded at build time

SemanticVersion := string where matches "^[0-9]+\.[0-9]+\.[0-9]+$"
// MAJOR.MINOR.PATCH — no pre-release suffixes in v1

TemplatePath := path
// The directory where deployment template files are found.
// Set at compile time via build variable TEMPLATE_DIR.
// Default (Linux): /usr/share/pcdp/templates/
// Runtime search order (first match wins):
//   1. TEMPLATE_DIR (compile-time default, read-only)
//   2. /etc/pcdp/templates/       (system administrator additions)
//   3. ~/.config/pcdp/templates/  (user additions)
//   4. ./.pcdp/templates/         (project-local)
// v1 supports Linux only. macOS and Windows paths deferred to v2.

DeploymentTemplate := one_of(
  "wasm" | "ebpf" | "kernel-module" | "verified-library" |
  "cli-tool" | "gui-tool" | "cloud-native" | "backend-service" |
  "library-c-abi" | "enterprise-software" | "academic" |
  "python-tool" | "enhance-existing" | "manual" | "template"
)
// "crypto-library" is retired as of 0.3.6. Use "verified-library" instead.
// "verified-library" covers all safety- and security-critical C-ABI libraries.
// "python-tool" is QM safety level only; Verification: none mandatory.
// "template" is used exclusively in deployment template definition files
// (*.template.md). A spec using Deployment: template is a template
// specification, not a translatable component.

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

The primary operation. Validates a specification file against the
structural rules defined in this specification.

INPUTS:
```
file:   SpecFile
strict: bool     // strict=true treats warnings as errors; default false
```

OUTPUTS:
```
result: LintResult
```

PRECONDITIONS:
- file exists and is readable
- file has `.md` extension
- strict is a valid boolean (true | false)

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

POSTCONDITIONS:
- exit_code = 0 always
- stdout contains exactly 15 lines, one per known DeploymentTemplate value
- each line format: "<template-name>  →  <default-language>"
- for enhance-existing: "<template-name>  →  (declare Language: in META)"
- for manual:           "<template-name>  →  (declare Target: in META)"
- for template:         "<template-name>  →  (template definition file, not translatable)"
- nothing written to stderr

---

## BEHAVIOR: lint-validation-rules

Defines the ordered set of structural rules applied during lint.
All rules are evaluated; lint does not stop at first error.

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
  - "## BEHAVIOR: <name>"          user-facing operation
  - "## BEHAVIOR/INTERNAL: <name>" internal implementation logic,
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

// pcdp-lint embeds the SPDX license list at build time.
// The embedded list version is reported in pcdp-lint version output.

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
             Run 'pcdp-lint list-templates' to see valid values."

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

An example block consists of:
  - a line matching "^EXAMPLE:" (example name declaration)
  - a line matching "^GIVEN:"   (precondition state)
  - a line matching "^WHEN:"    (operation)
  - a line matching "^THEN:"    (expected outcome)
  appearing in this order within the block.

if no example block found:
  emit Error, section="EXAMPLES",
    message="EXAMPLES section contains no example blocks. \
             Each example requires EXAMPLE:, GIVEN:, WHEN:, THEN: markers."

For each example block E:
  if E missing "GIVEN:":
    emit Error, section="EXAMPLES",
      message="Example '{n}' missing GIVEN: marker"
  if E missing "WHEN:":
    emit Error, section="EXAMPLES",
      message="Example '{n}' missing WHEN: marker"
  if E missing "THEN:":
    emit Error, section="EXAMPLES",
      message="Example '{n}' missing THEN: marker"

### RULE-07: EXAMPLES minimum content

Block boundaries are defined as follows:
  GIVEN block  := lines strictly between GIVEN: and WHEN: markers
  WHEN block   := lines strictly between WHEN: and THEN: markers
  THEN block   := lines after THEN: marker until one of:
                    - next EXAMPLE: marker at start of line
                    - next ## heading at start of line
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

- pcdp-lint does not modify any file on disk
- pcdp-lint does not make network calls
- pcdp-lint does not read environment variables for behaviour control
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

- GLOBAL: pcdp-lint is idempotent — running it twice on the same file
  produces identical output and identical exit code
- GLOBAL: all Error diagnostics produce exit_code ≥ 1
- GLOBAL: Warnings alone never produce exit_code = 1 unless strict=true
- GLOBAL: exit_code = 1 with only Warnings requires strict=true
- GLOBAL: exit_code = 2 indicates invocation error only,
  never a lint result
- GLOBAL: diagnostic line numbers are monotonically non-decreasing
  within a result
- GLOBAL: pcdp-lint never produces exit_code = 0 when any Error
  diagnostic is present, regardless of strict value
- GLOBAL: stderr receives diagnostics; stdout receives summary and
  list-templates output; these streams are never swapped

---

## EXAMPLES

EXAMPLE: valid_minimal_spec
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
  invocation: pcdp-lint spec.md
WHEN:
  result = lint(file, strict=false)
THEN:
  stderr = (empty)
  stdout = "✓ spec.md: valid"
  exit_code = 0

EXAMPLE: multiple_authors_valid
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
  invocation: pcdp-lint spec.md
WHEN:
  result = lint(file, strict=false)
THEN:
  stderr = (empty)
  stdout = "✓ spec.md: valid"
  exit_code = 0

EXAMPLE: invalid_spdx_license
GIVEN:
  META contains:
    License: MIT License
  all other META fields valid
  invocation: pcdp-lint spec.md
WHEN:
  result = lint(file, strict=false)
THEN:
  stderr contains one diagnostic:
    severity = Error
    message contains "License 'MIT License' is not a valid SPDX identifier"
    message contains "https://spdx.org/licenses/"
  exit_code = 1

EXAMPLE: invalid_version_format
GIVEN:
  META contains:
    Version: 1.0
  all other META fields valid
  invocation: pcdp-lint spec.md
WHEN:
  result = lint(file, strict=false)
THEN:
  stderr contains one diagnostic:
    severity = Error
    message contains "Version '1.0' is not valid semantic versioning"
  exit_code = 1

EXAMPLE: missing_author
GIVEN:
  META contains all required fields except no Author: line present
  invocation: pcdp-lint spec.md
WHEN:
  result = lint(file, strict=false)
THEN:
  stderr contains one diagnostic:
    severity = Error
    message = "Missing required META field: Author (at least one Author: line required)"
  exit_code = 1

EXAMPLE: missing_section
GIVEN:
  file is missing the ## INVARIANTS section
  all other required sections present and valid
  invocation: pcdp-lint spec.md
WHEN:
  result = lint(file, strict=false)
THEN:
  stderr contains one line:
    "ERROR  spec.md:{line}  [structure]  Missing required section: ## INVARIANTS"
  stdout = "✗ spec.md: 1 error(s), 0 warning(s)"
  exit_code = 1

EXAMPLE: unknown_deployment_template
GIVEN:
  file is valid except META contains: Deployment: serverless
  invocation: pcdp-lint spec.md
WHEN:
  result = lint(file, strict=false)
THEN:
  stderr contains one diagnostic:
    severity = Error
    message contains "Unknown deployment template: 'serverless'"
  exit_code = 1

EXAMPLE: deprecated_target_field_permissive
GIVEN:
  file is valid with META containing:
    Deployment: backend-service
    Target: Go
    Verification: none
    Safety-Level: QM
  invocation: pcdp-lint spec.md
WHEN:
  result = lint(file, strict=false)
THEN:
  stderr contains one diagnostic:
    severity = Warning
    message contains "META field 'Target' is deprecated since v0.3.0"
  stdout = "✓ spec.md: valid (1 warning(s))"
  exit_code = 0

EXAMPLE: deprecated_target_field_strict
GIVEN:
  same file as deprecated_target_field_permissive
  invocation: pcdp-lint strict=true spec.md
WHEN:
  result = lint(file, strict=true)
THEN:
  stderr contains one diagnostic:
    severity = Warning
    message contains "META field 'Target' is deprecated since v0.3.0"
  stdout = "✗ spec.md: 0 error(s), 1 warning(s) [strict mode]"
  exit_code = 1

EXAMPLE: enhance_existing_missing_language
GIVEN:
  file META contains:
    Deployment: enhance-existing
    Verification: none
    Safety-Level: QM
  META does not contain a Language: field
  invocation: pcdp-lint spec.md
WHEN:
  result = lint(file, strict=false)
THEN:
  stderr contains one diagnostic:
    severity = Error
    message = "Deployment 'enhance-existing' requires META field 'Language'"
  exit_code = 1

EXAMPLE: empty_given_block_permissive
GIVEN:
  file is structurally valid, but EXAMPLES contains:
    EXAMPLE: foo
    GIVEN:
    WHEN:
      result = foo()
    THEN:
      result = Ok
  GIVEN block contains no lines
  invocation: pcdp-lint spec.md
WHEN:
  result = lint(file, strict=false)
THEN:
  stderr contains one diagnostic:
    severity = Warning
    message contains "Example 'foo' has empty GIVEN block"
  stdout = "✓ spec.md: valid (1 warning(s))"
  exit_code = 0

EXAMPLE: multiple_errors
GIVEN:
  file is missing ## INVARIANTS and ## EXAMPLES sections
  META is present but Deployment field is absent
  invocation: pcdp-lint spec.md
WHEN:
  result = lint(file, strict=false)
THEN:
  stderr contains at least three diagnostics, all severity = Error:
    message = "Missing required section: ## INVARIANTS"
    message = "Missing required section: ## EXAMPLES"
    message = "Missing required META field: Deployment"
  stdout = "✗ spec.md: 3 error(s), 0 warning(s)"
  exit_code = 1

EXAMPLE: file_not_found
GIVEN:
  invocation: pcdp-lint missing.md
  missing.md does not exist
WHEN:
  pcdp-lint is invoked
THEN:
  stderr = "error: cannot open file: missing.md"
  stdout = (empty)
  exit_code = 2

EXAMPLE: unrecognised_option
GIVEN:
  invocation: pcdp-lint verbose=yes spec.md
WHEN:
  pcdp-lint is invoked
THEN:
  stderr = "error: unrecognised option: verbose"
  stdout = (empty)
  exit_code = 2

EXAMPLE: behavior_internal_recognised
GIVEN:
  file contains all required sections
  file contains:
    ## BEHAVIOR: lint
    ## BEHAVIOR/INTERNAL: precedence-resolution
  no plain "## BEHAVIOR" section without a name suffix
  invocation: pcdp-lint spec.md
WHEN:
  result = lint(file, strict=false)
THEN:
  stderr = (empty)
  stdout = "✓ spec.md: valid"
  exit_code = 0
  // BEHAVIOR/INTERNAL is treated as satisfying the BEHAVIOR requirement
  // and is not flagged as an unknown section

EXAMPLE: behavior_internal_unknown_variant
GIVEN:
  file contains:
    ## BEHAVIOR/PRIVATE: foo
  no standard BEHAVIOR or BEHAVIOR/INTERNAL section present
  invocation: pcdp-lint spec.md
WHEN:
  result = lint(file, strict=false)
THEN:
  stderr contains one diagnostic:
    severity = Error
    message contains "Missing required section: ## BEHAVIOR"
  exit_code = 1
  // BEHAVIOR/PRIVATE is not a recognised variant; does not satisfy
  // the BEHAVIOR requirement

EXAMPLE: list_templates
GIVEN:
  invocation: pcdp-lint list-templates
WHEN:
  list-templates is invoked
THEN:
  stdout contains exactly 15 lines
  each line contains template name and default language annotation
  for templates without a companion *.template.md file in the
    search path, annotation is "(template file not found)"
  stderr = (empty)
  exit_code = 0

EXAMPLE: non_md_extension
GIVEN:
  invocation: pcdp-lint myspec.txt
  myspec.txt exists and is readable
WHEN:
  pcdp-lint is invoked
THEN:
  stderr = "error: file must have .md extension: myspec.txt"
  stdout = (empty)
  exit_code = 2

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

Template search path (compile-time variable TEMPLATE_DIR):
  Default (Linux): /usr/share/pcdp/templates/
  Runtime search order (first match wins, later entries take precedence):
    1. TEMPLATE_DIR                          compiled-in default
    2. /etc/pcdp/templates/           system administrator
    3. ~/.config/pcdp/templates/      user
    4. ./.pcdp/templates/             project-local
  Platform: Linux only in v1.

Invocation:
  pcdp-lint <specfile.md>
  pcdp-lint strict=true <specfile.md>
  pcdp-lint list-templates

Key=value options (all optional, precede the file argument):
  strict=true     Treat warnings as errors; exit 1 on warnings
                  Default: strict=false

Commands (bare words, no file argument):
  list-templates  Print all known deployment templates and exit 0
                  Note: language defaults for templates other than
                  cli-tool require companion *.template.md files
                  to be present in the template search path.
                  If a companion file is absent, the annotation
                  "(template file not found)" is emitted for that entry.
  version         Print pcdp-lint version, Spec-Schema version, and
                  embedded SPDX list version, then exit 0.
                  Format: pcdp-lint {version} (schema {spec-schema}) spdx/{spdx-version}

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
  OBS package: pcdp-tools
  Available for: openSUSE Leap, SUSE Linux Enterprise, Fedora, Debian/Ubuntu
  No curl-based installation.
  Build variable: TEMPLATE_DIR must be set at OBS build time.
    openSUSE/SLES default: /usr/share/pcdp/templates/

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
