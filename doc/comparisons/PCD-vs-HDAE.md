# PCD vs Human-Directed Agentic Engineering

Version: 2026.11.03.01
Date:    2026-11-03
Author:  Matthias G. Eckermann <pcd@mailbox.org>
Status:  Draft, intended for public use
Scope:   doc/comparisons/

---

## 1. Purpose

This is the third entry in `doc/comparisons/`. It positions PCD against
Human-Directed Agentic Engineering (HDAE), the method Gregory M. Kurtzer used
to build the Ginux kernel with a coding agent, documented in his technical
report of October 2026 and since packaged as agent skills.

The two methods state the division of labor between human and agent in nearly
the same words. They differ on what the durable product is once a unit of
work has passed: HDAE keeps the code and governs the loop that produces it;
PCD keeps the specification and regenerates the code from it. That difference
is the subject of this document. It is a difference of object, not of
philosophy, and the two methods compose.

The entry exists because the question "is PCD the same thing as HDAE?" will
be asked. HDAE arrived in late October 2026 with a case study larger than
anything in this repository and an author whose reach exceeds that of the
other entries in this directory.

Sources are listed in section 8. Claims that rest on a single repository read
are flagged in section 7 as "to verify" before any public use.

---

## 2. What HDAE is

HDAE is a method for building and maintaining software with a coding agent.
The human owns product intent, domain decisions, architecture, constraints,
priorities and review boundaries. The agent turns that direction into
artifacts and code, runs implementation and diagnosis loops inside an approved
boundary, and presents evidence at agreed stopping points. The skill's own
framing excludes both approving every individual edit and unguided vibe
coding.

The method was first used, then formalized. Kurtzer's report describes how
Ginux was built: over eleven calendar days, one persistent OpenAI Codex task
using GPT-5.6 Sol produced a Rust-first x86-64 research kernel of
approximately 139,000 retained lines of implementation and test code. The
evaluated system boots directly through UEFI, supplies its own memory
management, scheduler, process model, filesystems, networking and devices,
runs selected unmodified Alpine programs, and completes a Rocky Linux 10
systemd boot under QEMU. The report pins the evaluated commit (dated
2026-10-23), gives an author-reported API cost of approximately USD 8,000,
and states the limits: a research prototype, QEMU only, partial Linux
compatibility, no independent security audit, and an aggregate `make test`
that did not pass completely at the evaluated checkpoint.

The `hdae` repository then packaged the process as portable agent skills.
Its README states that the Ginux process inspired the formalization.

| Item | Value |
|---|---|
| Origin | Process used on Ginux, formalized afterwards into skills |
| Author | Gregory M. Kurtzer (creator of CentOS and Rocky Linux) |
| License | Apache-2.0 (skills and Ginux); the report under its Zenodo record license |
| Language and runtime | POSIX shell installer; a SHA-256 tool required; Python explicitly not required |
| Distribution | `install.sh` fetched from the GitHub `main` branch and piped to `sh`; the same command updates |
| Repository artifacts | `.hdae/skills/`, `.hdae/manifest`, discovery links in `.agents/skills/` and `.claude/skills/`, a managed block in `AGENTS.md`, the `hdae/` control plane |
| Agent coverage | Codex and Claude Code, from one skill source |
| Governance | Discovery interview, gate alignment, checkpoint map, batch delegation rules, completion vocabulary, locked closed gates |
| Additional surfaces | `pr-issue-review` skill for scheduled, read-only issue and pull-request review with an explicit authorization policy |
| First publication | Technical report, October 2026; skills repository at version 0.1.1 when read on 2026-11-03 |

Three constructs account for most of the design.

The **gate** is the unit of work. A gate is one bounded unit of authorized
work with an observable goal, approved scope, non-goals and acceptance
criteria. Its verification plan maps every criterion to an automated test or
command and its expected result. It also records evidence, decisions and
limitations, a delivery strategy (branch, pull request, planned commits) and
a review policy. Each approved gate is one Markdown file under `hdae/`. The
directory is lifecycle state (`gates-todo/`, `gates-done/`,
`gates-deferred/`), the leading number in the filename is execution order,
and a decimal position inserts a gate between two existing ones. A file in
`gates-done/` is a locked historical record: it is never edited, renamed or
deleted, and later work gets a new gate. Not every request needs a gate;
small bounded work is planned, done and verified without one.

