# PCD Spec Composition (v0.4.0)

**Status:** Draft
**Version:** 0.4.0
**Author:** Matthias G. Eckermann <pcd@mailbox.org>
**Date:** 2026-05-18
**License:** CC-BY-4.0

---

## 1. Problem

Until v0.3.x, each PCD component had exactly one specification file. The
file declared all of the component's TYPES, BEHAVIORs, INVARIANTS,
EXAMPLES, and metadata. A translator consumed a single spec and a single
deployment template, and produced a complete implementation.

This worked well when components were isolated. It failed when two
components needed to share substantial structured logic.

The concrete instance: `pcd-lint` validates PCD specifications against
eighteen numbered RULES. `mcp-server-pcd` exposes a `lint_content` MCP
tool that performs the same validation, because external LLM clients
need to lint specs before consuming them. Both components describe the
same eighteen rules in their respective specs, with the same intended
behaviour. Both are translated by an LLM. Each translation independently
produces an implementation of all eighteen rules.

The result is two `internal/lint/lint.go` files with the same intent,
different implementations, and no structural mechanism to keep them
synchronised when the rules change. A bug fixed in one does not
propagate to the other. A new rule added to one spec is silently absent
from the other. The duplication is invisible to either spec individually,
and equally invisible to either LLM translator individually.

The pre-PCD codebase resolved this through a human refactoring pattern:
shared Go packages with cross-tool imports. PCD has no equivalent
mechanism at the spec layer, because the spec is the source of truth and
the spec didn't acknowledge the sharing.

## 2. Solution: Spec Composition

A spec may declare that it **includes** another spec. The included spec
contributes its TYPES, BEHAVIORs, INVARIANTS, and EXAMPLES to the host
spec's effective specification. The translator consumes the merged spec
as if it had been written inline.

Each consuming host produces its own implementation of the shared
content, in its own target language, using its own deployment template.
The implementations may be byte-different on disk — they live in
different packages, different files, different binaries — but they
implement the same behaviours, because the source spec text from which
they were generated is identical.

The mechanism is **language-neutral**. The shared spec describes
behaviours, types, examples, and invariants in PCD's structured
Markdown. Each translator projects these into its target language's
idioms. A shared spec written today can be consumed by a Go translator,
a Rust translator, and a Python translator with no changes — each
producing the same behaviour in its own language.

The mechanism is also **packaging-neutral**. Each consuming host carries
its own copy of the implementation; no shared binary, no shared library
artefact, no inter-package runtime dependency. A consumer's RPM, DEB,
or container image is self-sufficient — exactly the property required
for regulated supply-chain certification, where transitive dependencies
multiply the audit surface.

## 3. Syntax

A new META field:

```
Includes: <relative-path-to-included-spec>
```

Multiple `Includes:` lines are permitted; they resolve in order. Paths
are relative to the host spec file's location. For example, from
`tools/pcd-lint/spec/pcd-lint.md`:

```
## META
Deployment:  cli-tool
Version:     0.4.0
Spec-Schema: 0.4.0
Author:      Matthias G. Eckermann <pcd@mailbox.org>
License:     GPL-2.0-only
Verification: none
Safety-Level: QM
Includes:     ../../shared/spec/lint-rules.md
```

The included spec is itself a complete PCD spec file with META, TYPES,
BEHAVIORs, INVARIANTS, EXAMPLES. It may declare itself as Deployment:
none, indicating it is a composition target rather than a component
that produces an implementation directly. (Hosts cannot have Deployment:
none unless they exist solely as targets for inclusion.)

## 4. Merge Semantics

When pcd-lint, the translator prompt, or any other consumer processes a
host spec with `Includes:` directives, the following merge is performed
before any other processing.

### 4.1 What merges

- **TYPES**: all type definitions from included specs are added to the
  host's TYPES section. Order: included specs in the order their
  `Includes:` lines appear, then the host's own TYPES.
- **BEHAVIORs**: all `## BEHAVIOR:` and `## BEHAVIOR/INTERNAL:` sections
  merge into the host's set of BEHAVIORs.
