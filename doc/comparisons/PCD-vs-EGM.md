# PCD vs the Elephant-Goldfish Model

Version: 2026.11.05.01
Date:    2026-11-05
Author:  Matthias G. Eckermann <pcd@mailbox.org>
Status:  Draft, intended for public use
Scope:   doc/comparisons/

---

## 1. Purpose

This is the fourth entry in `doc/comparisons/`. It positions PCD against the
Elephant-Goldfish Model (EGM), a method for designing and reviewing software
changes with AI agents. Dave Rensin described the method in an article in
2026, and a team at Google released it as an open-source plugin in October
2026.

The two methods start from the same test. A design document is good enough
when a reader that has seen nothing except the document can work from it. EGM
calls that reader a Goldfish. PCD calls it the translator. Both state the bar
in similar terms: the reader must be able to do the work from the document
without making a design decision.

They differ in what the reader is asked to do, and in what the document is
afterwards. EGM asks fresh readers for findings on the design. The team's
coding agent then implements the approved design once, and from then on the
document and the code are both kept and compared. PCD asks a fresh reader for
the implementation, tests the result, and generates the code again from the
document when the document changes. That difference is the subject of this
entry. Section 6.1 describes how the two methods could be combined. No trial
of the combination has been run.

The entry exists because the question "why not use EGM instead of PCD?" will
be asked. EGM was published in late October 2026 with a large user base
inside Google, by the project's own statement. Its repository collected about
11,900 stars in two weeks.

Sources are listed in section 8. Claims that rest on a single repository read
are flagged in section 7 as "to verify" before any public use.

---

## 2. What EGM is

EGM is a method for getting a software change designed and reviewed before any
of it is implemented. The project's summary is "Design with an elephant.
Review with a goldfish." The method has three roles.

- The Elephant is an agent session that has the project context: the
  requirements, the earlier decisions and the code. It explores the problem
  with the human and drafts the design document.
- A Goldfish is a reviewer that runs in a fresh context. It reads the document
  and the source material the document refers to. It has no access to the
  conversation that produced the document, and it may not modify anything.
- The human decides the architectural trade-offs, approves the design and is
  responsible for the outcome.

The method came first and the plugin second. Rensin's article describes a
sequence of conversations that an engineer can run by hand. One long session
produces the design document. New sessions then test whether the document can
be understood, criticized and implemented without the earlier conversation.
A team at Google packaged that sequence as agent skills. The README states
that thousands of Google engineers use it. The first public commit is dated
2026-10-21, and the launch post followed on 2026-10-22. The license is Apache
2.0, and the README notes that this is not an officially supported Google
product. Version 0.1.0, which the project calls its first prerelease, was
tagged on 2026-10-27. This entry describes the main branch nine days later,
55 commits after that tag. Several rules described below were added in those
commits (section 7).

| Item | Value |
|---|---|
| Origin | Article by Dave Rensin (2026); plugin built inside Google and published in October 2026 |
| Authors | The launch post is signed by Brian Suchy, Jim Quinn, Elise Ewing, Oliver Newth and Dave Rensin |
| License | Apache-2.0; contributions require the Google contributor license agreement |
| Language and runtime | Markdown skills, personas and standards; no executable file in the runtime package; Python for packaging and tests in the repository |
| Distribution | Installed through each host's plugin mechanism, with the repository as the source; in addition, runtime archives per release with checksums and signed build provenance |
| Artifacts in a project that uses EGM | `docs/designs/<feature>.md`, an optional `SITEMAP.md`, README files per directory; working files `concepts_brief.md`, `design_draft.md` and the review reports |
| Agent coverage | Recorded end-to-end runs on Antigravity CLI, Claude Code, Codex and Muse; partial verification on Claude Cowork and ChatGPT Work; a community-verified installation on omp; experimental adapters for Cursor, OpenCode and Copilot CLI in the repository, support unverified |
| Governance | Six code owners; two contribution tracks; a feature needs an accepted issue, then a merged design document, then an implementation pull request |
| First publication | Article in 2026; tag v0.1.0 on 2026-10-27; 149 commits on the main branch when read on 2026-11-05 |

The plugin has seven skills, nine persona files and five document standards.
Three constructs account for most of the design.

