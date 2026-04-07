

# cloud-native.template

## META
Deployment:  template
Version:     0.3.19
Spec-Schema: 0.3.19
Author:      Matthias G. Eckermann <pcd@mailbox.org>
License:     CC-BY-4.0
Verification: none
Safety-Level: QM
Template-For: cloud-native

---

## TYPES

```
Constraint := required | supported | default | forbidden

TemplateRow := {
  key:        string where non-empty,
  value:      string where non-empty,
  constraint: Constraint,
  notes:      string         // human-readable explanation; may be empty
}

TemplateTable := List<TemplateRow>
// Rows with identical key are collected as a list for that key.
// Order within repeated keys is not significant.

Platform := Linux | arm64
// Cloud-native targets Kubernetes, which runs primarily on Linux

OutputFormat := OCI | HELM | KUSTOMIZE | MANIFEST | OPERATOR
// OCI = container image
// HELM = Helm chart package
// KUSTOMIZE = Kustomization files
// MANIFEST = raw Kubernetes YAML manifests
// OPERATOR = Kubernetes operator with CRDs

Language := Go | Rust | Java | Python
// Go is default for cloud-native/CNCF ecosystem

BaseImage := SLE-BCI | Distroless | Scratch
// SLE-BCI = SUSE Linux Enterprise Base Container Images from registry.suse.com
```

---

## BEHAVIOR: resolve
Constraint: required

Given a spec declaring `Deployment: cloud-native`, a translator reads this
template to determine defaults, constraints, and valid overrides before
generating any code or build configuration.

INPUTS:
```
template: TemplateTable
spec_meta: Map<string, string>    // the META fields from the spec
preset:    Map<string, string>    // merged preset (system + user + project)
```

OUTPUTS:
```
resolved: Map<string, string>     // effective settings for this build
warnings: List<string>            // advisory messages to surface
errors:   List<string>            // constraint violations; non-empty → reject
```

PRECONDITIONS:
- template is the cloud-native template (Template-For = "cloud-native")
- spec_meta contains at least Deployment, Verification, Safety-Level

STEPS:
1. Verify Template-For = "cloud-native"; on mismatch → error, halt.
2. Merge preset layers in order: vendor → system → user → project (last writer wins).
3. For each constraint=required key K: if not resolved → errors += violation.
4. For each constraint=default key K: apply preset value if present, else template default.
5. For each constraint=forbidden key K: if present in spec_meta or any preset → errors += violation.
6. For each constraint=supported key K: apply if declared in spec_meta or preset; skip silently if absent.
7. Apply LANGUAGE precedence: project preset > user preset > system preset > template default.
8. Validate cross-key constraints (e.g. BASE-IMAGE=Scratch requires LANGUAGE ∈ {Go, Rust};
   OUTPUT-FORMAT=OPERATOR requires CRDs declared in spec).
   On violation → errors += constraint description.
9. If errors non-empty → return errors (reject, do not return resolved).
   Else → return resolved.

POSTCONDITIONS:
- resolved contains an effective value for every required key
- for each key K with constraint=required: resolved[K] is set, else errors += violation
- for each key K with constraint=default: resolved[K] = preset[K] if present,
  else resolved[K] = template default value for K
- for each key K with constraint=forbidden: if spec_meta contains K,
  errors += "Key <K> is forbidden for Deployment: cloud-native"
- for each key K with constraint=supported: resolved[K] set only if
  spec_meta or preset declares it; no error if absent
- resolved["LANGUAGE"] follows precedence:
    project preset > user preset > system preset > template default

---

## BEHAVIOR/INTERNAL: precedence-resolution
Constraint: required

Defines how conflicting values across layers are resolved for any key.

