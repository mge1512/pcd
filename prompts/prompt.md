

I am providing the following input files, all present in the same
input directory alongside this prompt:

1. `<deployment-template>.template.md` — the deployment template defining
   conventions, constraints, defaults, and the full execution recipe for
   this component type.

2. `<spec-name>.md` — the specification for the component to implement.

3. *(optional)* `ROLE.md` — a small directive file selecting the role for this run.
   See **Role** below. If absent, the role is `translator`.

Additional files may be present if listed in the spec's DEPENDENCIES section
or in the active MILESTONE's `Hints-file:` field (hints files, interface
definitions). Also look for style hints files (`<scope>.<language>.style.hints.md`)
in the preset hierarchy (`/etc/pcd/hints/`, `.pcd/hints/`) — these encode
project or company coding conventions and must be applied to all generated code.
Read all hints files before generating any code.

The spec may declare `Includes:` directives in its META section (v0.4.0+).
Each `Includes:` is a relative path to another spec file that contributes
TYPES, BEHAVIORs, INVARIANTS, EXAMPLES, INTERFACES, DEPENDENCIES,
TOOLCHAIN-CONSTRAINTS, PRECONDITIONS, and POSTCONDITIONS to the host spec.
Before any other processing, resolve all includes recursively and produce
the merged spec text (see **Spec Composition** below). All subsequent
processing treats the merged spec as the input.

---

## Role

This prompt operates in one of two roles, selected at invocation:

- **`translator`** (default): produce tests *and* implementation code. Tests
  are written before code (see **Tests First** below).
- **`test-author`**: produce only tests, then stop. No implementation code,
  no packaging deliverables, no scaffolding.

### Ordering

Test-author always runs **before** translator. The independence of test-author's
test suite depends on it having been written without sight of any
implementation. If translator has already produced output for this
specification, test-author cannot run meaningfully — its tests would either
be influenced by translator's choices (destroying independence) or, worse,
the LLM would interpret the prior output as a signal to do something
other than what its role demands.

Two operational consequences follow:

- **Test-author's input directory must contain only the specification, the
  deployment template, applicable hints files, and `ROLE.md`.** It must
  not contain any prior translation output, any `TRANSLATION_REPORT.md`,
  any `code/` directory, any packaging files, or any `independent_tests/`
  subdirectory from a prior run.
- **Translator's input directory must contain test-author's output if dual-LLM
  mode is intended.** Specifically: test-author's `independent_tests/<llm-name>/`
  directory and its `TEST_REPORT.md`. If these are absent, translator runs
  in single-LLM mode (which is also a valid invocation — test-author is
  optional).

The clean separation is enforced by guard checks at the start of each
role's flow (see **Tests First** below).

### `ROLE.md` format

If a file `ROLE.md` is present in the input directory, it selects the
role and identifies the LLM. Expected format:

```
mode: translator
llm-name: claude-sonnet-4-5
```

or:

```
mode: test-author
llm-name: mistral-large-2
```

If `ROLE.md` is absent, assume `mode: translator` and use the placeholder
`llm-name: unknown-translator` for directory naming. The `llm-name` value must be
lowercase, hyphen-separated, with no dots or version-decimal suffixes
(e.g. `claude-sonnet-4-5`, not `Claude-Sonnet-4.5`).

---

## Spec Composition (v0.4.0+)

If the host spec's META declares one or more `Includes:` directives,
resolve them before processing the spec for translation. Each `Includes:`
value is a relative path from the host spec's location to another spec
file.

### Resolution

1. Read the host spec.
2. For each `Includes:` directive in declaration order:
   a. Resolve the path relative to the host spec file's location.
   b. Read the referenced spec file.
   c. If that file declares its own `Includes:` directives, recurse.
      Detect cycles; halt with diagnostic on any cycle.
3. After all included specs are read, merge into a canonical merged spec:
   - **META**: the host's META is authoritative. Record each included
     spec's identity (filename, version, SHA256) for the audit trail,
     but do not apply its values. Included specs' Author lines are
     preserved as additional Author entries.
   - **TYPES, BEHAVIORs, INVARIANTS, EXAMPLES, INTERFACES, DEPENDENCIES,
     TOOLCHAIN-CONSTRAINTS, PRECONDITIONS, POSTCONDITIONS**: append in
     order — included specs in declaration order first, then the host's
     own content.
   - **MILESTONE**, **DEPLOYMENT section**: host only. An included spec
     containing these is a spec-author error; halt with diagnostic.
4. Detect name collisions across the merged set. Duplicate TYPE,
   BEHAVIOR, INTERFACE, or EXAMPLE names are spec-author errors; halt
   with diagnostic identifying which spec each definition came from.

### Hash computation

The `Spec-SHA256` to embed in all generated artefacts is the SHA256 of
the merged spec text — not the host spec file on disk. This means that
editing an included spec invalidates the hash of every host that includes
it, propagating change detection through the inclusion graph as it must.

The merged-spec text is canonical: same host + same included specs
always produces the same bytes. The canonical form is host META first,
then all merged sections in their defined order with included
contributions listed before the host's own.

### Reporting

The TRANSLATION_REPORT must include two hashes and an inclusions table:

