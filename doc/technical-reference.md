# PCD Technical Reference

**Status:** Draft
**Version:** 0.4.0
**Author:** Matthias G. Eckermann <pcd@mailbox.org>
**Date:** 2026-05-18
**License:** CC-BY-4.0

This document explains the architectural and process decisions behind the
Post-Coding Development paradigm. It answers *why* the framework works the
way it does. For how to use the framework step by step, see `doc/user-guide.md`.
For the paradigm's goals, evidence, and strategic context, see `doc/whitepaper.md`.

---

## Table of Contents

1. Why a Constrained Specification Language
2. Why Specifications Must Not Declare a Target Language
3. Why Templates Exist: Enforcing the Intent/Implementation Separation
4. The Deployment Template System
5. The Hints File System
6. Why pcd-lint Must Run Before Translation
7. The Translation Process: Design Decisions
8. The Audit Bundle: What It Contains and Why
9. Translation Confidence and Independent Test Generation
10. The Specification Lifecycle: Full Regeneration vs. Incremental Update
11. The Decisions Hints File: Implementation Memory Without Spec Contamination
12. Spec Hash Embedding: Cryptographic Chain of Custody
13. Large Specifications: Why the Milestone Mechanism Exists
14. Formal Verification: When and Why
15. Dual-LLM Verification
16. License Compliance and Software Composition Analysis
17. Related Work and What Is Genuinely Novel
18. Empirical Testing Record

---

## 1. Why a Constrained Specification Language

The fundamental weakness in any AI-assisted generation system is the gap between
informal natural language and precise executable behavior. Freeform prompts produce
brittle, non-reproducible outputs because natural language is inherently ambiguous.
The same English sentence can mean different things to different translators. Over
multiple runs, different models, and different sessions, this ambiguity compounds.

PCD addresses this by requiring specifications to use a constrained format with
required sections, formal notation for types and invariants, and executable examples.
The constraint is not bureaucratic overhead — it is what makes specifications
machine-validatable before any AI translator is involved, and what makes translation
outputs comparable and reproducible across runs.

The design principles that drove the constrained format:

**Required sections with machine validation.** A specification that is missing a
TYPES section, or that declares behaviors without STEPS, is structurally incomplete.
`pcd-lint` catches these gaps before translation begins. This moves the error
detection point from "the AI produced something wrong" to "the specification is
incomplete" — a problem with a clear, human-fixable answer.

**Formal notation for critical properties.** `balance >= 0` is not ambiguous.
"balance should be positive" is. Invariants use mathematical notation or a
controlled English subset precisely because the goal is to eliminate the ambiguity
surface where hallucinations originate.

**Executable examples as the acceptance test.** GIVEN/WHEN/THEN examples in a
specification are not documentation — they are the acceptance criteria for the
translation. Generated code must pass all examples. This structure was borrowed
from the Behaviour-Driven Development tradition (Gherkin, 2008), which has two
decades of production validation behind it.

**Negative-path examples are required.** A specification that only covers the
happy path has not described the behavior — it has described one of many behaviors.
Any BEHAVIOR block whose STEPS contain error exits must include at least one example
whose THEN clause verifies the error outcome. `pcd-lint` enforces this as RULE-10.
The absence of negative-path examples was the single most common source of
specification ambiguity found during empirical testing.

**Controlled vocabulary.** Consistent keywords (`PRECONDITIONS` not "Requirements",
`BEHAVIOR` not "Function") reduce variance in how different AI models parse the
structure. This is especially important for smaller models with less instruction
following reliability.

The benefits of constrained format over freeform are: pre-translation linting
catches errors early; executable examples validate translation; formal notation
eliminates natural language ambiguity for safety-critical properties; and the
structured input produces more consistent translation output across different
models and sessions.

The tradeoff is a steeper learning curve for new spec authors. The interview
prompt (`prompts/interview-prompt.md`) was designed specifically to address this:
domain experts answer questions in plain language, and any capable LLM produces
the constrained specification from their answers. The format is learned by the
tool, not the human.

---

## 2. Why Specifications Must Not Declare a Target Language

Early versions of this paradigm required spec authors to declare a `Target:`
field in the META section. This was identified as an anti-pattern and removed
in v0.3.0.

The problem with declaring a target language in the spec is that it pulls the
spec author into implementation thinking. The moment an author writes
`Target: Go`, they have made a decision that is not theirs to make. They are
now thinking about Go semantics, Go packaging, and Go toolchain requirements —
none of which is their domain expertise. The specification becomes coupled to
a technology choice that may change: the organisation adopts Rust, a new
deployment context requires a different language, or the team wants to translate
the same spec to multiple languages for comparison.

The deeper issue is that target language is not a free variable — it is a
function of deployment context. Once you declare `Deployment: ebpf`, the
target language space collapses to restricted C; there is no meaningful choice
to make. Once you declare `Deployment: wasm`, the target is Rust. The spec
author was never deciding anything; they were being asked to transcribe a
decision that the deployment template had already encoded.

Removing the `Target:` field from META and encoding language defaults in
deployment templates was the most important design decision in the project's
evolution. Specifications written before v0.3.0 that declare a `Target:` field
are still processable — `pcd-lint` treats the field as unknown-but-harmless —
but the field has no effect on translation.

---

## 3. Why Templates Exist: Enforcing the Intent/Implementation Separation

The separation between specification (what) and template (how to build it) is
not a convenience — it is an architectural guarantee.

A specification describes behavior, types, invariants, and examples. It never
contains a language name, a compiler flag, a packaging format, or a delivery
phase sequence. A template encodes all of these for a specific deployment
context. Neither document refers to the concerns of the other.

This separation means:

**Specifications are stable across implementation changes.** The spec written
today for a Go binary remains valid if the organisation changes its default to
Rust in 2029. The spec does not change; only the template or preset changes.

