# PCD vs the Liatrio Spec-Driven Development workflow

Version: 2026.11.05.01
Date:    2026-11-05
Author:  Matthias G. Eckermann <pcd@mailbox.org>
Status:  Draft, intended for public use
Scope:   doc/comparisons/

---

## 1. Purpose

This is the fifth entry in `doc/comparisons/`. It positions PCD against the
Spec-Driven Development workflow that Liatrio publishes in the repository
`liatrio-labs/spec-driven-workflow`. "Spec-driven development" is also the
name of the family of methods, so this entry writes "Liatrio SDD" for this
workflow.

Liatrio SDD is a workflow in four phases for one engineer who works with an
AI coding assistant on one piece of work. The assistant writes a
specification, plans tasks, implements them with evidence, and validates the
result. The workflow is plain Markdown, and every rule of it can be read in
the repository.

The two methods agree on the first step. Intent is written down, questions
are answered and evidence is planned before any code exists. They differ in
what the specification is afterwards. Liatrio SDD says that its specification
guides one implementation run, and that the code and the tests are the
durable result. PCD says that the specification is the durable result, and
that the code is generated from it again when it changes. That difference is
the subject of this entry.

The entry exists because "we already work spec-driven" is a common answer to
PCD, and the term covers both positions. Liatrio SDD states the first
position in plain words, so it is a convenient reference for the
distinction.

Sources are listed in section 8. No trial of the workflow was run for this
entry. Every statement about its behavior comes from the text of the skill at
one commit and from the project's website. Claims to check before any public
use are listed in section 7.

---

## 2. What Liatrio SDD is

Liatrio SDD is a workflow for implementing one piece of work with an AI
coding assistant. The README describes it in one line: "A skill-first
spec-driven development workflow for collaborating with AI agents to deliver
reliable outcomes." One person runs four phases in order.

1. Generate a specification. The assistant checks the size of the request
   and reads the repository. It looks up current practice for the
   technologies involved and decides whether it has to ask questions. Then
   it writes the specification.
2. Generate a task list. The assistant derives parent tasks from the
   specification and waits for approval. It then adds sub-tasks and a table
   of the files to create or change, and audits its plan against six gates.
3. Manage tasks. The assistant implements the tasks one at a time. For each
   parent task it writes a proof file and creates one commit.
4. Validate the implementation. The assistant compares the changes with the
   specification and the task list, checks the proof files again, and writes
   a report with a PASS or FAIL result.

The project states its own scope. The playbook says that "SDD is
intentionally focused on the implementation loop". Research, prioritization
and architecture decisions come before that loop, in tickets or design
documents, and the review of the pull request comes after it. The playbook
names the work that fits: stories that are ready to pick up, small-to-medium
features of one user story, and work by one person with an assistant.

The first commit is dated 2025-10-02, and version 1.0.0 followed on
2025-10-10. The early versions contained an MCP server and a generator for
slash commands. Version 1.9.0 of 2025-11-17 reduced the project to its prompt
files. Version 1.13.0 of 2026-04-09 added the planning audit. Version 1.14.0
of 2026-05-26 added the skill, which the README now recommends over the
prompt files. This entry describes version 1.14.3 of 2026-08-12, which was
the head of the main branch on 2026-11-05. The license is Apache 2.0.

| Item | Value |
|---|---|
| Origin | Published by Liatrio under its GitHub organization `liatrio-labs`; first commit 2025-10-02. The README names the repository `snarktank/ai-dev-tasks` as the foundational example of such a workflow in Markdown |
| Authors | 60 commits on the main branch: 23 by the release automation, 26 by one maintainer and 11 by seven further authors |
| License | Apache-2.0 |
| Language and runtime | Markdown: one skill file, four phase references and four equivalent prompt files. One Python script reports the workflow state from the files in `docs/specs/`; the assistant runs it at the start of each invocation |
| Distribution | The skill is installed from the repository with the `skills` command-line tool. The prompt files are installed as slash commands with a separate Liatrio tool, or copied by hand. Releases are tags with a generated changelog |
| Artifacts in a project that uses Liatrio SDD | One directory per specification under `docs/specs/`: the specification, optional questions files, the task list, the audit report, one proof file per parent task and the validation report |
| Agent coverage | Any assistant that supports skills or slash commands. The README mentions Claude Code, Cursor, Windsurf and Codex and notes that support for skills varies. The repository records no test run per assistant |
| Governance | One code-owner team; conventional commits and a check of pull-request titles; an issue first for significant changes; automated releases from the main branch |
| First publication | Version 1.0.0 on 2025-10-10; 23 releases up to version 1.14.3; 94 stars and 10 forks when read on 2026-11-05 |

