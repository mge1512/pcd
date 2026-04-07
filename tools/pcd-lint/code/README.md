# pcd-lint

Post-Coding Development specification linter.

**Version:** 0.3.21 | **License:** GPL-2.0-only | **Schema:** 0.3.21

---

## Overview

`pcd-lint` validates Post-Coding Development (PCD) specification files
against the structural rules defined in the PCD spec schema. It checks:

- Required sections (META, TYPES, BEHAVIOR, PRECONDITIONS, POSTCONDITIONS, INVARIANTS, EXAMPLES)
- META field completeness and format (SPDX license, semantic versioning)
- BEHAVIOR block structure (STEPS required, Constraint: field validation)
- EXAMPLES structure (GIVEN/WHEN/THEN completeness, multi-pass support)
- INVARIANTS tagging ([observable] / [implementation])
- MILESTONE sections (structure, status, BEHAVIOR name existence)
- Negative-path EXAMPLE coverage for BEHAVIOR blocks with error exits

All rules are evaluated — linting does not stop at the first error.

---

## Installation

`pcd-lint` is distributed via [OBS (build.opensuse.org)](https://build.opensuse.org)
as part of the **pcd-tools** package.

**openSUSE / SUSE Linux Enterprise:**
```
zypper install pcd-tools
```

**Fedora:**
```
dnf install pcd-tools
```

**Debian / Ubuntu:**
```
apt install pcd-tools
```

The **pcd-templates** package is required and provides template files to
`/usr/share/pcd/templates/`.

> **Note:** No curl-based installation is supported (supply chain security requirement).

---

## Usage

```
pcd-lint [strict=true] <specfile.md>
pcd-lint list-templates
pcd-lint version
```

### Options

| Option | Description | Default |
|--------|-------------|---------|
| `strict=true` | Treat warnings as errors; exit 1 on any warning | `false` |

### Commands

| Command | Description |
|---------|-------------|
| `list-templates` | Print all known deployment templates with default language annotations |
| `version` | Print pcd-lint version, schema version, and SPDX list version |

---

## Exit Codes

| Code | Meaning |
|------|---------|
| `0` | Valid specification (no errors; no warnings unless strict=false) |
| `1` | Invalid specification (errors present; or strict=true with warnings) |
| `2` | Invocation error (bad arguments, file not found, wrong extension) |

---

## Output Streams

- **stderr**: Diagnostic lines (errors and warnings)
- **stdout**: Summary line (lint) or template list (list-templates)

### Diagnostic Format

```
{SEVERITY}  {file}:{line}  [{section}]  {message}
```

Examples:
```
ERROR    spec.md:1    [structure]  Missing required section: ## INVARIANTS
ERROR    spec.md:4    [META]       Missing required META field: Deployment
WARNING  spec.md:6    [META]       META field 'Target' is deprecated since v0.3.0
ERROR    spec.md:42   [EXAMPLES]   Example 'foo' missing THEN: marker
```

### Summary Format

```
✓ spec.md: valid
✓ spec.md: valid (N warning(s))
✗ spec.md: N error(s), M warning(s)
✗ spec.md: N error(s), M warning(s) [strict mode]
```

---

## Examples

**Lint a specification file:**
```
pcd-lint mycomponent.md
```

**Lint with strict mode:**
```
pcd-lint strict=true mycomponent.md
```

**List all known deployment templates:**
```
pcd-lint list-templates
```
Output (example):
```
wasm  →  (template file not found)
ebpf  →  (template file not found)
cli-tool  →  Go
...
```

**Print version:**
```
pcd-lint version
# pcd-lint 0.3.21 (schema 0.3.21) spdx/3.24.0
```

---

## Template Search Path

Template files are searched in ascending precedence order (last wins):

1. `/usr/share/pcd/templates/` — vendor default (pcd-templates package)
2. `/etc/pcd/templates/` — system administrator
3. `~/.config/pcd/templates/` — user
4. `./.pcd/templates/` — project-local

---

## Validation Rules Summary

| Rule | Description | Severity |
|------|-------------|----------|
| RULE-01 | Required sections present | Error |
| RULE-02 | META fields present and non-empty | Error |
| RULE-02b | Author field required | Error |
| RULE-02c | Version: MAJOR.MINOR.PATCH format | Error |
| RULE-02d | Spec-Schema: MAJOR.MINOR.PATCH format | Error |
| RULE-02e | License: valid SPDX identifier | Error |
| RULE-03 | Deployment: known template | Error |
| RULE-04 | Deprecated META fields (Target, Domain) | Warning |
| RULE-05 | Verification: known value | Warning |
| RULE-06 | EXAMPLES block structure (GIVEN/WHEN/THEN) | Error |
| RULE-07 | EXAMPLES block content non-empty | Warning |
| RULE-08 | BEHAVIOR blocks contain STEPS | Error |
| RULE-09 | INVARIANTS entries carry tags | Warning |
| RULE-10 | Negative-path EXAMPLE for BEHAVIOR with error exits | Error |
| RULE-11 | TOOLCHAIN-CONSTRAINTS structure | Warning |
| RULE-12 | Cross-section consistency | Error/Warning |
| RULE-13 | Constraint: field valid values | Error/Warning |
| RULE-14 | EXECUTION section in deployment templates | Warning |
| RULE-15 | MILESTONE structure and single-active | Error/Warning |
| RULE-16 | MILESTONE BEHAVIOR names exist | Error |
| RULE-17 | Scaffold milestone ordering and uniqueness | Error |

---

## Building from Source

```
git clone <repository>
cd pcd-lint
make build
make man
make install
```

Requirements: Go 1.21+, pandoc (for man page generation).

---

## License

GPL-2.0-only — see [https://spdx.org/licenses/GPL-2.0-only.html](https://spdx.org/licenses/GPL-2.0-only.html)
