# AI Agent Guidelines

This project accepts contributions from AI agents and AI-assisted workflows.
This document defines expectations, conventions, and architecture context.

## Principles

1. **Don't break what works.** Minimise changes. If a test suite exists,
   run it before submitting. If the compile gate fails, fix only the
   identified errors — do not rewrite unaffected files.

2. **Be honest about limitations.** "Applies and builds" is not
   "semantically correct." Do not overstate what a tool verifies.
   In `TRANSLATION_REPORT.md`, use the defined confidence levels honestly:
   High requires a named passing test; Low means reasoning only.
   Unverified claims must be listed explicitly — never silently omitted.

3. **Supply chain security is not optional.** This project targets
   regulated and safety-critical deployment contexts. Concretely:
   - Never use `curl` to download software or dependencies
   - Never use unqualified container image names (e.g. `golang:1.24`)
     Always use fully-qualified names with registry
     (e.g. `registry.suse.com/bci/golang:latest`)
   - Never fabricate dependency version strings or commit hashes
   - Use OBS (build.opensuse.org) for packaging; signed packages only
   - Formal certification frameworks (Common Criteria, ISO 26262, etc.)
     are reference points for the level of rigour expected — not
     necessarily required for every component

4. **No internal references.** No hostnames, IPs, internal paths,
   personal names, or employer-specific content in committed material.
   Use project-neutral framing. Product names are permitted where
   technically precise (e.g. base container image registries).

5. **Credit your work.** Use `Co-Authored-By: <Model> <contact>` in
   commits and documents where AI made a substantive contribution.

6. **Fix the spec, not the code.** This is the core PCD invariant.
   If generated code is wrong, update the specification and regenerate.
   Never hand-edit generated implementation files.

---

## What This Project Is

The **Post-Coding Development (PCD)** is an open specification
for a software development paradigm where:

- Domain experts write specifications in structured Markdown
- AI translates specifications into verified implementations
- Engineers never write implementation code directly
- The target language is derived from deployment templates — never
  declared by the spec author
- AI translation is probabilistic; correctness comes from multiple
  complementary mechanisms (human-reviewable specs, formal verification
  when used, independent tests, audit trails) — not from spec structure alone

**This is not vibe coding.** In vibe coding, if the output is wrong, you
edit the code. In PCD, you fix the specification and regenerate. The spec
is always the source of truth.

**Key artifacts:**
- `pcd-lint` — validates PCD specification files (RULE-01 through RULE-18)
- `mcp-server-pcd` — MCP server serving templates, prompts, and hints;
  exposes `lint_content` and `lint_file` tools for in-session validation
- Deployment templates — define target language, packaging, conventions,
  and the full translation execution recipe per deployment type

**Licenses:**
- Whitepaper, specs, templates, examples: CC-BY-4.0
- Tools (`pcd-lint`, `mcp-server-pcd`): GPL-2.0-only

---

## How This File Relates to PCD

`AGENTS.md` is one of four guidance layers in a PCD repository. They
compose; none replaces the others.

**This file** captures cross-cutting conventions that apply to any
contribution, regardless of which component is being changed: supply
chain rules, honesty expectations, the spec-not-code invariant. Generic
agent guidance lives here.

**Deployment templates** (in [`templates/`](templates/)) capture
conventions specific to one *kind* of component. The `cli-tool` template
fixes the language to Go by default, mandates man pages, and forbids
`curl` as an install method. The `python-tool` template mandates POSIX
`--flag` style. An agent writing a CLI tool reads the `cli-tool`
template; an agent writing a Python script reads `python-tool`. Per-kind
guidance lives there.

**Hints files** (in [`hints/`](hints/)) capture conventions specific to
one library, one language, or one project. The `mcp-server.go.mcp-go`
hints file records the verified `mcp-go` v0.46.0 API shapes. A project's
`style.hints.md` records its naming conventions. Per-library and
per-project guidance lives there.