Three constructs account for most of the design.

The **demoable unit** organizes every document. A specification contains two
to four demoable units of work. Each unit has a purpose, functional
requirements in the form "The system shall ...", and proof artifacts. A proof
artifact is a piece of observable evidence: the output of a command, a test
result, a screenshot, a URL. The task list turns each unit into a parent task
with its proof artifacts. During implementation the assistant collects the
evidence for a parent task in one proof file, before it creates the commit
for that task. The proof file starts with what the task proves and shows the
raw output after that. The validation phase checks each proof artifact
again.

The **gates** stop the assistant until a condition is met, and most of them
wait for the human. Phase 1 must declare whether it has enough information.
If it has not, it writes a questions file with options and a recommended
answer for each question. It then stops until the user has answered in the
file. A question that could change the specification may not be deferred.
The phase text says: "Do not answer a question by punting it to SDD2 or
later phases." Phase 2 waits for approval of the parent tasks before it
writes sub-tasks. Its planning audit has four required gates and two
advisory ones. Two required gates concern the evidence: every functional
requirement maps to a planned test, and the proof artifacts are observable.
The other two require that the plan agrees with the standards of the
repository and that no significant question is open. A gate that cannot be
evaluated counts as failed. The assistant proposes corrections, applies them
only after approval, and audits again. Phase 3 asks how often it should
stop: after each sub-task, after each parent task, or at the end. Phase 4
applies six gates. One gate fails when a requirement remains unverified. A
second fails when a changed source file maps to no requirement and no task.
A third fails when a proof file contains credentials. The last instruction
of the workflow asks the user for "a final code review" before the merge.

**State in files** connects the phases. Every phase writes its result into
the directory of the specification, and task states are check marks in the
task list. A script reads these files and reports which phase is due. The
skill handles one phase per invocation and tells the user what to ask for
next. Because the state is in files, the next phase can run in a new chat.
Each response starts with a short marker that names the phase. A missing
marker tells the user that the assistant may have lost its instructions. The
README calls the marker an indicator that the instructions are "probably"
being followed. The project took the technique from a conference talk of
November 2025.

Several smaller rules belong to single phases. Phase 1 reports whether the
request has the right size. Its examples of work that is too large include a
full authentication system. Its examples of work that is too small include
the fix of an off-by-one error, and for such work it suggests direct
implementation. Phase 1 also looks up current guidance for the technologies
involved, with a preference for primary sources, when that guidance could
change the specification. The specification must be "clear enough for a
junior developer to understand and implement". Each phase gives the assistant
a role, from product manager in phase 1 to quality assurance engineer in
phase 4. Commits follow the conventions of the repository and name the task
and the specification.

The project documents two runs. The repository contains the complete set of
documents for one of its own documentation pages. The validation report of
2026-03-17 lists 16 of 16 requirements as verified and names the model that
performed the validation. The playbook presents a second run in full, a
pre-commit hook for spell checking. It reports 11 of 11 requirements as
verified. By the git history the implementation took under six minutes, and
the playbook gives less than 15 minutes for the workflow from start to
finish.

---

## 3. What the specification is for

### 3.1 The agreement

Both methods put the work on intent first. The first phase of Liatrio SDD may
only create the specification. Its list of actions that the assistant must
never take starts with "Start implementing the spec". PCD authors and lints a
specification before a translator is invoked.

Both refuse to leave a significant decision for later. Liatrio SDD turns an
ambiguity that could change the specification into a question to the user,
and forbids its transfer to a later phase. Its open-questions section is
reserved for items that change neither scope nor requirements nor evidence.
PCD's bar is that the reader of the specification can produce the component
without deciding a behavior.

Both plan the evidence before the implementation. Every Liatrio requirement
needs a planned test, and every unit names the artifacts that will show it
working. PCD's interview asks for an example of every operation, and its
review before translation reports a BEHAVIOR without one. The tests are
written from the EXAMPLES.

Both say what was not verified. The validation of Liatrio SDD marks each
requirement as Verified, Failed or Unknown, and its audit fails a gate for
which it has no evidence. PCD's translation report rates how each EXAMPLE was
verified and lists unverified claims.