The **Goldfish review** has five stages, and one or more Goldfish run each
stage. A simplicity review asks whether the design solves the root problem in
the simplest way and reuses what the codebase already has. A comprehension
test has a fresh reader write, in its own words, what the document proposes
and how the existing system works. A section the reader cannot paraphrase is
a blocking finding, and a wrong paraphrase is a finding too. A critic review
examines the high-level part of the document against a feature specification
standard. A second simplicity review checks what the earlier stages added. A
readiness test examines the implementation plan. Two rules protect the
result. A review by the session that wrote the document does not count. The
skill says that "a review in the same context is not a Goldfish review", and
it stops when the host cannot create a fresh context. A reviewer that fails or
times out is recorded as not run, and never as a reviewer without findings.

The **document standards** state what a document must contain to pass. Two of
them state the completeness bar. The implementation readiness standard
describes its reader as an engineer, or a fresh AI session, who does not know
the codebase. A plan passes when that reader can execute it and produce the
correct changes at the first attempt, without guessing, asking or making a
design choice. Its zero-decision rule fails any plan that leaves a design or
architectural choice to the implementer. The re-implementation standard
applies the same test to a document that describes existing code: the reader
must be able to write functionally identical code from the document alone. At
the commit read, no stage of the five-stage review applies that standard. The
review loop accepts it by name.

The **living design** is the document after approval. An approved design is
checked in under `docs/designs/` and stays there. The team's coding agent
implements it. EGM has no implementation skill and leaves that phase to the
tools each team already uses. The guide allows the engineer to correct a minor
mistake of the agent directly in the code, or to ask the agent for a targeted
fix. When the result is far from the intent, the engineer returns to the
design document. A separate skill compares the checked-in design documents
and READMEs with the code and lists the differences, and the project
recommends a weekly run. For each difference a human decides whether the
document or the code changes.

Several smaller rules belong to single skills. The review loop ends after two
consecutive rounds without an open blocking or major finding. At five rounds,
or when a new finding would be resolved by reverting an earlier fix, it asks
the human whether to accept, to continue or to return to drafting. Inside the
five-stage review a stage is repeated once at most. In the resolve skill, the
session that wrote the document answers each finding in one of four ways. It
contests the finding with evidence the reviewer could have read, escalates it
to the human, proposes a fix, or declines a minor finding. It may not contest
a comprehension gap and may not change a severity. In the draft skill, a
design decision with more than one defensible answer is presented as up to
three options without a ranking, and the human chooses. The brainstorm, draft
and review skills forbid flattery and promotional adjectives. By the
architecture document, every skill writes long output to files and never into
the chat. The interactive skills ask one question at a time.

For an existing codebase, a bootstrap skill writes a README for each
directory, from the leaves upward. For a component directory the sections are
Interface, Invariants, Side Effects and Verification.

The project applies the method to itself. A feature for EGM needs an issue
that the maintainers have accepted, then a merged design document, then an
implementation pull request. The contribution policy expects the design
document to be drafted and reviewed with the EGM skills. Every pull request
needs a human author who can explain each changed line.

---

## 3. What the document is for

### 3.1 The agreement

Both methods distrust a document that only its authors have read. EGM's reason
is that the session which drafted a design assumes that what it remembers is
on the page. PCD's reason is that a specification is sufficient only if it
works without its author. Both therefore give the document to a reader that
has seen nothing else, and both define completeness in terms of that reader.
EGM: a plan passes when a fresh session can execute it without making a design
choice. PCD: a specification is finished when a translator that has never seen
the code can produce the binary from it.

The agreement goes further. Both say that human judgment belongs in the design
and that implementation is delegated. EGM's phrase is "Design is the New
Code"; PCD's README opens with "Specifications, not source code, are the human
deliverable". Both keep a named human responsible for the result. Both forbid
the reviewer to repair what it reviews. A Goldfish is read-only and reports
findings on the document. A PCD reviewer writes findings that cite
specification, hints or template text, and never writes a code patch.

The agreement is independent. By its changelog, PCD's first draft dates from
January 2026. Rensin's article appeared in 2026, and its abstract dates its
viewpoint to April 2026. The design documents for the core skills in the EGM
repository have June and July 2026 in their headers. Neither project cites
the other.

### 3.2 Both documents are structured

An EGM design document follows a template. Its high-level part has an
overview, the problem, goals, non-goals, open questions and critical user
journeys in Given / When / Then form. It continues with a technical plan with
diagrams, the alternatives considered with the reason each was rejected, and
non-functional requirements. The last part is a detailed implementation plan.
The plan lists every file to create or modify with the proposed change, and
ends with a verification plan. The review skills find the boundary between
the high-level part and the implementation plan by its heading and send each
part to different reviewers.

A PCD specification has seven required sections: META, TYPES, BEHAVIOR blocks
with ordered STEPS, PRECONDITIONS, POSTCONDITIONS, INVARIANTS and EXAMPLES.
INTERFACES and DEPENDENCIES are optional sections, and META selects the
deployment template. A linter checks the form of a specification against a
published rule set.

