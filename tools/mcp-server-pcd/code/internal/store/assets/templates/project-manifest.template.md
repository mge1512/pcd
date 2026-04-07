
# project-manifest.template

## META
Deployment:  template
Version:     0.3.19
Spec-Schema: 0.3.19
Author:      Matthias G. Eckermann <pcd@mailbox.org>
License:     CC-BY-4.0
Verification: none
Safety-Level: QM
Template-For: project-manifest
EXECUTION:   none

---

> **Status: Work in Progress**
> This template is planned for v0.3.9. The definition below is a stub
> capturing agreed design decisions from whitepaper A.16. It is not yet
> complete enough for production use.

---

## TYPES

```
ComponentRef := {
  name:      string where non-empty,
  spec:      path where extension = ".md",
  interface: path where extension = ".md" | none
  // If interface is none, the full spec is used for import resolution.
  // Recommended: define a separate *.interface.md for stable contracts.
}

Dependency := {
  from: string,        // component name (must be in Components list)
  to:   string,        // component name (must be in Components list)
  via:  string | none  // interface spec path if importing a specific interface
}

BuildOrder := List<string>
// Component names in topological order.
// pcd-lint v2 validates this matches the dependency graph.
// Circular dependencies are an error.

InterfaceVersion := string where matches "^[0-9]+\.[0-9]+\.[0-9]+$"
// Semantic version of an exported interface.
// Versioned independently of the component implementation.

SystemInvariant := string where non-empty
// A formal invariant that spans multiple components.
// Must reference only types exported by component interfaces.
```

---

## TEMPLATE-TABLE

| Key | Value | Constraint | Notes |
|-----|-------|------------|-------|
| VERSION | MAJOR.MINOR.PATCH | required | Version of this project manifest. |
| SPEC-SCHEMA | MAJOR.MINOR.PATCH | required | |
| AUTHOR | name <email> | required | Repeating field. Project architect(s). |
| LICENSE | SPDX identifier | required | License covering the manifest and interface specs. |
| LANGUAGE | N/A | forbidden | No code is generated from a project manifest. |
| VERIFICATION | none | required | Project manifests are not formally verified in v1. |
| SAFETY-LEVEL | (inherit) | required | Effective safety level is the highest of all component safety levels. |
| BUILD-TOOL | make | default | Build orchestration. Alternatives: cmake, cargo-workspace. |
| AUDIT-BUNDLE | project-level | required | Aggregates all component audit bundles. |

---

## DELIVERABLES

A project manifest does not generate executable code. Its deliverables
are documentation and build orchestration artifacts.

### Deliverables Table

| OUTPUT-FORMAT | Constraint | Required Deliverable Files | Notes |
|---|---|---|---|
| dependency-graph | required | `dependency-graph.mmd` | Mermaid diagram of component dependencies. Human-reviewable. Renders on GitHub. |
| build-config | required | `Makefile` | Orchestrates per-component translation in build order. |
| interface-index | required | `interfaces/index.md` | Lists all exported interfaces, their versions, and importers. |
| project-audit-bundle | required | `audit_bundle/project/` | Aggregates all component audit bundles. Includes system invariant documentation. |
| report | required | `TRANSLATION_REPORT.md` | Documents decomposition decisions, dependency analysis, and system invariant coverage. |

---

## BEHAVIOR: validate-project
Constraint: required

*(Full BEHAVIOR specification pending — v0.3.13 target)*

Validates the project manifest and all referenced component specs:
- All component specs pass pcd-lint individually
- All imports resolve to valid interface specs
- No circular dependencies in the dependency graph
- Build order is consistent with the dependency graph
- System invariants reference only exported types

INPUTS:
```
manifest: path    // pcd-project.md
strict:   bool
```

