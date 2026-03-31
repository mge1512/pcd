
# python-tool.template

## META
Deployment:  template
Version:     0.3.19
Spec-Schema: 0.3.19
Author:      Matthias G. Eckermann <pcd@mailbox.org>
License:     CC-BY-4.0
Verification: none
Safety-Level: QM
Template-For: python-tool

---

## TYPES

```
Language := Python
// Python is the only language for python-tool. No alternatives.
// Safety-Level: QM mandatory. Verification: none mandatory.

PythonVersion := string where matches "^3\.[0-9]+$"
// Minimum supported Python version. Default: 3.11.

OutputFormat := RPM | DEB | wheel | OCI
// RPM:   Linux RPM via OBS. Required.
// DEB:   Linux DEB via OBS. Required.
// wheel: Python wheel (.whl) for PyPI or internal distribution. Supported.
// OCI:   Container image. Supported.

PackageName := string where matches "^[a-z][a-z0-9_]*$"
// Python package name: hyphens in TOOL_NAME replaced with underscores.
```

---

## TEMPLATE-TABLE

| Key | Value | Constraint | Notes |
|-----|-------|------------|-------|
| VERSION | MAJOR.MINOR.PATCH | required | Semantic versioning. |
| SPEC-SCHEMA | MAJOR.MINOR.PATCH | required | PCD schema version. |
| AUTHOR | name <email> | required | Repeating field permitted. |
| LICENSE | SPDX identifier | required | Valid SPDX identifier. |
| LANGUAGE | Python | required | Fixed. No alternatives for python-tool. |
| PYTHON-VERSION | 3.11 | default | Minimum Python version. Override via preset. |
| SAFETY-LEVEL | QM | required | python-tool is QM only. Higher safety levels forbidden. |
| VERIFICATION | none | required | Formal verification forbidden for python-tool. |
| BINARY-TYPE | wheel | required | Installed as a Python wheel. No compiled binary. |
| CLI-ARG-STYLE | POSIX | required | Standard argparse --flag style. key=value style forbidden. |
| LAYOUT | src | required | src/ layout mandatory. Flat layout forbidden. |
| TYPECHECK | mypy-strict | required | mypy --strict must pass with zero errors. |
| LINTING | flake8 | required | flake8 must pass with zero warnings. |
| FORMAT | black | required | black formatting enforced. |
| TESTING | pytest | required | pytest + hypothesis for property-based tests. |
| BUILD-TOOL | uv | default | uv for dependency management and builds. |
| OUTPUT-FORMAT | RPM | required | OBS RPM package. noarch. |
| OUTPUT-FORMAT | DEB | required | OBS DEB package. |
| OUTPUT-FORMAT | wheel | supported | Python wheel for PyPI or internal distribution. |
| OUTPUT-FORMAT | OCI | supported | OCI container image. Base: registry.opensuse.org/opensuse/leap current. |
| INSTALL-METHOD | OBS | required | Primary Linux distribution via build.opensuse.org. |
| INSTALL-METHOD | curl | forbidden | Supply chain security requirement. |
| INSTALL-METHOD | pip-direct | forbidden | Direct pip install from URL forbidden. Use OBS packages. |
| SPDX-HEADERS | required | required | Every .py source file must carry SPDX-License-Identifier and SPDX-FileCopyrightText headers. |
| CONFIG-ENV-VARS | forbidden | forbidden | Behaviour must not be controlled via environment variables. |
| NETWORK-CALLS | forbidden | forbidden | Tool must not make network calls at runtime. |
| IDEMPOTENT | true | required | Running the tool twice on the same input produces identical output. |

---

## BEHAVIOR: resolve
Constraint: required

Given a spec declaring `Deployment: python-tool`, validate constraints
and derive the effective build configuration.

INPUTS:
```
spec_meta: Map<string, string>
preset:    Map<string, string>
```

STEPS:
1. Verify Template-For = "python-tool"; on mismatch → error, halt.
2. Check Safety-Level = QM; if not → error: "python-tool requires Safety-Level: QM".
3. Check Verification = none; if not → error: "python-tool requires Verification: none".
4. Merge preset layers: vendor → system → user → project (last writer wins).
5. For each constraint=required key: if not resolved → errors += violation.
6. For each constraint=forbidden key: if present in spec_meta or preset → errors += violation.
7. If errors non-empty → return errors; else return resolved map.

