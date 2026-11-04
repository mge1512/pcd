# Mode: spec-review

A review of a specification before translation money is spent. The
output is a list of findings, ordered by what each one costs if the
specification is translated with the finding left in. The specification
is not edited in this mode; the expert applies findings through the
change mode, one unit per run.

## 1. Read

Read the complete specification, its `CHANGELOG.md`, and the deployment
template it names (`templates/<deployment>.template.md`), in that
order. Read the hints files the template and the MILESTONE sections
name. Do not read generated code; a review against the code is the
translation skill's review mode.

## 2. Run the tools

Lint and slice check first, as `SKILL.md` section 4 says. Every lint
error is a finding of class 1; every warning and every slice-check
finding is a finding of its class below. Without an engine, say so at
the top of the findings and review by reading; the pull-request gate
will add what the tool finds.

## 3. Check, by cost class

Class 1 - halts the translation, or produces an artefact the pipeline
rejects:

- A lint error.
- A referenced contract that is not in `contracts/`, or a contract
  reproduced in prose instead of referenced.
- A MILESTONE structure the rules reject: scaffold not first, two
  active, a BEHAVIOR name that does not exist.
- A slice-check finding: a BEHAVIOR without the tag or binding the
  slicer needs, a heading the slicer cannot place.
- License text in the specification.

Class 2 - forces a full regeneration later, after code exists:

- A TYPE without the constraints its uses need (range, precision,
  uniqueness, allowed values), so that each BEHAVIOR will guess them
  differently.
- An INTERFACE whose methods, errors or test double are not stated.
- An INVARIANT that two BEHAVIORs can satisfy in incompatible ways.
- A scaffold milestone whose scope leaves a shared TYPE or INTERFACE
  for a later milestone.

Class 3 - produces wrong behaviour in one BEHAVIOR:

- A STEP without an explicit error exit, or with "handle the error".
- A STEP whose how is normative and is not marked `MECHANISM:`.
- PRECONDITIONS or POSTCONDITIONS that contradict the STEPS or an
  INVARIANT.
- An ERRORS entry no STEP can raise, or a STEP error no ERRORS entry
  names.
- A dependency without `do-not-fabricate: true` where no stable release
  exists.

Class 4 - leaves a gap the independent tests cannot close:

- A BEHAVIOR without a positive EXAMPLE, or with error exits and no
  negative EXAMPLE that names it with `(of: <behavior>)`.
- A multi-step operation without a multi-pass WHEN/THEN example.
- An EXAMPLE whose THEN cannot be verified from the outside (names an
  internal, not an observable).

Class 5 - costs review time, not correctness:

- A programming language, library, file path or function name in
  TYPES, STEPS or EXAMPLES.
- A run-time data file, directory or compiled-in fallback listed as a
  deliverable.
- Terminology that varies for one thing (two names for one TYPE or
  actor).
- A `Spec-Schema:` older than the schema of this skill, with no other
  finding caused by it.

## 4. Write the findings

One finding per line, ordered class 1 first, then by section order
within a class:

    <class> <section>:<line> <what is wrong> -> <what the specification should say>

The arrow part is specification text or a question to the expert, never
code. End with the counts per class and the three findings whose fix
would remove the most cost, named by line. Where the specification is
clean, say so in one line and list the warnings.

## 5. Hand over

As `SKILL.md` section 5, with `Specification:` unchanged, no changelog
entry, `Change impact: not applicable`, and the findings as the body.