STEPS:
*(Full STEPS specification pending — v0.3.13 target)*
1. Load and parse manifest file; on error → exit 2.
2. For each referenced component spec: run pcd-lint; collect errors.
3. Resolve all Imports; report unresolvable references.
4. Check dependency graph for cycles; report any found.
5. Compute topological build order; verify consistency.
6. Check system invariants reference only exported types.
7. If strict=true: treat warnings as errors.
8. If any errors → exit 1 with diagnostics. Else → exit 0.

POSTCONDITIONS:
- exit_code = 0 iff all checks pass
- exit_code = 1 iff any check fails
- exit_code = 2 iff manifest file not found or unreadable

---

## PRECONDITIONS

- A project manifest is authored by an architect, not a domain expert
- All component interfaces must be defined before component implementation
  specs are written — interfaces are the contract, implementations follow
- The manifest must declare a build order consistent with the dependency graph
- System invariants must reference only types exported via component interfaces
- Safety-Level of the manifest is the maximum of all component safety levels

---

## POSTCONDITIONS

- No code is generated from a project manifest directly
- Component specs are translated independently, in declared build order
- The project audit bundle aggregates all component audit bundles
- System invariants are documented in the project audit bundle
- A dependency graph Mermaid diagram is a required deliverable

---

## INVARIANTS

- [observable]      no circular dependencies permitted
- [observable]      build order must be a valid topological sort of the dependency graph
- [observable]      every component referenced in a Dependency must appear in Components
- [observable]      interface versions must be explicitly declared
- [observable]      a breaking interface change (major version bump) requires all importers
  to explicitly update their minimum version requirement
- [observable]      project Safety-Level = max(Safety-Level of all components)
- [observable]      template version is recorded in the project audit bundle

---

## EXAMPLES

EXAMPLE: minimal_two_component_project
GIVEN:
  pcd-project.md declares:
    Components:
      - name: account-service
        spec: components/account-service.md
        interface: interfaces/account-service.interface.md
      - name: transfer-service
        spec: components/transfer-service.md
    Dependencies:
      - from: transfer-service
        to: account-service
        via: interfaces/account-service.interface.md
    BuildOrder:
      - account-service
      - transfer-service
    SystemInvariants:
      - "GLOBAL: Σ(all Account.balance) is conserved across all services"
WHEN:
  pcd-lint validates the project manifest
THEN:
  all component specs pass pcd-lint individually
  all imports resolve correctly
  build order matches dependency graph
  system invariant references valid exported type (Account from account-service interface)
  exit_code = 0

EXAMPLE: circular_dependency_rejected
GIVEN:
  pcd-project.md Dependencies contains:
    - from: service-a, to: service-b
    - from: service-b, to: service-a
WHEN:
  pcd-lint validates the project manifest
THEN:
  stderr contains: "Circular dependency: service-a → service-b → service-a"
  exit_code = 1

EXAMPLE: build_order_inconsistent
GIVEN:
  Dependencies declare service-a depends on service-b
  BuildOrder declares service-a before service-b
WHEN:
  pcd-lint validates the project manifest
THEN:
  stderr contains: "Build order inconsistent with dependencies:
    service-a must come after service-b"
  exit_code = 1

---

## DEPLOYMENT

Runtime: this file is a template specification, not executable code.
Location: /usr/share/pcd/templates/project-manifest.template.md
Status: Work in progress — v0.3.9 target for completion.

Conventional project layout when using this template:

```
<project-root>/
├── pcd-project.md          ← the project manifest (uses this template)
├── interfaces/
│   ├── index.md             ← required deliverable: interface index
│   ├── account-service.interface.md
│   └── transfer-service.interface.md
├── components/
│   ├── account-service.md
│   └── transfer-service.md
├── dependency-graph.mmd     ← required deliverable: Mermaid diagram
├── Makefile                 ← required deliverable: build orchestration
└── audit_bundle/
    └── project/             ← required deliverable: aggregated audit bundle
```

The Makefile drives per-component translation in build order:

```makefile
.PHONY: all account-service transfer-service

all: account-service transfer-service

account-service:
	pcd-translate components/account-service.md

transfer-service: account-service
	pcd-translate components/transfer-service.md
```

