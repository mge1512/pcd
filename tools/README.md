# Tools

This directory holds the reference toolchain for PCD: one shared
engine and three thin front ends, plus the packaging that ships the
deployment templates, hints, and prompts.

The engine, `libpcd`, is the single translation unit for every
PCD-specific behaviour - spec loading and include resolution,
merged-hash computation, the structural lint rules (RULE-01 through
RULE-25), TYPE-table parsing and JSON Schema emission, milestone
editing, and change-impact assessment. The rule content lives in
includable fragments under `shared/spec/` and is composed into libpcd
via `Includes:`. The front ends do **not** include those fragments;
they bind to the built library through their `DEPENDENCIES` sections,
pinning its version and merged spec hash. Rationale, binding rules,
and provenance obligations: `doc/technical-reference.md`, section 20.

The tools are licensed under [GPL-2.0-only](../LICENSE-tools);
`pcd-templates` packages CC-BY-4.0 content (see below). All tools were
specified in PCD and translated by an LLM — zero hand-written
implementation code.

```
tools/
├── libpcd/           ← shared PCD engine (library, no executable)
├── pcd-lint/         ← specification validator (CLI front end)
├── pcd-emit-types/   ← TYPE-table schema emission (CLI front end)
├── mcp-server-pcd/   ← MCP server exposing the PCD toolchain
├── shared/spec/      ← includable rule fragments (composed into libpcd)
└── pcd-templates/    ← packaging for templates, hints, prompts
```

---

## libpcd

The shared PCD engine. A language-parameterised library
(`Deployment: library`, no executable entry point) whose merged
specification composes the three rule fragments:

| Fragment | Content |
|---|---|
| [`shared/spec/lint-rules.md`](shared/spec/lint-rules.md) | RULE-01 through RULE-21, rule-domain types, fence tracking |
| [`shared/spec/types-table-rules.md`](shared/spec/types-table-rules.md) | tabular TYPES grammar, cell vocabularies, RULE-22 through RULE-25 |
| [`shared/spec/types-emit-rules.md`](shared/spec/types-emit-rules.md) | TYPE-table to JSON Schema mapping, canonical output form |

Public interface: `SpecEngine` (see
[`libpcd/spec/libpcd.spec.md`](libpcd/spec/libpcd.spec.md)). Consumers
pin the library version in their spec's `DEPENDENCIES` and record its
merged spec hash in their `TRANSLATION_REPORT.md`; the build service
enforces version alignment across all consumers. Lint parity between
`pcd-lint` and `mcp-server-pcd` holds by construction: both call the
same engine.

libpcd is a build-time dependency; end users do not install it
directly.

---

## pcd-lint

The reference validator - a thin CLI front end over libpcd. Reads a
PCD specification file, applies the rules defined below via the shared
engine, and reports structural or semantic defects before any AI
translator is invoked. Every error caught here is an error that would
otherwise produce incorrect or unpredictable generated code.

### Installation

The supply-chain-secure path is the openSUSE Build Service:

```sh
zypper addrepo https://download.opensuse.org/repositories/.../pcd.repo
zypper install pcd-lint
```

The OBS path provides signed RPM packages with a trusted build
history. For development from a git clone, see
[`pcd-lint/spec/pcd-lint.spec.md`](pcd-lint/spec/pcd-lint.spec.md) and
the `Makefile` under `pcd-lint/code/`.

### Usage

```sh
pcd-lint <specfile>.md                # validate
pcd-lint strict=true <specfile>.md    # warnings become errors
pcd-lint check-report=true <specfile>.md
                                      # also verify TRANSLATION_REPORT.md
pcd-lint list-templates               # list known deployment templates
pcd-lint version                      # version and embedded SPDX list version
```

### Exit codes

| Code | Meaning |
|---|---|
| 0 | Validation passed. Warnings may be present unless `strict=true` |
| 1 | Validation failed. At least one Error, or one Warning under `strict=true` |
| 2 | Invocation error: file not found, missing argument, unrecognised key |