The **two loops** are the process. The outer loop is the human's: intent,
then architecture and constraints, then a gate batch with a checkpoint map,
then observed evidence, then direction at an agreed review boundary. The inner
loop is the agent's: implement, build and run, observe, diagnose, revise,
verify. The agent repeats the inner loop without approval of individual edits
while it stays inside the approved gate, and it returns to the outer loop at a
checkpoint, when a human decision is required, or when a stop condition is
met. Stall control is explicit: a retry must change a hypothesis, input,
implementation, environment or diagnostic, and repeating an unchanged attempt
is forbidden.

The **control plane** is what survives the session. `PROJECT.md`,
`ARCHITECTURE.md`, `PLAN.md` and `STATUS.md` under `hdae/` record intent,
boundaries, strategy and the resumable checkpoint. A context-recovery protocol
requires the agent, at task start and after any compaction or handoff, to
reload the complete skill, the repository instructions, the required
references and the checkpoint before resuming; the installer adds a short
managed block to `AGENTS.md` for that purpose. `STATUS.md` must answer which
outcome and invariants govern the work, which gate is active, what evidence
ran, what remains unverified, and which decision or next action is required.

Two smaller rules follow. Verification status has a fixed vocabulary:
passed, accepted, closed, failed, blocked, not run, partial, each defined, and
neither time spent nor code volume completes a gate. Discovered defects
outside the approved gate are not fixed silently; they are recorded and
proposed as a new corrective gate.

The report states what the method leaves behind, and this comparison is
built on that statement: for an HDAE project, what remains durable is the
generated code and, beside it, the specifications, the executable gates and
the decision record, since those three are what later changes and later
agent sessions are bound by.

After publication, the Ginux repository used the method on itself: externally
reported defects were reproduced against a baseline tag and the confirmed ones
worked through numbered gates. The ledger stood at 36 gates on 2026-11-03.

---

## 3. What each method makes durable

### 3.1 The agreement

The division of labor is the same in both methods, nearly word for word.
HDAE: humans direct intent, architecture, constraints, priorities and
acceptance; agents execute bounded implementation and verification loops
against explicit gates. PCD: human intent, machine implementation; domain
experts write the specification, the translator produces the code, and
engineers never write implementation code. Both reject unguided vibe coding
by name, both require evidence before anything counts as done, and both keep
a named human accountable for the result.

The agreement is independent. Kurtzer's report uses the HDAE label as a
descriptive name for the two-loop process it evaluates and makes no claim to
have originated the idea; PCD's first draft dates from January 2026; neither
cites the other.

### 3.2 Both keep structured control artifacts in the repository

A comparison that calls one method documented and the other improvised does
not survive a read of the files. HDAE's control plane has fixed files with
fixed responsibilities, and its gate file has a fixed section list: Source,
Goal, Approved scope, Non-goals, Completed work, Remaining work, Acceptance
criteria, Verification plan, Evidence, Decisions and limitations, Delivery
strategy, Review policy and outcome. The Ginux repository goes further than
the method requires. It keeps versioned interface contracts under `specs/`:
boot handoff versions 0 to 4, each marked superseded by the next, and
syscall, capability, IPC and device contracts. Its architecture document
lists specification-led development, with machine-checkable contracts and
independent validation, as a design principle.

PCD specifications are structured Markdown with mandatory sections: META,
TYPES, INTERFACES, BEHAVIOR blocks with ordered STEPS, PRECONDITIONS,
POSTCONDITIONS, INVARIANTS, EXAMPLES, DEPENDENCIES and DEPLOYMENT, validated
by a linter against a published rule set, with hints, templates and prompts
beside them.

Both are plain text, versioned, diffable and reviewed. The difference is
elsewhere.

### 3.3 The durable set

What separates the two methods is what happens to the code once a gate has
passed.

