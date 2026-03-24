# pcdp-lint

A command-line tool for validating Post-Coding Development Paradigm (PCDP) specification files.

## Installation

### Via Package Manager (Recommended)

#### openSUSE/SUSE Linux Enterprise
```bash
# Add the repository (if not already added)
zypper addrepo https://download.opensuse.org/repositories/devel:/tools/openSUSE_Leap_15.4/ pcdp-tools
zypper refresh
zypper install pcdp-tools
```

#### Fedora
```bash
dnf install pcdp-tools
```

#### Debian/Ubuntu
```bash
apt update
apt install pcdp-tools
```

### From Source
```bash
make build
sudo make install
```

## Usage

### Validate a specification file
```bash
pcdp-lint myspec.md
```

### Strict mode (treat warnings as errors)
```bash
pcdp-lint strict=true myspec.md
```

### List available deployment templates
```bash
pcdp-lint list-templates
```

### Get version information
```bash
pcdp-lint version
```

## Command-Line Arguments

- `strict=true` - Treat warnings as errors (default: false)
- `list-templates` - Show all known deployment templates with their default languages
- `version` - Show version information and exit

## Exit Codes

- **0**: Valid specification (no errors; no warnings when strict=true)
- **1**: Invalid specification (errors found; or warnings found when strict=true)
- **2**: Invocation error (bad arguments, file not found, unreadable file)

## Output Format

### Diagnostics (stderr)
```
ERROR    spec.md:42  [EXAMPLES]  Example 'foo' missing THEN: marker
WARNING  spec.md:6   [META]      META field 'Target' is deprecated since v0.3.0
```

### Summary (stdout)
```
✓ spec.md: valid
✓ spec.md: valid (1 warning(s))
✗ spec.md: 2 error(s), 1 warning(s)
✗ spec.md: 0 error(s), 1 warning(s) [strict mode]
```

## Validation Rules

pcdp-lint validates PCDP specification files against 13 structural rules:

1. **Required sections present** - Checks for META, TYPES, BEHAVIOR, PRECONDITIONS, POSTCONDITIONS, INVARIANTS, EXAMPLES
2. **META fields** - Validates required fields: Deployment, Verification, Safety-Level, Version, Spec-Schema, License, Author
3. **Deployment template resolution** - Validates deployment template values and special requirements
4. **Deprecated fields** - Warns about deprecated META fields like Target and Domain
5. **Verification field** - Validates verification backend values
6. **EXAMPLES structure** - Ensures proper EXAMPLE:/GIVEN:/WHEN:/THEN: structure
7. **EXAMPLES content** - Warns about empty example blocks
8. **BEHAVIOR steps** - Requires STEPS: blocks in all BEHAVIOR sections
9. **INVARIANTS tags** - Recommends [observable]/[implementation] tags on invariant entries
10. **Negative-path examples** - Requires error-case examples for behaviors with error exits
11. **TOOLCHAIN-CONSTRAINTS** - Validates constraint values in toolchain sections
12. **Cross-section consistency** - Checks identifier and type name consistency across sections
13. **BEHAVIOR constraints** - Validates Constraint: field values on BEHAVIOR headers

## Template Support

pcdp-lint recognizes the following deployment templates:

- `cli-tool` → Go
- `backend-service` → Go
- `verified-library` → C
- `library-c-abi` → C
- `python-tool` → Python
- `enhance-existing` → (declare Language: in META)
- `manual` → (declare Target: in META)
- And others (run `pcdp-lint list-templates` for the complete list)

## License

GPL-2.0-only

## Contributing

This tool is part of the Post-Coding Development Paradigm project. For bug reports and feature requests, please refer to the project documentation.

## Technical Notes

- pcdp-lint is a static binary with no runtime dependencies
- It does not make network calls or modify input files
- It follows the cli-tool deployment template constraints
- Signal handling: Clean exit on SIGTERM and SIGINT