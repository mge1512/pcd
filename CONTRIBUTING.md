
# Contributing to PCDP

Thank you for your interest in contributing to the Post-Coding Development Paradigm (PCDP) project!

## Project Overview

PCDP is a paradigm where:
- Domain experts write specifications in structured Markdown
- AI translates specifications into verified implementations  
- Engineers never write implementation code directly
- Target language is derived from deployment templates, not declared by spec authors

**Key distinction from "vibe coding":**
- Vibe coding: humans write code + AI suggests
- PCDP: humans write specs only; AI generates all implementation code
- If generated code is wrong: fix the spec, not the code

## Getting Started

### Prerequisites

- Git
- Go (for building tools)
- Basic understanding of Markdown
- Familiarity with specification writing

### Repository Structure

```
pcdp/
├── doc/                           # Documentation
│   ├── whitepaper.md             # Canonical whitepaper
│   └── executive-brief.md
├── templates/                     # Deployment templates
│   ├── cli-tool.template.md      # Complete, production-ready
│   ├── mcp-server.template.md    # Complete, production-ready
│   └── ...                       # Various template stubs
├── tools/                         # PCDP tooling
│   └── pcdp-lint/                # Specification validator
├── examples/                      # Example specifications
└── prompts/                       # AI translator prompts
```

## How to Contribute

### 1. Specification Writing

When writing PCDP specifications:

#### Required META Section Fields

Every PCDP spec must include:

```markdown
## META

Deployment:        <template-name>
Version:           MAJOR.MINOR.PATCH
Spec-Schema:       MAJOR.MINOR.PATCH
Author:            Name <email>
License:           SPDX-identifier
Verification:      none | lean4 | fstar | dafny | custom
Safety-Level:      QM | ASIL-A | ...
```

For multi-component projects, additionally:
```markdown
Interface-Version: MAJOR.MINOR.PATCH
Imports:           name: path#INTERFACE@>=version
```

#### Behavior Sections

Use these section types:
- `## BEHAVIOR: {n}` — user-facing operation
- `## BEHAVIOR/INTERNAL: {n}` — internal implementation logic

Every BEHAVIOR block **must** include `PRECONDITIONS:`, `STEPS:`, and `POSTCONDITIONS:`.
`STEPS:` are numbered, imperative, and include explicit error exits. They describe the
algorithm — not just the contract. A `MECHANISM:` annotation may follow any step where
the implementation pattern matters for correctness beyond what postconditions capture.

```markdown
## BEHAVIOR: {name}

INPUTS:
  ...

PRECONDITIONS:
  - [condition]

STEPS:
  1. [first action]; on failure → [error action].
  2. [next action].
     MECHANISM: [how, when the how matters for correctness]

POSTCONDITIONS:
  - [outcome]

ERRORS:
  - ERR_X if [condition]
```

#### Examples Section

Must include at least one complete test case. EXAMPLES may use multiple WHEN/THEN pairs
to express multi-pass or multi-step behaviour (e.g. Kubernetes reconcilers). Each
WHEN/THEN pair represents one invocation. Single-pass EXAMPLES remain valid.

```markdown
## EXAMPLES

EXAMPLE: {descriptive_name}
GIVEN:
  [Initial conditions]

WHEN:  [Action — optionally labelled, e.g. "reconcile runs (pass 1)"]
THEN:
  [Expected outcome]

WHEN:  [Action — next pass or step, if multi-pass]
THEN:
  [Expected outcome]
```

#### Interfaces Section (optional, recommended for complex components)

Declare module boundary contracts and required test doubles. This section prevents
translator discretion on abstraction layer decisions and makes independent tests
infrastructure-free.

```markdown
## INTERFACES

AdapterName {
  required-methods:
    MethodName(ctx, InputType) → (OutputType, error)
  implementations-required:
    production:  RealAdapter
    test-double: FakeAdapter {
      configurable fields: ...
      state machine: ...
    }
}
```

#### Dependencies Section (optional)

Declare external library requirements. Translators are bound by these rules and
must not fabricate dependency versions.

```markdown
## DEPENDENCIES

module/path:
  minimum-version: vX.Y.Z
  rationale: [why this version]
  do-not-fabricate: true     # translator must not invent commit hashes
  hints-file: template.language.library.hints.md
```

#### Invariants Section

Tag each invariant with `[observable]` (verifiable by external observation or the
independent test suite) or `[implementation]` (verifiable by code review or static
analysis only). Untagged invariants are a `pcdp-lint` warning.

```markdown
## INVARIANTS

- [observable]      [externally verifiable property]
- [implementation]  [code-review-only property]
```

### 2. Template Development

Deployment templates define how specifications are translated to code.

#### Template Structure
- Use `{curly_braces}` for placeholders (not `<angle_brackets>`)
- Include a DELIVERABLES section listing all files to generate
- Specify constraints and requirements clearly

#### Available Templates (v0.3.14)