STEPS:
1. Start with template defaults as the base map.
2. Merge /usr/share/pcd/presets/ values (vendor defaults); later entries override earlier.
3. Merge /etc/pcd/presets/ values (system admin); overrides vendor defaults.
4. Merge ~/.config/pcd/presets/ values (user); overrides system.
5. Merge <project-dir>/.pcd/ values (project-local); overrides user.
6. For each key in spec META: if constraint=supported → apply; if constraint=required or default →
   emit Warning: "Spec overrides template default for <K>. Ensure this is intentional."
7. If spec META declares a constraint=forbidden key → emit Error: "Key <K> is forbidden in cloud-native specs."
8. Return merged result.

Resolution order (last writer wins):
  1. template default
  2. /usr/share/pcd/presets/    (vendor default)
  3. /etc/pcd/presets/          (system administrator)
  4. ~/.config/pcd/presets/     (user)
  5. <project-dir>/.pcd/        (project-local, committed to git)
  6. spec META explicit override        (only permitted for constraint=supported keys)

If spec META declares a value for a constraint=required or constraint=default key,
emit Warning: "Spec overrides template default for <K>. Ensure this is intentional."

If spec META declares a value for a constraint=forbidden key,
emit Error: "Key <K> is forbidden in cloud-native specs."

---

## TEMPLATE-TABLE

| Key | Value | Constraint | Notes |
|-----|-------|------------|-------|
| VERSION | MAJOR.MINOR.PATCH | required | Semantic versioning. Spec author increments on every meaningful change. |
| SPEC-SCHEMA | MAJOR.MINOR.PATCH | required | Version of the Post-Coding spec schema this file was written against. |
| AUTHOR | name <email> | required | At least one Author: line required. Repeating key; multiple authors permitted. |
| LICENSE | SPDX identifier | required | Must be a valid SPDX license identifier or compound expression. Example: Apache-2.0. |
| LANGUAGE | Go | default | Default target language for cloud-native applications. Override via preset. |
| LANGUAGE-ALTERNATIVES | Rust | supported | May be selected via preset or project override. Good for performance-critical services. |
| LANGUAGE-ALTERNATIVES | Java | supported | May be selected via preset or project override. Common in enterprise environments. |
| LANGUAGE-ALTERNATIVES | Python | supported | May be selected via preset or project override. Suitable for data processing services. |
| BASE-IMAGE | SLE-BCI | default | SUSE Linux Enterprise Base Container Images from registry.suse.com. Supply chain security optimized. |
| BASE-IMAGE | Distroless | supported | Google distroless images for minimal attack surface. |
| BASE-IMAGE | Scratch | supported | Empty base image for static binaries only. |
| REGISTRY | registry.suse.com | default | Default container registry for base images. |
| CONTAINER-RUNTIME | containerd | required | OCI-compliant container runtime. Docker compatibility layer. |
| KUBERNETES-VERSION | v1.28+ | required | Minimum supported Kubernetes version. |
| PLATFORM | Linux | required | Linux is the primary and required platform for Kubernetes. |
| PLATFORM | arm64 | supported | ARM64 support for cloud efficiency and cost optimization. |
| OUTPUT-FORMAT | OCI | required | OCI container image. Primary deliverable for cloud-native applications. |
| OUTPUT-FORMAT | MANIFEST | required | Raw Kubernetes YAML manifests in deploy/ directory. Always required for kubectl apply. |
| OUTPUT-FORMAT | HELM | supported | Helm chart for complex deployments with dependencies. |
| OUTPUT-FORMAT | KUSTOMIZE | supported | Kustomization files for GitOps workflows. |
| OUTPUT-FORMAT | OPERATOR | supported | Kubernetes operator with Custom Resource Definitions (CRDs). |
| HEALTH-CHECKS | required | required | Kubernetes liveness, readiness, and startup probes must be implemented. |
| OBSERVABILITY | metrics | required | Prometheus-compatible metrics endpoint required. |
| OBSERVABILITY | logging | required | Structured logging to stdout/stderr required. |
| OBSERVABILITY | tracing | supported | OpenTelemetry tracing support. |
| CONFIG-METHOD | configmap | required | Configuration via Kubernetes ConfigMaps. |
| CONFIG-METHOD | env-vars | supported | Environment variable configuration for simple cases. |
| CONFIG-METHOD | files | forbidden | Direct file system configuration not permitted in containers. |
| SECRETS-METHOD | k8s-secrets | required | Sensitive data via Kubernetes Secrets. |
| SECRETS-METHOD | vault | supported | HashiCorp Vault integration for advanced secret management. |
| SECRETS-METHOD | files | forbidden | Direct file system secrets not permitted. |
| NETWORK-POLICY | required | required | Kubernetes NetworkPolicy definitions required for security. |
| SERVICE-MESH | istio | supported | Istio service mesh integration. |
| SERVICE-MESH | linkerd | supported | Linkerd service mesh integration. |
| RBAC | required | required | Kubernetes RBAC (Role-Based Access Control) definitions required. |
| SECURITY-CONTEXT | non-root | required | Containers must run as non-root user. |
| SECURITY-CONTEXT | read-only-fs | required | Root filesystem must be read-only. |
| GRACEFUL-SHUTDOWN | SIGTERM | required | Handle SIGTERM for graceful shutdown within 30 seconds. |
| RESOURCE-LIMITS | required | required | CPU and memory limits must be defined. |
| HORIZONTAL-SCALING | HPA | supported | Horizontal Pod Autoscaler configuration. |
| VERTICAL-SCALING | VPA | supported | Vertical Pod Autoscaler configuration. |
| PERSISTENCE | none | default | Stateless by default. Use external storage services. |
| PERSISTENCE | pvc | supported | Persistent Volume Claims for stateful applications. |
| INSTALL-METHOD | helm | supported | Installation via Helm package manager. |
| INSTALL-METHOD | kubectl | required | Direct kubectl apply installation. |
| INSTALL-METHOD | curl | forbidden | curl-based installation scripts are not permitted. Supply chain security requirement. |