Both are plain Markdown, kept in the repository and reviewed through pull
requests.

### 3.3 A review of the change, or a build of the component

The two methods differ in what the document describes and in what the fresh
reader is asked for.

An EGM design document describes a change. It is written against a repository
that every reader can open. The reviewers check its file paths and signatures
against the code, and the implementing agent works in that codebase. The plan
lists the files to touch, so the document belongs to one repository and to
one programming language. The fresh readers are asked for findings. Whether
the plan leaves a decision open is a reviewer's judgment against the readiness
standard.

A PCD specification describes a component. A file path, a function name or a
programming language in a specification is a review finding. The fresh reader
is asked for the implementation: in a full run the translator receives the
specification, the deployment template and the hints, and no existing code.
Whether the specification leaves a decision open is visible in the run. The
result must compile and must pass tests that were written from the
specification before the implementation existed, and the report lists every
ambiguity the translator met.

The two documents are also written for different readers. EGM's resolve skill
tells the author not to add context only to satisfy a reviewer. Context is
added when a human reader of the document needs it, and a gap that a
zero-context agent alone would meet is escalated to the human. The first
reader of an EGM design is a person. In PCD the reader without context is the
translator, so whatever it needs has to be in the specification, the template
or the hints.

The methods differ again after the first implementation. Under EGM the code is
the product and is maintained. The design document is kept beside it, and a
periodic comparison reports where the two differ. Under PCD the code is a
build output. A correction edits the specification, the hints or the template,
and the code is generated again: in full, or for an isolated change in the
affected part only.

| | EGM | PCD |
|---|---|---|
| The document describes | A change: the design of a feature and the plan to implement it in an existing repository | A component: types, interfaces, ordered steps, conditions, invariants, examples, deployment |
| Files and functions of the implementation | Listed in the plan; every file to create, modify or delete | Absent; a file path or function name in a specification is a finding |
| The fresh reader | Reviewers that read the document and the referenced code, then an implementing agent with the codebase in front of it | A translator that, in a full run, has no codebase to read |
| The fresh reader is asked for | Findings on the document | The implementation, which is then compiled and tested |
| The document is written for | A human reader; the author does not add context that a zero-context agent alone would need | The translator; whatever a reader without context needs is in the specification, the template or the hints |
| Completeness is established by | Reviewers who judge the document against a written standard | A translation run that passes the compile gate and the tests written from the specification |
| After the first implementation | Code and document are both kept; a periodic report lists where they differ, and a human decides which one changes | The code is generated again from the specification; a correction edits the specification, the hints or the template |
| Durable product | Code, plus design documents and README files | Specification, plus hints, templates and procedures |
| Finished when | The review loop has had two consecutive rounds without an open blocking or major finding, or the human accepts what is left; then a human approves | A machine that has never seen the code can produce the binary from the specification |

Stated as one sentence: **EGM governs the review; PCD governs the artifact.**
EGM asks a fresh reader whether the document is sufficient. PCD has a fresh
reader build from the document and tests what it built.

Birgitta Böckeler distinguishes three levels of specification-driven
development: spec-first, spec-anchored and spec-as-source. In those terms EGM
is spec-anchored, because the document is kept and maintained beside the
code, and PCD is spec-as-source. Her article predates the publication of EGM,
so the classification of EGM is this entry's.

Section 4 follows from this difference. A document that describes a change
against a repository cannot be the audit point for the component, because the
repository is part of its meaning (4.1). A method that keeps document and code
has to compare them, and a method that derives the code has to record the
derivation (4.2). A review estimates whether a document is sufficient, and a
translation run tests it (4.3). A plan that lists files and signatures has a
programming language (4.4).

The reverse follows as well. EGM does not ask anyone to describe a component
in full. Only the change has to be decided completely, and the reader may
consult the code for everything else. For that reason EGM applies to a
repository of any size and age. PCD applies to components that are small and
stable enough to be written down in full.

### 3.4 What the two designs share

The following moves appear in both methods. Each project developed them
independently.