**The universal translation prompt stays language-agnostic.** `prompts/prompt.md`
contains universal translation principles that apply regardless of target language.
It never mentions Go, Rust, or any other language. Language-specific delivery
phases, compiler invocations, and compile gate commands live in the template's
`## EXECUTION` section. This split — introduced in v0.3.16 — means new templates
can be added without modifying the universal prompt, and the universal prompt
remains readable and maintainable as a paradigm document rather than an
operational configuration file.

**Type bindings are a template concern.** Logical types from the spec
(`Duration`, `Timestamp`, `Condition`) map to concrete language types in the
template's `## TYPE-BINDINGS` section. The spec declares `Duration` as a
logical type; the template maps it to `metav1.Duration` for Go in a cloud-native
context. The spec author never names a Go type. Translators apply the binding
table mechanically, eliminating the divergence that arises from translator
discretion.

The template system follows a systemd-style preset layering model. Templates
define defaults; organisation, user, and project presets can override those
defaults within the permitted set. The resolution order is: template default →
system preset (`/etc/pcd/presets/`) → user preset (`~/.config/pcd/presets/`) →
project preset (`.pcd/presets/`). This means organisations can standardise on
Go without modifying any template, and individual projects can override to Rust
without affecting anyone else. The spec author participates in none of this.

---

## 4. The Deployment Template System

PCD ships nine deployment templates covering the translator deployment contexts:

`cli-tool` is the translator template for command-line tools. Default language: Go.
Valid alternatives: Rust, C, C++, C#. Produces a static binary, RPM package, DEB
package, man page, and README documenting OBS installation. Man pages are a
required deliverable (section 1 commands, section 3 libraries). `pandoc` is a
required build dependency.

`mcp-server` covers MCP protocol servers with stdio and streamable-HTTP transports.
Default language: Go using the `mcp-go` library. Requires the `mcp-server.go.mcp-go.hints.md`
hints file for verified API shapes; the library has no tagged releases and
fabricated pseudo-versions are disqualifying.

`backend-service` covers 12-factor app backend services. Default language: Go.
Valid alternative: Rust. Includes systemd service unit as a required deliverable.

`cloud-native` covers Kubernetes operators and cloud-native controllers. Default
language: Go. Produces CRDs, RBAC manifests, Helm charts, and Containerfile.
Includes Kubernetes ecosystem TYPE-BINDINGS. The `cloud-native.go.go-libvirt.hints.md`
and `cloud-native.go.golang-crypto-ssh.hints.md` hints files cover the two main
library dependencies used in early reference implementations.

`gui-tool` covers desktop GUI applications. Default language is OS-dependent:
C (GTK) or Go on Linux, C# on Windows, Swift on macOS. Qt6/Tauri/Flutter are
supported alternatives. No formal verification path; EXECUTION: none.

`python-tool` covers Python tools and automation scripts. QM safety level only;
formal verification is not supported for Python. Produces `pyproject.toml` as a
required deliverable. `--flag` style (argparse) is mandatory; key=value CLI
style is forbidden in this template.

`library-c-abi` covers C-ABI shared libraries. Default language: C. Valid
alternative: Rust via `cbindgen`. Section 3 man pages required. Stable ABI
and C-compatible headers are mandatory constraints.

`verified-library` covers safety- and security-critical C-ABI libraries where
formal verification is required or strongly recommended. Default language: C.
QM safety level is not permitted; this template is for ASIL-B through ASIL-D
and equivalent security criticality levels.

`project-manifest` is an architect artifact. No code is generated; it produces
a project-level audit bundle covering multi-component system definitions. No
man pages; EXECUTION: none.

The `enhance-existing` template allows adding PCD-generated components to an
existing codebase in a language declared by the spec author. This is the only
template where the spec must declare a language — because the existing codebase
already fixes the language choice, and the template cannot know it.

The `project-manifest`, `cloud-native`, and `gui-tool` templates are at v0.3.19
and need to be bumped to v0.3.20 for consistency with the man pages update.

---

## 5. The Hints File System

Hints files contain implementation knowledge that belongs neither in the spec
(which must be language-agnostic) nor in the template (which covers language
and deployment conventions, not library internals). They are advisory — a
translator that follows a hints file produces better output, but hints cannot
override spec invariants or template constraints.

The hints file system has five layers with a naming convention that encodes
the scope of each layer:

`<template>.<language>.milestones.hints.md` — scaffold-first patterns for
milestone-based translation. These files contain the structural patterns that
should be established in the scaffold milestone: package layout, file naming,
interface shapes, stub conventions. They are reusable across all components
using the same template and language combination.

`<component>.implementation.hints.md` — component-specific, language-neutral
implementation knowledge. These files capture domain-specific patterns that
are specific to one component type but not to a particular language.

`<template>.<language>.<library>.hints.md` — library API shapes and known
gotchas for a specific library in a specific language and template context.
The `mcp-server.go.mcp-go.hints.md` file is the canonical example: it documents
the verified `mcp-go` v0.46.0 API shapes, the correct function to use for
streamable HTTP servers (`NewStreamableHTTPServer`, not `NewSSEServer`), and
the correct error return pattern (`NewToolResultError` for domain errors, not
Go error returns). Without this file, translators fabricate API calls that
compile but fail at runtime, or use deprecated alternatives.

`<scope>.<language>.style.hints.md` — coding style and architectural philosophy
for a specific language within a given scope. This layer addresses the objection
that AI-generated code cannot adhere to project or company coding conventions.
The scope is either a project name or a company/organisation name, and the file
lives at the corresponding level in the preset hierarchy:

- `/etc/pcd/hints/suse.go.style.hints.md` — company-wide Go style; applies to
  all Go translations on that machine
- `.pcd/hints/myproject.go.style.hints.md` — project-specific style; applies
  only to that project

When both are present, the project-level file takes precedence over the
company-level file, consistent with the preset layering model. The style hints
file is authored by the project maintainer or organisation, not generated by the
translator. It captures: architectural conventions (flat structs vs. full OOP,
interface naming patterns), framework idioms (Spring conventions, dependency
injection patterns), naming standards, and forbidden patterns.