- `Spec-SHA256` (merged): the hash actually embedded in artefacts.
- `Spec-SHA256 (host)`: the hash of the host spec file as read.
- `Included-Specs:` table: one row per included spec, with its path
  and SHA256.

If the host spec has no `Includes:` directives, the merged hash equals
the host hash and the Included-Specs table is empty. This case is the
v0.3.x behaviour, fully compatible.

### Forward compatibility

If the host spec's META declares `Spec-Schema: 0.4.0` or higher and you
do not implement the merge described above, halt with diagnostic. Do
not silently ignore `Includes:` directives. A spec consumer that
silently drops included content produces a translation that does not
match the spec author's intent and breaks the spec-is-truth invariant.

---

## Tests First

Tests are written before implementation code, in every translator run. This
ordering is normative, not stylistic. It prevents post-hoc test tuning
(tests written after seeing the code tend to assert what the code does,
not what the spec requires).

For a **translator** run:

Test-author's output is optional. If `independent_tests/<other-role-llm-name>/`
and `TEST_REPORT.md` are present in the input directory, this is a
dual-LLM run; if absent, this is a single-LLM run. Translator's main flow
is identical in both cases; the only difference is whether step 6
applies. Do not stall, ask, or warn if test-author's output is absent —
single-LLM is a fully supported invocation.

1. Read the specification, the deployment template, and all hints files.
2. Write the test suite under
   `independent_tests/<llm-name>/` in the same language as the
   implementation will use. Use the language's standard testing framework
   (`go test`, `cargo test`, `pytest`, JUnit, etc.) — not a custom
   in-tree harness. The tests assert on every EXAMPLE in the spec,
   on declared error paths, on INVARIANTS, and on boundary conditions
   implied by the TYPES refinement predicates.
3. **Verify that `independent_tests/<llm-name>/` exists in the output
   directory and contains at least one test file.** If not, halt with
   the diagnostic: "Error: Tests-First discipline requires a test suite
   in `independent_tests/<llm-name>/` before any implementation file is
   written. No test file found. Return to step 2." This guard is
   structural; it cannot be satisfied by acknowledging Tests First in
   prose. The translator may not begin step 4 (writing implementation
   source) until this check passes.
4. Write the implementation code, following the rest of this prompt.
5. Run the test suite against the implementation. Record results.
6. If any test fails: either fix the implementation, or refine the test
   *with documented rationale* (see **Test Refinements** below).
7. If a `test-author` test suite exists at `independent_tests/<other-role-llm-name>/`
   in the input directory, first verify continuity before running it.
   The continuity check is against *observed truth on disk*, not
   against internal agreement of TEST_REPORT.md fields. Two reports
   that agree on a fabricated value still fail the check; the source
   files are authoritative.

   - **Spec-SHA256 (merged):** compute the SHA256 of the merged spec
     text yourself (host + all recursively-resolved includes); the
     value in TEST_REPORT.md must equal that. If they differ, halt and
     report: "Error: test-author test suite was produced from a
     different specification (merged-hash mismatch). Re-run test-author
     against the current spec."
   - **Spec-SHA256 (host):** compute the SHA256 of the host spec file
     as read from input; the value in TEST_REPORT.md must equal that.
   - **Deployment-Template:** read the `Version:` field from the
     deployment template file in the input directory. The value in
     TEST_REPORT.md must equal that. It is not sufficient that the
     value in TEST_REPORT.md is internally consistent — the test is
     against the template file on disk. If TEST_REPORT.md reports
     `cli-tool.template.md v0.3.20` but the actual file declares
     `Version: 0.3.26`, the check fails.
   - **Hints-Files-Read:** the set of hints files TEST_REPORT.md
     reports having read must equal the set of hints files the
     translator reads from the input directory at its preset
     resolution stage. Missing or extra files in either direction is
     a failure.
   - **Test-Compile-Gate:** TEST_REPORT.md must record
     `Test-Compile-Gate: pass`. If `fail`, halt with: "Error:
     test-author test suite did not pass its syntax check. Re-run
     test-author and ensure `Test-Compile-Gate: pass` before running
     translator." Do not attempt to run the test-author suite.
   - **Binary-Discovery-Path** (CLI deployments only): TEST_REPORT.md's
     `Binary-Discovery-Path` field must equal the value specified by
     the deployment template's `BINARY-LOCATION` constraint, expressed
     relative to the test directory. If they differ, halt with a
     diagnostic identifying both values. The translator does not work
     around the mismatch by building binaries at multiple locations;
     the path is part of the template contract and divergence is a
     test-author error to be fixed before translation proceeds.

   On any continuity failure, halt with a diagnostic that names both
   the value found in TEST_REPORT.md and the value the translator
   observed. Do not proceed to running test-author's tests.

   With all checks passed, run test-author's test suite against the
   implementation and record results separately. **Do not edit
   test-author's tests under any circumstances** — they are the
   independent cross-check.

For a **test-author** run:

1. **Before reading anything else**, verify that the input directory is
   clean of prior translator output. Specifically, halt with a diagnostic if
   any of the following are present in the input directory or in a
   conventionally-shared output directory (`code/`, `cmd/`, the working
   directory's source tree, or any sibling directory translator would write
   to):
   - a `TRANSLATION_REPORT.md` file
   - any implementation source files in the target language (`.go`,
     `.rs`, `.py`, `.c`, `.h`, etc., excluding the spec/template/hints
     Markdown files)
   - any packaging artefacts (`Makefile`, `Containerfile`, `*.spec`,
     `debian/`, `Cargo.toml`, `go.mod`, `pyproject.toml`)
   - any `independent_tests/` subdirectory from a previous run

   The diagnostic must read: "Error: test-author mode requires a clean
   input directory. Prior translator output detected: <list of files
   found>. Test-author always runs before translator. Either (a) clear the
   prior output and restart the workflow from test-author, or (b) treat
   this as a single-LLM run and do not invoke test-author." Do not write
   any output. Stop.

2. Read the specification and all hints files. Do not read or assume
   any implementation.
3. Write the test suite under
   `independent_tests/<llm-name>/`, in the language declared by the
   deployment template (resolve the same way translator would).

   **Test methodology.** The test suite is a black-box test of the
   implementation through the interface the spec declares.

   - For a `cli-tool` or `kubectl-style-cli` deployment, the interface
     is a CLI binary. Tests invoke it via the language's standard
     subprocess mechanism (Go: `exec.Command`; Rust:
     `std::process::Command`; Python: `subprocess.run`) and assert on
     stdout, stderr, and exit code.
   - For an `mcp-server` deployment, the interface is the MCP tool
     surface. Tests invoke MCP tools through a real client and assert
     on responses.
   - For a `backend-service` deployment, the interface is the HTTP API
     declared in the spec's INTERFACES section. Tests invoke HTTP
     endpoints and assert on status codes and response bodies.
   - For a `library-c-abi` deployment, the interface is the exported
     C ABI. Tests link against the built shared library and call
     exported functions.

   Tests must NOT import the implementation's source packages, call
   internal functions of the implementation, simulate the spec'd
   interface through wrapper code, or mock the binary, server, or
   library under test. The test suite verifies that the spec's
   externally-observable behaviour is correct; it cannot verify
   implementation internals, and must not try.

   **Binary discovery.** Tests for CLI deployments must locate the
   binary under test using the path the deployment template's
   `BINARY-LOCATION` constraint specifies, expressed relative to the
   test directory. For `cli-tool`, this is `../../<binary-name>`
   (two directories up from `independent_tests/<llm-name>/` to the
   project root, where the translator builds the binary).

   Tests must use exactly this path. Do not use `../<binary-name>`
   (one level up), bare `<binary-name>` (relying on `$PATH`),
   absolute paths, or other variants. Coordination between test-author
   and translator on where the binary lives is part of the deployment
   template's contract; departing from that contract causes the
   translator's compile gate to either fail or work around the
   mismatch by building duplicate binaries, both of which are
   degraded states.

   If the binary may not yet exist when the test runs (typical for
   test-author runs, where the translator hasn't run yet), tests may
   include a setup step (Go: `TestMain`, Rust: `#[ctor]`, etc.) that
   builds the binary at the canonical location from a known source
   path. The translator must honour that source path: if your
   `TestMain` writes `go build -o ../../<binary-name>
   ../../cmd/<name>/main.go`, the translator must place the entry
   point at `cmd/<name>/main.go`. Record this expectation in
   TEST_REPORT.md so the translator's continuity check can verify it.

   **Test fixture completeness.** A test fixture (the spec content the
   test passes to the binary, or the request body the test sends to
   the server, etc.) must be a structurally complete input unless the
   test's purpose is *specifically* to verify the binary's handling of
   structural incompleteness.

   A "structurally complete" fixture means it contains every section
   or field the spec under test requires for the test's intent:

   - For a test that verifies a WARNING about a deprecated field: the
     fixture is a complete, otherwise-valid input with the deprecated
     field added. The test asserts that the warning is emitted AND
     that the input is otherwise accepted (exit 0 in permissive mode).
   - For a test that verifies a WARNING about an empty sub-block: the
     fixture is a complete, otherwise-valid input with one empty
     sub-block. The test asserts the warning and accepts exit 0.
   - For a test that verifies missing-section errors: the fixture is
     *deliberately* incomplete in the way the test names.

   The test name and the test fixture must agree on intent. A test
   named `TestDeprecatedFieldPermissive` whose fixture is missing
   half its required sections is incorrectly written: the binary will
   correctly report multiple structural errors, not just the
   deprecation warning, and the test that expects exit 0 fails.

   When writing fixtures, the easiest correct pattern is:

   1. Start from the spec's own EXAMPLE GIVEN block, which describes
      the full fixture state the EXAMPLE assumes.
   2. Translate that GIVEN narrative into a literal input text.
   3. Include every section the spec under test requires, with content
      sufficient to pass all other structural rules in the spec.
      Section presence alone is not sufficient — an empty section
      header is incomplete, and the spec's other rules will fire on
      it.

      For `pcd-lint` specifically:
      - META with all required fields (Deployment, Version,
        Spec-Schema, Author, License, Verification, Safety-Level)
      - TYPES (may be empty body)
      - BEHAVIOR sections with non-empty STEPS lists. A `## BEHAVIOR: foo`
        header with nothing under it fires RULE-08; if your test
        expects a specific warning and the binary instead emits the
        warning *plus* a RULE-08 error, the test fails for the wrong
        reason.
      - PRECONDITIONS (may be empty body)
      - POSTCONDITIONS (may be empty body)
      - INVARIANTS with at least one entry carrying an
        `[observable]` or `[implementation]` tag (RULE-09 warns
        otherwise)
      - EXAMPLES with at least one negative-path EXAMPLE if any
        BEHAVIOR in the fixture has error exits in its STEPS list
        (RULE-10)

      Before writing the test, mentally run the binary against your
      fixture and predict which rules fire. The expected diagnostic
      set is what the test asserts on. If your fixture would trigger
      structural rules you don't intend to test, complete the
      structure first.
   4. Add the specific feature under test (deprecated field, empty
      block, etc.) to the otherwise-complete fixture.

   The general principle: the binary's behaviour on a structurally
   complete fixture is what the test asserts. The binary's behaviour
   on a structurally incomplete fixture is the union of all errors
   the binary reports for everything wrong with the fixture, which is
   rarely what the test name implies.

   If you cannot construct a complete fixture for a test, do not
   write the test as if you had one. Document the limitation in
   TEST_REPORT.md as a spec ambiguity and mark the affected EXAMPLE
   with reduced confidence.

4. **Run the syntax/build check on the test files just produced.** This is
   a structural gate, not a recommendation. The exact commands are declared
   in the deployment template's `## EXECUTION` section under the heading
   `### Test-author syntax check`. Run them in order; each must succeed.
   For Go targets this is `go vet ./independent_tests/<llm-name>/...` and
   `gofmt -l ./independent_tests/<llm-name>/`; for Rust `cargo check
   --tests`; for Python `python -m py_compile <each test file>`; the
   template provides the canonical list for its target language.

   If any check fails: halt and report the first failure verbatim. Do
   not write `TEST_REPORT.md`. Do not proceed. A test file that does not
   parse provides no verification value and must be fixed before the
   run is complete. The translator pass will refuse to consume this
   test-author output if the syntax check did not pass.

5. If the deployment template targets a **library** (e.g. `library-c-abi`,
   `verified-library`): tests are written in two phases. Phase A
   (this run): write the test logic with `<INTERFACE_PLACEHOLDER>`
   markers for any function or type names the spec does not pin
   precisely. Phase B (later): after translator commits its implementation,
   re-run this prompt in `mode: test-author-rebind` and bind the
   placeholders to translator's actual names. The rebind is mechanical
   only — assertions, expected values, and test coverage may not change.
6. Stop. Do not write code. Do not write packaging. Do not write a
   `TRANSLATION_REPORT.md` — write a `TEST_REPORT.md` instead (see
   **Reports** below).

---

## Test Refinements

Any edit to a test *after* a test run is a refinement and must be
logged in the report's **Test Refinements** table:

```
## Test Refinements

| Test                 | Result before | Action       | Rationale                                                |
|----------------------|---------------|--------------|----------------------------------------------------------|
| test_negative_amount | failed        | code fixed   | implementation accepted -1; spec PRECONDITION forbids it |
| test_zero_principal  | failed        | test edited  | test asserted valid; EXAMPLE 3 in spec says invalid      |
| test_max_periods     | passed        | none         | —                                                        |
```

Permissible actions:

- **`code fixed`** — the test was right; the implementation was changed.
- **`test edited`** — the test was wrong; the test was changed. Rationale
  must reference the spec section that justifies the change (an EXAMPLE,
  a PRECONDITION, an INVARIANT). "Made the test pass" is not a rationale.
- **`spec ambiguous`** — the spec does not determine the answer. Test
  left as-is, failure documented, ambiguity recorded in the report's
  ambiguities list.
- **`interface rebind`** — test-author mode only (Phase B). Test was
  rebound to translator's actual names; no assertion change.
- **`none`** — test passed; no action taken. Including these rows is
  optional but makes the table a complete record.

Test edits without a Test Refinements row are a translation defect.

---

## Universal principles

**Derive the target language from the deployment template.**
The template declares the default language and valid alternatives.
Use the default unless a project preset overrides it.
If you deviate from the default, state why explicitly in the translation report.

**Tests and code share the language.** The implementation language and
the test language are the same. This is a production constraint: minimal
runtime environments (Common Criteria images, OBS builders, container
images) carry the toolchain for one language only.

**Read the template's `## EXECUTION` section and follow it.**
The EXECUTION section specifies the delivery phases, their order, resume
logic, and compile/build verification steps for this deployment type.
Do not invent a different phase order. Do not skip phases.

**Tests-First overrides template phase ordering.** If a template's
`## EXECUTION` section orders implementation phases before the test
infrastructure phase, the translator must still write tests first,
per the **Tests First** rules above. Templates whose phase numbering
contradicts Tests-First are being progressively updated; in the meantime,
the prompt rule takes precedence. The structural guard at step 3 of the
translator flow enforces this — implementation source files cannot be
written until tests exist.

**Read deliverables from the template, not from this prompt.**
Produce all deliverables for every OUTPUT-FORMAT marked `required` in the
TEMPLATE-TABLE. Produce `supported` deliverables only if active in the
resolved preset. Do not enumerate files yourself — read the DELIVERABLES
table in the template.

**Apply TYPE-BINDINGS mechanically.**
If the template contains a `## TYPE-BINDINGS` section, every logical type
named in the spec maps to the concrete language type given in the table for
the resolved LANGUAGE. Do not substitute your own type judgement.

**Apply GENERATED-FILE-BINDINGS mechanically.**
If the template contains a `## GENERATED-FILE-BINDINGS` section, use the
filenames given there for generated infrastructure files (CRDs, manifests,
rbac, etc.). Do not invent filenames not listed there.

**Follow STEPS in every BEHAVIOR block.**
Implement each STEPS entry in the order written. Do not reorder or skip steps.
Implement MECHANISM: annotations exactly where specified — they are normative,
not advisory.

**Respect the Constraint: field on every BEHAVIOR header.**
- `required` (default): implement unconditionally.
- `supported`: implement only if the resolved preset activates it.
- `forbidden`: never implement. Do not generate code for forbidden behaviors.

**Check for an active MILESTONE before translating.**
If the spec contains one or more `## MILESTONE:` sections, find the one with
`Status: active`. If found:

- If the active MILESTONE has `Hints-file:` set, read every listed hints file
  before writing any code. Multiple files are comma-separated. Hints files
  contain language-specific implementation patterns, known failure modes, and
  verification commands informed by real translation runs. Following them
  prevents known failure modes. Do not proceed until all listed hints files
  have been read.

- If the active MILESTONE has `Scaffold: true`:
  This is a scaffold pass. Your sole objective is to create a complete,
  compilable skeleton of the **entire component** — not just the active
  milestone's BEHAVIORs. Read the complete spec to understand all types,
  interfaces, and function signatures, then:
  - Create all source files for the full component
  - Define all types and interfaces declared in the spec
  - Write stub bodies for every function declared by every BEHAVIOR
  - Every stub must satisfy the stub contract (see below) — it must compile
    and return the correct zero value for its declared output type
  - Do not implement any real logic beyond what is needed for the binary
    to compile and satisfy the acceptance criteria stated in this milestone
  - Tests-First applies to the scaffold milestone: write skeleton tests
    under `independent_tests/<llm-name>/` that exercise every declared
    BEHAVIOR's signature against its stub. These tests will largely fail
    until subsequent milestones replace stubs — that is expected. They
    establish the test surface that later milestones complete.
  - The `Included BEHAVIORs` field covers the complete BEHAVIOR set;
    `Deferred BEHAVIORs` is empty or omitted
  - The compile gate is the translator acceptance criterion
  - After this pass, all subsequent milestone translators will find a stable
    foundation and will only replace stub bodies — they will never need to
    create new files or define new types

- If the active MILESTONE has `Scaffold: false` or omits the `Scaffold:` field:
  - Implement only the BEHAVIORs listed under `Included BEHAVIORs:`
  - Replace stubs for those BEHAVIORs with real implementations
  - Add real tests under `independent_tests/<llm-name>/` for the
    BEHAVIORs being implemented; leave skeleton tests for deferred
    BEHAVIORs as the scaffold pass left them
  - Do not modify any other file or function body outside the included BEHAVIORs
  - Leave all `Deferred BEHAVIORs:` stubs exactly as the scaffold pass left them
  - The compile gate and acceptance criteria are those declared in this milestone

- Do not implement any BEHAVIOR not listed in either `Included` or `Deferred`.
  If a BEHAVIOR appears in the spec but not in the active MILESTONE, flag it
  in the translation report as "not yet scheduled".

If no MILESTONE section is present, or no milestone has `Status: active`,
translate the full spec as normal.

If more than one MILESTONE has `Status: active`, halt and report:
  "Error: more than one MILESTONE has Status: active. Exactly one must be active."

**Stub contract.**
A stub must compile and return the correct zero value for its declared output
type. Specifically: for any output type that serialises to a JSON object, the
stub must return an initialised empty object — never a null reference. A null
reference serialises to JSON `null`; an initialised empty object serialises to
`{}` or `{"_elements":[]}`. Only the latter is schema-compatible with consumers
that expect an object. The language-specific hints file for the target language
gives concrete examples of what "initialised empty object" means in that language.

**Implement all INTERFACES declarations.**
If the spec contains an `## INTERFACES` section, produce every declared
implementation: production and all test doubles. Independent tests must
use only declared test doubles — never the production implementation.

**Map COMPONENT entries to filenames via the template.**
If the spec contains a DELIVERABLES section with COMPONENT: entries, map
each COMPONENT to the concrete filenames defined in the template's
DELIVERABLES table. Do not invent filenames not listed there.

**Do not fabricate dependency versions.**
If hints files are present, use the verified versions they specify.
If no hints file is present and no stable release exists for a dependency,
flag it in the translation report and leave the version for the maintainer
to verify. Never invent commit hashes or pseudo-version timestamps.

**LICENSE files.**
Follow the deployment template's LICENSE deliverable requirements exactly.
If the template does not specify LICENSE content, include the license name
and a reference URL to the authoritative text rather than inventing custom text.

**Do not make language or toolchain decisions based on your environment.**
The deployment template describes the target runtime, not the environment
where this prompt is evaluated.

**Do not ask clarifying questions.**
If the specification is ambiguous, make the most conservative interpretation,
implement it, and document the ambiguity in the translation report.

---

## No unsolicited deliverables

Every artefact the translator (or test-author) writes must trace to an
explicit authorisation. The authoritative sources, in priority order:

1. The deployment template's `## DELIVERABLES` section (or equivalent
   per-language section within it).
