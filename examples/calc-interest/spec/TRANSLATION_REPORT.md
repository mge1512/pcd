# Translation Report: calc-interest Reverse Specification

## Summary

This report documents the reverse-engineering of the COBOL program `calc-interest.cob`
into a PCD specification `calc-interest.spec.md` following the `reverse-prompt.md` process.

---

## Source Material Analysed

| File | Description |
|------|-------------|
| `cobol/calc-interest.cob` | GnuCOBOL source, 21 lines |
| `cobol/Makefile` | Build and test targets |
| `pcd-artifacts/reverse-prompt.md` | PCD reverse-engineering prompt |
| `pcd-artifacts/cli-tool.template.md` | PCD cli-tool deployment template |

---

## Phase 1 — Extraction from Source

### What was found

| Item | Extracted value | Source |
|------|----------------|--------|
| Component name | `calc-interest` | `PROGRAM-ID. CALC-INTEREST` |
| License | Not present in source → Apache-2.0 applied per instruction | — |
| Author | Not present in source → `Unknown` | — |
| Language | COBOL (GnuCOBOL) | file extension, Makefile `cobc` invocation |
| Deployment type | `cli-tool` | main program, reads stdin, writes stdout, no server |
| Operations | 1 (calculate-simple-interest) | `PROCEDURE DIVISION` |
| External interfaces | None | no file I/O, no network, no subprocess |
| Dependencies | None | Makefile: only `cobc` (compiler, not runtime dep) |

### COBOL data types mapped to spec TYPES

| COBOL PIC | Meaning | Spec type |
|-----------|---------|-----------|
| `9(7)V99` | 7 integer + 2 decimal digits | Principal, Interest, Total |
| `9(3)V9999` | 3 integer + 4 decimal digits | Rate |
| `9(3)` | 3 integer digits | Periods |

### Formula extracted

From line 16: `WS-INTEREST = WS-PRINCIPAL * WS-RATE * WS-PERIODS`
→ Simple (flat) interest, not compound interest.

From line 17: `WS-TOTAL = WS-PRINCIPAL + WS-INTEREST`

---

## Phase 2 — Identity and Intent Decisions

### Deployment type

Detected: `cli-tool`
Evidence: `PROCEDURE DIVISION` with `ACCEPT`/`DISPLAY`/`STOP RUN` — no server,
no daemon, no long-running loop. Reads three values, writes two, exits.

### Language decision

Original language: COBOL (GnuCOBOL).
Target language for spec: not fixed — the spec is language-agnostic per PCD convention.
The COBOL source becomes the behavioral reference; a translator will regenerate in the
target language chosen at translation time (default: Go per cli-tool template).

### Requested delta

None — pure reverse-engineering output. No DELTA section added.

---

## Phase 3 — Gap-fill

### Items resolved

| Gap | Resolution |
|-----|-----------|
| License | Not found in source; Apache-2.0 applied per task instruction |
| Author | Not found in source; recorded as `Unknown` |
| Error handling | COBOL `ACCEPT` has no explicit error handling; spec adds input validation and error exits consistent with cli-tool template requirements |
| Rate semantics | Makefile test uses `0.0350` for 3.5% — confirmed as decimal fraction, not percentage integer |
| Output format | COBOL `DISPLAY` produces labels `"INTEREST: "` and `"TOTAL:    "` (4 trailing spaces on TOTAL for alignment); preserved in spec INVARIANTS |
| Periods meaning | Inferred as time periods (e.g. months) from context; no compounding |

---

## Phase 4 — Specification Written

Output file: `./spec/calc-interest.spec.md`

Sections produced:
- META (all 7 required fields)
- TYPES (5 types extracted from COBOL PIC clauses)
- BEHAVIOR: calculate-simple-interest (1 operation, all 5 sub-sections)
- PRECONDITIONS (global)
- POSTCONDITIONS (global)
- INVARIANTS (5, all tagged [observable] or [implementation])
- EXAMPLES (5: 1 success + 4 negative-path)
- DEPENDENCIES (none)
- DEPLOYMENT

---

## Phase 5 — Milestone Design

The specification has 1 BEHAVIOR. This is well below the 10-BEHAVIOR threshold.
No milestones were added; the spec is translatable in a single pass.

---

## Phase 6 — Self-Check Results

| Check | Result |
|-------|--------|
| META contains all 7 required fields | ✓ |
| Spec-Schema is 0.3.21 | ✓ |
| License present (Apache-2.0) | ✓ |
| Author present | ✓ (Unknown — not found in source) |
| Deployment type is valid PCD type | ✓ (cli-tool) |
| Every BEHAVIOR has INPUTS, PRECONDITIONS, STEPS, POSTCONDITIONS, ERRORS | ✓ |
| Every STEP has explicit "on failure" exit | ✓ |
| Every INVARIANT tagged [observable] or [implementation] | ✓ |
| Every EXAMPLE has GIVEN, WHEN, THEN | ✓ |
| Every BEHAVIOR with error exits has negative-path EXAMPLE | ✓ (4 negative examples) |
| INTERFACES section present if external systems identified | N/A — no external systems |
| DEPENDENCIES section present | ✓ (none declared) |
| No language-specific constructs in TYPES/BEHAVIOR/INVARIANTS | ✓ |
| No invented type names or versions | ✓ |
| DELTA section absent (no changes requested) | ✓ |
| No contradictions | ✓ |

---

## pcd-lint Result

```
✓ ./spec/calc-interest.spec.md: valid
```

---

## Notes

1. **License**: The COBOL source contains no SPDX header, copyright notice, or LICENSE file.
   Apache-2.0 was applied as instructed. If this program has a different actual license,
   the META must be updated before use.

2. **Author**: No author information was found in the source, Makefile, or any README.
   Recorded as `Unknown`. Must be updated before publication.

3. **Error handling**: The original COBOL program has no error handling — invalid input
   causes undefined COBOL runtime behavior. The spec adds explicit validation steps
   consistent with the cli-tool template's exit code requirements (exit 0/1/2).

4. **Formula verification**: The Makefile test `echo -e "10000.00\n0.0350\n12" | ./calc-interest`
   with principal=10000, rate=0.035, periods=12 gives:
   - interest = 10000 × 0.035 × 12 = 4200.00 ✓
   - total = 10000 + 4200 = 14200.00 ✓
   This is confirmed as simple (flat) interest, not compound interest.