The critical distinction from the decisions hints file: the style hints file
captures *what the project or organisation requires* regardless of which
translator runs. The decisions hints file captures *what the prior translator
decided*. The style hints file is stable and maintained; the decisions hints file
is generated and disposable.

`<specname>.<language>.decisions.hints.md` — the decisions hints file (see
section 11 below). Lives next to the spec, not in `hints/`. Language-specific
and disposable.

The reasoning behind putting hints files outside the spec: the spec must be
language-agnostic and stable. Library API shapes change between versions. A
hints file can be updated when a library releases a breaking change without
touching the spec. The spec captures intent; the hints file captures the current
state of the implementation ecosystem.

---

## 6. Why pcd-lint Must Run Before Translation

`pcd-lint` validates specification structure before any AI translator is
involved. This ordering is not optional — it is the mechanism that prevents
the AI from receiving ambiguous or structurally incomplete input.

Every error that `pcd-lint` catches before translation is an error that would
otherwise produce incorrect or unpredictable generated code, possibly without
any visible signal. A missing EXAMPLES section does not cause a compiler error
— it causes the generated code to be untested against the specification's
acceptance criteria. A BEHAVIOR without STEPS does not cause a parse failure
— it causes the translator to invent an implementation without specification
guidance.

`pcd-lint` implements 18 rules. RULE-01 through RULE-09 cover structural
completeness (required sections, META fields, TYPES, EXAMPLES format). RULE-10
covers negative-path example requirements. RULE-11 covers TOOLCHAIN-CONSTRAINTS
structure. RULE-12 covers cross-section consistency. RULE-13 covers BEHAVIOR
Constraint: field values. RULE-14 covers EXECUTION section presence in deployment
templates. RULE-15 through RULE-17 cover the MILESTONE mechanism. RULE-18
detects spec hash drift between the current specification and the recorded
hash in the adjacent TRANSLATION_REPORT.md (requires `check-report=true`).

The rules are implemented in `internal/lint/lint.go` as an importable Go
library, not only as a command-line tool. This is why `mcp-server-pcd` can
perform inline lint validation without shelling out to the `pcd-lint` binary —
it imports the same rule engine. Both tools were generated from their own
PCD specifications; the shared library was an architectural decision made
by the translators independently, not specified explicitly. Both Sonnet and
Haiku converged on the same package structure (`internal/lint/`) given the
same input.

---

## 7. The Translation Process: Design Decisions

**Why the prompt is split into two layers.** The universal prompt
(`prompts/prompt.md`) contains principles that apply to every translation
regardless of template or language: how to read MILESTONE sections, the stub
contract, the delivery mode decision, the translation report requirements.
The template's `## EXECUTION` section contains everything language- and
context-specific: the delivery phases, the compile gate commands, the resume
logic. This split means the universal prompt is stable and readable as a
paradigm document, while templates can define their own build verification
without touching it. Before this split (pre-v0.3.16), the prompt contained
language-specific commands that had to be updated every time a new language
or template was added.

**Why the stub contract specifies zero values.** When the scaffold milestone
creates stub implementations, each stub must return the correct typed zero
value — not null, not a placeholder string. For output types that serialise
to JSON objects, the stub must return an initialised empty object (`{}`), never
null. A null reference serialises to JSON `null`, which is schema-incompatible
with consumers that expect an object. This caused silent failures in early
scaffold runs: the binary compiled, but API clients received invalid responses.
The explicit stub contract prevents this class of error.

**Why the EXECUTION section specifies a compile gate.** The compile gate —
`go build ./...` for Go, `cargo build` for Rust — is the minimum acceptance
criterion for any translation run. A translation that does not compile is not
a deliverable. Making the compile gate explicit in the EXECUTION section means
the AI translator cannot complete a translation run without verifying that the
output compiles. Before this was explicit, some translators would deliver source
files that contained syntax errors, leaving the error to be discovered by the
human receiving the output.

**Why the translation report is always the last deliverable.** The
TRANSLATION_REPORT.md must be produced after all other deliverables are written
and the compile gate has passed. This ordering is enforced by the EXECUTION
section. A translation report written before the compile gate has passed cannot
accurately document the compile gate result. A translation report written before
all files are delivered cannot accurately document what was produced. The ordering
requirement was added after early runs produced reports that described planned
deliverables rather than actual ones.

**Why tests are written before implementation code.** Within a single translator
translation run, the prompt mandates that test functions under
`independent_tests/<llm-name>/` are written before the implementation. The
purpose is to prevent post-hoc test tuning — tests written after seeing the
code tend to assert what the code does rather than what the spec requires.
The ordering is the same TDD discipline practised by human engineers, applied
to the LLM workflow. Tests may be refined when the implementation reveals a
gap, but every refinement must be logged in the report's Test Refinements
table with documented rationale; an unlogged test edit after a test run is a
translation defect.

**Why the prompt has two roles.** The translator prompt operates in either
`translator` mode (the default — produce tests and code) or `test-author` mode
(produce only tests, then stop). The role is selected at invocation via an
optional `ROLE.md` file in the input directory. A single prompt with two
roles is simpler than two separate prompts: the rules about derived language,
test framework, hints, and ambiguity handling apply identically in both
roles, and only one prompt has to be maintained, embedded in
`mcp-server-pcd`, and reasoned about. The role selection is the only
difference between a single-LLM translation and the test-generation phase of
a dual-LLM verification run.

**Why the spec hash must be computed before generating any output.** The SHA256
of the specification file must be computed once, at the start of the translation
run, before any output files are written. This ensures all artifacts from the
same translation run carry the same hash — the hash of the specification as it
existed at translation time. If the spec were modified during a translation run
(unlikely but possible in an agentic workflow), computing the hash at the end
would embed a different hash than computing it at the start.

---

## 8. The Audit Bundle: What It Contains and Why

The audit bundle is the artifact that makes PCD suitable for regulated domain
certification. It is the complete, traceable record of a translation run.