- **INVARIANTS**: all INVARIANTS entries merge into a single combined
  list.
- **EXAMPLES**: all EXAMPLES merge. Each EXAMPLE keeps its name (which
  must be unique across the merged set).
- **INTERFACES**: if present, merge.
- **DEPENDENCIES**: if present, merge.
- **TOOLCHAIN-CONSTRAINTS**: if present, merge.
- **PRECONDITIONS** and **POSTCONDITIONS**: merge as combined lists.

### 4.2 What does not merge

- **META fields**: the host's META is authoritative. The included spec's
  META is recorded as *provenance* in the audit trail but does not
  affect translation. Specifically:
  - The host's `Deployment:`, `Verification:`, `Safety-Level:`,
    `Version:`, `Spec-Schema:`, and `License:` apply to the merged
    result. The included spec's values are recorded in the
    TRANSLATION_REPORT but not used.
  - The included spec's `Author:` is preserved in the audit trail
    (multiple Author lines from the included specs append to the host's).
- **MILESTONE sections**: included specs do not contribute milestones.
  Milestones are an orchestration concern of the host component. If an
  included spec contains MILESTONE sections, this is an error
  (RULE-21).
- **DEPLOYMENT section** (the prose section, not the META field): host
  only. Included specs may not contain a `## DEPLOYMENT` section, or it
  is ignored.

### 4.3 Collision handling

The following are spec-author errors, caught by pcd-lint (see RULE-20):

- Two TYPE definitions with the same name from different specs in the
  merged set.
- Two BEHAVIORs with the same name.
- Two INTERFACES with the same name.
- Two EXAMPLEs with the same name.

There is no implicit "host wins" rule. Spec authors must resolve
collisions explicitly — either by renaming in the host, by removing the
included content, or by structuring the included spec so collisions
cannot occur.

The motivation for strict collision handling: implicit precedence rules
are the source of audit-trail ambiguity. If `RULE-01` is defined in both
`lint-rules.md` and `pcd-lint.md`, and pcd-lint silently uses one,
a reviewer cannot tell from the host spec alone what's actually
implemented. Forcing an explicit resolution makes the audit trail
unambiguous.

### 4.4 Provenance

The merged spec retains the origin of every contribution. The
TRANSLATION_REPORT must include an `Included-Specs:` section listing
each included spec with its SHA256:

```
## Included Specs

| Path                                | SHA256                                                            |
|-------------------------------------|-------------------------------------------------------------------|
| ../../shared/spec/lint-rules.md     | a1b2c3d4...                                                       |
```

The hash of the included spec is the SHA256 of its file contents as
read. The host spec's hash is computed *after* merge (see §5).

## 5. Hash Semantics

### 5.1 Merged-text hash

The `Spec-SHA256` embedded in generated artefacts is the SHA256 of the
**merged spec text** — the spec as the translator effectively saw it,
not the host spec file as it sits on disk.

This means: editing `lint-rules.md` invalidates the merged-text hash
of every host that includes it, which in turn invalidates downstream
artefacts (binaries, RPMs, container images) just as editing the host
would. Stale artefacts are detected through the same spec-hash
discipline already in place; no new mechanism is required.

### 5.2 Computation

The merge is deterministic. Given a host spec and its included specs,
the merge produces a single canonical text by:

1. Reading host spec and recording its META.
2. For each `Includes:` directive in order: reading the included spec,
   recursively resolving its own `Includes:`, and computing the
   transitive contribution to TYPES, BEHAVIORs, INVARIANTS, EXAMPLES,
   etc.
3. Writing a canonical merged-spec text with all sections in their
   defined order, all included content inlined in the order specified
   above (§4.1), and the host's META preserved.
4. Computing SHA256 of that canonical text.

Implementations of the merge (pcd-lint, mcp-server-pcd, future tooling)
must produce byte-identical merged text. The canonical form is
specified in the technical reference.

### 5.3 Audit trail

A consumer reading a generated artefact's `Spec-SHA256` cannot
distinguish a single-spec translation from a composed translation. This
is by design: the merged spec is what was translated, and the hash
identifies it.

However, the TRANSLATION_REPORT preserves the structure:

```
## Spec-SHA256 (merged)
4e5f6a7b...

## Spec-SHA256 (host, pre-merge)
9c8d7e6f...

## Included Specs
| Path                                | SHA256       |
|-------------------------------------|--------------|
| ../../shared/spec/lint-rules.md     | a1b2c3d4...  |
```

This gives a reviewer both views: the merged hash for verification of
the artefact, and the component hashes for traceability of what
contributed.

## 6. Lint Rules

Three new rules in pcd-lint (numbered after RULE-18):

### RULE-19: Includes path resolves

For every `Includes:` directive in a spec being linted, the referenced
file must exist and be readable.

- Severity: Error
- Diagnostic: "Includes path does not resolve: {path}"

### RULE-20: Merged spec has no name collisions

After merging all included specs, the resulting set of TYPES, BEHAVIORs,
INTERFACES, and EXAMPLES must have no duplicate names.

- Severity: Error
- Diagnostic: "Name collision after merge: {kind} {name} appears in both
  {origin-1} and {origin-2}"

### RULE-21: Inclusion graph is acyclic and well-formed

- The inclusion graph must contain no cycles. A → B → C → A is an error.
- An included spec may not declare a MILESTONE section.
- An included spec may not declare a `## DEPLOYMENT` section.

- Severity: Error
- Diagnostic:
  - "Inclusion cycle: {spec-A} → {spec-B} → ... → {spec-A}"
  - "Included spec must not declare MILESTONE: {path}"
  - "Included spec must not declare DEPLOYMENT section: {path}"

Inclusion depth is not capped. Practical depth in real specs is expected
to be 1 or 2.

## 7. Translator Behaviour

The translator prompt is updated:

> When reading the host spec, first resolve all `Includes:` directives
> recursively. Compute the canonical merged-spec text per §5.2 of the
> spec composition design. Compute its SHA256; this is the
> `Spec-SHA256` to embed in all generated artefacts. Treat the merged
> spec as the input for all subsequent translation phases.

A translator that does not implement the merge (because it predates
v0.4.0, or because it is operating outside the PCD prompt) will fail
loudly: the host spec's META declares `Spec-Schema: 0.4.0`, and an
older translator will not recognise the `Includes:` field. Forward
compatibility is intentional — a pre-v0.4.0 translator should refuse,
not silently ignore.

## 8. Worked Example: pcd-lint and mcp-server-pcd

### 8.1 The shared spec

A new file at `tools/shared/spec/lint-rules.md`:

```
# PCD Lint Rules (Shared)

## META
Deployment:   none
Version:      0.4.0
Spec-Schema:  0.4.0
Author:       Matthias G. Eckermann <pcd@mailbox.org>
License:      CC-BY-4.0
Verification: none
Safety-Level: QM

## TYPES
[type definitions for diagnostic severity, rule identifier, ...]

## BEHAVIOR: lint-validation-rules
[the orchestration of all 18 rules — moved verbatim from pcd-lint.md]

### RULE-01: Required sections present
[content moved verbatim from pcd-lint.md]

### RULE-02: META fields present and non-empty
[...]

[... all 18 rules ...]

## INVARIANTS
[invariants applicable to rule processing]

## EXAMPLES
[positive and negative examples for every rule]
```

### 8.2 pcd-lint after composition

`tools/pcd-lint/spec/pcd-lint.md` becomes shorter:

```
# pcd-lint

## META
Deployment:   cli-tool
Version:      0.4.0
Spec-Schema:  0.4.0
Author:       Matthias G. Eckermann <pcd@mailbox.org>
License:      GPL-2.0-only
Verification: none
Safety-Level: QM
Includes:     ../../shared/spec/lint-rules.md

## TYPES
[only pcd-lint-specific types: CLI arguments, configuration]

## BEHAVIOR: lint
[the top-level CLI behaviour: parse arguments, read file, invoke
 lint-validation-rules (from shared spec), format output, exit]

## BEHAVIOR: list-templates
[unchanged from current pcd-lint.md]

## BEHAVIOR/INTERNAL: code-fence-tracking
[unchanged from current pcd-lint.md]

## PRECONDITIONS
## POSTCONDITIONS
## INVARIANTS
## EXAMPLES
[pcd-lint-specific examples; rule-specific examples come from shared]

## DEPLOYMENT
[unchanged]
```

### 8.3 mcp-server-pcd after composition

`tools/mcp-server-pcd/spec/mcp-server-pcd.md` adds an `Includes:`:

```
## META
Deployment:   mcp-server
Version:      0.4.0
Spec-Schema:  0.4.0
[...]
Includes:     ../../shared/spec/lint-rules.md
```

The host's BEHAVIORs change: instead of re-describing rule validation,
the `lint_content` and `lint_file` MCP tools delegate to the
`lint-validation-rules` BEHAVIOR from the shared spec, which is now part
of the merged spec.

### 8.4 What this produces

After translation:

- `tools/pcd-lint/code/internal/lint/lint.go` — Go implementation of
  all 18 rules, generated from the merged spec.
- `tools/mcp-server-pcd/code/internal/lint/lint.go` — Go implementation
  of all 18 rules, generated from the merged spec.

Both files contain rule implementations of identical semantics, because
both were generated from the same merged-spec text. The files differ in
their package names, imports, and any host-specific surrounding code,
but the *rule logic* is the same.

When `lint-rules.md` is edited — say to add RULE-19 (an entirely
different RULE-19 from this design document's RULE-19, in a future
revision) — both tools' merged-spec hashes change, both regenerate, and
both produce the updated implementation. No drift is possible.

### 8.5 What about Rust?

If `pcd-lint` were re-spec'd to Rust by changing its deployment template
from `cli-tool` to a Rust-targeting variant, the same `lint-rules.md`
would be consumed and translated to Rust. The shared spec's BEHAVIORs
remain language-neutral; the Rust translator projects them into Rust
idioms (`Result<_, Diagnostic>` instead of `(Output, error)`, `match`
instead of `switch`, and so on).

The mcp-server-pcd, still in Go, would continue to consume the same
shared spec. The two host components are now in different languages but
share their rule definitions. Updates to `lint-rules.md` regenerate both
without modification.

## 9. Migration

The transition from v0.3.x to v0.4.0 is opt-in. Existing specs with no
`Includes:` directives are valid v0.4.0 specs and require no changes.
They behave exactly as they did in v0.3.x.

The migration of `pcd-lint` and `mcp-server-pcd` to use the shared
`lint-rules.md` is a separate piece of work, planned as the first real
use of the mechanism. It should land after the v0.4.0 framework changes
are stable.

## 10. Out of Scope

The following are explicitly not addressed by this design and may be
added in future schema revisions:

- **Conditional inclusion.** No mechanism to include a spec only when a
  preset or template condition is met. If two host components need
  conditional access to shared content, they should structure the
  shared content into smaller specs and include selectively.
- **Versioned inclusion.** No mechanism to declare "include version 1.2
  of lint-rules.md." The included spec is whatever lives at the
  referenced path. Versioning is handled through git history and the
  inclusion graph's recorded hashes.
- **Cross-repository inclusion.** Paths are filesystem-relative. There
  is no `Includes: github.com/other-org/spec.md` or similar. Specs
  consumed across repositories must be vendored.
- **Override or extension.** No mechanism for a host spec to override a
  BEHAVIOR from an included spec, or to extend a TYPE with additional
  fields. If the host needs different behaviour, it should not include
  the spec.
- **Inclusion of partial sections.** No `Includes: lint-rules.md
  sections=BEHAVIORs` syntax. Whole specs only.

These boundaries are intentional. Spec composition is a sharp tool and
its sharpness comes from being predictable; the more expressive the
inclusion mechanism, the harder the merge becomes to reason about and
the more audit-trail ambiguity creeps in.