Under HDAE the code stays. It is the product; the brief, the architecture,
the gates and the decision records constrain how it changes. The maintenance
operation handles bounded defects directly in the code, larger ones through a
new gate, and the revise step of the inner loop edits code. The versioned
contracts in the Ginux repository describe interfaces the code implements;
nothing in the method regenerates code from them, and a contract is marked
superseded when the interface changes.

Under PCD the code is excluded from the durable set. The specification is the
sole human-authored artifact, the code is a build output regenerated from it,
and a correction edits the specification, the hints or the template, never the
code. The reproducibility unit is the tuple of specification, resolved
language, hints and template set, and the specification hash is embedded in
every generated artifact.

| | HDAE | PCD |
|---|---|---|
| Durable product | Code, plus architecture, gates and decision history | Specification, plus hints, templates and prompts |
| The control artifact describes | Intent, boundaries, authorized work and acceptance | The component: types, interfaces, ordered steps, conditions, invariants, examples, deployment |
| Role of the code | The product the gates protect | A derivative the specification produces |
| Reader of the control artifact | A human directing and an agent resuming, both with the codebase in front of them | A translator with no codebase to read |
| Unit of human review | Gate evidence at a checkpoint | The specification and the reports |
| A correction edits | Code, under a new gate | Specification, hints or template, then regenerate |
| Finished when | Every criterion ran and passed, and the human accepted the evidence | A machine that has never seen the code can produce the binary from the specification |

Stated as one sentence: **HDAE governs the loop; PCD governs the artifact.**
In HDAE the gates protect the code. In PCD the specification produces it.

That single difference explains the rest of this document. A method that
keeps the code as the product records how the code came to be, not what
would reproduce it (4.2). It can verify with the same agent that wrote the
code, because the human reviews observed behavior rather than a derivation
(4.3). It leaves the implementation language to the architect, because the
language is a property of the product (4.4). And it presents evidence to a
certifier in the form the certifier already knows: a codebase and the
controls under which it was produced (4.6).

It also explains the reverse. An HDAE gate needs an observable goal and
executable criteria, not a complete description of the component. That lower
bar on the description is why the method could produce a kernel of 139,000
lines in eleven days, and why PCD's own target is a component whose behavior
can be written down in full.

### 3.4 What the two designs share

The durable set is one difference against a wide base of agreement. The
following moves appear in both methods, arrived at independently.

| Shared move | In HDAE | In PCD |
|---|---|---|
| Intent is written down and approved before any code exists | Discovery interview, confirmed brief, architecture, gate map approved before implementation | Specification authored and linted before a translator is invoked |
| The durable artifact is plain text in the repository | `hdae/` control plane, one Markdown file per gate | Specification, hints, templates and prompts, all diffable and reviewable |
| Explicit project memory against session amnesia | `STATUS.md` checkpoint, context-recovery protocol, managed block in `AGENTS.md` | Decisions and milestone hints files, deployment template, translator prompt |
| Large work is decomposed into bounded units | Numbered gates with decimal insertion, batches with a checkpoint map | MILESTONE blocks, per-behavior slicing, composition through `Includes:` |
| Done is defined in advance as executable criteria | Verification plan: every criterion mapped to a command and an expected result | EXAMPLES as an executable oracle, a compile gate, an independent test run |
| History is append-only | Closed gates are locked; a correction is a new gate | Changelog rows are append-only; a correction row names the prior row's error |
| Verification status has a fixed vocabulary | passed, accepted, closed, failed, blocked, not run, partial | Confidence levels per claim in the translation report; unverified claims listed explicitly |
| Interview-first authoring | Discovery interview in small rounds, repository inspected between rounds, summary confirmed by the human | Interview prompt that conducts a structured interview and produces a complete specification |
| Tooling without an interpreter runtime | POSIX shell installer, a SHA-256 tool, no Python | Static Go binaries with assets embedded at build time |
| More than one agent front end from one source | Codex and Claude Code | Any MCP host through `mcp-server-pcd`; a translation skill |
| A human stays accountable; autonomy is bounded | Checkpoints; delegation only for objective, reversible work; never a push to the default branch | A human certifies the specification; the reviewer step is a required gate |
| The framework is built with itself | The `hdae` repository runs its own gate ledger | The PCD tools are generated from their own specifications |

