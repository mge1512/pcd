# abap-report — Canonical Translation Rules

This document is the authoritative reference for translating a PCD spec into
an `abap-report` package. The AI translator follows these rules; deviations
are bugs in the translator.

A PCD deployment template that translates a spec into an abapGit-compatible
ABAP package suitable for execution on a modern SAP application server.

Targets

- SAP S/4HANA (on-premise, private cloud, public cloud)
- SAP BTP ABAP Environment ("Steampunk")
- Any NetWeaver 7.57+ system reachable via the ABAP Cloud development model

When to use this template

- The behaviour belongs on the SAP application server itself: it operates on
  SAP master data, business objects, or other ABAP-resident state.
- The customer wants delivery via `abapGit pull` from a Git repository.
- The component runs in customer namespace (`Z*`, `Y*`, or a registered SAP
  namespace `/COMPANY/`).
- The component is invocable as an executable program (transaction code
  optional), a function module, or a class method exposed via RFC/OData.

When NOT to use this template

| Situation                                                | Use instead          |
|----------------------------------------------------------|----------------------|
| Frontend / SAPUI5 / Fiori Elements                       | (future) `fiori-app` |
| Linux-side CLI tool talking to SAP via RFC or OData      | `cli-tool` with SAP as INTERFACE |
| Pure HANA-side computation (CDS, AMDP) with no ABAP entry | (future) `cds-view`  |
| Modification of SAP standard objects                     | not in scope — violates clean core |

## META

Deployment:  template
Version:     0.3.29
Spec-Schema: 0.4.0
Author:      Matthias G. Eckermann <pcd@mailbox.org>
License:     CC-BY-4.0
Verification: none
Safety-Level: QM
Template-For: abap-report

---

## 1. Naming

### 1.1 Spec name → ABAP object names

Given a spec named `<Name>` (CamelCase), the translator derives:

| Object              | ABAP name              | Limit  |
|---------------------|------------------------|--------|
| abapGit package     | `Z<NAME>`              | ≤30    |
| Report              | `Z_<NAME>`             | ≤30    |
| Application class   | `ZCL_<NAME>`           | ≤30    |
| Exception class     | `ZCX_<NAME>`           | ≤30    |
| Message class       | `ZMSG_<NAME>`          | ≤20    |
| Interface (per ext) | `ZIF_<NAME>_<SYS>`     | ≤30    |

Conversion: `<Name>` → upper-snake-case (`MyReportOnBees` → `MYREPORTONBEES`).
If a name would exceed its limit, the translator abbreviates per the rules in
`abap-report.abap.style.hints.md` §1.

### 1.2 Namespace

Default prefix is `Z`. The spec may override via:

```
## META
...
ABAP-Namespace: /SUSE/
```

When a namespace is declared, `Z`/`ZCL`/`ZCX`/`ZMSG`/`ZIF` are replaced with
`/SUSE/`, `/SUSE/CL_`, `/SUSE/CX_`, `/SUSE/MSG_`, `/SUSE/IF_` respectively.
Length limits apply to the full name including the namespace.

---

## 2. Object mapping

| Spec element              | ABAP object                                          |
|---------------------------|------------------------------------------------------|
| The spec itself           | One abapGit package                                  |
| `BEHAVIOR: <op>`          | One method on the application class                  |
| `TYPES.<X>`               | One `TYPES` declaration in the application class     |
| `INTERFACES.<sys>`        | One interface class + one default implementation     |
| `INVARIANTS` (observable) | One test method per invariant in the test class      |
| `INVARIANTS` (implementation) | Documented only; no generated code               |
| `EXAMPLES`                | One test method per EXAMPLE in the test class        |
| `ERRORS`                  | One message per ERROR + one constant on the exception class |
| `DEPENDENCIES`            | Listed in package interfaces (released APIs only)    |
| `PRECONDITIONS` (global)  | Asserted in the application class `constructor`      |
| `POSTCONDITIONS` (global) | Asserted at end of each BEHAVIOR method              |
| `DEPLOYMENT`              | Rendered into `documentation/z_<name>.doc`           |

