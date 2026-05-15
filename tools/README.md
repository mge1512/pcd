# Tools

This directory holds the reference toolchain for PCD: a linter that
validates specifications, an MCP server that makes the toolchain
available to any MCP-capable LLM host, and the packaging that ships
the deployment templates, hints, and prompts to those tools.

All three are licensed under [GPL-2.0-only](../LICENSE-tools). All three
were specified in PCD and translated by an LLM — zero hand-written
implementation code.

```
tools/
├── pcd-lint/         ← specification validator
├── mcp-server-pcd/   ← MCP server exposing the PCD toolchain
└── pcd-templates/    ← packaging for templates, hints, prompts
```

---

## pcd-lint

The reference validator. Reads a PCD specification file, applies the
rules defined below, and reports structural or semantic defects before
any AI translator is invoked. Every error caught here is an error that
would otherwise produce incorrect or unpredictable generated code.

### Installation

The supply-chain-secure path is the openSUSE Build Service:

```sh
zypper addrepo https://download.opensuse.org/repositories/.../pcd.repo
zypper install pcd-lint
```

The OBS path provides signed RPM packages with a trusted build
history. For development from a git clone, see
[`pcd-lint/spec/pcd-lint.md`](pcd-lint/spec/pcd-lint.md) and the
`Makefile` under `pcd-lint/code/`.

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

`pcd-lint` implements 18 numbered rules. Sub-letters indicate refinements
of an existing rule (e.g. RULE-02b covers the Author field within
RULE-02's META coverage). RULE-18 is conditional: it runs only when
`check-report=true` is passed.

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
| **RULE-18** | 0.3.22 | Warning | TRANSLATION_REPORT.md contains a `Spec-SHA256:` field matching the current spec hash. Runs only with `check-report=true` |

For the normative definition of each rule, including the exact
diagnostic message text, see
[`pcd-lint/spec/pcd-lint.md`](pcd-lint/spec/pcd-lint.md).

---

## mcp-server-pcd

An MCP server that exposes the full PCD toolchain — templates, prompts,
hints, the linter, milestone state, and change-impact analysis — to any
MCP-capable LLM host. The host gets everything in one session, without
local file copies of the supporting material.

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
[`mcp-server-pcd/spec/mcp-server-pcd.md`](mcp-server-pcd/spec/mcp-server-pcd.md).

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

Both `pcd-lint` and `mcp-server-pcd` are specified in PCD and translated
to Go by an LLM. The specifications live alongside the code:

```
tools/pcd-lint/spec/pcd-lint.md                ← the spec
tools/pcd-lint/code/                           ← LLM-generated Go

tools/mcp-server-pcd/spec/mcp-server-pcd.md    ← the spec
tools/mcp-server-pcd/code/                     ← LLM-generated Go
```

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