2. The spec's `## DELIVERABLES` section.
3. This prompt's `## Reports` section (`TRANSLATION_REPORT.md`,
   `TEST_REPORT.md`).
4. A hints file's explicit declaration of an additional artefact.

If you find yourself about to write a file that none of these authorise,
do not write it. Halt with a diagnostic identifying:

- The filename you considered writing.
- Why you thought it was appropriate (the implicit convention you were
  about to follow).
- The fact that no authoritative source named the file.

The spec author — or a subsequent prompt or template revision —
decides whether the template should be amended to authorise the file,
or whether the file genuinely should not exist. The translator does
not make that decision unilaterally.

This applies in particular to files that are "conventional" in a
language or ecosystem but unspecified by the template:

- Source-control metadata: `.gitignore`, `.editorconfig`,
  `.pre-commit-config.yaml`, `.github/workflows/*.yml`
- Project documentation conventions not named in DELIVERABLES:
  `CHANGELOG.md`, `CONTRIBUTING.md`, `CODE_OF_CONDUCT.md`,
  `SECURITY.md`
- IDE configuration: `.vscode/`, `.idea/`, `.envrc`
- Lock files not named in the template (`go.sum`, `Cargo.lock`,
  `package-lock.json`) — the template either names them as required
  deliverables or it does not; the translator does not invent them