---

## 3. I/O dialect mapping

ABAP has no Unix-style stdio. The spec's I/O contract is mapped as follows:

| Spec concept                              | ABAP equivalent                            |
|-------------------------------------------|--------------------------------------------|
| Read from `argv` / positional arg         | `PARAMETERS` on the selection screen       |
| Subcommand / operation selection          | Radio button group `op`                    |
| Write to stdout                           | `WRITE: /` (list output)                   |
| Diagnostic + non-zero exit                | `MESSAGE <token> TYPE 'E'`                 |
| Exit code 0                               | Implicit on end of `START-OF-SELECTION`    |
| Exit code non-zero                        | `MESSAGE TYPE 'E'` or `LEAVE PROGRAM`      |
| Environment variable                      | Not available — forbidden in spec for this template |
| File path argument                        | Permitted only when spec declares `OPEN DATASET` INTERFACE |

### 3.1 Error tokens

Every entry in a BEHAVIOR's `ERRORS` list MUST appear as a message in the
message class. The message ID is the literal token from the spec, transformed
to ABAP message-class conventions:

- Hyphens → underscores: `no-rate` → `NO_RATE`
- Upper-cased
- Truncated to 30 characters if needed

The exception class `ZCX_<NAME>` carries the original spec token verbatim as
a `token` attribute, so external callers (RFC, OData) see the unmangled
spec-level identifier.

### 3.2 Output rendering

- Numeric outputs: `WRITE: ... DECIMALS n` where `n` matches the spec-declared
  precision (e.g. one decimal for `BeeCount`).
- Multi-line outputs: one `WRITE: /` per line.
- No color codes, terminal escapes, or progress indicators. List output only.
- If the spec requires structured output (JSON, XML), the translator emits a
  call to `/ui2/cl_json=>serialize` or `cl_sxml_string_writer` and writes the
  result with `WRITE: /`.

---

## 4. Selection screen synthesis

### 4.1 One BEHAVIOR

All INPUTS become `PARAMETERS` in spec declaration order. No radio group.

### 4.2 Multiple BEHAVIORs

The translator synthesizes a radio button group `op` with one button per
BEHAVIOR. Inputs for each BEHAVIOR are placed in conditionally-visible
blocks via `AT SELECTION-SCREEN OUTPUT`:

```abap
SELECTION-SCREEN BEGIN OF BLOCK b_op WITH FRAME TITLE TEXT-t01.
  PARAMETERS:
    p_<op1> RADIOBUTTON GROUP op DEFAULT 'X' USER-COMMAND op_changed,
    p_<op2> RADIOBUTTON GROUP op,
    ...
SELECTION-SCREEN END OF BLOCK b_op.

SELECTION-SCREEN BEGIN OF BLOCK b_in_<op1> WITH FRAME TITLE TEXT-t02.
  PARAMETERS: <op1 inputs> MODIF ID i1.
SELECTION-SCREEN END OF BLOCK b_in_<op1>.

AT SELECTION-SCREEN OUTPUT.
  LOOP AT SCREEN.
    IF screen-group1 = 'I1' AND p_<op1> <> 'X'.
      screen-active = '0'.
      MODIFY SCREEN.
    ENDIF.
    "... repeat per op
  ENDLOOP.
```

### 4.3 Type mapping for PARAMETERS

