# Mode: test-author

Produces the independent test suite for a specification before any
implementation exists, as `modes/translation-procedure.md` prescribes
for its `test-author` role. Its value is independence: the tests are written
without sight of any implementation, so that a translator cannot tune
code and tests to each other.

## Inputs, and nothing else

The input directory contains `spec.d/`, the template, the hints files
in scope and `ROLE.md` with `mode: test-author`. Before reading
anything else, list the input directory. If it contains any of

- implementation source files or a `code/` directory
- packaging files, build files, a `Makefile`
- a `TRANSLATION_REPORT.md`
- an `independent_tests/` directory from an earlier run

STOP with "pcd-translate: test-author input contaminated, refusing" and
name the files. Do not delete them, do not ignore them; the operator
stages a clean directory.

## Procedure

Follow `modes/translation-procedure.md`, test-author role: black-box tests from the
EXAMPLES and BEHAVIORs of each bundle, hermetic, no hollow tests, the
syntax check the template names for the language. Read bundles as
`SKILL.md` section 4 says; one bundle per step.

## Report

`TEST_REPORT.md` as the procedure defines it, with the provenance lines of
`SKILL.md` section 3. The translator verifies these lines against its
own inputs before it runs the suite, and a mismatch aborts the
translator's run. Every value is therefore taken from the files as
read, never typed from memory.
