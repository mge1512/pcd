# PCD User Guide

**Version:** 0.3.22
**Author:** Matthias G. Eckermann <pcd@mailbox.org>
**Date:** 2026-04-13
**License:** CC-BY-4.0

---

## Introduction — Human intent. Machine implementation.

The leverage of LLMs is in describing intent precisely enough that the model
can take it from there. For personal workflows this is now widespread —
people write a markdown file describing what they want done, drop it in a
directory, and the model handles the rest. The pattern is widely known as
"skills."

Post-Coding Development (PCD) is what this pattern looks like when the
deliverable isn't a personal helper but software shipping to others — software
that has to survive packaging, signing, translation to a different language,
an audit. Same insight that the markdown is the leverage; different
constraints on what it has to carry.

In practical terms, PCD changes one thing about how software is built: humans
write specifications describing what a component should do, and AI translates
those specifications into the complete deliverable — source code, packaging,
documentation, test infrastructure, and the audit record of every translation
decision. Engineers no longer maintain these artifacts by hand. The
specification is the source of truth. If the generated code is wrong, the
specification is fixed and everything downstream regenerates from it — the
code is never edited directly.

This is not AI-assisted coding. In vibe coding, humans write code and AI
suggests improvements. In PCD, humans write specifications and AI generates
the full set of deliverables. The specification is what you version-control,
review, and certify. The code is a build artifact.

**The key rule, stated once:** if the output is wrong, fix the specification
and regenerate. Never edit the generated code.

This guide covers everything you need to use PCD — from choosing your entry
point through writing specifications, translating to code, and managing the
specification lifecycle. It is written for domain experts who know what a
system should do and for engineers who want to understand how to integrate
PCD into their workflow.

For the reasoning behind the design decisions described here, see
`doc/technical-reference.md`. For the paradigm's goals and evidence, see
`doc/whitepaper.md`.

---

## Contents

