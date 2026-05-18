


# cli-tool.template

## META
Deployment:  template
Version:     0.3.26
Spec-Schema: 0.4.0
Author:      Matthias G. Eckermann <pcd@mailbox.org>
License:     CC-BY-4.0
Verification: none
Safety-Level: QM
Template-For: cli-tool

---

## TYPES

```
Constraint := required | supported | default | forbidden

TemplateRow := {
  key:        string where non-empty,
  value:      string where non-empty,
  constraint: Constraint,
  notes:      string         // human-readable explanation; may be empty
}

TemplateTable := List<TemplateRow>
// Rows with identical key are collected as a list for that key.
// Order within repeated keys is not significant.

Platform := Linux | macOS | Windows

OutputFormat := RPM | DEB | OCI | PKG | binary
// binary = raw executable, no packaging

Language := Go | Rust | C | CPP | CSharp
```

---

## BEHAVIOR: resolve
Constraint: required

Given a spec declaring `Deployment: cli-tool`, a translator reads this
template to determine defaults, constraints, and valid overrides before
generating any code or build configuration.

INPUTS:
```
template: TemplateTable
spec_meta: Map<string, string>    // the META fields from the spec
preset:    Map<string, string>    // merged preset (system + user + project)
```

OUTPUTS:
```
resolved: Map<string, string>     // effective settings for this build
warnings: List<string>            // advisory messages to surface
errors:   List<string>            // constraint violations; non-empty → reject
```

PRECONDITIONS:
- template is the cli-tool template (Template-For = "cli-tool")
- spec_meta contains at least Deployment, Verification, Safety-Level

STEPS:
1. Verify Template-For = "cli-tool"; on mismatch → error, halt.
2. Merge preset layers in order: vendor → system → user → project (last writer wins).
3. For each constraint=required key K: if not resolved → errors += violation.
4. For each constraint=default key K: apply preset value if present, else template default.
5. For each constraint=forbidden key K: if present in spec_meta or any preset → errors += violation.
6. For each constraint=supported key K: apply if declared in spec_meta or preset; skip silently if absent.
7. Apply LANGUAGE precedence: project preset > user preset > system preset > template default.
8. Validate cross-key constraints (e.g. BINARY-TYPE vs LANGUAGE, PLATFORM vs OUTPUT-FORMAT).
   On violation → errors += constraint description.
9. If errors non-empty → return errors (reject, do not return resolved).
   Else → return resolved.

POSTCONDITIONS:
- resolved contains an effective value for every required key
- for each key K with constraint=required: resolved[K] is set, else errors += violation
- for each key K with constraint=default: resolved[K] = preset[K] if present,
  else resolved[K] = template default value for K
- for each key K with constraint=forbidden: if spec_meta contains K,
  errors += "Key <K> is forbidden for Deployment: cli-tool"
- for each key K with constraint=supported: resolved[K] set only if
  spec_meta or preset declares it; no error if absent
- resolved["LANGUAGE"] follows precedence:
    project preset > user preset > system preset > template default

---

## BEHAVIOR/INTERNAL: precedence-resolution
Constraint: required

Defines how conflicting values across layers are resolved for any key.

STEPS:
1. Start with template defaults as the base map.
2. Merge /usr/share/pcd/presets/ values (vendor defaults); later entries override earlier.
3. Merge /etc/pcd/presets/ values (system admin); overrides vendor defaults.
4. Merge ~/.config/pcd/presets/ values (user); overrides system.
5. Merge <project-dir>/.pcd/ values (project-local); overrides user.
6. For each key in spec META: if constraint=supported → apply; if constraint=required or default →
   emit Warning: "Spec overrides template default for <K>. Ensure this is intentional."
7. If spec META declares a constraint=forbidden key → emit Error: "Key <K> is forbidden in cli-tool specs."
8. Return merged result.

Resolution order (last writer wins):
  1. template default
  2. /usr/share/pcd/presets/    (vendor default)
  3. /etc/pcd/presets/          (system administrator)
  4. ~/.config/pcd/presets/     (user)
  5. <project-dir>/.pcd/        (project-local, committed to git)
  6. spec META explicit override        (only permitted for constraint=supported keys)

If spec META declares a value for a constraint=required or constraint=default key,
emit Warning: "Spec overrides template default for <K>. Ensure this is intentional."

