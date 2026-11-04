# library.template

## META
Deployment:  template
Version:     0.1.0
Spec-Schema: 0.4.0
Author:      Matthias G. Eckermann <pcd@mailbox.org>
License:     CC-BY-4.0
Verification: none
Safety-Level: QM
Template-For: library

---

A general-purpose, language-parameterised library: a linkable component
with a stable public interface and no executable entry point, consumed
by other PCD components at build time. Unlike `library-c-abi` (stable C
ABI across a shared-object boundary) and `verified-library` (safety- or
security-critical, formal verification path), this template produces a
library in the consumer's own language, linked natively.

The template deliberately declares **no default language**. A library
exists to be linked; its realisation language follows its consumers.
The language is chosen per binding via preset or hints, and a build
without a resolved LANGUAGE is rejected (see BEHAVIOR: resolve, step 8).

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

Language := Go | Rust | C | CPP
```

---

## BEHAVIOR: resolve
Constraint: required

Given a spec declaring `Deployment: library`, a translator reads this
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
- template is the library template (Template-For = "library")
- spec_meta contains at least Deployment, Verification, Safety-Level

STEPS:
1. Verify Template-For = "library"; on mismatch → error, halt.
2. Merge preset layers in order: vendor → system → user → project
   (last writer wins).
3. For each constraint=required key K: if not resolved → errors +=
   violation.
4. For each constraint=default key K: apply preset value if present,
   else template default.
5. For each constraint=forbidden key K: if present in spec_meta or any
   preset → errors += violation.
6. For each constraint=supported key K: apply if declared in spec_meta
   or preset; skip silently if absent.
7. Apply LANGUAGE precedence: project preset > user preset > system
   preset > language hints file.
8. If LANGUAGE is not resolved after step 7 → errors +=
   "Deployment 'library' has no default language.
    Declare LANGUAGE via preset or a language hints file."
9. If errors non-empty → return errors (reject, do not return
   resolved). Else → return resolved.

POSTCONDITIONS:
- resolved.LANGUAGE is one of the supported Language values
- no resolved key carries a forbidden value

---

## TEMPLATE-TABLE

| Key | Value | Constraint | Notes |
|-----|-------|------------|-------|
| VERSION | MAJOR.MINOR.PATCH or YYYY.MM.DD.VV | required | Semantic versioning or dated versioning. Spec author increments on every meaningful change. |
| SPEC-SCHEMA | MAJOR.MINOR.PATCH | required | Version of the Post-Coding spec schema this file was written against. |
| AUTHOR | name <email> | required | At least one Author: line required. Repeating key; multiple authors permitted. |
| LICENSE | SPDX identifier | required | Must be a valid SPDX license identifier or compound expression. |
| LANGUAGE | Go | supported | Selected via preset or hints. Per-language deliverable details apply (see DELIVERABLES). No template default; see BEHAVIOR: resolve, step 8. |
| LANGUAGE | Rust | supported | Selected via preset or hints. |
| LANGUAGE | C | supported | Selected via preset or hints. |
| LANGUAGE | C++ | supported | Selected via preset or hints. |
| ARTIFACT-TYPE | language-native-library | required | The build produces the target language's natural library artifact: a Go module, a Rust library crate, a C or C++ static archive with headers. No executable. |
| ENTRY-POINT | none | forbidden | Libraries have no executable entry point and no CLI. A `main` function or equivalent is a constraint violation. |
| BINARY-COUNT | 0 | required | No binaries are produced. Test executables built by `make test` are transient build artifacts, not deliverables. |
| PUBLIC-API-SURFACE | stable-across-translations | required | The exported names and signatures form the public interface. It must remain stable across translations of the same spec at the same Version. Additions permitted; removals or renames require a spec Version increment. |
| PUBLIC-API-SURFACE | recorded-in-report | required | The translator records the public interface in `TRANSLATION_REPORT.md` under `## Public API Surface`, one entry per exported symbol with full signature, grouped by module. The next translation reads this section as input and verifies continuity. |
| SOURCE-PARTITIONING | modular | required | Implementation source partitioned into multiple modules per the target language's convention. A single monolithic source file is not permitted. |
| MODULE-IDENTITY | host-specified | required | Module identity from an authoritative source, in priority order: spec META `Module:` field, language hints file, existing manifest from a prior translation. Never inferred from repository guesswork. |
| MODULE-IDENTITY | propagated | required | The identity appears consistently across manifests, packaging metadata, and documentation. |
| MODULE-IDENTITY | conflict-halts | required | Conflicting authoritative sources halt translation with a diagnostic naming all sources and values. |
| CONSUMPTION | build-time-linkage | required | Consumers bind through their spec's DEPENDENCIES section (kind: pcd-artifact), pinning the library version and recording the library's merged spec hash in their TRANSLATION_REPORT. The build service builds the library from source and enforces version alignment across all consumers. |
| API-DOCS | generated | required | API reference generated by the target language's documentation convention from source-level documentation comments. |
| MAN-PAGES | none | supported | No section-1 man page (there is no executable). Section-3 API pages are optional per language convention. This template is a documented exception to the CLI man-page requirement. |
| RUNTIME-DEPS | none | required | No runtime dependencies beyond the target language's standard library, unless the spec's DEPENDENCIES section declares otherwise. |
| OUTPUT-FORMAT | RPM | required | Library and development packaging per the target language's convention on the build service. |
| OUTPUT-FORMAT | DEB | required | Library and development packaging per the target language's convention. |
| INSTALL-METHOD | OBS | required | Primary distribution via build.opensuse.org. curl-based install is forbidden. |
| INSTALL-METHOD | curl | forbidden | curl-based installation scripts are not permitted. Supply chain security requirement. |
| PLATFORM | Linux | required | Linux is the primary and required platform. |
| PLATFORM | macOS | supported | Optional; follows the consuming front ends. |
| CONFIG-ENV-VARS | forbidden | forbidden | Library behaviour must not be controlled via environment variables. |
| NETWORK-CALLS | forbidden | forbidden | The library must not make network calls. |
| PRESET-SYSTEM | systemd-style | required | Preset layering follows systemd conventions. See whitepaper A.11. |