Three qualifications, so the table is not read as more than it is.

Project memory is the same instinct pointed at different targets. A
`STATUS.md` records where the loop is so an agent can resume; a decisions
hints file records realization choices so a translator does not re-derive
them. Neither file would work in the other system unchanged.

The author of an implementation may not be its reviewer is absent from the
table. HDAE has no second-model rule: the agent that wrote the code runs the
verification commands, and the human reviews the evidence. The Ginux report
names the resulting evaluator bias, the author having set the architecture
and judged the behavior, as a threat to validity. PCD and Spec Kitty both
have the rule.

Fix it in the artifact rather than in the code is absent for the opposite
reason. HDAE routes a discovered defect into a new gate and then fixes it in
the code. PCD routes it into the specification. That is 3.3.

---

## 4. Where PCD differs

The differences all descend from the durable set in 3.3. They are listed in
the order a skeptical reader will raise them.

### 4.1 The audited object

Under HDAE the code is the object a reviewer, an auditor or a certifier
examines. The gate ledger is the evidence about how it was produced: for each
unit of work, the authorized scope, the criteria, the commands that ran, their
results and the human decision. The human at a checkpoint reviews that
evidence and the observed behavior. The method does not ask the human to read
every line, and the Ginux report is explicit that the author did not type the
implementation. Responsibility for the claims, the release decision and the
resulting code is placed with the named human author, and the agent is not
listed as an author.

Under PCD the specification is the object under examination and the binary is
its derivative. The specification changes when behavior changes and not
otherwise, which makes the reviewed object stable in a way a codebase is not.
PCD's assurance argument rests on this. On authorship the two differ by
convention rather than by method: PCD credits a model's substantive
contribution in the commit, HDAE keeps the human as sole author.

### 4.2 Provenance of decisions versus provenance of derivation

HDAE records provenance of decisions. The Ginux repository keeps a curated
steering chronology that classifies each design outcome as human-specified,
agent-proposed, collaboratively refined, test-driven correction or
implementation detail, with commits as evidence. The report pins the evaluated
commit, archives the surviving benchmark logs with checksums, and states what
it could not preserve: exact host and tool versions, token use, and a verbatim
conversation export. Its own conclusion is that the checked-in material
permits inspection and rerunning but cannot reconstruct the original
trajectory, and its threats-to-validity section lists the absence of an
independent reconstruction from the original requirements.

PCD records provenance of derivation. The translation report lists a labeled
hash of every file consumed as a translation input. The specification hash is
embedded in source headers, in the binary's version output, in package
metadata and in container labels, so the question "was this artifact produced
from the certified specification?" is answered by recomputing a hash rather
than by asking a person. Reconstruction from the specification is the point
of the pipeline, and the input tuple exists so that a second reconstruction
can be compared with the first.

This is a difference of purpose, not a defect. A project that keeps its code
needs to know why the code is the way it is; a project that regenerates its
code needs to know what it was regenerated from. The report lists the absence
of an independent reconstruction as a limitation; producing that
reconstruction is what PCD's input tuple is for.

### 4.3 Verification: one agent with gates, or separated roles

HDAE verifies with the agent that implemented. The gates are executable and
independent of the agent's opinion (a boot reaches a target, a test passes,
a command exits zero). The test harness that runs them was built in the same
project, largely by the same agent. The human at the checkpoint judges
semantics, usability and system-wide effects from the evidence. The Ginux
report notes that unmodified Alpine and Rocky Linux userspaces exposed
failures that the synthetic tests had missed, which is the method's own
answer to the question of who checks the checker.

PCD separates the roles across models. A test author produces an independent
test suite from the EXAMPLES before the translator writes anything. A reviewer
produces findings that must be resolved in the specification, the hints, the
template or the prompt, never in the code. A tiebreaker, invoked by a human
for ambiguous failures, is forbidden from declaring a winning implementation.
The placement follows from 3.3: a reviewer that is permitted to patch code has
repaired the symptom and left the specification wrong, and under PCD the
specification is the product.

