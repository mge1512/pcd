# PCD vs Spec Kitty

Version: 2026.09.01.02
Date:    2026-09-01
Author:  Matthias G. Eckermann <pcd@mailbox.org>
Status:  Draft, intended for public use
Scope:   doc/comparisons/

---

## 1. Purpose

This is the second entry in `doc/comparisons/`. It positions PCD against
Spec Kitty, an open-source workflow and governance layer for agentic
coding.

The two projects agree on where software development is going: humans express
intent, machines implement it. They disagree about what a specification has to
contain today, and that disagreement is the subject of this document. It is a
difference of completeness and of intended reader, not of philosophy, and the
two formats can compose.

An earlier revision of this entry framed the difference as a disagreement
about which artifact is authoritative. That framing came from the published
article and was too strong; section 3 records what replaced it and why.

Spec Kitty also has more users than the other entries in this directory, so a
reader may already know its arguments.

Sources are listed in section 8. Claims that rest on a single article or a
single repository read are flagged in section 7 as "to verify" before any
public use.

---

## 2. What Spec Kitty is

Spec Kitty is a command-line tool that wraps a governed workflow around AI
coding agents. It turns product intent into a repeatable loop, keeps the
context of that loop in the repository, cuts implementation into work
packages that agents execute, and isolates each package in a git worktree so
parallel work does not collide.

The loop is: charter, specify, plan, tasks, next, implement, review, accept,
merge, retrospective. Each step has a slash command generated for the user's
agent, and a corresponding command in the CLI.

| Item | Value |
|---|---|
| Origin | Fork of GitHub Spec Kit, since heavily diverged |
| License | MIT |
| Language and runtime | Python, requires 3.11 or newer |
| Distribution | Python package index, installed with `pipx`, `uv` or `pip` |
| Repository artifacts | `kitty-specs/`, `.kittify/config.yaml`, `.worktrees/` |
| Agent coverage | Roughly fifteen agent front ends, including Claude Code, Codex, Cursor, Gemini CLI, Copilot, Windsurf, OpenCode, Qwen, Kiro |
| Governance | Per-project charter, doctrine layer (directives, tactics, profiles), architecture decision records, `advise` / `ask` / `do`, trail model |
| Additional surfaces | Local kanban dashboard, mission retrospectives, optional hosted workspace |
| First release | 2025-11-02 (`spec-kitty-cli` 0.2.20) |
| Current release | 3.2.5, 2026-07-08; pre-release 3.2.6rc2, 2026-08-21 |
| Release count | 165 published versions in ten months |

Three constructs account for most of the design.

The **mission** is the unit of work: a purpose-specific workflow with defined
inputs, process, outcomes and states. Release 3.2.5 ships four mission kinds -
software development, research, documentation and planning - each with its own
specification template, task template and step contracts, resolved through a
mission-type repository. Mission kinds are an extension point rather than a
fixed list.

A second extension point sits beside it. Release 3.2.5 ships a **doctrine
pack** implementing the Thoughtworks REASONS Canvas, activatable per project,
with its own activation and charter-context contracts and a template renderer.
A foreign specification format has already been added to Spec Kitty this way
once.

The **charter** is a per-project governance document that records what an
agent cannot infer from the code. The published example includes which of two
container workflows is policy rather than convenience, which continuous
integration failure is expected after a merge and must not be treated as a
defect, which domain noun is current and which is a legacy alias, and which
user-authored files an upgrade must preserve. Every project gets its own
charter rather than inheriting a generic rule set.

**Cross-review with a fresh context** is the verification step. When an
implementation lands, a second agent, or the same agent with its context
stripped, reviews it. The stated reason is that an agent which has just
finished writing code has spent the preceding hour convincing itself the code
is correct, and is therefore the worst available reviewer of it.

The project publishes a security position dated 2026-02-18. It states that
the core is a deterministic workflow tool which does not set approval,
sandbox or network flags on the agent front ends it drives; that autonomous
orchestration is a separate, optional component that policy can disallow;
that workflow state changes must go through the host commands; and that there
is no direct write path from an agent provider to the hosted service. The
document reads as a response to earlier criticism about bundled autonomous
execution.

---

## 3. What each project means by "specification"