POSTCONDITIONS:
- resolved["LANGUAGE"] = "Python"
- resolved["SAFETY-LEVEL"] = "QM"
- resolved["VERIFICATION"] = "none"
- curl and pip-direct are never accepted install methods
- errors is empty iff all required keys are resolved and no forbidden keys present

ERRORS:
- ERR_SAFETY    if Safety-Level ≠ QM
- ERR_VERIFY    if Verification ≠ none
- ERR_FORBIDDEN if a forbidden key is present
- ERR_MISSING   if a required key is absent after preset merge

---

## BEHAVIOR/INTERNAL: precedence-resolution
Constraint: required

Defines how conflicting values across preset layers are resolved.

STEPS:
1. Start with template defaults as the base map.
2. Merge /usr/share/pcd/presets/ values; later entries override earlier.
3. Merge /etc/pcd/presets/ values; overrides vendor defaults.
4. Merge ~/.config/pcd/presets/ values; overrides system.
5. Merge ./.pcd/presets/ values; overrides user.
6. For each constraint=forbidden key present in any layer → emit Error.
7. Return merged map.

---

## PRECONDITIONS

- Spec META must declare Safety-Level: QM
- Spec META must declare Verification: none
- Tool must not make network calls at runtime
- All .py source files must carry SPDX license and copyright headers

---

## POSTCONDITIONS

- Every spec using Deployment: python-tool is governed by this template
- LANGUAGE is always Python — no override permitted
- curl and pip-direct are never accepted install methods, regardless of preset
- Generated code carries SPDX headers in every .py file

---

## INVARIANTS

- [observable]   Safety-Level QM is the only permitted value; higher levels rejected
- [observable]   Verification: none is the only permitted value
- [observable]   curl and pip-direct install methods are rejected at all preset layers
- [observable]   LANGUAGE is always Python; no alternative is ever resolved
- [observable]   every .py source file carries SPDX-License-Identifier header
- [observable]   RPM and DEB are required output formats
- [implementation] src/ layout is enforced; flat layout never generated
- [implementation] cli.py contains only argparse logic; business logic in separate modules
- [implementation] mypy --strict passes with zero errors before delivery
- [implementation] flake8 passes with zero warnings before delivery

---

## EXAMPLES

EXAMPLE: default_resolution
GIVEN:
  spec META declares:
    Deployment: python-tool
    Safety-Level: QM
    Verification: none
  no preset overrides
WHEN:
  resolve runs
THEN:
  resolved["LANGUAGE"] = "Python"
  resolved["PYTHON-VERSION"] = "3.11"
  resolved["BUILD-TOOL"] = "uv"
  errors = []

EXAMPLE: non_qm_safety_rejected
GIVEN:
  spec META declares:
    Deployment: python-tool
    Safety-Level: ASIL-A
    Verification: none
WHEN:
  resolve runs
THEN:
  errors contains: "python-tool requires Safety-Level: QM"
  resolved is not produced

EXAMPLE: forbidden_curl_rejected
GIVEN:
  spec META declares Deployment: python-tool
  preset declares INSTALL-METHOD = curl
WHEN:
  resolve runs
THEN:
  errors contains: "Key INSTALL-METHOD=curl is forbidden for Deployment: python-tool"
  resolved is not produced

---

## DELIVERABLES

### Project Structure

The translator must produce a src/ layout with the following structure:

```
{TOOL_NAME}/
├── pyproject.toml
├── LICENSE
├── README.md
├── src/
│   └── {PACKAGE_NAME}/
│       ├── __init__.py          ← SPDX header + __version__
│       ├── __main__.py          ← enables: python -m {PACKAGE_NAME}
│       ├── cli.py               ← argparse only; no business logic
│       └── {CORE_MODULE}.py     ← business logic; no CLI concerns
├── tests/
│   ├── __init__.py
│   ├── test_{CORE_MODULE}.py    ← pytest + hypothesis
│   └── test_cli.py
├── packaging/
│   ├── {TOOL_NAME}.spec         ← RPM spec for OBS
│   └── debian/
│       ├── control
│       ├── changelog
│       ├── rules
│       └── copyright            ← DEP-5 format
├── Containerfile                ← OCI build (if OCI active in preset)
└── Makefile
```

### Deliverables Table