| Spec type signature                       | ABAP type                                  |
|-------------------------------------------|--------------------------------------------|
| `text(maxlen=N)`                          | `TYPE c LENGTH N`                          |
| `text` (unbounded)                        | `TYPE string`                              |
| `identifier(len=2)` (e.g. country code)   | `TYPE c LENGTH 2`                          |
| `integer`                                 | `TYPE i`                                   |
| `positive integer`                        | `TYPE i` + `AT SELECTION-SCREEN` check     |
| `decimal(precision=P, scale=S)`           | `TYPE p LENGTH P DECIMALS S`               |
| `decimal` (unspecified)                   | `TYPE p LENGTH 16 DECIMALS 2`              |
| `date`                                    | `TYPE d`                                   |
| `time`                                    | `TYPE t`                                   |
| `boolean`                                 | `TYPE abap_bool`                           |
| `enum(<v1>, <v2>, ...)`                   | `TYPE c LENGTH <max>` + `VALUE-REQUEST`    |

### 4.4 Input validation

- **Shape validation** (range, regex, length, type): emitted in
  `AT SELECTION-SCREEN`, raises `MESSAGE TYPE 'E'` with token
  `malformed_<param>`.
- **Semantic validation** (e.g. "country code must exist in price table"):
  emitted in the application class method, raises `lcx_<name>` with the
  spec-declared token.

---

## 5. Application class structure

### 5.1 Skeleton

```abap
CLASS zcl_<name> DEFINITION FINAL CREATE PUBLIC.
  PUBLIC SECTION.
    CONSTANTS:
      c_spec_hash    TYPE string VALUE `<64-hex>`,    " A.20 spec hash
      c_spec_version TYPE string VALUE `<version>`,
      c_rate_date    TYPE d      VALUE '<yyyymmdd>'.  " if applicable

    METHODS:
      constructor RAISING zcx_<name>,
      <one method per BEHAVIOR — see §5.3>.

  PRIVATE SECTION.
    DATA:
      <reference data tables — see §5.2>,
      <external system handles — see §6>.

    METHODS:
      load_data,
      <helper methods, one per private operation>.
ENDCLASS.
```

### 5.2 Compile-time reference data

If the spec declares static tables (price tables, exchange rates, lookup
sets), they are initialized in `load_data` using `VALUE #( ... )`:

```abap
mt_prices = VALUE #(
  ( country = 'AT' currency = 'EUR' price = '9.00'   )
  ( country = 'CZ' currency = 'CZK' price = '100.00' )
  ...
).
```

Threshold rule: if a table has more than 100 rows, the translator emits a
DDIC table instead and loads it via `SELECT * FROM z<name>_<table>`. The
spec MUST declare this with `Storage: ddic` on the TYPE.

### 5.3 BEHAVIOR method signature

```abap
METHODS <behavior>
  IMPORTING <inputs in spec order, with iv_/it_/is_ prefix>
  RETURNING VALUE(<rv|rt|rs>_result) TYPE <type>
  RAISING   zcx_<name>.
```

For BEHAVIORs whose result is "write to stdout":
- The method returns the output structure/table.
- The `START-OF-SELECTION` block calls the method and renders the result
  via `WRITE: /`.

This separation enables ABAP Unit to assert on the return value without
having to parse list output.

---

## 6. INTERFACES — external systems

### 6.1 Allowed integration patterns

| Spec INTERFACE kind         | ABAP binding                                       |
|------------------------------|----------------------------------------------------|
| `database`                   | Open SQL against customer-namespace DDIC table     |
| `rfc(<system>)`              | Synchronous RFC via referenced destination         |
| `http(<service>)`             | HTTP destination + `cl_web_http_client_manager`    |
| `file(<location>)`           | `OPEN DATASET` — classic ABAP only, forbidden in ABAP Cloud |
| `email`                      | `cl_bcs_message` (BCS framework)                   |
| `workflow(<event>)`          | Released BO event raise                            |
| `released-api(<api>)`        | Direct call to released API (verified §6.3)        |

### 6.2 Interface class generation

For each INTERFACE in the spec, the translator generates:

```abap
INTERFACE zif_<name>_<sys> PUBLIC.
  METHODS:
    <one method per operation in the INTERFACE>.
ENDINTERFACE.
```

Plus two implementations:

