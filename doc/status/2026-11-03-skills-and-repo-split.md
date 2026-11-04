# PCD as Skills - Decision Record and Path

Version:     2026.11.04.05
Refers-to:   2026-06-09-pcd-consistency-decisions.md (D-6, D-7, D-10, T-20, T-23);
             github.com/mge1512/skill-claudism-pass (skill repository convention)
Decided-by:  Matthias G. Eckermann, 2026-11-03 and 2026-11-04
Recorded-by: Claude (Fable 5.1), same days
Baseline:    commit 6392d57 (2026-10-16), framework 0.4.10
Status:      Draft; P-8 awaits a ruling, everything else is ready to execute

<!-- variant: eu -->

---

## 1. Decisions (maintainer)

| ID | Date | Question | Decision |
|---|---|---|---|
| D-1 | 11-03 | Primary distribution form | PCD is distributed as skills. A skill is a directory with `SKILL.md` at its root, packaged as a `.skill` ZIP for claude.ai and as a signed RPM for SLES. |
| D-2 | 11-03 | Skill content | Each skill is self-contained: it restates the rules, and it ships the templates and the shipped hints tiers it needs. A skill can stand in for the MCP server. |
| D-3 | 11-03 | MCP server and CLIs | `mcp-server-pcd`, `pcd-lint` and `pcd-slice` remain available in addition to the skills. |
| D-4 | 11-03 | Engine discovery in a session | Two options inside every skill. A: use the MCP server when it is present. B: a shell script in the skill tests whether the CLIs are installed. |
| D-5 | 11-03 | Spec-author without an engine | The spec-author role always proceeds ("speed matters"). |
| D-6 | 11-03 | Git in the workflow | Git records the workflow and its steps (direction set; mechanics in P-8, P-9, Phase D). |
| D-7 | 11-03 | Spec authoring | A specification is written across several issues and pull requests, not in one run (direction set; mechanics in P-9). |
| D-8 | 11-04 | Skill granularity | Two skills, not one per role: `pcd-spec` for every task that changes a specification, `pcd-translate` for every task that works from a specification. Roles are modes inside a skill. |
| D-9 | 11-04 | Skill repositories | The convention of `skill-claudism-pass` applies: one git repository per skill, `SKILL.md` at the repository root, installed by cloning into the skills directory of any framework; a `.skill` ZIP attached to each release. |
| D-10 | 11-04 | MCP server and tools, scope | Kept as they are for now. No tool or resource is dropped. |
| D-11 | 11-04 | Binaries in the skill | `pcd-lint` and `pcd-slice` are not shipped inside `pcd-translate`. A manual override is allowed: the operator declares that lint and slice were run before the upload. |
| D-12 | 11-04 | Slice in the MCP server | Direction: slicing may be added to the MCP server, so that an agentic tool, a skill and the MCP server can work autonomously (mechanics in P-11). |

## 2. Proposals

Withdrawn: P-6, dropping the asset-serving MCP tools (D-10). Withdrawn: P-7, splitting the tools into a repository of their own; one source repository suffices once the skills are published per D-9.

Awaiting a ruling: P-8.

Ready to execute: the rest; P-9 and O-3 apply as defaults until overruled.