| OUTPUT-FORMAT | Constraint | Required Files | Notes |
|---|---|---|---|
| source | required | `src/`, `pyproject.toml` | src/ layout mandatory. SPDX headers in every .py file. |
| build | required | `Makefile` | Targets: `lint`, `typecheck`, `test`, `build`, `clean`. |
| docs | required | `README.md` | Installation via OBS (zypper/apt/dnf). Must not document curl-based installation. |
| license | required | `LICENSE` | SPDX identifier from spec META + authoritative URL. Never reproduce the full license text. |
| RPM | required | `packaging/{TOOL_NAME}.spec` | OBS RPM spec. BuildArch: noarch. pip install --no-deps in %install. |
| DEB | required | `packaging/debian/control`, `packaging/debian/changelog`, `packaging/debian/rules`, `packaging/debian/copyright` | DEP-5 copyright. dh-python build sequence. |
| wheel | supported | built by `uv build` | Produced by build step; not a static file. |
| OCI | supported | `Containerfile` | Base image: `registry.opensuse.org/opensuse/leap:15.6` or current release. Never Alpine, never Debian. pip3 install --no-cache-dir. ENTRYPOINT as exec form. |
| report | required | `TRANSLATION_REPORT.md` | Last deliverable. Must include: toolchain gate result (uv sync, flake8, mypy, pytest). |

### Deliverable Content Requirements

**pyproject.toml:**
- `[build-system]` must use hatchling: `requires = ["hatchling"]`
- `license = { text = "{LICENSE}" }` (PEP 639 old style for 3.11 compat)
- Dev dependencies in `[dependency-groups]` (PEP 735), not in `[project.optional-dependencies]`
- `[project.scripts]` entry: `{TOOL_NAME} = "{PACKAGE_NAME}.cli:main"`

**Every .py source file:**
- First two lines must be SPDX headers:
  ```
  # SPDX-License-Identifier: {LICENSE}
  # SPDX-FileCopyrightText: {YEAR} {AUTHOR_NAME} <{AUTHOR_EMAIL}>
  ```

**RPM spec:**
- `BuildArch: noarch`
- `%install` must use `pip install --no-deps --root %{buildroot}`
- `%files` must use `%{python3_sitelib}`, not `%{python3_sitearch}`

**Containerfile:**
- Base image: `registry.opensuse.org/opensuse/leap:15.6` (or current Leap release)
- Install Python via: `zypper --non-interactive install --no-recommends python311`
- Install wheel via: `pip3 install --no-cache-dir {TOOL_NAME}-{VERSION}-py3-none-any.whl`
- `ENTRYPOINT ["{TOOL_NAME}"]` — exec form, never shell form

---

## DEPLOYMENT

Runtime: this file is a template specification, not executable code.
Location: /usr/share/pcd/templates/python-tool.template.md

---

## EXECUTION

### Input files

The translator receives in the working directory:
- `python-tool.template.md` — this deployment template
- `<spec-name>.md` — the component specification

### Resume logic

Before writing any file, list the output directory.
If a listed deliverable already exists and is non-empty, skip it.
Report which files were found and which are being produced.

### Delivery phases

**Phase 1 — Core source**
- `src/{PACKAGE_NAME}/__init__.py`
- `src/{PACKAGE_NAME}/__main__.py`
- `src/{PACKAGE_NAME}/cli.py`
- `src/{PACKAGE_NAME}/{CORE_MODULE}.py`
- `pyproject.toml`

**Phase 2 — Tests**
- `tests/__init__.py`
- `tests/test_{CORE_MODULE}.py`
- `tests/test_cli.py`

**Phase 3 — Build and packaging**
- `Makefile`
- `packaging/{TOOL_NAME}.spec`
- `packaging/debian/control`, `changelog`, `rules`, `copyright`
- `Containerfile` (if OCI active in preset)
- `LICENSE`

**Phase 4 — Documentation**
- `README.md`

**Phase 5 — Toolchain gate**

Run in order. If environment cannot execute shell commands, document
explicitly in TRANSLATION_REPORT.md.

```
uv sync
uv run flake8 src/ tests/
uv run mypy src/
uv run pytest tests/ -v
uv build
```

Record pass/fail for each command. Proceed to Phase 6 only after all pass
or after explicitly documenting why a step could not be executed.

**Phase 6 — Report (last)**
- `TRANSLATION_REPORT.md`