### 3.1 The agreement

Both projects agree that the future of software is humans expressing intent
and machines implementing it. Spec Kitty's author states this in his founding
article, before the argument that follows from it, and PCD is built on the
same premise. The agreement is recorded first because the published
disagreement is easy to read as a disagreement about the destination, when it
is a disagreement about what is achievable in an existing codebase now.

His position on that second question is that for brownfield work - millions of
lines, decisions made years ago by people who have moved on, constraints
written down nowhere - intent alone is not yet enough, so the code remains the
authority on what the system currently does and the specification states the
intended change. PCD's target is a component whose behavior can be written
down in full, where that constraint does not bind.

### 3.2 Both formats are structured

A comparison that calls one format formal and the other informal does not
survive contact with the files. Spec Kitty's specification template for
software-development missions, shipped in release 3.2.5, is a structured
requirements document with mandatory sections and typed identifiers:

- User stories with explicit priorities, each required to be independently
  testable and independently deliverable
- Acceptance Scenarios written as GIVEN / WHEN / THEN, the same
  behavior-driven-development lineage from which PCD took its EXAMPLES section
- Functional requirements, non-functional requirements and constraints in
  three separate tables with unique typed identifiers and a status column
- A rule that every non-functional requirement states a measurable threshold
- A constraints table with a regulatory category
- Key entities, described without implementation
- Success criteria with their own identifiers, required by the template to be
  technology-agnostic and measurable

PCD specifications are also structured Markdown with mandatory sections: META,
TYPES, INTERFACES, BEHAVIOR blocks with ordered STEPS, PRECONDITIONS,
POSTCONDITIONS, INVARIANTS, EXAMPLES, DEPENDENCIES and DEPLOYMENT, validated
by a linter against a published rule set.

Both are constrained natural language. The difference lies elsewhere.

### 3.3 The completeness bar

What separates the two formats is how much has to be written down before the
document is finished, and that follows from who reads it.

| | Spec Kitty | PCD |
|---|---|---|
| Describes | What should change, and how you know it worked | What the component is |
| Contains | User stories, requirement and constraint tables, entities, success criteria | Types, interfaces, operations with ordered steps, pre- and postconditions, invariants, examples, deployment |
| Reader | A team, plus an agent that also reads the codebase | A translator with no codebase to read |
| Finished when | The team agrees what done means and the criteria are measurable | A machine that has never seen the code can produce the binary from it |
| Status of the document | The record of a decision | The certified artifact |

Stated as one sentence: **Spec Kitty specifies the requirement, PCD specifies
the component.** A Spec Kitty specification is sufficient for a team to agree
and for an agent to act against code that already exists. A PCD specification
is sufficient on its own, which is the property that makes it signable.

That single difference explains the rest of this document. A requirements
document cannot be the audit point, because it does not describe the
component (4.1). It cannot be hashed into an artifact as a claim about
behavior, because it does not determine behavior (4.2). It has no language to
be neutral about, because the language is a property of the repository it
describes (4.4).

It also explains the reverse: a component specification is expensive, and
buying that completeness is only rational where the component is small enough
to describe and durable enough to repay the cost.

## 4. Where PCD differs

The differences all descend from the completeness bar in 3.3. They are listed
in the order a skeptical reader will raise them.

### 4.1 The audited object

Under Spec Kitty the merged code remains the object a reviewer, an auditor or
a certifier reads. The charter, the review lane and the merge record are
evidence about how that code was produced, and an evaluator can work with
them, but the artifact under examination is still a codebase that changes
with every mission.

Under PCD the specification is the object under examination and the binary is
its derivative. The specification changes when behavior changes and not
otherwise, which makes the reviewed object stable in a way a codebase is not.
PCD's assurance argument rests on this. Neither project chose its position
carelessly: a requirements document cannot serve as the audit point for a
component it does not describe, so Spec Kitty keeps the audit at the code
because that is where the description sits.

### 4.2 Reproducibility and attestation

PCD pins the input tuple - specification, resolved language, hints and
template set - and records a labeled hash of every file consumed as a
translation input. The specification hash is embedded in source headers, in
the binary's version output, in package metadata and in container labels, so
the question "was this binary produced from the certified specification?" is
answered by recomputing a hash rather than by asking a person.