### 4.4 Language neutrality

HDAE leaves the implementation language to the human as an architectural
decision. Ginux is Rust-first by design principle, with assembly and `unsafe`
restricted to small reviewed boundaries, and that choice is recorded in the
architecture document like any other constraint. Nothing in the method asks
for language neutrality, and nothing needs it, because the code is the
product and the product has a language.

PCD keeps language-specific realization out of the behavioral specification
and places it in hints and deployment templates, so that one specification
produces implementations in several languages without specification edits.
This only has meaning if the specification, rather than the code, is the
durable artifact.

### 4.5 Supply chain and sovereignty

HDAE's skills are installed by fetching `install.sh` from the repository's
`main` branch and piping it to a shell; running the same command later
updates from `main`. There is no release pin and no signature. The installer
records a SHA-256 digest of the installed skills in `.hdae/manifest`, which
lets the next run detect that the skills were modified locally and refuse to
overwrite them; it does not establish where the skills came from. The update
check in the skill clones `main` into a temporary directory and runs
`diff -qr`. The case study used a hosted agent; the skills are written for
Codex and Claude Code, both hosted services, and nothing in the method itself
depends on a particular provider.

None of this is unusual for a skill package. The product built with the
method is more careful than the method's own distribution: Ginux pins its
Rocky image by URL and digest, verifies OCI layers before assembling a root,
and its architecture separates firmware trust from administrator
configuration trust.

PCD's target is a pipeline an organization can run itself: tools shipped as
signed packages from a controlled build service, assets embedded at build
time rather than fetched at run time, offline and reproducible builds, and an
SPDX software bill of materials. Models can be hosted in a chosen
jurisdiction or air-gapped entirely. The PCD translation skill refuses to
build, fetch or substitute a missing tool; it stops and reports. For a
customer whose constraint is sovereignty rather than velocity, this is where
the two diverge furthest in practice, and it is a distribution question rather
than a paradigm question.

### 4.6 Certification: the same goal, a different distribution of evidence

Neither approach makes certification impossible, and the comparison is easy
to overstate in PCD's favor. The familiar route keeps the code as the
evaluated object and treats the agent as part of the development environment
under controls. HDAE's gate ledger is evidence in the form that route
requires: authorized scope, criteria, a verification plan, the commands that
ran and their results, the human decision, locked after closure, with a
vocabulary that forbids reporting partial work as passed. An evaluator who
asks what controls governed the agent can be shown the ledger. The cost is
the one already recorded for Spec Kitty: the evaluated object changes with
every gate, and the evidence that a change was understood is a human
acceptance at a checkpoint.

The Ginux case also shows the limit of gates as a security argument. The
report states that no independent security audit was completed and that
security properties beyond the tested mechanisms are outside its evidence.
After publication, externally reported defects, among them in the syscall
entry path and in user-pointer handling, were reproduced against a baseline
tag, classified as confirmed or rejected, and the confirmed ones fixed through
numbered gates. The method handled them in an orderly way; it did not prevent
them, and it does not claim to.

Under PCD the primary human-reviewed object moves to the specification and
the specification-to-artifact link is made machine-checkable. PCD does not
assert that a certifiable system must be built this way; it redistributes the
evidence, trading code-level review effort for provenance and pipeline
evidence, and betting that the trade is accepted. That acceptance has not yet
been demonstrated at certification scale, as the whitepaper records. Both
routes are open.

---

## 5. Where HDAE leads

Stated plainly, so the comparison is not read as one-sided.

- Established projects and maintenance. HDAE has a maintenance operation:
  bounded defects are fixed directly and larger ones converted into gates;
  work originating from an issue or pull request always goes on a dedicated
  branch with a recorded source; a scheduled read-only review skill can act
  under a standing, human-confirmed delegation policy. PCD has a
  reverse-engineering prompt and specification composition, and no loop for
  a codebase that is maintained instead of regenerated.
- Scale of the published case. One case study produced a bootable kernel of
  approximately 139,000 lines in eleven days. PCD's published translation
  runs are single components. The report is careful about what the number
  does not measure, and the size still stands.