---

## PRECONDITIONS

- This template is applied only when spec META declares
  Deployment: library
- Preset files must be valid TOML

---

## POSTCONDITIONS

- A resolved build declares exactly one LANGUAGE
- The produced artifact set contains no executable

---

## INVARIANTS

- [observable]      the library never controls behaviour via
  environment variables and never makes network calls
- [observable]      the public interface recorded in the translation
  report is a superset-compatible continuation of the previous
  translation's record at the same spec Version
- [implementation]  packaging follows the target language's library
  convention on the build service; no ad-hoc installation path exists

---

## EXAMPLES

### EXAMPLE: language_resolved_from_preset
GIVEN:
  spec META declares Deployment: library
  project preset declares LANGUAGE = Rust
WHEN:
  resolve runs
THEN:
  errors = (empty)
  resolved.LANGUAGE = Rust

### EXAMPLE: missing_language_rejected
GIVEN:
  spec META declares Deployment: library
  no preset layer and no hints file declares LANGUAGE
WHEN:
  resolve runs
THEN:
  errors contains one entry:
    message contains "no default language"
  resolved is not returned

### EXAMPLE: curl_install_rejected
GIVEN:
  a preset declares INSTALL-METHOD = curl
WHEN:
  resolve runs
THEN:
  errors contains one entry:
    message contains "curl"
  resolved is not returned

---

## DELIVERABLES

Defines the files a translator must produce. Delivery order:
1. Core implementation files (library source, module manifest,
   `VERSION` file, Makefile, README.md, LICENSE)
2. Packaging artifacts (RPM, DEB)
3. TRANSLATION_REPORT.md last, after all other files are written and
   verified

| Deliverable | Constraint | Required Files | Notes |
|---|---|---|---|
| source | required | Library modules plus the language's module manifest. Per-language layout below. | Per SOURCE-PARTITIONING: modular. No executable entry point (ENTRY-POINT: none). |
| public-api | required | `TRANSLATION_REPORT.md` section `## Public API Surface` | Per PUBLIC-API-SURFACE: recorded-in-report. |
| build | required | `Makefile` | Targets: `build`, `test`, `clean`, `dist`. `make test` is executable and exits non-zero on any test failure. `dist` produces the release source tarball (and vendor tarball for vendoring languages). |
| api-docs | required | per language convention | Generated from documentation comments; wired into `make` where the language toolchain supports it. |
| docs | required | `README.md` | Documents the public interface, the consumers, installation via the build service, and the version-alignment rule. |
| license | required | `LICENSE` | SPDX identifier from spec META + authoritative URL. Never reproduce the full license text. |
| RPM | required | `<n>.spec` | Library and development packaging per the target language's convention on OBS. |
| DEB | required | `debian/control`, `debian/changelog`, `debian/rules`, `debian/copyright` | Development-package layout per language convention; DEP-5 copyright. |
| report | required | `TRANSLATION_REPORT.md` | Includes the translation-input provenance block (see cli-tool.template.md for the canonical field list) and the Public API Surface section. Written last. |
| spec-hash | required | embedded in all artifacts | SHA256 of the merged spec text embedded in: source file header comments, `TRANSLATION_REPORT.md` `Spec-SHA256:` field, `Makefile` `SPEC_SHA256` variable, RPM `.spec` comment, DEB `control` `X-PCD-Spec-SHA256:` field. There is no binary `--version` carrier; source headers and packaging metadata are the artifact-side carriers. |

**Per-language source layout (`source` deliverable):**

| LANGUAGE | Implementation location | Manifest file |
|---|---|---|
| Go | packages under the module root (no `main` package) | `go.mod` |
| Rust | `src/lib.rs` plus modules under `src/` | `Cargo.toml` (`[lib]`) |
| C | one header and one source file per logical unit under `include/` and `src/` | `Makefile` |
| C++ | headers under `include/`, sources under `src/` | `Makefile` or `CMakeLists.txt` |

### Naming Convention

`<n>` refers to the component name as declared in the specification
title (first `#` heading): lowercase, hyphen-separated, no version
suffix in the filename.

---

## EXECUTION

### Input files

The spec (with Includes resolved), this template, the resolved
language's hints files, and the translator prompt.

### Delivery phases

1. Resolve template and language (BEHAVIOR: resolve); halt on errors.
2. Produce the `source`, `build`, `docs`, `api-docs`, and `license`
   deliverables.
3. Compile gate (below); on failure, fix within this phase before
   proceeding.
4. Produce packaging deliverables (RPM, DEB).
5. Write `TRANSLATION_REPORT.md` last.

### Compile gate

`make build && make test` from a clean checkout must succeed: the
library compiles and the full test suite passes, exiting non-zero on
any failure. For spec-defined golden outputs (byte-exact EXAMPLES), the
gate includes byte comparison.

### Resume logic

On resumption, verify which deliverables already exist on disk and are
consistent with the spec hash, then continue with the first incomplete
phase. Never regenerate completed phases silently; note resumption in
the translation report.