**The PCD specification** (the file under translation) captures the
actual requirements of one component: its types, behaviours, invariants,
and examples. Per-component requirements live there.

An agent making a change reads all four layers. This file tells the
agent what is universally true about PCD contributions. The template
tells it what is true for components of this kind. The hints files tell
it what is true for the libraries and projects involved. The
specification tells it what is true for this component.

The distinction matters in practice. A rule like "no `curl`" belongs in
this file — it applies to every contribution. A rule like "use
`NewStreamableHTTPServer`, not `NewSSEServer`" belongs in a hints file —
it applies to one library in one language. A rule like "Balance must
never be negative" belongs in the specification — it applies to one
component. Putting a per-library rule in this file makes the file grow
without bound; putting a universal rule in a hints file means it gets
missed.

---

## What You Can Do

- Fix bugs in `pcd-lint` rule implementations
- Add or improve deployment templates
- Add or improve library hints files
- Improve `mcp-server-pcd` tool and resource implementations
- Update documentation — whitepaper, README, slide content
- Add EXAMPLES to existing specs
- Translate a spec to code using the standard translator prompt

## What Requires Human Review

- Changes to pcd-lint RULE definitions — rules affect all downstream
  translation runs and must be reviewed for correctness and consistency
- New deployment templates — the EXECUTION section governs how AI
  translators behave; errors here affect every translation for that type
- Changes to the spec schema (required sections, field names, constraints)
- Any claim about model accuracy, security properties, or certification
  readiness
- Publication decisions affecting external visibility — announcements,
  conference submissions, public positioning statements

---

## Architecture Quick Reference

### Spec format

Required sections: `META`, `TYPES`, `BEHAVIOR`, `PRECONDITIONS`,
`POSTCONDITIONS`, `INVARIANTS`, `EXAMPLES`

Optional sections: `INTERFACES`, `DEPENDENCIES`, `TOOLCHAIN-CONSTRAINTS`,
`DELIVERABLES`, `MILESTONE`, `DELTA`

Every `BEHAVIOR` block requires:
- `STEPS:` — ordered algorithm with explicit error exits on each step
- `Constraint:` — `required` (default) | `supported` | `forbidden`
- Optional `MECHANISM:` annotation where the *how* is normative

Every `INVARIANTS` entry should carry `[observable]` or `[implementation]`.

### Two-layer prompt architecture

```
prompts/prompt.md              — universal, language-agnostic principles
                                 delegates execution recipe to the template

templates/<n>.template.md      — deployment-specific ## EXECUTION section:
  ## EXECUTION                   input files, delivery phases (ordered),
    ### Input files              resume logic, compile gate
    ### Delivery phases
    ### Resume logic
    ### Compile gate
```

`pcd-lint` RULE-14 validates that every deployment template has
a `## EXECUTION` section with the required subsections. RULE-15
through RULE-17 validate MILESTONE structure; RULE-18 validates
spec-hash embedding in `TRANSLATION_REPORT.md`.

### Deployment templates (current)

| Template | Default language | Notes |
|---|---|---|
| `cli-tool` | Go | `--flag`-free CLI; Rust, C, C++, C# alternatives |
| `kubectl-style-cli` | Go | Multi-verb CLI; `--flag` style; Rust alternative |
| `mcp-server` | Go | stdio + streamable-HTTP |
| `backend-service` | Go | 12-factor; systemd unit required |
| `cloud-native` | Go | Kubernetes operators; Go-only |
| `gui-tool` | OS-dependent | Qt6, Tauri, Flutter |
| `cockpit-module` | HTML + JS + CSS | Cockpit web-admin plugins |
| `python-tool` | Python | QM only; POSIX `--flag` style mandatory |
| `library-c-abi` | C | Rust alternative via `cbindgen` |
| `verified-library` | C | Lean4/F\*/Dafny required; QM forbidden |
| `spack-package` | Python (Spack DSL) | Declarative; no compile gate |
| `project-manifest` | N/A | Architect artifact; no code generated |

