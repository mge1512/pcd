# Examples

Two worked examples illustrate different aspects of the PCD paradigm.

[`account-transfer/`](account-transfer/) is a specification-only example.
It shows what a `backend-service` specification looks like when formal
verification is required and the safety level is financial-integrity-critical.
No generated code is included — the point is the specification, not the
implementation.

[`calc-interest/`](calc-interest/) is a full end-to-end migration example.
An original COBOL program is reverse-engineered into a PCD specification,
which is then translated to both Rust and Java. Spec, original COBOL,
two generated implementations, and the prompts used are all in the
directory — everything needed to reproduce or audit the run.

The third worked example, `sitar`, lives in a separate repository at
[github.com/mge1512/sitar](https://github.com/mge1512/sitar). At 35
BEHAVIORs and ~3000 spec lines, it was too large to ship inside `pcd/`
without making the repository structure misleading. Read it after the
two examples here.

---

## account-transfer

A `backend-service` specification for a money-transfer operation between
two accounts. The specification declares `Verification: lean4` and
`Safety-Level: financial-integrity-critical` — a context where the
implementation must be formally verified, not merely tested.

The example demonstrates several PCD features that matter most in
regulated and safety-critical contexts:

- **TYPES with refinement predicates.** `Balance := i64 where balance >= 0`
  is a type, not a precondition. The non-negativity is part of the
  type's identity, and the implementation is expected to enforce it
  at every boundary.
- **PRECONDITIONS, POSTCONDITIONS, INVARIANTS as separate contracts.**
  Money conservation, non-negative balances, and atomicity each appear
  where they semantically belong. Conservation is an INVARIANT (it must
  hold across every state transition); same-account rejection is a
  PRECONDITION (it must hold at entry).
- **Verified ERROR codes.** `INSUFFICIENT_FUNDS`, `SAME_ACCOUNT`, and
  `INVALID_AMOUNT` are declared in TYPES and referenced in BEHAVIOR.
  RULE-10 enforces that a BEHAVIOR with error exits has at least one
  negative-path EXAMPLE.

Reading the file:

```
examples/account-transfer/account-transfer.md
```

This example does not ship generated code. The deployment template is
`backend-service` and `Verification: lean4` requires a formal-methods
toolchain that is out of scope for a repository demonstration.

---

## calc-interest

A complete migration loop from COBOL to two modern languages, via PCD.
The deliverables are organised so that each stage of the workflow is
inspectable independently:

```
calc-interest/
├── cobol/                       ← the original
│   ├── calc-interest.cob
│   └── Makefile
├── spec/                        ← the PCD specification
│   ├── calc-interest.spec.md
│   └── TRANSLATION_REPORT.md
├── pcd-artifacts/               ← the prompts and template used
│   ├── reverse-prompt.md
│   ├── cli-tool.template.md
│   └── prompt.md
└── code/                        ← the generated implementations
    ├── rs/                      ← Rust
    └── java/                    ← Java
```

### What the example demonstrates

**Reverse engineering.** The COBOL program in `cobol/calc-interest.cob`
is 21 lines: it reads a principal, rate, and number of periods, computes
simple interest, and prints the result. The PCD specification in
`spec/calc-interest.spec.md` was produced from that COBOL via the
[`reverse-prompt.md`](../../prompts/reverse-prompt.md) workflow. The
spec captures the numeric precision and ranges from the COBOL
`PICTURE` clauses as refinement predicates on the TYPES — a `PIC 9(7)V99`
becomes `decimal where value > 0 and value <= 9999999.99`.

**Language portability.** The same specification produced both a Rust
implementation under `code/rs/` and a Java implementation under
`code/java/`. The target language is not declared in the specification;
it is chosen at translation time. Both implementations carry their own
RPM and Debian packaging, man page, and `TRANSLATION_REPORT.md`.

**Reproducibility.** The `pcd-artifacts/` directory contains the exact
template and prompts used for this translation run. Anyone with access
to a capable LLM can reproduce the run, or compare a fresh translation
against the committed output. This is what auditability looks like in
practice: not "we trust the AI", but "here is the input, here is the
output, here is the model that produced it, here is the verification
record."

### Reading order

Start with `cobol/calc-interest.cob` — twenty lines of COBOL is enough
to understand what the program does. Then read
`spec/calc-interest.spec.md` to see what PCD captures from that. Then
pick either `code/rs/` or `code/java/` to see what the translator
produced, and compare the two `TRANSLATION_REPORT.md` files to see how
the same specification was resolved differently into each language.

---

## Choosing what to read

| If you want to see... | Read |
|---|---|
| What a regulated-domain spec looks like | `account-transfer/` |
| The full reverse-engineer → translate loop | `calc-interest/` |
| A non-trivial, real-world spec with milestones | `github.com/mge1512/sitar` |

For the format reference and the rationale behind each section, see
[`doc/user-guide.md`](../doc/user-guide.md) and
[`doc/technical-reference.md`](../doc/technical-reference.md).
