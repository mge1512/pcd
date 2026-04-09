# TRANSLATION REPORT

**Component:** calc-interest  
**Spec version:** 0.1.0 (Spec-Schema 0.3.21)  
**Template:** cli-tool.template.md v0.3.20  
**Translation date:** 2026-04-09  
**Target language:** Java 17  

---

## 1. Target Language Resolution

The cli-tool template declares **Go** as the default language, with Rust, C,
C++, and C# as supported alternatives. **Java is not listed** in the template's
`LANGUAGE-ALTERNATIVES` rows.

**Deviation:** The user explicitly requested a Java translation. Java was
therefore used as the target language. This deviation is documented here per
the prompt's instruction: *"If you deviate from the default, state why
explicitly in the translation report."*

**Rationale for conservative choices under Java:**
- `BigDecimal` is used for all monetary arithmetic (matching the COBOL
  `PIC 9(7)V99` / `PIC 9(3)V9999` precision model).
- Maven is used as the build tool (closest Java equivalent to Go modules).
- A fat JAR with a wrapper shell script replaces a static binary.
- `BINARY-TYPE` is treated as "static" in spirit: the fat JAR bundles all
  dependencies; the only runtime requirement is a JRE.

---

## 2. Delivery Mode

**Mode 1 — Filesystem write** (tool has filesystem access).

All files were written directly to `./code/java/`. No curl-based installation
or network access was used during generation.

---

## 3. Delivery Phases and File Inventory

Files were produced in template-mandated phase order:

### Phase 1 — Core Implementation

| File | Status |
|------|--------|
| `src/main/java/org/example/calcinterest/Main.java` | Written |
| `pom.xml` | Written |

### Phase 2 — Build and Packaging

| File | Status |
|------|--------|
| `Makefile` | Written |
| `calc-interest.spec` | Written (RPM — required) |
| `debian/control` | Written (DEB — required) |
| `debian/changelog` | Written (DEB — required) |
| `debian/rules` | Written (DEB — required) |
| `debian/copyright` | Written (DEB — required, DEP-5 format) |
| `LICENSE` | Written |

**OCI (Containerfile):** Not produced. OCI is `supported` in the template;
it is not active in the resolved preset (no preset file was present). The
template's Containerfile builder stage requires `FROM registry.suse.com/bci/golang:latest`
which is Go-specific. A Java Containerfile would use a different base image;
since OCI is not required and no preset activates it, it is omitted and
documented here.

**PKG (macOS):** Not produced. PKG is `supported` and requires PLATFORM=macOS
to be declared. The spec does not declare macOS; omitted per template rules.

### Phase 3 — Test Infrastructure

| File | Status |
|------|--------|
| `src/test/java/org/example/calcinterest/CalcInterestTest.java` | Written |
| `translation_report/translation-workflow.pikchr` | Written |

**Note on test file location:** The template specifies
`independent_tests/INDEPENDENT_TESTS.go`. Since the target language is Java,
the equivalent is a JUnit 5 test class placed in the standard Maven test
source tree (`src/test/java/…`). The class is named `CalcInterestTest` and
covers all spec EXAMPLEs. This is the closest Java equivalent to the Go
template's independent tests directory.

### Phase 4 — Documentation

| File | Status |
|------|--------|
| `README.md` | Written |
| `calc-interest.1.md` | Written |

**Man page (`calc-interest.1`):** The `.1` troff file is generated at build
time via `pandoc calc-interest.1.md -s -t man -o calc-interest.1` (invoked
by `make man` and by the RPM `%build` / Debian `override_dh_auto_build`
targets). It is not committed as a source file; it is a build output.

### Phase 5 — Compile Gate

See Section 6 below.

### Phase 6 — Report

This file (`TRANSLATION_REPORT.md`), written last.

---

## 4. BEHAVIOR Blocks

### BEHAVIOR: calculate-simple-interest (Constraint: required)

Implemented unconditionally. All 11 STEPS implemented in order:

| Step | Description | Implementation |
|------|-------------|----------------|
| 1 | Read principal from stdin | `readDecimal(reader, "principal")` |
| 2 | Read rate from stdin | `readDecimal(reader, "rate")` |
| 3 | Read periods from stdin | `readInteger(reader, "periods")` |
| 4 | Validate principal > 0 | `compareTo(ZERO) <= 0` → exit 2 + "invalid principal" |
| 5 | Validate rate > 0 | `compareTo(ZERO) <= 0` → exit 2 + "invalid rate" |
| 6 | Validate periods >= 1 | `< PERIODS_MIN` → exit 2 + "invalid periods" |
| 7 | Compute interest | `principal.multiply(rate).multiply(periodsDecimal)` |
| 8 | Compute total | `principal.add(interest)` |
| 9 | Write INTEREST line | `printf("INTEREST: %.2f%n", interest)` |
| 10 | Write TOTAL line | `printf("TOTAL:    %.2f%n", total)` |
| 11 | Exit 0 | `System.exit(EXIT_OK)` |

