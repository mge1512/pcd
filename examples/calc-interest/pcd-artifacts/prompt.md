
I am providing the following input files, all present in the same
input directory alongside this prompt:

1. `<deployment-template>.template.md` — the deployment template defining
   conventions, constraints, defaults, and the full execution recipe for
   this component type.

2. `<spec-name>.md` — the specification for the component to implement.

Additional files may be present if listed in the spec's DEPENDENCIES section
or in the active MILESTONE's `Hints-file:` field (hints files, interface
definitions). Read them before generating any code.

---

## Universal principles

**Derive the target language from the deployment template.**
The template declares the default language and valid alternatives.
Use the default unless a project preset overrides it.
If you deviate from the default, state why explicitly in the translation report.

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
  - The `Included BEHAVIORs` field covers the complete BEHAVIOR set;
    `Deferred BEHAVIORs` is empty or omitted
  - The compile gate is the primary acceptance criterion
  - After this pass, all subsequent milestone translators will find a stable
    foundation and will only replace stub bodies — they will never need to
    create new files or define new types

- If the active MILESTONE has `Scaffold: false` or omits the `Scaffold:` field:
  - Implement only the BEHAVIORs listed under `Included BEHAVIORs:`
  - Replace stubs for those BEHAVIORs with real implementations
  - Do not modify any other file or function body
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

---

## Translation report

Produce a `TRANSLATION_REPORT.md` covering:

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
- Per-example confidence as a table:

  | EXAMPLE | Confidence | Verification method | Unverified claims |

  Confidence definitions:
  - **High** = a named test function in `independent_tests/` passes without
    any live external service
  - **Medium** = some paths tested; other paths require live services or
    are untested
  - **Low** = no test function covers this; reasoning or code review only

  A claim is verified only if it references a specific named test function
  that passes without a live external service. Unverified claims must be
  listed explicitly — never silently omitted.

Write `TRANSLATION_REPORT.md` last, after all other deliverables are
complete and the compile gate has passed (or has been explicitly
documented as not executed — see template EXECUTION section).