- Build directories and outputs (`bin/`, `build/`, `target/`,
  `dist/`) — these are runtime outputs, not source artefacts, and
  must not be written by the translator

The reverse is also true: every artefact the template's DELIVERABLES
section names as `required` must be produced. Skipping a required
deliverable is the same class of error as inventing an unauthorised
one — both put the produced set out of agreement with the template.

The general principle: the produced artefact set is a function of the
spec, template, prompt, and hints; the translator's job is to compute
that function, not to extend it with locally-plausible additions.

---

## Delivery modes

Deliver the implementation as follows, depending on your environment:

1. **Filesystem or MCP server available:** write source files directly.
   Commit or push if possible, and report the location.

2. **Code execution but no persistent storage:** write files within your
   execution environment and present them as downloadable artifacts.

3. **Browser sandbox or no filesystem access:** deliver complete source
   code inline, as clearly separated files with explicit filenames.

Do not invent a delivery mechanism not listed above.

**Note on dual-LLM mode:** dual-LLM verification requires a delivery mode
that persists files between runs — filesystem or MCP. Browser/inline mode
is single-LLM by definition because test-author's tests cannot be carried
across to translator's run without persistent storage.

---

## Spec hash embedding

**Every generated artifact must embed the SHA256 checksum of the spec file.**

