# pcdp-lint

A command-line linter and validator for Post-Coding Development Paradigm (PCDP) specifications.

## Overview

`pcdp-lint` validates specification files written in the Post-Coding Development Paradigm format. It enforces structural rules, semantic validation, and cross-section consistency checks according to the PCDP specification schema (v0.3.13).

## Features

- **Comprehensive validation** of PCDP specification files
- **14 validation rules** covering structure, metadata, examples, and behavior definitions
- **Strict mode** for treating warnings as errors
- **Multiple output formats** for integration with CI/CD pipelines
- **Template discovery** with `list-templates` command
- **No external dependencies** — single static binary

## Installation

### From OBS (openSUSE Build Service)

#### openSUSE Leap / SUSE Linux Enterprise

```bash
sudo zypper install pcdp-lint
```

#### Fedora

```bash
sudo dnf install pcdp-lint
```

#### Debian / Ubuntu

```bash
sudo apt-get install pcdp-lint
```

### From Source

```bash
git clone https://github.com/mge1512/pcdp-lint.git
cd pcdp-lint
make build
sudo make install
```

## Usage

### Basic Validation

Validate a specification file:

```bash
pcdp-lint myspec.md
```

### Strict Mode

Treat warnings as errors:

```bash
pcdp-lint strict=true myspec.md
```

### List Available Templates

Display all known deployment templates:

```bash
pcdp-lint list-templates
```

### Version Information

Display version and schema information:

```bash
pcdp-lint version
```

## Command-Line Options

| Option | Description | Default |
|--------|-------------|---------|
| `strict=true` | Treat warnings as errors | `false` |
| `strict=false` | Warnings do not affect exit code | (default) |
| `list-templates` | Print all known deployment templates | N/A |
| `version` | Print version information | N/A |

## Exit Codes

| Code | Meaning |
|------|---------|
| `0` | Specification is valid (no errors; no warnings if strict=true) |
| `1` | Specification is invalid (contains errors, or strict mode with warnings) |
| `2` | Invocation error (bad arguments, missing file, file not readable) |

## Validation Rules

`pcdp-lint` enforces the following validation rules in order:

- **RULE-01**: Required sections present (META, TYPES, BEHAVIOR, PRECONDITIONS, POSTCONDITIONS, INVARIANTS, EXAMPLES)
- **RULE-02**: META fields present and non-empty
- **RULE-02b**: Author field (at least one required)
- **RULE-02c**: Version semantic versioning format
- **RULE-02d**: Spec-Schema semantic versioning format
- **RULE-02e**: License SPDX identifier validation
- **RULE-03**: Deployment template resolution
- **RULE-04**: Deprecated META fields detection
- **RULE-05**: Verification field value validation
- **RULE-06**: EXAMPLES section structure validation
- **RULE-07**: EXAMPLES minimum content validation
- **RULE-08**: BEHAVIOR blocks must contain STEPS
- **RULE-09**: INVARIANTS entries should carry observable/implementation tags
- **RULE-10**: Negative-path EXAMPLE required for BEHAVIOR with error exits
- **RULE-13**: Constraint field value validation
- **RULE-14**: EXECUTION section required in deployment templates

## Output Format

### Diagnostic Line Format

```
SEVERITY  file:line  [section]  message
```

Example:
```
ERROR    account_transfer.md:1    [structure]  Missing required section: ## INVARIANTS
WARNING  account_transfer.md:6    [META]       META field 'Target' is deprecated since v0.3.0
```

### Summary Line Format

```
✓ file: valid                                        (exit 0, no warnings)
✓ file: valid (N warning(s))                         (exit 0, warnings present)
✗ file: N error(s), M warning(s)                     (exit 1, errors present)
✗ file: N error(s), M warning(s) [strict mode]       (exit 1, strict mode)
```

## Examples

### Valid Specification

```bash
$ pcdp-lint valid-spec.md
✓ valid-spec.md: valid
```

### Specification with Warnings

```bash
$ pcdp-lint spec-with-warnings.md
✓ spec-with-warnings.md: valid (1 warning(s))
```

### Invalid Specification

```bash
$ pcdp-lint invalid-spec.md
ERROR    invalid-spec.md:1    [structure]  Missing required section: ## INVARIANTS
✗ invalid-spec.md: 1 error(s), 0 warning(s)
```

### Strict Mode

```bash
$ pcdp-lint strict=true spec-with-warnings.md
WARNING  spec-with-warnings.md:6  [META]  META field 'Target' is deprecated
✗ spec-with-warnings.md: 0 error(s), 1 warning(s) [strict mode]
```

## Specification Format

PCDP specification files are Markdown files (`.md` extension) with the following required sections:

- `## META` — Metadata (Deployment, Version, Author, License, etc.)
- `## TYPES` — Type definitions
- `## BEHAVIOR` — Operation definitions (one or more)
- `## PRECONDITIONS` — Pre-conditions for the operation
- `## POSTCONDITIONS` — Post-conditions for the operation
- `## INVARIANTS` — Invariants that must hold
- `## EXAMPLES` — Examples demonstrating the operation

See https://github.com/mge1512/pcdp-spec for the complete PCDP specification.

## Signal Handling

`pcdp-lint` exits cleanly on SIGTERM and SIGINT (Ctrl-C). No partial output is produced.

## Platform Support

- **Linux** (primary)
- **macOS** (supported)
- **Windows** (not supported in v1)

## License

GNU General Public License v2.0 (GPL-2.0-only)

See LICENSE file for details.

## Contributing

Contributions are welcome. Please ensure all tests pass and follow the existing code style.

```bash
make test
```

## Author

Matthias G. Eckermann <pcdp@mailbox.org>

## References

- [PCDP Specification](https://github.com/mge1512/pcdp-spec)
- [SPDX License List](https://spdx.org/licenses/)
- [Semantic Versioning](https://semver.org/)