A complete audit bundle contains: the specification (human-authored, CC-BY-4.0);
the translation report (`TRANSLATION_REPORT.md`) documenting every decision the
translator made; the generated source code; the packaging artifacts (RPM, DEB,
Containerfile); the independent test suite if generated; and the `metadata.json`
traceability record including the spec hash, translator model version, and
timestamp.

The reasoning behind each element: The specification is the artifact that is
certified — human-reviewed, human-approved, and the only document the spec author
is responsible for. The translation report is the closest equivalent to a compiler
log for the AI translation step; it documents what the AI decided and why, making
the translation decision auditable by a human reviewer. The generated code is what
is deployed and what is verified by automated tools (compile gate, examples,
independent tests, SCA). The spec hash in `metadata.json` and embedded in all
artifacts provides the cryptographic link from certified specification to deployed
artifact.

For Common Criteria and ISO 26262, the 4-eyes principle requires that code be
reviewed by a human before release. In PCD, this requirement applies to the
specification, not the generated code. A human reviews the specification — the
document that defines what the system does — and signs off on it. The generated
code is verified automatically. This is the correct application of the 4-eyes
principle: it requires human comprehension and sign-off, and human comprehension
of a structured Markdown specification is tractable in a way that human
comprehension of 5000 lines of generated Go is not.

The Pikchr workflow diagram (`translation-workflow.pikchr`) is a machine-generated
visualisation of the specific translation run — inputs, decisions, outputs — that
renders to SVG. It serves as a machine-generated audit trail, not a hand-drawn
diagram, making it reproducible and version-controllable.

---

## 9. Translation Confidence and Independent Test Generation

AI translation is probabilistic. The same specification translated twice by the
same model may produce different implementations — each correct, but making
different architectural decisions. Over multiple runs, this variance accumulates.
Three mechanisms address this: the per-EXAMPLE confidence table in
TRANSLATION_REPORT.md, the tests-first discipline within every translator run, and (optionally) an independent test suite produced by a
separate LLM acting as test-author, before the translator runs.

**The translation confidence table.** Every translation report must include a
per-example confidence table with three levels: High (a named test function in
`independent_tests/<llm-name>/` passes without any live external service;
when a test-author suite is present, both suites pass), Medium (some paths
tested, others require live services), and Low (no test covers this, code
review only). A claim is verified only if it references a specific named test
function. Unverified claims must be listed explicitly. This discipline was
introduced after early translation reports claimed high confidence for examples
that had no corresponding tests — the confidence was the translator's assessment
of its own work, not an empirical measurement.

**The tests-first discipline.** In every translation run, the translator
writes test functions under `independent_tests/<llm-name>/` *before* writing
implementation code. The ordering matters: tests written after implementation
tend to assert what the code does rather than what the spec requires, because
the code's existence anchors the test author's interpretation of the spec.
Writing tests first forces a direct reading of the EXAMPLES, PRECONDITIONS,
POSTCONDITIONS, and INVARIANTS without the bias of an existing implementation.
The same discipline that has made TDD valuable in human engineering for two
decades is what makes tests-first valuable for LLM-driven translation.

**Test refinement discipline.** Tests may be refined after a test run reveals
a gap, but every refinement must be recorded in a `Test Refinements` table in
the TRANSLATION_REPORT.md, with one row per refinement and an explicit
rationale. The permitted actions are `code fixed` (the test was right; the
implementation was changed), `test edited` (the test was wrong; rationale
must reference a spec section, never "made the test pass"), `spec ambiguous`
(the spec does not determine the answer; failure documented), and
`interface rebind` (test-author mode only). An unlogged test edit after a test
run is a translation defect. The table is the audit artifact that makes the
test-tuning failure mode visible to a reviewer: a passing build with no Test
Refinements section after a failing test is a red flag.

**Independent test generation by a second LLM.** A second LLM, acting as test-author, can
read only the specification — not any translation's code — and
generate a test suite. This test suite is then run against the translator's code in addition to the translator's own tests. Failures indicate
specification ambiguity (the two LLMs read the spec differently) or
translation error (the translator pass produced code that does not satisfy the spec's
semantics). The second agent has no access to the translation; its
tests are truly independent. Section 16 covers the operational workflow and
when this escalation is appropriate.

Second-LLM tests consistently find edge cases and boundary conditions that
the spec author did not think to include in the EXAMPLES section. The tests
serve two purposes: as a validation mechanism for the current translation,
and as a candidate addition to the spec's EXAMPLES section for future
translations.

**Why the same language for tests and code.** Tests and the implementation
are written in the same language. This is a production constraint, not a
methodological preference: minimal runtime environments — Common Criteria
images, OBS build workers, certified container images — carry the toolchain
for one language only. A Go implementation tested by Python tests would
require both runtimes in the certified image, doubling the supply-chain
surface for no operational benefit. The cost of forgoing cross-language
independence is accepted; tests are made independent by being written
*against the spec alone*, not by being written *in a different language*.

---

## 10. The Specification Lifecycle: Full Regeneration vs. Incremental Update

A specification is not written once and frozen. Requirements change, behaviors
are added, types evolve. The question of how to handle specification changes —
full regeneration or incremental update — is a practical operational decision
with significant consequences for codebase consistency.

**Why full regeneration is the safe default.** AI translation is probabilistic.
Two runs from the same specification with the same model may make different
structural decisions — naming conventions, error handling patterns, package
layout. Over multiple incremental updates, a codebase accumulates the
independent decisions of multiple translation runs. Each run was internally
consistent; the combination may not be. Full regeneration from a clean
specification produces a codebase that was produced in a single pass from a
single source of truth, and is guaranteed internally consistent by construction.

**The scaffold boundary as the decision point.** The scaffold milestone (see
section 13) establishes the package structure, file layout, and interface shapes
for the entire component. Any change to the scaffold — a new package, a
restructured interface, a new type referenced throughout the codebase — requires
full regeneration, because subsequent milestones were built on the scaffold's
decisions. A change isolated to the STEPS of one or two behaviors, with no
effect on shared types or interfaces, is a candidate for incremental update.