`pcd-lint` is idempotent, makes no network calls, reads no environment
variables for behavioural control, and never modifies the file it is
validating. Diagnostics go to stderr; summary and `list-templates`
output go to stdout.

### Rule reference

The engine implements 25 numbered rules; `pcd-lint` surfaces all of
them. Sub-letters indicate refinements of an existing rule (e.g.
RULE-02b covers the Author field within RULE-02's META coverage).
RULE-18 is conditional: it runs only when `check-report=true` is
passed. RULE-19 through RULE-21 run when the spec declares `Includes:`;
RULE-22 through RULE-25 run when the merged TYPES section contains
TYPE tables.

| Rule | Since | Severity | Validates |
|---|---|---|---|
| **RULE-01** | 0.3.0 | Error | All required sections present: META, TYPES, BEHAVIOR, PRECONDITIONS, POSTCONDITIONS, INVARIANTS, EXAMPLES |
| **RULE-02** | 0.3.0 | Error | Required META fields present and non-empty: Deployment, Verification, Safety-Level, Version, Spec-Schema, License |
| RULE-02b | 0.3.0 | Error | At least one Author field; multiple Author lines permitted |
| RULE-02c | 0.3.0 | Error | Version follows semantic versioning `MAJOR.MINOR.PATCH` |
| RULE-02d | 0.3.0 | Error | Spec-Schema follows semantic versioning |
| RULE-02e | 0.3.0 | Error | License is a valid SPDX identifier or compound expression |
| **RULE-03** | 0.3.0 | Error | Deployment template name resolves to a known template; per-template constraints satisfied (e.g. `python-tool` requires Safety-Level: QM and Verification: none) |
| **RULE-04** | 0.3.0 | Warning | Deprecated META fields (`Target`, `Domain`) trigger migration advice |
| **RULE-05** | 0.3.0 | Warning | Verification value is one of: none, lean4, fstar, dafny, custom |
| **RULE-06** | 0.3.0 | Error | EXAMPLES section structure: at least one block, each with EXAMPLE/GIVEN/WHEN/THEN markers; multi-pass WHEN/THEN pairs supported |
| **RULE-07** | 0.3.0 | Warning | EXAMPLES content: GIVEN, WHEN, and THEN blocks non-empty |
| **RULE-08** | 0.3.12 | Error | Every BEHAVIOR contains a STEPS: block |
| **RULE-09** | 0.3.12 | Warning | INVARIANTS entries tagged `[observable]` or `[implementation]` |
| **RULE-10** | 0.3.13 | Error | Every BEHAVIOR with error exits in STEPS has at least one negative-path EXAMPLE |
| **RULE-11** | 0.3.13 | Warning | TOOLCHAIN-CONSTRAINTS section uses valid constraint values (required, forbidden) |
| **RULE-12** | 0.3.13 | Mixed | Cross-section consistency |
| RULE-12a | 0.3.13 | Warning | INTERFACES identifiers referenced verbatim in BEHAVIOR STEPS |
| RULE-12b | 0.3.13 | Error | TYPES are not redefined in BEHAVIOR sections |
| RULE-12c | 0.3.13 | Warning | Files referenced in BEHAVIOR/INTERNAL are declared in DELIVERABLES |
| **RULE-13** | 0.3.13 | Error | BEHAVIOR `Constraint:` value is one of: required, supported, forbidden; `forbidden` requires a `reason:` annotation |
| **RULE-14** | 0.3.16 | Warning | Deployment templates declare an `## EXECUTION` section with Delivery phases, Compile gate (or `COMPILE-GATE: none`), and Resume logic — unless META declares `EXECUTION: none` |
| **RULE-15** | 0.3.21 | Mixed | MILESTONE structure: required fields, valid Status values, at most one active milestone |
| **RULE-16** | 0.3.21 | Error | BEHAVIOR names in MILESTONE Included/Deferred lists exist in the spec |
| **RULE-17** | 0.3.21 | Error | At most one milestone has `Scaffold: true`, and the scaffold milestone appears first |
| **RULE-18** | 0.3.22 | Warning | TRANSLATION_REPORT.md contains a `Spec-SHA256:` field matching the current merged spec hash. Runs only with `check-report=true` |
| **RULE-19** | 0.4.0 | Error | Every `Includes:` path resolves to a readable file |
| **RULE-20** | 0.4.0 | Error | Merged spec has no name collisions across TYPES, BEHAVIORs, INTERFACES, EXAMPLES |
| **RULE-21** | 0.4.0 | Error | Inclusion graph is acyclic; included specs declare no MILESTONE or DEPLOYMENT sections |
| **RULE-22** | 0.4.5 | Error | TYPE block and table structure: heading, Kind, column contract per kind, divider, row widths |
| **RULE-23** | 0.4.5 | Error | TYPE table cell vocabulary: names, TypeRefs, Req values, constraint tokens, defaults |
| **RULE-24** | 0.4.5 | Error | TYPE table references resolve; names unique across tables and against prose TYPES |
| **RULE-25** | 0.4.5 | Warning | TYPE table lifecycle discipline: Since/Deprecated/Replaced-by values, deprecation points to a successor |

For the normative definition of each rule, including the exact
diagnostic message text, see
[`shared/spec/lint-rules.md`](shared/spec/lint-rules.md) (RULE-01
through RULE-21) and
[`shared/spec/types-table-rules.md`](shared/spec/types-table-rules.md)
(RULE-22 through RULE-25), composed into
[`libpcd/spec/libpcd.spec.md`](libpcd/spec/libpcd.spec.md).

---

## pcd-emit-types

A thin CLI front end over libpcd that lowers the TYPE tables of a
specification to a canonical JSON Schema 2020-12 document - the
deterministic emission path for the declarative data-contract layer.
Same table, same bytes: output is byte-identical across runs and
embeds the merged spec hash as `x-pcd-spec-sha256`.

### Installation

```sh
zypper addrepo https://download.opensuse.org/repositories/.../pcd.repo
zypper install pcd-emit-types
```

### Usage

```sh
pcd-emit-types <specfile>.md                  # schema to stdout
pcd-emit-types out=types.schema.json <specfile>.md
pcd-emit-types strict=true <specfile>.md      # RULE-25 warnings become errors
pcd-emit-types version
```

### Exit codes

| Code | Meaning |
|---|---|
| 0 | Schema emitted. Warnings may be present unless `strict=true` |
| 1 | Refused: TYPE-table or include Errors, no TYPE tables, or a Warning under `strict=true` |
| 2 | Invocation error: file not found, missing argument, unrecognised key, unwritable destination |

In stdout mode the schema is the only content on stdout; all
diagnostics go to stderr. The mapping and the canonical output form
are normative in
[`shared/spec/types-emit-rules.md`](shared/spec/types-emit-rules.md);
the tool spec is
[`pcd-emit-types/spec/pcd-emit-types.spec.md`](pcd-emit-types/spec/pcd-emit-types.spec.md).

---

## mcp-server-pcd

An MCP server that exposes the full PCD toolchain — templates, prompts,
hints, the linter, milestone state, change-impact analysis, and schema
emission — to any MCP-capable LLM host. A thin front end over libpcd
(same engine as the CLIs), it gives the host everything in one session
without local file copies of the supporting material.

### Installation

```sh
zypper addrepo https://download.opensuse.org/repositories/.../pcd.repo
zypper install mcp-server-pcd
```

### Transports

```sh
mcp-server-pcd stdio                  # for mcphost, Claude Desktop, KIT
mcp-server-pcd http                   # listens on 127.0.0.1:8080
mcp-server-pcd http listen=:9090      # custom address
```

Both transports serve identical content. The stdio transport is the
production default for desktop LLM hosts; the streamable-HTTP transport
is for shared deployments and CI integration.

### Tools

| Tool | Purpose |
|---|---|
| `list_templates` | Enumerate available deployment templates |
| `get_template` | Retrieve a template by name |
| `lint_content` | Validate specification content passed inline |
| `lint_file` | Validate a specification file by path |
| `get_schema_version` | Return the current spec schema version |
| `set_milestone_status` | Advance milestone pipeline state (pending → active → released, or failed) |
| `assess_change_impact` | Recommend full regeneration or incremental update for a spec change |
| `verify_spec_hash` | Check whether a generated artifact is current with respect to the spec |
| `emit_types` | Emit the canonical JSON Schema for a spec's TYPE tables |
| `list_resources` | Enumerate available MCP resources |

### Resources

| URI pattern | Contents |
|---|---|
| `pcd://templates/{name}` | Full deployment template Markdown |
| `pcd://prompts/interview` | Interview prompt (spec authoring via dialogue) |
| `pcd://prompts/translator` | Universal translator prompt |
| `pcd://prompts/reverse` | Reverse-engineering prompt |
| `pcd://hints/{key}` | Library and milestone hints files |

### Configuration

For `mcphost`:

```yaml
mcpServers:
  pcd:
    command: mcp-server-pcd
    args: [stdio]
```

For Claude Desktop:

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

The server never modifies a file on disk except via the
`set_milestone_status` tool (which edits exactly the `Status:` line of
one named milestone), never makes outbound network calls, and never
reads environment variables for behavioural control. All MCP responses
are valid JSON-RPC 2.0.

For the normative definition, see
[`mcp-server-pcd/spec/mcp-server-pcd.spec.md`](mcp-server-pcd/spec/mcp-server-pcd.spec.md).

---

## pcd-templates

Packaging metadata that ships the deployment templates, hints files,
and prompts to the install locations expected by `pcd-lint` and
`mcp-server-pcd`. There is no executable code here — just a `Makefile`,
an RPM spec, and Debian packaging that copy the content from the
[`templates/`](../templates/), [`hints/`](../hints/), and
[`prompts/`](../prompts/) directories at the repository root.

Default install layout (Linux):

```
/usr/share/pcd/templates/       deployment templates
/usr/share/pcd/hints/           shipped hints files
/usr/share/pcd/prompts/         interview, translator, reverse prompts
```

The preset hierarchy permits organisations and projects to override any
of these without touching the vendor-shipped files. See
[`templates/README.md`](../templates/README.md) and
[`hints/README.md`](../hints/README.md) for the full layering model.

`pcd-templates` is published under [CC-BY-4.0](../LICENSE), matching the
license of the content it packages, not the GPL-2.0-only that covers
the tools themselves.

---

## Self-Hosting

The engine and every front end are specified in PCD and translated by
an LLM; the target language is resolved from each component's
deployment template, not hard-coded here. The specifications live
alongside the code:

```
tools/libpcd/spec/libpcd.spec.md                   ← shared engine spec
tools/libpcd/code/                                 ← LLM-generated

tools/pcd-lint/spec/pcd-lint.spec.md               ← the spec
tools/pcd-lint/code/                               ← LLM-generated

tools/pcd-emit-types/spec/pcd-emit-types.spec.md   ← the spec
tools/pcd-emit-types/code/                         ← LLM-generated

tools/mcp-server-pcd/spec/mcp-server-pcd.spec.md   ← the spec
tools/mcp-server-pcd/code/                         ← LLM-generated
```

The rule content the engine composes lives in `shared/spec/` and is
not a component of its own - it has no `code/` directory because it is
never translated alone; it is included into libpcd and translated as
part of it.

When a tool is regenerated, the workflow is preserved as three
commits: the pre-regeneration baseline, the LLM-generated translation
(with the model named in the commit message), and any post-translation
corrections. This makes the translation step auditable in the git
history without losing the ability to compare runs.

The generated code under `tools/*/code/` is marked `linguist-generated`
in `.gitattributes`, so GitHub's language statistics reflect the
specifications — Markdown — rather than the implementation language of
any one translation run. Diffs in `tools/*/code/` are collapsed by
default in the GitHub web UI; review changes by reading the spec diff
and the `TRANSLATION_REPORT.md`, not the generated code.