- The context-recovery protocol. After every compaction or handoff the agent
  reloads the complete skill, the repository instructions and the checkpoint.
  A managed instruction block triggers the reload, and the checkpoint format
  names the active skill and the exact next action. PCD's translation skill
  has no such procedure. It is a direct candidate for adoption.
- The gate ledger as a filesystem convention. Lifecycle by directory, order by
  filename, insertion by decimal, closed records locked, and the shell
  commands to list and select gates given in the reference. PCD keeps
  milestone status inside the specification, maintained by the pipeline; the
  ledger separates authorization and evidence from the artifact.
- Convergence rules for the inner loop. Never repeat an unchanged attempt,
  prefer the action with the highest information value, and a listed set of
  stop conditions. PCD's templates define halts and a compile gate; they have
  no rule for an agent that keeps trying.
- Completion vocabulary. Seven defined states for a gate. PCD has confidence
  levels per claim in its reports and no state vocabulary for a unit of work.
- Delivery strategy as part of the plan. Branch, intended base, pull request
  and planned commits are recorded in the gate before implementation and
  reconciled at closure.
- Reach. The author created CentOS and Rocky Linux; the report was covered by
  the trade press within days; third-party repositories adopted the HDAE
  wording with attribution within a week. A single-author framework cannot
  manufacture that, and the gap belongs in the comparison rather than outside
  it.

---

## 6. Net positioning

The two methods answer different questions.

- HDAE answers: how does a human keep an autonomous agent inside a boundary,
  get evidence back, and resume after a context loss? It applies to any
  codebase of any size, greenfield or established, and the code is retained.
- PCD answers: what artifact can be certified today and regenerated in ten
  years? It applies to components whose behavior can be written down in full.

Most systems contain both kinds of work: a specifiable core and a large body
of integration and maintenance work that nobody will describe completely.
Nothing forces a team to choose one method for an entire system.

### 6.1 A named integration point

A PCD translation run fits inside an HDAE gate. The gate's sections map onto
artifacts PCD already produces:

| HDAE gate section | PCD source |
|---|---|
| Goal | The MILESTONE or BEHAVIOR under translation |
| Approved scope | The per-behavior bundle produced by `pcd-slice` |
| Non-goals | BEHAVIORs with `Constraint: forbidden` |
| Acceptance criteria | The EXAMPLES section |
| Verification plan | `pcd-lint` clean, the template's compile gate, the independent test run, `verify_spec_hash` |
| Evidence | `TRANSLATION_REPORT.md` and `TEST_REPORT.md`, with their input hashes |
| Decisions and limitations | The decisions hints file; the unverified-claims list in the report |
| Review policy and outcome | Reviewer findings resolved in the inputs; human sign-off on the specification |

One adaptation is required on the HDAE side. Inside such a gate, the revise
step of the inner loop edits the specification, the hints or the template and
regenerates; it does not edit code. HDAE's own rule that a discovered defect
is recorded and proposed as a gate, not fixed on the spot, is the same
instinct; under PCD the corrective gate's scope becomes a specification
change.

The composition runs in the other direction too. HDAE's discovery interview
can front PCD's interview prompt, so that the confirmed brief becomes the
specification and not a separate document. And PCD's input-hash reports are
evidence an HDAE gate can record verbatim.

Both methods ship as agent skills in the same format, so the composition is
a question of which skill governs which step, not an integration project.

### 6.2 Framing for talks

HDAE governs the loop; PCD governs the artifact. HDAE shows what a directed
agent can build in eleven days. PCD asks what of that can still be verified in
ten years, and answers by keeping the specification instead of the code.
Where the component can be described in full, the specification can be signed
and the binary checked against it. Where it cannot, a human-directed loop with
executable gates is the better available method, and saying so costs PCD
nothing.

---

## 7. Open questions and claims to verify

Confirm the following against the skill as installed and a working trial,
rather than against the report and the repository read alone.

- Can a project rule under HDAE forbid code edits inside a gate and route the
  revise step to a specification? The skill handles bounded defects directly
  in code. This decides whether 6.1 is a configuration or a fork of the skill.