- `zcl_<name>_<sys>` — production implementation (real RFC / HTTP / SQL)
- `ltd_<sys>` — local test double (in the test class file)

The application class constructor takes `io_<sys> TYPE REF TO zif_<name>_<sys>`
with the production class as default. Tests inject the double.

### 6.3 Clean core enforcement

The translator MUST only emit calls to APIs marked "released" in the SAP API
Hub catalog. If the spec describes a behaviour that would require a
non-released API, the translator MUST:

1. Stop emitting code.
2. Report the required API by name (function module, class, table).
3. Suggest one of:
   - Find a released equivalent
   - Add a custom released wrapper as an INTERFACE
   - Mark the spec as `Clean-Core: false` (downgrades target to classic ABAP, blocks ABAP Cloud deployment)

The translator MUST NOT silently emit a non-released call.

---

## 7. Persistence

Persistent state requires a customer-namespace DDIC table. The spec declares
this with a `Persisted: true` annotation on the TYPE:

```
## TYPES

### MeasurementRecord
Persisted: true
Storage: ddic
Fields:
  - id: identifier(len=20), primary
  - timestamp: timestamp
  - value: decimal(precision=15, scale=3)
```

The translator generates:

- `z<name>_<type>.tabl.xml` — table definition
- `z<name>_<type>e_<field>.dtel.xml` — data element per field
- A typed table type `tt_<type>` in the application class

Activation happens via abapGit pull; no transport handling.

### 7.1 CDS views

If the type is annotated `View: <name>`, the translator additionally emits:

- `z<name>_<view>.ddls.asddls` — CDS view DDL source
- The view is annotated with the spec hash for drift detection

---

## 8. Test scaffolding (ABAP Unit)

### 8.1 Test class layout

```abap
CLASS ltc_<behavior> DEFINITION FINAL FOR TESTING
  DURATION SHORT
  RISK LEVEL HARMLESS.

  PRIVATE SECTION.
    DATA mo_cut TYPE REF TO zcl_<name>.

    METHODS:
      setup,
      <one method per EXAMPLE> FOR TESTING.
ENDCLASS.
```

### 8.2 EXAMPLE → test method

| EXAMPLE kind   | Test pattern                                                      |
|----------------|-------------------------------------------------------------------|
| Happy path     | Direct method call → `cl_abap_unit_assert=>assert_equals`         |
| Error path     | `TRY ... CATCH zcx_<name> INTO DATA(lx)` → assert `lx->token`      |
| Multi-step     | Sequential calls, asserting intermediate state between them       |

### 8.3 INVARIANTS → tests

For each `[observable]` invariant, the translator emits one test method that:
1. Runs every EXAMPLE input through the relevant BEHAVIOR.
2. Asserts the invariant holds against each result.

For `[implementation]` invariants, no test is emitted; the invariant text
is reproduced in `documentation/z_<name>.doc` under a "DESIGN INVARIANTS"
heading.

### 8.4 Test data

Test fixtures are declared in the spec's EXAMPLES section. The translator
embeds them as constants in the test class — no external CSV / JSON files.
If a test needs more than 20 rows of fixture data, the translator emits a
DDIC table with prefix `Z<NAME>_TEST_<n>` and a `setup` that loads it.

---

## 9. Documentation

Every `abap-report` package MUST include `documentation/z_<name>.doc`.
This is the man-page equivalent. It is generated from the spec.

### 9.1 Required sections

```
TITLE
  <spec name> — <one-line purpose>

VERSION
  <version> (spec hash <first 12 chars>)

PURPOSE
  <Multi-paragraph derived from spec META and DEPLOYMENT sections>

OPERATIONS
  <For each BEHAVIOR:>
    <name> — <one-line summary>
      Inputs:   <list>
      Output:   <description>
      Errors:   <list of error tokens>

REFERENCE DATA
  <For each compile-time constant table: name, source, effective date>

ERROR TOKENS
  <Alphabetical list of all error tokens with explanation>

DESIGN INVARIANTS
  <All [implementation] invariants verbatim>

AUTHORIZATION
  <RequiresAuth declarations, or note that authorization is transaction-level>

CHANGES
  <Spec version + hash + date — appended on each translation>
```