| ID | Proposal | Depends on |
|---|---|---|
| P-1 | `pcd-spec` modes: interview, reverse, change, milestone (the architect work: project-manifest, milestone design), spec-review (the pre-translation review, findings ordered by cost-if-translated-unfixed). `pcd-translate` modes: translator, test-author, review, security-review, tiebreak. `SKILL.md` is a short dispatcher; each mode is a file under `modes/`, selected by `ROLE.md` in the input directory or by the skill argument. The line between the two skills is checkable: a PR from a translation branch that changes `*.spec.md` fails CI. | D-8 |
| P-2 | Skill directories are build outputs of `make skills` from the canonical tree. Rule documents, templates and hints are copied verbatim, and `make check-skills` fails on drift, by the mechanism of the AUTO blocks and `make check-spec`. `.gitattributes` marks the copies generated, as `tools/*/code` is today. The skill text says: the tool decides, the rule text explains a finding. | D-2 |
| P-3 | `SKILL.md` and the mode files become the canonical home of each role's procedure. `prompts/` retires as a source directory. What the MCP server serves under `pcd://prompts/` is derived from the mode files at build time (D-10: the server keeps serving it). | P-2 |
| P-4 | Guard, per skill. `pcd-spec` never requires an engine: it lints when one is present, otherwise it proceeds and the hand-over message and the PR body state `Lint: not run`. `pcd-translate` requires an engine in every mode. Order: B first (CLI, shared filesystem, exit codes), A second (MCP server, for hosts without a shell). Override (D-11): when no engine is present and the operator declares that lint and slice were run before the upload, the guard verifies what it can without the tools: `spec.d/` and `MANIFEST.tsv` present, and the source hash on the first line of `spec.d/preamble.md` equal to `sha256sum` of the specification. Hashes equal: proceed, and record the override. Hashes unequal: stop, as stale. The override never enters the spec bytes; the PR gate re-runs both tools and is the authority. | D-4, D-5, D-11 |
| P-5 | Provenance lines in `TRANSLATION_REPORT.md` and `TEST_REPORT.md`: `Skill-SHA256: <skill>/SKILL.md <hash>`, `Skill-Manifest-SHA256: <hash>`, `Toolchain: <tool> <version> <cli or mcp>` (one line per tool used), and `Guard-Override: <declaration, verified by sha256sum, tools not run in session>` when P-4's override was used. The existing `Mode:` line names the mode. Contract change of the T-23 class: one framework row, technical reference section 12. | P-2 |
| P-8 | Milestone state moves from the `Status:` line in the spec into git references: a tag per released milestone, a branch per active one. This supersedes D-10 of 2026-06-10. Reason: the `Status:` line sits inside the hashed bytes, so every advance moves the spec hash and makes the last report stale. Alternative with a smaller footprint: keep the line and exclude it from the hash in libpcd. | D-6 |
| P-9 | The unit of a spec change is one BEHAVIOR, or one TYPE or INVARIANT change: one issue, one branch, one PR. One PR is one Version move, one changelog row, one lint run and one `assess_change_impact` result recorded in the PR body. Translation runs per milestone or tag, never per PR. Default; reversible. | D-7 |
| P-10 | Publishing. A release workflow in `pcd`, modelled on `release.yml` of `skill-claudism-pass`, runs on a version tag: `make check-skills`, the `check-tools.sh` self-test, the `.skill` ZIPs built with the same rsync-and-zip step, `SHA256SUMS`, the GitHub release, then a mirror push of each built skill directory to `pcd-spec` and `pcd-translate` (one commit per release, same tag). The skill repositories are never hand-edited; issues are filed in `pcd`. The `pcd-skills` RPM is built in the build service from the same release tarball and installs to `/usr/share/pcd/skills/<name>/`. | D-9 |
| P-11 | Slicing in the engine. The behaviours of `pcd-slice` move into libpcd (via `Includes:`, as the lint and types rules did), so that the CLI and the MCP server are both front ends over one engine. The server gains `slice_spec` and `check_slices`, with the inputs of the CLI commands `slice` and `check`. `pcd-translate` option A then covers every guard step. Batched into the next re-translation (Phase C). | D-12 |

Open points:

| ID | Status |
|---|---|
| O-1 | Closed by D-8: `security-reviewer.md` becomes the security-review mode of `pcd-translate`. |
| O-2 | Closed by D-11. The guard script still looks on PATH first and in the skill's own `bin/` second, so shipping binaries later needs no guard change. |
| O-3 | Default: a skill ships the hints tiers its modes read (milestones, library and style hints under `hints/`); project-local tiers (decisions hints, `.pcd/hints/`) stay in the project. |

---

## 3. Target layout