Both keep a named human responsible for the result, and both keep their
documents as plain Markdown in the repository.

The agreement is independent. Liatrio's first commit dates from October 2025,
and its README names an earlier Markdown workflow as its model. By its
changelog, PCD's first draft dates from January 2026. Neither project cites
the other.

### 3.2 Both documents are structured

A Liatrio specification follows a fixed template. It has an introduction,
goals, user stories, the demoable units of work with their functional
requirements and proof artifacts, and non-goals. Then follow design
considerations, repository standards, technical considerations, security
considerations, success metrics and open questions. The task list has its own
form: parent tasks with proof artifacts, sub-tasks, and a table of relevant
files. The audit report and the validation report are tables with a status
per gate or per requirement.

A PCD specification has seven required sections: META, TYPES, BEHAVIOR blocks
with ordered STEPS, PRECONDITIONS, POSTCONDITIONS, INVARIANTS and EXAMPLES.
INTERFACES and DEPENDENCIES are optional sections, and META selects the
deployment template. A linter checks the form of a specification against a
published rule set.

Both are plain Markdown, kept in the repository. The form is checked in
different ways. The Liatrio template is an instruction to the assistant. The
script of the skill reads which files exist, the check marks of the tasks and
the PASS or FAIL lines. No program checks the structure of a Liatrio
specification. PCD checks the form of a specification with a program.

### 3.3 A document for one run, or the source of the component

The two methods differ in what the specification describes, in who reads it,
and in what it is after the implementation.

A Liatrio specification describes one piece of work in an existing
repository. Its scope check names the fitting size: one command-line flag,
one API endpoint, one user story. Larger work is split into several
specifications. The specification is written for a junior developer, and the
user reviews it "as the source of truth for implementation". It contains
technical considerations and the standards of the repository, and the task
list that follows names the files to change. The assistant that planned the
work implements it in that repository, task by task, and the user can stop
it at each checkpoint.

After the run the documents have served their purpose. The playbook says:
"Specs, tasks, proofs, and validation reports are process artifacts, not
living documentation." It continues: "After that loop completes, the code and
tests become the durable source of truth." A team may keep the documents,
archive them or remove them. Feedback on details is handled in the same run.
A change of direction, by the playbook, "usually means starting a new spec
and a new implementation run".

A PCD specification describes a component. A file path, a function name or a
programming language in a specification is a review finding. The reader is
the translator: in a full run it receives the specification, the deployment
template and the hints, and no existing code. The result must compile and
must pass tests that were written from the specification before the
implementation existed. After the first translation the specification stays
the source. A correction edits the specification, the hints or the template,
and the code is generated again.

| | Liatrio SDD | PCD |
|---|---|---|
| The document describes | One piece of work in an existing repository; the fitting size is one command-line flag, one API endpoint or one user story | A component: types, interfaces, ordered steps, conditions, invariants, examples, deployment |
| Written for | A junior developer; the user reviews it as the source of truth for the implementation | The translator; whatever a reader without context needs is in the specification, the template or the hints |
| Files and technical choices of the implementation | Technical considerations and repository standards in the specification; a table of relevant files in the task list | Absent; a file path or function name in a specification is a finding |
| The implementer | The assistant that planned the work, in the repository, task by task, with human checkpoints | A translator that, in a full run, has no codebase to read |
| Completeness is established by | The user's review of the specification and an audit of the plan by the assistant | A translation run that passes the compile gate and the tests written from the specification |
| After the implementation | The documents are process artifacts and may be archived or removed; a change of direction starts a new specification | The code is generated again from the specification; a correction edits the specification, the hints or the template |
| Durable product | Code and tests | Specification, plus hints, templates and procedures |
| Finished when | The validation gates pass and the user has reviewed the code before the merge | A machine that has never seen the code can produce the binary from the specification |

Stated as one sentence: **Liatrio SDD governs the task; PCD governs the
artifact.** Liatrio SDD organizes one piece of work, from the request to the
validated change. PCD keeps a document from which the component can be built
again.

Birgitta Böckeler distinguishes three levels of specification-driven
development: spec-first, spec-anchored and spec-as-source. In those terms
Liatrio SDD is spec-first: the specification is written first and used for
the task, and the project says that it is not maintained afterwards. PCD is
spec-as-source. Her article does not discuss the Liatrio workflow, so the
classification is this entry's.

