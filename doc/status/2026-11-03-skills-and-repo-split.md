# PCD as Skills - Decision Record and Path

Version:     2026.11.03.01
Refers-to:   2026-06-09-pcd-consistency-decisions.md (D-6, D-7, D-10, T-20, T-23)
Decided-by:  Matthias G. Eckermann, 2026-11-03
Recorded-by: Claude (Fable 5.1), same day
Baseline:    commit 6392d57 (2026-10-16), framework 0.4.10
Status:      Draft, for maintainer ruling on section 2

<!-- variant: eu -->

---

## 1. Decisions (maintainer, 2026-11-03)

| ID | Question | Decision |
|---|---|---|
| D-1 | Primary distribution form | PCD is distributed as skills. A skill is a directory with `SKILL.md`, packaged as a ZIP for claude.ai and as a signed RPM for SLES. |
| D-2 | Skill granularity | One skill per role. Each skill is self-contained: it restates the rules, and it ships the templates and the shipped hints tiers it needs. A skill can stand in for the MCP server. |
| D-3 | MCP server and CLIs | `mcp-server-pcd`, `pcd-lint` and `pcd-slice` remain available in addition to the skills. |
| D-4 | Engine discovery in a session | Two options inside every skill. A: use the MCP server when it is present. B: a shell script in the skill tests whether the CLIs are installed. |
| D-5 | Spec-author without an engine | The spec-author role always proceeds ("speed matters"). |
| D-6 | Git in the workflow | Git records the workflow and its steps (direction set; mechanics in section 2). |
| D-7 | Spec authoring | A specification is written across several issues and pull requests, not in one run (direction set; mechanics in section 2). |

## 2. Proposals awaiting ruling

| ID | Proposal | Depends on |
|---|---|---|
| P-1 | Role set: `pcd-spec-author`, `pcd-translate` (exists), `pcd-test-author`, `pcd-review`, `pcd-tiebreak`. The architect work (project-manifest, milestone design) stays inside spec-author until it needs a skill of its own. | D-2 |
| P-2 | Skill directories are build outputs of `make skills` from the canonical tree. Rule documents are copied verbatim, and `make check-skills` fails on drift, by the mechanism of the AUTO blocks and `make check-spec`. The skill text says: the tool decides, the rule text explains a finding. | D-2 |
| P-3 | `SKILL.md` becomes the canonical home of each role's procedure. `prompts/` retires as a source directory; what the MCP server serves under `pcd://prompts/` is derived from the skills or dropped (P-6). | P-2 |
| P-4 | Guard order: B first (CLI), A second (MCP server). Translate and test-author stop when neither is present. Spec-author, review and tiebreak proceed; the hand-over message and the PR body state `Lint: not run`. The marker never enters the spec bytes; the PR gate is the authority. | D-4, D-5 |
| P-5 | Provenance labels in `TRANSLATION_REPORT.md` and `TEST_REPORT.md`: `Skill-SHA256: <role>/SKILL.md <hash>`, `Skill-Manifest-SHA256: <hash>`, `Toolchain: <tool> <version> <cli or mcp>`, one line per tool used. Contract change of the T-23 class: one framework row, technical reference section 12. | P-2 |
| P-6 | `mcp-server-pcd` keeps the engine tools and drops the asset-serving tools and resources (`list_templates`, `get_template`, `list_resources`, `read_resource`, `pcd://templates`, `pcd://prompts`, `pcd://hints`) after one release that marks them deprecated. It gains content variants of the path-based tools (`verify_spec_hash`, `emit_types`), so that a remote server can serve a sandboxed host that has no shared filesystem. | D-3, D-4 |
| P-7 | Two repositories. `pcd` keeps the name and the paradigm: skills, templates, hints, rule documents, docs, examples, comparisons; CC-BY-4.0. `pcd-tools` contains libpcd, pcd-lint, pcd-slice, pcd-emit-types, mcp-server-pcd, each with its spec and generated code side by side; GPL-2.0-only. `pcd-tools` vendors `pcd` as a submodule pinned to a signed tag. The contract between the two is the Spec-Schema version. | none |
| P-8 | Milestone state moves from the `Status:` line in the spec into git references: a tag per released milestone, a branch per active one. This supersedes D-10 of 2026-06-10. Reason: the `Status:` line sits inside the hashed bytes, so every advance moves the spec hash and makes the last report stale. | D-6 |
| P-9 | The unit of a spec change is one BEHAVIOR, or one TYPE or INVARIANT change: one issue, one branch, one PR. One PR is one Version move, one changelog row, one lint run and one `assess_change_impact` result recorded in the PR body. Translation runs per milestone or tag, never per PR. | D-7 |

