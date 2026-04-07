% pcd-lint(1) pcd-lint 0.3.21
% Matthias G. Eckermann <pcd@mailbox.org>
% April 2026

# NAME

pcd-lint — Post-Coding Development specification linter

# SYNOPSIS

**pcd-lint** [*strict=true*] *specfile.md*

**pcd-lint** **list-templates**

**pcd-lint** **version**

# DESCRIPTION

**pcd-lint** validates Post-Coding Development (PCD) specification files
against the structural rules defined in the PCD spec schema (v0.3.21).

It checks required sections, META fields, SPDX license identifiers,
BEHAVIOR blocks, EXAMPLES structure, INVARIANTS tags, MILESTONE sections,
and more. All rules are evaluated — linting does not stop at the first error.

Diagnostics are written to **stderr**. The summary line is written to **stdout**.

# COMMANDS

**list-templates**
:   Print all known deployment templates with their default language annotation.
    Outputs exactly 17 lines. Exits 0.

**version**
:   Print pcd-lint version, spec-schema version, and embedded SPDX list version.
    Format: `pcd-lint {version} (schema {spec-schema}) spdx/{spdx-version}`

# OPTIONS

**strict=true**
:   Treat warnings as errors. Exit code 1 if any warning is present.
    Default: strict=false.

# EXIT CODES

**0**
:   Valid specification. No errors; no warnings (or strict=false and only warnings).

**1**
:   Invalid specification. At least one error; or strict=true and at least one warning.

**2**
:   Invocation error. Bad arguments, file not found, or unreadable file.

# OUTPUT STREAMS

Diagnostics (errors and warnings) are written to **stderr**.

The summary line and list-templates output are written to **stdout**.

# DIAGNOSTIC FORMAT

```
{SEVERITY}  {file}:{line}  [{section}]  {message}
```

Examples:

```
ERROR    spec.md:1    [structure]  Missing required section: ## INVARIANTS
ERROR    spec.md:4    [META]       Missing required META field: Deployment
WARNING  spec.md:6    [META]       META field 'Target' is deprecated since v0.3.0
```

# SUMMARY FORMAT

```
✓ spec.md: valid
✓ spec.md: valid (N warning(s))
✗ spec.md: N error(s), M warning(s)
✗ spec.md: N error(s), M warning(s) [strict mode]
```

# INVOCATION EXAMPLES

Lint a specification file:

```
pcd-lint mycomponent.md
```

Lint with strict mode (warnings treated as errors):

```
pcd-lint strict=true mycomponent.md
```

List all known deployment templates:

```
pcd-lint list-templates
```

Print version information:

```
pcd-lint version
```

# TEMPLATE SEARCH PATH

Template files are searched in the following directories (ascending precedence):

1. `/usr/share/pcd/templates/` — vendor default (pcd-templates package)
2. `/etc/pcd/templates/` — system administrator additions
3. `~/.config/pcd/templates/` — user additions
4. `./.pcd/templates/` — project-local

Directories that do not exist are silently skipped.

# VALIDATION RULES

**RULE-01**: Required sections present (META, TYPES, BEHAVIOR, PRECONDITIONS,
POSTCONDITIONS, INVARIANTS, EXAMPLES).

**RULE-02**: META fields present and non-empty (Deployment, Verification,
Safety-Level, Version, Spec-Schema, License, Author).

**RULE-02c/02d**: Version and Spec-Schema must match MAJOR.MINOR.PATCH.

**RULE-02e**: License must be a valid SPDX identifier or compound expression.

**RULE-03**: Deployment value must be a known deployment template.

**RULE-04**: Deprecated META fields (Target, Domain) emit warnings.

**RULE-05**: Verification field must be one of: none, lean4, fstar, dafny, custom.

**RULE-06/07**: EXAMPLES section must contain complete EXAMPLE/GIVEN/WHEN/THEN blocks.

**RULE-08**: Every BEHAVIOR block must contain a STEPS: block.

**RULE-09**: INVARIANTS entries should carry [observable] or [implementation] tags.

**RULE-10**: BEHAVIOR blocks with error exits must have a negative-path EXAMPLE.

**RULE-13**: Constraint: field on BEHAVIOR headers must be required, supported, or forbidden.

**RULE-14**: Deployment templates must have an EXECUTION section.

**RULE-15/16/17**: MILESTONE sections are validated for structure, status, and BEHAVIOR name existence.

# INSTALLATION

**pcd-lint** is distributed via OBS (build.opensuse.org) as part of the
**pcd-tools** package.

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

No curl-based installation is supported.

# FILES

`/usr/share/pcd/templates/`
:   Vendor-default template files (provided by pcd-templates package).

`/etc/pcd/templates/`
:   System administrator template additions.

`~/.config/pcd/templates/`
:   User template additions.

`./.pcd/templates/`
:   Project-local template overrides.

# SEE ALSO

The PCD specification schema documentation is available at the project
repository. Template files follow the naming convention `{name}.template.md`.

# LICENSE

GPL-2.0-only. See https://spdx.org/licenses/GPL-2.0-only.html