Section 4 follows from this difference. A document that is set aside after
the run cannot be the object of an audit of the component (4.1). A method
that keeps the code links each change to its task, and a method that derives
the code has to record the derivation (4.2). A plan that names the files of
one repository has the programming language of that repository (4.4).

The reverse follows as well. Liatrio SDD asks nobody to describe a component
in full. The specification covers one change, and the assistant reads the
repository for everything else. For that reason the workflow applies to any
repository and to any change of the fitting size, and by the project's
example a run takes minutes. PCD applies to components that are small and
stable enough to be written down in full.

### 3.4 What the two designs share

The following moves appear in both methods. Each project developed them
independently.

| Shared move | In Liatrio SDD | In PCD |
|---|---|---|
| Intent is written down and approved before any code exists | Phase 1 may only create the specification; the user reviews it before planning starts | Specification authored and linted before a translator is invoked |
| Significant decisions are closed in the document | An ambiguity that could change the specification becomes a question to the user and may not be deferred to a later phase; the open-questions section is for non-blocking items only | A specification is finished when a translator that has never seen the code can produce the binary |
| Requirements are testable statements | "The system shall ..." with the instruction "clear, testable, unambiguous" | BEHAVIOR blocks with ordered STEPS; EXAMPLES in Given / When / Then |
| Evidence is planned before the implementation | Proof artifacts per demoable unit; the audit fails a requirement without a planned test | EXAMPLES in the specification; the linter requires a negative-path EXAMPLE for a BEHAVIOR with error exits; tests are written from the EXAMPLES |
| Authoring starts with questions | A questions file with options and a recommended answer per question; the user answers in the file | Interview mode: one question at a time, each phase confirmed |
| Existing material is accepted as input | An idea, a ticket or an issue | Notes, a design document or a ticket, in the gap-fill option of the interview |
| Large work is cut into parts that each show a result | Two to four demoable units per specification; one commit per parent task | Milestones; the first is a scaffold that must compile |
| What was not verified is reported | Validation marks each requirement Verified, Failed or Unknown; an audit gate without evidence fails | The report rates each EXAMPLE and lists unverified claims; the lint line reads clean, findings or not run |
| A record links the result to its document | Commit messages name the task and the specification; the validation report cites file lines, commits and proof files | The specification hash in every generated artifact; a hash of every input in the translation report |
| The method reads the conventions of its target | Repository standards are read from the guidance files of the repository and recorded in a table | The deployment template and the hints state the conventions of the target |
| Distribution as agent skills | One skill that routes to four phase references | Two skills, `pcd-spec` and `pcd-translate`, with modes |
| The framework is used on itself | The repository contains the document set of one of its own documentation pages | The PCD tools were generated from their own specifications |

The table needs three qualifications.

The evidence is planned early in both methods. The tests are written at
different times and by different authors. In Liatrio SDD the assistant that
implements a task also tests it, with the testing approach of the
repository. The sub-task steps of phase 3 list the implementation before the
test. A hand-over line in phase 2 announces test-driven work, and the text
of phase 3 does not require it (section 7). In PCD every run writes the
tests before the implementation, and in dual-LLM mode a model from another
family writes an independent suite first.

No rule of Liatrio SDD separates the author from the checker. The planning
audit runs in the phase that wrote the plan and ends with a self-check. The
validation can run in the session that implemented the work. Because the
state is in files, each phase can also run in a new session or with another
assistant. The workflow permits that and does not require it. The validation
report records which model performed it. PCD separates roles by input. Each
role receives a defined set of files, and the skill refuses a test-author
run whose input directory contains an implementation. PCD does not test
whether a session started without an earlier conversation.

"Used on itself" means two different things. The Liatrio example is a
documentation page that was written with the workflow. Three PCD tools were
generated from their own specifications, and two of them are older than the
current text of their specification.

---

## 4. Where PCD differs

The differences follow from 3.3. They are listed in the order a skeptical
reader will raise them.

### 4.1 The audited object

Under Liatrio SDD the code is what ships, and it is what a reviewer or an
auditor examines. The workflow adds work in front of the code and evidence
beside it. The user reviews the specification before planning starts. The
proof files and the validation report are written for a human reviewer. The
last instruction of the workflow asks the user for a final code review
before the merge, so the review of the code remains. After the run the
documents have no further role. The playbook calls them process artifacts
and leaves it to the team whether they are kept.