One source repository, `pcd`, with the tools in place as today:

```
skills/
  pcd-spec/
    SKILL.md              dispatcher; front matter metadata: version,
                          spec-schema, minimum tool versions
    modes/                interview.md, reverse.md, change.md,
                          milestone.md, spec-review.md
    SOURCES               the copies, one glob and target subdirectory per line
    scripts/              check-tools.sh (copy of scripts/check-tools.sh)
    references/           lint-rules, types-table-rules, spec format (copies)
    templates/            all deployment templates (copies)
    MANIFEST.sha256
  pcd-translate/
    SKILL.md              today's guard, extended per P-4 and P-5
    modes/                translator.md, test-author.md, review.md,
                          security-review.md, tiebreak.md
    SOURCES
    scripts/              check-tools.sh (copy)
    references/
    templates/
    hints/                shipped tiers (O-3)
    MANIFEST.sha256
templates/  hints/        canonical, unchanged
tools/                    unchanged (D-10); tools/shared/spec/ stays the home
                          of the rule documents
doc/  examples/
tools/pcd-skills/         RPM and deb, next to tools/pcd-templates/
.github/workflows/release.yml   P-10
```

Everything under `skills/*/` except `SKILL.md`, `SOURCES` and `modes/` is written by `make skills`; the guard script has one source, `scripts/check-tools.sh` at the repository root, and the copy in each skill is derived like the rest. `check-tools.sh` runs `sha256sum -c MANIFEST.sha256` before any other check. A runtime that rewraps or truncates fetched text (the failure mode recorded in the claudism-pass README) then fails at session start, not inside a translation.

Mirror repositories, written by the release workflow only:

```
pcd-spec/              SKILL.md at the root, the built directory as above
pcd-translate/         same
```

Install on any framework: `git clone <repo> <skills-dir>/pcd-spec`, the target directory named after the skill. Commands: `/pcd-spec`, `/pcd-translate`.

---

## 4. The path, as pull requests

One PR per row. Each PR includes its own Version moves and changelog rows (P-9 applied to the framework itself). Phases A and D touch no generated code. Phase C is the one re-translation, and everything that changes a tool spec is batched into it.

### Phase A - skills in the repository

| PR | Scope | Gate | Version effect |
|---|---|---|---|
| A-1 | This record under `doc/status/`. | none | none |
| A-2 | Scaffolding: `make skills` and `make check-skills`, `scripts/check-tools.sh` (`/bin/sh`, exit 0/1/2, manifest check first, tool versions on stdout, PATH then `bin/`), front matter metadata, `MANIFEST.sha256` generation, `.gitattributes` entries, `tools/pcd-skills/`. No skill content yet beyond the existing `pcd-translate`. | `make check-skills` clean; `make check-versions` clean | framework row |
| A-3 | `pcd-spec`: modes built from `interview-prompt.md`, `reverse-prompt.md`, `change-impact.md` and the pre-translation review practice; P-4 proceed rule; writes the Version move and the `CHANGELOG.md` entry beside the spec; records the change-impact result for the PR body. | a spec written with the skill passes `pcd-lint` in CI; `make check-skills` clean | framework row |
| A-4 | `pcd-translate`: modes from the translator and test-author roles of `prompt.md`, `reviewer.md`, `security-reviewer.md`, `tiebreaker.md`; guard per P-4 with override; template and hints read from the skill; P-5 lines in the reports; incremental mode with `UPGRADE.md` as a committed file outside the hash scope. | `make check-skills` clean; one dry run against `pcd-slice.spec.md` bundles without a translation; one dry run with the override path | framework row |
| A-5 | Retire `prompts/` as a source directory (P-3). `materialise-prompt.sh` and the pcd-lint bundle target read from `skills/pcd-translate/`; the MCP server's embedded prompts are derived from the mode files; `.sha256` sidecars regenerate. Technical reference section 12 gains the P-5 labels; user guide section 3 gains "Using the PCD skills" and lists the MCP server as the integration path; `AGENTS.md` replaces the two-layer prompt architecture with the skill layer; whitepaper row. | `make check-spec` clean after `make spec` (the pcd-lint bundle hash moves; note it in the row, re-translation deferred to C-2) | framework 0.5.0 (contract change: prompt retired, labels added) |
| A-6 | Publishing (P-10): `release.yml`, the two mirror repositories created and first filled, the `pcd-skills` package in the build service. | a tag produces the release with both `.skill` files and `SHA256SUMS`; `rpm -qpl` shows `/usr/share/pcd/skills/`; a clone of each mirror passes `sha256sum -c MANIFEST.sha256` | package versions |
| A-7 | Validation without translation money: one spec change on `pcd-slice.spec.md` through `pcd-spec` in claude.ai, delivered as a PR with the P-9 body. | PR gate: `pcd-lint`, `check-versions`, `check-spec` | spec Version move only |

