# verified-library.template

## META
Deployment:  template
Version:     0.3.7
Spec-Schema: 0.3.7
Author:      Matthias G. Eckermann <pcdp@mailbox.org>
License:     CC-BY-4.0
Verification: none
Safety-Level: QM
Template-For: verified-library

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
- RPM spec, Debian package files
- `TRANSLATION_REPORT.md`

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

- GLOBAL: Safety-Level QM is rejected at spec-lint time
- GLOBAL: Verification: none is rejected at spec-lint time
- GLOBAL: verified-library inherits all library-c-abi constraints
  and adds security/safety constraints on top
- GLOBAL: template version is recorded in every audit bundle

---

## EXAMPLES

*(Pending — to be completed in v0.3.8)*
*(Reference examples: AES-256-GCM, ECDSA P-256, SHA-3, mutex primitive)*

---

## DEPLOYMENT

Runtime: this file is a template specification, not executable code.
Location: /usr/share/pcdp/templates/verified-library.template.md
Status: Work in progress — v0.3.8 target for completion.
Supersedes: crypto-library template (retired in v0.3.6).
