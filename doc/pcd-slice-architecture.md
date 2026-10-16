# pcd-slice architecture

Status: describes pcd-slice 0.1.1. The specification at
tools/pcd-slice/spec/pcd-slice.spec.md is normative; where this
document and the specification disagree, the specification wins.

## 1. The problem: where translation tokens go

A PCD translation reads the specification many times: once per
milestone pass, again for the test-author role, again for each
reviewer. Harnesses read files in mechanical windows - kit's read tool
returns line ranges capped at 2000 lines or 50KB, and the model
chooses the offsets - so a window that cuts a behavior in half or
lacks the types it references forces another window or a guess.
Measured on a production corpus (a planner spec: 273K bytes, 41
behaviors, 46K of changelog, 36K of examples), a full-context pass is
on the order of 70K tokens, multiplied by passes and roles. The
generated code has an irreducible size; the reading does not.

## 2. The decision: derive the semantic form, keep the source

The obvious alternative - reorder the source so each behavior sits
beside its types and examples - fails three ways. Shared types
(references used by dozens of behaviors) cannot sit beside every
consumer without duplication, and duplicated definitions drift. The
source order is doing other work: one file, one Version header, review
by reading top to bottom, an append-only changelog. And reordering
changes nothing about what is transmitted, because the windowing is
mechanical.

So the semantic form is a build output. pcd-slice reads the
specification and its hints files, computes per-behavior bundles
deterministically, and writes them beside the source. The source
remains the durable artifact; the bundles are regeneratable,
hash-anchored, and disposable.

## 3. The consumer contract: preamble plus bundle

A translation pass for behavior X reads exactly two files:

    out/preamble.md     what every behavior shares
    out/X.md            what only X needs

The preamble holds the version line, the introductory prose, the full
TYPES section, the global invariants, and hints content attributable
to no single behavior. The bundle holds the behavior block verbatim,
the examples that exercise it, the invariants bound to it, the hints
blocks written for it, and a Requires-Types line listing its type
closure by name.

The closure is listed, not embedded. Embedding the referenced type
definitions in each bundle would make bundles self-contained, and it
was rejected for two reasons: every embedded copy is a copy that can
drift from the TYPES section, and the preamble's value depends on
being byte-identical across all bundles of one run - a stable prefix
that API-side prompt caching prices at a fraction of fresh input. With
caching, one full TYPES section read cheaply per call beats a
different closure read at full price per call. Without caching the
trade reverses; if that ever becomes the operating mode, revisit this
section first.

## 4. Data flow

Five stages, one pass over the input:

1. Parse. A line classifier with three pieces of state: inside a code
   fence, current section, current behavior. Headings count only at
   column 0 outside fences.
2. Assign. Every source line receives exactly one assignment -
   behavior, preamble, or excluded (the changelog). The completeness
   invariant is this array: outputs are projections of it, so nothing
   is dropped or duplicated by construction.
3. Close. Type definitions are collected from the TYPES fence;
   references are whole-word matches; the closure is transitive over
   definitions.
4. Attribute. A four-rung ladder, first rung wins: nesting; an
   explicit heading suffix `(of: behavior)` or `(of: shared)`; the
   identifier invoked on the WHEN line; a unique whole-word behavior
   name anywhere in the example's own text. Hints blocks attach when
   their heading names a behavior as a whole word. What attaches
   nowhere, and a rung-four tie, is a finding, never a silent drop:
   0.1.1 chose strict classification over shared-by-default, because
   an example that rides in the preamble instead of its bundle, with
   nothing saying so, is the silent-degradation class this project spends review
   rounds hunting. The first translation of 0.1.0 proved the need on
   the target corpus itself - 66 of its top-level
   examples were unattributable under the two-rung grammar - and the
   suffix is the escape hatch for genuinely cross-behavior
   narratives.
5. Emit. All outputs are computed in memory, then written: preamble,
   one file per behavior, MANIFEST.tsv last as the commit marker.
   Every markdown file begins with one provenance comment carrying
   the source path and SHA-256, the hints paths and SHA-256 values,
   and the tool version. No timestamps, no hostnames, no absolute
   paths beyond those given: two runs over identical inputs are
   byte-identical.

Staleness is a hash comparison. `check out=<dir>` recomputes the
provenance values for the current sources and compares them with what
the emitted files carry; any difference is one finding per file. This
is the drift-detection posture of the reproducibility tuple applied to
a derived artifact: the tuple's members are hashed into the output,
so the output can testify about its own currency.

## 5. The grammar boundary

pcd-slice recognizes a deliberately small structure, stated normatively
in the specification's SECTION GRAMMAR: column-0 headings outside
fences, the TYPES fence with `Name := ...` definitions, behavior
blocks to the next `## ` heading, invariant lines with an optional
trailing `(binds: name, name)`, a changelog section excluded from
everything, and everything else as preamble. It covers both dialects
in use - Spec-Schema specifications with a top-level EXAMPLES section,
and the house dialect with examples nested inside behavior blocks.

