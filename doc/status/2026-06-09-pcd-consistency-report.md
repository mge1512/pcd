# PCD Consistency Check - Findings and Tasks

Version:  2026.06.10.01
Scope:    consistency-check-for-fable-5.zip (148 files; pcd/, kvm-manager/, zypper-declarative/; snapshot 2026-06-09)
Method:   read of normative core and project artefacts; SHA256 recomputation of all pinned translation inputs against TRANSLATION_REPORT claims; cross-grep of every version, schema, and convention marker
Authority baseline (as agreed): prompts/prompt.md + templates/*.template.md + tools/shared/spec/lint-rules.md are normative; doc/ is descriptive
Hash notation: first 16 hex characters throughout

---

## 1. Verdicts

**Q1 - Is the use of PCD in kvm-manager and zypper-declarative up-to-date and consistent?**
Largely yes at process level: both projects pin spec, prompt, template, and hints; all four TRANSLATION_REPORTs carry the full v0.4.3 labelled per-input provenance; the pinned `cli-tool.template.md` and the Go/Rust milestones hints are byte-identical to the framework masters; reports contain Tests-First compliance, continuity tables, Test Refinements, and per-EXAMPLE confidence. Three classes of inconsistency remain: (a) kvm-manager operates outside three template constraints (linking, runtime deps, env vars) that the template has not yet relaxed despite the spec's own TEMPLATE-FEEDBACK requesting it; (b) the two projects use two different feedback and changelog conventions; (c) one live provenance mismatch in kvm-manager (decisions hints) and stale pinned prompts in both projects.

**Q2 - Has PCD absorbed the lessons from kvm-manager and zypper-declarative?**
This is the largest gap. The projects encode at least 19 generalizable lessons; roughly half are absent from every framework artefact. Most striking: the "no hollow tests" and "clean up after yourself" rules exist only in zypper-declarative's pinned copy of `cli-tool.cpp.milestones.hints.md` - the framework master is newer but lacks both sections, and zypper's decisions hints reference a "general rule in the milestones scaffold" that the master no longer states. The hermetic-isolation directive (kvm-manager), the VERSION single-source and `make dist` rules, the RPM directory-ownership procedure, and the `go vet` gate are likewise project-only. Section 4 contains the complete harvest table.

**Q3 - Is PCD internally consistent?**
Mostly, with a short list of genuine defects: the embedded-hash semantics contradict themselves between the composition section and the hash-embedding section of the prompt; the cli-tool template disagrees with itself about its own language set and carries a stale version literal; `lint-rules.md` is invalid under its own RULE-01 and RULE-03; the `DeploymentTemplate` enum is missing four shipping templates; two templates violate RULE-14; the universal prompt contains pcd-lint-specific instructions, contradicting the two-layer design the whitepaper itself states; and RULE-02c rejects kvm-manager's (house-convention) version string. The whitepaper lags the normative core by three minor versions and has a duplicated appendix number.

---

## 2. Provenance verification (evidence)

### 2.1 kvm-manager (report 2026-06-09, llm claude-opus-4-8, incremental per UPGRADE brief)

| Input | Report claim | On disk (spec/) | Verdict |
|---|---|---|---|
| kvm-manager.spec.md | 12c83b5052de6715 | 12c83b5052de6715 | match |
| kvm-manager.go.decisions.hints.md | 9abb33ecff46a8fb | c7294de30a803c20 | MISMATCH |
| cli-tool.go.milestones.hints.md | c210c2f1404a72fc | c210c2f1404a72fc | match |
| cli-tool.template.md | c8447ba8f1e63f36 | c8447ba8f1e63f36 | match (== master) |
| UPGRADE.md | 3a2ed2bf71990ff2 | 3a2ed2bf71990ff2 | match (label `Upgrade-Brief-SHA256` is a project invention, not yet canonical) |
| prompt.md | not recorded (no label exists) | 1172916a43585141 | gap |
| test-isolation.directive.md | not recorded | 18ce15c7c81a68c8 | gap |

The decisions-hints mismatch means the tuple of the *current* build is not verifiable from the tree, although the report (12:51) postdates the file timestamp (11:24). Recorded only; see Section 7.

### 2.2 zypper-declarative (reports: go and rs 2026-06-02, cpp 2026-06-03)

| Input | cpp claim | go claim | rs claim | On disk | Verdict |
|---|---|---|---|---|---|
| zypper-declarative.spec.md | aafbb3158415b5c8 | 51284526723dc923 | 51284526723dc923 | 20bbef896f83de4f | all mismatch - expected: spec advanced to 0.6.10, UPGRADE.md brief pending |
| <lang> decisions hints | fd815bece1004a16 | c31330c25e7b90ad | b00395c6d6cc6af4 | fd815bec / 9f99205c / 1171ccc4 | cpp match; go and rs edited after translation (pending pass) |
| cli-tool.<lang>.milestones.hints.md | c6e80c18bbc4a726 | c210c2f1404a72fc | 811af43339b25e4f | identical | all match |
| cli-tool.template.md | c8447ba8f1e63f36 | same | same | c8447ba8f1e63f36 | all match (== master) |
| prompt.md | not recorded | not recorded | not recorded | 087d5db88c56600d | gap |

Note: the cpp artefact was translated from a different spec state (aafbb315, post-0.6.9-clarifications) than go and rs (51284526). The three language implementations have never been built from one spec hash; the pending 0.6.10 pass will align them.

### 2.3 Framework master vs project pins

| File | Master | kvm-manager pin | zypper pin | Verdict |
|---|---|---|---|---|
| cli-tool.template.md | c8447ba8f1e63f36 | identical | identical | clean |
| cli-tool.go.milestones.hints.md | c210c2f1404a72fc | identical | identical | clean |
| cli-tool.rs.milestones.hints.md | 811af43339b25e4f | - | identical | clean |
| cli-tool.cpp.milestones.hints.md | c3ab7ae9d8f680bf | - | c6e80c18bbc4a726 | MASTER REGRESSION: master (newer mtime) lacks two whole sections present in the project pin; see F-C1/F-C2 |
| prompt.md | c059dca7f7aa4dbc | 1172916a43585141 | 087d5db88c56600d | both pins predate the v0.4.3 provenance blocks; runs complied only because the template carries the contract |

---

## 3. Findings - normative core (Q3)

Classes: **Defect** (artefact contradicts itself or a normative rule), **Drift** (out-of-sync copies/statements), **Gap** (practice or lesson not codified), **Recommendation**.

**F-A1 (Defect) - Embedded-hash semantics contradict themselves.**
`prompt.md` "Spec Composition / Hash computation" (lines 121-131) and "Reporting" (135-144): the hash embedded in all artefacts is the SHA256 of the *merged* spec text. `prompt.md` "Spec hash embedding" (678-704): "Compute `sha256sum <specname>.md`" (682) and "the SHA256 of the spec file *as provided as input* - not of any transformed or post-processed version" (698-699). For any spec with `Includes:` these demand different hashes. The same "as provided" wording appears in `cli-tool.template.md:588` and the "Computed once before any output" row at :458. `lint-rules.md` RULE-18 (:660) also says "Run: sha256sum <specname>.md".

**F-A2 (Defect) - cli-tool template disagrees with itself about its language set.**
TYPES (template:39): `Language := Go | Rust | C | CPP | CSharp`. TEMPLATE-TABLE adds `Java` (:141) and `Lean4` (:142) as supported LANGUAGE rows, and every per-language table (source layout :500-501, build :518-519, Containerfile :534-535, syntax check :744-745, compile gate, manifest rules) covers them. PRECONDITIONS (:194): "LANGUAGE value in resolved output must be one of: Go, Rust, C, C++, C#". LANGUAGE-ALTERNATIVES rows (:143-146) omit Java/Lean4, and POSTCONDITIONS (:205) binds resolution to LANGUAGE-ALTERNATIVES or the default.

**F-A3 (Defect) - Stale embedded version literal.**
`cli-tool.template.md:613` "Current version: 0.3.13" inside DEPLOYMENT/Versioning; META (:8) declares `Version: 0.3.29`.

**F-A4 (Defect) - lint-rules.md is invalid under its own rules.**
META declares `Deployment: none` (:4); the `DeploymentTemplate` enum (:63-69) contains no `none`, so RULE-03 fires "Unknown deployment template". The file lacks `## PRECONDITIONS`, `## POSTCONDITIONS`, `## EXAMPLES`, so RULE-01 fires three errors. The composition-target convention (`Deployment: none`) is described only in descriptive docs (spec-composition.md:95, user-guide.md:901), never in a normative artefact.

**F-A5 (Defect) - DeploymentTemplate enum out of sync with templates/.**
The enum lacks four shipping templates: `kubectl-style-cli`, `spack-package`, `cockpit-module`, `abap-report` (all exist under pcd/templates/). A spec declaring any of them is rejected by RULE-03. Conversely the enum lists `wasm`, `ebpf`, `enterprise-software`, `academic` for which no template file exists (planned entries, unannotated), plus the template-less modes `enhance-existing` and `manual` (legitimate).

**F-A6 (Defect) - Project-specific content in the universal prompt.**
`prompt.md:358-380` carries a block "For `pcd-lint` specifically:" enumerating pcd-lint fixture sections and rule numbers. This contradicts the two-layer architecture the whitepaper states as design rationale (A.13, whitepaper:2037-2049: "the generic prompt contains only universal principles").

**F-A7 (Defect) - RULE-02c rejects the house date-version scheme.**
RULE-02c (lint-rules:227-233) requires `^[0-9]+\.[0-9]+\.[0-9]+$`; the TEMPLATE-TABLE VERSION row (template:132) likewise demands MAJOR.MINOR.PATCH. `kvm-manager.spec.md` META declares `Version: 2026.06.09.03` (the maintainer's YYYY.MM.DD.VV convention) and therefore fails lint. zypper-declarative (0.6.10) conforms.

**F-A8 (Defect) - RULE-14 violations inside pcd/templates/.**
`abap-report.template.md` (Spec-Schema 0.4.0, newest template): no `## EXECUTION`, no `EXECUTION: none` in META. `spack-package.template.md`: same state (whitepaper A.11 even notes "no EXECUTION phase" for it, so `EXECUTION: none` is the intended declaration). `cockpit-module.template.md`: has EXECUTION but zero TRANSLATION_REPORT mentions - no report contract at all.

**F-A9 (Drift) - Spec-Schema scatter across templates.**
0.3.19 (project-manifest, cloud-native), 0.3.20 (verified-library, python-tool, gui-tool, backend-service, library-c-abi, mcp-server), 0.3.21 (kubectl-style-cli), 0.3.22 (spack-package, cockpit-module), 0.4.0 (cli-tool, abap-report). If Spec-Schema in a template means "schema this file was written against", the scatter is legal but signals ten templates never revalidated against 0.4.0. Three of the lagging templates also lack the v0.4.3 provenance block (abap-report, cockpit-module, spack-package); the other ten carry it.

**F-A10 (Gap) - Provenance label set incomplete relative to practice.**
The canonical required lines (prompt:759-779, template:460-479, tech-ref section 12) are Spec (merged/host), Decisions-Hints, Milestones-Hints, Template, plus generic Style/Library lines. Practice already exceeds this: kvm's report records `Upgrade-Brief-SHA256` (report:10, ad-hoc label); `test-isolation.directive.md` was consumed but is recorded under no label; the pinned `prompt.md` is an input in both projects yet no `Prompt-SHA256` label exists and no report records it. The reproducibility tuple as defined - (spec, resolved language, hints/template set) - silently excludes the prompt, briefs, and directives.

**F-A11 (Gap) - File-naming convention codified nowhere.**
Practice: all four live specs use `.spec.md`; templates `.template.md`; hints `.hints.md`; kvm-manager additionally introduced `.directive.md`. Normative wording: prompt:10 and template EXECUTION input list (:629) say `<spec-name>.md`. The pinned project prompts instantiate inconsistently: kvm-manager correctly as `kvm-manager.spec.md`, zypper as `zypper-declarative.md` - a filename that does not exist. The shared fragment is named `lint-rules.md` (bare .md), in tension with the ".spec.md always" practice. Also `examples/account-transfer/account-transfer.md` uses bare `.md`.

**F-A12 (Gap) - Black-box methodology binds only the test-author flow.**
The "Test methodology" block (prompt:271-294: real binary via subprocess, never import internals, never simulate or mock the unit under test) sits inside the test-author flow. The translator's own suite (step 2, prompt:172-179) is not explicitly bound. zypper's decisions hints treat black-box as generic for both suites (`[pcd]` tags: cpp:441, go:343, rs:332).

---

## 4. Findings - projects vs framework (Q1)

**F-B1 (Defect, conformance conflict) - kvm-manager operates outside three cli-tool constraints.**
(a) Linking: template forbids dynamic for Go ("Not permitted for Go or Rust", :148; PRECONDITION :197); kvm's Go decisions hints select cgo + dynamic libsqlite3 as the default and state "Static linking is not required (this relaxes the earlier default)" - a hints file relaxing a template-required constraint inverts the authority order. (b) RUNTIME-DEPS `none` required (:161) vs the runtime dependency on the platform's libsqlite3. (c) CONFIG-ENV-VARS forbidden (:181) vs the `KVM_MANAGER_HOST` behaviour switch (spec DELTA 24). The spec's own TEMPLATE-FEEDBACK item 4 (spec:2243-2250) requests exactly the linking change with the supply-chain rationale (packaged, signed, centrally patched library beats a vendored static blob). Resolution direction: change the template (T-08..T-11), after which kvm-manager conforms.

**F-B2 (Drift) - Two feedback conventions.**
kvm-manager: `## TEMPLATE-FEEDBACK` in-spec (spec:2219-2251, four open items). zypper-declarative: `[pcd]` tags inside decisions hints (legend at cpp:10, go:8, rs:10) plus CHANGELOG narration. TEMPLATE-FEEDBACK is documented only in the user-guide and reverse-prompt; prompt.md, templates, and lint-rules do not know it. No harvesting step exists anywhere (CONTRIBUTING has none), which is precisely how F-C1..C16 accumulated.

**F-B3 (Drift) - Two changelog conventions, one of them internally mixed.**
zypper keeps an external `spec/CHANGELOG.md` with explicit rationale (CHANGELOG:3-7: beside the spec, not read by the translator, not covered by the spec hash) - but the file mixes two formats: dash entries 0.6.5 down to 0.1.0 at the top (lines 10-130), then `## Version 0.6.10` ... `0.6.6` sections below (line 233 onward), so the newest entries sit in the middle. kvm-manager has no changelog at all; history lives in UPGRADE briefs and DELTA "(DELIVERED)" markers. The framework codifies neither convention.

**F-B4 (Drift, recorded only) - kvm-manager decisions-hints provenance mismatch.**
See table 2.1. Either the report recorded a wrong hash or the file changed post-run with a preserved-looking timestamp. No task (implementations out of scope); resolve at the next translation; checking the file's git history against the run log will identify which.

**F-B5 / F-B6 (Drift, expected and properly briefed) - zypper 0.6.10 pending.**
Spec ahead of all three translations; go/rs decisions hints edited after their runs; project docs pinned to 0.6.9 (README:22,41; architecture:7,565); UPGRADE.md brief ready with `recommendation: incremental`. This is the intended PCD state between passes. The cross-language spec-hash split (2.2 note) is the one item worth calling out in the next pass's report.

**F-B7 (Drift) - Pinned prompts stale, refresh undisciplined.**
Both project prompts predate the v0.4.3 provenance blocks (table 2.3); kvm's was touched on 2026-06-09 (run header `Language: GO`, `Target Directory: ...`) without refreshing the body. Because no report records a prompt hash (F-A10), nothing documents which prompt text a run actually used.

**F-B8 (Drift) - MILESTONE Status fields unmaintained in kvm-manager.**
All eight MILESTONE sections are `Status: pending` (spec:2255-2445); none active, none released, although work through the 0.5.0-era milestones has shipped. Actual progress is tracked via DELTA "(DELIVERED)" markers - two mechanisms, one dead. TYPES in lint-rules says "Status is managed by the agent pipeline" (:88); the pipeline does not do so. Runs remained valid (no active milestone -> full-spec translation per prompt:558), but the pipeline state is fiction.

**Positive verdicts worth stating:** `.spec.md` naming in all live specs; full input pinning in both projects; labelled per-file provenance in all four reports (never collapsed into a combined hash); template pins byte-identical to master; continuity-check tables against on-disk truth; Test Refinements tables with spec-referencing rationales; reviewer/tiebreaker/security-reviewer prompts all enforce "never a code patch" (reviewer:22-25, 165, 203, 244); the lint-rules extraction and Includes migration for pcd-lint (0.4.1) and mcp-server-pcd (0.4.0) is complete at spec level with resolving relative paths and no duplicated rule bodies.

---

## 5. Findings - lessons not yet absorbed (Q2, feeds Goal 1)

Harvest table. "Coverage" = where the lesson currently lives in pcd/ masters.

| # | Lesson | Evidence | Coverage in pcd/ | Target artefact |
|---|---|---|---|---|
| C1 | No hollow tests: a test for an EXAMPLE asserts that EXAMPLE's THEN (stdout/stderr content, exit code, state); `exit_code == 0`-only is forbidden "green theater"; genuinely untestable behaviours are explicit SKIPs counted separately | zypper pinned cli-tool.cpp.milestones.hints.md, section after the /tmp/err passage (~:267-292); referenced as "the general rule" by cpp decisions hints:454 | ABSENT (master cpp hints lost it; nowhere else) | restore to master cpp hints (T-01); generalize into prompt.md (T-02) |
| C2 | Clean-run discipline, both roles: per-test temp fixtures deleted incl. failure path; translator leaves no scratch/intermediates; build output excluded from dist; `git status` clean; re-runnable from identical state | zypper pinned cpp milestones (~:294-318) | ABSENT | restore (T-01); generalize (T-02) |
| C3 | Hermetic test isolation: sandbox HOME/config root; every side effect confined to per-test temp dir; absolute paths in fixtures/EXAMPLEs are illustrative, never write targets - rewrite into the sandbox; assertions computed from sandbox paths; defence-in-depth assert in any file-writing helper; acceptance test (suite passes for an unprivileged user with no access to any fixture path). Motivated by a generated helper truncating a real disk image | kvm-manager/spec/test-isolation.directive.md (whole file) | ABSENT | prompt.md test methodology (T-02); cross-ref in templates |
| C4 | VERSION single-source: one-line top-level VERSION file feeds RPM `Version:`, tarball name/dir, and binary version output - guaranteed identical | `[pcd]` zypper cpp:424, go:327, rs:318 | ABSENT | cli-tool + kubectl templates (T-03) |
| C5 | `make dist` is a required Makefile target: tarball `<n>-X.Y.Z.tar.gz` with single top dir `<n>-X.Y.Z/`, excludes build/ and VCS, so rpmbuild %setup works; Source0 references it | `[pcd]` cpp:430, go:332, rs:324; zypper pinned cpp milestones:78 | template requires Source0 = local tarball (:554) but the Makefile row (:448) lacks `dist` - the tarball has no defined producer | templates (T-03) |
| C6 | RPM directory ownership: OBS `50-check-filelist` fails post-build on unowned dirs (invisible to local rpmbuild); decision procedure: `rpm -qf <dir>` on target -> owned by a dependency: `Requires:` it, do NOT also %dir (duplicate-ownership rpmlint complaint); unowned: `%dir` every level | zypper cpp decisions hints:409-423 | ABSENT | templates RPM section (T-04) |
| C7 | SLE devel-package naming: BuildRequires names `<name>-devel` (jsoncpp-devel, yaml-cpp-devel); `lib<name><soname>` packages are runtime libs and never belong in BuildRequires | cpp:396-406 | ABSENT | master cpp milestones hints (T-07) |
| C8 | No root at build time: fetch/vendor/build as unprivileged user (OBS reality); Go module cache in build user's HOME; Rust `cargo vendor` + offline | go:22, rs:22 | ABSENT | prompt universal principles (T-05) |
| C9 | Feature-detect external C/C++ library API versions in the build system (pkg-config module version -> compile definition -> `#if defined`), record detected version and compiled branch in the report; never guess header macros | cpp:170-177 (snapper case) | ABSENT in generic form | master cpp milestones hints (T-07) |
| C10 | `go vet` belongs in the compile gate: a printf verb/type mismatch rendered `%!s(int=201)` into MAC/UUID templates; qemu rejected the VM after the tool had already reported success; `go vet` catches the class | kvm UPGRADE.md (02 -> 03) items 1-2; spec DELTA 28 | gate is `go build` only (template:790); vet exists only in the test-author syntax check (:740) | templates compile gate + master go milestones hints (T-06) |
| C11 | Long-running EXAMPLE policy: examples driving O(installed-packages) operations can take minutes; the harness must allow a generous per-test timeout and must not kill a still-running long example | zypper CHANGELOG 0.6.9(b) | ABSENT | prompt.md test methodology (T-02e) |
| C12 | Pin physical schema names in INTERFACES; test fixtures must derive from the same canonical schema as production (root cause of "no such column: d.os_version": underspecified column, translator invented a fixture schema) | kvm DELTA 21 | ABSENT | user-guide spec-authoring guidance (T-16) |
| C13 | PERSISTENCE is undeclared in cli-tool (SQLite-backed tools exist) | kvm TEMPLATE-FEEDBACK 1 | ABSENT | template key (T-08) |
| C14 | CLI-ARG-STYLE hybrid: positional subcommands + key=value (git/systemctl pattern) deserves a named style | kvm TEMPLATE-FEEDBACK 2 | partially (bare-words supported row, :163) | template (T-10) |
| C15 | Per-host configuration as a declarable property | kvm TEMPLATE-FEEDBACK 3 | ABSENT | decision D-4, then template (T-12) |
| C16 | Linking mode is a per-component decision; dynamic against a platform-packaged, signed library is the supply-chain-friendly choice for FFI persistence | kvm TEMPLATE-FEEDBACK 4; DELTA 26; go decisions hints section 2 | template forbids (F-B1) | template (T-09) |
| C17 | N-language translation as spec-ambiguity probe: three-way Go/C++/Rust comparisons settled the config_files emission rules and exposed a real Go-vs-C++ divergence (alternatives symlinks) | zypper CHANGELOG 0.6.3-0.6.6 | absent from whitepaper methodology | whitepaper A.18/A.19 (T-14) |
| C18 | Model-selection lessons (session-sourced, in scope per maintainer): LLM scale beats domain fine-tuning (20B-30B local models failed on base capability; a 120B EU-hosted model succeeded without fine-tuning); prompt engineering beats model swapping (a methodology directive fixed a simulator-vs-real-binary default without changing the model) | working sessions; the directive itself is C3's source | absent | whitepaper A.18 (T-14) |
| C19 | UPGRADE.md KIT change brief: header `<spec> <from> -> <to>`; fields recommendation / structural impact / blast radius; per-behaviour directives; consumed as a translator input for incremental passes | both projects practice it; kvm report labels it Upgrade-Brief-SHA256; change-impact.md produces assessments but defines no artefact | uncodified | prompt + change-impact + user-guide (T-21) |
| C20 | Reviewer never proposes code patches; all feedback targets spec, hints, template, or prompt | reviewer.md:22-25, 165, 203, 244 | PRESENT | none - already absorbed |
| C21 | Black-box methodology (real binary, never simulator) | prompt.md:271-294 | PRESENT for test-author only | hoist role-neutral (T-02a, F-A12) |

---

## 6. Findings - documentation alignment (Goal 2)

**F-D1 (Defect)** - Duplicate appendix number: two `## A.2` sections (whitepaper:559 "Specification Format Details" and :876 "Complete Example"). All later appendix references are suspect until renumbered.

**F-D2 (Drift)** - whitepaper header `Version: 0.4.0` (:5) vs changelog rows through 0.4.3 (:3563); technical-reference header `Version: 0.4.0` (:4) although its section 12 was updated in 0.4.3. No doc-version policy exists.

**F-D3 (Gap)** - whitepaper body predates 0.4.1-0.4.3 content: zero coverage of the Translation-Inputs provenance contract and the corrected reproducibility tuple, of UPGRADE briefs, of TEMPLATE-FEEDBACK, of hermetic/hollow-test rules; MILESTONE pipeline appears twice, ROLE.md once. A.13 embeds a full copy of the translator prompt that will drift from prompts/prompt.md.

**F-D4 (Drift)** - A.11 deployment-template table omits abap-report, cockpit-module, and kubectl-style-cli; the cli-tool row's alternatives ("Rust, C, C++, C#") match the template's PRECONDITIONS but not its TEMPLATE-TABLE (ties F-A2); "Single static binary preferred" needs rewording if D-2 is accepted.

**F-D5 (Drift)** - user-guide claims "Current Spec-Schema: 0.3.21" (:1369) and uses 0.3.21 in its skeletons (:443, :501 - the latter even commented "use current").

**F-D6 (Drift)** - interview-prompt.md (:272, :438) and reverse-prompt.md (:177, :365) hardcode Spec-Schema 0.3.21 in output skeletons and checklists; both are silent about Includes/composition.

**F-D7 (Drift)** - example set stale: calc-interest.spec.md declares Spec-Schema 0.3.22 and `Author: Unknown`; examples/account-transfer/account-transfer.md uses bare `.md`.

**F-D8 (minor)** - doc/presentation/pcd-intro-90-appendix.md is an empty (0-byte) file.

---

## 7. Recorded, explicitly out of scope (no tasks)

Per the agreed constraint, implementation drift is recorded only:

1. **pcd-lint**: spec at 0.4.1 (2026-06-09) vs last translation 2026-05-18; embedded Spec-SHA256 constants stale; re-translation pending (known item).
2. **mcp-server-pcd**: embedded store assets (code/internal/store/assets/, dated 2026-04-07..05-18) older than the 2026-06-01 masters (e.g. its cli-tool.template.md asset is the pre-0.4.x 24 KB version vs the 50 KB master); the spec is now 0.4.0 with Includes; re-translation refreshes the assets.
3. **kvm-manager decisions-hints provenance mismatch** (F-B4): investigate file history vs run log at the next pass.
4. **zypper-declarative 0.6.10 pass** per UPGRADE.md, which also unifies the three languages on one spec hash and refreshes the 0.6.9 references in README and architecture doc.
5. **Project-side housekeeping after T-22 sets the convention**: unify zypper CHANGELOG's two formats; create kvm-manager CHANGELOG.md.

---

## 8. Tasks

Each task is self-contained and executable by Claude Opus or Sonnet against the repository: target files, action, acceptance criteria, finding references. Tasks marked **[D-n]** require the listed decision first (Section 9). No diffs here by request; every task names its anchors precisely enough to be executed without this report's author.

### 8.1 Goal 1 - PCD templates and hints absorb the project findings

**T-01 - Restore the two lost sections to the master cpp milestones hints.**
Target: `pcd/hints/cli-tool.cpp.milestones.hints.md`.
Action: copy verbatim from `zypper-declarative/spec/cli-tool.cpp.milestones.hints.md` (sha256 c6e80c18...) the sections "Every test must assert the EXAMPLE's actual outcome (no hollow tests)" and "## Clean up after yourself (test-author and translator)", inserting at the same position (after the /tmp/err passage, before "## Output path construction").
Acceptance: `diff -u` between the project pin and the master is empty (today the pin is exactly master plus these two sections).
Refs: F-C1, F-C2, table 2.3.

**T-02 - Consolidate test discipline in prompts/prompt.md (one editing pass).**
Target: `pcd/prompts/prompt.md`.
Action: (a) hoist the "Test methodology" black-box rules (currently inside test-author step 3, :271-294) into a role-neutral section that binds the translator's own suite and the test-author suite alike; (b) add the no-hollow-tests rule in generic wording (assert the EXAMPLE's THEN; exit-0-only assertions forbidden; genuinely untestable behaviours are explicit SKIPs and the reports count skipped separately from passed); (c) add "Hermetic test isolation" incorporating the five rules and the acceptance test from `kvm-manager/spec/test-isolation.directive.md`, including the rationale that absolute paths in fixtures and EXAMPLEs are illustrative and never write targets; (d) add the clean-run discipline for both roles (per-test temp fixtures removed including on failure; translator leaves no scratch or intermediates; second identical run starts from the same clean state); (e) add the long-example timeout rule (harness allows generous per-test timeouts for EXAMPLEs driving O(system) operations and never kills a still-running long example; specs may annotate expected-long EXAMPLEs).
Acceptance: all five disciplines stated once, role-neutrally; the test-author flow references rather than restates them; the kvm directive's content is fully subsumed (project may then mark its directive superseded - separate, project-side).
Refs: F-A12, F-C1, F-C2, F-C3, F-C11, F-C21.

**T-03 - VERSION single-source and `make dist` in the CLI templates.**
Target: `pcd/templates/cli-tool.template.md`, `pcd/templates/kubectl-style-cli.template.md`.
Action: extend the Makefile deliverable row (cli-tool :448) to required targets `build, test, install, clean, man, dist`; add a Deliverable Content rule "Version single source": a one-line top-level `VERSION` file is the sole version authority - the RPM `Version:` reads it, `make dist` derives the tarball name and single top-level directory `<n>-X.Y.Z/` from it, and the build embeds it so binary version output, RPM `Version:`, and tarball directory are guaranteed identical; define dist content (single top dir, excludes build artefacts and VCS directories, satisfies rpmbuild's default %setup); couple the RPM rule "`Source0:` must reference a local tarball" (:554) to the dist target as its producer.
Acceptance: both templates name `dist` as required, define the VERSION file, and Source0 references the dist output.
Refs: F-C4, F-C5.

**T-04 - RPM directory-ownership rule in the CLI templates.**
Target: same two templates, RPM spec content section (cli-tool :543-554).
Action: add: every directory the package installs into must be owned - by this package (`%dir` entries in `%files`) or by a package in `Requires:`; include the decision procedure (`rpm -qf <dir>` on the target; owned by a dependency -> add the `Requires:`, do not also `%dir` it, duplicate ownership draws an rpmlint complaint; unowned -> `%dir` every directory level); note that OBS post-build checks (`50-check-filelist`) run after a successful rpmbuild and are invisible to a local build.
Acceptance: rule and procedure present in both templates.
Refs: F-C6.

**T-05 - "No root at build time" universal principle.**
Target: `pcd/prompts/prompt.md`, Universal principles.
Action: add: every fetch, vendor, dependency-resolution, and build step runs as an unprivileged user (OBS builders and CI are unprivileged); per-language notes: Go module cache under the build user's HOME; Rust via `cargo vendor` and offline build; no `sudo` in any Makefile target.
Acceptance: principle present with the per-language notes.
Refs: F-C8.

**T-06 - `go vet` into the compile gate; printf pitfall into the Go hints.**
Target: compile gate Step 2 Go rows in `cli-tool.template.md` (:790) and `kubectl-style-cli.template.md`; `pcd/hints/cli-tool.go.milestones.hints.md`.
Action: gate row becomes `go build ./...` then `go vet ./...`, both must pass. Hints: add the format-verb/type-mismatch pitfall with the concrete symptom (an integer passed through `%s` renders `%!s(int=201)`; when filling string templates from numeric values, render the value to the intended string form first - e.g. four uppercase hex digits split into two-character halves - and validate assembled identifiers such as MACs and UUIDs against their grammars before exec; `go vet` catches the verb mismatch at the gate).
Acceptance: vet in both gates; pitfall in the hints with the symptom string.
Refs: F-C10.

**T-07 - Two generic C/C++ patterns into the master cpp milestones hints.**
Target: `pcd/hints/cli-tool.cpp.milestones.hints.md` (after T-01).
Action: (a) build-system feature detection for external library API versions: detect via the pkg-config module version in CMake/meson, define a project compile definition, branch in code on that macro; record the detected version and the compiled branch in TRANSLATION_REPORT.md; never guess header macros; (b) SLE devel-package naming: `BuildRequires:` names `<name>-devel` packages; `lib<name><soname>` packages are runtime shared libraries and never belong in BuildRequires.
Acceptance: both patterns present, generic (no zypper/snapper specifics; the project file keeps those).
Refs: F-C7, F-C9.

**T-08 - PERSISTENCE key in the cli-tool template.**
Target: `cli-tool.template.md` TEMPLATE-TABLE.
Action: add `PERSISTENCE | none | default` and `PERSISTENCE | sqlite-local | supported` (likewise json-local, toml-local); notes: declaring sqlite-local interacts with the linking rows (see BINARY-TYPE) and packaging (a `Requires:` on the platform SQLite when dynamically linked).
Acceptance: rows present with the cross-references.
Refs: F-C13, kvm TEMPLATE-FEEDBACK 1.

**T-09 [D-2] - Linking-mode relaxation.**
Target: `cli-tool.template.md` BINARY-TYPE rows (:147-148), RUNTIME-DEPS (:161), PRECONDITIONS (:196-197), INVARIANTS (:220-221); whitepaper A.11 cli-tool row (with T-15).
Action (on D-2 accept): static remains the default for every language; dynamic becomes supported for any language when a declared PERSISTENCE/FFI capability binds to a platform-packaged, signed shared library; RUNTIME-DEPS reworded to "none beyond platform-packaged shared libraries explicitly declared in packaging (Requires/Depends)"; PRECONDITIONS and INVARIANTS updated to match; A.11 reworded ("single static binary default; dynamic link against packaged platform libraries permitted for declared persistence/FFI bindings").
Acceptance: kvm-manager's documented choice (cgo + dynamic libsqlite3, `Requires: libsqlite3-0`) is expressible without violating any required or forbidden row; specs declaring nothing keep today's behaviour.
Refs: F-B1, F-C16, kvm TEMPLATE-FEEDBACK 4.

**T-10 - Name the hybrid CLI argument style.**
Target: `cli-tool.template.md` CLI-ARG-STYLE rows (:162-163).
Action: add `CLI-ARG-STYLE | subcommand | supported`: positional subcommands (git/systemctl pattern) combined with key=value options; clarify the bare-words row covers single-word commands; POSIX `--flag` remains forbidden for new options.
Acceptance: kvm-manager's surface (subcommands + key=value) and zypper's (verbs + key=value) both map to named rows.
Refs: F-C14, kvm TEMPLATE-FEEDBACK 2.

**T-11 [D-3] - Test-injection carve-out for CONFIG-ENV-VARS.**
Target: `cli-tool.template.md` TEMPLATE-TABLE.
Action (on D-3 accept): add `TEST-INJECTION | single-declared-env-var | supported`: a spec may declare exactly one named environment variable whose sole purpose is hermetic test isolation (redirecting host identity or configuration root); it must be named in the spec and documented in the man page; it is not a configuration channel, and CONFIG-ENV-VARS forbidden stands for behaviour configuration.
Acceptance: `KVM_MANAGER_HOST` becomes expressible; the forbidden row's intent intact.
Refs: F-B1(c), F-C3 rule 1, kvm DELTA 24.

**T-12 [D-4] - Per-host configuration declarability.**
Target: `cli-tool.template.md`.
Action (only if D-4 accepts): add a `PER-HOST-CONFIG | supported` row acknowledging host-keyed configuration as a declarable property (kubectl contexts, ssh Host blocks class), with a note that the storage format follows PERSISTENCE.
Refs: F-C15, kvm TEMPLATE-FEEDBACK 3.

### 8.2 Goal 2 - whitepaper and templates aligned

**T-13 - Whitepaper appendix renumbering.**
Action: renumber the second `## A.2` (:876) and every subsequent appendix; sweep all `A.<n>` cross-references in pcd/ (including `cli-tool.template.md:185` "See whitepaper A.11") and fix any that shift.
Acceptance: appendix numbers unique and monotonic; every cross-reference resolves to the intended section.
Refs: F-D1.

**T-14 - Whitepaper content update to the 0.4.3 state.**
Action: (a) document the Translation-Inputs provenance contract and the corrected reproducibility tuple (mirroring prompt ## Reports and technical-reference section 12, including the never-collapse-into-one-hash rationale); (b) add the UPGRADE.md change-brief workflow (after T-21); (c) add the feedback/harvest channel (after T-19); (d) summarize the hermetic and no-hollow-test disciplines in the testing material (after T-02); (e) expand MILESTONE pipeline coverage and the ROLE/dual-mode summary; (f) in A.13, replace the embedded full prompt copy with a summary plus a pointer declaring prompts/prompt.md normative, or stamp the copy with the prompt's version and hash plus a sync note; (g) add C17 (N-language translation as ambiguity probe, with the zypper three-way example) and C18 (scale beats fine-tuning; prompt engineering beats model swapping) to A.18; (h) bump the header Version per D-5.
Acceptance: every 0.4.1-0.4.3 changelog row has corresponding body coverage; A.13 cannot silently drift; C17/C18 present.
Refs: F-D2, F-D3, F-C17, F-C18.

**T-15 - A.11 table reconciliation.**
Action: add rows for abap-report, cockpit-module, kubectl-style-cli; align the cli-tool row's languages with the D-1 outcome and its linking wording with T-09; verify every row against its template's META and TEMPLATE-TABLE defaults.
Acceptance: one A.11 row per file in pcd/templates/, values consistent with each template.
Refs: F-D4, F-A5.

**T-16 - user-guide refresh.**
Action: replace the three 0.3.21 currency statements (:443, :501, :1369) - preferably with a single pointer to one authoritative "current schema" location so the scatter cannot recur; add the spec-authoring guidance from C12 (pin physical schema names in INTERFACES; fixtures derive from the canonical schema); host the naming-convention (T-20) and changelog-convention (T-22) subsections.
Refs: F-D5, F-C12.

**T-17 - interview-prompt and reverse-prompt schema refresh.**
Target: `pcd/prompts/interview-prompt.md` (:272, :438), `pcd/prompts/reverse-prompt.md` (:177, :365).
Action: bump skeleton and checklist Spec-Schema values to 0.4.0; add one sentence each on composition (interview output does not generate `Includes:`; reverse-engineering may propose composition for shared fragments) so both prompts are explicit rather than silent about 0.4.0.
Refs: F-D6.

**T-18 - Example set refresh.**
Action: `examples/calc-interest/spec/calc-interest.spec.md` to Spec-Schema 0.4.0 with a real Author line; rename `examples/account-transfer/account-transfer.md` per the T-20 convention and fix in-tree references; stamp the historical TRANSLATION_REPORTs under examples/ as "pre-0.4.3 report format" rather than rewriting them.
Refs: F-D7, F-A11.

### 8.3 Normative core repairs

**T-19 - Standardize the feedback channel and the harvest step.**
Target: user-guide, `pcd/CONTRIBUTING.md`, `pcd/prompts/prompt.md`.
Action: (a) user-guide: `## TEMPLATE-FEEDBACK` is the in-spec channel for template change requests; `[pcd]`-tagged entries in decisions hints are the in-hints channel for framework-level constraints discovered during translation; both advisory, both ignored by translator and lint; (b) prompt.md, one sentence in the spec-reading rules: sections named TEMPLATE-FEEDBACK are advisory to template maintainers and must not influence translation; (c) CONTRIBUTING release checklist: before any template version bump, harvest open TEMPLATE-FEEDBACK sections and `[pcd]` tags from known consumer projects.
Acceptance: channels documented, translator behaviour defined, harvest step in the checklist (this is the structural fix that prevents Section 5 from recurring).
Refs: F-B2.

**T-20 [D-6] - Codify the file-naming convention.**
Target: CONTRIBUTING (or user-guide) for the statement; `pcd/prompts/prompt.md:10`; template EXECUTION input lists (cli-tool :629 and equivalents).
Action: state the invariant once: component specs `<name>.spec.md`; deployment templates `<name>.template.md`; hints `<scope>.<language>.<class>.hints.md`; composition fragments per D-6; change the normative wording `<spec-name>.md` to `<name>.spec.md` everywhere it names the spec input.
Acceptance: convention stated normatively once; prompt and template wording match it; the zypper pinned-prompt misnaming class (`zypper-declarative.md`) can no longer arise from the master text.
Refs: F-A11.

**T-21 - Codify the UPGRADE.md change brief.**
Target: `pcd/prompts/prompt.md` (input list and ## Reports), `pcd/prompts/change-impact.md`, user-guide.
Action: (a) prompt input list: optional `UPGRADE.md`, a KIT change brief directing an incremental pass; when present the translator follows its blast radius and records it under `Upgrade-Brief-SHA256:` in the provenance block; (b) change-impact.md: define UPGRADE.md as the assessment's persisted output (header `<spec> <from-version> -> <to-version>`; fields recommendation, structural impact, blast radius; per-behaviour directives); (c) user-guide: a short workflow section (spec change -> change-impact -> UPGRADE.md -> incremental translation).
Acceptance: the artefact is defined where it is produced and where it is consumed; the label becomes canonical (with T-23).
Refs: F-C19, F-A10.

**T-22 - Codify the external spec changelog.**
Target: user-guide (+ CONTRIBUTING pointer).
Action: a spec's history lives in `CHANGELOG.md` beside the spec; not part of the spec, not read by the translator, not covered by the spec hash; newest entries first; entry form `- YYYY-MM-DD: Version X. <narrative>`; state that this is canonical and that DELTA remains the forward-looking work list (kvm-manager's current no-changelog state and zypper's mixed-format file become project-side follow-ups, Section 7 item 5).
Refs: F-B3.

**T-23 [D-7] - Extend the provenance label canon.**
Target: `pcd/prompts/prompt.md` ## Reports (both modes), the provenance blocks of all report-bearing templates, technical-reference section 12.
Action: add the required line `Prompt-SHA256: <filename> <hash>` (the translator prompt file as read); add the conditional canonical labels `Upgrade-Brief-SHA256` (when a brief is consumed) and `Directive-SHA256` (one line per directive file consumed, retained for ad-hoc directives even after T-02 subsumes the test-isolation one); per D-7, state in technical-reference 12 whether the prompt joins the reproducibility tuple or is recorded-but-excluded, with rationale.
Acceptance: labels identical in all three places; the tuple definition is unambiguous; a future run of either project records its prompt hash.
Refs: F-A10, F-B7.

**T-24 - Fix the embedded-hash semantics.**
Target: `pcd/prompts/prompt.md` ## Spec hash embedding (:678-704); `cli-tool.template.md` :458 and :588; the same "as provided" wording in the other nine provenance-bearing templates (sweep); `lint-rules.md` RULE-18 message (:660).
Action: define once: the embedded `Spec-SHA256` is the SHA256 of the merged spec text (host plus recursively resolved Includes), equal to the host file hash when no `Includes:` is declared; remove "sha256sum <specname>.md" and "as provided as input - not of any transformed or post-processed version"; point at the Spec Composition section as the single definition; align RULE-18's remediation text.
Acceptance: a grep for "as provided" across prompt, templates, and lint-rules returns nothing that contradicts the merged-hash definition.
Refs: F-A1.

**T-25 [D-1] - Reconcile the cli-tool language set.**
Target: `cli-tool.template.md` TYPES (:39), LANGUAGE-ALTERNATIVES rows (:143-146), PRECONDITIONS (:194); A.11 via T-15.
Action (on D-1 keep): TYPES gains `Java | Lean4`; PRECONDITIONS list gains them; LANGUAGE-ALTERNATIVES rows added with the existing caveats (Java jlink runtime image; Lean4 capability warning). (On D-1 drop: remove the two LANGUAGE rows and every Java/Lean4 row from the per-language tables.)
Acceptance: TYPES, TEMPLATE-TABLE, PRECONDITIONS, POSTCONDITIONS, and all per-language tables enumerate one identical language set.
Refs: F-A2.

**T-26 - Remove the stale version literal.**
Target: `cli-tool.template.md:613`.
Action: delete "Current version: 0.3.13" or replace with "Current version: see META Version above" so it cannot drift again.
Refs: F-A3.

**T-27 - Move the pcd-lint block out of the universal prompt.**
Target: `pcd/prompts/prompt.md:358-380`; a pcd-lint-scoped hints file (e.g. `pcd/tools/pcd-lint/spec/pcd-lint.fixtures.hints.md`).
Action: keep the generic fixture-completeness rules (start from the GIVEN; structurally complete; predict which rules fire) in the prompt; move the block under "For `pcd-lint` specifically:" verbatim into the project-scoped file; leave one generic sentence in the prompt ("a spec-consuming tool's hints file enumerates the sections its fixtures require").
Acceptance: the universal prompt contains no project names in normative rules; the pcd-lint guidance survives in project scope and is listed among that project's translation inputs.
Refs: F-A6.

**T-28 - lint-rules: composition targets and template inventory.**
Target: `pcd/tools/shared/spec/lint-rules.md`.
Action: (a) add `none` to DeploymentTemplate with a comment defining composition targets; (b) define rule behaviour for `Deployment: none`: a reduced RULE-01 required set for composition targets (META, TYPES, and at least one of BEHAVIOR/INVARIANTS; PRECONDITIONS/POSTCONDITIONS/EXAMPLES optional), documented in the rule; (c) add `kubectl-style-cli`, `spack-package`, `cockpit-module`, `abap-report` to the enum; (d) annotate `wasm`, `ebpf`, `enterprise-software`, `academic` as planned/no-template-yet so the enum's intent is explicit; (e) align RULE-18's message per T-24; (f) append the changelog entry.
Acceptance: lint-rules.md is valid under its own rule set; every file in pcd/templates/ has its Deployment name in the enum; both consumer specs re-lint clean against the extended fragment.
Refs: F-A4, F-A5.

**T-29 [D-8] - Version grammar.**
Target: `lint-rules.md` RULE-02c (:227-233) and the SemanticVersion type comment (:60-61); the VERSION rows of all templates (cli-tool :132).
Action (on D-8 accept): RULE-02c pattern becomes the alternation `^([0-9]+\.[0-9]+\.[0-9]+|[0-9]{4}\.[0-9]{2}\.[0-9]{2}\.[0-9]{2})$`, message naming both accepted schemes (semver MAJOR.MINOR.PATCH, or dated YYYY.MM.DD.VV); template VERSION rows note both.
Acceptance: kvm-manager.spec.md (2026.06.09.03) and zypper-declarative.spec.md (0.6.10) both pass RULE-02c.
Refs: F-A7.

**T-30 [D-9] - Bring the three lagging templates to baseline.**
Target: `abap-report.template.md`, `spack-package.template.md`, `cockpit-module.template.md`.
Action: abap-report: add `## EXECUTION` (delivery phases for report generation) or `EXECUTION: none` in META with rationale, plus the standard Translation-Inputs provenance block; spack-package: add `EXECUTION: none` to META (A.11 already states it has no execution phase) and add the report deliverable plus provenance block (a TRANSLATION_REPORT is meaningful for `package.py` output); cockpit-module: add the report deliverable plus provenance block; all three: restamp Spec-Schema per D-9 after a review pass against 0.4.0.
Acceptance: RULE-14 clean across pcd/templates/; every template with a report contract carries the provenance block.
Refs: F-A8, F-A9.

**T-31 [D-10] - Milestone status policy.**
Target: user-guide (process section), optionally one sentence in prompt.md.
Action: per D-10 either (a) the pipeline advances `Status: pending -> active -> released` as part of each milestone pass and the user-guide documents the operator step, or (b) MILESTONE Status is declared informational for solo-maintainer projects and DELTA "(DELIVERED)" markers are the progress record. Recommendation: (a), noting that full-spec translation remains valid when no milestone is active.
Refs: F-B8.

### 8.4 Suggested execution order

1. Decisions D-1..D-10 (Section 9).
2. T-01 (pure restore, zero risk) and T-26 (one line).
3. T-02 and T-24, T-27 together (one prompt.md pass), then T-23 (prompt + templates + tech-ref).
4. T-03..T-07 (templates and hints, Goal 1 core), then T-08..T-12 per decisions.
5. T-28, T-29 (lint-rules), T-30 (lagging templates), re-lint both project specs.
6. T-19..T-22 (conventions), T-16..T-18 (docs small), then T-13..T-15 (whitepaper last, after the normative state it must describe is final).

---

## 9. Decisions required from the maintainer

| ID | Question | Options | Recommendation |
|---|---|---|---|
| D-1 | cli-tool language set | keep Java + Lean4 as supported / drop them | keep: the per-language deliverable, build, and gate rows already exist and are coherent |
| D-2 | Linking relaxation | dynamic-against-packaged-lib supported / keep static-only | accept: kvm precedent, supply-chain rationale already written in your own hints |
| D-3 | Test-injection env var | one spec-declared variable supported / keep absolute prohibition (kvm then needs a config-path argument instead) | accept the single declared variable |
| D-4 | PER-HOST-CONFIG key | add / defer | defer unless a second project needs it; TEMPLATE-FEEDBACK keeps the request on record |
| D-5 | Doc version policy | doc header tracks framework version at last content revision / independent doc versions | header tracks framework version |
| D-6 | Composition fragment naming | `lint-rules.md` stays bare .md (fragments are their own class) / rename to `lint-rules.spec.md` (uniform ".spec.md always") | uniform .spec.md; cost is one `Includes:` line in each consumer spec plus changelog entries |
| D-7 | Prompt in the reproducibility tuple | extend tuple to include the prompt / record Prompt-SHA256 but keep it outside the tuple, with rationale | extend: both projects already pin the prompt, which is a de-facto vote |
| D-8 | Version grammar | accept YYYY.MM.DD.VV alongside semver / migrate kvm-manager to semver | accept the alternation; it codifies your stated convention |
| D-9 | Spec-Schema restamp policy for templates | restamp all to 0.4.0 after review / leave as written-against markers | restamp after review, then the value means "validated against" |
| D-10 | MILESTONE Status | pipeline-maintained / informational | pipeline-maintained |

---

## Changelog

- 2026.06.10.01 - Initial report. Corpus consistency-check-for-fable-5.zip (snapshot 2026-06-09). 12 normative-core findings (F-A1..A12), 8 project findings (F-B1..B8), 21-row lessons harvest (F-C1..C21), 8 documentation findings (F-D1..D8), provenance verification for all four TRANSLATION_REPORTs, 31 tasks (T-01..T-31), 10 decisions (D-1..D-10), 5 recorded out-of-scope items.