Under PCD the specification is the object under examination and the binary is
its derivative. The human reviews the specification and the reports of the
pipeline, and does not review the generated code. The pipeline checks the
code: the compile gate, the tests written from the specification, the
reviewer findings, and the hash that records which specification the artifact
was derived from.

Liatrio SDD adds a specification to a process that still ends in a code
review. PCD replaces the human review of the code with a human review of the
specification.

### 4.2 Linking document and code

Liatrio SDD links document and code by reference, at the time of the work.
Each commit names its task and its specification. The task list contains a
table of the files that the plan expects to change. The validation phase
lists the changed files from the git history and maps each to a requirement
or a task. It fails when a source file was changed without such a link. Its
coverage matrix cites file lines, commits and proof files for every
requirement.

These links describe the state at the end of the run. The specification
template has no version field and no hash, and the workflow has no later
step that compares a specification with the code. That is consistent with
the project's position: a document that is not maintained is not compared
again.

PCD links document and code by derivation. The translation report lists a
labeled hash of every file consumed as an input. The specification hash is
embedded in source headers, in the binary's version output, in package
metadata and in container labels. The version of a specification must change
whenever its bytes change. Recomputing a hash answers one question: which
specification text did the run that produced this artifact record? That the
code implements that text is for the tests and the review to show.

Each method has a check that the other lacks. Liatrio SDD checks the changed
files against the plan. PCD records a hash of every input of a run. It
records no hash of the generated files, so a hash comparison does not find a
later edit of a generated file (section 9).

### 4.3 Verification: who checks, and by what means

Liatrio SDD verifies in four ways. The human approves at fixed points: the
specification, the parent tasks, the corrections after the audit, the
checkpoints during implementation, and the code before the merge. The
assistant audits its plan against written gates. The quality checks of the
repository run for every task: tests, linting, pre-commit hooks. The
validation phase checks the proof artifacts again, by running the commands
and opening the references they name.

Every check that a model makes can be made by the assistant that did the
work. The workflow asks for no second model and for no reader in a fresh
context (3.4). Its independence comes from the human at the checkpoints and
from the tools of the repository.

PCD uses a second model in its dual-LLM mode, and it uses execution. The
translator must produce an implementation that passes the compile gate and
the tests that were written from the EXAMPLES before it existed. A review and
a security review of the generated artifacts are defined as further steps in
the translation skill. The public repositories contain no report from a run
of them yet.

Neither project has published a measurement of how many defects its checks
find. Liatrio's evidence is the two runs described in section 2. Its testing
notes say that the automated tests cover the state script and the text of
the skill, and that evaluations with a live assistant are tracked
separately.

### 4.4 Language neutrality

A Liatrio specification has a section for technical considerations and one
for the standards of the repository, and the task list names files. The
implementation follows the conventions of the codebase it is written in. The
document belongs to one repository and to the language of that repository.

One rule of phase 1 asks for general wording. Before the specification is
shown to the user, the assistant checks that the requirements make no
project-specific assumption. The stated aim is that "another repository
could reuse the structure with only context substitutions". The rule
concerns the wording of the requirements. The plan and the result stay
specific to one codebase.

PCD keeps language-specific realization out of the specification and places it
in hints and deployment templates. In June 2026 one specification text was
translated to C++, Go and Rust, and the three translation reports record the
same specification hash. This only has meaning where the specification is the
durable artifact.

### 4.5 Supply chain and sovereignty

The runtime of Liatrio SDD is Markdown and one Python script. The skill file
declares no MCP server and no hook, and the workflow calls no service of the
project. The assistant runs the script with the Python interpreter of the
user's machine at the start of each invocation.

The documented installation fetches from the network at install time. The
recommended command installs the skill from the repository with a
third-party command-line tool. The README notes that this tool collects
telemetry unless an environment variable disables it. The second command
runs Liatrio's slash-command tool from the default branch of its repository
and takes the prompt files from the main branch of this one. Neither command
names a release. The third way is to copy a prompt file by hand.

Releases are tags that an automated workflow creates from the commit
messages. The tags are not signed, and the release configuration builds no
archive, no checksum file and no build provenance (section 7). The release
workflow pins its token action by commit and its other actions by version
tag.

Phase 1 requires web research when a technology could change the
specification, and the assistant consults external sources for it. In an
environment without network access that step cannot run. The phase text
allows the statement that no research was needed.

