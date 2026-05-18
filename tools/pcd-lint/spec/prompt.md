

I am providing the following input files, all present in the same
input directory alongside this prompt:

1. cli-tool.template.md — the deployment template defining
   conventions, constraints, defaults, and the full execution recipe for
   this component type.

2. pcd-lint.md — the specification for the component to implement.

3. ROLE.md` — a small directive file selecting the
   translator role for this run. See **Role** below. If absent, the role
   is `primary`.

Additional files may be present if listed in the spec's DEPENDENCIES section
or in the active MILESTONE's `Hints-file:` field (hints files, interface
definitions). Also look for style hints files (`<scope>.<language>.style.hints.md`)
in the preset hierarchy (`/etc/pcd/hints/`, `.pcd/hints/`) — these encode
project or company coding conventions and must be applied to all generated code.
Read all hints files before generating any code.

---

## Role

This prompt operates in one of two roles, selected at invocation:

- **`primary`** (default): produce tests *and* implementation code. Tests
  are written before code (see **Tests First** below).
- **`secondary`**: produce only tests, then stop. No implementation code,
  no packaging deliverables, no scaffolding.

If a file `ROLE.md` is present in the input directory, it selects the
role and identifies the LLM. Expected format:

```
mode: primary
llm-name: claude-sonnet-4-5
```

or:

```
mode: secondary
llm-name: mistral-large-2
```

If `ROLE.md` is absent, assume `mode: primary` and use the placeholder
`llm-name: primary` for directory naming. The `llm-name` value must be
lowercase, hyphen-separated, with no dots or version-decimal suffixes
(e.g. `claude-sonnet-4-5`, not `Claude-Sonnet-4.5`).

---

## Tests First

Tests are written before implementation code, in every primary run. This
ordering is normative, not stylistic. It prevents post-hoc test tuning
(tests written after seeing the code tend to assert what the code does,
not what the spec requires).

For a **primary** run:

1. Read the specification, the deployment template, and all hints files.
2. Write the test suite under
   `independent_tests/<llm-name>/` in the same language as the
   implementation will use. Use the language's standard testing framework
   (`go test`, `cargo test`, `pytest`, JUnit, etc.) — not a custom
   in-tree harness. The tests assert on every EXAMPLE in the spec,
   on declared error paths, on INVARIANTS, and on boundary conditions
   implied by the TYPES refinement predicates.
3. Write the implementation code, following the rest of this prompt.
4. Run the test suite against the implementation. Record results.
5. If any test fails: either fix the implementation, or refine the test
   *with documented rationale* (see **Test Refinements** below).
6. If a `secondary` test suite exists at `independent_tests/<other-llm-name>/`
   in the input directory, first verify continuity before running it:
   - Read `TEST_REPORT.md` (produced by secondary). Confirm its
     `Spec-SHA256` matches the SHA256 of the current spec file. If they
     differ, halt and report: "Error: secondary test suite was produced
     from a different specification (hash mismatch). Re-run secondary
     against the current spec." Do not run secondary's tests against
     the implementation.
   - Confirm the deployment template, preset resolution, and hints
     files listed in `TEST_REPORT.md` match those in scope for this
     run. On any mismatch, halt with the same diagnostic pattern.
   - With both checks passed, run secondary's test suite against the
     implementation and record results separately. **Do not edit
     secondary's tests under any circumstances** — they are the
     independent cross-check.

For a **secondary** run:

1. Read the specification and all hints files. Do not read or assume
   any implementation.
2. Write the test suite under
   `independent_tests/<llm-name>/`, in the language declared by the
   deployment template (resolve the same way primary would).
3. If the deployment template targets a **library** (e.g. `library-c-abi`,
   `verified-library`): tests are written in two phases. Phase A
   (this run): write the test logic with `<INTERFACE_PLACEHOLDER>`
   markers for any function or type names the spec does not pin
   precisely. Phase B (later): after primary commits its implementation,
   re-run this prompt in `mode: secondary-rebind` and bind the
   placeholders to primary's actual names. The rebind is mechanical
   only — assertions, expected values, and test coverage may not change.
4. Stop. Do not write code. Do not write packaging. Do not write a
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
- **`interface rebind`** — secondary mode only (Phase B). Test was
  rebound to primary's actual names; no assertion change.
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

**Read the template's `## EXECUTION` section and follow it exactly.**
The EXECUTION section specifies the delivery phases, their order, resume
logic, and compile/build verification steps for this deployment type.
Do not invent a different phase order. Do not skip phases.

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
  - The compile gate is the primary acceptance criterion
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
is single-LLM by definition because secondary's tests cannot be carried
across to primary's run without persistent storage.

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