Open points that need the maintainer and are not proposals yet:

| ID | Question |
|---|---|
| O-1 | `security-reviewer.md`: a mode of `pcd-review`, or a skill of its own. |
| O-2 | Whether a skill may ship static `pcd-lint` and `pcd-slice` binaries for hosts that can install nothing. P-4 allows it later: the guard script looks on PATH first and in the skill's own `bin/` second. |
| O-3 | Which hints tiers are "shipped": today `hints/` contains milestones, library and style hints for several templates and languages. A role skill ships the tiers its role reads; project-local tiers (decisions hints, `.pcd/hints/`) stay in the project. |

---

## 3. Target layout

Content repository `pcd` (after P-7; before P-7 the same tree inside the current repository, with `tools/` still present):

```
skills/
  pcd-spec-author/
    SKILL.md              procedure: interview, reverse, change mode
    VERSION               skill version, Spec-Schema, required tool versions
    scripts/check-tools.sh
    references/           lint-rules.md, types-table-rules.md, spec format
    templates/            all deployment templates (the author picks one)
  pcd-translate/
    SKILL.md              today's guard, extended by template and hints reads,
                          provenance lines, incremental mode (UPGRADE.md)
    VERSION
    scripts/check-tools.sh
    references/
    templates/
    hints/                shipped tiers (O-3)
  pcd-test-author/        same shape as pcd-translate
  pcd-review/
  pcd-tiebreak/
templates/                canonical, unchanged
hints/                    canonical, unchanged
rules/                    lint-rules.spec.md, types-table-rules.spec.md,
                          types-emit-rules.spec.md (moved from tools/shared/spec,
                          renamed per D-6 of 2026-06-10, executes T-20)
doc/
examples/
packaging/pcd-skills/     RPM and deb: /usr/share/pcd/skills/<role>/
```

Every built skill directory also contains `MANIFEST.sha256`, written by `make skills`, one line per file. The ZIP for claude.ai and the RPM payload are the same bytes.

Tools repository `pcd-tools` (after P-7):

```
pcd/                      submodule, pinned to a signed tag of the content repository
libpcd/spec/              Includes: ../../pcd/rules/lint-rules.spec.md (path moves)
pcd-lint/{spec,code}/
pcd-slice/{spec,code}/
pcd-emit-types/spec/
mcp-server-pcd/{spec,code}/
pcd-templates/            packaging, reads templates and hints from the pin
Makefile                  spec, check-spec, check-versions across the pin, dist
```

`make dist` copies the pinned snapshot into the release tarball, so OBS builds each tools package from one self-contained source.

---

## 4. The path, as pull requests

One PR per row. Each PR includes its own Version moves and changelog rows (P-9 applied to the framework itself). Phases A and D touch no generated code. Phase C is the one re-translation, and everything that changes a tool spec is batched into it.

### Phase A - skills in the current repository

| PR | Scope | Gate | Version effect |
|---|---|---|---|
| A-1 | This record under `doc/status/`. | none | none |
| A-2 | Scaffolding: `make skills` and `make check-skills`, `scripts/check-tools.sh` (`/bin/sh`, exit 0/1/2, prints tool versions on stdout), `VERSION` and `MANIFEST.sha256` generation, `packaging/pcd-skills/`. No skill content yet beyond the existing `pcd-translate`. | `make check-skills` clean; `make check-versions` clean | framework row |
| A-3 | `pcd-spec-author`: built from `interview-prompt.md`, `reverse-prompt.md` and `change-impact.md`; two modes, bootstrap and change; P-4 proceed rule; writes the Version move and the `CHANGELOG.md` entry beside the spec; records the change-impact result for the PR body. | a spec written with the skill passes `pcd-lint` in CI; `make check-skills` clean | framework row |
| A-4 | `pcd-translate` extended: template and hints read from the skill, P-5 lines in the report, incremental mode with `UPGRADE.md` as a committed file outside the hash scope. | `make check-skills` clean; one dry run against `pcd-slice.spec.md` bundles without a translation | framework row |
| A-5 | `pcd-test-author`, `pcd-review`, `pcd-tiebreak`: from the test-author role of `prompt.md`, `reviewer.md` (O-1 ruled first) and `tiebreaker.md`. Each with its reading discipline and forbidden inputs. | `make check-skills` clean | framework row |
| A-6 | Retire `prompts/` as a source directory (P-3). `materialise-prompt.sh` and the pcd-lint bundle target in the Makefile read from `skills/pcd-translate/SKILL.md` instead of `prompts/prompt.md`; `.sha256` sidecars regenerate. Technical reference section 12 gains the P-5 labels; user guide section 3 gains "Using the PCD skills" and moves the MCP server to the integration path; `AGENTS.md` replaces the two-layer prompt architecture with the skill layer; whitepaper row. | `make check-spec` clean after `make spec` (the pcd-lint bundle hash moves; note it in the row, re-translation deferred to C-2) | framework 0.5.0 (contract change: prompt retired, labels added) |
| A-7 | Validation without translation money: one spec change on `pcd-slice.spec.md` through the spec-author skill in claude.ai, delivered as a PR with the P-9 body. | PR gate: `pcd-lint`, `check-versions`, `check-spec` | spec Version move only |

