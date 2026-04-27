# Post-Coding Development
## Executive Brief

**Author:** Matthias G. Eckermann <pcd@mailbox.org>
**Date:** 2026-03-23
**Status:** Draft

---

## The Problem

Artificial intelligence has made software synthesis cheap. Every major
technology company now offers AI coding assistants, and developer
productivity gains are real. Yet the largest and most lucrative software
markets — automotive, aviation, medical devices, industrial control,
finance, and government — cannot use any of these tools.

The reason is auditability. Regulatory frameworks including ISO 26262
(automotive), DO-178C (aviation), IEC 62304 (medical devices), and
Common Criteria (security certification) require that every line of
deployed software can be traced to a reviewable, auditable source.
AI-generated code cannot meet this requirement. It is produced by a
probabilistic process that cannot explain its own output. Regulators
reject it outright.

This represents a fundamental market failure. The automotive software
market alone exceeds $50 billion annually. The combined regulated
software market — automotive, aviation, medical, defence, finance,
government — is orders of magnitude larger. None of it can today
benefit from AI productivity gains.

---

## The Opportunity

The Post-Coding Development (PCD) solves this problem.

PCD is not AI-assisted coding, where engineers write code and AI
suggests completions. It is a fundamentally different approach: domain
experts write structured natural-language specifications describing
*what* a system should do, and AI generates the complete deliverable
from those specifications — source code, packaging, documentation,
test infrastructure, and the audit record of every translation
decision. The specifications — not the generated code —
are the auditable artifact.

This distinction is decisive. Regulators can review a specification
written in structured English. They can verify that the specification
correctly captures the requirements. They can trace every line of
generated code back to a specific specification clause. The AI is a
translator, not an author. Its output is verifiable because the input
is human-readable and formally structured.

PCD unlocks AI productivity gains across the entire regulated software
market — a market that has been waiting for exactly this capability.

---

## How It Works

A domain expert — an engineer who understands the system's purpose but
need not be a programmer — writes a specification in structured Markdown.
The specification describes data types, behaviours, preconditions,
postconditions, invariants, and concrete examples. No programming
language knowledge is required.

A deployment template resolves the target programming language
automatically. The engineer declares the deployment context
(`cli-tool`, `backend-service`, `verified-library`, and so on) and
the template applies the appropriate language, packaging conventions,
and safety constraints. Language choice is an implementation detail,
not an authoring decision.

An AI translator converts the specification into the full set of
deliverables — source code, packaging artifacts, documentation, test
infrastructure, and a translation report documenting every decision
made during translation. Engineers no longer maintain these artifacts
by hand; the specification is the source of truth, and everything
downstream regenerates from it. For safety-critical components, an
optional formal verification path produces mathematical proofs of key
properties — memory safety, conservation invariants, state machine
correctness.

The output is an audit bundle: specification, generated code, proofs
(where applicable), and translation report. This bundle is the
deliverable for regulatory review. It is complete, traceable, and
human-reviewable at every layer.

---

## Empirical Validation

PCD is not a theoretical proposal. The reference validator — `pcd-lint`,
a command-line tool that checks specification files for structural
correctness — was itself specified and generated using the paradigm.
Zero implementation code was written by hand.

The specification and deployment template were submitted to different AI models across three continents, ranging from frontier
commercial models to 120-billion-parameter open-weight models running
at a regional European provider. Every model independently resolved
the target programming language from the deployment template, without
being told. Every model produced working source code, packaging
artifacts, and a self-assessment report.

The model running on European infrastructure, with no dependency on
US cloud services, produced the most complete set of deliverables of
any tested model — including container images and macOS packaging
alongside the Linux artifacts.

This finding has direct implications for digital sovereignty. Regulated
industries in Germany, France, and the broader EU increasingly require
that sensitive development processes not depend on non-European
infrastructure. PCD works with locally-hosted AI models. The
paradigm's core mechanism — target language resolution from deployment
templates — functions correctly regardless of which AI model performs
the translation.

---

## Strategic Positioning

PCD occupies a position that no existing product addresses.

Current AI coding assistants (GitHub Copilot, Cursor, and their
competitors) are explicitly prohibited in regulated domains. They
produce opaque code that cannot satisfy auditability requirements.
Their market ceiling is the non-regulated software sector.

Traditional formal methods (TLA+, Coq, F*, Dafny) provide the
mathematical rigour that regulators require, but they demand
specialised expertise that most engineering organisations do not
have. Their adoption has been limited to elite research groups and
a handful of high-assurance projects.

PCD bridges this gap. Specifications are written in structured
natural language — accessible to domain experts, not just formal
methods specialists. Formal verification is available as an optional
layer for the highest-assurance components. The same specification
format works for a command-line tool with no certification requirement
and for an automotive safety function requiring ASIL-D certification.

The project is open source under a deliberate dual-licensing strategy.
The specification format and deployment templates are published under
Creative Commons Attribution 4.0, allowing any organisation to
implement the paradigm — including proprietary implementations. The
reference tooling is published under GNU General Public License
version 2 (GPLv2-only), following the Linux kernel model: organisations that
modify and distribute the validator must contribute their changes
back. This mirrors the dynamic that made Linux the dominant platform
for regulated industry embedded software.

---

## What Exists Today

The following artifacts are publicly available in the project repository:

**Specification format.** A stable, versioned Markdown schema with
formal validation tooling. Covers all required sections, SPDX license
validation, and deployment template resolution.

**Deployment templates.** Production-ready template for command-line
tools. Stub templates for safety/security-critical libraries, general
C-ABI libraries, Python tooling, multi-component project manifests,
and MCP server components.

**Reference implementation.** `pcd-lint`, the specification validator,
generated from its own specification. Available as RPM, DEB, and
container image via the openSUSE Build Service.

**Translator prompt.** A versioned, model-agnostic prompt that drives
AI translation of specifications into complete deliverable sets.
Tested across multiple commercial and open-weight models.

**Worked example.** A complete account transfer specification
demonstrating the full paradigm, including formal pre- and
postconditions, state invariants, and executable examples.

---

## Why Now

Three conditions converge in 2026 that make this moment the right
one to establish this paradigm.

First, AI model capability has crossed the threshold where structured
specification-to-code translation is reliable enough for production
use. The empirical evidence from the `pcd-lint` development confirms
this for a representative real-world component.

Second, the EU Cyber Resilience Act and strengthening of Common
Criteria requirements are increasing the cost of traditional manual
certification processes. The demand for an auditable AI development
pathway is accelerating, not receding.

Third, the open-weight model ecosystem has matured to the point where
organisations can run capable AI models on their own infrastructure.
The digital sovereignty argument is no longer theoretical.

The window to establish the open standard — and the toolchain that
enforces it — is open now. Once proprietary vendors recognise the
market opportunity, the standard will fragment unless an open,
community-controlled reference exists first.

---

## Next Steps

The project is seeking collaborators in three areas:

**Regulated-industry pilots.** Early adopters willing to apply PCD
to a real component — a cryptographic primitive, a state machine, a
device driver — and document the certification cost and effort
reduction.

**Standards engagement.** Organisations with existing relationships
with ISO TC22, EUROCAE, IEC TC62, or Common Criteria schemes to
validate the audit bundle format against actual certification
requirements.

**Toolchain contribution.** Engineering teams who can contribute
deployment templates for their domain — additional safety standards,
programming languages, packaging formats, or verification back-ends.

Contact: pcd@mailbox.org

---

*The Post-Coding Development is an independent open-source
project. The author is affiliated with SUSE but this document does not
represent SUSE's official position.*
