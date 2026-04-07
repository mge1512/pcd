# backend-service.template

## META
Deployment:  template
Version:     0.3.20
Spec-Schema: 0.3.20
Author:      Matthias G. Eckermann <pcd@mailbox.org>
License:     CC-BY-4.0
Verification: none
Safety-Level: QM
Template-For: backend-service

---

## TYPES

```
Constraint := required | supported | default | forbidden

TemplateRow := {
  key:        string where non-empty,
  value:      string where non-empty,
  constraint: Constraint,
  notes:      string
}

TemplateTable := List<TemplateRow>

Platform := Linux | macOS | Windows

OutputFormat := RPM | DEB | OCI | binary

Language := Go | Rust | C | CPP | CSharp
```

---

## BEHAVIOR: resolve
Constraint: required

Given a spec declaring `Deployment: backend-service`, a translator reads
this template to determine defaults, constraints, and valid overrides
before generating any code or packaging.

INPUTS:
```
template: TemplateTable
spec_meta: Map<string, string>
preset:    Map<string, string>
```

OUTPUTS:
```
resolved: Map<string, string>
warnings: List<string>
errors:   List<string>
```

PRECONDITIONS:
- template is the backend-service template (Template-For = "backend-service")
- spec_meta contains at least Deployment, Verification, Safety-Level

STEPS:
1. Verify Template-For = "backend-service"; on mismatch → error, halt.
2. Merge preset layers in order: vendor → system → user → project (last writer wins).
3. For each constraint=required key K: if not resolved → errors += violation.
4. For each constraint=default key K: apply preset value if present, else template default.
5. For each constraint=forbidden key K: if present in spec_meta or any preset → errors += violation.
6. For each constraint=supported key K: apply if declared in spec_meta or preset; skip silently if absent.
7. Apply LANGUAGE precedence: project preset > user preset > system preset > template default.
8. Validate cross-key constraints (e.g. BINARY-TYPE vs LANGUAGE, PLATFORM vs OUTPUT-FORMAT).
   On violation → errors += constraint description.
9. If errors non-empty → return errors (reject, do not return resolved).
   Else → return resolved.

POSTCONDITIONS:
- resolved contains an effective value for every required key
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
7. If spec META declares a constraint=forbidden key → emit Error: "Key <K> is forbidden in backend-service specs."
8. Return merged result.

---

## TEMPLATE-TABLE

| Key | Value | Constraint | Notes |
|-----|-------|------------|-------|
| VERSION | MAJOR.MINOR.PATCH | required | Semantic versioning. |
| SPEC-SCHEMA | MAJOR.MINOR.PATCH | required | Version of the Post-Coding spec schema. |
| AUTHOR | name <email> | required | At least one Author: line required. |
| LICENSE | SPDX identifier | required | Must be a valid SPDX license identifier or compound expression. |
| LANGUAGE | Go | default | Default target language for Linux services. |
| LANGUAGE-ALTERNATIVES | Rust | supported | Via preset or project override. |
| LANGUAGE-ALTERNATIVES | C | supported | Via preset or project override. |
| LANGUAGE-ALTERNATIVES | C++ | supported | Via preset or project override. |
| LANGUAGE-ALTERNATIVES | C# | supported | Primarily for Windows service deployments. |
| BINARY-TYPE | static | default | Default: one static binary. |
| BINARY-TYPE | dynamic | supported | Permitted for C, C++, and C# only. |
| BINARY-COUNT | 1 | required | Exactly one service binary per spec. |
| RUNTIME-MODE | daemon | required | Long-running process managed by service supervisor. |
| HTTP-ENDPOINTS | local-only | default | Service may bind loopback or explicitly declared interfaces. |
| NETWORK-CALLS | declared-only | supported | Runtime network calls permitted only when declared by the spec's BEHAVIOR/DEPLOYMENT sections. |
| CONFIG-METHOD | file | required | Configuration is read from a declared file path. |
| CONFIG-METHOD | key=value-args | supported | CLI overrides may be supported if declared in spec DEPLOYMENT. |
| CONFIG-ENV-VARS | forbidden | forbidden | Behaviour must not be controlled via environment variables. |
| SIGNAL-HANDLING | SIGTERM | required | Graceful shutdown required. |
| SIGNAL-HANDLING | SIGINT | required | Graceful shutdown required. |
| OBSERVABILITY | stdout-stderr | required | Logs go to stdout/stderr for capture by journald or supervisor. |
| SERVICE-MANAGER | systemd | default | Primary Linux deployment target. |
| SERVICE-USER | dedicated | default | Run as a dedicated non-root user unless spec states otherwise. |
| OUTPUT-FORMAT | RPM | required | Linux RPM package. OBS build target. |
| OUTPUT-FORMAT | DEB | required | Linux DEB package. OBS build target. |
| OUTPUT-FORMAT | OCI | supported | OCI image for containerised single-service deployment. |
| OUTPUT-FORMAT | binary | supported | Raw binary for direct installation. |
| INSTALL-METHOD | OBS | required | Primary distribution via build.opensuse.org. |
| INSTALL-METHOD | curl | forbidden | curl-based installation scripts are not permitted. |
| PLATFORM | Linux | required | Primary and required platform. |
| PLATFORM | macOS | supported | Optional development/runtime target. |
| PLATFORM | Windows | supported | Optional alternative platform. |
| FILE-MODIFICATION | declared-state-only | supported | Service may modify only explicitly declared state files or sockets. |
| IDEMPOTENT | startup-config | required | Given the same config and inputs, startup behaviour is deterministic. |
| PRESET-SYSTEM | systemd-style | required | Preset layering follows systemd conventions. |