The files are few and readable: about 1,750 lines in the four phase
references and 104 lines in the skill file. Under the Apache license an
organization can review them, keep its own copy and distribute that copy in
its own way.

PCD's target is a pipeline an organization can run itself. The target includes
tools shipped as signed packages from a controlled build service, assets
embedded at build time, offline and reproducible builds, and an SPDX software
bill of materials. It also includes models hosted in a chosen jurisdiction or
run without network access. Today a PCD skill checks its own files against a
manifest of hashes at the start of a session. It refuses to build, fetch or
substitute a missing tool.

Liatrio SDD depends on no model provider: it uses the assistant and the model
of its user. For a customer whose constraint is sovereignty, the open points
are the installation path and the web research step. The method itself
raises none.

### 4.6 Certification: the same goal, a different distribution of evidence

Neither approach makes certification impossible, and the comparison is easy to
overstate in PCD's favor. Liatrio SDD follows the familiar route: the code is
the evaluated object, and the assistant is part of the development
environment. It adds a trail that an evaluator can read for each change. The
trail consists of the specification with its non-goals, the questions and
their answers, the audit of the plan, the proof files, the validation
report, and commits that name their task. The playbook describes the result
as "a self-documenting, auditable process" and mentions compliance tracking.

Three limits apply. The trail is written by the assistant whose work it
documents, and no rule requires a second reader. The trail covers one change
and says nothing about the component afterwards; by the project's own
statement the documents need not be kept. The evaluated object changes with
every change, and each change enters code-level review again, as recorded
for the other entries in this directory.

Security is a section of the specification and a gate of the validation. The
template asks for security requirements and for the handling of sensitive
data. The gate concerns the proof files, which must contain no credentials.
The workflow has no separate security review of the code. It leaves that to
the checks of the repository and to the human review.

Under PCD the human-reviewed object is the specification, and a machine can
check which specification an artifact was derived from. PCD does not assert
that a certifiable system must be built this way. It redistributes the
evidence: it trades code-level review effort for provenance and pipeline
evidence, and it depends on a certifier accepting that trade. That acceptance
has not yet been demonstrated at certification scale, as the whitepaper
records. Both routes are open.

---

## 5. Where Liatrio SDD leads

Liatrio SDD is ahead of PCD in the following respects.

- Applicability and cost of adoption. One skill serves any repository and any
  change of the fitting size, with the assistant the engineer already uses.
  By the project's example a run takes minutes. PCD needs a component whose
  behavior can be written down in full, a deployment template for its type
  and its two tools. It has no method for a change to a codebase that is
  maintained and not regenerated.
- A fit check before any writing. Phase 1 reports whether the request is too
  large, too small or of the fitting size, and it pauses when the size is
  wrong. PCD's interview proposes milestones when a specification has more
  than ten BEHAVIORs. It does not ask at the start whether the work suits the
  method.
- Questions in a file. The questions file gives options, a recommended answer
  and a reason for each question, and the user answers in the file. The file
  stays beside the specification, so the record shows what was asked and what
  was decided. PCD's interview is a conversation. Its answers enter the
  specification, and the questions are not kept. The Spec Kitty and EGM
  entries recorded a related gap for the reasons behind decisions.
- A stop rule for unverified requirements. The validation fails while any
  functional requirement is marked Unknown. PCD's translation report rates
  every EXAMPLE as High, Medium or Low, where Low means that no test covers
  it, and lists unverified claims. The procedures record the rating. They
  contain no rule that a Low rating stops the run or the review.
- A check of the changed files against the plan. The validation classifies
  every changed file and fails on a source file without a link to a
  requirement or a task. PCD states the rule that generated code is not
  edited by hand. It has no check that would find such an edit (4.2).
- Evidence written for a reviewer. A proof file starts with what the task
  proves, explains each artifact, and shows the raw output after that. PCD's
  translation report is a record of the run, and its required content has no
  summary for a reader with little time.
- Reading the target repository. The workflow reads the guidance files of the
  repository and records what it found in a table. Its audit fails when
  `AGENTS.md` or the root README exists and was not read. That step makes the
  workflow usable in an existing codebase. PCD states conventions in
  templates and hints and starts without a codebase.
- A signal for lost instructions. The marker at the start of each response
  shows whether the phase instructions are still in effect, and the state in
  files lets a new chat continue. PCD has no such signal. The HDAE entry
  lists a context-recovery protocol among its follow-up actions.