| Shared move | In EGM | In PCD |
|---|---|---|
| Intent is written down and approved before any code exists | Brief, design document, review and human approval come before implementation; the design skills may not write production code | Specification authored and linted before a translator is invoked |
| A reader without context is the test | Goldfish reviewers in fresh contexts; a review in the authoring context does not count | The test author receives the specification bundles, the template and the hints, and no other input; a full translation run starts without existing code |
| The bar is that no decision is left to the implementer | Implementation readiness standard with its zero-decision rule; re-implementation standard for existing code | A specification is finished when a translator that has never seen the code can produce the binary |
| The durable document is plain text in the repository | `docs/designs/`, README files per directory | Specification, hints, templates and procedures |
| Interview-first authoring, one question at a time | Brainstorm with open questions over several turns and a readiness gate; one question per gate | Interview mode: one question at a time, each phase confirmed |
| Findings go to the document | Reviewers are read-only; the resolve step edits the design document | Reviewers cite specification, hints or template text and never write a code patch |
| The author of a text does not grade it | The drafting session may not grade its own draft against the simplicity standard; a fresh reviewer does | The test author is a separate role, and in dual-LLM mode a model from another family |
| The human decides contested questions | Architecture choices are presented as up to three unranked options; a finding that the agent cannot refute from readable evidence is escalated | A human certifies the specification; the tiebreak never declares a winning implementation |
| "Not run" is never reported as "passed" | A failed reviewer is recorded as not run, and the round then does not count as clean | The lint line reads clean, findings or not run; unverified claims are listed in the report |
| Existing code has an entry path | Bootstrap: README files written bottom-up per directory | Reverse mode: a specification derived from source code |
| Distribution as agent skills from one source | Seven skills, one source for every host | Two skills, `pcd-spec` and `pcd-translate`, installed by cloning |
| Shipped files are compared with recorded hashes | Release archives with checksums and signed build provenance | A manifest of file hashes that each skill compares at the start of a session; it detects damaged files, and signed packages are the target |
| The framework is built with itself | A feature of EGM starts with an EGM design document | The PCD tools were generated from their own specifications |

The table needs three qualifications.

The fresh reader is isolated in two different ways. EGM isolates by context: a
Goldfish starts without the conversation, and the project tests this on the
hosts with recorded runs. Its documentation cites reports that two hosts add
project instruction files to the reviewer's context. PCD isolates by input.
Each role receives a defined set of files, and the skill refuses a
test-author run whose input directory contains an implementation. PCD does
not test whether the session itself started without an earlier conversation.

A second model is absent from the EGM column. The Claude adapters set each
reviewer to the model of the session, and no rule of the method requires
another model. PCD's dual-LLM mode gives the test-author role to a model from
another family.

"Findings go to the document" ends at approval in EGM. From the implementation
phase on, a defect may be corrected in the code, and the document is updated
when a later comparison shows the difference. In PCD the rule has no end.

---

## 4. Where PCD differs

The differences follow from 3.3. They are listed in the order a skeptical
reader will raise them.

### 4.1 The audited object

Under EGM the code is what ships, and it is what a reviewer, an auditor or a
certifier examines. EGM adds a review in front of it. Goldfish review the
design, and then a human reviews it through a pull request, before any
implementation. EGM does not prescribe what happens to the review of the
code. The guide leaves the implementation phase to the tools and rules of
each team. The FAQ expects the attention of engineers to shift from code
toward the design documents, and the project's index page calls human review
of every change expensive. For contributions to EGM itself the policy is
strict: a human author must have reviewed, tested and verified every changed
line. The FAQ also states the cost of the method: design review adds work and
can become a bottleneck. The README recommends EGM for changes that need
design decisions.

Under PCD the specification is the object under examination and the binary is
its derivative. The human reviews the specification and the reports of the
pipeline, and does not review the generated code. The pipeline checks the
code: the compile gate, the tests written from the specification, the
reviewer findings, and the hash that records which specification the artifact
was derived from. The specification changes when behavior changes and not
otherwise.

EGM adds a human review of the design and leaves the review of the code to
each team. PCD removes the human review of the code by rule, and puts a human
review of the specification and pipeline evidence in its place.

### 4.2 Linking document and code

EGM links document and code by comparison. The drift report is an agent run.
For each design document the agent reads the text, locates the source files
and reads their current state. It lists four kinds of difference: missing
features, incomplete data models, hardcoded shortcuts and architectural
divergence. The design recommends a weekly run and says that the cadence is
not enforced. The document has no version field and no hash, and the status
list of the template ends at "Approved". A footer line records that the
document was drafted with EGM, and the design describes that line as
best-effort provenance. The method prescribes no reference from the code back
to the design it implements.

The EGM document records what PCD does not: the alternatives. Every design
must list the alternatives that were considered and why each was rejected, and
it keeps its open questions.