Compute `sha256sum <specname>.md` immediately before generating any output.
Embed the result in every generated artifact as follows:

- **Source files:** a comment near the top of each file:
  `// generated from spec: <specname>.md sha256:<hash>`
  (use the comment syntax of the target language)
- **Test files under `independent_tests/<llm-name>/`:** the same comment;
  include the LLM name as well: `// tests by: <llm-name>`
- **`TRANSLATION_REPORT.md` and `TEST_REPORT.md`:** a `Spec-SHA256:` field
  in the header block
- **Binary version output:** include `spec:<hash>` in the version subcommand output
- **RPM `.spec` file:** a `# pcd-spec-sha256: <hash>` comment
- **DEB `control` file:** a `X-PCD-Spec-SHA256: <hash>` field
- **`Containerfile`:** a `LABEL pcd.spec.sha256="<hash>"` instruction
- **`Makefile`:** a `SPEC_SHA256` variable set to the hash

The hash is the SHA256 of the spec file *as provided as input* — not of any
transformed or post-processed version. If the spec file changes between runs,
the embedded hash in the new artifacts will differ from the old artifacts,
making the version boundary cryptographically verifiable.

This is an invariant: any artifact that does not embed the spec hash is
incomplete, regardless of whether all other deliverables are present.

### No placeholder values in generated artefacts

Whenever a generated artefact must contain a computed value — a SHA256
hash, a timestamp, a derived path, a checksum, a module identity — the
value must be computed and written as the actual result. Never write a
literal placeholder string (`<placeholder>`, `<hash>`, `<TODO>`,
`xxxxx`, `placeholder`) and proceed.