- Tests of the workflow's own mechanics. The script that determines the phase
  has 17 tests, and five further tests check the text of the skill. Both
  groups run in CI. PCD's skills have a manifest check and dry runs.
- Teaching material. The playbook publishes one run in full: the
  conversation, the specification, the task list, the proof files, the
  validation report and an analysis of the git history. The project presents
  the workflow as a way to learn how to work with an assistant.
- Adoption and maintenance. On 2026-11-05 the repository showed 94 stars and
  10 forks, eight commit authors besides the release automation, and 23
  releases in ten months. PCD has one maintainer.

---

## 6. Net positioning

The two methods answer different questions.

- Liatrio SDD answers: how does one engineer complete one piece of work with
  an AI assistant in the repository as it is, with a plan, evidence and
  checkpoints? The code is retained, and the documents may be discarded.
- PCD answers: what artifact can be certified today and regenerated in ten
  years? It applies to components whose behavior can be written down in full.

A team can use both. Liatrio SDD's scope check excludes work that is too
large for one specification, and PCD's scope excludes whatever cannot be
specified completely. Most systems contain many small changes to retained
code and a few components that justify a complete specification.

Liatrio SDD does not keep its specification and does not claim to. On
regeneration PCD's nearer neighbor is Codeplain, the subject of the first
entry in this directory. For a design that is kept beside the code, the
neighbor is the Elephant-Goldfish Model, the subject of the fourth.

### 6.1 A named integration point

The two methods can meet in two places. Neither has been tried.

A Liatrio specification can be the input of a PCD specification. The
interview mode of `pcd-spec` has an option that starts from existing
material: notes, a ticket or a design document. It then asks only for what
is missing. A Liatrio specification is such a document. It describes a
change. The interview therefore has to add what a component specification
needs and what a change request leaves to the repository: the types, the
ordered steps and the error exits.

| Liatrio SDD element | PCD counterpart |
|---|---|
| The specification as a whole | Input for the gap-fill option of the interview mode of `pcd-spec` |
| Goals and user stories | Not part of a specification; the reason for a change goes to the changelog beside it |
| Functional requirement, "The system shall ..." | BEHAVIOR with ordered STEPS and error exits |
| Proof artifact | EXAMPLE in Given / When / Then; the tests are written from it |
| Non-goals | BEHAVIORs with `Constraint: forbidden` |
| Repository standards and technical considerations | Deployment template and hints |
| Open questions | Closed before translation |
| Task list with relevant files | Not written; the translator produces the file list from the template and the hints |
| Planning audit | Linter and spec-review mode |
| Proof files and validation report | Test report, and translation report with the rating per EXAMPLE |
| Gate on changed files | No counterpart today (section 9) |

The second place is the repository-standards step. Phase 2 of Liatrio SDD
must read `AGENTS.md` and the root README of the repository, and its audit
fails when it has not. A validation gate requires that the implementation
follows the standards found there. The PCD repository states in its
`AGENTS.md` that generated files are never edited by hand and that a wrong
result is corrected in the specification. A repository that is developed
with PCD can state the same rule. By the text of both projects, a Liatrio SDD
run in such a repository would then have to plan a change to the
specification and a new translation. It could not plan a direct code edit.
Its proof artifacts would be the lint result, the translation report and the
test results. Whether an assistant follows that reading is open. The steps
of the two skills would be started by the user.

### 6.2 Framing for talks

Liatrio SDD governs the task; PCD governs the artifact. Liatrio SDD shows
how little is needed to bring order to AI-assisted work on a small change.
It needs a specification, a plan, evidence per task and a validation, all in
Markdown and all in the repository. It also says what the specification is
when the task is done: a process artifact. PCD asks what a specification
needs if it is to be kept as the source. Its answer is completeness for a
reader without the repository, tests written from the specification, and a
record of what the code was derived from. Where the code is kept and
reviewed in any case, the lighter method is the proportionate one.

---

## 7. Open questions and claims to verify

Confirm the following against an installed copy of the skill and a working
trial, and not against the repository read alone.

- Does the `skills` command-line tool install a named release, or the current
  state of the default branch? The documented commands name no version. The
  answer decides how the installation is described in 4.5.
- Do the release pages on GitHub offer files beyond the generated source
  archives? The release configuration builds none, and the tags in the clone
  are unsigned. The release pages were not read for this entry.
- The README says that phase 2 creates a baseline planning commit. The phase
  text at the commit read has no such step. Which one is current?