PCD links document and code by derivation. The translation report lists a
labeled hash of every file consumed as an input. The specification hash is
embedded in source headers, in the binary's version output, in package
metadata and in container labels. The version of a specification must change
whenever its bytes change. Recomputing a hash answers one question: which
specification text did the run that produced this artifact record? That the
code implements that text is for the tests and the review to show.

A project that keeps its code has to find out where code and document have
diverged. A project that regenerates its code has to show what the code was
generated from.

### 4.3 Verification: three kinds of independence

The two methods use different kinds of independence, and each has one that the
other lacks.

EGM uses a fresh context. A fresh context removes the author's unstated
assumptions. The EGM design documents describe that failure: the model that
drafted a plan assumes that what it remembers is on the page. As a review
before implementation, EGM's has more stages and more viewpoints than PCD's.
It has five stages, seven reviewer personas, written standards for four of
the five stages, a paraphrase test, and rules for when a loop has converged.

PCD uses a second model in its dual-LLM mode. A second model addresses errors
that one model family makes consistently, with or without a fresh context.
PCD also uses execution. The readiness question is whether an implementer can
work from the document without deciding anything. In EGM a reviewer answers
it by reading the plan. In PCD a translator answers it by producing the
implementation. The implementation then has to pass the compile gate and the
tests that were written from the EXAMPLES before it existed.

EGM states the limits of a review itself. The README says that a review does
not replace tests or human judgment. The FAQ asks whether a document can be
complete enough for a zero-context agent and answers "sometimes, but not
always". An open issue in the project's tracker proposes to benchmark review
quality against a baseline. Its reason is that a completed workflow does not
show that the review was good.

After approval EGM has no verification step of its own. Tests and code review
belong to the team, and the drift report is a later comparison. PCD's
verification starts at that point. Every run writes the tests before the
implementation. For higher assurance a second model, as test author, writes
an independent suite first. A reviewer and a security reviewer produce
findings that must be resolved in the inputs, and a tiebreak may not declare
a winning implementation. These three steps are defined in the translation
skill. The public repositories contain no report from a run of them yet.

The two methods also check document structure differently. PCD checks the form
of a specification with a linter, and the result is the same on every run.
EGM gives heading hierarchy, required sections, table syntax and links to a
reviewer persona. The EGM repository tests its own documents for links, tables
and YAML headers with unit tests, and a user's design document is checked by
the persona.

### 4.4 Language neutrality

An EGM implementation plan lists repository-relative file paths, function
signatures and schemas, and the readiness standard requires them. The file
list is meant to serve as the list of files that the implementation phase may
write. The simplicity standard asks the design to follow the conventions of
the codebase and to reuse its existing utilities. The document is specific to
one codebase by design, and its language is the language of that codebase.

PCD keeps language-specific realization out of the specification and places it
in hints and deployment templates. In June 2026 one specification text was
translated to C++, Go and Rust, and the three translation reports record the
same specification hash. This only has meaning where the specification is the
durable artifact.

### 4.5 Supply chain and sovereignty

EGM's runtime consists of Markdown files: skills, personas and standards. The
project states that the runtime operates no hosted service, sends no
telemetry, declares no MCP server and installs no hook. Each release provides
two runtime archives, a checksum file, and signed build provenance that a
user can check. The archives have fixed entry order, timestamps and
permissions, so that identical inputs give identical archives on the same
build runtime. The workflow that builds them pins its actions by commit. The
documented installation uses each host's plugin mechanism with the repository
as the source. The catalog files refer to the repository and specify no
release and no digest (section 7).

The recorded hosts are Antigravity CLI, Claude Code, Codex and Muse, with
partial records for two cloud products and a community report for one more
tool. The method does not depend on a provider: the plugin uses the models and
the permissions of its host.

PCD's target is a pipeline an organization can run itself. The target includes
tools shipped as signed packages from a controlled build service, assets
embedded at build time, offline and reproducible builds, and an SPDX software
bill of materials. It also includes models hosted in a chosen jurisdiction or
run without network access. Today a PCD skill checks its own files against a
manifest of hashes at the start of a session. It refuses to build, fetch or
substitute a missing tool.

The EGM runtime is plain text under a permissive license, so it can be
packaged and distributed in the same way. For a customer whose constraint is
sovereignty, the open points are distribution and host support. The method
itself raises none.

### 4.6 Certification: the same goal, a different distribution of evidence

Neither approach makes certification impossible, and the comparison is easy to
overstate in PCD's favor. EGM follows the familiar route: the code is the
evaluated object, and the agent is part of the development environment. EGM
adds evidence to that route which an evaluator can read. There is a design
document with its rationale and rejected alternatives. There are review
reports with findings by severity and the answer to each, and a human
approval by pull request. The costs are those recorded for the other entries
in this directory. The evaluated object changes with every change, and each
change enters code-level review again. Two points are specific to EGM. The
correspondence between design and code depends on a periodic comparison by an
agent. The reviews are model judgments, and the project has not yet published
a measurement of how many defects they find.