### Phase B - repository split (if P-7 is accepted)

| PR | Scope | Gate | Version effect |
|---|---|---|---|
| B-1 | `pcd-tools` created with `git subtree split --prefix=tools`; history kept, `pcd` not rewritten. `tools/` removed from `pcd` in the same step. | both repositories build their docs; `tools/README.md` moves | none (file moves) |
| B-2 | Submodule pin in `pcd-tools`; `Includes:` paths of `libpcd.spec.md` and the Makefile paths point at the pin; `rules/` created in `pcd` with the T-20 rename. | `make check-spec` in `pcd-tools` | spec hashes of libpcd, pcd-lint and mcp-server-pcd move (Includes path and bundle hash); re-translation in C-2 |
| B-3 | OBS: `pcd-skills` builds from the `pcd` release tarball; the tools packages build from `make dist` of `pcd-tools` with the vendored snapshot. | packages build in OBS; `rpm -qpl` shows `/usr/share/pcd/skills/` | package versions |
| B-4 | `check-versions` extended to the pin: a Spec-Schema change in `pcd` without a matching engine release fails the tools build. | CI run | none |

### Phase C - engine, one re-translation

| PR | Scope | Gate | Version effect |
|---|---|---|---|
| C-1 | `mcp-server-pcd.spec.md`: content variants of `verify_spec_hash` and `emit_types`; asset-serving tools marked deprecated (P-6); the `mcp-server` template's provenance block gains the P-5 labels. | `pcd-lint` clean; `assess_change_impact` result in the PR body (TYPES touched: expect full regeneration) | spec Version moves; template Version moves |
| C-2 | Re-translate mcp-server-pcd, pcd-lint and libpcd consumers in one batch, baseline / generation / fixes commits each, reports with the new labels. | compile gate, independent tests, `verify_spec_hash` current for every spec | `TRANSLATION_REPORT.md` per tool |
| C-3 | Drop the deprecated asset tools one release later. | as C-2 | spec Version move; re-translation |

### Phase D - git as the workflow record (D-6, D-7)

| PR | Scope | Gate | Version effect |
|---|---|---|---|
| D-1 | Hooks and CI: pre-commit runs `pcd-lint` on `*.spec.md`; the PR workflow runs `check-spec`, `check-versions`, `verify_spec_hash`. | CI green on a deliberate failing fixture | none |
| D-2 | Issue and PR templates for P-9: one BEHAVIOR or TYPE per issue; PR body with Version move, changelog row, change-impact result, `Lint:` line. | first PR filed with the template | none |
| D-3 | Milestone state in git references (P-8, after ruling): `set_milestone_status` becomes a documented git operation; RULE-15 single-active check reads refs or is retired; the `Status:` line leaves the spec. | `pcd-lint` clean on all specs; `verify_spec_hash` stays current across a milestone advance | lint-rules Version move; engine change batched into the next re-translation |
| D-4 | Worktree per translation run documented in user guide section 3.3 (stage 3): `git worktree add <dir> <baseline-tag>` is the sandbox the translator sees. | docs only | user guide Version move |

Order inside a phase is fixed by the dependencies. Between phases: A runs first, B and D can interleave, and C closes. C-2 is the single expensive step, and it collects every hash move from A-6, B-2 and C-1.

---

## 5. What does not change

- The spec is the sole durable human artefact; generated code is never hand-edited.
- The reproducibility tuple keeps its members and gains the skill and toolchain lines (P-5). Per-file hashes as read, never a combined hash.
- Changelog rows stay append-only; a Version moves whenever bytes move.
- Signed packages from the build service for every installable artefact. The ZIP for claude.ai is an output of the same build, published with its sha256 and signature.
- No internal names, hosts or review markers in anything under `doc/` that is meant for the public repository.

---

## Changelog

- 2026.11.03.01 - Initial record of D-1 to D-7, proposals P-1 to P-9, open points O-1 to O-3, the target layout and the PR path.