---

## TYPE-BINDINGS

Maps logical spec types to ecosystem-canonical Go types when LANGUAGE=Go.
The translator applies this table mechanically after resolving LANGUAGE.
Spec authors must never reference these language-specific type names directly.

| Spec Type   | LANGUAGE=Go                                    | Notes                                     |
|-------------|------------------------------------------------|-------------------------------------------|
| Duration    | metav1.Duration (k8s.io/apimachinery/meta/v1)  | Serialises as "60s", "5m", "1h" etc.     |
| Timestamp   | metav1.Time                                    | Serialises as RFC3339                     |
| Condition   | metav1.Condition                               | Use standard k8s condition type           |
| List\<T\>   | []T                                            |                                           |
| ObjectRef   | corev1.ObjectReference                         | Cross-namespace references                |
| LabelSet    | map[string]string                              | Standard Kubernetes label map             |
| bytes       | []byte                                         | Raw byte sequences (e.g. SSH key material)|

## GENERATED-FILE-BINDINGS

Maps logical TOOLCHAIN-CONSTRAINTS generated file names to language-specific
filenames and tools when LANGUAGE=Go. When the declared tool is unavailable in
the translation environment, the translator must hand-author a functionally
correct equivalent — the file must always be present.

| Logical name              | LANGUAGE=Go filename                    | Generator tool                          |
|---------------------------|-----------------------------------------|-----------------------------------------|
| type-marshaling-deepcopy  | `api/v1alpha1/zz_generated.deepcopy.go` | `controller-gen object paths="./..."`   |
| dependency-lock-file      | `go.sum`                                | `go mod tidy`                           |

---

## PRECONDITIONS