Security review is not a separate step in EGM. Security is one item among the
risks. The feature specification standard requires a design to identify its
security implications, and the design critic is told to look for
vulnerabilities the design does not address. The six brainstorm viewpoints
and the nine persona files include no security reviewer, and the five
standards include no security standard. Personas and standards can be added
by path, so a team can supply its own.

Under PCD the human-reviewed object is the specification, and a machine can
check which specification an artifact was derived from. PCD does not assert
that a certifiable system must be built this way. It redistributes the
evidence: it trades code-level review effort for provenance and pipeline
evidence, and it depends on a certifier accepting that trade. That acceptance
has not yet been demonstrated at certification scale, as the whitepaper
records. Both routes are open.

---

## 5. Where EGM leads

EGM is ahead of PCD in the following respects.

- Applicability. EGM works on a repository of any size and age, and on any
  change that needs design decisions. PCD needs a component whose behavior can
  be written down in full, and it has no method for a codebase that is
  maintained and not regenerated.
- Depth of review before implementation. EGM has five stages, seven reviewer
  personas and written standards for four of the stages. PCD has one review
  pass by one reader, ordered by cost, and the linter.
- The comprehension test. A fresh reader paraphrases the document before it
  lists findings. EGM's argument is that a list of undefined terms does not
  find a document whose terms are all defined and whose parts do not connect,
  and that a paraphrase does. The step is on the main branch since 2026-10-28
  and in no release yet. PCD has no equivalent step before translation.
- Convergence rules for a review loop, on the main branch since 2026-10-28. A
  finding that was contested, or accepted without a fix, is closed and is not
  reopened when a later reviewer raises it again. A finding that was fixed is
  reopened when it returns. Two consecutive clean rounds end the loop, and at
  five rounds the loop asks the human how to proceed. PCD's review of
  generated artifacts ends each run with one of three recommendations:
  accept, edit and regenerate, or human triage. Its review of a specification
  before translation is one pass. Neither review keeps closed findings from
  one run to the next.
- A procedure for answering findings. EGM has four classes of answer. A
  contest needs evidence that the reviewer could have read, a comprehension
  gap cannot be contested, and a severity cannot be changed. PCD states where
  a finding must be resolved and leaves the rest to the maintainer.
- Isolation as a tested property. For the hosts with recorded runs, the
  project tests whether a reviewer saw the parent conversation. It treats
  instruction files that a host adds as a separate question and publishes
  what is unverified. PCD assumes that its roles are separate.
- Rationale in the document. Alternatives and the reasons for rejecting them
  are mandatory, and open questions are kept. PCD records realization choices
  in a hints file and the reason for each change in the changelog beside the
  specification. It has no record of the alternatives that were considered.
  The Spec Kitty entry in this directory recorded a related gap.
- A challenge before the specification. The brainstorm challenges the idea
  from six viewpoints before a document is written, and one of them asks
  whether the feature should exist at all. PCD's interview starts after that
  decision.
- Existing codebases. README files written bottom-up give an agent a map of a
  large repository, and the drift report gives the documentation a
  maintenance routine.
- Testing of the skills themselves. A host counts as supported only after
  recorded end-to-end runs. Each skill also has a file of behavior scenarios;
  the repository checks their structure and does not yet run them. PCD's
  skills have a manifest check and dry runs.
- Adoption. The README reports thousands of engineers at Google. On
  2026-11-05, two weeks after the launch post, the repository showed about
  11,900 stars, 829 forks, six code owners and commits from more than twenty
  authors. PCD has one maintainer.

---

## 6. Net positioning

The two methods answer different questions.

- EGM answers: is this design good enough to build, and does the code still
  match it? It applies to any change that needs design decisions, in any
  codebase. The code is retained.
- PCD answers: what artifact can be certified today and regenerated in ten
  years? It applies to components whose behavior can be written down in full.

A team can use both. EGM's own scope note excludes small and clear fixes, and
PCD's scope excludes whatever cannot be specified completely. Most systems
contain work of each kind.

EGM does not regenerate code and does not claim to. On that question PCD's
nearer neighbor is the subject of the first entry in this directory, Codeplain
with its specification language Plain.

### 6.1 A named integration point