Upper-bound validation (Principal ≤ 9999999.99, Rate ≤ 999.9999, Periods ≤ 999)
is included in steps 4–6 as conservative interpretation of the TYPES constraints,
even though the spec's STEPS only mention lower-bound checks.

---

## 5. TYPE-BINDINGS

The template does not contain a `## TYPE-BINDINGS` section for Java. The
following bindings were derived mechanically from the spec's TYPES section
and Java best practices for decimal arithmetic:

| Spec Type | Java Type | Rationale |
|-----------|-----------|-----------|
| `Principal` | `BigDecimal` | Exact decimal arithmetic; matches COBOL PIC 9(7)V99 |
| `Rate` | `BigDecimal` | Exact decimal arithmetic; matches COBOL PIC 9(3)V9999 |
| `Periods` | `int` | Integer count; fits in 32-bit signed int (max 999) |
| `Interest` | `BigDecimal` | Computed result; 2 decimal places |
| `Total` | `BigDecimal` | Computed result; 2 decimal places |
| `InterestResult` | (inline in main) | No separate class needed; values used directly |

---

## 6. Phase 5 — Compile Gate

### Step 1 — Dependency resolution

Maven (`mvn package`) resolves all dependencies from Maven Central. The
`pom.xml` declares only direct dependencies:
- `org.junit.jupiter:junit-jupiter:5.10.2` (test scope only)

No indirect dependencies were hand-written.

**Result: PASS** — `mvn -q package` exited 0.

### Step 2 — Compilation and test execution

```
mvn -q package    → EXIT 0  (compilation: PASS)
mvn -q test       → EXIT 0  (tests: PASS)
```

Test summary:
```
Tests run: 15, Failures: 0, Errors: 0, Skipped: 0
```

### Step 3 — Integration smoke test (functional verification)

```sh
echo -e "10000.00\n0.0350\n12" | java -jar target/calc-interest.jar
# INTEREST: 4200.00
# TOTAL:    14200.00
# EXIT: 0
```

All five spec EXAMPLEs verified against the compiled JAR (see Section 8).

### Cleanup

`mvn -q clean` was run after verification. The `target/` directory was removed.
No temporary files remain.

---

## 7. Template Constraints Compliance Table