### Phase B - removed

The repository split (P-7) is withdrawn. T-20 (rename of the rule documents to `.spec.md`, D-6 of 2026-06-10) stays an independent housekeeping item and is not part of this path.

### Phase C - engine, one re-translation

| PR | Scope | Gate | Version effect |
|---|---|---|---|
| C-1 | Slicing in the engine (P-11): `pcd-slice.spec.md` behaviours into libpcd via `Includes:`, `slice_spec` and `check_slices` in `mcp-server-pcd.spec.md`, the `pcd-slice` CLI spec reduced to a front end; the `mcp-server` template's provenance block gains the P-5 labels. | `pcd-lint` clean; `assess_change_impact` result in the PR body (TYPES touched: expect full regeneration) | spec Version moves; template Version moves |
| C-2 | Re-translate libpcd, pcd-lint, pcd-slice and mcp-server-pcd in one batch, baseline / generation / fixes commits each, reports with the new labels. | compile gate, independent tests, `verify_spec_hash` current for every spec | `TRANSLATION_REPORT.md` per tool |
| C-3 | Deferred (D-10): content variants of the path-based MCP tools for remote hosts without a shared filesystem. Revisit when such a host is in use. | - | - |

### Phase D - git as the workflow record (D-6, D-7)

| PR | Scope | Gate | Version effect |
|---|---|---|---|
| D-1 | Hooks and CI: pre-commit runs `pcd-lint` on `*.spec.md`; the PR workflow runs `check-spec`, `check-versions`, `verify_spec_hash`, and fails a translation-branch PR that changes `*.spec.md` (P-1). | CI green on a deliberate failing fixture | none |
| D-2 | Issue and PR templates for P-9: one BEHAVIOR or TYPE per issue; PR body with Version move, changelog row, change-impact result, `Lint:` line, `Guard-Override:` line where used. | first PR filed with the template | none |
| D-3 | Milestone state in git references (P-8, after the ruling): `set_milestone_status` becomes a documented git operation; RULE-15's single-active check reads refs or is retired; the `Status:` line leaves the spec. | `pcd-lint` clean on all specs; `verify_spec_hash` stays current across a milestone advance | lint-rules Version move; engine change batched into the next re-translation |
| D-4 | Worktree per translation run documented in user guide section 3.3 (stage 3): `git worktree add <dir> <baseline-tag>` is the sandbox the translator sees. | docs only | user guide Version move |

Order inside a phase is fixed by the dependencies. Between phases: A runs first, D can interleave with it from D-1 on, and C closes. C-2 is the single expensive step, and it collects every hash move from A-5 and C-1.

---

## 5. What does not change

- The spec is the sole durable human artefact; generated code is never hand-edited.
- The reproducibility tuple keeps its members and gains the skill and toolchain lines (P-5). Per-file hashes as read, never a combined hash.
- Changelog rows stay append-only; a Version moves whenever bytes move.
- Signed packages from the build service for every installable artefact. The `.skill` ZIP is an output of the same release, published with `SHA256SUMS`.
- No internal names, hosts or review markers in anything under `doc/` that is meant for the public repository.

