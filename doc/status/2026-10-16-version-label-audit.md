# PCD Version-Label Audit

Version:     2026.10.16.02
Refers-to:   git history of the repository up to 2943cd3
Decided-by:  Matthias G. Eckermann, 2026-10-16
Executed-by: Claude, same day, per maintainer instruction

## 1. Rule

A version label names one content. When the bytes of a specification,
template, hints file or pinned copy change, its `Version:` field moves.

Document headers follow D-5 instead: a header states the framework
version at the document's last content revision. The framework version is
the latest row of the changelog table in `doc/whitepaper.md`. Every
framework change gets a row there.

## 2. Method

For every commit, each changed `*.md` outside `*/code/*` with a `Version:`
field in its first 40 lines was compared before and after the commit. An
unchanged field with changed bytes is a finding: from that commit on, the
label names two contents. The first run used 12 lines and missed
`templates/abap-report.template.md`, whose `Version:` field is on line 38.
A second run with 40 lines added one finding for it.
`scripts/check-version-bump.sh` (`make check-versions`) applies the same
comparison to the working tree.

Not covered by this audit:

- Files without a version of their own: the prompts, and seven of the ten
  hints files. The `Version:` field in `hints/mcp-server.go.mcp-go.hints.md`
  records the version of the mcp-go library, not of the file.
- The staged copies in the mcp-server-pcd asset store, which are under
  `*/code/*`.

Prompts and hints are still identified exactly: since 0.4.3, a
TRANSLATION_REPORT records the SHA-256 of every translation input.

## 3. Labels that name more than one content

Each row lists the commits that changed the file after the label was set.
Paths are given as they were at the time. **Bold** marks a label that was
still the current one before 0.4.9.

### 3.1 Templates

| File | Label | Changed in place by |
|---|---|---|
| `templates/abap-report.template.md` | **0.3.29** | 35210db (2026-06-01) |
| `templates/backend-service.template.md` | 0.3.16 | 34a1d7c (2026-03-28) |
| `templates/backend-service.template.md` | 0.3.20 | acb3613 (2026-04-13), 8201123 (2026-05-18), 06da863 (2026-05-18), 35210db (2026-06-01) |
| `templates/cli-tool.template.md` | 0.3.7 | a22741c (2026-03-19) |
| `templates/cli-tool.template.md` | 0.3.13 | b7201ad (2026-03-25), 34a1d7c (2026-03-28) |
| `templates/cli-tool.template.md` | 0.3.20 | acb3613 (2026-04-13), 8201123 (2026-05-18), 06da863 (2026-05-18) |
| `templates/cli-tool.template.md` | 0.3.29 | 35210db (2026-06-01) |
| `templates/cloud-native.template.md` | 0.3.14 | e4f92a9, a7de58f, 980751f, b7201ad (2026-03-25), 34a1d7c (2026-03-28) |
| `templates/cloud-native.template.md` | 0.3.19 | 8ae12e9 (2026-04-13), 8201123 (2026-05-18), 06da863 (2026-05-18), 35210db (2026-06-01) |
| `templates/cockpit-module.template.md` | **0.1.0** | bd73c6f (2026-04-27) |
| `templates/gui-tool.template.md` | 0.3.18 | 34a1d7c (2026-03-28) |
| `templates/gui-tool.template.md` | 0.3.20 | 8ae12e9 (2026-04-13), 06da863 (2026-05-18), 35210db (2026-06-01) |
| `templates/kubectl-style-cli.template.md` | 0.1.0 | 8201123 (2026-05-18), 06da863 (2026-05-18), 35210db (2026-06-01) |
| `templates/library-c-abi.template.md` | 0.3.7 | a22741c (2026-03-19) |
| `templates/library-c-abi.template.md` | 0.3.13 | 34a1d7c (2026-03-28) |
| `templates/library-c-abi.template.md` | 0.3.20 | 8ae12e9 (2026-04-13), 06da863 (2026-05-18), 35210db (2026-06-01) |
| `templates/mcp-server.template.md` | 0.3.13 | b7201ad (2026-03-25), d0e090a (2026-03-26), 34a1d7c (2026-03-28) |
| `templates/mcp-server.template.md` | 0.3.20 | acb3613 (2026-04-13), 8201123 (2026-05-18), 06da863 (2026-05-18), 35210db (2026-06-01) |
| `templates/mcp-server.template.md` | **0.3.22** | 71b4804 (2026-08-05) |
| `templates/project-manifest.template.md` | 0.3.13 | 34a1d7c (2026-03-28) |
| `templates/project-manifest.template.md` | 0.3.19 | 8ae12e9 (2026-04-13), 06da863 (2026-05-18), 35210db (2026-06-01) |
| `templates/python-tool.template.md` | 0.3.7 | a22741c (2026-03-19) |
| `templates/python-tool.template.md` | 0.3.13 | 34a1d7c (2026-03-28) |
| `templates/python-tool.template.md` | 0.3.20 | 8ae12e9 (2026-04-13), 8201123 (2026-05-18), 06da863 (2026-05-18), 35210db (2026-06-01) |
| `templates/verified-library.template.md` | 0.3.7 | a22741c (2026-03-19) |
| `templates/verified-library.template.md` | 0.3.13 | 34a1d7c (2026-03-28) |
| `templates/verified-library.template.md` | 0.3.20 | 8ae12e9 (2026-04-13), 06da863 (2026-05-18), 35210db (2026-06-01) |

