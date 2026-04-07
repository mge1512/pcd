
# library-c-abi.template

## META
Deployment:  template
Version:     0.3.20
Spec-Schema: 0.3.20
Author:      Matthias G. Eckermann <pcd@mailbox.org>
License:     CC-BY-4.0
Verification: none
Safety-Level: QM
Template-For: library-c-abi
EXECUTION:   none

---

> **Status: Work in Progress**
> This template is planned for v0.3.8. The definition below is a stub
> capturing agreed design decisions. It is not yet complete enough for
> production use. Use `manual` deployment type with explicit META fields
> until this template is finalised.

---

## TYPES

```
Language := C | Rust
// Rust via cbindgen for C-ABI compatible output.

ABIStability := stable | unstable
// stable: ABI must not break across versions (soname versioning required)
// unstable: internal library, ABI may change between releases
```

---

## TEMPLATE-TABLE

| Key | Value | Constraint | Notes |
|-----|-------|------------|-------|
| VERSION | MAJOR.MINOR.PATCH | required | |
| SPEC-SCHEMA | MAJOR.MINOR.PATCH | required | |
| AUTHOR | name <email> | required | Repeating field permitted. |
| LICENSE | SPDX identifier | required | |
| LANGUAGE | C | default | Primary language for C-ABI libraries. |
| LANGUAGE-ALTERNATIVES | Rust | supported | Via cbindgen. Preset override required. |
| VERIFICATION | none | default | Formal verification optional. |
| VERIFICATION-ALTERNATIVES | lean4 | supported | For higher-assurance general libraries. |
| ABI-STABILITY | stable | default | Soname versioning required for stable ABI. |
| BINARY-TYPE | static | default | .a archive preferred. |
| BINARY-TYPE | dynamic | supported | .so with soname versioning for stable ABI. |
| NETWORK-CALLS | forbidden | forbidden | No network access in library code. |
| OUTPUT-FORMAT | static-library | required | .a archive. |
| OUTPUT-FORMAT | shared-library | supported | .so with versioned soname. |
| OUTPUT-FORMAT | pkg-config | required | `<n>.pc` for pkg-config consumers. |
| OUTPUT-FORMAT | cps | required | `<n>.cps` for CMake 4.3+ consumers. |
| OUTPUT-FORMAT | RPM | required | OBS RPM package. |
| OUTPUT-FORMAT | DEB | required | OBS DEB package. |
| INSTALL-METHOD | OBS | required | No curl-based installation. |
| INSTALL-METHOD | curl | forbidden | Supply chain security requirement. |

---

## DELIVERABLES

*(Full DELIVERABLES section pending — to be completed in v0.3.8)*

Required deliverables will include:
- Source files (`lib<n>.c` / `lib<n>.h` or Rust with cbindgen)
- `<n>.pc` — pkg-config descriptor
- `<n>.cps` — CPS descriptor (CMake 4.3+, JSON format)
- `<n>.h` — public C header
- `<n>.3.md` — man page source (Markdown); `<n>.3` — generated via `pandoc`
  (section 3: library functions). `BuildRequires: pandoc` in RPM spec;
  `pandoc` in DEB `Build-Depends`. Install to `%{_mandir}/man3/` (RPM)
  and `usr/share/man/man3/` (DEB).
- RPM spec, Debian package files
- `TRANSLATION_REPORT.md`

### CPS File Note

As of CMake 4.3, CPS (Common Package Specification) files are required
for CMake ecosystem consumers. The `.cps` file is a JSON artifact
describing artifacts, components, headers, and link requirements in a
vendor-neutral format. It is a required deliverable for all library-c-abi
components. Reference: https://cps-org.github.io/cps/

---

## BEHAVIOR: resolve
Constraint: required

Given a spec declaring `Deployment: library-c-abi`, derive the effective
build configuration from template defaults and preset overrides.

INPUTS:
```
spec_meta: Map<string, string>
preset:    Map<string, string>
```

STEPS:
1. Verify Template-For = "library-c-abi"; on mismatch → error, halt.
2. Merge preset layers: vendor → system → user → project (last writer wins).
3. For each constraint=required key: if not resolved → errors += violation.
4. For each constraint=forbidden key: if present → errors += violation.
5. If errors non-empty → return errors; else return resolved map.

POSTCONDITIONS:
- resolved contains an effective value for every required key
- LANGUAGE is one of: C, Rust
- curl is never an accepted install method

ERRORS:
- ERR_FORBIDDEN if a forbidden key (e.g. INSTALL-METHOD=curl) is present
- ERR_MISSING   if a required key is absent after preset merge

---

## PRECONDITIONS

- Public header file must be declared in spec TYPES section
- ABI stability level must be declared (stable | unstable)
- If ABI-STABILITY = stable: soname versioning required in shared library

---

## POSTCONDITIONS

*(Pending — to be completed in v0.3.8)*

---

## INVARIANTS

- [observable]      library-c-abi is for general-purpose C-ABI libraries
- [observable]      for safety/security-critical libraries use verified-library
- [observable]      CPS file is always a required deliverable
- [observable]      template version is recorded in every audit bundle

---

## EXAMPLES

EXAMPLE: default_c_static_library
GIVEN:
  spec META declares Deployment: library-c-abi
  no preset overrides
WHEN:
  resolve runs
THEN:
  resolved["LANGUAGE"] = "C"
  resolved["BINARY-TYPE"] = "static"
  resolved["ABI-STABILITY"] = "stable"
  errors = []

EXAMPLE: forbidden_curl_rejected
GIVEN:
  spec META declares Deployment: library-c-abi
  preset declares INSTALL-METHOD = curl
WHEN:
  resolve runs
THEN:
  errors contains: "Key INSTALL-METHOD=curl is forbidden for Deployment: library-c-abi"
  resolved is not produced

---

## DEPLOYMENT

Runtime: this file is a template specification, not executable code.
Location: /usr/share/pcd/templates/library-c-abi.template.md
Status: Work in progress — v0.3.8 target for completion.
Note: For safety/security-critical C libraries, use verified-library instead.