---

## Changelog

- 2026.11.03.01 - Initial record of D-1 to D-7, proposals P-1 to P-9, open points O-1 to O-3, the target layout and the PR path.
- 2026.11.04.01 - D-8 to D-12 recorded. Two skills instead of one per role (D-8); P-1 rewritten with modes. Skill repositories follow the claudism-pass convention and are release mirrors of `pcd` (D-9, P-10); P-7 withdrawn, Phase B removed, one source repository. MCP server and tools kept as they are (D-10); P-6 withdrawn, C-3 deferred. No binaries in the translation skill, manual override with hash verification (D-11, P-4, P-5 `Guard-Override:`); O-2 closed. Slicing in the engine as the next engine change (D-12, P-11, C-1). O-1 closed, O-3 given a default, P-9 marked as default. Front matter metadata replaces the `VERSION` file; manifest check first in `check-tools.sh`.
- 2026.11.04.02 - Executed with PR A-2. Mirror repositories named `pcd-spec` and `pcd-translate` (maintainer ruling of 2026-11-04: the repository name equals the skill name, so a clone needs no rename); P-10 and section 3 updated. Packaging under `tools/pcd-skills/`, next to `tools/pcd-templates/`, instead of `packaging/`. The guard script has one source, `scripts/check-tools.sh`, copied into each skill by `make skills`; `scripts/` inside a skill is therefore derived content. The hand-authored set is `SKILL.md`, `SOURCES` and `modes/`. `SOURCES` added to the layout.
- 2026.11.04.03 - Executed with PR A-3. `pcd-spec` with the modes interview, reverse, change, milestone and spec-review. The interview, reverse and change-impact prompts are copied into the skill under `prompts/` by `make skills` until A-5 moves them into `modes/`; the three new modes are hand-authored under `modes/`. The external spec changelog of T-22 (2026-06-10) is applied by the skill (`CHANGELOG.md` beside the specification, newest first, `- YYYY-MM-DD: Version X. <narrative>`); the user-guide text of T-22 follows in A-5. The skill ships the milestones hints tier for the milestone mode (O-3). pcd-skills packaging moves with the framework version until A-6 derives it from the whitepaper table.
- 2026.11.04.04 - Executed with PR A-4. `pcd-translate` 0.2.0 with the modes translator, test-author, review, security-review and tiebreak. Guard with the D-11 override: bundle hashes verified by `sha256sum`, recorded as `Guard-Override:`. The P-5 lines in the reports. `UPGRADE.md` beside the specification as the committed upgrade brief, outside the specification hash. The four translation-side prompts are copied into the skill under `prompts/` until A-5. Correction to the pre-existing skill text: pcd-slice produces no source-region line numbers, so review modes cite bundle name and section heading. Pulled forward from A-6: `scripts/mirror-skill.sh`, so that the mirror repositories can be filled from a committed state before the release workflow exists; A-6 wraps it in the workflow. `check-tools.sh` ignores `.git` in an installed copy (found by testing a clone of a mirror).
- 2026.11.04.05 - Executed as the follow-up commit to PR A-4, after the maintainer had pushed A-4. Publication mechanics corrected: the mirror tag is the skill's version (`v0.1.0` for metadata version 0.1.0), not a tag copied from `pcd`, and a version the mirror has published is never republished with other content. New `tools/pcd-skills/mirror-release.yml`, installed into each mirror by `scripts/mirror-skill.sh`: on the tag, the mirror verifies the manifest, builds the reproducible `<name>.skill` archive with `git archive` and attaches it with `SHA256SUMS` to a GitHub release, with the mirror's own token, so the mirrors need no credential from `pcd` beyond push. A-6 keeps only the orchestration (run the script for every skill on a `pcd` tag, build the RPM). `check-tools.sh` also ignores `.github` in an installed copy. The row 2026.11.04.04 above describes the pushed state and is unchanged.
