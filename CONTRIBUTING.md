# Contributing to PCD

This document is for people who want to **work on PCD itself** — improving
the specification format, adding deployment templates, fixing `pcd-lint`,
writing examples, or extending the tooling.

If you want to **use PCD** to specify and generate your own software, see
[`doc/user-guide.md`](doc/user-guide.md).

For the reasoning behind design decisions, see
[`doc/technical-reference.md`](doc/technical-reference.md).

For agent-specific conventions (AI-assisted contributions, supply-chain
rules), see [`AGENTS.md`](AGENTS.md).

---

## Getting Started

Prerequisites: Git, Go (for building tools), basic Markdown familiarity.
Read `doc/user-guide.md` before contributing.

---

## Before You Open a Pull Request

Some changes can land directly via PR; others should be discussed first
in an issue. The bar is whether the change affects translation behaviour
for components other than your own.

**Direct PR is fine for:** typo and prose fixes; new EXAMPLES in existing
specs; new hints files for libraries you have actually used; additional
test cases for `pcd-lint` rules; documentation improvements; fixing a
spec defect you uncovered while using PCD.

**Open an issue first for:** new deployment templates; changes to
`pcd-lint` RULE definitions; changes to the spec schema (required
sections, field names, constraints); changes to the prompts in
`prompts/`; anything that would require regenerating either tool.

The reason for the second list is that these changes affect every
downstream user of PCD. A new RULE in `pcd-lint` might invalidate every
spec written before it. A change to `prompts/prompt.md` changes every
translation run. These need discussion before code.

---

## What to Work On

### Deployment templates (`templates/`)

Templates define how specs are translated for a given deployment context.
Every template must include a `## DELIVERABLES` table, an `## EXECUTION`
section (or `EXECUTION: none` in META), and pass `pcd-lint` with zero
errors. Use `{curly_braces}` for placeholders.

The current set of twelve templates is listed in
[`templates/README.md`](templates/README.md) and
[`doc/technical-reference.md`](doc/technical-reference.md) section 4.

**Before proposing a new template**, check that the need cannot be met by
configuring an existing one through a preset. The `cli-tool` template
covers Go, Rust, C, C++, and C# CLIs by changing the preset's default
language — adding a `rust-cli-tool` would be wrong. New templates are
warranted when the *deliverables* differ (Cockpit modules ship HTML+JS
into `/usr/share/cockpit/`, not a binary; `spack-package` produces a
declarative Python class with no compile gate), not when only the
language differs.

The workflow for a new template:

1. Open an issue describing the deployment context and why no existing
   template fits.
2. Use the closest existing template as your starting point — `cli-tool`
   for any compiled binary; `cockpit-module` for any template that ships
   non-compiled assets; `python-tool` for any declarative or
   interpreted target.
3. Set `Template-For:` in META to match your filename stem (so
   `kubectl-style-cli.template.md` has `Template-For: kubectl-style-cli`).
4. Demonstrate the template with at least one real component
   translation. A template that has never produced working code is not
   ready for inclusion.

### Hints files (`hints/`)

Hints files contain library-specific implementation knowledge. Five-layer
naming convention — see [`hints/README.md`](hints/README.md) for the
full layering model and [`doc/technical-reference.md`](doc/technical-reference.md)
section 5 for the lifecycle of each layer.

New hints files are usually written in response to a translator failure
you have just diagnosed. The most valuable hints files are concise: one
fact per section, with the right and wrong patterns shown as code
snippets rather than explained as prose.

### pcd-lint (`tools/pcd-lint/`)

The reference validator. The spec lives in
[`tools/pcd-lint/spec/pcd-lint.md`](tools/pcd-lint/spec/pcd-lint.md);
the full rule reference is in [`tools/README.md`](tools/README.md).

To add or change a validation rule:

1. Add or modify the rule definition under the `### RULE-N:` pattern in
   the spec.
2. Add EXAMPLES covering both the positive and negative paths. RULE-10
   itself requires this for any BEHAVIOR with error exits.
3. Update the STEPS list in `BEHAVIOR: lint-validation-rules`.
4. Regenerate the tool. Do not hand-edit the generated implementation.

A good lint rule has a clear normative justification (a translator
without this rule would produce wrong output), a precise structural
trigger (no fuzzy "looks like" matching), and an actionable diagnostic
message (the user knows what to fix). Rules that warn about *style* or
*convention* without affecting correctness belong in hints files, not in
the linter.

### mcp-server-pcd (`tools/mcp-server-pcd/`)

Spec lives in
[`tools/mcp-server-pcd/spec/mcp-server-pcd.md`](tools/mcp-server-pcd/spec/mcp-server-pcd.md).
Same rule: fix the spec, regenerate. Asset embedding (templates, hints,
prompts) happens at build time via the Makefile's `embed-assets` target;
the embedded inventory must stay in sync with the repository root
directories.

### Examples (`examples/`)

Every example must pass `pcd-lint` with zero errors, use the current
`Spec-Schema:` version in META, and include at least one non-trivial
BEHAVIOR with error exits and a corresponding negative-path EXAMPLE.

A new example is worth adding when it demonstrates something the
existing examples do not — a new deployment template, a verification
path, a domain pattern. Adding another simple CLI tool when
`calc-interest` already exists is duplication.

### Prompts (`prompts/`)

Changes here affect every user. Test with at least one real translation
run before submitting, using a representative spec from `examples/` or
`tools/`. A "real translation run" means: pass the modified prompt and a
spec to an LLM, observe the generated output, confirm it compiles and
passes its own EXAMPLES. Report the model and version in your PR
description.

---

## Reporting Translator Failures