No comparable mechanism is documented in Spec Kitty. Nothing in the published
material records which model, which prompt version and which specification
revision produced a given work package in a form that survives the merge.
Provenance in that system is the git history plus the decision ledger, which
is a record of intent, not an attestation of derivation.

This is a difference of purpose, not a defect. A team shipping product
features does not need a derivation proof.

### 4.3 Verification: the same instinct, placed differently

Both systems refuse to let the author of an implementation be its reviewer,
and both use a second model with a clean context to enforce that. The
argument for it is the same in both projects, and Spec Kitty states it well.

The placement differs. Spec Kitty's reviewer reads a diff and decides whether
to accept it; the reviewer may propose changes to that diff, and the accepted
diff is merged. PCD's second model does not review the code at all. It
authors an independent test suite from the EXAMPLES section of the
specification, before the translator writes anything, and any finding it
raises must be resolved in the specification, the hints, the template or the
translator prompt. A reviewer that is permitted to patch code has repaired
the symptom and left the specification wrong.

### 4.4 Language neutrality

Spec Kitty operates on the repository as it already is. The target language,
frameworks and libraries are properties of the existing codebase, which is
appropriate for its target case and leaves no retargeting question to answer.

PCD keeps language-specific realization out of the behavioral specification
and places it in hints and deployment templates, so that one specification
produces implementations in several languages without specification edits.
This is what makes a specification outlive the language ecosystem it was
first translated into, and it only has meaning if the specification, rather
than the code, is the durable artifact.

### 4.5 Supply chain and sovereignty

Spec Kitty is distributed through a language package index and installed with
a language-ecosystem installer. Its runtime is an interpreter with a version
floor, its agent front ends are largely hosted services, and a hosted
workspace is available as an option. Nothing here is unusual, and the
security position document gives reasonable deployment guidance for
organizations that need to control the agent front ends.

PCD's target is a pipeline an organization can run itself: signed packages
from a controlled build service, offline and reproducible builds, an SPDX
software bill of materials, and models that can be hosted in a chosen
jurisdiction or air-gapped entirely. For a customer whose constraint is
sovereignty rather than velocity, this is where the two diverge furthest in
practice, and it is a distribution and deployment question rather than a
paradigm question.

### 4.6 Certification: the same goal, a different distribution of evidence

Neither approach makes certification impossible, and the comparison is easy
to overstate in PCD's favor. Common Criteria evaluations of hand-written code
are routine and have been for decades. At EAL4 the developer supplies the
design documentation, the development-environment controls and the
implementation representation of the security functionality, and the
evaluator examines a sample of it. Introducing an AI coding agent into that
process does not close the path. It changes what the developer has to show.

Under Spec Kitty's model the code stays the evaluated object and the agent
becomes part of the development environment. The team must show that its
review step, its work-package boundaries and its acceptance gates are
adequate controls over what the agent produced. That arrangement is familiar to evaluators,
and the charter, the review lane and the merge record are usable evidence. The
cost is that the evaluated object keeps moving: every agent-authored change
re-enters code-level review scope, and the evidence that a given change was
understood is a human sign-off rather than a machine-checkable link.

Under PCD the primary human-reviewed object moves to the specification and
the specification-to-artifact link is made machine-checkable. The claim is
narrower than it is sometimes read to be. PCD does not assert that a
certifiable system must be built this way; it redistributes the evidence,
trading code-level review effort for provenance and pipeline evidence, and
betting that the trade is accepted.

The residue is stated honestly in the whitepaper and repeated here: PCD's
route depends on a certifier accepting specification-level comprehension plus
an auditable pipeline in place of code-level comprehension, and that has not
yet been demonstrated at certification scale. Spec Kitty's route is the
proven one and is expensive per change. PCD's route is cheaper per change if
it is accepted, and acceptance is exactly what is still open.

---

## 5. Where Spec Kitty leads

Stated plainly, so the comparison stays honest.

- Brownfield work. Spec Kitty is built for a codebase with twenty years of
  decisions in it, most of them undocumented. PCD assumes a component whose
  behavior can be written down in full, which is a much narrower starting
  condition. Where a component specification cannot be written, Spec Kitty
  has an answer and PCD does not.