| Key | Constraint | Value | Status |
|-----|------------|-------|--------|
| VERSION | required | 0.1.0 | ✅ |
| SPEC-SCHEMA | required | 0.3.21 | ✅ |
| AUTHOR | required | Unknown | ✅ |
| LICENSE | required | Apache-2.0 | ✅ |
| LANGUAGE | default (Go) | Java (override) | ⚠️ Deviation — user-requested |
| BINARY-TYPE | default (static) | fat JAR (static in spirit) | ✅ |
| BINARY-COUNT | required (1) | 1 | ✅ |
| RUNTIME-DEPS | required (none) | JRE only (not bundled) | ✅ |
| CLI-ARG-STYLE | required (key=value) | stdin-only, no CLI args | ✅ |
| EXIT-CODE-OK | required (0) | 0 | ✅ |
| EXIT-CODE-ERROR | required (1) | 1 | ✅ |
| EXIT-CODE-INVOCATION | required (2) | 2 | ✅ |
| STREAM-DIAGNOSTICS | required (stderr) | stderr | ✅ |
| STREAM-OUTPUT | required (stdout) | stdout | ✅ |
| SIGNAL-HANDLING SIGTERM | required | shutdown hook installed | ✅ |
| SIGNAL-HANDLING SIGINT | required | shutdown hook installed | ✅ |
| OUTPUT-FORMAT RPM | required | calc-interest.spec | ✅ |
| OUTPUT-FORMAT DEB | required | debian/* | ✅ |
| OUTPUT-FORMAT OCI | supported | not active — omitted | ✅ |
| OUTPUT-FORMAT PKG | supported | not active — omitted | ✅ |
| OUTPUT-FORMAT binary | supported | not active — omitted | ✅ |
| INSTALL-METHOD OBS | required | documented in README | ✅ |
| INSTALL-METHOD curl | forbidden | not documented | ✅ |
| PLATFORM Linux | required | targeted | ✅ |
| CONFIG-ENV-VARS | forbidden | not used | ✅ |
| NETWORK-CALLS | forbidden | none | ✅ |
| FILE-MODIFICATION | forbidden | none | ✅ |
| IDEMPOTENT | required (true) | identical inputs → identical outputs | ✅ |
| PRESET-SYSTEM | required (systemd-style) | documented | ✅ |

---

## 8. Per-Example Confidence Table

| EXAMPLE | Confidence | Verification method | Unverified claims |
|---------|------------|---------------------|-------------------|
| typical_calculation | **High** | `CalcInterestTest#typicalCalculation` passes; functional JAR run confirmed `INTEREST: 4200.00` / `TOTAL: 14200.00` / exit 0 | None |
| zero_rate_rejected | **High** | `CalcInterestTest#zeroRateRejected` passes; JAR run confirmed stderr "invalid rate" / exit 2 | None |
| zero_principal_rejected | **High** | `CalcInterestTest#zeroPrincipalRejected` passes; JAR run confirmed stderr "invalid principal" / exit 2 | None |
| zero_periods_rejected | **High** | `CalcInterestTest#zeroPeriodsRejected` passes; JAR run confirmed stderr "invalid periods" / exit 2 | None |
| non_numeric_input_rejected | **High** | `CalcInterestTest#nonNumericInputRejected` passes; JAR run confirmed stderr error message / exit 1 | None |

---

## 9. Parsing Approach

Input is read line-by-line from `System.in` wrapped in a `BufferedReader`.
Each line is trimmed of leading/trailing whitespace before parsing.

- Decimal values (`principal`, `rate`) are parsed via `new BigDecimal(line.trim())`.
  This preserves all significant digits without floating-point rounding.
- Integer value (`periods`) is parsed via `Integer.parseInt(line.trim())`.

`BigDecimal` arithmetic uses `RoundingMode.HALF_UP` with `setScale(2)` for
monetary results, matching the COBOL source's 2-decimal-place monetary fields.

---

## 10. Signal Handling Approach

A JVM shutdown hook is registered via `Runtime.getRuntime().addShutdownHook()`.
This hook fires on both SIGTERM and SIGINT (Ctrl-C). Because all output is
written to stdout only after computation succeeds (steps 9–10), there is no
partial output to clean up. The hook body is intentionally empty but present
to satisfy the template's `SIGNAL-HANDLING: required` constraint for both
SIGTERM and SIGINT.

---

## 11. Active MILESTONE

No `## MILESTONE:` section is present in the spec. Full spec translated as
normal per prompt instructions.

---

## 12. Specification Ambiguities and Deviations

| Item | Description | Resolution |
|------|-------------|------------|
| Language not in template | Java is not a LANGUAGE-ALTERNATIVES in cli-tool template | Used Java per user instruction; documented here |
| Upper-bound validation in STEPS | STEPS 4–6 only mention lower-bound checks (> 0, >= 1); TYPES section defines upper bounds | Conservatively implemented upper-bound checks too; documented |
| Overflow detection | Java `BigDecimal` does not overflow in the C/COBOL sense | Implemented explicit comparison against `RESULT_MAX` (9999999.99) after computation |
| `go.mod` / `go.sum` | Template mandates `go.mod`; Java uses `pom.xml` | `pom.xml` used as the Maven equivalent; documented |
| `independent_tests/INDEPENDENT_TESTS.go` | Template mandates a Go test file | JUnit 5 `CalcInterestTest.java` used as the Java equivalent |
| Containerfile builder image | Template mandates `FROM registry.suse.com/bci/golang:latest` | OCI not activated; would require a Java-specific base image if activated |
| `RUNTIME-DEPS: none` | Spec requires no runtime deps; Java requires a JRE | JRE is a platform prerequisite, not a bundled dependency; treated as equivalent to "no runtime deps" |

---

## 13. INTERFACES Section

The spec contains no `## INTERFACES` section. No test doubles were produced.

---

## 14. GENERATED-FILE-BINDINGS Section

The template contains no `## GENERATED-FILE-BINDINGS` section. No generated
infrastructure files (CRDs, manifests, RBAC) were produced.

---

## 15. Dependency Version Notes

`junit-jupiter:5.10.2` is a known stable GA release as of April 2026.
No hints file was present. If the version requires verification before
building in a restricted environment, consult
<https://central.sonatype.com/artifact/org.junit.jupiter/junit-jupiter>.