Most of these come from four rows of the framework changelog. Each of
them changed six to eleven templates without moving their versions: 0.3.22
(acb3613, 8ae12e9), 0.3.25 (8201123), 0.4.0 (06da863) and 0.4.3
(35210db). The mcp-server-pcd asset store was staged from the templates
on 2026-05-18 (66255fd), after 8201123 and 06da863. Its copies therefore
have these labels but contain the later contents.

### 3.2 Specifications and pinned copies

| File | Label | Changed in place by |
|---|---|---|
| `examples/account-transfer.md` | 0.3.7 | a22741c (2026-03-19) |
| `examples/account-transfer.md` | 0.3.8 | 34a1d7c (2026-03-28) |
| `examples/account-transfer/account-transfer.md` | **0.3.8** | 06da863 (2026-05-18) |
| `examples/calc-interest/spec/calc-interest.spec.md` | **0.2.0** | 06da863 (2026-05-18) |
| `examples/calc-interest/pcd-artifacts/cli-tool.template.md` | **0.3.20** | 963c4b7 (2026-04-13), 06da863 (2026-05-18) |
| `tools/pcdp-lint/spec/pcdp-lint.md` | 0.3.7 | a22741c (2026-03-19), 56d4ef6 (2026-03-20), cebbb3a (2026-03-23) |
| `tools/pcdp-lint/spec/pcdp-lint.md` | 0.3.13 | 76e5c49, 300ddac (2026-03-24), b7201ad (2026-03-25), f15839b (2026-03-27) |
| `tools/pcd-lint/spec/pcd-lint.md` | 0.3.13 | 6b97acb (2026-04-02) |
| `tools/pcd-lint/spec/pcd-lint.md` | 0.3.21 | acb3613 (2026-04-13), 8c7d70b (2026-04-13) |
| `tools/pcd-lint/spec/pcd-lint.md` | 0.4.0 | d126864 (2026-05-18), c8b5455 (2026-05-18) |
| `tools/mcp-server-pcdp/spec/mcp-server-pcdp.md` | 0.1.0 | f15839b (2026-03-27) |
| `tools/mcp-server-pcd/spec/mcp-server-pcd.md` | 0.1.0 | 4d9e434 (2026-03-30), 6b97acb (2026-04-02) |
| `tools/mcp-server-pcd/spec/mcp-server-pcd.md` | 0.3.2 | 06da863 (2026-05-18) |
| `tools/mcp-server-pcd/spec/mcp-server-pcd-spec.md` | 0.3.2 | 06da863 (2026-05-18) |
| `tools/mcp-server-pcd/spec/mcp-server-pcd.spec.md` | 0.3.2 | 4e23b63 (2026-05-22) |
| `tools/shared/spec/lint-rules.md` | 0.4.0 | 2cf7d17 (2026-06-09) |

The calc-interest translations of 2026-04-13 record Spec-SHA256
`8279a6f935e0a8c1f7e3caa355a553dcb6470960bb2d87b6b4cda99caa48f941` and
name `cli-tool.template v0.3.20`. After 06da863, neither the spec nor the
pinned template in `pcd-artifacts/` contains the bytes those runs read.

06da863 changed only the example headings, from `EXAMPLE: <name>` to
`### EXAMPLE: <name>`, which RULE-06 requires since 0.4.0. The April spec
can be rebuilt from that diff and hashes to `8279a6f9...` again, but it
fails pcd-lint with eight errors. The current spec fails one rule as well.
It has two behaviours, so RULE-10 attributes examples by name, and no
negative example names `calculate-simple-interest`. pcd-slice reports the
same five examples as unassignable.

### 3.3 Hints

| File | Label | Changed in place by |
|---|---|---|
| `hints/mcp-server.go.mcp-go.hints.md` | v0.46.0 | 34a1d7c (2026-03-28), 4d9e434 (2026-03-30) |

This field is the library version (see section 2), so these rows show a
missing file version rather than a moved one.

