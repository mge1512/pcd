# PCD vs Codeplain / Plain

Version: 2026.06.26.01
Date:    2026-06-26
Author:  Matthias G. Eckermann <pcd@mailbox.org>
Status:  Draft for internal review
Scope:   doc/comparisons/

---

## 1. Purpose

This is the first entry in `doc/comparisons/`. It positions PCD against
Codeplain and its open-source specification language Plain, the closest
externally-funded expression of the same core thesis: the specification is
the durable artifact, generated code is disposable output that is
regenerated rather than maintained.

The aim is not to claim PCD is "better" in the abstract. Codeplain and PCD
agree on the paradigm. They optimize different ends of it. This document
records what is shared, where PCD differs, and where Codeplain currently
leads, so the differences can be stated accurately in talks, proposals, and
conversations with skeptics.

Sources are listed in section 8. Claims that rest on a single article or a
single repository read are flagged in section 7 as "to verify" before any
public use.

---

## 2. What Codeplain and Plain are

Codeplain (founded early 2025 in Ljubljana, Slovenia; CEO Dusan Omercevic,
CTO Predrag Radenkovic) launched publicly in September 2025 with the promise
of spec-driven, production-ready code generation. It has raised USD 3M to
date from GapMinder VC and Silicon Gardens. Johan Rosenkilde, a creator of
GitHub Next's SpecLang research project and a founding member of the GitHub
Copilot team, sits on the board and was an early investor.

The stack has three distinct layers, and the distinction matters:

1. Plain - an open-source specification language (MIT, plainlang.org,
   github.com/plainlang), described as combining natural-language
   efficiency with code-like control. A Plain project is organized into
   tiers: functional specs, implementation requirements, test requirements,
   and definitions, plus acceptance tests, linked external resources
   (schemas, API specs), and reusable Liquid templates. Specs compose
   through a module system (import modules and requires modules), and a
   per-spec complexity limit (200 lines) is enforced with tooling that
   detects over-complex specs and conflicts between specs.

2. The Codeplain renderer - the proprietary, hosted component that turns
   Plain specs into code. Per the reporting, the renderer uses faster,
   cheaper models (Gemini Flash) for the code-generation step rather than a
   frontier model, on the argument that codegen is a specialized,
   compiler-like task while the frontier model is reserved for research and
   spec work.

3. plain-forge - a newly released, open-source (MIT) conversational
   spec-writing tool packaged as a Claude Code / OpenCode plugin. It runs a
   four-phase structured interview (what are we building; what technologies
   - language, framework, storage, testing tools; how it works - entities,
   features, flows, business rules; write the specs), confirming each phase
   before proceeding, and emits `.plain` files. It also ships skills for
   per-language test-script generation (`run_unittests`,
   `run_conformance_tests`, `prepare_environment`) and a `debug-specs`
   workflow that traces a bug in generated code back to the spec and fixes
   only the `.plain` files.

The philosophical framing Codeplain leans on is Chad Fowler's "Regenerative
Software" / "Phoenix Architecture" series: code is abundant and cheap,
therefore ephemeral; the spec and the reasoning behind it are what is worth
preserving; manually editing generated code severs the record of intent,
accumulating what Fowler calls "provenance debt". Fowler explicitly invites
many independent implementations of this "missing layer" rather than a
single canonical one.

Codeplain places itself in a lineage: SpecLang (GitHub Next, 2023) ->
Amazon Kiro (2025) -> GitHub Spec Kit, with itself as a production
expression of the same direction. A named customer, Incode (identity
verification), uses it for third-party integrations and regenerates code
from the unchanged spec when an external API change breaks an integration.

---

## 3. Shared premises

PCD and Plain/Codeplain agree at the level of the core thesis, and the
overlap is substantial:

- The specification is the single source of truth; generated code is
  ephemeral and is never hand-maintained.
- A defect is fixed in the spec, not in the code; the code is then
  regenerated. (Codeplain's `debug-specs` skill is explicit about tracing
  the bug back to the spec and editing only the spec.)
- When an external system changes and breaks an integration, the response is
  to regenerate from the same spec - the spec did not break, only the code
  did. This is exactly the METEORA argument for migration and third-party
  interfaces.
- The code-generation step does not require a frontier model. Codeplain
  uses Gemini Flash for rendering; PCD emits declarative spec sections
  deterministically and uses the LLM only for procedural BEHAVIOR bodies.
  Both refuse to spend frontier-model effort on mechanical translation.
- Specs compose from multiple files with an include/module mechanism (Plain
  import/requires modules; PCD host-spec inclusion of shared specs).