- This template is applied only when spec META declares Deployment: cloud-native
- Kubernetes cluster must be version 1.28 or later
- If BASE-IMAGE is Scratch, LANGUAGE must be Go or Rust (static linking required)
- If PERSISTENCE includes pvc, spec must declare storage requirements
- If SERVICE-MESH is declared, corresponding mesh must be installed in cluster
- If OUTPUT-FORMAT includes OPERATOR, spec must define custom resources
- LANGUAGE value in resolved output must be one of: Go, Rust, Java, Python

---

## POSTCONDITIONS

- Every spec using Deployment: cloud-native is governed by this template
- A spec may not declare LANGUAGE directly in META unless using Deployment: manual
- Resolved LANGUAGE is always one of the LANGUAGE-ALTERNATIVES or the default
- curl is never an accepted install method, regardless of preset override
- Forbidden constraints cannot be overridden by any preset or spec declaration
- All containers run as non-root with read-only filesystem
- Health checks and observability are always configured
- Kubernetes manifests are always generated in deploy/ directory

---

## INVARIANTS

- [observable]      constraint=forbidden rows cannot be overridden at any preset layer
- [observable]      constraint=required rows must resolve to a value; missing value is an error
- [observable]      LANGUAGE resolution always produces exactly one value
- [observable]      OUTPUT-FORMAT required rows must appear in every build configuration
- [observable]      a spec declaring Deployment: cloud-native inherits all required constraints
  whether or not the spec author is aware of them
- [observable]      template version is recorded in every audit bundle that references it
- [observable]      BASE-IMAGE=Scratch is only valid when LANGUAGE ∈ {Go, Rust}
- [observable]      containers must always run as non-root user
- [observable]      health checks are mandatory for all cloud-native applications
- [observable]      deploy/ directory must contain all required Kubernetes manifests

---

## EXAMPLES

*Note: specs using Deployment: cloud-native frequently target Kubernetes reconcilers.
EXAMPLES in such specs may use multiple WHEN/THEN pairs to express multi-pass
reconciliation behaviour. Each WHEN/THEN pair represents one reconciler invocation.
Single-pass EXAMPLES remain valid for non-reconciler operations.*

EXAMPLE: minimal_cloud_native_spec
GIVEN:
  spec META contains:
    Deployment: cloud-native
    Verification: none
    Safety-Level: QM
  no preset files present (system defaults only)
WHEN:
  resolved = resolve(template, spec_meta, preset={})
THEN:
  resolved["LANGUAGE"] = "Go"
  resolved["BASE-IMAGE"] = "SLE-BCI"
  resolved["REGISTRY"] = "registry.suse.com"
  resolved["OUTPUT-FORMAT"] = "OCI"
  resolved["HEALTH-CHECKS"] = "required"
  resolved["SECURITY-CONTEXT"] = "non-root"
  errors = []
  warnings = []

EXAMPLE: rust_with_distroless_override
GIVEN:
  spec META contains:
    Deployment: cloud-native
    Verification: none
    Safety-Level: QM
  /etc/pcd/presets/org.toml contains:
    [templates.cloud-native]
    language = "rust"
    base_image = "distroless"
WHEN:
  resolved = resolve(template, spec_meta, preset={LANGUAGE: "Rust", BASE-IMAGE: "Distroless"})
THEN:
  resolved["LANGUAGE"] = "Rust"
  resolved["BASE-IMAGE"] = "Distroless"
  errors = []
  warnings = []

EXAMPLE: forbidden_file_config_rejected
GIVEN:
  spec META contains:
    Deployment: cloud-native
    CONFIG-METHOD: files
WHEN:
  resolved = resolve(template, spec_meta, preset={})
THEN:
  errors contains:
    "Key CONFIG-METHOD=files is forbidden for Deployment: cloud-native"
  resolved is not produced (errors non-empty → reject)

EXAMPLE: scratch_base_requires_static_language
GIVEN:
  spec META contains:
    Deployment: cloud-native
    Verification: none
    Safety-Level: QM
  preset declares BASE-IMAGE = Scratch and LANGUAGE = Java