### 9.2 Format

Plain text, ≤72 columns wide. No markup; SE38 documentation renders as
preformatted text.

---

## 10. abapGit metadata

### 10.1 `.abapgit.xml`

```xml
<?xml version="1.0" encoding="utf-8"?>
<asx:abap xmlns:asx="http://www.sap.com/abapxml" version="1.0">
  <asx:values>
    <DATA>
      <MASTER_LANGUAGE>E</MASTER_LANGUAGE>
      <STARTING_FOLDER>/src/</STARTING_FOLDER>
      <FOLDER_LOGIC>FULL</FOLDER_LOGIC>
      <IGNORE>
        <item>/.gitignore</item>
        <item>/LICENSE</item>
        <item>/README.md</item>
        <item>/.github/*</item>
      </IGNORE>
    </DATA>
  </asx:values>
</asx:abap>
```

### 10.2 `<package>.devc.xml`

Generated from spec META. Defaults:

- `CTEXT` (short description): spec one-liner
- `AS4USER`: `<no_user>` (set by abapGit on pull)
- `DEVCLASS`: package name
- `PARENTCL`: `$TMP` (local) or declared parent

The spec may override `PARENTCL` via:

```
ABAP-Parent-Package: ZSUSE_PCD
```

---

## 11. Authorization

If any BEHAVIOR declares `RequiresAuth: <object>(<field>=<value>, ...)`,
the translator emits as the first step of that BEHAVIOR's method:

```abap
AUTHORITY-CHECK OBJECT '<object>'
  ID '<field1>' FIELD '<value1>'
  ID '<field2>' FIELD '<value2>'.
IF sy-subrc <> 0.
  RAISE EXCEPTION TYPE zcx_<name>
    EXPORTING token = `no_authorization`.
ENDIF.
```

`no_authorization` is added to the message class automatically.

If no BEHAVIOR declares `RequiresAuth`, the documentation MUST state:
"No explicit authorization check; relies on transaction-level authorization
(S_TCODE) and the system's role assignment."

---

## 12. Forbidden patterns

The translator MUST NOT emit:

- `MODIFY` against SAP standard tables
- `INSERT REPORT` (dynamic code generation)
- `CALL TRANSACTION` against SAP standard transactions
- `SUBMIT` of SAP standard reports
- `FIELD-SYMBOLS` of `TYPE ANY` combined with dynamic `ASSIGN COMPONENT`
- `PERFORM` (use methods)
- `EXEC SQL` (use Open SQL or AMDP)
- DDIC modifications outside customer namespace
- Calls to non-released APIs (§6.3)

A translator that needs to emit any of these MUST fail loudly and request a
spec change.

---

## 13. Spec hash embedding (PCD A.20)

Per the PCD spec-hash rule, the SHA256 of the spec source MUST be embedded
in the generated package:

```abap
CONSTANTS c_spec_hash TYPE string VALUE `<64-char-hex>`.
```

This constant is declared on the application class. It is referenced by:

1. The synthesized `version` operation (radio button or subcommand
   `--version` equivalent), which writes the hash to the list.
2. `documentation/z_<name>.doc`, where the first 12 characters appear in
   the VERSION header.
3. The package's README.md as a Git-friendly verification anchor.

Drift detection (RULE-18 in pcd-lint) compares the embedded hash against
the on-disk spec at audit time.

---

## 14. Companion files

- `abap-report.abap.style.hints.md` — naming, formatting, idiomatic patterns
- `abap-report.abap.milestones.hints.md` — scaffold and acceptance patterns

The translator MUST consult both before emitting code.