**The blast radius analysis.** Before choosing incremental update, the change
impact must be assessed: how many files, BEHAVIORs, and functions are affected?
One or two isolated BEHAVIORs with no shared type changes: incremental viable.
Three to five BEHAVIORs, or a shared type changed: judgment call. Five or more
BEHAVIORs, or an INTERFACE changed: full regeneration.

**The `assess_change_impact` tool.** The `mcp-server-pcd` tool `assess_change_impact`
automates this analysis. Given the specification change (as a diff or plain-language
description) and optionally the existing generated code, it applies the decision
framework and returns a structured recommendation: full regeneration or incremental
update, with translator factor and reasoning. When existing code is not provided,
the tool biases conservative — it cannot verify blast radius without the code.
The recommendation in this case explicitly notes the limitation.

**The cost asymmetry.** Writing a specification is the expensive part. Running
a translator is cheap — one LLM session at approximately 128K tokens, taking
minutes. This asymmetry means the default should always bias toward full
regeneration when in doubt. The cost of an unnecessary full regeneration is one
translator run. The cost of an incremental update that silently introduces
inconsistency may be discovered weeks later and require a full regeneration
anyway, plus debugging time.

---

## 11. The Decisions Hints File: Implementation Memory Without Spec Contamination

When a change impact assessment recommends full regeneration, it may produce a
list of implementation decisions from the existing code that are worth preserving
in the next translation run: the chosen package layout, the error convention
pattern, the routing structure, the asset embedding approach. These decisions
are not in the specification — they are not behavioral requirements — but they
represent accumulated good judgment that would be wasteful to discard.

Writing these decisions into the specification is wrong: specifications must be
language-agnostic and permanent. A Go-specific package layout decision does not
belong in a document that may be translated to Rust tomorrow. Writing them into
a hints file in `hints/` is also wrong: hints files in `hints/` are shared across
components, not specific to one spec's implementation history.

The decisions hints file (`<specname>.<language>.decisions.hints.md`) solves this:

It lives next to the specification (not in `hints/`), so it is spec-scoped.
It carries the language name in its filename, so it is explicitly language-specific
and disposable when switching languages. It is generated by the translator as a
required deliverable of every translation run that makes architectural decisions
not inferable from the spec — produced alongside `TRANSLATION_REPORT.md`. It is
read by the translator at the start of a guided regeneration or incremental update
run, and ignored on clean full regeneration from scratch. It is not a specification
artifact: it does not affect `pcd-lint` validation and is not reviewed in
certification.

The three-state translation model that the decisions hints file enables:

A *clean full regeneration* reads only the spec and the template. The translator
starts with no prior knowledge. This is the correct choice after a breaking
structural change, after switching languages, or when the existing codebase's
quality is unknown.

A *guided regeneration* reads the spec, the template, and the decisions hints
file. The translator starts with knowledge of prior architectural decisions and
produces an implementation that is consistent with them. This is the correct
choice after a non-structural spec change where the existing architecture was
good.

An *incremental update* reads the spec diff, the existing code, and the decisions
hints file. The translator touches only the changed behaviors. This is the correct
choice for isolated, low-blast-radius changes.

---

## 12. Spec Hash Embedding: Cryptographic Chain of Custody

Every generated artifact must embed the SHA256 hash of the specification file
it was produced from. This embedding creates a cryptographically verifiable
link between the certified specification and the deployed artifact.

The decision to embed the hash in every artifact, rather than only in the
translation report, was deliberate. A hash only in the translation report
requires that the translation report accompany every artifact. A hash embedded
in the artifact itself — in a source file comment, in the version subcommand output, in
a Containerfile `LABEL`, in RPM metadata — means any single artifact can be
verified independently, without access to the audit bundle.

The hash is computed once, before any output is written, from the specification
file as provided. All artifacts from the same translation run carry the same
hash. If the specification changes and a new translation run produces new
artifacts, the new artifacts carry a different hash. The version boundary is
cryptographically visible without inspecting build logs or commit history.

For regulated domain compliance, this answers the audit question "was this
binary produced from the certified specification?" without depending on human
attestation or trusting the build pipeline. The hash embedded in the binary
either matches `sha256sum <specname>.md` or it does not. There is no middle
ground.

For the 4-eyes principle: the reviewer signs off on the specification. The
hash in the artifacts proves that what was signed off on is what was built.
The chain is: human certifies spec → spec hash → artifacts embed hash → hash
is verifiable. No link in this chain requires trusting any tool, any pipeline,
or any person after the initial certification.

`pcd-lint` RULE-18 (planned) will detect hash drift: if a `TRANSLATION_REPORT.md`
exists adjacent to the specification and its `Spec-SHA256:` field does not match
the current specification's hash, `pcd-lint check-report=true` emits a warning.
This surfaces the "spec has changed since last translation" condition before
the developer starts a new translation run.

---

## 13. Large Specifications: Why the Milestone Mechanism Exists

For specifications with more than approximately ten behaviors or five hundred
lines of estimated output, single-pass translation produces unreliable results.
The translator must hold the full specification in context while generating the
full implementation, and the output window approaches or exceeds the model's
practical limits. Early large specification translations produced truncated
deliverables — the translator ran out of output budget and stopped mid-file,
sometimes without signalling the truncation.

The milestone mechanism partitions a specification into sequential translation
passes, each producing a defined, verifiable subset of the implementation.

**Why the scaffold milestone must always be first.** The scaffold milestone
(Scaffold: true) creates all files, all types, all function signatures, and all
stubs in a single pass. Its only acceptance criterion is a clean compile. It
does not implement any real logic. Every subsequent milestone finds a stable
foundation — the same file structure, the same type definitions, the same
function signatures — and fills in stub bodies. Without the scaffold milestone,
different milestone runs may create different file structures or define the same
type differently, producing a codebase that is internally inconsistent.

**Why the scaffold must compile.** The compile gate on the scaffold milestone
is the guarantee that the skeleton is sound. A scaffold that compiles confirms:
all import paths are correct, all types are internally consistent, all function
signatures are well-typed. Subsequent translators can fill in bodies without
restructuring anything. The `sitar` tool — 35 behaviors, 2900-line specification —
was translated to both Go and Rust using this pattern. The scaffold held without
modification through all seven milestones in both languages.