WHEN:
  resolved = resolve(template, spec_meta, preset={BASE-IMAGE: "Scratch", LANGUAGE: "Java"})
THEN:
  errors contains:
    "BASE-IMAGE Scratch requires LANGUAGE ∈ {Go, Rust} for static linking"
  resolved is not produced

EXAMPLE: operator_format_requires_crds
GIVEN:
  spec META contains:
    Deployment: cloud-native
    Verification: none
    Safety-Level: QM
  preset declares OUTPUT-FORMAT includes OPERATOR
  spec DEPLOYMENT section does not describe custom resources
WHEN:
  translator processes spec
THEN:
  warnings contains:
    "OUTPUT-FORMAT=OPERATOR declared but no custom resource definitions found in spec. \
     Ensure CRDs and controller logic are specified."

---

## DELIVERABLES

Defines the files a translator must produce for each OUTPUT-FORMAT
declared as `required` or `supported` in the TEMPLATE-TABLE.
A translator must produce all deliverables for every `required`
OUTPUT-FORMAT. For `supported` OUTPUT-FORMATs, deliverables are
produced only if that format is active in the resolved preset.

The prompt to the translator must not enumerate these files —
the translator derives them from this section.

### Delivery Order

Deliverables must be produced in the following order:
1. Core implementation files (source, go.mod, Containerfile, README.md, LICENSE)
2. Required Kubernetes artifacts (deploy/ directory with all manifests)
3. Supported packaging artifacts if preset active (Helm, Kustomize, Operator)
4. Independent tests in independent_tests/ subdirectory
5. Pikchr workflow diagram in translation_report/ subdirectory
6. TRANSLATION_REPORT.md last, after all other files are written and verified

### Deliverables Table

| OUTPUT-FORMAT | Constraint | Required Deliverable Files | Notes |
|---|---|---|------|
| source | required | `main.go`, `go.mod` (Go) or equivalent for other languages | Application source code. Must implement health check endpoints. |
| container | required | `Containerfile` | OCI-compliant container build file. Multi-stage build required. Must use declared BASE-IMAGE. |
| docs | required | `README.md` | Must document: deployment to Kubernetes, configuration options, health check endpoints, scaling considerations. |
| license | required | `LICENSE` | SPDX identifier from spec META + authoritative URL to the full license text. Never reproduce the full license text. |
| OCI | required | none beyond container | Primary deliverable is the Containerfile. |
| MANIFEST | required | `deploy/deployment.yaml`, `deploy/service.yaml`, `deploy/configmap.yaml`, `deploy/networkpolicy.yaml`, `deploy/rbac.yaml` | Core Kubernetes resources in deploy/ directory. Ready for kubectl apply -f deploy/. |
| HELM | supported | `helm/Chart.yaml`, `helm/values.yaml`, `helm/templates/` | Helm chart structure. templates/ must contain templated versions of Kubernetes manifests. |
| KUSTOMIZE | supported | `kustomize/kustomization.yaml`, `kustomize/base/`, `kustomize/overlays/` | Kustomization structure for GitOps. Base contains common resources, overlays for environments. |
| OPERATOR | supported | `deploy/crd.yaml`, `controllers/` | Kubernetes operator with Custom Resource Definitions. Controllers directory contains operator logic. The operator's own deployment manifest is `deploy/deployment.yaml` (already produced by the MANIFEST row); no separate `deploy/operator.yaml` is required. |
| independent-tests | required | `independent_tests/INDEPENDENT_TESTS.go` | Specification verification tests using declared INTERFACES test doubles. For LANGUAGE=Go: test functions must reside in a companion `independent_tests/independent_tests_test.go` file (Go requires `_test.go` suffix for test execution); `INDEPENDENT_TESTS.go` serves as the package documentation file. |
| workflow-diagram | required | `translation_report/translation-workflow.pikchr` | Pikchr diagram documenting the translation process and decisions. |
| report | required | `TRANSLATION_REPORT.md` | AI translator self-evaluation with cloud-native specific considerations. |