**Part 1: Getting Started**
1. [Choosing your entry point](#1-choosing-your-entry-point)
2. [Logical entry points](#2-logical-entry-points)
3. [Technical entry points](#3-technical-entry-points)

**Part 2: Writing Specifications**

4. [Writing your first specification](#4-writing-your-first-specification)
5. [Specification structure reference](#5-specification-structure-reference)
6. [Milestones — phased translation for large components](#6-milestones)
7. [Hints files](#7-hints-files)
8. [Language neutrality](#8-language-neutrality)

**Part 3: Translating to Code**

9. [Translating to code](#9-translating-to-code)
10. [Reverse-engineering existing code](#10-reverse-engineering-existing-code)
11. [Prompts reference](#11-prompts-reference)

**Part 4: The Specification Lifecycle**

12. [When the specification changes](#12-when-the-specification-changes)
13. [Verifying artifact provenance](#13-verifying-artifact-provenance)

**Part 5: Reference**

14. [Validating your specification](#14-validating-your-specification)
15. [Quick reference — spec schema](#15-quick-reference--spec-schema)

---

# Part 1: Getting Started

## 1. Choosing Your Entry Point

PCD can be adopted from different starting points and at different levels of
infrastructure investment. Before diving into the details, identify which
situation applies to you.

**Logical entry point** — where you are starting from in terms of existing
work:

| Your situation | Entry point |
|---|---|
| No existing code or spec; starting fresh | Start from scratch via the interview prompt |
| You have a PCD spec that needs extending | Enhance an existing specification |
| You have a codebase and want to migrate parts of it | Add PCD incrementally |
| You want to make PCD the canonical source of truth for an existing project | Rearchitect with PCD |

**Technical entry point** — what infrastructure you want to use:

| Your infrastructure | Entry point |
|---|---|
| No installation; just want to try it | Manual / chatbot |
| Local or CI integration via MCP | Use mcp-server-pcd |
| Automated pipeline from spec to deployment | CI/CD integration |

These two dimensions are independent. A domain expert starting from scratch
(logical entry point 1) might use a chatbot interface today (technical entry
point 1) and migrate to MCP server integration later (technical entry point 2).

---

## 2. Logical Entry Points

### 2.1 Starting from Scratch

You are a domain expert who knows what a component should do but has no
existing code, no spec, and no familiarity with the PCD specification format.
This is the most common starting point.

Use the interview prompt. You do not need to learn the spec format first. The
interview prompt instructs any capable LLM to conduct a structured conversation
with you — asking questions one at a time about the component's purpose,
inputs, outputs, error cases, and deployment context — and produce a complete
specification from your answers.

```bash
# With a local model:
ollama run llama3.2 "$(cat prompts/interview-prompt.md)"

# Or paste prompts/interview-prompt.md as the system prompt
# in any chat interface and start the conversation
```

The interview covers nine phases: component identity, inputs and outputs,
behaviors, error handling, invariants, deployment context, dependencies,
examples, and milestone design for large components. You can stop at any
phase if the component is simple enough.

After the interview, you will have a draft specification. Run `pcd-lint` on
it, review it, and correct any misunderstandings before translating. The
specification is the artifact you are responsible for — the interview produces
a draft, not a final document.

### 2.2 Enhancing an Existing Specification

You already have a PCD specification and want to add behaviors, fix errors,
or extend existing behaviors. This is straightforward because the specification
is the source of truth — you edit it directly.

For small, targeted changes, use the `## DELTA` section to capture the change
as a work order before editing the BEHAVIOR sections:

```markdown
## DELTA

- Add --json output flag to the list subcommand
- Fix error handling in transport layer (currently silently drops errors)
- Add BEHAVIOR: export for bulk data export
```

The `DELTA` section tells the translator what changed without requiring it to
diff the full specification. It is ephemeral — remove it after a successful
translation pass.

For larger changes that affect types, interfaces, or the scaffold, do not use
DELTA. Edit the specification directly, run `pcd-lint`, and assess the impact
of the change before translating (see Part 4 for the change impact workflow).

### 2.3 Adding PCD Incrementally to an Existing Project

You have an existing codebase and want to adopt PCD for new components or for
refactoring high-value existing components, without disrupting the working code.

This is the recommended adoption path for most teams. The key principle is
that PCD components and hand-written components can coexist in the same
project. You do not need to commit to a full migration to start.

**Phase 0: Identify a candidate component.** Choose something small,
well-understood, and relatively self-contained: a crypto primitive, a state
machine, a data validation library, a CLI subcommand. The component should
have clear inputs, outputs, and invariants — if you cannot describe it
precisely, the specification will be weak.

**Phase 1: Write and validate the specification.** Use the interview prompt
or write directly. Run `pcd-lint` until the spec is clean. Keep the existing
implementation running; do not modify it yet.

**Phase 2: Generate and compare.** Translate the specification to the target
language. Run the existing test suite against the generated code. Compare
behavior. The goal at this phase is to validate translation quality, not to
deploy the generated code.

**Phase 3: Shadow deployment.** Deploy the generated code in a non-critical
path or test environment. Monitor behavior. Keep the existing implementation
as a fallback.

**Phase 4: Replace.** Replace the existing implementation with the generated
code. The specification is now the source of truth for this component.
Future changes go through the specification — not the code.

**Phase 5: Expand.** Repeat for additional components. As specifications
accumulate, the value of the approach compounds: new engineers can understand
the system by reading specifications, and cross-component interfaces become
explicit and versioned.

For components where you want the generated code to integrate with an existing
codebase, use the `enhance-existing` deployment template and declare the
existing language explicitly:

```markdown
## META
Deployment:   enhance-existing
Language:     Java
Version:      0.1.0
```

This is the only deployment template where a language declaration in the spec
is required — because the existing codebase already fixes the language choice.

### 2.4 Rearchitecting with PCD as Source of Truth

You want to make PCD the canonical source of truth for an existing project:
the existing code is replaced by generated code, the specification is what
is version-controlled and reviewed, and all future changes go through the
specification.

This is the most committed adoption path. Use the reverse-engineering workflow:

```bash
# Paste prompts/reverse-prompt.md as the system prompt,
# then share your existing source code
```

The reverse prompt reads the existing code and produces a PCD specification
that captures the current behavior. It asks three questions: is the detected
deployment type correct? Should the language stay the same or change? What
do you want to change or add?

After producing the specification, run `pcd-lint`, review it carefully — the
reverse prompt is accurate but not infallible — and translate. The generated
code now replaces the existing implementation. The specification is the new
source of truth.

The `## DELTA` section in the reverse-prompt output captures any changes you
requested. A `## MILESTONE` chain is proposed automatically if the component
is large.

This path is particularly powerful for legacy code: a COBOL program, for
example, can be reverse-engineered to a PCD specification and then translated
to Go, Rust, Java, or any other language supported by the relevant deployment
template. The COBOL is discarded; the specification survives; the target
language is chosen at translation time, not at specification time.

---

## 3. Technical Entry Points

### 3.1 Manual / Chatbot (No Installation Required)

The simplest way to use PCD requires no software installation. Everything you
need is available as Markdown files in the repository.

To write a specification: copy the contents of `prompts/interview-prompt.md`
and use it as the system prompt in any capable LLM chat interface. Answer the
questions and save the produced specification to a `.md` file.

To validate a specification: if you have Go available, install `pcd-lint`
from the repository and run it locally. If not, paste the specification into
a chat interface with the following instruction: "Validate this PCD specification
against the schema described in pcd-lint.md. Report all errors and warnings."
This is a reasonable approximation but is not a substitute for `pcd-lint`.

To translate a specification: copy the contents of `prompts/prompt.md`,
set it as the system prompt, then provide the specification and the appropriate
deployment template (e.g. `templates/cli-tool.template.md`) as input. The
translator will produce the implementation inline.

This path is suitable for evaluation, for initial specification drafts, and
for contexts where software installation is restricted. For production use,
the MCP server path is recommended.

### 3.2 Using mcp-server-pcd

`mcp-server-pcd` is the production-grade path. It is an MCP server that
serves all templates, prompts, and hints to any MCP-capable LLM host. The
translator connects to the server and has everything it needs in one session,
without requiring local file copies of templates or hints.

**Installation:**

```bash
# From OBS (recommended — signed package, supply chain secure):
zypper addrepo https://download.opensuse.org/repositories/...
zypper install mcp-server-pcd

# Run:
mcp-server-pcd stdio   # for mcphost, Claude Desktop, KIT
mcp-server-pcd http    # listens on 127.0.0.1:8080 by default
```

**Configuration for mcphost:**

```yaml
# ~/.config/mcphost/config.yaml
mcpServers:
  pcd:
    command: mcp-server-pcd
    args: [stdio]
```

**Configuration for Claude Desktop:**

```json
{
  "mcpServers": {
    "pcd": {
      "command": "mcp-server-pcd",
      "args": ["stdio"]
    }
  }
}
```

Once connected, the LLM host can access all PCD resources natively:

| Resource URI | Contents |
|---|---|
| `pcd://templates/{name}` | Full deployment template Markdown |
| `pcd://prompts/interview` | Interview prompt |
| `pcd://prompts/translator` | Universal translation prompt |
| `pcd://prompts/reverse` | Reverse-engineering prompt |
| `pcd://hints/{key}` | Library hints files |

**Available tools:**

| Tool | Purpose |
|---|---|
| `list_templates` | List all available deployment templates |
| `get_template` | Retrieve a specific template by name |
| `lint_content` | Validate specification content inline |
| `lint_file` | Validate a specification file on disk |
| `get_schema_version` | Return current spec schema version |
| `set_milestone_status` | Advance milestone pipeline state |
| `assess_change_impact` | Recommend full regen vs. incremental update |
| `verify_spec_hash` | Check if artifacts are current with the spec |
| `list_resources` | List all available MCP resources |

### 3.3 CI/CD Integration

A PCD-integrated CI/CD pipeline automates the full workflow from specification
change to verified artifact. The following describes the logical stages; the
specific implementation depends on your CI system (GitHub Actions, GitLab CI,
OBS, Jenkins, etc.).

**Stage 1: Validate.** On every commit that touches a `.md` file in the
spec directory, run `pcd-lint` against the changed specifications. Fail the
pipeline if any errors are present. This is the gate that prevents invalid
specifications from reaching the translation stage.

**Stage 2: Assess change impact.** For specification changes that pass
validation, assess whether the change warrants full regeneration or incremental
update. Use `assess_change_impact` via `mcp-server-pcd`, or run the
`prompts/change-impact.md` prompt manually. Record the recommendation in the
pipeline log.

**Stage 3: Translate.** Run the AI translator with the specification and the
appropriate deployment template. The translator produces source code, packaging
artifacts, and `TRANSLATION_REPORT.md`. Store the output in the repository
or as pipeline artifacts.

**Stage 4: Verify.** Run the compile gate (`go build ./...`, `cargo build`,
etc.), the generated example tests, and any independent tests from
`independent_tests/`. Run `pcd-lint check-report=true` to verify the spec hash
in `TRANSLATION_REPORT.md` matches the current specification.

**Stage 5: Publish audit bundle.** Package the specification, translation
report, generated code, and test results into the audit bundle. Store it as
a versioned artifact. For regulated deployments, this bundle is the
certification evidence.

**Stage 6: Build and package.** Build the final binary and packaging artifacts
(RPM, DEB, container image) from the generated source. The spec hash embedded
in each artifact provides the cryptographic link to the certified specification.

The key design constraint for CI/CD: the translation stage (Stage 3) involves
an LLM call, which is non-deterministic and potentially slow. Structure the
pipeline so that translation only runs when the specification actually changes
— not on every commit. A hash comparison between the current specification
and the `Spec-SHA256:` field in the existing `TRANSLATION_REPORT.md` is a
lightweight check that avoids unnecessary translation runs.

---

# Part 2: Writing Specifications

## 4. Writing Your First Specification

### Option A — AI-assisted interview (recommended)

You do not need to learn the spec format first. Use the interview prompt with
any capable LLM:

```bash
# With a local model:
ollama run llama3.2 "$(cat prompts/interview-prompt.md)"

# Or paste prompts/interview-prompt.md as the system prompt
# in any chat interface
```

The model asks questions one at a time and produces a complete spec at the
end. It handles both new components (full interview) and existing material
(gap-fill from notes, emails, design docs).

### Option B — Reverse-engineering existing code

If you have existing source code you want to analyse, refactor, or port:

```bash
# Paste prompts/reverse-prompt.md as the system prompt,
# then share your source code
```

The model reads the code, extracts the spec structure, confirms the deployment
type and target language with you, and asks what changes you want to make.

### Option C — Write directly

Every spec follows this skeleton:

```markdown
# My Component

## META
Deployment:   cli-tool
Version:      0.1.0
Spec-Schema:  0.3.21
Author:       Your Name <you@example.org>
License:      Apache-2.0
Verification: none
Safety-Level: QM

## TYPES
...

## BEHAVIOR: my-operation
Constraint: required
INPUTS: ...
PRECONDITIONS: ...
STEPS:
  1. [action]; on failure → [error].
  2. [next action].
POSTCONDITIONS: ...
ERRORS: ...

## PRECONDITIONS
...

## POSTCONDITIONS
...

## INVARIANTS
- [observable]      ...
- [implementation]  ...

## EXAMPLES

EXAMPLE: success_case
GIVEN: ...
WHEN:  ...
THEN:  ...

EXAMPLE: error_case
GIVEN: ...
WHEN:  ...
THEN:  result = Err(...)
```

Validate with `pcd-lint myspec.md` before translating.

---

## 5. Specification Structure Reference

### Required sections

All seven must be present or `pcd-lint` will report errors.

#### `## META`

```markdown
## META
Deployment:   <template>       # cli-tool | mcp-server | cloud-native | ...
Version:      0.1.0            # your spec version (MAJOR.MINOR.PATCH)
Spec-Schema:  0.3.21           # PCD schema version — use current
Author:       Name <email>     # repeatable; multiple Author: lines allowed
License:      Apache-2.0       # SPDX identifier
Verification: none             # none | lean4 | fstar | dafny | custom
Safety-Level: QM               # QM | ASIL-A | ASIL-B | ...
```

The `Deployment:` field determines the target language, packaging format, and
build conventions automatically. You never declare a language in the spec —
that decision belongs to the deployment template and your organisation's
presets.

| Deployment type | What it produces |
|---|---|
| `cli-tool` | Static binary, RPM, DEB, man page |
| `mcp-server` | MCP server binary (stdio + HTTP), RPM, DEB |
| `cloud-native` | Kubernetes operator, CRDs, Helm chart |
| `backend-service` | Linux service, systemd unit, RPM, DEB |
| `gui-tool` | Desktop application |
| `python-tool` | Python package (QM only) |
| `library-c-abi` | C-ABI shared library |
| `verified-library` | Safety/security-critical C library |
| `project-manifest` | Multi-component project definition |
| `enhance-existing` | Integration with existing codebase |

#### `## TYPES`

Declare all data types the component works with. Use pseudocode notation —
no programming language syntax.

```markdown
## TYPES

Account := {
  id:      string where non-empty,
  balance: int    where balance >= 0
}

TransferResult := Ok | Err(ErrorCode)

ErrorCode := InsufficientFunds | InvalidAccount | SameAccount
```

#### `## BEHAVIOR: {name}`

One block per operation. Every block must have INPUTS, PRECONDITIONS, STEPS,
POSTCONDITIONS, and ERRORS.

```markdown
## BEHAVIOR: transfer
Constraint: required

INPUTS:
  from:   Account
  to:     Account
  amount: int where amount > 0

PRECONDITIONS:
- from.id ≠ to.id
- from.balance >= amount

STEPS:
  1. Validate that from.id ≠ to.id; on failure → ERR_SAME_ACCOUNT.
  2. Validate that from.balance >= amount; on failure → ERR_INSUFFICIENT_FUNDS.
  3. Deduct amount from from.balance.
  4. Add amount to to.balance.

POSTCONDITIONS:
- from.balance decreased by amount
- to.balance increased by amount
- sum of all balances is unchanged

ERRORS:
- ERR_SAME_ACCOUNT if from.id = to.id
- ERR_INSUFFICIENT_FUNDS if from.balance < amount
```

**STEPS rules:**
- Numbered, imperative sentences
- Every step that can fail must say: `on failure → [error action]`
- Use `MECHANISM:` annotation when the *how* matters for correctness,
  not just the *what*

**Constraint values:**
- `required` (default): always implement
- `supported`: implement only if active in the resolved preset
- `forbidden`: never implement (add a `reason:` annotation)

Use `## BEHAVIOR/INTERNAL: {name}` for implementation logic not directly
exposed to users. Same structural rules apply.

#### `## PRECONDITIONS`

Global preconditions that apply before the component can run at all.

#### `## POSTCONDITIONS`

Global postconditions guaranteed after any successful operation.

#### `## INVARIANTS`

Rules that must always hold. Tag each entry:

```markdown
## INVARIANTS

- [observable]      sum of all account balances never changes across transfers
- [implementation]  SSH private key bytes are never written to any file path
```

- `[observable]` — verifiable by external observation or the test suite
- `[implementation]` — verifiable only by code review or static analysis

#### `## EXAMPLES`

At least one complete example. Every BEHAVIOR with error exits in STEPS must
have at least one negative-path example (THEN shows an error outcome).

```markdown
## EXAMPLES

EXAMPLE: successful_transfer
GIVEN:
  account A has balance 100
  account B has balance 50
WHEN:
  transfer(A, B, 30)
THEN:
  A.balance = 70
  B.balance = 80
  result = Ok

EXAMPLE: insufficient_funds
GIVEN:
  account A has balance 20
WHEN:
  transfer(A, B, 50)
THEN:
  result = Err(ERR_INSUFFICIENT_FUNDS)
  A.balance unchanged
  B.balance unchanged
```

Multi-pass examples for reconcilers, retry loops, and state machines:

```markdown
EXAMPLE: graceful_stop
GIVEN:
  VM "test-01" is Running
  desiredState = Stopped
WHEN:  reconcile runs (pass 1)
THEN:
  domain.Shutdown() called
  result = RequeueAfter(10s)

WHEN:  reconcile runs (pass 2); domain now Shutoff
THEN:
  status.phase = Stopped
  result = RequeueAfter(60s)
```

### Optional sections

#### `## INTERFACES`

Declare external system boundaries and their test doubles. This prevents the
translator from making ad-hoc abstraction decisions and keeps tests
infrastructure-free.

```markdown
## INTERFACES

Store {
  required-methods:
    Load(id string) → (Record, error)
    Save(r Record) → error
  implementations-required:
    production:  PostgresStore
    test-double: FakeStore {
      configurable fields: records map[string]Record, loadErr error
    }
}
```

#### `## DEPENDENCIES`

Declare external library requirements. The translator must not fabricate
version strings or commit hashes.

```markdown
## DEPENDENCIES

github.com/some/library:
  minimum-version: v1.2.3
  rationale: required for X feature
  do-not-fabricate: true
  hints-file: cli-tool.go.some-library.hints.md
```

#### `## TOOLCHAIN-CONSTRAINTS`

Spec-specific overrides for OCI builds, generated files, or toolchain
constraints that the deployment template does not cover.

#### `## DELIVERABLES`

For multi-component projects: declare logical COMPONENT entries. The
translator maps these to concrete filenames via the deployment template.

#### `## DELTA`

A single-pass work order for the next translation. Lists changes not yet
reflected in the BEHAVIOR sections. Remove it after a successful translation.

```markdown
## DELTA

- Add --json output flag to the list subcommand
- Fix error handling in the transport layer (currently silently drops errors)
```

#### `## MILESTONE`

For large components: defines phased translation. See Section 6.

---

## 6. Milestones

Use milestones when your component is too large to translate in one pass —
roughly when you have more than 10 BEHAVIORs or expect more than 500 lines
of generated code.

### The scaffold-first pattern

The first milestone must always be a scaffold pass. It creates all files,
all types, all function signatures, and all stub bodies for the entire
component. The only acceptance criterion is a clean compile. All subsequent
milestones fill in real implementations without touching the file structure.

This pattern has been validated empirically: a 35-BEHAVIOR, 2900-line
specification was translated to both Go and Rust in single sessions each.
The scaffold held without modification through all seven implementation
milestones in both languages.

### Milestone syntax

```markdown
## MILESTONE: 0.0.0
Status: pending
Scaffold: true
Hints-file: cli-tool.go.milestones.hints.md, mycomponent.implementation.hints.md

Included BEHAVIORs:
  operation-a, operation-b, operation-c, operation-d, operation-e

Acceptance criteria:
  ./mycomponent --version | grep -q "^mycomponent "
  ./mycomponent --help | grep -q "usage:"

## MILESTONE: 0.1.0
Status: pending

Included BEHAVIORs:
  operation-a, operation-b

Deferred BEHAVIORs:
  operation-c, operation-d, operation-e

Acceptance criteria:
  ./mycomponent run | jq '.result | length > 0' | grep -q true

## MILESTONE: 0.2.0
Status: pending

Included BEHAVIORs:
  operation-c, operation-d

Deferred BEHAVIORs:
  operation-e

Acceptance criteria:
  ./mycomponent full | jq '.items | length > 3' | grep -q true
```

### Status values

| Status | Meaning | Set by |
|---|---|---|
| `pending` | Not yet attempted | Author (initial) |
| `active` | Currently being translated | Agent pipeline |
| `failed` | Gates did not pass | Agent pipeline |
| `released` | All gates passed, frozen | Agent pipeline |

Exactly one milestone may be `active` at a time. The `set_milestone_status`
tool in `mcp-server-pcd` advances the cursor. You intervene only on failures.

### Field reference

| Field | Required | Description |
|---|---|---|
| `Status:` | Yes | Pipeline state |
| `Scaffold:` | No | `true` = scaffold pass (default: false) |
| `Hints-file:` | No | Comma-separated hints files to read before translating |
| `Included BEHAVIORs:` | Yes | BEHAVIORs to implement fully in this milestone |
| `Deferred BEHAVIORs:` | No (omit for scaffold) | BEHAVIORs to leave as stubs |
| `Acceptance criteria:` | Recommended | Shell commands; exit 0 = pass |

### Rules

- At most one `Scaffold: true` milestone per spec (`pcd-lint` RULE-17)
- The scaffold milestone must appear first in document order (RULE-17)
- Every BEHAVIOR listed in Included or Deferred must exist in the spec (RULE-16)
- Exactly one milestone may have `Status: active` at any time (RULE-15)

### Acceptance criteria format

Write criteria as shell commands that exit 0 on pass. This makes them
automatable without parsing:

```
./sitar version | grep -q "^sitar "
./sitar all outdir=/tmp/test && test -s /tmp/test/general.json
jq '.cpu._elements | length > 0' /tmp/test/json/cpu.json | grep -q true
```

For components requiring elevated privileges, M0 criteria must be runnable
without privilege. M1+ criteria may require a privileged environment.

---

## 7. Hints Files

Hints files contain implementation knowledge that belongs in neither the spec
(which must be language-agnostic) nor the template (which covers conventions,
not library internals). They are advisory and cannot override spec invariants.

| File pattern | Contents | Lives in |
|---|---|---|
| `<template>.<lang>.milestones.hints.md` | Scaffold patterns, stub conventions, file layout. Reusable across all components using this template and language. | `hints/` or preset hierarchy |
| `<component>.implementation.hints.md` | Component-specific, language-neutral. File groupings, required field names, known failure modes. | `hints/` |
| `<template>.<lang>.<library>.hints.md` | Library API shapes, verified version strings, known gotchas for a specific library. | `hints/` or preset hierarchy |
| `<scope>.<lang>.style.hints.md` | Coding style and architectural conventions for a language within a project or company scope. Authored by the project maintainer. | `/etc/pcd/hints/` (company) or `.pcd/hints/` (project) |
| `<specname>.<lang>.decisions.hints.md` | Implementation decisions from prior translation runs. Generated by the translator. | Next to the spec |

**Style hints files** address integration of PCD-generated components into
existing projects with established coding conventions — OOP patterns, naming
standards, framework idioms. The `<scope>` is a project or company name:

```
# Company-wide Go style — applies to all Go translations on this machine:
/etc/pcd/hints/acme.go.style.hints.md

# Project-specific override — takes precedence over company style:
.pcd/hints/myproject.go.style.hints.md
```

The style hints file is authored and maintained by your team, not generated
by the translator. It is read on every translation run for the matching
language — unlike the decisions hints file, which is only read on guided
regeneration and incremental updates.

Reference hints files from a MILESTONE via the `Hints-file:` field, or from
your spec via the DEPENDENCIES section.

---

## 8. Language Neutrality

A spec that may be translated to more than one language must contain no
language-specific constructs in TYPES, BEHAVIOR, INTERFACES, INVARIANTS,
or MILESTONE acceptance criteria.

**Correct:**
```
STEPS:
  1. Create the output directory recursively if it does not exist.
  2. Write the result as JSON to {outdir}/result.json.
```

**Wrong:**
```
STEPS:
  1. Call os.MkdirAll(outdir, 0755).
  2. json.Marshal the result and write to outdir/result.json.
```

**Acceptance criteria — correct:**
```
test -d /tmp/out && test -s /tmp/out/result.json
jq '.status' /tmp/out/result.json | grep -q '"ok"'
```

**Acceptance criteria — wrong:**
```
go build ./... && ./mytool run
cargo test --release
```

A useful test: can a developer who knows the domain but has not decided on a
target language read and understand this section? If not, language-specific
content has leaked into the spec.

---

# Part 3: Translating to Code

## 9. Translating to Code

Once your spec passes `pcd-lint` with zero errors:

```bash
pcd-lint myspec.md   # must show zero errors before proceeding
```

Provide the spec and the appropriate deployment template to an AI translator
using `prompts/prompt.md` as the system prompt:

```
Input files in the same directory:
  cli-tool.template.md    (the deployment template)
  myspec.md               (your specification)
```

The translator derives the target language from the template, follows the
template's EXECUTION phases, produces all required deliverables, and writes
a `TRANSLATION_REPORT.md` documenting every decision.

**If a milestone is active:** the translator reads the active milestone,
implements only its Included BEHAVIORs, generates stubs for Deferred
BEHAVIORs, and verifies the acceptance criteria.

**If the output is wrong:** fix the specification and regenerate. Do not edit
the generated code.

**Choosing a translator:** any capable LLM works. The paradigm has been
validated on frontier cloud models, 120B open-weight models at EU providers,
and 30B local models (Ollama). Larger context windows produce more complete
deliverable sets. For `mcp-server-pcd` style components with many embedded
assets, full input (all templates, hints, prompts) is required for correct
embedding — partial input produces placeholder assets.

---

## 10. Reverse-Engineering Existing Code

Use `prompts/reverse-prompt.md` to produce a PCD spec from existing source:

1. Paste `reverse-prompt.md` as the system prompt in any chat interface
2. Share your source code (and any design docs, README, or partial specs)
3. The model extracts the spec structure and asks three questions:
   - Is the detected deployment type correct?
   - Should the language stay the same, or change?
   - What do you want to change or add?
4. After gap-fill, it writes the complete spec
5. A `## DELTA` section captures requested changes
6. A `## MILESTONE` chain is proposed if the component is large

The output is a first-class PCD spec. Run `pcd-lint` on it, review it
carefully, then translate normally. The reverse prompt is accurate but not
infallible — the spec is your responsibility, not the model's.

---

## 11. Prompts Reference

| Prompt | Purpose | MCP resource |
|---|---|---|
| `prompts/interview-prompt.md` | New component: guided interview → spec | `pcd://prompts/interview` |
| `prompts/reverse-prompt.md` | Existing code: reverse-engineer → spec | `pcd://prompts/reverse` |
| `prompts/prompt.md` | Translate spec → code (universal translator) | `pcd://prompts/translator` |
| `prompts/change-impact.md` | Assess impact of a spec change | — |

---

# Part 4: The Specification Lifecycle

## 12. When the Specification Changes

Specifications evolve. When a specification changes, two questions must be
answered: is the change valid (does the updated spec pass `pcd-lint`?) and
what does it mean for the generated code (full regeneration or incremental
update)?

### Assessing change impact

Use the `assess_change_impact` tool in `mcp-server-pcd`, or use the
`prompts/change-impact.md` prompt with any LLM. Provide the change
description (a unified diff of the spec or a plain-language description)
and optionally the existing generated code.

The assessment returns a recommendation: **full regeneration** or
**incremental update**, with the primary deciding factor and reasoning.

The decision rules in brief:

| Change affects | Recommendation |
|---|---|
| TYPES, INTERFACES, or INVARIANTS | Full regeneration |
| Scaffold milestone (M0) | Full regeneration |
| A released milestone | Full regeneration |
| STEPS of 1–2 isolated BEHAVIORs | Incremental update |
| EXAMPLES only | Incremental update |
| META only (version bump, license) | Manual edit, no regeneration |

**When in doubt, regenerate.** Writing the spec is the expensive part.
Running the translator is cheap.

### The decisions hints file

When full regeneration is recommended, the assessment produces a list of
implementation decisions from the existing code worth preserving — package
layout, error convention patterns, routing structure. These go into
`<specname>.<language>.decisions.hints.md` alongside the spec before the
regeneration run.

The translator reads this file during guided regeneration and produces an
implementation consistent with the prior decisions. The file is:

- Named with the language qualifier (`myspec.go.decisions.hints.md`) — making
  clear it is language-specific and disposable when switching languages
- Produced by the translator as a deliverable alongside `TRANSLATION_REPORT.md`
- Ignored on clean full regeneration from scratch
- Not a spec artifact — it does not affect `pcd-lint` validation

### The three-state translation model

| Mode | Translator reads | When to use |
|---|---|---|
| Clean full regeneration | Spec + template only | Breaking structural change; language switch; unknown prior quality |
| Guided regeneration | Spec + template + decisions hints file | Non-structural change; good prior architecture worth preserving |
| Incremental update | Spec diff + existing code + decisions hints file | Isolated STEPS/EXAMPLES change; low blast radius confirmed |

---

## 13. Verifying Artifact Provenance

Every artifact generated by a PCD translation run embeds the SHA256 hash of
the specification it was produced from: in source file comments, in binary
version subcommand output, in RPM metadata, in DEB control fields, and in Containerfile
labels.

To check whether a binary is current with the current specification:

```bash
# Get the current spec hash:
sha256sum myspec.md

# Compare with the hash embedded in the binary:
./mybinary --version
# output includes: spec:abc123...

# Or check via mcp-server-pcd:
# tool: verify_spec_hash
# input: spec_path = "path/to/myspec.md"
# output: status = "current" | "stale" | "no-report" | "no-hash-in-report"
```

To check whether the translation report is current:

```bash
pcd-lint check-report=true myspec.md
# Warns if TRANSLATION_REPORT.md Spec-SHA256 does not match current spec hash
```

A `stale` status means the specification has changed since the last translation
run. The artifacts are not current with the current specification. Run the
change impact assessment and translate again.

---

# Part 5: Reference

## 14. Validating Your Specification

```bash
# Basic validation
pcd-lint myspec.md

# Strict mode — warnings become errors
pcd-lint strict=true myspec.md

# Check translation report hash currency
pcd-lint check-report=true myspec.md

# List all known deployment templates
pcd-lint list-templates
```

**Exit codes:**
- `0` — valid (no errors; no warnings in strict mode)
- `1` — invalid (errors present, or warnings in strict mode)
- `2` — invocation error (bad arguments, file not found)

**Diagnostic format:**
```
ERROR   myspec.md:1   [structure]  Missing required section: ## INVARIANTS
WARNING myspec.md:6   [META]       META field 'Target' is deprecated since v0.3.0
```

**Common errors and fixes:**

| Error | Fix |
|---|---|
| Missing required section | Add the missing section |
| BEHAVIOR missing STEPS: | Add a numbered STEPS: block |
| BEHAVIOR has error exits but no negative-path EXAMPLE | Add an EXAMPLE whose THEN: shows an error outcome |
| INVARIANT missing tag | Prefix with `- [observable]` or `- [implementation]` |
| License not valid SPDX | Check https://spdx.org/licenses/ |
| MILESTONE lists unknown BEHAVIOR | The BEHAVIOR name must match exactly what is declared in the spec |
| More than one active MILESTONE | Set all but one to `pending`, `failed`, or `released` |
| Scaffold milestone not first | Move the `Scaffold: true` milestone to appear first in the file |

---

## 15. Quick Reference — Spec Schema

```
Required sections:   META, TYPES, BEHAVIOR, PRECONDITIONS,
                     POSTCONDITIONS, INVARIANTS, EXAMPLES

Optional sections:   INTERFACES, DEPENDENCIES, TOOLCHAIN-CONSTRAINTS,
                     DELIVERABLES, MILESTONE, DELTA

BEHAVIOR variants:   ## BEHAVIOR: {name}
                     ## BEHAVIOR/INTERNAL: {name}

BEHAVIOR fields:     INPUTS, PRECONDITIONS, STEPS, POSTCONDITIONS, ERRORS
                     Constraint: required | supported | forbidden

MILESTONE fields:    Status: pending | active | failed | released
                     Scaffold: true | false  (default: false)
                     Hints-file: {comma-separated filenames}
                     Included BEHAVIORs: {names}
                     Deferred BEHAVIORs: {names}
                     Acceptance criteria: {shell commands}

INVARIANT tags:      [observable] | [implementation]

Hints file layers:   <template>.<lang>.milestones.hints.md
                     <component>.implementation.hints.md
                     <template>.<lang>.<library>.hints.md
                     <scope>.<lang>.style.hints.md         (company or project; preset hierarchy)
                     <specname>.<lang>.decisions.hints.md  (next to spec; generated)

Current Spec-Schema: 0.3.21
```

---

*This document is CC-BY-4.0. Canonical location: `doc/user-guide.md`.*
