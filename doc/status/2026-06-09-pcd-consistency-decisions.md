# PCD Consistency Check - Decisions and Task Dispositions

Version:  2026.06.10.01
Refers-to: pcd-consistency-report.md 2026.06.10.01
Decided-by: Matthias G. Eckermann, 2026-06-10
Executed-by: Claude (Fable 5), same day, per maintainer instruction

---

## 1. Decisions D-1 .. D-10

| ID | Question | Decision |
|---|---|---|
| D-1 | cli-tool language set | KEEP Java and Lean4 as supported |
| D-2 | Linking relaxation | ACCEPT dynamic linking; preference for C/C++, optional for Go and Rust |
| D-3 | Test-injection env var | ACCEPT one spec-declared variable |
| D-4 | PER-HOST-CONFIG key | DEFER |
| D-5 | Doc version policy | Header tracks the framework version at last content revision |
| D-6 | Composition fragment naming | Uniform specification names: ".spec.md" (lint-rules.md to be renamed lint-rules.spec.md when T-20 executes) |
| D-7 | Prompt in the reproducibility tuple | EXTEND the tuple to include the prompt: (spec, resolved language, hints and template set, prompt) |
| D-8 | Version grammar | ACCEPT YYYY.MM.DD.VV alongside semver |
| D-9 | Spec-Schema restamp policy | Restamp after review; the value then means "validated against" |
| D-10 | MILESTONE Status | Pipeline-maintained |

## 2. Task dispositions (maintainer ruling of 2026-06-10)

Executed in this batch:

| Task | Ruling | Amendment |
|---|---|---|
| T-01 | Restore | none |
| T-26 | Agreed | none |
| T-02 | Yes | none |
| T-24 | Yes | "assuming this follows what we have done for kvm-manager and zypper-declarative" - confirmed: the merged-hash definition is exactly what both projects' reports already record (merged = host when no Includes) |
| T-27 | Yes | prompts as generic as possible |
| T-23 | Yes | tuple extended per D-7 |
| T-03 | Yes | AMENDED: for languages with vendoring, the vendored files go into a separate tarball `<name>-<version>-vendor.tar.<extension>` |
| T-05 | Yes | AMENDED: keep the door open for future exceptions (e.g. ephemeral VMs) |
| T-06 | Agreed | AMENDED: Go-specific advice MUST NOT live in the language-neutral templates; it lives in language-specific hints files. Interpretation applied: the per-language compile-gate table row gains only the bare command `go vet ./...` (the template already carries per-language command rows, e.g. the test-author syntax check); every explanation and pitfall lives in cli-tool.go.milestones.hints.md. If the intent was zero template change, reverting is one line. |
| T-07 | Yes | AMENDED: prefer CMake; Meson as an option; automake in exceptional cases |
| T-08 | Yes | PERSISTENCE key expected to be needed more in the future |
| T-09 | Per D-2 | see D-2 |
| T-10 | OK | none |
| T-11 | Per D-3 | see D-3 |
| T-25 | Executed (second pass, same day, per D-1) | Java and Lean4 added to TYPES, PRECONDITIONS, and LANGUAGE-ALTERNATIVES of cli-tool.template.md (0.4.1); per-language tables already carried both |
| T-29 | Executed (second pass, same day, per D-8) | RULE-02c alternation in lint-rules.md (0.4.2) with new SpecVersion type; SemanticVersion stays semver-only for Spec-Schema (RULE-02d unchanged); VERSION rows of all ten templates note both schemes |

Deferred by ruling:

| Task | Ruling |
|---|---|
| T-04 | DEFER (RPM directory-ownership rule stays project-side for now) |
| T-12 | DEFER (PER-HOST-CONFIG, per D-4) |

Decided but deliberately NOT executed in this batch (decision recorded, implementing task not in today's approved list):

| Decision | Pending task |
|---|---|
| D-6 (uniform .spec.md) | T-20 codifies the naming convention and renames lint-rules.md |
| D-9 (restamp policy) | T-30 restamps the lagging templates after review |
| D-10 (pipeline-maintained Status) | T-31 documents the operator step |

All remaining tasks (T-13..T-22, T-28, T-30, T-31) are approached later.

## 3. Framework version designation (proposal, flagged for review)

The T-23 provenance-contract extension (Prompt-SHA256 required line; tuple
extended per D-7; canonical labels Upgrade-Brief-SHA256 and Directive-SHA256)
is a contract change of the same class as v0.4.3. This batch designates it
**v0.4.4**. Applied to: prompts/prompt.md ## Reports (both modes), the
Translation-Inputs provenance block of all ten report-bearing templates, and
doc/technical-reference.md section 12 (header bumped per D-5). The whitepaper
changelog row follows with T-14.

---

## Changelog

- 2026.06.10.02 - T-25 (per D-1) and T-29 (per D-8) executed on maintainer
  instruction in a second same-day pass; dispositions updated. cli-tool
  template at 0.4.1, lint-rules at 0.4.2, the other nine templates
  patch-bumped for the VERSION-row note.
- 2026.06.10.01 - Initial record of D-1..D-10 and the task dispositions of
  2026-06-10, including the four amendments (T-03 vendor tarball, T-05
  exception door, T-06 hints-only advice, T-07 CMake preference) and the
  v0.4.4 designation proposal.