- Format plurality. Spec Kitty treats the specification format as a property
  of the mission kind and of the active doctrine pack, so more than one format
  can coexist in one repository. PCD has one format and a linter that enforces
  it, which is the right choice for a certifiable artifact and the wrong one
  for a team that wants to adopt gradually.
- The charter as a named artifact. Project policy that the agent reads, kept
  separate from the specification and from the framework's own rules, is a
  construct PCD lacks under any name. PCD has hints, templates, prompts and
  contribution rules, which cover parts of the same ground without being
  recognized as one thing.
- Parallel execution. Work packages with lifecycle lanes, worktree isolation
  per package and a dashboard across all of them are an orchestration layer
  PCD has never built. The per-behavior slicing PCD does for token cost is
  the same decomposition applied to a different problem, and the two could
  meet.
- Multi-agent reach. Command surfaces generated for roughly fifteen agent
  front ends from one source is the pattern any framework needs if it wants
  adoption beyond its author's own toolchain.
- Interview-first authoring. Spec Kitty interviews the human until the
  boundaries and contracts of the intended change are settled, and presses
  when the answers are vague or contradictory. This is the second independent
  arrival at conversational specification authoring recorded in
  `doc/comparisons/`, which moves it from an idea to a pattern.
- Retrospectives as artifacts. Every completed mission produces one by
  default, and there is machinery to synthesize across them. PCD produces
  findings and change briefs per translation run but has no loop that feeds
  outcomes back into the framework.
- Adoption. The project is roughly ten months old with 165 published
  releases, an active issue tracker, third-party training offerings and
  visible community use. A single-author framework cannot manufacture that,
  and the gap belongs in the comparison rather than outside it.

---

## 6. Net positioning

The two projects sit at different points on one scale, and the scale is how
completely a component can be described.

- Spec Kitty addresses systems that cannot be described in full. The
  achievable win is that agents stop misreading intent and the team stops
  losing its decisions, with human accountability kept at the code. That is
  the conservative position and the currently provable one.
- PCD addresses components that can be described in full and have to stay
  verifiable for a decade. The achievable win is that the binary remains a
  checkable function of a certified specification, produced by a pipeline the
  operator can run.

Most systems contain both kinds of component: a long-lived, specifiable core
and a large periphery of integration code that nobody will ever describe
completely. Nothing forces a team to choose one approach for an entire system.

### 6.1 A named integration point

Spec Kitty's architecture already accepts additional specification formats, in
two ways described in section 2: each mission kind has its own specification
template and step contracts, and doctrine packs can be activated per project,
as the shipped REASONS Canvas pack demonstrates.

That makes the composition concrete rather than aspirational. A mission kind
whose specification template is the PCD format, whose plan step is a milestone
design rather than a task breakdown, and whose accept gate is a linter run
plus a translation producing a passing binary, is the same kind of addition
Spec Kitty has already made once for a foreign format.

The mapping at the front is close to direct. Acceptance Scenarios in
GIVEN / WHEN / THEN are PCD EXAMPLES. The requirement, constraint and success
criteria tables are the rationale and acceptance layer that PCD
specifications lack, and that a reviewer of a PCD specification currently has
to reconstruct. What a PCD mission would add is the back half of
the table in 3.3: types, interfaces, ordered steps, pre- and postconditions,
invariants and deployment.

Whether the mission-kind mechanism admits a third-party validator and a
programmatic accept gate, rather than only a template and prose contracts, is
the question that decides whether this is a weekend or a project. It is listed
in section 7.

### 6.2 Framing for talks

Both projects agree that intent has to be written down and that machines will
do the implementing. They differ on how complete the writing has to be before
the machine can be trusted with it, and the honest answer depends on whether
the component can be described in full. Where it can, a specification can be
signed and the binary checked against it. Where it cannot, a requirement plus
a governed agent is the better available answer, and saying so costs PCD
nothing.

## 7. Open questions and claims to verify

Confirm the following against the documentation, the source distribution and a
working trial rather than against the founding article and the repository page
alone.

- Does the mission-kind mechanism admit a third-party specification format
  with its own validator and a programmatic accept gate, or only a template
  and prose contracts? This decides the size of the integration in 6.1 and is
  the first question to put to the project.