**Why milestone status is managed by the pipeline, not the spec author.**
The milestone state machine (pending → active → released/failed) is managed by
the agent pipeline, not the human author. The `set_milestone_status` tool in
`mcp-server-pcd` advances the cursor. The human intervenes only when a milestone
fails — which signals a specification problem that requires human judgment. This
keeps the operational workflow automated while preserving the human as the
decision point for failures.

**Why at most one scaffold milestone is permitted.** The scaffold establishes
the package structure. Having two scaffold milestones would mean the package
structure was reconsidered mid-translation, invalidating everything built on
the first scaffold. `pcd-lint` RULE-17 enforces uniqueness.

---

## 14. Spec Composition: Sharing Behaviours Across Components

Two PCD components occasionally need to share substantial structured
behaviour. The driving example: `pcd-lint` validates PCD specifications
against the framework's numbered RULES, and `mcp-server-pcd` exposes an
MCP tool that performs the same validation for LLM clients. Both
components need the same eighteen-plus rule implementations, derived
from the same definitions, with no drift over time.

Until v0.4.0, PCD had no spec-layer mechanism for this. Each component's
spec independently described every rule it needed, and each translator
independently produced an implementation. Two `internal/lint/lint.go`
files resulted, with the same intent and divergent code — drift was
empirically observed within weeks of both tools existing.

v0.4.0 introduces **spec composition**. A host spec's META declares
`Includes:` directives pointing to other spec files. Each included spec
contributes its TYPES, BEHAVIORs, INVARIANTS, EXAMPLES, INTERFACES,
DEPENDENCIES, TOOLCHAIN-CONSTRAINTS, PRECONDITIONS, and POSTCONDITIONS
to the host. The translator consumes the merged spec as if it had been
written inline, producing a self-contained implementation of all the
included content alongside the host's own code.

The mechanism is language-neutral. The shared spec describes behaviours
in PCD's structured Markdown; each translator projects those behaviours
into its target language's idioms. The same `lint-rules.md` produces a
Go package when consumed by a Go-targeting host and a Rust module when
consumed by a Rust-targeting host. No shared library artefact exists at
the binary level; every host is self-sufficient.

**Hash semantics.** The `Spec-SHA256` embedded in generated artefacts
is the SHA256 of the merged spec text — what the translator actually
read, not the host file on disk. Editing an included spec invalidates
the merged hash of every host that includes it, which propagates change
detection through the existing spec-hash discipline. No new mechanism is
required; the existing chain-of-custody invariant simply extends to one
more boundary.

**Strict collision handling.** Two TYPE, BEHAVIOR, INTERFACE, or EXAMPLE
definitions with the same name across the merged set are a spec-author
error caught by `pcd-lint` (RULE-20). There is no implicit precedence
rule. The motivation: implicit precedence is a source of audit-trail
ambiguity. A reviewer reading the host spec alone should not have to
mentally simulate the merge to know what's actually implemented; the
spec must be explicit.

**Acyclicity and depth.** The inclusion graph must be acyclic
(RULE-21). Practical depth is expected to be one or two levels; no cap
is imposed. Included specs may not declare orchestration sections —
MILESTONE, DEPLOYMENT — that belong to host components.

**Audit trail.** The TRANSLATION_REPORT records both the merged hash
and the host hash, plus a table of included specs with their individual
hashes. A reviewer reading an artefact's embedded hash can verify the
artefact against the merged spec; a reviewer tracing provenance can
follow the recorded inclusion table back to each contributing source.

**Migration is opt-in.** A spec without `Includes:` directives is a
valid v0.4.0 spec and behaves exactly as in v0.3.x. The merged hash
equals the host hash; the inclusions table is empty. Existing components
do not need to change unless they want to share content with another
component.

**Why not shared Go packages?** The pre-PCD codebase resolved this kind
of duplication by factoring out shared Go packages. PCD rejects that
solution for several reasons. It would mean baking a language choice
into the spec layer — a `pcd-rules` Go package can't be consumed by a
Rust host. It would mean cross-component build coupling, where building
one tool requires another tool's package to be present. It would
require a runtime artefact distribution mechanism for the shared
library that adds an audit-surface dependency. Spec composition avoids
all three: the sharing happens at the spec layer, before translation,
and each host produces a self-contained implementation in its own
package and binary.

The full design — merge rules, hash computation, lint rules, worked
example — is in `doc/spec-composition.md`.

---

## 15. Formal Verification: When and Why

PCD supports an optional formal verification path: specification → meta-language
(Lean 4, F*, Dafny) → target language. This path is not the default and not
required for most use cases. It exists for contexts where mathematical guarantees
are required or where the cost of a runtime defect is high enough to justify
the verification investment.

The meta-language layer was designed as pluggable from the start. The key
lesson from early experimentation with ATS2 (a powerful linear type system)
was that LLM training data coverage is non-negotiable for an AI-native paradigm.
ATS2's syntax is underrepresented in LLM training data; multiple models
consistently produced syntactically incorrect ATS2. Lean 4 was chosen as the
translator reference meta-language because it combines strong verification power
with broad LLM training data coverage and active community support.

Lean 4 is a strong candidate for theorem proving requirements (ISO 26262 ASIL-C/D,
DO-178C DAL-A/B). F* (Microsoft Research) is proven in production at scale —
the HACL* cryptographic library used in Firefox, the Linux kernel, and WireGuard
was produced using F* extraction. Dafny has the lowest learning curve of the
three and is accessible to engineers without a formal methods background. Coq
provides maximum proof power for research and academic contexts.

For most PCD use cases, the direct path (specification → AI → code, validated
by examples and independent tests) provides sufficient confidence. The formal
verification path should be chosen when the component handles financial
transactions with conservation invariants, implements cryptographic primitives
with constant-time requirements, targets safety-critical automotive or aviation
functions, or requires formal certification evidence that runtime testing alone
cannot provide.