### Naming Convention

`<n>` in the above table refers to the component name as declared
in the specification title (first `#` heading). It must be:
- lowercase
- hyphen-separated (no underscores)
- no version suffix in the filename itself (version lives inside the file)

### Deliverable Content Requirements

**Containerfile:**
- Must use multi-stage build: builder stage + minimal final stage
- Final stage must use the resolved BASE-IMAGE value
- Must not run as root user (USER directive required)
- Must not expose unnecessary ports
- Must NOT include HEALTHCHECK instruction (not supported by OCI format; use Kubernetes liveness/readiness probes in deployment.yaml instead)
- Layer order in builder stage must be: copy dependency manifest AND dependency lock
  file first, then run dependency download, then copy source. This enables layer
  caching and ensures the build inside the container matches the verified local build.
  (For LANGUAGE=Go: `COPY go.mod go.sum ./` then `RUN go mod download` then `COPY . .`)
- For SLE-BCI base: `FROM registry.suse.com/bci/golang:latest AS builder`

**deploy/deployment.yaml:**
- Must include livenessProbe, readinessProbe, and startupProbe
- Must specify resource requests and limits
- Must use non-root securityContext
- Must set readOnlyRootFilesystem: true
- Must reference ConfigMap and Secrets appropriately

**deploy/service.yaml:**
- Must expose only necessary ports
- Must include appropriate service type (ClusterIP default)
- Must include proper selectors matching deployment labels

**deploy/networkpolicy.yaml:**
- Must implement least-privilege network access
- Must deny all traffic by default, allow only required connections
- Must document allowed ingress and egress rules

**deploy/rbac.yaml:**
- Must include ServiceAccount, Role, and RoleBinding
- Must follow principle of least privilege
- Must grant only permissions required by the application

**deploy/crd.yaml (if OPERATOR format active):**
- Must define Custom Resource Definitions for the operator
- Must include proper OpenAPI v3 schema validation
- Must follow Kubernetes API conventions

**TRANSLATION_REPORT.md:**
- Must include cloud-native specific sections:
  - Container image selection rationale
  - Kubernetes resource sizing decisions
  - Security context configuration
  - Health check implementation approach
  - Configuration and secrets management strategy
  - Network policy design decisions
  - RBAC permission justification
  - Operator design decisions (if OPERATOR format active)

---

## EXECUTION

The translator must read this section before generating any code.
It specifies the exact delivery phases, resume logic, and compile/build
gate for cloud-native components. Follow it exactly.

### Input files

The translator receives in the working directory:
- `cloud-native.template.md` — this deployment template
- `<spec-name>.md` — the component specification

If the spec's DEPENDENCIES section references hints files, they are also
present. Read them before writing `go.mod` or any code that uses those
libraries — they contain verified dependency version strings.

### Resume logic

Before writing any file, list the output directory.
If a listed deliverable already exists and is non-empty, skip it — treat
it as complete and move to the next missing file. Report which files were
found and which are being produced.

### Delivery phases

Produce files in this exact order. Complete each phase before starting
the next. Do not produce `TRANSLATION_REPORT.md` until Phase 5 is done.

**Phase 1 — Core implementation**
- All source files for the resolved LANGUAGE (e.g. `main.go` + helpers for Go)
- `go.mod` — direct dependencies only; see Compile gate below

**Phase 2 — Build and packaging**
- `Containerfile`
- `LICENSE`

**Phase 3 — Kubernetes manifests**
- `deploy/deployment.yaml`
- `deploy/service.yaml`
- `deploy/configmap.yaml`
- `deploy/networkpolicy.yaml`
- `deploy/rbac.yaml`
- `deploy/crd.yaml` (if OPERATOR format is active)
- `deploy/operator.yaml` (if OPERATOR format is active)
- Helm chart under `helm/` (if HELM format is active)
- Kustomize under `kustomize/` (if KUSTOMIZE format is active)

