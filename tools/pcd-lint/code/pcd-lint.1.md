# pcd-lint.1.md — man page source

% PCD-LINT(1) pcd-lint 0.4.0
% Matthias G. Eckermann
% May 2026

# NAME

pcd-lint — validate Post-Coding Development specification files

# SYNOPSIS

**pcd-lint** [*strict=true*] [*check-report=true*] *specfile.md*

**pcd-lint** *list-templates*

**pcd-lint** *version*

# DESCRIPTION

**pcd-lint** validates a specification file against the structural rules
defined in the pcd-lint specification (Spec-Schema 0.4.0).

It applies RULE-01 through RULE-21, collects all diagnostics (no
short-circuiting), and reports a summary line.

# OPTIONS

**strict=true**
: Treat warnings as errors; exit 1 on any warning. Default: false.

**check-report=true**
: Evaluate RULE-18: look for TRANSLATION_REPORT.md adjacent to the spec
  and verify the Spec-SHA256 field. Emits warnings only. Default: false.

# COMMANDS

**list-templates**
: Print all 17 known deployment templates and their default language
  annotations, then exit 0.

**version**
: Print pcd-lint version, schema version, and embedded SPDX list version.

# EXIT STATUS

**0**
: Valid (no errors; no warnings when strict=true).

**1**
: Invalid (at least one Error, or strict=true and at least one Warning).

**2**
: Invocation error (bad arguments, file not found, wrong extension).

# DIAGNOSTICS

Diagnostic lines are written to **stderr** in the format:

    SEVERITY  file:line  [section]  message

The summary line is written to **stdout**.

# EXAMPLES

    pcd-lint myspec.md
    pcd-lint strict=true myspec.md
    pcd-lint check-report=true myspec.md
    pcd-lint list-templates

# INSTALLATION

Available as the **pcd-tools** OBS package for openSUSE Leap, SUSE Linux
Enterprise, Fedora, and Debian/Ubuntu. Requires the **pcd-templates**
package.

    https://build.opensuse.org/

# SEE ALSO

The pcd-lint specification: **pcd-lint.md**

# LICENSE

GPL-2.0-only. See <https://spdx.org/licenses/GPL-2.0-only.html>
