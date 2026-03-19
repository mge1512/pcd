# python-tool.template

## META
Deployment:  template
Version:     0.3.7
Spec-Schema: 0.3.7
Author:      Matthias G. Eckermann <pcdp@mailbox.org>
License:     CC-BY-4.0
Verification: none
Safety-Level: QM
Template-For: python-tool

---

> **Status: Work in Progress**
> This template is planned for v0.3.8. The definition below is a stub
> capturing agreed design decisions. It is not yet complete enough for
> production use. Use `manual` deployment type with explicit META fields
> until this template is finalised.

---

## TYPES

```
Language := Python
// Python is the only valid language for python-tool.
// No alternatives permitted.

PythonMinVersion := "3.9" | "3.10" | "3.11" | "3.12" | "3.13"
// Minimum Python version must be declared in DEPLOYMENT section.

DistributionFormat := wheel | sdist
// Both are required deliverables.
```

---

## IMPORTANT CONSTRAINTS

**python-tool is QM safety level only.**
Python is not suitable for safety-critical components. The absence of
a formal verification path, the dynamic type system, and the interpreted
execution model make Python incompatible with ISO 26262, DO-178C,
IEC 62304, and Common Criteria certification requirements.

If you need a Python-like developer experience for a safety-critical
component, consider:
- `cli-tool` with Go (statically typed, compiled, formally verifiable)
- `verified-library` with C (for components requiring formal proof)

**python-tool is for:**
- Development tooling and automation
- Data pipelines and analysis scripts
- Test infrastructure
- Glue code and integration scripts
- Domain-specific tools where Python ecosystem libraries are required

---

## TEMPLATE-TABLE

| Key | Value | Constraint | Notes |
|-----|-------|------------|-------|
| VERSION | MAJOR.MINOR.PATCH | required | |
| SPEC-SCHEMA | MAJOR.MINOR.PATCH | required | |
| AUTHOR | name <email> | required | Repeating field permitted. |
| LICENSE | SPDX identifier | required | |
| LANGUAGE | Python | required | No alternatives. |
| VERIFICATION | none | required | Formal verification path does not exist for Python. |
| SAFETY-LEVEL | QM | required | No other safety level permitted. |
| PYTHON-MIN-VERSION | 3.11 | default | Minimum Python version. Override via preset. |
| TYPE-CHECKING | mypy | default | mypy or pyright. Replaces compile-time type safety. |
| RUNTIME-DEPS | declared in pyproject.toml | required | All dependencies pinned in pyproject.toml. |
| CLI-ARG-STYLE | key=value | required | Consistent with pcdp conventions. |
| NETWORK-CALLS | context-dependent | supported | Declare explicitly in DEPLOYMENT section if used. |
| OUTPUT-FORMAT | wheel | required | Python wheel (.whl) for pip installation. |
| OUTPUT-FORMAT | sdist | required | Source distribution for audit and OBS builds. |
| OUTPUT-FORMAT | RPM | required | OBS RPM package wrapping the wheel. |
| OUTPUT-FORMAT | DEB | required | OBS DEB package wrapping the wheel. |
| INSTALL-METHOD | OBS | required | Primary distribution via OBS. |
| INSTALL-METHOD | pip | supported | pip install from wheel. Acceptable for tooling. |
| INSTALL-METHOD | curl | forbidden | Supply chain security requirement. |
| PYPROJECT-TOML | required | required | pyproject.toml with full metadata and pinned deps. |
| IDEMPOTENT | true | required | Running tool twice on same input produces identical output. |

---

## DELIVERABLES

*(Full DELIVERABLES section pending — to be completed in v0.3.8)*

Required deliverables will include:
- `<n>/` — Python package directory with `__init__.py`
- `<n>/__main__.py` — entry point for `python -m <n>`
- `pyproject.toml` — build metadata, dependencies, entry points
- `README.md` — installation (zypper/apt/dnf/pip), usage, examples
- `LICENSE` — full license text
- RPM spec, Debian package files
- `TRANSLATION_REPORT.md`

---

## PRECONDITIONS

- Safety-Level must be QM — any other value is rejected by pcdp-lint
- Verification must be none — any other value is rejected by pcdp-lint
- pyproject.toml is a required deliverable
- Python minimum version must be declared in DEPLOYMENT section

---

## POSTCONDITIONS

*(Pending — to be completed in v0.3.8)*

---

## INVARIANTS

- GLOBAL: Safety-Level anything other than QM is rejected at pcdp-lint time
- GLOBAL: Verification anything other than none is rejected at pcdp-lint time
- GLOBAL: python-tool components may not be used in safety-critical systems
- GLOBAL: template version is recorded in every audit bundle

---

## EXAMPLES

*(Pending — to be completed in v0.3.8)*
*(Reference example: a pcdp-validate tool that runs pcdp-lint and
  spec-validate as a Python wrapper)*

---

## DEPLOYMENT

Runtime: this file is a template specification, not executable code.
Location: /usr/share/pcdp/templates/python-tool.template.md
Status: Work in progress — v0.3.8 target for completion.
Note: python-tool is QM only. Not suitable for safety-critical components.
