# Mode: change

One change to one existing specification: one BEHAVIOR added, removed
or altered; or one TYPE, one INTERFACE or one INVARIANT changed. The
rules of `SKILL.md` section 3 apply; this file adds the procedure.

## 1. Fix the unit

Restate the request as one unit before touching the file:

    Unit: <BEHAVIOR: name | TYPE: name | INTERFACE: name | INVARIANT: text> - <add | remove | alter>

If the request contains more than one unit, say
"pcd-spec: request bundles several units, asking which first" and ask.
A TYPE change that forces edits in several BEHAVIORs is still one unit.
The TYPE is the unit; the BEHAVIOR edits follow from it and are listed
in the changelog entry.

## 2. Read

Read, in this order and nothing more:

1. `## META`, for Version, Spec-Schema, Includes and Deployment.
2. The section that is the unit.
3. Every TYPE the unit uses, and every BEHAVIOR that uses a TYPE or
   INTERFACE the unit changes (search the file for the name).
4. The EXAMPLES whose heading names the unit's BEHAVIOR with
   `(of: <behavior>)`, or that refer to it.
5. The MILESTONE that lists the BEHAVIOR, if any, and its `Status:`.
6. `CHANGELOG.md` beside the specification, the first three entries.

Do not read the generated code. Where the code did something the
specification did not say, the specification is what changes, and the
expert's words are the source, not the code.

## 3. Ask, limited to the unit

Ask the gap questions of the interview prompt's PHASE 4 to PHASE 6 for
this unit only, one at a time. They cover the inputs and their types,
the preconditions, the steps with an explicit error exit each, the
postconditions, the rules that must hold, and the examples. Before each
question, name the section it belongs to ("For the STEPS of
compute-interest: ..."). End with a summary of what will change and ask
"Is this correct?".

## 4. Edit

- Apply the change to the unit's section.
- Keep the examples current. Every altered BEHAVIOR has at least one
  positive EXAMPLE that exercises the change. Where it has an error
  exit, it also has one negative EXAMPLE whose THEN verifies the error.
  A new BEHAVIOR gets both. A removed BEHAVIOR takes its examples with
  it.
- Mark a step whose how is normative with `MECHANISM:`; otherwise state
  the what.
- A new INVARIANT is tagged `[observable]` or `[implementation]`.
- A BEHAVIOR added to a specification with MILESTONE sections is
  listed under `Included BEHAVIORs:` of a milestone with
  `Status: pending`, or under `Deferred BEHAVIORs:` of the active one.
  Never add to a released milestone: its scope is audit evidence. If
  the change has to touch a released milestone's scope, say so in the
  hand-over. The path is then full regeneration under a new milestone,
  and the maintainer decides.
- Move `Version:` (section 3 of `SKILL.md`: PATCH or MINOR by the
  unit's class).
- Add the `CHANGELOG.md` entry: date, Version, what changed, why, and
  which BEHAVIORs were touched as a consequence of a TYPE or INTERFACE
  change.

## 5. Check

Lint, then slice check, as `SKILL.md` section 4 says. Fix findings in
the specification and re-run until clean, or record "not run" when no
engine is reachable.

## 6. Assess the impact

The change-impact assessment tells the maintainer whether the next
translation is a full regeneration or an incremental update.

- With the MCP server: call `assess_change_impact` with the change
  description, the old and the new specification text, and, when the
  session may read it, the existing code. Record the recommendation and
  the primary factor.
- Without it: apply `prompts/change-impact.md` to the same inputs and
  produce its output format. Say in the hand-over that the assessment
  was made by the session, not by the tool.
- The decision rules, for orientation. Full regeneration: a TYPES,
  INTERFACES or INVARIANTS change, anything that touches the scaffold
  milestone, or a change inside a released milestone's scope.
  Incremental update: a change confined to the STEPS, ERRORS or
  EXAMPLES of one or two BEHAVIORs, with no shared type changed. When
  in doubt, full regeneration. An unnecessary regeneration costs one
  translator run; an inconsistent incremental update costs a
  regeneration later, plus the debugging.

When the recommendation is incremental, the assessment's "what to
change" list is the upgrade brief (`UPGRADE.md`) for the translation
skill's incremental mode. Where that file is staged is stated by that
skill. When the recommendation is full regeneration, the "decisions
hints to preserve" list goes into the hand-over. The maintainer decides
whether it enters a decisions hints file.

## 7. Hand over

The block of `SKILL.md` section 5, plus the pull-request body. The
changelog entry and the change-impact recommendation are quoted
verbatim in the body; the maintainer pastes it.
