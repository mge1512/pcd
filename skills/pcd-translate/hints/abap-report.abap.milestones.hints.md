# abap-report Milestone Hints

Scaffold pattern and acceptance criteria for `abap-report` milestones.

This file is referenced from the scaffold milestone's `Hints-file:` field.

---

## 1. Scaffold milestone (0.0.0)

The scaffold milestone produces a compileable but functionally empty
package. It is the foundation for all later milestones.

### 1.1 Generated artifacts

| File                                            | Content                                    |
|-------------------------------------------------|--------------------------------------------|
| `.abapgit.xml`                                  | abapGit configuration                      |
| `README.md`                                     | Repo intro with spec hash and version      |
| `LICENSE`                                       | License text from spec META                |
| `src/z<name>/package.devc.xml`                  | Package definition                         |
| `src/z<name>/z_<name>.prog.abap`                | Report with `version` radio button only    |
| `src/z<name>/z_<name>.prog.xml`                 | Program metadata                           |
| `src/z<name>/zcl_<name>.clas.abap`              | Application class — constructor + `version` method |
| `src/z<name>/zcl_<name>.clas.xml`               | Class metadata                             |
| `src/z<name>/zcl_<name>.clas.testclasses.abap`  | Empty test class with one passing smoke test |
| `src/z<name>/zcx_<name>.clas.abap`              | Exception class with `token` attribute     |
| `src/z<name>/zcx_<name>.clas.xml`               | Exception class metadata                   |
| `src/z<name>/zmsg_<name>.msag.xml`              | Message class with one message: `system_error` |
| `src/z<name>/documentation/z_<name>.doc`        | Stub documentation                         |

### 1.2 Scaffold report content

```abap
REPORT z_<name> LINE-SIZE 80.

PARAMETERS p_ver RADIOBUTTON GROUP op DEFAULT 'X'.

START-OF-SELECTION.
  IF p_ver = 'X'.
    DATA(go_app) = NEW zcl_<name>( ).
    WRITE: / go_app->c_spec_version,
             go_app->c_spec_hash+0(12).
  ENDIF.
```

### 1.3 Acceptance criteria (scaffold)

```sh
# 1. Static analysis passes
abaplint src/ --format=summary | grep -q "^Errors: 0$"

# 2. Spec hash is embedded in the application class
grep -q "c_spec_hash TYPE string VALUE" \
  src/z<name>/zcl_<name>.clas.abap

# 3. Spec version constant matches META
grep -q "c_spec_version TYPE string VALUE \`<version>\`" \
  src/z<name>/zcl_<name>.clas.abap

# 4. README references the spec hash
grep -q "<spec-hash-first-12>" README.md

# 5. abapGit metadata is well-formed
xmllint --noout .abapgit.xml
xmllint --noout src/z<name>/*.xml
```

When a live SAP system is available (e.g. SAP Cloud SDK CLI configured
against a BTP ABAP trial tenant):

```sh
# 6. abapGit pull succeeds
abapgit-cli pull-zip --target $TENANT --package Z<NAME> --branch main

# 7. Version operation prints expected output
abap-cli run-report z_<name> --param=p_ver=X | \
  grep -qE "^<version>\s+[0-9a-f]{12}$"
```

---

## 2. Per-BEHAVIOR milestones

Each subsequent milestone groups 3–5 BEHAVIORs. The grouping prefers
behavioral cohesion over arbitrary balance — `validate_*` BEHAVIORs go in
one milestone, `process_*` in the next.

### 2.1 Acceptance pattern: list output

For BEHAVIORs that write a value to the list:

```sh
abap-cli run-report z_<name> \
  --param=op=<behavior> \
  --param=p_country=CZ \
  --param=p_amt=250 \
  | grep -q "^5.0\s*$"
```

### 2.2 Acceptance pattern: error token

For BEHAVIORs that raise a spec ERROR:

```sh
abap-cli run-report z_<name> \
  --param=op=<behavior> \
  --param=p_amt=-50
RC=$?
test "$RC" -ne 0 && \
  abap-cli last-message --format=raw | \
  grep -q "^E\s.*NON_POSITIVE_AMOUNT"
```

### 2.3 Acceptance pattern: ABAP Unit

For invariants and EXAMPLE-driven tests:

```sh
abap-cli run-unit-test \
  --package=Z<NAME> \
  --class=LTC_<BEHAVIOR> \
  --format=junit > /tmp/junit-<behavior>.xml

grep -q 'failures="0"' /tmp/junit-<behavior>.xml
grep -q 'errors="0"' /tmp/junit-<behavior>.xml
```

### 2.4 Acceptance pattern: persistence

For BEHAVIORs that write to a DDIC table:

```sh
abap-cli run-report z_<name> --param=op=<behavior> --param=...

abap-cli sql --query="SELECT COUNT(*) FROM z<name>_<table> WHERE ..." \
  | grep -q "^1$"
```

---

## 3. Static-only acceptance (no live system)

For CI environments without an SAP backend, every milestone's acceptance
reduces to:

```sh
# Static checks
abaplint src/ --format=summary | grep -q "^Errors: 0$"

# XML well-formedness
find src -name '*.xml' -exec xmllint --noout {} \;

# Spec hash drift check (PCD RULE-18)
pcd-lint --rule=RULE-18 <spec-path>

# Offline ABAP Unit (requires embedded runtime — see §4)
abap-cli run-unit-test --offline --package=Z<NAME> --format=junit \
  > /tmp/junit-offline.xml
grep -q 'failures="0"' /tmp/junit-offline.xml
```

This is the **default acceptance bar** for `abap-report`. Live-system
runs are an optional escalation.

---

## 4. Offline ABAP Unit setup

To run ABAP Unit without a hosted SAP system, the CI environment provides
one of:

| Option                                | Notes                                  |
|---------------------------------------|----------------------------------------|
| SAP ABAP Platform Developer Edition (Docker image) | Free, ABAP 7.57 SP00, sufficient for unit tests |
| SAP BTP ABAP trial environment        | Free tier, suitable for clean-core code |
| `abaplint` + JS-side mock execution    | Static analysis + linter-driven unit tests; no real ABAP runtime |

The `abap-cli --offline` flag selects whichever is configured locally.
The spec MUST NOT assume a specific option — the abstraction layer is in
the CI runner configuration.

---

## 5. Test-double posture per milestone

Each milestone declares which INTERFACEs are mocked vs. live:

```
## MILESTONE: 0.2.0
Status: pending
INTERFACES-mocked:
  - rfc_external_system
  - http_pricing_api
INTERFACES-live:
  (none)
```

For the scaffold milestone, all INTERFACEs are mocked.
For the final milestone before release, all INTERFACEs should be live in
the acceptance run — unless the spec declares an INTERFACE as
`Mock-Only: true` for security or licensing reasons.

---

## 6. Milestone version numbering

| Version | Status                                          |
|---------|-------------------------------------------------|
| 0.0.0   | Scaffold — empty, compiles                      |
| 0.x.0   | Subset of BEHAVIORs, mocks for external systems |
| 0.x.y   | Bug fixes within the same BEHAVIOR set          |
| 1.0.0   | All BEHAVIORs implemented, all INTERFACEs live  |
| 1.x.y   | Post-release iterations                         |

The spec's `Version:` field in META must match the highest milestone
reached. The `version` radio button writes this verbatim.

---

## 7. Worked acceptance example

For the reference spec, milestone `0.1.0` (estimate-bees only,
list-countries deferred) acceptance:

```sh
# Static
abaplint src/ --format=summary | grep -q "^Errors: 0$"

# Happy path
abap-cli run-report z_myreportonbees \
  --param=p_est=X --param=p_cntry=CZ --param=p_amt=250 \
  | grep -q "^5.0\s*$"

# Error: non-positive amount
abap-cli run-report z_myreportonbees \
  --param=p_est=X --param=p_cntry=CZ --param=p_amt=-50
test "$?" -ne 0 && \
  abap-cli last-message --format=raw | grep -q "NON_POSITIVE_AMOUNT"

# Error: sanity cap
abap-cli run-report z_myreportonbees \
  --param=p_est=X --param=p_cntry=CZ --param=p_amt=30000
test "$?" -ne 0 && \
  abap-cli last-message --format=raw | grep -q "EXCEEDS_CAP"

# Unit tests
abap-cli run-unit-test --package=ZMYREPORTONBEES \
  --class=LTC_ESTIMATE_BEES --format=junit > /tmp/junit.xml
grep -q 'failures="0"' /tmp/junit.xml
grep -q 'errors="0"' /tmp/junit.xml
```