---

## PRECONDITIONS

- This template is applied only when spec META declares Deployment: backend-service
- Preset files must be valid TOML
- LANGUAGE value in resolved output must be one of: Go, Rust, C, C++, C#
- If LANGUAGE is C#, PLATFORM must include Windows
- If BINARY-TYPE is dynamic, LANGUAGE must be one of: C, C++, C#
- If LANGUAGE is Go or Rust, BINARY-TYPE must be static

---

## POSTCONDITIONS

- Every spec using Deployment: backend-service is governed by this template
- A spec may not declare LANGUAGE directly in META unless using Deployment: manual
- Resolved LANGUAGE is always one of the LANGUAGE-ALTERNATIVES or the default
- curl is never an accepted install method, regardless of preset override
- Forbidden constraints cannot be overridden by any preset or spec declaration

---

## INVARIANTS

- [observable]  constraint=forbidden rows cannot be overridden at any preset layer
- [observable]  constraint=required rows must resolve to a value; missing value is an error
- [observable]  LANGUAGE resolution always produces exactly one value
- [observable]  OUTPUT-FORMAT required rows must appear in every build configuration
- [observable]  a spec declaring Deployment: backend-service inherits all required constraints
- [observable]  BINARY-TYPE=dynamic is only valid when LANGUAGE ∈ {C, C++, C#}
- [observable]  BINARY-TYPE=static is the only valid value when LANGUAGE ∈ {Go, Rust}
- [observable]  services log to stdout/stderr rather than private log files

---

## EXAMPLES

EXAMPLE: minimal_service_resolution
GIVEN:
  spec META contains:
    Deployment: backend-service
    Verification: none
    Safety-Level: QM
  no preset files present
WHEN:
  resolved = resolve(template, spec_meta, preset={})
THEN:
  resolved["LANGUAGE"] = "Go"
  resolved["BINARY-TYPE"] = "static"
  resolved["SERVICE-MANAGER"] = "systemd"
  resolved["CONFIG-METHOD"] = "file"
  errors = []
  warnings = []

EXAMPLE: forbidden_curl_rejected
GIVEN:
  spec META contains:
    Deployment: backend-service
    INSTALL-METHOD: curl
WHEN:
  resolved = resolve(template, spec_meta, preset={})
THEN:
  errors contains:
    "Key INSTALL-METHOD=curl is forbidden for Deployment: backend-service"
  resolved is not produced

EXAMPLE: csharp_requires_windows
GIVEN:
  spec META contains:
    Deployment: backend-service
    Verification: none
    Safety-Level: QM
  preset declares LANGUAGE = "C#"
  preset declares PLATFORM = "Linux"
WHEN:
  resolved = resolve(template, spec_meta, preset={LANGUAGE: "C#"})
THEN:
  errors contains:
    "LANGUAGE C# requires PLATFORM: Windows"
  resolved is not produced

---

## DELIVERABLES

Defines the files a translator must produce for each OUTPUT-FORMAT
declared as `required` or active `supported`.

### Delivery Order

1. Core implementation files
2. Service and packaging artifacts
3. Supported packaging artifacts if preset active
4. Test infrastructure
5. Documentation
6. TRANSLATION_REPORT.md last, after all other files are written and verified

### Deliverables Table

| OUTPUT-FORMAT | Constraint | Required Deliverable Files | Notes |
|---|---|---|---|
| source | required | `main.go` or `cmd/<n>/main.go`, `go.mod` | Single binary service. Split files only when complexity requires it. |
| build | required | `Makefile` | Must include `build`, `test`, `install`, `clean`, `man` targets. `man` target: `pandoc <n>.1.md -s -t man -o <n>.1`. |
| service | required | `<n>.service` | systemd unit for Linux service deployment. |
| docs | required | `README.md` | Must document installation, config file path, arguments, endpoints, and signal handling. |
| man | required | `<n>.1.md`, `<n>.1` | Markdown source converted to troff via `pandoc`. Section 1 (user commands). Install to `%{_mandir}/man1/` (RPM) and `usr/share/man/man1/` (DEB). |
| license | required | `LICENSE` | SPDX identifier from spec META + authoritative URL to the full license text. Never reproduce the full license text. |
| RPM | required | `<n>.spec` | OBS RPM spec file. Must install the binary, service unit, and default config path if spec declares one. |
| DEB | required | `debian/control`, `debian/changelog`, `debian/rules`, `debian/copyright` | Debian package metadata. Must install the service unit. |
| OCI | supported | `Containerfile` | Multi-stage build. Builder stage: `FROM registry.suse.com/bci/golang:latest AS builder` for Go — never unqualified names (supply chain security requirement). Final stage: `FROM scratch`. Expose only ports declared in spec DEPLOYMENT. |
| binary | supported | none | Raw binary only. |
| report | required | `TRANSLATION_REPORT.md` | Must include service startup/shutdown notes, config-loading notes, and compile gate result. |

### Systemd Service Unit

For Linux/systemd deployment, install a service unit in the package:

```ini
[Unit]
Description={component name}
After=network.target

[Service]
Type=simple
ExecStart=/usr/bin/{n} config=/etc/{n}/config.json
Restart=on-failure
NoNewPrivileges=true
User={n}
Group={n}

[Install]
WantedBy=multi-user.target
```

Adjust `ExecStart` and config path only if the spec DEPLOYMENT section declares
different values. Do not add environment-variable based configuration.

### Deliverable Content Requirements

**RPM spec (`<n>.spec`):**
- `BuildRequires:` must include `pandoc` (for man page generation)
- `%build` must include: `pandoc <n>.1.md -s -t man -o <n>.1`
- `%files` must include: `%{_mandir}/man1/<n>.1*`

**debian/control:**
- `Build-Depends` must include `pandoc`

**debian/rules:**
- Must include man page build step: `pandoc <n>.1.md -s -t man -o <n>.1`
- Man page must be installed to `usr/share/man/man1/<n>.1`

**Containerfile (when OCI is active in preset):**
- Must use multi-stage build: builder stage + minimal final stage
- Builder stage must use `FROM registry.suse.com/bci/golang:latest AS builder`
  for Go. Never use unqualified names such as `golang:1.24` or `docker.io/golang`.
  This is a supply chain security requirement, not a preference.
- Final stage must be `FROM scratch` (static binary; no runtime dependencies)
- Layer order in builder stage: `COPY go.mod go.sum ./` → `RUN go mod download`
  → `COPY . .` → `RUN CGO_ENABLED=0 go build`
- Expose only ports declared in the spec DEPLOYMENT section
- Do not include a package manager in the final image

---

## DEPLOYMENT

Runtime: long-running service process, typically supervised by systemd.

Naming convention:
  Binary name: {n}
  Service unit: {n}.service
  Default config directory: /etc/{n}/

Packaging:
  RPM and DEB are mandatory.
  OCI is optional and produced only when active in preset.

---

## EXECUTION

The translator must read this section before generating any code.
It specifies the exact delivery phases, resume logic, and compile gate
for backend-service components. Follow it exactly.

### Input files

The translator receives in the working directory:
- `backend-service.template.md` — this deployment template
- `<spec-name>.md` — the component specification

If the spec's DEPENDENCIES section references hints files, they are also
present. Read them before writing `go.mod` or any code that uses those
libraries.

### Resume logic

Before writing any file, list the output directory.
If a listed deliverable already exists and is non-empty, skip it — treat
it as complete and move to the next missing file. Report which files were
found and which are being produced.

### Delivery phases

Produce files in this exact order. Complete each phase before starting
the next. Do not produce `TRANSLATION_REPORT.md` until Phase 5 is done.

**Phase 1 — Core implementation**
- All source files
- `go.mod`

**Phase 2 — Service and packaging**
- `Makefile`
- `<n>.service`
- `<n>.spec`
- `debian/control`, `debian/changelog`, `debian/rules`, `debian/copyright`
- `Containerfile` (if OCI is active in preset)
- `LICENSE`

**Phase 3 — Test infrastructure**
- `independent_tests/INDEPENDENT_TESTS.go`
- `translation_report/translation-workflow.pikchr`

**Phase 4 — Documentation**
- `README.md`

**Phase 5 — Compile gate**

**Phase 6 — Report (last)**
- `TRANSLATION_REPORT.md`

### Compile gate

Execute after Phase 4 and before Phase 6. If your environment cannot
execute shell commands, document this explicitly in `TRANSLATION_REPORT.md`.

**Step 1 — Dependency resolution**

Run: `go mod tidy`

If `go mod tidy` cannot be run:
- Produce `go.mod` with direct dependencies only, no `go.sum`
- Note in `TRANSLATION_REPORT.md` that `go mod tidy` must be run before building

**Step 2 — Compilation**

Run: `go build ./...`

If compilation fails, fix only the identified errors and re-run.

**Step 3 — Record result**

Record pass/fail for each step in `TRANSLATION_REPORT.md`.
Once all steps pass, do not modify any source files further.