EGM's review loop takes its reviewers and its standards as parameters, and a
name that contains a slash is read as a file path. By the skill text, a PCD
specification can therefore be the target of that loop, with a PCD
completeness standard as the rubric. EGM does not have to change for it. The
five-stage review is not open in this way: its standards are fixed, and it
expects the two parts of the design template. The other direction needs no
mechanism. The interview mode of `pcd-spec` accepts existing notes and design
documents as its input, and an approved EGM design is such a document.

| EGM element | PCD counterpart |
|---|---|
| `concepts_brief.md` | Input for the interview mode of `pcd-spec` |
| Goals and non-goals | The list of BEHAVIORs; BEHAVIORs with `Constraint: forbidden` |
| Critical user journeys in Given / When / Then | EXAMPLES |
| Technical plan and alternatives considered | Realization choices go to the decisions hints file; the alternatives have no counterpart today (section 9) |
| Detailed implementation plan with files and signatures | Not written; the translator produces it from the template and the hints |
| Comprehension test (stage 6) | Run on the specification before translation |
| Readiness test against the zero-decision rule (stage 9) | The translation run; its report lists the ambiguities the translator met |
| Drift report | `verify_spec_hash`; the artifact states the specification it was produced from |

The combination requires two adaptations. The detailed implementation plan is
not written, because under PCD the translator produces the file list and the
signatures from the template and the hints. After approval the document is
translated, where EGM implements it once and keeps both. A wrong result leads
back to the specification.

Three limits apply when EGM reviewers read a specification. The personas are
written for design documents. Several of them compare file paths with a
codebase, or object to the level of detail that a specification must have.
Two of the seven are candidates: the comprehension reader and the design
critic. The resolve step can edit its target. For a specification the run has
to end with a report, and accepted findings are applied through the change
mode of `pcd-spec`. EGM reviewers can read every file they can reach, so the
run needs a directory that contains the specification, the template and the
hints, and no other file.

A PCD translation run also produces a record for a question that EGM answers
by review. The translation report lists the ambiguities the translator met
and how it resolved each. The list is the translator's own account, so it is
a lower bound for the decisions the document left open.

Both methods ship as agent skills. EGM's skills are started by the user and
cannot be called by another skill, so the combination is a sequence of steps
that the user runs.

### 6.2 Framing for talks

EGM governs the review; PCD governs the artifact. EGM shows how a design is
reviewed before any code exists, at the scale of a large engineering
organization. PCD asks what happens to that design afterwards. Under EGM it
is kept beside the code and compared with it. Under PCD it is the source, and
the code is generated from it again whenever it changes. The second
arrangement is available where a component can be described in full and has to
pass an audit or outlast its implementation language. For all other work, a
reviewed design and a retained codebase is the better available method.

---

## 7. Open questions and claims to verify

Confirm the following against a working trial and against the article, and not
against the repository read alone.

- Does the review loop accept a `.spec.md` file as its target, with a standard
  given by path? Do the reviewers then produce findings that the PCD linter
  and two passes of the spec-review mode do not? The claim in 6.1 rests on
  the skill text. No trial has been run.
- How does each host resolve the plugin source at installation: the default
  branch at that moment, or a tagged release? The catalog files refer to the
  repository without a release or a digest. Confirm before citing 4.5.
- Which version is meant? This entry describes the main branch at commit
  0dfb5bf. Four rules were added on 2026-10-28, after the tag v0.1.0. They
  are the paraphrase test, the convergence rules, the "not run" rule, and the
  rule that a draft is not graded by its own session. The release archives of
  0.1.0 do not contain them. State the commit when citing these rules.
- The article by Dave Rensin was not read in full for this entry. The
  description of the manual sequence in section 2 and the dating in 3.1 rest
  on the abstract and on secondary coverage. Read the article before citing
  either.
- "Thousands of Google engineers" is the project's own statement. No
  measurement of the effect has been published, and an open issue proposes a
  benchmark. Cite it as the project's statement.
- Which hosts keep project instruction files out of a reviewer's context? The
  test record treats this separately from isolation against the parent
  conversation and reports leaks on two hosts. The answer affects the first
  qualification in 3.4.
- Can a reviewer be given another model than the session on each host? The
  Claude adapters set `model: inherit`. The answer affects the second
  qualification in 3.4.
- Is the drift report run weekly inside Google as recommended, and with what
  result? The design says that the cadence is a recommendation.
- Is EGM in use under any regulated review, and what is presented as evidence
  that the code corresponds to the approved design?
- Stars, forks, commit count and host versions are as read on 2026-11-05 and
  will be out of date quickly. Cite them with the date.

Four statements about PCD need the same care.