---

## Reports

### TRANSLATION_REPORT.md (primary mode)

Produce a `TRANSLATION_REPORT.md` covering:

- **Spec-SHA256:** `<hash>` — SHA256 of `<specname>.md` as provided
- **LLM-Name:** `<llm-name>` — from `ROLE.md` or placeholder
- **Mode:** `primary`
- Target language resolved, and whether any preset overrides the template default
- Delivery mode used and why
- How STEPS ordering was applied for each BEHAVIOR block
- Which INTERFACES test doubles were produced (if INTERFACES section present)
- How TYPE-BINDINGS were applied (if present in template)
- How GENERATED-FILE-BINDINGS were applied (if present in template)
- Which BEHAVIOR blocks had Constraint: supported or forbidden, and how
  that affected code generation
- Which COMPONENT entries from spec DELIVERABLES mapped to which filenames
- Specification ambiguities encountered
- Rules that could not be implemented exactly as written, and why
- Active MILESTONE (if any): name, `Scaffold:` value, hints files read,
  BEHAVIORs included and deferred, stubs produced, acceptance criteria
  result (pass/fail per criterion)
- Compile gate result (see template EXECUTION section)
- **Test results — primary suite:**
  every test in `independent_tests/<llm-name>/` with pass/fail/skip
- **Test results — secondary suite** (if present at input):
  every test in `independent_tests/<other-llm-name>/` with pass/fail/skip,
  and a note: "secondary tests are the independent cross-check; they
  were not edited"
- **Test Refinements** table (see Test Refinements above)
- Per-example confidence as a table:

  | EXAMPLE | Confidence | Verification method | Unverified claims |

  Confidence definitions:
  - **High** = a named test function in `independent_tests/<llm-name>/`
    *and*, if present, `independent_tests/<other-llm-name>/`, both pass
    without any live external service
  - **Medium** = primary tests pass; secondary tests absent, or some
    paths require live services and are untested
  - **Low** = no test function covers this; reasoning or code review only

  A claim is verified only if it references a specific named test function
  that passes without a live external service. Unverified claims must be
  listed explicitly — never silently omitted.

Write `TRANSLATION_REPORT.md` last, after all other deliverables are
complete and the compile gate has passed (or has been explicitly
documented as not executed — see template EXECUTION section).

### TEST_REPORT.md (secondary mode)

Produce a `TEST_REPORT.md` covering:

- **Spec-SHA256:** `<hash>`
- **LLM-Name:** `<llm-name>`
- **Mode:** `secondary` (or `secondary-rebind`)
- **Deployment-Template:** template filename and version (e.g.
  `cli-tool.template.md v0.3.21`)
- **Preset-Resolution:** any preset overrides that affected the run,
  in the order they were applied (system → user → project)
- **Hints-Files-Read:** list of hints files in scope, with versions
  where applicable
- Target language resolved (the same way primary would resolve it)
- Tests produced: one row per test function, with the EXAMPLE/BEHAVIOR/
  INVARIANT it covers
- INTERFACE_PLACEHOLDER markers used (library templates only)
- Specification ambiguities encountered
- Note: this report does not include a compile gate result, an
  implementation, or a confidence table — those are primary's deliverables.

The `Spec-SHA256`, `Deployment-Template`, `Preset-Resolution`, and
`Hints-Files-Read` fields are mandatory because primary will verify them
against its own scope before running secondary's tests. Mismatch on any
of these aborts primary's run.