---

## 16. Dual-LLM Verification

PCD distinguishes two operational modes for translation. The default is
**single-LLM mode**: one LLM acting as translator writes tests under
`independent_tests/<llm-name>/` first, then the implementation, then runs
the tests. The optional escalation is **dual-LLM mode**: a second LLM acting as test-author writes a test suite first, with no access to any implementation;
translator then runs as in single-LLM mode and additionally runs test-author's
test suite against its implementation.

Most translations run in single-LLM mode. Dual-LLM mode is the appropriate
escalation when one or more of the following applies: the component has
`Safety-Level` above QM; the component has `Verification:` not equal to
`none`; the spec is novel or has a high suspected ambiguity surface; the
component is part of the PCD toolchain itself (where a regression would
affect every downstream user). The gate is operational, not declarative —
no META field controls it. The user runs test-author, or they do not.

**Milestone-level dual-LLM.** For large specifications partitioned into
milestones, dual-LLM mode may be applied per milestone rather than across
the full specification. This restricts the additional translation cost to
the milestones with the highest verification value — typically the scaffold
milestone (where the package structure and interface shapes are committed)
and any milestone covering safety-critical or security-critical behaviours
— while keeping the bulk of the work in single-LLM mode. The mechanism is
the same as the full-spec case: test-author runs first against the milestone-
scoped translation, translator runs second against the same scope, and the
hash and template continuity checks apply at the milestone level. The
choice of which milestones to escalate is recorded in the audit bundle by
the presence or absence of test-author's directory next to each milestone's
TRANSLATION_REPORT.md.

**The operational flow in dual-LLM mode.**

1. Test-author runs first. Its `ROLE.md` declares `mode: test-author` and an
   `llm-name`. Test-author reads the spec and the deployment template, writes
   tests under `independent_tests/<test-author-llm-name>/` in the same
   language the implementation will use, and produces a `TEST_REPORT.md`.
   Test-author does not write any implementation code, scaffolding, or
   packaging.
2. Translator runs next. It finds test-author's tests already present in the
   input directory. Translator first verifies that the `Spec-SHA256` recorded
   in test-author's `TEST_REPORT.md` matches the current specification's
   hash. If the hashes differ — the specification has been edited between
   test-author's run and translator's run — translator aborts with a diagnostic
   and the user re-runs test-author against the current specification.
   Stale test-author tests are not run against translator's implementation.
   Translator then verifies that the deployment template, preset resolution,
   and hints files in scope are the same as those test-author used; mismatch
   on any of these is treated the same way as a hash mismatch. With both
   checks passed, translator writes its own tests under
   `independent_tests/<translator-llm-name>/` *before* reading test-author's
   tests, then writes the implementation, then runs both test suites.
3. Translator may refine its own tests after the test run, subject to the
   refinement discipline. Translator **may not** edit test-author's tests under
   any circumstances. Test-author's tests are the independent cross-check;
   the property that gives them value is that they were written without
   knowledge of any implementation.
4. If test-author's tests fail on translator's implementation: the failure is
   diagnostic, not stop-the-world. Translator records the failure in
   TRANSLATION_REPORT.md. The reviewer determines the cause — translator
   bug, test-author test misreads spec, or spec ambiguity — and acts
   accordingly. Translator may fix the implementation, the spec may be
   clarified and both LLMs rerun, or test-author may be re-invoked to
   regenerate tests *from a clarified spec*. Test-author's tests are never
   adjusted to match translator's behaviour.

**The library-interface special case.** For deployment templates targeting
shared libraries (`library-c-abi`, `verified-library`), test-author's tests
need to reference function and type names that the spec may not pin
precisely. Test-author writes tests in two phases. Phase A (before translator):
test logic is written with `<INTERFACE_PLACEHOLDER>` markers for any
function or type name the spec does not pin. Phase B (after translator): a
mechanical rebind pass replaces placeholders with translator's actual names.
The rebind is logged as `interface rebind` in the Test Refinements table
and may not change assertions, expected values, or coverage. The
independence property is preserved because the test content — what is
asserted, against which examples, with which expected outputs — was
determined before translator committed any code.

**Same language for tests and code.** Both test-author and translator write
tests in the language declared by the deployment template. This is the
same production constraint described in section 9: certified runtime
images carry the toolchain for one language only. The cost of forgoing
cross-language independence is accepted. Independence is provided by
test-author writing *against the spec alone*, not by writing in a different
language.

**Directory naming.** Test directories are named after the LLM that
produced them: `independent_tests/<llm-name>/`. Names are lowercase,
hyphen-separated, and contain no version-decimal suffix
(`claude-sonnet-4-5`, not `Claude-Sonnet-4.5`). The directory listing
under `independent_tests/` is itself an audit artifact: it tells a
reviewer which models were involved without parsing any metadata. Dates
and full version strings live in TRANSLATION_REPORT.md, not in directory
names — otherwise the directory accumulates copies on every regeneration.

**Cost asymmetry.** Single-LLM mode costs one translation run. Dual-LLM
mode costs two runs and one additional test-execution pass against the
translator implementation. For a component where a runtime defect has
safety, security, or regulatory consequences, the additional cost is
justified by the spec-ambiguity probe and the independent test suite.
For components without such consequences, the cost is overhead. The
operational gate keeps the choice explicit and visible.

**Delivery-mode constraint.** Dual-LLM verification requires a delivery
mode that persists files between runs — filesystem or MCP. Browser/inline
mode is single-LLM by definition because test-author's tests cannot be
carried across to translator's run without persistent storage.

---

## 17. License Compliance and Software Composition Analysis

No LLM can provide a legal guarantee that generated code is free of patterns
derived from differently-licensed training data. This is an unsolved problem
in the field. The `License:` META field and SPDX validation in `pcd-lint` are
necessary but not sufficient for license compliance.

PCD is better positioned than generic AI coding assistants because: the
`License:` META field declares intent upfront; the translator receives an
explicit license constraint and acknowledges it in the translation report; the
generated source code is available for SCA scanning; and the translation report
documents any known license-relevant deviations.