- The hand-over line of phase 2 says that phase 3 executes the task list
  "using TDD". The sub-task steps of phase 3 list the implementation before
  the test. Which order is intended?
- The four prompt files and the four phase references of the skill differ in
  38 to 69 lines per phase. This entry describes the phase references of the
  skill.
- Is the workflow in use under a regulated review, and what is presented
  there as evidence? The playbook mentions compliance tracking and gives no
  case.
- `docs/faq.md` and `docs/press-release.md` in the repository are marked as
  hypothetical examples. Do not cite the customer statements and figures in
  them.
- The run times and counts in section 2 come from two runs that the project
  selected and reported itself.
- Four unmerged branches indicate possible directions. One adds a subagent
  option for the implementation phase (last commit 2026-09-10), and one adds
  telemetry at phase level (2026-06-16). Two more are named for parallel
  execution (2026-02-16) and for evaluations with a live assistant
  (2025-10-14); their content was not read. None is on the main branch.
  Check their state before describing the runtime as free of telemetry or
  the workflow as single-threaded.
- Stars, forks, commit and release counts are as read on 2026-11-05. Cite
  them with the date.

Three statements about PCD need the same care.

- A review and a security review are defined in the translation skill, and
  the public repositories contain no report from a run of them. Describe
  them as steps of the method until a run is recorded (4.3).
- Three PCD tools were generated from their own specifications. Two of them
  are older than the current text of their specification. Say so when citing
  the last row of the table in 3.4.
- The statement that PCD has no stop rule for a Low rating rests on the
  procedure files of the translation skill. The deployment templates were
  searched for such a rule, and none was found.

---

## 8. Sources

- github.com/liatrio-labs/spec-driven-workflow, main branch, read on
  2026-11-05 at commit 7010cb8 (2026-08-12), which is the tag v1.14.3. Read:
  README; `skill/SKILL.md`, the four phase references under
  `skill/references/` and the head of the state script; the four prompt
  files under `prompts/` by comparison with the references; `CHANGELOG.md`;
  `CONTRIBUTING.md`, `MIGRATION.md` and `docs/TESTING.md` in part; the
  release configuration and the workflow files; the tests by count; the
  document set under `docs/specs/`; the note on context verification
  markers.
- The SDD Playbook, liatrio-labs.github.io/spec-driven-workflow, read on
  2026-11-05 from the `docs/` directory of the same commit: the index page,
  the comparison page and the common-questions page in full; the other pages
  by their headings.
- The GitHub page of the repository, read on 2026-11-05. Used for stars and
  forks.
- Birgitta Böckeler, "Understanding Spec-Driven-Development: Kiro, spec-kit,
  and Tessl", martinfowler.com, 2025-10-15. Used for the three levels in 3.3.
- github.com/mge1512/pcd at commit 8d04918 and
  github.com/mge1512/zypper-declarative, read on 2026-11-05. Used for the
  statements about PCD's own artifacts and procedures.

---

## 9. Follow-up actions (non-normative)

- Decide what a Low rating of an EXAMPLE means for acceptance: a stop, or a
  reason that a human signs.
- Add a check for generated files: a commit that changes one of them also
  contains a translation report, or the check fails.
- Consider two changes to the interview mode: a fit question at the start,
  and a questions file with recommended answers as an option. Decide the
  second together with the rationale record from the Spec Kitty and EGM
  entries.
- Put a short summary for a reviewer at the top of the translation report:
  what the run shows and what it leaves unverified.
- Try a response marker in the long modes of the two skills, together with
  the context-recovery item of the HDAE entry.
- Use the sentence from 3.3 and the paragraph in 6.2 in talks.

---

## Changelog

- 2026.11.05.01 - Initial draft. Sources: the main branch of the
  `spec-driven-workflow` repository at commit 7010cb8 (tag v1.14.3), the
  project's playbook site and the GitHub page of the repository, all read on
  2026-11-05. No trial of the workflow was run. Statements about PCD's own
  artifacts and procedures were compared with the `pcd` and
  `zypper-declarative` repositories on the same day. Written for public use:
  no internal project names appear. Certification wording follows the Spec
  Kitty, HDAE and EGM entries: both routes are open and distribute the
  evidence differently. Mirror updates to the whitepaper (A.12, Comparative
  Summary, References) and to the technical reference (section 18,
  References) are not yet applied.
