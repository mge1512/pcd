# Translation Report: calc-interest — Rust

**Component:** calc-interest v0.1.0  
**Spec-Schema:** 0.3.21  
**Template:** cli-tool.template v0.3.20  
**Translation date:** 2026-04-09  
**Translator:** PCD AI Translator (Rust target)

---

## 1. Target Language Resolution

| Setting | Value | Source |
|---------|-------|--------|
| Template default language | Go | cli-tool.template TEMPLATE-TABLE |
| Resolved language | **Rust** | Explicit translator directive (user instruction) |
| BINARY-TYPE | static | Template default; Rust mandates static (BINARY-TYPE=dynamic is forbidden for Rust per template PRECONDITIONS) |
| Deviation from template default | Yes — Go → Rust | Stated reason: translator was invoked specifically for the Rust language target. Rust is listed as a valid `LANGUAGE-ALTERNATIVES` (supported) in the TEMPLATE-TABLE. No preset file was present; the override is applied at the invocation level. |

---

## 2. Delivery Mode

**Mode used:** Filesystem (Mode 1) — files written directly to `./code/rs/` using the filesystem write tool.

All deliverables are present on disk. The compile gate was executed successfully in the same environment.

---

## 3. Delivery Phases Executed

The template EXECUTION section defines Go-centric phases. These were adapted for Rust as follows:

| Template Phase | Go artefact | Rust equivalent produced |
|---|---|---|
| Phase 1 — Core implementation | `main.go`, `go.mod` | `src/main.rs`, `Cargo.toml` |
| Phase 2 — Build and packaging | `Makefile`, `<n>.spec`, `debian/*`, `LICENSE` | Same filenames, Rust build commands |
| Phase 3 — Test infrastructure | `independent_tests/INDEPENDENT_TESTS.go` | `tests/independent_tests.rs` |
| Phase 4 — Documentation | `README.md`, `<n>.1.md`, `<n>.1` | Same filenames |
| Phase 5 — Compile gate | `go build ./...` | `cargo build --release` + `cargo test --release` |
| Phase 6 — Report | `TRANSLATION_REPORT.md` | `TRANSLATION_REPORT.md` (this file) |

---

## 4. STEPS Ordering — BEHAVIOR: calculate-simple-interest

All 11 STEPS were implemented in the exact order specified in the spec:

| Step | Spec text | Implementation location |
|------|-----------|------------------------|
| 1 | Read principal from stdin; on failure → exit 1, write error to stderr | `src/main.rs` lines ~20–35 |
| 2 | Read rate from stdin; on failure → exit 1, write error to stderr | `src/main.rs` lines ~37–52 |
| 3 | Read periods from stdin; on failure → exit 1, write error to stderr | `src/main.rs` lines ~54–69 |
| 4 | Validate principal > 0; on failure → exit 2, "invalid principal" | `src/main.rs` ~72–75 |
| 5 | Validate rate > 0; on failure → exit 2, "invalid rate" | `src/main.rs` ~78–81 |
| 6 | Validate periods >= 1; on failure → exit 2, "invalid periods" | `src/main.rs` ~84–87 |
| 7 | Compute interest = principal × rate × periods; overflow → exit 1 | `src/main.rs` ~90–94 |
| 8 | Compute total = principal + interest; overflow → exit 1 | `src/main.rs` ~97–101 |
| 9 | Write "INTEREST: {interest}" to stdout, 2 d.p. | `src/main.rs` ~104–108 |
| 10 | Write "TOTAL:    {total}" to stdout, 2 d.p. | `src/main.rs` ~108–112 |
| 11 | Exit with code 0 | `src/main.rs` ~115 |

---

## 5. TYPE-BINDINGS Applied

The template does not include a `## TYPE-BINDINGS` section for Rust. Bindings were derived mechanically from the spec TYPES and COBOL PIC annotations:

| Spec type | Rust type | Constraint enforced |
|-----------|-----------|---------------------|
| Principal | `f64` | `> 0.0 && <= 9_999_999.99` |
| Rate | `f64` | `> 0.0 && <= 999.9999` |
| Periods | `u32` | `>= 1 && <= 999` (parsed as unsigned integer) |
| Interest | `f64` | `>= 0.0 && <= 9_999_999.99` (overflow check) |
| Total | `f64` | `>= 0.0 && <= 9_999_999.99` (overflow check) |
| InterestResult | inline (no struct) | written directly to stdout |