When a translation run produces wrong, incomplete, or fabricated output,
open an issue tagged `translator-failure`. Include all of the following.
This is exactly the audit trail PCD is meant to produce, applied
recursively to PCD itself.

**The inputs:**

- The specification file (or a link to it in the repo if it lives there).
- The deployment template version used.
- The prompt file used (`prompts/prompt.md`, `prompts/reverse-prompt.md`,
  etc.) and its version.
- Any hints files in scope.
- The model name and version (e.g. "Claude Sonnet 4.5", "Gemini 2.5 Pro
  via Vertex", "Qwen3-Coder 30B via Ollama"). If a self-hosted model,
  the quantisation and context window.
- The `max_tokens` setting if applicable.

**The output:**

- What the translator produced — the generated files, or the inline
  output if no filesystem was available.
- The `TRANSLATION_REPORT.md` if one was produced.
- Any error messages from the compile gate.

**The diagnosis:**

- What was expected — which BEHAVIOR or DELIVERABLE was wrong, what
  the spec says, what the translator did instead.
- Your hypothesis about the cause, if you have one. Common patterns:
  fabricated library version, missing hints file, conditional language
  in the spec being interpreted as optional, files listed in
  DELIVERABLES but missing from the output.

A translator failure that cannot be reproduced is not actionable. A
translator failure with complete inputs and outputs is the most
valuable kind of issue you can open — it turns into either a hints
file, a spec clarification, or a prompt improvement.

---

## CLI Conventions

These conventions apply to the PCD toolchain itself (`pcd-lint`,
`mcp-server-pcd`). They do *not* apply to every PCD-generated
component — `python-tool` mandates POSIX `--flag` style, and
`kubectl-style-cli` uses subcommand-flag style.

For PCD's own tools:

- Key=value syntax: `pcd-lint strict=true spec.md`
- Bare words for commands: `pcd-lint list-templates`
- stderr for diagnostics; stdout for summaries
- Exit codes: 0 = valid, 1 = errors, 2 = invocation error

---

## Working with Generated Code

The implementation code under `tools/pcd-lint/code/` and
`tools/mcp-server-pcd/code/` is generated by LLM translation from the
specifications in `tools/*/spec/`. Do not hand-edit it. If you find a
defect in the generated code, the cause is in the specification, the
template, the prompt, or a missing hints file — fix the cause and
regenerate.

When regenerating a tool, preserve the workflow as three commits:

1. **Baseline** — the pre-regeneration state. Tag this if the tool was
   previously working.
2. **Generation** — the LLM-produced output. The commit message names
   the model and version that produced the translation, and any preset
   overrides used.
3. **Fixes** — any post-translation corrections (typically module path,
   import fixes, build-system tweaks). Keep these minimal and explain
   each one. If the list of fixes grows long, the underlying problem is
   in the spec or template — go back and address it there.

This makes every regeneration auditable in `git log` without losing the
ability to compare two runs against each other.

The generated code is also marked `linguist-generated` in
`.gitattributes`. Diffs in `tools/*/code/` are collapsed by default in
the GitHub web UI; PR reviewers should read the spec diff and the new
`TRANSLATION_REPORT.md`, not the generated code.

---

## Derived Spec Content

Specifications under `tools/*/spec/` are human-authored, with one
exception: small sections that enumerate filesystem state — such as the
list of templates currently shipped — are mechanically derived by the
top-level `make spec` target, not written by hand. The same target also
regenerates the bundled translation inputs (`tools/pcd-lint/spec/cli-tool.template.md`,
`tools/pcd-lint/spec/prompt.md`) and their `.sha256` sidecars, so the
audit chain captures the exact upstream versions used for a translation.

Derived sections inside a spec file are delimited by HTML-comment
markers:

```
<!-- BEGIN AUTO: <name> -->
... content owned by `make spec` ...
<!-- END AUTO: <name> -->
```

Content strictly between these markers is overwritten on every
`make spec`. Hand-edits inside auto-blocks are caught by `make check-spec`
in CI, which fails if `make spec` produces any diff against the
committed tree. Content outside the markers — including the markers'
surrounding prose — is human-authored as usual.

The assembled spec, derivations included, is what gets hashed by the
translator and recorded in `TRANSLATION_REPORT.md`. Adding a new
template under `templates/` therefore changes the spec hash of every
tool whose spec discovers templates from that directory; this is the
intended behaviour and means a new template triggers a re-translation
of the affected tools. Treat such changes as a normal regeneration
(baseline / generation / fixes commits, per the previous section).

Hand-authored content should remain the majority of every spec. The
auto-block mechanism is a tool for sweeping up mechanical enumerations
that mirror directory listings — not a path toward fully-generated
specs. A reviewer should be able to tell at a glance which parts of a
spec express semantics (BEHAVIOR, TYPES, INVARIANTS, EXAMPLES) and
which parts are derived listings; the markers exist to make that
distinction visible.

---

## Licensing

| Artifact | License |
|---|---|
| Specs, templates, examples, docs | CC-BY-4.0 |
| Tools (`pcd-lint`, `mcp-server-pcd`) | GPL-2.0-only |

Contributions are accepted under these licenses. By submitting a pull
request you agree that your contribution is released under the license
of the directory it lands in.

---

## Pull Requests

1. Fork, create a feature branch.
2. Run `pcd-lint` on any spec you modify; zero errors required.
3. If you regenerated a tool, include the new `TRANSLATION_REPORT.md`
   and follow the three-commit workflow above.
4. Commit with `component: brief description` format
   (e.g. `pcd-lint: add RULE-19 for X`, `templates: cockpit-module v0.2.0`).
5. Submit pull request. Reference the issue if one was opened first.

**Contact:** pcd@mailbox.org