| Template | Status | Default Lang | Notes |
|---|---|---|---|
| `cli-tool` | Complete | Go | Production-ready |
| `mcp-server` | Complete | Go | Production-ready |
| `cloud-native` | Complete | Go | TYPE-BINDINGS; kit findings fixed (v0.3.14) |
| `verified-library` | Stub | C | Safety/security-critical C-ABI |
| `library-c-abi` | Stub | C | General-purpose C-ABI |
| `python-tool` | Stub | Python | QM only, no formal verification |
| `project-manifest` | Stub | N/A | Multi-component projects |

#### Hints Files

Library-specific API shapes, version selection rules, and known gotchas
live in `hints/` files, separate from templates and specs.
Naming convention: `<template>.<language>.<library>.hints.md`

Current shipped hints:
- `hints/cloud-native.go.go-libvirt.hints.md`
- `hints/cloud-native.go.golang-crypto-ssh.hints.md`

Specs reference hints files via their `## DEPENDENCIES` section.
Hints are advisory only — they cannot override spec invariants.

### 3. Tool Development

#### pcdp-lint Validation Rules

The `pcdp-lint` tool validates specifications against these rules:

- **RULE-01**: Required sections present (META, TYPES, BEHAVIOR, PRECONDITIONS, POSTCONDITIONS, INVARIANTS, EXAMPLES)
- **RULE-02**: META fields validation
- **RULE-03**: Deployment template resolution
- **RULE-04**: Deprecated META fields detection
- **RULE-05**: Verification field validation
- **RULE-06**: EXAMPLES structure validation (GIVEN/WHEN/THEN; multi-pass permitted)
- **RULE-07**: EXAMPLES content validation
- **RULE-08**: BEHAVIOR blocks must contain STEPS (v0.3.12+)
- **RULE-09**: INVARIANTS entries must carry `[observable]` or `[implementation]` tag (v0.3.12+, warning)

#### CLI Conventions

All PCDP tools follow these conventions:
- Key=value syntax: `pcdp-lint strict=true spec.md`
- Bare words for commands: `pcdp-lint list-templates`
- **NO `--flag` style ever** (firm decision)
- stderr for diagnostics; stdout for summaries and lists
- Exit codes: 0 = valid, 1 = errors/strict warnings, 2 = invocation error

### 4. Code Style and Standards

- **Go**: Follow standard Go conventions
- **Markdown**: Use consistent formatting
- **Filenames**: No version numbers (Git handles versioning)
- **Placeholders**: Always use `{curly_braces}`

## Licensing

The project uses a dual-license model:

| Artifact Type | License |
|---|---|
| Whitepaper, specs, templates, examples | CC-BY-4.0 |
| Tools (pcdp-lint, etc.) | GPL-2.0-only |

**Rationale for GPL-2.0-only on tools:** Mirrors the Linux kernel model. Encourages everybody to contribute changes back to the validator toolchain.

## Submission Guidelines

### Pull Request Process

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/your-feature`
3. Make your changes following the guidelines above
4. Test your changes with `pcdp-lint` if applicable
5. Commit with clear, descriptive messages
6. Push to your fork and submit a pull request

### Commit Message Format

Use clear, descriptive commit messages:
```
component: Brief description of change

Longer explanation of what changed and why, if necessary.
Reference any relevant issues.
```

### Testing

- Run `pcdp-lint` on any specifications you modify
- For Go code: ensure `go build ./...` succeeds
- Test templates with example specifications

## Development Priorities (v0.3.14+)

### v0.3.14 completed
- cloud-native template: INDEPENDENT_TESTS Go naming note, operator.yaml dedup,
  HEALTHCHECK contradiction fixed, CRD scope note, go.sum as generated file
- Compiler gate (Phase 7) added to translator prompt
- Two hints files shipped: cloud-native.go.go-libvirt, cloud-native.go.golang-crypto-ssh
- remote-kvm-operator.md revised to language-neutral v0.3.0

### Carry-forward (templates)
1. Complete `verified-library.template.md`
2. Complete `library-c-abi.template.md`
3. Complete `python-tool.template.md`
4. Complete `project-manifest.template.md` (full BEHAVIOR STEPS beyond stub)
5. Add `independent_tests/` deliverable to cli-tool and mcp-server templates

### Tooling
6. Regenerate `pcdp-lint` implementation from v0.3.13 spec (RULE-08–13 new)
7. Update generic `prompts/prompt.md` (A.13) to include compile gate, TYPE-BINDINGS
   guidance, and v0.3.13 confidence table format

## Communication

- **Email**: pcdp@mailbox.org
- **Issues**: Use GitHub issues for bug reports and feature requests
- **Discussions**: Use GitHub discussions for questions and ideas

## Code of Conduct

We are committed to providing a welcoming and inclusive environment for all contributors. Please be respectful and professional in all interactions.

## Questions?

If you have questions about contributing, please:
1. Check existing issues and discussions
2. Review the whitepaper in `doc/whitepaper.md`
3. Contact us at pcdp@mailbox.org

Thank you for contributing to PCDP!