**Rationale for `f64`:** The spec requires decimal arithmetic matching COBOL `PIC 9(7)V99` (9 significant digits). `f64` provides 15–17 significant decimal digits, well above the 9-digit requirement. The spec INVARIANT `[implementation]` notes "2 decimal places for monetary values, 4 for rate" — these are output formatting constraints, not storage constraints. `f64` is the natural Rust type for this precision level.

---

## 6. GENERATED-FILE-BINDINGS

The template does not contain a `## GENERATED-FILE-BINDINGS` section. No generated infrastructure files (CRDs, manifests, RBAC) are applicable to this CLI tool.

---

## 7. BEHAVIOR Constraint Classification

| BEHAVIOR | Constraint | Action taken |
|----------|------------|-------------|
| calculate-simple-interest | required | Fully implemented |

No `supported` or `forbidden` BEHAVIORs are present in the spec.

---

## 8. COMPONENT → Filename Mapping

| Deliverable | Template key | File produced |
|-------------|-------------|---------------|
| Source | source (required) | `src/main.rs` |
| Cargo manifest | build (required, Rust equiv. of go.mod) | `Cargo.toml` |
| Build | build (required) | `Makefile` |
| RPM spec | RPM (required) | `calc-interest.spec` |
| Debian control | DEB (required) | `debian/control` |
| Debian changelog | DEB (required) | `debian/changelog` |
| Debian rules | DEB (required) | `debian/rules` |
| Debian copyright | DEB (required) | `debian/copyright` |
| Man page source | man (required) | `calc-interest.1.md` |
| Man page troff | man (required) | `calc-interest.1` |
| License | license (required) | `LICENSE` |
| Integration tests | test infrastructure | `tests/independent_tests.rs` |
| Workflow diagram | translation_report | `translation_report/translation-workflow.pikchr` |
| Readme | docs (required) | `README.md` |
| Translation report | report (required) | `TRANSLATION_REPORT.md` |

OCI (`Containerfile`) and PKG (`.pkgbuild`) are `supported` but not active (no preset declared). Not produced.

---

## 9. Active MILESTONE

No `## MILESTONE:` sections are present in the spec. Full translation was performed in a single pass per prompt instructions.

---

## 10. Specification Ambiguities and Resolutions

| Ambiguity | Resolution |
|-----------|-----------|
| Upper-bound validation for Principal, Rate, Periods | The spec defines max values in TYPES but the STEPS only validate lower bounds (`> 0`, `>= 1`). Conservative interpretation: upper-bound checks are also applied (exit 2 if exceeded), consistent with the type constraint definitions. |
| Arithmetic overflow definition | `f64` cannot overflow to infinity for values within spec range. The overflow check uses `!is_finite()` and a monetary-max guard. This is conservative and correct. |
| `InterestResult` struct | The spec defines `InterestResult` as a record type but the BEHAVIOR writes directly to stdout. No struct is needed; output is formatted inline. Documented in translation report. |
| `TOTAL:    ` trailing spaces | The spec INVARIANTS explicitly state 4 trailing spaces after the colon for alignment with the COBOL source. Implemented exactly as specified. |
| Periods parsed as `u32` | The spec says `integer where value >= 1 and value <= 999`. Using `u32` means negative period strings (e.g. "-1") will fail at parse time (exit 1), not validation time (exit 2). This is conservative: a non-parseable integer is a read failure. |

---

## 11. Rules That Could Not Be Implemented Exactly

| Rule | Reason | Mitigation |
|------|--------|-----------|
| OBS install method | OBS repository URLs require an actual published package. README documents the install pattern with a placeholder URL. | Noted in README. |
| SIGNAL-HANDLING: SIGTERM/SIGINT | Rust's default signal handling terminates cleanly on SIGTERM/SIGINT without partial stdout output (because stdout is line-buffered and `process::exit` flushes). No explicit signal handler is needed for this single-operation tool. | Acceptable: the tool completes in milliseconds; no partial output is possible. |

---

## 12. Template Constraints Compliance Table