### mcp-server-pcd

9 tools: `list_templates`, `get_template`, `list_resources`,
`lint_content`, `lint_file`, `get_schema_version`,
`set_milestone_status`, `assess_change_impact`, `verify_spec_hash`

Native MCP resources (browseable without tool calls):
- `pcd://templates/{name}` — full template Markdown
- `pcd://prompts/interview` — interview prompt (embedded at build time)
- `pcd://prompts/translator` — translator prompt (embedded at build time)
- `pcd://prompts/reverse` — reverse-engineering prompt
- `pcd://hints/{template}.{lang}.{lib}` — library hints

Transports — same binary, bare-word selection:
```bash
mcp-server-pcd stdio   # for mcphost, Claude Desktop, VS Code
mcp-server-pcd http    # default: 127.0.0.1:8080
```

### Repository layout

```
prompts/          — translator, interview, reverse, and change-impact prompts
templates/        — deployment templates (*.template.md)
hints/            — library hints files (*.hints.md — not PCD specs)
tools/
  pcd-lint/
    spec/         — canonical pcd-lint specification
    code/         — generated Go implementation
  mcp-server-pcd/
    spec/         — canonical mcp-server-pcd specification
    code/         — generated Go implementation
doc/              — whitepaper, executive brief, presentation slides
examples/         — example PCD specs
```

---

## Conventions

### CLI style (PCD toolchain itself)
- Applies to `pcd-lint` and `mcp-server-pcd` — not to every PCD-generated
  component. Per-template CLI conventions are defined by each deployment
  template; `python-tool` mandates POSIX `--flag` style and forbids
  `key=value`, and `kubectl-style-cli` uses subcommand-flag style.
- For the PCD toolchain: `key=value` for options, bare words for commands
- `stderr` for diagnostics, `stdout` for summaries
- Exit codes: `0` = valid/success, `1` = errors, `2` = invocation error

### Containerfiles
- Builder stage: `FROM registry.suse.com/bci/golang:latest`
- Final stage: `FROM scratch` (static binary, no runtime deps)
- Never use unqualified names (`golang:1.24`, `docker.io/golang`)
- Layer order: `COPY go.mod go.sum` → `RUN go mod download` → `COPY . .`

### Go modules
- Declare direct dependencies only in `go.mod`
- Never hand-write indirect dependencies — use `go mod tidy`
- Never fabricate pseudo-versions or commit hashes for untagged modules
- Use verified versions from hints files when available

### Diagrams
- README.md: Mermaid (GitHub native rendering)
- Whitepaper and audit bundles: Pikchr
- Slides: Pikchr (converted via `pikchr --svg-only | sed ... | magick`)

### Hints files
- Named: `<template>.<language>.<library>.hints.md`
- Live in `hints/` — they are **not** PCD specs
- Running `pcd-lint` against a hints file produces expected errors;
  this is correct behaviour, not a bug
- Advisory only — cannot override spec invariants or template constraints

---

## Tooling Notes

- **pcd-lint:** `pcd-lint myspec.md` / `pcd-lint strict=true myspec.md` /
  `pcd-lint check-report=true myspec.md` / `pcd-lint list-templates`.
  Full rule reference in [`tools/README.md`](tools/README.md).
- **mcp-go API gotchas:** see
  [`hints/mcp-server.go.mcp-go.hints.md`](hints/mcp-server.go.mcp-go.hints.md)
  — version pin, correct constructors, error return convention.
- **Pikchr font fix and diagram conventions:** see the project's
  diagram-tooling notes; Pikchr is a system dependency installed via OBS.
- **Slides:** pandoc → pdflatex. Use `\textrightarrow{}` not `$\rightarrow$`
  in list contexts. UTF-8 em-dashes require `---`. Consider XeLaTeX for
  native UTF-8 support.
- **max_tokens:** ≥ 32000 for complete translation runs
- **Filesystem MCP:** must allow subdirectory creation for packaging artifacts