- Complexity is governed at the spec level (Plain's 200-line limit and
  conflict detection; PCD's pcd-lint rule set).
- Empirically, developers resist writing specs but are willing to read
  them. This is the same observation PCD makes as "comprehension relocates
  to the spec," now reported from another team's user testing.

The practical takeaway: the regenerate-not-maintain paradigm is no longer a
single-author position. An independently funded company, a board member from
the SpecLang/Copilot lineage, and Fowler's framework arrive at the same
conclusion. In PCD's own terms, this confirms the problem framing, not PCD
specifically.

---

## 4. Where PCD differs

The differences cluster on assurance, sovereignty, and verification - the
properties that matter for compliance-critical and air-gapped targets.

### 4.1 Provenance: cryptographic attestation vs intent preservation

This is the headline difference, and it turns on two different meanings of
"provenance".

- Fowler / Codeplain provenance is about preserving the why: the intent and
  reasoning behind a change, lost when code is hand-edited ("provenance
  debt"; "the conversation is the commit"). This is an intent layer.
- PCD provenance is, in addition, cryptographic attestation. The spec
  SHA256 is embedded in every artifact (source-file headers, the binary's
  version output, RPM and DEB metadata, container labels). The reproducible
  unit - the tuple of spec, resolved language, and hints/template set - is
  recorded by labeled per-file hash in the translation report. The chain is:
  a human certifies the spec -> the spec hash -> artifacts embed the hash ->
  the hash is independently verifiable, with no trust required in any tool,
  pipeline, or person after certification.

The audit question for Common Criteria EAL4+/EUCC is "was this binary
produced from the certified specification?" PCD answers it without human
attestation. Codeplain's provenance concept, as documented, does not. PCD
carries both layers: the spec is the intent record, and the hash is the
attestation.

### 4.2 Verification: adversarial dual-LLM vs single generation path

Both systems generate tests; this is not a "they have no tests" claim. Plain
emits per-language conformance and acceptance tests and detects spec
conflicts. The difference is structural:

- Codeplain, as documented, generates and validates along a single
  translation path (the renderer, validated against its own conformance
  tests and renderer checks).
- PCD treats the translator as untrusted and wraps it: two independent
  translations, two independent test suites authored from the EXAMPLES, a
  cross-validation matrix, a tie-break step, and a reviewer that is
  forbidden from patching code - every proposed fix targets the spec,
  hints, template, or prompt.

The adversarial second LLM is PCD's structural answer to the hallucination
objection. A single generation path, however well validated against its own
tests, does not provide an independent reading of the spec.

### 4.3 Reproducibility: tuple pinning vs stochastic regeneration

- PCD pins the full input tuple and records every input hash, with the goal
  of reproducible builds for audit, and is explicit that the spec hash alone
  does not determine the binary.
- Codeplain regenerates from the same spec using a stochastic model (Gemini
  Flash) with no tuple pinning evident. This is appropriate for the
  high-churn, high-tolerance integration work it targets (the Incode case
  explicitly tolerates things breaking), but it is not a reproducible-build
  posture.

Different assurance targets, not a defect on either side.

### 4.4 Sovereignty: full self-hostable pipeline vs hosted renderer

The open/closed split locates the sovereignty gap precisely:

- Open in Codeplain's stack: the Plain language (MIT) and the plain-forge
  authoring tool (MIT). A team can author Plain specs offline.
- Hosted/proprietary: the renderer that actually turns specs into code,
  which runs on a US-cloud model (Gemini Flash).

So Plain specs can be authored sovereignly, but rendering them to code
depends on a hosted, non-EU pipeline. PCD's entire pipeline is
self-hostable: pcd-lint, the KIT translator harness on EU-jurisdiction or
air-gapped models, and OBS builds with signed RPMs and `GOPROXY=off` for
reproducible offline builds. For digital-sovereignty and air-gapped
deployment, the gap is specifically at the renderer, not at the language.

### 4.5 Spec neutrality: behavior/realization separation

Both parameterize the target language. The difference is where the
realization choices live.

- Plain captures technology choices (language, framework, storage, testing
  tools) and non-functional requirements inside the spec - the plain-forge
  technology phase and the implementation-requirements tier.
- PCD deliberately pushes language-specific realization (filenames, library
  and tool names, command syntax) out of the behavioral spec and into hints
  and templates, so the same behavioral spec retargets to another language
  without spec edits. PCD specs do carry DEPLOYMENT and DELIVERABLES
  sections, so this is a difference of degree and intent, not absolutes.

PCD's stricter separation of normative behavior from realization choice is
what enables the demonstrated three-language-from-one-spec property
(C++, Go, Rust) and clean language ports.

### 4.6 Regeneration policy: three-state model vs pure regeneration

On the exact axis that draws the most skepticism - regenerate vs incremental
in production - Codeplain takes the more absolutist line ("you don't even
tweak the specs, you just regenerate"). PCD is the more nuanced position:

- Clean full regeneration (spec + template only)
- Guided regeneration (spec + template + decisions-hints file)
- Incremental update (spec diff + existing code + decisions-hints file),
  bounded by an enforced Public-API-surface continuity check

PCD therefore already accommodates the common engineering instinct to make
isolated, low-blast-radius changes in production without a full re-roll,
while keeping the spec normative. Codeplain, as reported, does not.

---

## 5. Where Codeplain leads

Stated plainly, so the comparison stays honest:

- Authoring ergonomics. plain-forge's phase-gated conversational interview,
  with per-phase confirmation, directly attacks the "developers will not
  write specs" adoption barrier. PCD has pcd-lint and change-impact
  assessment but no packaged conversational authoring front-end that drafts
  and validates a spec before a human reviews it. This is the clearest gap
  to consider closing (see section 9).
- Adoption insight. The "build relations with the spec" approach -
  incremental, one feature at a time rather than a 200-line upfront dump -
  is a usability principle PCD can adopt for onboarding.
- Ecosystem momentum and external validation. Funding, a board member from
  the SpecLang/Copilot lineage, named customers, and a place in a broader
  movement (SpecLang -> Kiro -> Spec Kit; Fowler's Regenerative Software)
  give the paradigm credibility that a single-author framework cannot
  manufacture.
- Packaged spec-quality tooling. Conflict detection between specs and an
  enforced complexity limit are shipped as first-class skills; PCD has
  comparable pcd-lint capability but can learn from how these are surfaced.

---

## 6. Net positioning

PCD and Plain/Codeplain share one thesis: the spec is the durable artifact,
and code is regenerated rather than maintained. They optimize different ends
of it.

- Codeplain optimizes authoring ergonomics and commercial delivery velocity
  on a hosted renderer, for fast-moving integration work with a high
  tolerance for breakage.
- PCD optimizes verifiable provenance, adversarial dual-LLM verification,
  reproducible builds, and a sovereign / air-gapped supply chain, for
  compliance-critical targets (Common Criteria EAL4+/EUCC, SPDX SBOM).

One-line framing for talks: PCD is a high-assurance instance of the same
regenerative-software paradigm - the end of the spectrum where the binary
must remain a verifiable function of a certified spec, produced by a supply
chain you can run yourself.

---

## 7. Open questions and claims to verify

Before any of section 4 is used publicly, confirm the following against
plainlang.org/docs and a renderer trial, not against the press article
alone:

- Does Codeplain offer an on-premises or self-hosted renderer, or a choice
  of model and jurisdiction? The reporting implies SaaS plus Gemini Flash;
  an enterprise/on-prem option, if it exists, would narrow 4.4.
- Does the renderer or the Plain toolchain offer any artifact hashing,
  attestation, or reproducibility guarantee? None is evident, but absence in
  a press piece is not proof of absence.
- Are Plain's conformance tests independently authored or cross-validated,
  or generated along the same path as the code? The repository suggests a
  single path; confirm before relying on 4.2.
- What is the renderer's actual target-language and framework matrix?
- Is the implementation-requirements tier genuinely in-spec technology
  binding, or separable in a way that preserves retargeting? This affects
  4.5.

---

## 8. Sources

- Paul Sawers, "Code should be regenerated, not maintained: Codeplain makes
  the case for spec-driven development," The New Stack, 2026-06-25.
- plainlang.org - Plain language landing page and documentation index.
- github.com/Codeplain-ai/plain-forge - README and repository structure
  (MIT license), read 2026-06-26.
- Chad Fowler, "Regenerative Software" / "Phoenix Architecture" series,
  aicoding.leaflet.pub (referenced via the article; not independently read
  for this draft).

---

## 9. Follow-up actions (non-normative)

- Consider a conversational, pcd-lint-validated spec-authoring front-end as a
  PCD tooling item, learning from plain-forge's phased interview, but
  keeping the spec - not the conversation - as the durable artifact.
- Complete section 7 verification before citing 4.x publicly.
- Reuse sections 3 and 6 as the backbone of the AGNTCon / MCPCon Europe talk
  comparing PCD with the Plain / Kiro / Spec Kit lineage.

---

## Changelog

- 2026.06.26.01 - Initial draft. Based on The New Stack article (2026-06-25),
  plainlang.org, and the plain-forge repository. Differentiators in section
  4 corrected against the repository read: Plain is language-parameterized
  (not Python-only) and does generate conformance tests; the verification
  difference is restated as single-path vs adversarial dual-LLM rather than
  tests-vs-no-tests. Open questions recorded in section 7.
