# abap-report — PCD Deployment Template

A PCD deployment template that translates a spec into an abapGit-compatible
ABAP package suitable for execution on a modern SAP application server.

## Targets

- SAP S/4HANA (on-premise, private cloud, public cloud)
- SAP BTP ABAP Environment ("Steampunk")
- Any NetWeaver 7.57+ system reachable via the ABAP Cloud development model

## When to use this template

- The behaviour belongs on the SAP application server itself: it operates on
  SAP master data, business objects, or other ABAP-resident state.
- The customer wants delivery via `abapGit pull` from a Git repository.
- The component runs in customer namespace (`Z*`, `Y*`, or a registered SAP
  namespace `/COMPANY/`).
- The component is invocable as an executable program (transaction code
  optional), a function module, or a class method exposed via RFC/OData.

## When NOT to use this template

| Situation                                                | Use instead          |
|----------------------------------------------------------|----------------------|
| Frontend / SAPUI5 / Fiori Elements                       | (future) `fiori-app` |
| Linux-side CLI tool talking to SAP via RFC or OData      | `cli-tool` with SAP as INTERFACE |
| Pure HANA-side computation (CDS, AMDP) with no ABAP entry | (future) `cds-view`  |
| Modification of SAP standard objects                     | not in scope — violates clean core |

## Output shape

```
README.md
LICENSE
.abapgit.xml
src/
  z<name>/
    package.devc.xml
    z_<name>.prog.abap                       (the report)
    z_<name>.prog.xml
    zcl_<name>.clas.abap                     (application class)
    zcl_<name>.clas.xml
    zcl_<name>.clas.testclasses.abap         (ABAP Unit tests)
    zcx_<name>.clas.abap                     (domain exception class)
    zcx_<name>.clas.xml
    zmsg_<name>.msag.xml                     (message class for error tokens)
    documentation/
      z_<name>.doc                           (report documentation — man-page equivalent)
```

For each external system declared in the spec's INTERFACES section, an
additional interface class `zif_<name>_<system>.intf.abap` is generated to
enable test doubles.

## Required deliverables

1. Executable report (`Z_<NAME>`)
2. Application class encapsulating all BEHAVIORs
3. Domain exception class carrying error tokens
4. Message class with one message per ERROR
5. ABAP Unit test class covering EXAMPLES and observable INVARIANTS
6. Report documentation (`z_<name>.doc`) — **mandatory**, per PCD man-page rule
7. abapGit metadata (`.abapgit.xml`, `package.devc.xml`)

## Companion files

| File                                          | Purpose                                    |
|-----------------------------------------------|--------------------------------------------|
| `TEMPLATE.md`                                 | Canonical translation rules (authoritative) |
| `abap-report.abap.style.hints.md`             | Code style and naming conventions          |
| `abap-report.abap.milestones.hints.md`        | Scaffold milestone and acceptance patterns |

## Spec-Schema compatibility

0.3.21 and later.

## Status

Initial draft. Not yet a registered PCD deployment template — see PCD
template registry for promotion criteria.
