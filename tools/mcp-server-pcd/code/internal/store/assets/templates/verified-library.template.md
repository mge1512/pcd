
# verified-library.template

## META
Deployment:  template
Version:     0.3.20
Spec-Schema: 0.3.20
Author:      Matthias G. Eckermann <pcd@mailbox.org>
License:     CC-BY-4.0
Verification: none
Safety-Level: QM
Template-For: verified-library
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

VerificationLevel := lean4 | fstar | dafny | custom
// Verification: none is NOT permitted for verified-library.
// Minimum: lean4 or equivalent for EAL4+/ASIL-C/D/DAL-A/B contexts.

SafetyLevel := one_of(
  "ASIL-A" | "ASIL-B" | "ASIL-C" | "ASIL-D" |
  "DAL-A" | "DAL-B" | "DAL-C" | "DAL-D" |
  "IEC-62443-SL1" | "IEC-62443-SL2" | "IEC-62443-SL3" | "IEC-62443-SL4" |
  "EAL2" | "EAL3" | "EAL4" | "EAL4+" | "EAL5" | "EAL6" | "EAL7"
)
// Safety-Level: QM is NOT permitted for verified-library.
// Use library-c-abi for general-purpose C libraries without
// safety or security certification requirements.
```

---

## TEMPLATE-TABLE

| Key | Value | Constraint | Notes |
|-----|-------|------------|-------|
| VERSION | MAJOR.MINOR.PATCH | required | |
| SPEC-SCHEMA | MAJOR.MINOR.PATCH | required | |
| AUTHOR | name <email> | required | Repeating field permitted. |
| LICENSE | SPDX identifier | required | |
| LANGUAGE | C | default | Primary language for verified-library. |
| LANGUAGE-ALTERNATIVES | Rust | supported | Via preset override only. |
| VERIFICATION | lean4 | default | Formal verification required. none is forbidden. |
| VERIFICATION-ALTERNATIVES | fstar | supported | F* for effect tracking and SMT-based verification. |
| VERIFICATION-ALTERNATIVES | dafny | supported | Dafny for accessible SMT verification. |
| SAFETY-LEVEL | (must declare) | required | QM is forbidden. See SafetyLevel type above. |
| BINARY-TYPE | static | required | Static library (.a) or shared (.so). No executables. |
| CONSTANT-TIME | required | required | No secret-dependent branching or memory access. |
| SIDE-CHANNEL-RESISTANCE | required | required | Timing, cache, power side-channels must be mitigated. |
| KEY-ZEROIZATION | required | required | All key material zeroized after use. |
| MEMORY-SAFETY | required | required | Guaranteed by meta-language type system. |
| NETWORK-CALLS | forbidden | forbidden | No network access in library code. |
| DYNAMIC-ALLOCATION | context-dependent | supported | Must be formally bounded or forbidden per standard. |
| OUTPUT-FORMAT | static-library | required | .a archive for static linking. |
| OUTPUT-FORMAT | shared-library | supported | .so for dynamic linking. Include .pc and .cps. |
| OUTPUT-FORMAT | RPM | required | OBS RPM package. |
| OUTPUT-FORMAT | DEB | required | OBS DEB package. |
| INSTALL-METHOD | OBS | required | No curl-based installation. |
| INSTALL-METHOD | curl | forbidden | Supply chain security requirement. |

---

## DELIVERABLES

*(Full DELIVERABLES section pending — to be completed in v0.3.8)*

Required deliverables will include:
- Source files in target language
- Meta-language IR (Lean 4 / F* / Dafny) — AI-generated, not human-authored
- Formal proof artifacts
- Test vectors (from applicable standard: NIST CAVP, ISO, etc.)
- `<n>.pc` pkg-config file
- `<n>.cps` CPS file (CMake 4.3+)
- `<n>.3.md` — man page source (Markdown); `<n>.3` — generated via `pandoc`
  (section 3: library functions). `BuildRequires: pandoc` in RPM spec;
  `pandoc` in DEB `Build-Depends`. Install to `%{_mandir}/man3/` (RPM)
  and `usr/share/man/man3/` (DEB).
- RPM spec, Debian package files
- `TRANSLATION_REPORT.md`

---

## BEHAVIOR: resolve
Constraint: required

Given a spec declaring `Deployment: verified-library`, validate that
safety and verification constraints are met and derive the build configuration.

INPUTS:
```
spec_meta: Map<string, string>
preset:    Map<string, string>
```

STEPS:
1. Verify Template-For = "verified-library"; on mismatch → error, halt.
2. Check Safety-Level ≠ QM; on violation →
   error: "Safety-Level: QM is forbidden for Deployment: verified-library".
3. Check Verification ≠ none; on violation →
   error: "Verification: none is forbidden for Deployment: verified-library".
4. Merge preset layers: vendor → system → user → project (last writer wins).
5. For each constraint=required key: if not resolved → errors += violation.
6. For each constraint=forbidden key: if present → errors += violation.
7. If errors non-empty → return errors; else return resolved map.

POSTCONDITIONS:
- resolved contains an effective value for every required key
- Safety-Level is not QM
- Verification is one of: lean4, fstar, dafny, custom
- curl is never an accepted install method

ERRORS:
- ERR_QM_SAFETY    if Safety-Level = QM
- ERR_NO_VERIFY    if Verification = none
- ERR_FORBIDDEN    if a forbidden key is present
- ERR_MISSING      if a required key is absent after preset merge

---

## PRECONDITIONS

- Safety-Level must not be QM
- Verification must not be none
- Spec must include a SECURITY-PROPERTIES section declaring:
  Constant-Time, Side-Channel-Resistance, Key-Zeroization requirements
- Applicable standard must be declared (FIPS-140-3, NIST, RFC, ISO 26262, etc.)

---

## POSTCONDITIONS

*(Pending — to be completed in v0.3.8)*

---

## INVARIANTS

- [observable]      Safety-Level QM is rejected at spec-lint time
- [observable]      Verification: none is rejected at spec-lint time
- [observable]      verified-library inherits all library-c-abi constraints
  and adds security/safety constraints on top
- [observable]      template version is recorded in every audit bundle

---

## EXAMPLES

EXAMPLE: valid_asil_b_lean4
GIVEN:
  spec META declares:
    Deployment: verified-library
    Safety-Level: ASIL-B
    Verification: lean4
  no preset overrides
WHEN:
  resolve runs
THEN:
  resolved["LANGUAGE"] = "C"
  resolved["VERIFICATION"] = "lean4"
  errors = []

EXAMPLE: qm_safety_level_rejected
GIVEN:
  spec META declares:
    Deployment: verified-library
    Safety-Level: QM
    Verification: lean4
WHEN:
  resolve runs
THEN:
  errors contains: "Safety-Level: QM is forbidden for Deployment: verified-library"
  resolved is not produced

EXAMPLE: no_verification_rejected
GIVEN:
  spec META declares:
    Deployment: verified-library
    Safety-Level: ASIL-A
    Verification: none
WHEN:
  resolve runs
THEN:
  errors contains: "Verification: none is forbidden for Deployment: verified-library"
  resolved is not produced

---

## DEPLOYMENT

Runtime: this file is a template specification, not executable code.
Location: /usr/share/pcd/templates/verified-library.template.md
Status: Work in progress — v0.3.8 target for completion.
Supersedes: crypto-library template (retired in v0.3.6).

