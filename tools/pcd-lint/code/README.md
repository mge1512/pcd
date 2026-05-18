# pcd-lint

**pcd-lint** validates Post-Coding Development (PCD) specification files
against the structural rules defined in the pcd-lint specification.

## Installation

### openSUSE / SUSE Linux Enterprise

```
zypper install pcd-tools pcd-templates
```

### Fedora

```
dnf install pcd-tools pcd-templates
```

### Debian / Ubuntu

```
apt install pcd-tools pcd-templates
```

Packages are distributed via [build.opensuse.org](https://build.opensuse.org/).
No curl-based installation is provided.

## Usage

```
pcd-lint <specfile.md>
pcd-lint strict=true <specfile.md>
pcd-lint check-report=true <specfile.md>
pcd-lint strict=true check-report=true <specfile.md>
pcd-lint list-templates
pcd-lint version
```

## Key=value options

| Option | Default | Description |
|--------|---------|-------------|
| `strict=true` | false | Treat warnings as errors; exit 1 on any warning |
| `check-report=true` | false | Evaluate RULE-18: verify Spec-SHA256 in TRANSLATION_REPORT.md |

## Exit codes

| Code | Meaning |
|------|---------|
| 0 | Valid (no errors; no warnings when strict=true) |
| 1 | Invalid (error present, or strict mode and warning present) |
| 2 | Invocation error (bad arguments, file not found, wrong extension) |

## Output

- **stderr**: diagnostic lines in format `SEVERITY  file:line  [section]  message`
- **stdout**: summary line

## Requirements

- Requires `pcd-templates` package (provides template files at `/usr/share/pcd/templates/`)

## License

GPL-2.0-only — see [LICENSE](LICENSE)
