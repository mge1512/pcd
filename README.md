# PCDP — Post-Coding Development Paradigm

**Human Intent, Machine Implementation.**

PCDP is an open specification for a new software development paradigm: domain experts write structured natural-language specifications; AI generates all implementation code. Engineers never write implementation code directly.

This is not "AI-assisted coding" where developers write code with AI suggestions. This is **post-coding development** where specifications are the primary artifact and code is a generated output.

---

## Core Idea

```
Domain expert writes:          AI generates:
┌─────────────────────┐        ┌──────────────────────┐
│  Specification      │  ───▶  │  Source code         │
│  (Markdown)         │        │  Packaging artifacts  │
│                     │        │  Translation report   │
│  TYPES              │        │  Audit bundle         │
│  BEHAVIOR           │        └──────────────────────┘
│  INVARIANTS         │
│  EXAMPLES           │
└─────────────────────┘
```

The target language is never declared in the specification. It is derived automatically from the **deployment template** — a structured definition of the target environment's conventions, constraints, and defaults.

---

## Repository Layout

```
pcdp/
├── README.md                          ← this file
├── LICENSE                            ← CC-BY-4.0 (specs, templates, whitepaper)
├── LICENSE-tools                      ← GPL-2.0-only (tools/)
├── CONTRIBUTING.md
│
├── whitepaper/
│   └── whitepaper.md                  ← canonical whitepaper
│
├── templates/
│   ├── cli-tool.template.md           ← CLI tool deployment template
│   ├── verified-library.template.md   ← safety/security-critical C-ABI libraries
│   ├── library-c-abi.template.md      ← general-purpose C-ABI libraries
│   └── python-tool.template.md        ← Python tooling (QM only)
│
├── tools/
│   └── pcdp-lint/                     ← GPL-2.0-only
│       ├── spec/
│       │   └── pcdp-lint.md           ← specification for pcdp-lint
│       └── code/                      ← generated implementation
│
├── examples/
│   └── account-transfer/
│       └── account-transfer.md        ← worked example from whitepaper
│
└── prompts/
    └── prompt.md                      ← standard translator prompt (A.13)
```

---

## Quick Start

### 1. Validate a specification

```bash
# Install pcdp-lint (openSUSE / SLES)
zypper install pcdp-lint

# Install pcdp-lint (Debian / Ubuntu)
apt install pcdp-lint

# Install pcdp-lint (Fedora)
dnf install pcdp-lint

# Validate a specification file
pcdp-lint myspec.md

# Strict mode (warnings treated as errors)
pcdp-lint strict=true myspec.md

# List available deployment templates
pcdp-lint list-templates
```

### 2. Write a specification

Every specification follows this structure:

```markdown
# My Component

## META
Deployment:  cli-tool
Version:     0.1.0
Spec-Schema: 0.3.7
Author:      Your Name <you@example.org>
License:     Apache-2.0
Verification: none
Safety-Level: QM

## TYPES
...

## BEHAVIOR: my-function
INPUTS: ...
PRECONDITIONS: ...
POSTCONDITIONS: ...

## PRECONDITIONS
...

## POSTCONDITIONS
...

## INVARIANTS
...

## EXAMPLES

EXAMPLE: basic_case
GIVEN:
  ...
WHEN:
  ...
THEN:
  ...
```

### 3. Translate a specification to code

Use the standard translator prompt from `prompts/prompt.md` with any capable LLM. The prompt instructs the LLM to:

- Derive the target language from the deployment template (never declared in the spec)
- Produce all required deliverables defined in the template's DELIVERABLES section
- Write a `TRANSLATION_REPORT.md` documenting decisions and confidence levels

---

## Key Concepts

**Deployment templates** define what a target environment requires — language defaults, binary type, packaging formats, installation method, CLI conventions. The spec author declares `Deployment: cli-tool` and the template resolves all implementation details automatically.

**Verification paths** are optional and pluggable:
- *Direct path:* Specification → Go/C/Rust — fast iteration, lower assurance
- *Verified path:* Specification → Lean 4/F*/Dafny → Go/C — formal proofs, highest assurance

**Audit bundles** are first-class outputs: specification + generated code + proofs (if any) + translation report + metadata. Designed for regulatory compliance with ISO 26262, DO-178C, IEC 62304, and Common Criteria.

---

## Licensing

| Artifact | License |
|---|---|
| Whitepaper, specifications, templates | [CC-BY-4.0](LICENSE) |
| `pcdp-lint` and tools | [GPL-2.0-only](LICENSE-tools) |

The CC-BY-4.0 license on specifications and templates means anyone may implement the paradigm — including proprietary translators and commercial tools — provided attribution is given. The GPL-2.0-only license on `pcdp-lint` ensures the reference validator remains community-controlled and open.

---

## Status

Current version: **0.3.7** (draft)

This project is in active development. The specification format, deployment templates, and tooling are stabilising toward a v1.0 release. Feedback, issue reports, and contributions are welcome.

---

## Author

Matthias G. Eckermann — [pcdp@mailbox.org](mailto:pcdp@mailbox.org)