## 4. Document headers (D-5)

D-5 allows several revisions under one header value, provided that value
is the framework version at the last revision. Deviations found:

| Document | Header | Deviation | Status |
|---|---|---|---|
| `doc/whitepaper.md` | 0.4.0 | left at 0.4.0 while rows 0.4.1 to 0.4.4 were added | brought forward in 0.4.5 |
| `doc/technical-reference.md` | 0.4.0, later 0.5.x | left at 0.4.0 until at least 2026-06-01, later a 0.5.x line of its own (0.5.3 last) | restamped to 0.4.8 in 0.4.9 |
| `doc/spec-composition.md` | 0.4.0 | kept 0.4.0 after its revision of 2026-07-30 (framework 0.4.5) | restamped to 0.4.5 in 0.4.9 |
| `README.md` | 0.3.23 | status line not maintained | set to 0.4.9 in 0.4.9 |

## 5. Framework changes without a row

These changes reached templates, prompts, hints or tools but have no row
of their own in the framework changelog. The list covers changes whose
commit subject names them; it is not exhaustive.

| Date | Commits | Change | Table stood at |
|---|---|---|---|
| 2026-04-20 | 1bd9b03 | cockpit-module template | 0.3.22 |
| 2026-04-24/25 | d880c9e, 02dbee9, dae02a3 | kubectl-style-cli template, lint registration, milestone hints | 0.3.22 |
| 2026-05-18/19 | c8b5455, 7dd3730, 19d3199, fdb46c3 | prompt and template changes labelled v0.4.1 to v0.4.4 in their commit subjects | 0.4.0 |
| 2026-05-22 | 8c19c55, 4e23b63 | security-reviewer prompt | 0.4.0 or 0.4.1 |
| 2026-06-01 | 4d9600f | cli-tool C++ milestone hints | 0.4.2 |
| 2026-06-09/10 | 2cf7d17, d9a3f6f | lint-rules.md as the single rule source; consistency batch, including T-23 (designated v0.4.4) | 0.4.3 |
| 2026-07-29 | 530d36a | libpcd restructure: shared engine, RULE-22 to RULE-25, pcd-emit-types (treated as 0.5.0 in `tools/README.md` and the technical reference) | 0.4.4 or 0.4.5 |
| 2026-08-05 | 71b4804 | systemd unit as a file deliverable of the mcp-server template | 0.4.5 |
| 2026-08-21/24 | 4f7f564, 1f3995c, 1970974, d69c377, c758463 | pcd-slice 0.1.0 and 0.1.1, pcd-translate skill | 0.4.5 |

Row 0.4.4 (2026-06-15) records a whitepaper change, not T-23. Rows 0.4.1
to 0.4.4 do not correspond to the commit labels of 2026-05-18/19. The
table itself has two rows labelled 0.3.19, dated 2026-03-27 and
2026-03-31.

## 6. Dispositions

| Item | Disposition |
|---|---|
| `templates/mcp-server.template.md` 0.3.22 | moved to 0.3.23 |
| `templates/cockpit-module.template.md` 0.1.0 | moved to 0.1.1 |
| `templates/abap-report.template.md` 0.3.29 | moved to 0.3.30 (0.4.10) |
| `examples/account-transfer/account-transfer.md` 0.3.8 | moved to 0.3.9 |
| `examples/calc-interest/spec/calc-interest.spec.md` 0.2.0 | moved to 0.2.1 (0.4.10); each example heading names its behaviour with `(of: <behavior>)`; Spec-Schema restamped from 0.3.22 to 0.4.0 after clean runs of pcd-lint 0.4.0 and pcd-slice 0.1.1 (D-9) |
| `examples/calc-interest/pcd-artifacts/cli-tool.template.md` | restored to the bytes of the 2026-04-13 run (0.4.10) |
| Document headers (section 4) | restamped per D-5 |
| `tools/README.md` RULE-22 to RULE-25 | "Since" 0.5.0 corrected to 0.4.5 |
| Historical labels (section 3) | kept; resolve a reference by commit or by SHA-256 |
| Recurrence | `make check-versions`, added in 0.4.9 |

Open:

- The calc-interest Rust and Java code is still the translation of spec
  0.2.0 (Spec-SHA256 `8279a6f9...`); spec 0.2.1 has to be translated
  again.
- Hints files without a file version of their own.
- The mcp-server-pcd asset store (tracked, staged 2026-05-18) is replaced
  by the next mcp-server-pcd translation.

## Changelog

- 2026.10.16.01 - Initial record of the audit and of the 0.4.9
  dispositions.
- 2026.10.16.02 - abap-report finding from the second audit run (40-line
  window); calc-interest dispositions. Both shipped in framework 0.4.10.