- Does any part of the pipeline record model identity, prompt version and
  specification revision per work package, in a form that survives the merge?
  Absence from the documentation is not proof of absence, and this single
  answer decides 4.2.
- How much of the charter is mechanically enforced? The source distribution
  ships term-neutrality lists, a language-scoped allowlist and an error-code
  catalog beside the charter defaults, so part of it clearly is. The split
  between the enforced part and the advisory part decides how much of 4.6
  stands.
- Are there users under regulated audit, and what do they present in place of
  code-level review evidence?
- Is the hosted workspace available self-hosted, with a choice of model and
  jurisdiction? This affects 4.5 and nothing else.
- Confirm the fork lineage from GitHub Spec Kit from a primary source. Several
  secondary sources agree, which is suggestive rather than conclusive.
- Confirm the release and version figures at the time of use. They were read
  on 2026-09-01 and this project changes weekly.

## 8. Sources

- Robert Douglass, "Why I Built Spec Kitty", Medium, 2026-04-27. The shared
  premise in 3.1 and the brownfield argument come from here.
- Direct conversation between the author of this document and the project's
  founder, 2026-09. The agreement recorded in 3.1 was confirmed there.
- `spec-kitty-cli` 3.2.5 source distribution, read 2026-09-01. Specification
  template facts in 3.2 come from
  `src/doctrine/missions/software-dev/templates/spec-template.md`; the mission
  kinds and the REASONS doctrine pack in section 2 and 6.1 from
  `src/doctrine/missions/` and `src/doctrine/spdd_reasons/`; the charter
  enforcement material in section 7 from `src/charter/`.
- github.com/Priivacy-ai/spec-kitty - README, `SECURITY-POSITION.md`
  (dated 2026-02-18), repository structure. Read 2026-09-01.
- docs.spec-kitty.ai - documentation index and glossary, version 3.2.
  Read 2026-09-01.
- Python package index metadata for `spec-kitty-cli`, retrieved 2026-09-01:
  license, Python floor, release list and upload dates.
- Public profile material for the project's founders, used only for section 2
  attribution.

---

## 9. Follow-up actions (non-normative)

- Put the first question in section 7 to the project directly. The answer
  decides whether 6.1 is a prototype or a design discussion.
- Decide whether PCD wants a named per-project governance artifact in the
  sense of section 5, distinct from hints and templates, or whether the
  existing layers already cover it under different names.
- Consider adopting a rationale and acceptance layer in PCD specifications,
  in the spirit of the requirement and success-criteria tables in 3.2. A
  reviewer signing a PCD specification currently reconstructs the why from
  the what.
- Complete section 7 verification before citing 4.x publicly.
- Use the one-sentence framing from 3.3 and the closing paragraph of 6.2 in
  the conference talk. Both position PCD by scope rather than by claim.
- Write the boundary note the end of section 6 points at: which components in
  a mixed system are specifiable, and what the interface between a specified
  core and an unspecified periphery looks like.

---

## Changelog

- 2026.09.01.02 - Section 3 rewritten. The first revision framed the
  difference as a disagreement about which artifact is authoritative, taken
  from the founding article. A direct conversation with the project's founder
  established that both projects share the same view of where software
  development is going, and a read of the shipped specification template
  established that both formats are structured natural language with
  mandatory sections and typed identifiers. The accurate axis is the
  completeness bar: Spec Kitty specifies the requirement, PCD specifies the
  component. Section 6 rewritten from opposition to complementarity and given
  a named integration point, grounded in the mission-kind and doctrine-pack
  mechanisms the project already ships. Sections 1, 2, 4 intro, 4.1, 5, 7, 8
  and 9 adjusted to match. No claim in 4.2 to 4.6 changed.

- 2026.09.01.01 - Initial draft. Based on the founder's article of 2026-04-27,
  the repository README and security position document, the documentation
  index, and package index metadata, all read on 2026-09-01. Written for
  public use: no third-party or internal project names appear. Section 4.6
  rewritten from an earlier draft that claimed certification was closed to
  hand-maintained AI-generated code; the accurate statement is that both
  routes are open and distribute the evidence differently.