If spec META declares a value for a constraint=forbidden key,
emit Error: "Key <K> is forbidden in cli-tool specs."

---

## TEMPLATE-TABLE

| Key | Value | Constraint | Notes |
|-----|-------|------------|-------|
| VERSION | MAJOR.MINOR.PATCH | required | Semantic versioning. Spec author increments on every meaningful change. |
| SPEC-SCHEMA | MAJOR.MINOR.PATCH | required | Version of the Post-Coding spec schema this file was written against. |
| AUTHOR | name <email> | required | At least one Author: line required. Repeating key; multiple authors permitted. |
| LICENSE | SPDX identifier | required | Must be a valid SPDX license identifier or compound expression. Example: Apache-2.0. |
| LANGUAGE | Go | default | Default target language. Override via preset. Valid alternatives: Rust, C, C++, C#. |
| LANGUAGE-ALTERNATIVES | Rust | supported | May be selected via preset or project override. |
| LANGUAGE-ALTERNATIVES | C | supported | May be selected via preset or project override. |
| LANGUAGE-ALTERNATIVES | C++ | supported | May be selected via preset or project override. |
| LANGUAGE-ALTERNATIVES | C# | supported | Primary use case: Windows platform. Requires .NET runtime. |
| BINARY-TYPE | static | default | Default: single static binary. No shared library dependencies at runtime. |
| BINARY-TYPE | dynamic | supported | Permitted for C, C++, and C# only. Dynamic linking may be preferable when system libraries are large, versioned independently, or required by platform conventions. Not permitted for Go or Rust (use static). |
| SOURCE-PARTITIONING | modular | required | Implementation source must be partitioned into multiple modules. A single monolithic source file containing all logic is not permitted. The partitioning unit follows the target language's package/module/namespace convention (Go: packages under `internal/`; Rust: modules under `src/`; C: separate header and source files per logical unit; C#: namespaces with one type per file). |
| SOURCE-PARTITIONING | one-entry-one-implementation | required | At minimum: one entry-point module containing only CLI dispatch (argument parsing, top-level error reporting, calling into the implementation), and at least one implementation module containing the spec's behaviours. The entry-point module does not implement behaviours directly. |
| SOURCE-PARTITIONING | by-behaviour-domain | supported | Implementation may be further partitioned by behavioural domain when the spec defines multiple distinct BEHAVIORs. Example: a linter may partition into parsing, rule-application, and formatting modules. The partitioning is the translator's choice, constrained only by the minimum above. |
| MODULE-IDENTITY | host-specified | required | The implementation declares its module identity (Go: `go.mod` module line; Rust: `Cargo.toml` package.name; C: pkg-config name; C#: AssemblyName) using a value supplied by an authoritative source. The translator MUST NOT invent or substitute the identity. Authoritative sources, in priority order: (1) spec META `Module:` field, (2) language-specific hints file, (3) existing manifest from a prior translation in the output directory. If no source provides an identity, the translator halts with a diagnostic and requests one. The translator never infers identity from project name, repository URL guesswork, or heuristic from spec content. |
| MODULE-IDENTITY | propagated | required | The module identity, once chosen, must appear consistently across all packaging artefacts: source manifests, package metadata (RPM `Source:` URL, DEB `Source:` field, `DH_GOPKG` in `debian/rules`), documentation (man page Homepage line, README install commands), and any internal package import paths. A reviewer must be able to grep the identity once and find all consumers. |
| MODULE-IDENTITY | conflict-halts | required | If multiple authoritative sources provide module identity and they disagree, the translator halts with a diagnostic identifying all conflicting sources and their values. The translator does not silently choose one. |
| PUBLIC-API-SURFACE | stable-across-translations | required | The names and signatures of functions and types exposed by implementation modules to other components (entry-point, tests, other tools that import the module) form a public API surface. This surface must remain stable across translations of the same spec at the same Version. A translation may add to the surface; it may not remove or rename existing entries without a spec Version increment. |
| PUBLIC-API-SURFACE | recorded-in-report | required | The translator records the public API surface in `TRANSLATION_REPORT.md` under a `## Public API Surface` section, listing each exported symbol with its full signature, grouped by module. The next translation reads this section as input and verifies that the new implementation preserves it. |
| BINARY-COUNT | 1 | required | Exactly one binary per spec. Multi-binary tools require separate specs. |
| RUNTIME-DEPS | none | required | No runtime dependencies permitted. All dependencies linked statically. |
| CLI-ARG-STYLE | key=value | required | Argument parsing uses key=value pairs. POSIX --flag style is forbidden for new options. v2 note: relax to default= and add supported alternatives (POSIX, subcommand) when real use cases require it. |
| CLI-ARG-STYLE | bare-words | supported | Bare word commands (e.g. list-templates) are permitted alongside key=value. |
| EXIT-CODE-OK | 0 | required | Success exit code is always 0. |
| EXIT-CODE-ERROR | 1 | required | Logical error (validation failure, lint error) exits 1. |
| EXIT-CODE-INVOCATION | 2 | required | Invocation error (bad arguments, missing file) exits 2. |
| STREAM-DIAGNOSTICS | stderr | required | Errors and warnings are written to stderr. |
| STREAM-OUTPUT | stdout | required | Normal output (summaries, lists, results) is written to stdout. |
| SIGNAL-HANDLING | SIGTERM | required | Clean exit on SIGTERM. No partial output. |
| SIGNAL-HANDLING | SIGINT | required | Clean exit on SIGINT (Ctrl-C). No partial output. |
| OUTPUT-FORMAT | RPM | required | Linux RPM package. OBS build target. |
| OUTPUT-FORMAT | DEB | required | Linux DEB package. OBS build target. |
| OUTPUT-FORMAT | OCI | supported | OCI container image. Useful for CI pipeline integration. |
| OUTPUT-FORMAT | PKG | supported | macOS installer package. Required if macOS platform is declared. |
| OUTPUT-FORMAT | binary | supported | Raw binary for platforms without package manager integration. |
| INSTALL-METHOD | OBS | required | Primary distribution via build.opensuse.org. curl-based install is forbidden. |
| INSTALL-METHOD | curl | forbidden | curl-based installation scripts are not permitted. Supply chain security requirement. |
| PLATFORM | Linux | required | Linux is the primary and required platform. |
| PLATFORM | macOS | supported | macOS support is optional. If declared, PKG output format is required. |
| PLATFORM | Windows | supported | Windows support is not targeted in v1 templates. |
| CONFIG-ENV-VARS | forbidden | forbidden | Behaviour must not be controlled via environment variables. Use key=value args or preset files. |
| NETWORK-CALLS | forbidden | forbidden | Tool must not make network calls at runtime. |
| FILE-MODIFICATION | input-files | forbidden | Tool must not modify its input files. |
| IDEMPOTENT | true | required | Running the tool twice on the same input must produce identical output. |
| PRESET-SYSTEM | systemd-style | required | Preset layering follows systemd conventions. See whitepaper A.11. |

---

## PRECONDITIONS

- This template is applied only when spec META declares Deployment: cli-tool
- Preset files must be valid TOML
- If PLATFORM includes macOS, OUTPUT-FORMAT must include PKG
- LANGUAGE value in resolved output must be one of: Go, Rust, C, C++, C#
- If LANGUAGE is C#, PLATFORM must include Windows (C# targets .NET runtime)
- If BINARY-TYPE is dynamic, LANGUAGE must be one of: C, C++, C#
- If LANGUAGE is Go or Rust, BINARY-TYPE must be static

---

## POSTCONDITIONS

- Every spec using Deployment: cli-tool is governed by this template
- A spec may not declare LANGUAGE directly in META unless using Deployment: manual
- Resolved LANGUAGE is always one of the LANGUAGE-ALTERNATIVES or the default
- curl is never an accepted install method, regardless of preset override
- Forbidden constraints cannot be overridden by any preset or spec declaration

---

## INVARIANTS

- [observable]  constraint=forbidden rows cannot be overridden at any preset layer
- [observable]  constraint=required rows must resolve to a value; missing value is an error
- [observable]  LANGUAGE resolution always produces exactly one value
- [observable]  OUTPUT-FORMAT required rows must appear in every build configuration
- [observable]  a spec declaring Deployment: cli-tool inherits all required constraints
  whether or not the spec author is aware of them
- [observable]  template version is recorded in every audit bundle that references it
- [observable]  BINARY-TYPE=dynamic is only valid when LANGUAGE ∈ {C, C++, C#}
- [observable]  BINARY-TYPE=static is the only valid value when LANGUAGE ∈ {Go, Rust}
- [observable]  every generated artifact embeds the SHA256 of the spec file
  it was produced from; an artifact without an embedded spec hash is incomplete

---

## EXAMPLES

### EXAMPLE: minimal_spec_resolution
GIVEN:
  spec META contains:
    Deployment: cli-tool
    Verification: none
    Safety-Level: QM
  no preset files present (system defaults only)
WHEN:
  resolved = resolve(template, spec_meta, preset={})
THEN:
  resolved["LANGUAGE"] = "Go"
  resolved["BINARY-TYPE"] = "static"
  resolved["CLI-ARG-STYLE"] = "key=value"
  resolved["EXIT-CODE-OK"] = "0"
  resolved["INSTALL-METHOD"] = "OBS"
  errors = []
  warnings = []

### EXAMPLE: org_preset_overrides_language
GIVEN:
  spec META contains:
    Deployment: cli-tool
    Verification: none
    Safety-Level: QM
  /etc/pcd/presets/org.toml contains:
    [templates.cli-tool]
    default_language = "rust"
WHEN:
  resolved = resolve(template, spec_meta, preset={LANGUAGE: "Rust"})
THEN:
  resolved["LANGUAGE"] = "Rust"
  errors = []
  warnings = []

### EXAMPLE: forbidden_curl_rejected
GIVEN:
  spec META contains:
    Deployment: cli-tool
    INSTALL-METHOD: curl
WHEN:
  resolved = resolve(template, spec_meta, preset={})
THEN:
  errors contains:
    "Key INSTALL-METHOD=curl is forbidden for Deployment: cli-tool"
  resolved is not produced (errors non-empty → reject)

### EXAMPLE: macos_platform_requires_pkg
GIVEN:
  spec META contains:
    Deployment: cli-tool
    Verification: none
    Safety-Level: QM
  preset declares PLATFORM includes macOS
  preset does not declare OUTPUT-FORMAT = PKG
WHEN:
  resolved = resolve(template, spec_meta, preset={PLATFORM: "macOS"})
THEN:
  errors contains:
    "PLATFORM macOS requires OUTPUT-FORMAT: PKG"
  resolved is not produced

### EXAMPLE: env_var_control_rejected
GIVEN:
  spec DEPLOYMENT section describes behaviour controlled via
  environment variable SPEC_LINT_STRICT
WHEN:
  translator processes spec
THEN:
  errors contains:
    "CONFIG-ENV-VARS is forbidden for Deployment: cli-tool. \
     Use key=value arguments or preset files instead."

### EXAMPLE: source_partitioning_modular_required
GIVEN:
  spec declares multiple BEHAVIORs (lint, list-templates, code-fence-tracking)
  translator produces a single source file containing all logic in package main
WHEN:
  translator phase 6 compile-gate runs
THEN:
  translator halts with diagnostic:
    "SOURCE-PARTITIONING constraint violated: implementation must be \
     partitioned into entry-point module and at least one implementation \
     module. Found: 1 file in package main."

### EXAMPLE: module_identity_must_come_from_authoritative_source
GIVEN:
  spec META does not declare a Module: field
  no hints file provides module identity
  no existing manifest in output directory
WHEN:
  translator phase 1 setup runs
THEN:
  translator halts with diagnostic:
    "MODULE-IDENTITY required but no authoritative source provides it. \
     Supply via spec META 'Module:' field, language-specific hints file, \
     or pre-existing manifest in output directory."
  translator does not infer identity from spec content or repository URL

### EXAMPLE: module_identity_conflict_halts
GIVEN:
  existing go.mod in output directory declares module github.com/foo/bar
  hints file declares module identity github.com/foo/baz
WHEN:
  translator phase 1 setup runs
THEN:
  translator halts with diagnostic identifying both conflicting sources
  and their values
  translator does not silently choose one source over the other

### EXAMPLE: public_api_surface_preserved_across_translation
GIVEN:
  prior TRANSLATION_REPORT.md exists with "## Public API Surface" section
  listing N exported symbols with signatures
WHEN:
  current translation produces new implementation
THEN:
  every symbol from the prior surface appears in the new implementation
  with compatible signature
  new symbols may be added
  if any prior symbol is missing or has a renamed/incompatible signature,
  translator halts with diagnostic identifying the affected symbols

### EXAMPLE: public_api_surface_recorded
GIVEN:
  successful translation produces internal/<n>/ package with exported symbols
WHEN:
  TRANSLATION_REPORT.md is written
THEN:
  report contains "## Public API Surface" section
  section lists every exported symbol with full signature
  symbols grouped by module/package

---

## DELIVERABLES

Defines the files a translator must produce for each OUTPUT-FORMAT
declared as `required` or `supported` in the TEMPLATE-TABLE.
A translator must produce all deliverables for every `required`
OUTPUT-FORMAT. For `supported` OUTPUT-FORMATs, deliverables are
produced only if that format is active in the resolved preset.

The prompt to the translator must not enumerate these files —
the translator derives them from this section.

### Delivery Order

Deliverables must be produced in the following order:
1. Core implementation files (source, go.mod, Makefile, README.md, LICENSE)
2. Required packaging artifacts (RPM, DEB) in table order
3. Supported packaging artifacts if preset active (OCI, PKG, binary)
4. TRANSLATION_REPORT.md last, after all other files are written and verified

### Deliverables Table

| OUTPUT-FORMAT | Constraint | Required Deliverable Files | Notes |
|---|---|---|---|
| source | required | `main.go` (or `cmd/<n>/main.go`) + at least one implementation file in `internal/<n>/` or equivalent, `go.mod` | Per `SOURCE-PARTITIONING: modular` and `one-entry-one-implementation`, a single-file implementation is not permitted. The entry-point file contains only CLI dispatch; behaviour implementation lives in a separate package or namespace. Translator documents the chosen partitioning in the translation report. |
| public-api | required | `TRANSLATION_REPORT.md` section `## Public API Surface` | Per `PUBLIC-API-SURFACE: recorded-in-report`. One row per exported symbol, with full signature, grouped by module. The next translation reads this section to verify continuity. |
| build | required | `Makefile` | Must include: `build`, `test`, `install`, `clean`, `man` targets. `build` target must set `CGO_ENABLED=0` for Go, `-static` for C/C++. `man` target: `pandoc <n>.1.md -s -t man -o <n>.1`. |
| docs | required | `README.md` | Must document: installation via OBS (zypper, apt, dnf), usage, flags, exit codes. Must not document curl-based installation. |
| man | required | `<n>.1.md`, `<n>.1` | Markdown source converted to troff via `pandoc`. Section 1 (user commands). Install to `%{_mandir}/man1/` (RPM) and `debian/<n>/usr/share/man/man1/` (DEB). |
| license | required | `LICENSE` | SPDX identifier from spec META + authoritative URL to the full license text. Never reproduce the full license text. |
| RPM | required | `<n>.spec` | OBS RPM spec file. Must include: Name, Version, License (SPDX), Summary, BuildRequires, %build, %install, %files sections. |
| DEB | required | `debian/control`, `debian/changelog`, `debian/rules`, `debian/copyright` | Standard Debian source package layout. `debian/copyright` must use DEP-5 machine-readable format with SPDX license identifier. |
| OCI | supported | `Containerfile` | OCI-compliant container build file. Named `Containerfile` not `Dockerfile`. Multi-stage build required. Builder stage: `FROM registry.suse.com/bci/golang:latest AS builder` for Go — never unqualified names (supply chain security requirement). Final stage: `FROM scratch`. Must not expose ports unless spec DEPLOYMENT declares them. |
| PKG | supported | `<n>.pkgbuild` | macOS installer package descriptor. Required when PLATFORM includes macOS. Minimal skeleton acceptable; document in translation report. |
| binary | supported | none | Raw binary only. No packaging descriptor required. |
| report | required | `TRANSLATION_REPORT.md` | AI translator self-evaluation. Must be Markdown. Must include: language resolution rationale, delivery mode, template constraints compliance table, ambiguities, deviations, per-example confidence levels with reasoning, parsing approach, signal handling approach. Written last after all other files verified on disk. |
| spec-hash | required | embedded in all artifacts | SHA256 of the spec file embedded in: source file header comments, `TRANSLATION_REPORT.md` `Spec-SHA256:` field, binary `--version` output, RPM `.spec` comment, DEB `control` `X-PCD-Spec-SHA256:` field, `Containerfile` `LABEL pcd.spec.sha256=`, `Makefile` `SPEC_SHA256` variable. Computed once before any output is written. |

### Naming Convention

`<n>` in the above table refers to the component name as declared
in the specification title (first `#` heading). It must be:
- lowercase
- hyphen-separated (no underscores)
- no version suffix in the filename itself (version lives inside the file)

### Deliverable Content Requirements

**RPM spec (`<n>.spec`):**
- `License:` field must use the SPDX identifier from the spec META
- `BuildRequires:` must not include curl or any network fetch tool
- `BuildRequires:` must include `pandoc` (for man page generation)
- `%build` must include: `pandoc <n>.1.md -s -t man -o <n>.1`
- `%files` must include: `%{_mandir}/man1/<n>.1*`
- `%build` section must set `CGO_ENABLED=0` for Go, `-static` for C/C++
- `Source0:` must reference a local tarball, not a URL

**debian/copyright:**
- Must use DEP-5 machine-readable format
- `License:` field must use the SPDX identifier from the spec META

**debian/control Build-Depends:**
- Must include `pandoc` in `Build-Depends`

**debian/rules:**
- Must include man page build step: `pandoc <n>.1.md -s -t man -o <n>.1`
- Man page must be installed to `usr/share/man/man1/<n>.1`

**Containerfile:**
- Must use multi-stage build: builder stage + minimal final stage
- Builder stage must use `FROM registry.suse.com/bci/golang:latest AS builder`
  for Go. Never use unqualified names such as `golang:1.24` or `docker.io/golang`.
  This is a supply chain security requirement, not a preference.
- Final stage must be `FROM scratch` (static binary; no runtime dependencies)
- Layer order in builder stage: `COPY go.mod go.sum ./` → `RUN go mod download`
  → `COPY . .` → `RUN CGO_ENABLED=0 go build`
- Must not expose any ports unless the spec DEPLOYMENT section declares them
- Must not include a package manager in the final image

**TRANSLATION_REPORT.md:**
- Must be a Markdown file (not .txt or other format)
- Must include a `Spec-SHA256:` field in the header (SHA256 of the spec file as provided)
- Must include a template constraints compliance table
- Must include per-example confidence levels with reasoning
- Must document parsing approach chosen
- Must document signal handling approach
- Must be written to disk via filesystem tool, not output to terminal

---

## DEPLOYMENT

Runtime: this file is a template specification, not executable code.
It is read by pcd-lint (for template resolution validation) and by
AI translators (for code generation context).

Location in preset hierarchy:
  /usr/share/pcd/templates/cli-tool.template.md

Versioning:
  Template version is declared in META (Version: field).
  Specs reference the template by name (Deployment: cli-tool).
  Audit bundles record the template version used at generation time.
  Breaking changes to a template increment the minor version.
  Additions of supported rows are non-breaking.
  Changes to required or forbidden rows are breaking.
  Current version: 0.3.13



---

## EXECUTION

The translator must read this section before generating any code.
It specifies the exact delivery phases, resume logic, and compile
gate for cli-tool components. Follow it exactly.

### Input files

The translator receives in the working directory:
- `cli-tool.template.md` — this deployment template
- `<spec-name>.md` — the component specification

If the spec's DEPENDENCIES section references hints files, they are also
present. Read them before writing `go.mod` or any code that uses those
libraries — they contain verified dependency version strings.

### Module identity resolution

Before any code or manifest is generated, resolve the module identity
required by `MODULE-IDENTITY: host-specified`. The translator reads
authoritative sources in this order:

1. The spec's META `Module:` field, if present.
2. Any language-specific hints file (e.g. `<spec-name>.go.hints.md`)
   declaring a module name.
3. The pre-existing manifest in the output directory from a prior
   translation (Go: `go.mod`; Rust: `Cargo.toml`).

If exactly one source provides an identity, use it. If two or more
sources provide identities and they agree, use the agreed value. If
they disagree, halt with a diagnostic identifying every source and its
value. If no source provides an identity, halt with a diagnostic
requesting one.

Identity is never inferred from project name, repository URL guesswork,
or heuristic from spec content. This is the same discipline applied to
the spec hash — derived values must come from a deterministic source.

Once resolved, the identity propagates to all artefacts where it
appears (per `MODULE-IDENTITY: propagated`).

### Resume logic

Before writing any file, list the output directory.
If a listed deliverable already exists and is non-empty, skip it — treat
it as complete and move to the next missing file. Report which files were
found and which are being produced.

### Delivery phases

Produce files in this exact order. Complete each phase before starting
the next. Do not produce `TRANSLATION_REPORT.md` until Phase 6 is done.

**Phase 1 — Translator test suite (Tests First)**
- `independent_tests/<llm-name>/<spec-name>_test.go` (and additional
  files as needed). The directory is named after the translator LLM
  per `ROLE.md` (e.g. `independent_tests/claude-sonnet-4-5/`).
- Tests must cover every EXAMPLE, every declared error path, every
  `[observable]` INVARIANT, and the boundary conditions implied by TYPE
  refinement predicates.
- Tests use Go's standard testing package; no custom in-tree harness.
- This phase MUST complete before any non-test source file is written.
  The prompt's Tests-First guard halts the translator if this directory
  is empty when Phase 2 begins.

**Phase 2 — Core implementation**
- All `.go` source files (typically `main.go`, or `cmd/<n>/main.go` for
  larger tools, plus any additional `.go` files for interfaces and helpers)
- `go.mod` — direct dependencies only; see Compile gate below

**Phase 3 — Build and packaging**
- `Makefile`
- `<n>.spec` (RPM spec)
- `debian/control`, `debian/changelog`, `debian/rules`, `debian/copyright`
- `Containerfile` (if OCI is active in preset)
- `<n>.pkgbuild` (if PKG/macOS is active in preset)
- `LICENSE`

**Phase 4 — Auxiliary artefacts**
- `translation_report/translation-workflow.pikchr`

**Phase 5 — Documentation**
- `README.md`

**Phase 6 — Compile gate** (see below)

**Phase 7 — Report (last)**
- `TRANSLATION_REPORT.md`

### Test-author syntax check

When this template is consumed in `mode: test-author`, the test-author's
syntax check (mandated by the universal prompt) consists of these commands,
run in order. Each must succeed; the first failure halts the run before
`TEST_REPORT.md` is written.

```
$ go vet ./independent_tests/<llm-name>/...
$ gofmt -l ./independent_tests/<llm-name>/
```

`go vet` must exit 0. `gofmt -l` must produce no output (an empty list of
unformatted files). A non-empty `gofmt -l` indicates a parse failure or
formatting drift; both are blockers. If the test files reference packages
or symbols that the translator will provide (e.g. `exec.Command("./pcd-lint")`
where `pcd-lint` does not yet exist), this is acceptable — the test-author
syntax check verifies that the test files themselves parse and conform to
Go's syntax, not that they compile against a complete implementation.

### Compile gate

Execute after Phase 5 and before Phase 7. If your environment cannot
execute shell commands, document this explicitly under the heading
"Phase 6 — Compile gate not executed" in TRANSLATION_REPORT.md and
state why. Do not silently omit this phase.

**Step 1 — Dependency resolution**

Run: `go mod tidy`

This resolves all direct and indirect dependencies and writes `go.sum`.
Do not hand-write indirect dependencies — they must come from `go mod tidy`.

If `go mod tidy` cannot be run:
- Produce `go.mod` with direct dependencies only, no `go.sum`
- Note in TRANSLATION_REPORT.md that `go mod tidy` must be run before building

**Step 2 — Compilation**

Run: `go build ./...`

If compilation fails, fix only the identified errors and re-run.
Do not rewrite unaffected files. Repeat until compilation succeeds
or all reasonable fixes are exhausted.

**Step 3 — Translator test run**

Run: `go test ./independent_tests/<llm-name>/...`

If tests fail, either fix the implementation (logged in Test Refinements
as `code fixed`) or refine the test with documented rationale referencing
the spec (logged as `test edited`). Never edit a test without justification.

**Step 4 — Test-author test run** (dual-LLM mode only)

If a `independent_tests/<other-role-llm-name>/` directory exists and the
continuity checks in the prompt's step 7 passed, run:
`go test ./independent_tests/<other-role-llm-name>/...`

Record results separately. Do not edit test-author's tests under any
circumstances.

**Step 5 — Record result**

Record pass/fail for each step in TRANSLATION_REPORT.md.
Once all steps pass, do not modify any source files further.
Proceed immediately to Phase 7.

### go.mod rules

- Declare only direct dependencies (those your code imports directly)
- Do NOT hand-write indirect dependencies (resolved by `go mod tidy`)
- Do NOT fabricate pseudo-versions or commit hashes for untagged modules
  If hints files are present: use the verified versions they provide
  If no hints file: flag the dependency in TRANSLATION_REPORT.md as
  requiring manual version verification before building