Software Composition Analysis is recommended in the CI pipeline after code
generation and before deployment sign-off. Recommended tools: REUSE (FSFE) for
SPDX header enforcement per file; FOSSology for deep license scanning and snippet
detection in regulated or commercial deployments; Black Duck for enterprise SCA
and policy enforcement.

The licensing model for PCD itself follows the Linux ecosystem pattern:
specifications, templates, examples, and documentation are CC-BY-4.0 (maximum
adoption, no barrier to building on the format); reference tools (`pcd-lint`,
`mcp-server-pcd`) are GPL-2.0-only (forces collaboration on the validation
toolchain, prevents proprietary forking of the compliance layer). The GPL-2.0-only
reference implementation is a strategic choice: it is the same mechanism that
made Linux's platform layer vendor-neutral.

---

## 18. Related Work and What Is Genuinely Novel

PCD combines several established ideas in a novel way. Each ingredient has
precedent; the combination does not exist as a productised, accessible,
regulated-domain-ready system.

OpenAPI/AsyncAPI describes interfaces, not full component behavior. It has no
formal verification layer and no deployment template concept. Gherkin/BDD
provided the GIVEN/WHEN/THEN example structure used in PCD EXAMPLES sections —
a direct borrowing from a twenty-year-old tradition with proven value. TLA+
and Alloy are formal specification languages used in industry, but humans write
them directly; there is no AI translation layer and no pathway from specification
to deployable code. F*/HACL* is the closest existing work to PCD's verified path:
HACL* (used in Firefox, the Linux kernel, and WireGuard) was produced from F*
specifications. The difference is that humans write F* directly; PCD places AI
as the translator so domain experts author the translator artifact. Dafny compiles
verified code to multiple targets and is accessible without a formal methods
background; it is a candidate meta-language within PCD, not a competing paradigm.

AWS Kiro (2025) is a proprietary, IDE-integrated, AWS-hosted product built
around writing specifications before AI generates code. It does not define a
portable, lintable specification format, has no deployment template abstraction,
no formal verification path, no supply chain or packaging conventions, and no
pathway to regulated-domain certification. The paradigms are complementary. The
convergence of Werner Vogels' re:Invent 2025 keynote with the core thesis of
this work — independently developed — is external validation of the problem
framing.

What is genuinely novel: natural language as the translator artifact (structured
Markdown, not a programming language or formal language); deployment templates
as a first-class concept (target language is not a human decision); formal
verification as optional and pluggable; regulated-domain certification as a
design goal from the start; and self-hosting from the first artifact.

---

## 19. Empirical Testing Record

The paradigm has been validated empirically across multiple models, environments,
and specification sizes. This section records the key findings. The full test
data is maintained in the translator test log.

**Universal finding: language resolution.** Every model tested resolved the
target language by reading the deployment template, without being told explicitly.
All cited the template's `LANGUAGE | Go | default` entry as the source of their
decision. This was tested across eight runs, three continents, and multiple
model families. The core design claim held in every case.

**Model capability classes tested.** Frontier cloud models (US providers), a
120B open-weight model at a regional EU provider (digital sovereignty proof of
concept), a 30B open-weight coder model on local hardware (Ollama), and a small
frontier model via direct API. The 120B EU-hosted model produced the most complete
deliverable set of any single run in early testing — validating digital
sovereignty as a practical option, not just a theoretical one.

**pcd-lint v0.3.21 regeneration (2026-04-07).** Three approaches tested:
Haiku fresh (13/17 rules, missing RULE-10/11/12/16), Sonnet incremental (stalled
at 16K token output limit), Sonnet fresh (17/17 rules, correct `internal/lint/`
architecture). Sonnet fresh was chosen. Both models independently chose the
same package structure (`internal/lint/lint.go` as an importable library), without
explicit specification of that requirement. The module path was a systematic gap
in both runs — translators cannot infer the author's GitHub username from the
spec. This was fixed in a post-generation commit and identified as a candidate
for a new META field.

**mcp-server-pcd v0.2.0 regeneration (2026-04-07).** Two rounds. Round 1 without
full asset input: all translators produced placeholder or partial embedded assets.
Round 2 with full asset input (all 9 templates, 6 hints files, 3 prompts): Sonnet
fresh produced 18 of 18 embedded assets correctly. Haiku produced 4 of 18 assets
and 13 session housekeeping files (`COMPLETION_SUMMARY.txt` etc.) — disqualifying
for a public repository. Incremental runs produced correct logic but failed to
embed assets in both rounds.

**Key infrastructure findings from regeneration runs.** The hard API limit for
Sonnet 4.5/4.6 is 128,000 output tokens (131,072 returns HTTP 400). Filesystem
restriction to `/tmp/` is essential in agentic mode — Sonnet explores the
filesystem and reads existing code if not constrained. No root access during
translation runs; `go mod vendor` is the correct approach. Haiku stays within
provided input files without prompting; Sonnet requires explicit restriction.

**COBOL demo validation (2026-04-10).** The `calc-interest` COBOL program was
translated via the reverse prompt to a PCD specification, then translated to
Go, Rust, and Java in independent runs. All three target language translations
produced correct, compiling implementations. Language was chosen at translation
time from the deployment template in all cases. This validates the claim that
the same specification can produce idiomatic implementations in multiple languages
without any specification change.

---

## References

- Vogels2025: Werner Vogels, AWS re:Invent 2025 keynote. "Specifications are the
  new code." December 2025.
- REUSE: FSFE REUSE Specification v3.3 — standardised method for declaring
  copyright and licensing in software projects using SPDX identifiers.
  https://reuse.software/
- ISO 26262: Road vehicles — Functional safety. ISO, 2018.
- DO-178C: Software Considerations in Airborne Systems and Equipment
  Certification. RTCA, 2011.
- Common Criteria: Common Criteria for Information Technology Security
  Evaluation. ISO/IEC 15408.

---

*This document is CC-BY-4.0. Canonical location: `doc/technical-reference.md`.*