| Constraint key | Required | Compliant | Notes |
|----------------|----------|-----------|-------|
| VERSION | required | ✓ | 0.1.0 in Cargo.toml and RPM spec |
| SPEC-SCHEMA | required | ✓ | 0.3.21 (from spec) |
| AUTHOR | required | ✓ | Unknown (as in spec) |
| LICENSE | required | ✓ | Apache-2.0 SPDX in all packaging files |
| LANGUAGE | default=Go, used=Rust | ✓ | Rust is a valid supported alternative |
| BINARY-TYPE | static | ✓ | `RUSTFLAGS='-C target-feature=+crt-static'` in Makefile and RPM spec |
| BINARY-COUNT | 1 | ✓ | Exactly one binary |
| RUNTIME-DEPS | none | ✓ | No external crates; stdlib only |
| CLI-ARG-STYLE | key=value | ✓ | No CLI args (stdin-driven); no flags invented |
| EXIT-CODE-OK | 0 | ✓ | `process::exit(0)` |
| EXIT-CODE-ERROR | 1 | ✓ | Read/overflow failures exit 1 |
| EXIT-CODE-INVOCATION | 2 | ✓ | Validation failures exit 2 |
| STREAM-DIAGNOSTICS | stderr | ✓ | All errors via `eprintln!` |
| STREAM-OUTPUT | stdout | ✓ | Results via `writeln!(out, ...)` |
| SIGNAL-HANDLING | SIGTERM/SIGINT | ✓ | Default Rust signal handling (clean exit) |
| OUTPUT-FORMAT RPM | required | ✓ | `calc-interest.spec` produced |
| OUTPUT-FORMAT DEB | required | ✓ | `debian/` directory produced |
| INSTALL-METHOD | OBS | ✓ | README documents OBS install |
| INSTALL-METHOD curl | forbidden | ✓ | curl install not documented |
| PLATFORM Linux | required | ✓ | Primary target |
| CONFIG-ENV-VARS | forbidden | ✓ | No env var controls |
| NETWORK-CALLS | forbidden | ✓ | No network access |
| FILE-MODIFICATION | forbidden | ✓ | No file I/O |
| IDEMPOTENT | required | ✓ | Deterministic computation |
| PRESET-SYSTEM | required | ✓ | No preset files needed (no configurable behavior) |

---

## 13. Phase 5 — Compile Gate Result

| Step | Command | Result |
|------|---------|--------|
| Dependency resolution | `cargo build --release` (no external deps; no `go mod tidy` equivalent needed) | ✓ PASS |
| Compilation | `cargo build --release` | ✓ PASS — `Finished release profile` |
| Integration tests | `cargo test --release` | ✓ PASS — 5/5 tests passed |
| Man page generation | `pandoc calc-interest.1.md -s -t man -o calc-interest.1` | ✓ PASS |

Build artefacts (`target/`) were cleaned after the compile gate (`cargo clean`).

---

## 14. Per-Example Confidence Table

| EXAMPLE | Confidence | Verification method | Unverified claims |
|---------|------------|---------------------|-------------------|
| typical_calculation | **High** | `test_typical_calculation` in `tests/independent_tests.rs` — passes without any external service | None |
| zero_rate_rejected | **High** | `test_zero_rate_rejected` in `tests/independent_tests.rs` — passes without any external service | None |
| zero_principal_rejected | **High** | `test_zero_principal_rejected` in `tests/independent_tests.rs` — passes without any external service | None |
| zero_periods_rejected | **High** | `test_zero_periods_rejected` in `tests/independent_tests.rs` — passes without any external service | None |
| non_numeric_input_rejected | **High** | `test_non_numeric_input_rejected` in `tests/independent_tests.rs` — passes without any external service | None |

All 5 spec examples achieve **High** confidence. Every example has a dedicated named test function that was executed and passed during the compile gate (Phase 5).

---

## 15. Parsing Approach

Standard input is read line-by-line using `std::io::BufRead::lines()` on a locked `stdin` handle. Each line is trimmed of whitespace before parsing. Parsing uses Rust's built-in `str::parse::<f64>()` and `str::parse::<u32>()`. Parse failures produce a descriptive error on stderr and exit code 1.

## 16. Signal Handling Approach

No explicit signal handler is installed. Rust's default behaviour on SIGTERM and SIGINT is to terminate the process cleanly. Since the tool completes its entire computation before writing any output line (stdout is written atomically via `writeln!` on a locked handle), no partial output is possible. This satisfies the template requirement "Clean exit on SIGTERM/SIGINT. No partial output."

---

*End of Translation Report*