**Phase 4 — Test infrastructure**
- `independent_tests/INDEPENDENT_TESTS.go`
- `translation_report/translation-workflow.pikchr`

**Phase 5 — Documentation**
- `README.md`

**Phase 6 — Compile gate and build gate** (see below)

**Phase 7 — Report (last)**
- `TRANSLATION_REPORT.md`

### Compile gate

Execute after Phase 5 and before Phase 7. If your environment cannot
execute shell commands, document this explicitly under the heading
"Phase 6 — Compile gate not executed" in TRANSLATION_REPORT.md and
state why. Do not silently omit this phase.

**Step 1 — Dependency resolution**

Run: `go mod tidy`

Do not hand-write indirect dependencies — they must come from `go mod tidy`.

If `go mod tidy` cannot be run:
- Produce `go.mod` with direct dependencies only, no `go.sum`
- Note in TRANSLATION_REPORT.md that `go mod tidy` must be run before building

**Step 2 — Compilation**

Run: `go build ./...`

If compilation fails, fix only the identified errors and re-run.
Do not rewrite unaffected files. Repeat until compilation succeeds
or all reasonable fixes are exhausted.

**Step 3 — Build gate (OCI)**

After compilation succeeds, consult the `## BUILD-GATE` section of this
template for container image build verification. Execute the steps there.

**Step 4 — Record result**

Record pass/fail for each step in TRANSLATION_REPORT.md.
Once all steps pass, do not modify any source files further.
Proceed immediately to Phase 7.

### go.mod rules

- Declare only direct dependencies (those your code imports directly)
- Do NOT hand-write indirect dependencies (resolved by `go mod tidy`)
- Do NOT fabricate pseudo-versions or commit hashes for untagged modules
  If hints files are present: use the verified versions they provide
  If no hints file: flag the dependency in TRANSLATION_REPORT.md as
  requiring manual version verification before building


---

## BUILD-GATE

Specifies which output formats require an actual build attempt as part of
translation verification, beyond the compile gate. The translator prompt
instructs the translator to consult this section after compilation succeeds.

If no suitable tool is available for a required build, the translator must
document this explicitly in TRANSLATION_REPORT.md under
"Build gate — \<FORMAT\> not executed" and state which tools were tried.
Silent omission is not permitted.

Once a build succeeds, do not modify any source files further.

### OCI (if OUTPUT-FORMAT: OCI is active)

Attempt to build the container image from the `Containerfile`.

**Acceptable tools** (try in order; use the first available):

| Tool | Command |
|---|---|
| `podman` | `podman build -f Containerfile -t <name>:<version> .` |
| `docker` | `docker build -f Containerfile -t <name>:<version> .` |
| `buildah` | `buildah bud -f Containerfile -t <name>:<version> .` |

where `<name>` is the component name (lowercase, hyphenated) and `<version>`
is the version from the spec META section.

If the project also contains a compose file (`compose.yml` or
`docker-compose.yml`), the following tools may be used instead:

| Tool | Command |
|---|---|
| `podman compose` | `podman compose build` |
| `docker compose` | `docker compose build` |

If the build fails, treat it as a compile gate failure: fix only the
identified files (typically the Containerfile or build configuration)
and retry. Do not rewrite source files.

---

## DEPLOYMENT

Runtime: this file is a template specification, not executable code.
It is read by pcd-lint (for template resolution validation) and by
AI translators (for code generation context).

Location in preset hierarchy:
  /usr/share/pcd/templates/cloud-native.template.md

Versioning:
  Template version is declared in META (Version: field).
  Specs reference the template by name (Deployment: cloud-native).
  Audit bundles record the template version used at generation time.
  Breaking changes to a template increment the minor version.
  Additions of supported rows are non-breaking.
  Changes to required or forbidden rows are breaking.
  Current version: 0.3.14