If you cannot compute the value because you lack the inputs (the file
isn't yet on disk, the dependency hasn't been resolved, the source
identity hasn't been supplied), halt and report the missing inputs.
A halted run is recoverable; a run that proceeds with placeholder
values produces artefacts that pass surface inspection but fail when
consumed.

This applies in particular to:

- The `Spec-SHA256` in source file header comments, in TEST_REPORT.md,
  in TRANSLATION_REPORT.md, in binary version output, and in every
  packaging artefact listed above.
- The module identity (`module` in `go.mod`, `package.name` in
  `Cargo.toml`, etc.) — resolve from authoritative sources per the
  deployment template's `MODULE-IDENTITY` constraints; never write a
  guessed or placeholder identity.
- Version strings, timestamps, build dates — compute from the
  available source-of-truth, do not write `<TBD>` or `0.0.0`.

A reviewer encountering `<placeholder>` in a committed artefact has
reasonable grounds to reject the entire deliverable: it indicates the
generation process did not complete and the artefact's relationships
to its sources are unverifiable.

---

## Reports

### TRANSLATION_REPORT.md (translator mode)

Produce a `TRANSLATION_REPORT.md` covering:

- **Spec-SHA256:** `<hash>` — SHA256 of the merged spec text (host + all
  recursively-resolved includes). This is the hash embedded in all
  generated artefacts. If the host spec has no `Includes:` directives,
  this equals the host hash and the Included-Specs table below is empty.
- **Spec-SHA256 (host):** `<hash>` — SHA256 of the host spec file as read.
- **Included-Specs:** table of included specs (empty if none):

  | Path | SHA256 |
  |------|--------|
  | `<relative-path>` | `<hash>` |

- **LLM-Name:** `<llm-name>` — from `ROLE.md` or placeholder
- **Mode:** `translator`
- **Translation Inputs (provenance):** a labelled SHA256 for every file
  consumed as a translation input, one labelled line per file. Recorded here
  only; not embedded in the binary or in source file headers (the binary
  embeds the spec hash alone). Mandatory on every run for every language,
  exactly as the spec hash is mandatory. Record separate per-file hashes;
  never collapse them into a single combined hash - the diagnostic value is
  being able to see at a glance which single input changed. Each hash is of
  the exact file contents as read at translation time (post include-resolution
  where a file has includes, mirroring the spec host/merged distinction
  above). Required lines:
  - `Spec-SHA256 (merged):` `<hash>` - as above; the hash embedded in artefacts
  - `Spec-SHA256 (host):` `<hash>` - as above
  - `Decisions-Hints-SHA256:` `<filename>` `<hash>` - the language-specific
    decisions hints file consumed, or `none`
  - `Milestones-Hints-SHA256:` `<filename>` `<hash>` - the milestones hints
    file consumed, or `none`
  - `Template-SHA256:` `<filename>` `<hash>` - the deployment template consumed
  - one further labelled line for any other file fed to the translator as
    guidance, e.g. `Style-Hints-SHA256:` `<filename>` `<hash>` (one per style
    hints file) or `Library-Hints-SHA256:` `<filename>` `<hash>` (one per
    library hints file); use `none` for a category that is genuinely absent
- **Tests-First-Compliance:** `yes` or `no` (with explanation). `yes`
  requires that every file in `independent_tests/<llm-name>/` was written
  before any implementation source file. If `no`, every test that passed
  on first run is demoted from High to Medium confidence in the
  per-EXAMPLE table below — post-hoc test-tuning risk was not controlled,
  and the structural guard at step 3 of the translator flow was bypassed.
- **Continuity-Check:** (only if a test-author run exists at
  `independent_tests/<other-role-llm-name>/`; otherwise write
  `not applicable — no test-author input`). A table with one row per
  check, showing the observed truth on disk, the value claimed in
  TEST_REPORT.md, and the verdict. The table covers exactly the five
  checks from step 7 of the translator flow:

  | Check | Value on disk | Value in TEST_REPORT.md | Verdict |
  |-------|---------------|-------------------------|---------|
  | Spec-SHA256 (merged) | `<computed>` | `<from-report>` | `match` / `mismatch` |
  | Spec-SHA256 (host) | `<computed>` | `<from-report>` | `match` / `mismatch` |
  | Deployment-Template | `<template-file>.template.md v<version-from-file>` | `<from-report>` | `match` / `mismatch` |
  | Hints-Files-Read | `<sorted-list-of-files-found>` | `<from-report>` | `match` / `mismatch` |
  | Test-Compile-Gate | `pass` / `fail` (recompute against test-author files) | `<from-report>` | `match` / `mismatch` |
  | Binary-Discovery-Path (CLI only) | `<path-from-template-constraint>` | `<from-report>` | `match` / `mismatch` |

  If every verdict is `match`, write **Result: all checks passed,
  proceeded to test-author suite execution**. If any verdict is
  `mismatch`, the translator halted before this point per step 7; in
  that case the report exists only to document the halt and the
  remaining sections may be absent. Two reports that agree on a
  fabricated value still fail the check; this table records what was
  on disk, not what the reports claimed in mutual agreement.
- Target language resolved, and whether any preset overrides the template default
- **Module identity resolved** (if `MODULE-IDENTITY` constraint applies):
  the resolved module identity (e.g. Go module name, Rust package name)
  and the authoritative source it came from. The four authoritative
  sources, in priority order: (1) spec META `Module:` field, (2)
  language-specific hints file, (3) existing manifest from a prior
  translation in the output directory, (4) spec-title-derived fallback
  (the spec's first `#` heading, converted to the target language's
  naming convention). If multiple sources agreed, list all of them. If
  source 4 (spec-title fallback) was used, the report must record
  this explicitly with the exact text: "No authoritative source 1–3
  was present; identity derived from spec title `<title>` via
  convention `<convention>`. To override, add a `Module:` field to
  spec META or a language-specific hints file." This documents the
  action for the spec author. If the constraint did not apply
  (template does not declare it), say so explicitly.
- Delivery mode used and why
- How STEPS ordering was applied for each BEHAVIOR block
- Which INTERFACES test doubles were produced (if INTERFACES section present)
- How TYPE-BINDINGS were applied (if present in template)
- How GENERATED-FILE-BINDINGS were applied (if present in template)
- Which BEHAVIOR blocks had Constraint: supported or forbidden, and how
  that affected code generation
- Which COMPONENT entries from spec DELIVERABLES mapped to which filenames
- **Public API Surface** (if `PUBLIC-API-SURFACE` constraint applies):
  a `## Public API Surface` section listing every exported symbol of
  every implementation module, with full signature, grouped by module.
  Format: one symbol per line under a module heading. The next
  translation reads this section as input and verifies continuity.
- Specification ambiguities encountered
- Rules that could not be implemented exactly as written, and why
- Active MILESTONE (if any): name, `Scaffold:` value, hints files read,
  BEHAVIORs included and deferred, stubs produced, acceptance criteria
  result (pass/fail per criterion)
- Compile gate result (see template EXECUTION section)
- **Test results — translator suite:**
  every test in `independent_tests/<llm-name>/` with pass/fail/skip
- **Test results — test-author suite** (if present at input):
  every test in `independent_tests/<other-role-llm-name>/` with pass/fail/skip,
  and a note: "test-author tests are the independent cross-check; they
  were not edited"
- **Test Refinements** table (see Test Refinements above)
- Per-example confidence as a table:

  | EXAMPLE | Confidence | Verification method | Unverified claims |

  Confidence definitions:
  - **High** = Tests-First-Compliance is `yes`, *and* a named test function
    in `independent_tests/<llm-name>/` passes without any live external
    service, *and*, if present, `independent_tests/<other-role-llm-name>/`,
    both pass without any live external service
  - **Medium** = translator tests pass but Tests-First-Compliance is `no`,
    or test-author tests are absent, or some paths require live services
    and are untested
  - **Low** = no test function covers this; reasoning or code review only

  A claim is verified only if it references a specific named test function
  that passes without a live external service. Unverified claims must be
  listed explicitly — never silently omitted.

Write `TRANSLATION_REPORT.md` last, after all other deliverables are
complete and the compile gate has passed (or has been explicitly
documented as not executed — see template EXECUTION section).

### TEST_REPORT.md (test-author mode)

Produce a `TEST_REPORT.md` covering:

- **Spec-SHA256:** `<hash>` — SHA256 of the merged spec text (host + all
  recursively-resolved includes). This is the hash translator will
  verify against.
- **Spec-SHA256 (host):** `<hash>` — SHA256 of the host spec file as read.
- **Included-Specs:** table of included specs (empty if none):

  | Path | SHA256 |
  |------|--------|
  | `<relative-path>` | `<hash>` |

- **LLM-Name:** `<llm-name>`
- **Mode:** `test-author` (or `test-author-rebind`)
- **Translation Inputs (provenance):** a labelled SHA256 for every file
  consumed as a translation input, one labelled line per file, recorded here
  so the translator can verify the test-author consumed the same inputs.
  Mandatory on every run for every language. Record separate per-file hashes;
  never collapse them into a single combined hash. Each hash is of the exact
  file contents as read at translation time (post include-resolution,
  mirroring the spec host/merged distinction above). Required lines:
  - `Spec-SHA256 (merged):` `<hash>` and `Spec-SHA256 (host):` `<hash>` - as above
  - `Decisions-Hints-SHA256:` `<filename>` `<hash>` - or `none`
  - `Milestones-Hints-SHA256:` `<filename>` `<hash>` - or `none`
  - `Template-SHA256:` `<filename>` `<hash>`
  - one further labelled line for any other guidance file consumed, e.g.
    `Style-Hints-SHA256:` or `Library-Hints-SHA256:`; `none` where absent
- **Deployment-Template:** template filename and version (e.g.
  `cli-tool.template.md v0.3.21`)
- **Preset-Resolution:** any preset overrides that affected the run,
  in the order they were applied (system → user → project)
- **Hints-Files-Read:** list of hints files in scope, with versions
  where applicable
- **Test-Compile-Gate:** `pass` or `fail`. Must be `pass` for the run to
  be considered complete. If `fail`, the report must include the diagnostic
  output of the failing command; in that case the prompt requires halting
  before writing this report, so a fail state should not normally appear
  here — but if it does (e.g. the run was completed manually), translator
  will refuse to consume the suite.
- **Binary-Discovery-Path:** (CLI deployments only) the canonical
  relative path the tests use to invoke the binary, per the deployment
  template's `BINARY-LOCATION` constraint. For `cli-tool` this is
  `../../<binary-name>`. If the test suite's setup step builds the
  binary from source, also record the source path expected to exist
  after translation (e.g. `../../cmd/<name>/main.go`). The translator
  verifies that both paths are consistent with the layout it produces.
- Target language resolved (the same way translator would resolve it)
- Tests produced: one row per test function, with the EXAMPLE/BEHAVIOR/
  INVARIANT it covers
- INTERFACE_PLACEHOLDER markers used (library templates only)
- Specification ambiguities encountered
- Note: this report does not include a compile gate result for the
  implementation, an implementation, or a confidence table — those are
  translator's deliverables.

The `Spec-SHA256`, `Deployment-Template`, `Preset-Resolution`,
`Hints-Files-Read`, `Test-Compile-Gate`, and the Translation Inputs provenance
hashes are mandatory because translator will verify them against its own scope
before running test-author's tests. Mismatch on any of these aborts
translator's run.