It does not depend on libpcd. libpcd's model is the Spec-Schema; the
corpus pcd-slice must serve includes specifications that predate or
sidestep that schema, and a parser that recognizes more than the
grammar admits would make the slicing depend on behavior this
specification does not state. The seam is named in DEPENDENCIES:
when the dialects converge, LoadSpec replaces the classifier and this
section shrinks to a pointer.

## 6. What stays out, and why

- Contract shape cards - compact type listings generated from JSON
  schemas so a bundle can reference contract fragments instead of
  whole files - are a separate tool with a separate input grammar.
- Example-to-fixture compilation, which would shrink the test-author
  role, reads the same blocks but writes a different artifact; it can
  share the parser later.
- Serving bundles over MCP (a get_behavior_bundle tool on
  mcp-server-pcd) is the better delivery for harness integration, and
  it is exactly the files this tool emits behind a lookup; the CLI
  comes first because files are inspectable and diffable.
- No configuration file, no environment variables, no network, Linux
  only in v1.

## 7. How-to

### 7.1 Build

    cd tools/pcd-slice && go build ./...

Ships like the other PCD tools: a single static binary from the
project's build service as a signed package. Nothing is fetched at
build or run time.

### 7.2 First look

    pcd-slice list spec=tools/mytool/spec/mytool.spec.md \
              hints=tools/mytool/spec/mytool.go.hints.md

One line per behavior with counts: types in the closure, bound
invariants, attributed examples, attributed hints blocks. Zeros in the
hints column usually mean the hints headings do not name the behavior;
zeros in the examples column on a Spec-Schema spec usually mean the
WHEN lines invoke a different identifier than the behavior heading.

### 7.3 Check before slicing

    pcd-slice check spec=... hints=...

Findings, one per line on stderr, each with a rule name:

    unknown-binding       an invariant binds a behavior that does not
                          exist; usually a typo in the tag
    unassignable-example  a top-level example names no behavior; put
                          the behavior's name on the WHEN line
    unassignable-hints    a hints heading looks like an attribution
                          but names no known behavior
    duplicate-behavior    two behavior headings share a name
    undefined-type        a type definition references a type that is
                          defined nowhere
    no-behaviors          the file has no behavior headings

Exit 0 is clean, 1 is findings, 2 is invocation trouble.

### 7.4 Tag examples that the ladder cannot place

check names every unplaced or tied example. For each: a scenario that
exercises one behavior gets `(of: that-behavior)` on its EXAMPLE
heading; a genuinely cross-behavior narrative gets `(of: shared)` and
lands in the preamble. The edit is mechanical, and each edited
specification moves its version - one row - per the repository's
bytes-move-version rule.

### 7.5 Tag invariants, incrementally

Untagged invariants are global and land in the preamble, so an
untagged specification slices today with no edits. Tightening is one
suffix at a time:

    - Reset destroys nothing unfetched.               before
    - Reset destroys nothing unfetched. (binds: reset)   after

Tag the invariants that constrain one or two behaviors; leave the
genuinely global ones untagged. check catches tag typos immediately.

### 7.6 Slice

    pcd-slice slice spec=... hints=... out=tools/mytool/spec/spec.d

Produces:

    spec.d/preamble.md
    spec.d/<behavior>.md      one per behavior
    spec.d/MANIFEST.tsv       name and sha256 of every file

slice runs check first and writes nothing on findings. It refuses an
out directory containing any file it did not write - move your notes
elsewhere.

### 7.7 Point the harness at it

For the behavior under work, the translation prompt names two files
instead of the specification:

    Read spec.d/preamble.md, then spec.d/build_plan.md.
    Translate the behavior it contains. Do not open the full
    specification; the bundle is complete for this behavior.

Keep the preamble first in every call so its bytes are a stable prefix
for prompt caching.

### 7.8 Keep it current

After any edit to the specification or a hints file:

    pcd-slice check spec=... hints=... out=spec.d

reports stale-output per file; re-run slice. Whether spec.d is
committed is a project choice: committing gives offline harnesses and
reviewable bundle diffs anchored by MANIFEST.tsv; ignoring it keeps
the repository to durable artifacts and regenerates in CI. Either
way the provenance headers make a stale tree detectable, and only the
source specification is ever edited by hand.

### 7.9 When something looks wrong

- A bundle is missing content you expected: run list; if the count is
  zero, the attribution heading or WHEN line does not name the
  behavior.
- Two runs differ: they must not; report it with both MANIFEST.tsv
  files - determinism is an invariant, not a goal.
- exit 2 with "foreign file": the out directory holds a file without
  a pcd-slice provenance line; it is yours, and the tool will not
  touch it.