- Was the `hdae/` gate ledger in use during the eleven-day Ginux run, or was
  it formalized afterwards? The report names the architecture document, a
  working-goal file, use cases and decision records as the persistent state;
  the first gate in the Ginux ledger is a post-publication baseline. The
  answer affects how much of section 2 describes the case study and how much
  describes the later package.
- What does "one persistent Codex task" mean in terms of context? The report
  says one task; the skill's recovery protocol assumes compaction and
  handoff. The answer affects any claim about context length.
- Does the installer verify origin in any form (signature, pinned release) in
  a later version? Read on 2026-11-03, it does not. Confirm before citing 4.5.
- Is HDAE in use outside the author's own repositories under any regulated
  review, and what is presented in place of code-level review evidence? The
  third-party adoptions found were documentation changes, not evidence of
  audit.
- The line count, the cost figure and the dates are taken from the report and
  the repositories as read on 2026-11-03; the Ginux repository has grown since
  the evaluated commit. Cite the report's figures only, and confirm the Zenodo
  record version at the time of use.

---

## 8. Sources

- Gregory M. Kurtzer, "Ginux: Building an Independent Kernel That Runs Linux
  Userspace through Human-Directed Agentic Engineering", technical report,
  October 2026, DOI 10.5281/zenodo.23682548. Read from the LaTeX source in
  the Ginux repository on 2026-11-03: abstract, Related Work, Methods
  (experimental setting and provenance, staged construction, the closed-loop
  model), Evaluation (artifact, verification and reproducibility), Threats to
  Validity, Discussion.
- github.com/gmkurtzer/hdae, version 0.1.1, read 2026-11-03: README,
  `install.sh`, `skills/hdae-development/SKILL.md` and
  `references/hdae-method.md`, `skills/pr-issue-review/SKILL.md`, the
  repository's own `hdae/` control plane and gate ledger, `.hdae/manifest`.
  The gate file format in 3.2 and the completion vocabulary in section 2 come
  from the method reference.
- github.com/gmkurtzer/ginux, read 2026-11-03: README (research background,
  current demonstrated state, verification targets, licensing),
  `ARCHITECTURE.md` (design principles), `specs/` (versioned contracts),
  `docs/paper/STEERING-PROVENANCE.md`, and the `hdae/` ledger (`STATUS.md`,
  `gates-done/`).
- The author's public announcement of Ginux and HDAE, late October 2026, and
  trade-press coverage of it, used only to confirm dates and the author's own
  framing.
- Pull requests in third-party repositories adopting the HDAE wording with
  attribution to the report, read 2026-11-03; used only for the adoption
  item in section 5.

---

## 9. Follow-up actions (non-normative)

- Adopt a context-recovery protocol in the PCD translation skill: a managed
  instruction block that triggers a reload, and a checkpoint that names the
  active skill, the specification, the bundle and the exact next action.
- Decide whether PCD wants a gate ledger beside the specification, with
  authorization and evidence separated from the artifact and closed records
  locked, or whether the MILESTONE status field and the reports already cover
  it.
- Add a convergence rule to the translator prompt or the templates: a retry
  must change an input or a hypothesis, and a listed set of stop conditions.
- Prototype 6.1 on one component: a PCD translation run inside an HDAE gate
  with the revise step redirected to the inputs, and a record of what the gate
  file adds beyond the translation report.
- Put the first question in section 7 to the project directly.
- Use the one-sentence framing from 3.3 and the closing paragraph of 6.2 in
  the conference talk.

---

## Changelog

- 2026.11.03.01 - Initial draft. Based on the Ginux technical report (LaTeX
  source in the Ginux repository), the `hdae` repository at version 0.1.1,
  the Ginux repository, and the author's public announcement, all read on
  2026-11-03. Written for public use: no third-party or internal project
  names appear. Certification wording follows the Spec Kitty entry: both
  routes are open and distribute the evidence differently. Mirror updates
  applied to the whitepaper (A.12, Comparative Summary, References, row
  0.4.11) and the technical reference (section 18, References).