- PCD's review, security review and tiebreak are defined in the translation
  skill, and the public repositories contain no report from a run of them.
  Describe them as steps of the method until a run is recorded (4.1, 4.3).
- Three PCD tools were generated from their own specifications. Two of them
  are older than the current text of their specification, and two further
  tools are specified and not yet generated. Say so when citing the last row
  of the table in 3.4.
- The linter in the repository implements 21 of the 25 published rules. The
  other four belong to a library that is specified and not yet generated.
  Give the number that applies when citing the linter.
- The README states that one specification produced implementations in Go,
  Rust and Java. The committed artifacts show Rust and Java for one example,
  Go and Rust for one tool, and C++, Go and Rust for another. Cite what the
  artifacts show (4.4).

---

## 8. Sources

- github.com/google/elephant-goldfish-model, main branch, read on 2026-11-05
  at commit 0dfb5bf (2026-11-04). That is 55 commits after the tag v0.1.0
  (b3a3595); the manifest gives version 0.1.0 at both. Read: README; `docs/`
  (getting started, index, FAQ, hosts, privacy, releases, directory packages,
  and the section on reviewer isolation in the Claude guide); `docs/designs/`
  (architecture, brainstorm, draft, review, synthesizer and drift report in
  full; bootstrap, runtime packaging and additional hosts in part); the seven
  `SKILL.md` files (bootstrap in part); the nine persona files under
  `agents/`; `resources/` (execution instructions and the five document
  standards); the design document template; `TESTING.md`, `CONTRIBUTING.md`,
  `AI_POLICY.md`, `AGENTS.md`; the plugin manifests, the catalog files and
  the release workflow.
- Brian Suchy, Jim Quinn, Elise Ewing, Oliver Newth and Dave Rensin,
  "Introducing the Elephant-Goldfish Model plugin: scaling design-driven AI
  software development", Google Open Source Blog, 2026-10-22.
- Dave Rensin, "Elephants, Goldfish, and the New Golden Age of Software
  Engineering", Medium, 2026. Read through its abstract on the Google Research
  publications page and through the report in I Programmer of 2026-06-12, not
  in full.
- egm.dev, the project's website, read on 2026-11-05.
- The project's issue tracker and discussions, read on 2026-11-05: the issue
  that proposes a benchmark of review quality, the issue that proposes a
  whiteboard check, and the announcement of version 0.1.0. Used for the state
  of verification and for the release facts.
- Help Net Security, 2026-10-26, and The Agent Report, 2026-10-27. Used for
  dates and for the observation that security is not a separate review step.
- Birgitta Böckeler, "Understanding Spec-Driven-Development: Kiro, spec-kit,
  and Tessl", martinfowler.com, 2025-10-15. Used for the three levels in 3.3.
- github.com/mge1512/pcd at commit 8d04918 and
  github.com/mge1512/zypper-declarative, read on 2026-11-05. Used for the
  statements about PCD's own artifacts in 4.3, 4.4 and section 7.

---

## 9. Follow-up actions (non-normative)

- Write a completeness standard for PCD specifications in the form the EGM
  review loop accepts. Run the loop on two existing specifications, with a
  report as its only output. Compare its findings with those of the linter
  and of two passes of the spec-review mode.
- After that trial, decide on three candidates. The first is a paraphrase
  step in the spec-review mode. The second is a declared fresh session for
  every mode that works from a specification. The third is a file that keeps
  closed findings from one review run to the next.
- Decide whether PCD wants a rationale record beside the specification,
  outside the hashed bytes, for the alternatives considered and the open
  questions.
- Record first runs of the review, security review and tiebreak, and commit
  the reports.
- Offer the EGM project the result of the trial as input to its benchmark
  proposal.
- Put the question about the plugin source in section 7 to the project
  directly.
- Use the sentence from 3.3 and the paragraph in 6.2 in talks.

---

## Changelog

- 2026.11.05.01 - Initial draft. Sources: the main branch of the
  `elephant-goldfish-model` repository at commit 0dfb5bf, the launch post on
  the Google Open Source Blog, and the project's website, issue tracker and
  discussions, all read on 2026-11-05. The article by Dave Rensin was read
  through its abstract and secondary coverage. Statements about PCD's own
  artifacts were compared with the `pcd` and `zypper-declarative`
  repositories on the same day. Written for public use: no internal project
  names appear. Certification wording follows the Spec Kitty
  and HDAE entries: both routes are open and distribute the evidence
  differently. Mirror updates to the whitepaper (A.12, Comparative Summary,
  References) and to the technical reference (section 18, References) are not
  yet applied.
