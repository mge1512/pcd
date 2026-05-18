



# kubectl-style-cli.template

## META
Deployment:   template
Version:      0.1.0
Spec-Schema:  0.3.21
Author:       François-Xavier Houard <fx.houard@gmail.com>
License:      CC-BY-4.0
Verification: none
Safety-Level: QM
Template-For: kubectl-style-cli

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

Platform := Linux | macOS | Windows

PackageFormat := RPM | DEB | OCI | PKG | MSI | binary
// binary = raw executable, no packaging

Language := Go | Rust

CLIArgStyle := subcommand-flag | key=value
// subcommand-flag = `<bin> verb resource --flag=value` (kubectl convention)
// key=value       = `<bin> verb key=value` (simple-cli-tool convention)

FlagAssignment := equals | space | both
// equals = only `--flag=value`
// space  = only `--flag value`
// both   = accept either form

OutputFormat := human | json | yaml | jsonpath | wide | name | custom

AuthStorage := os-keyring | xdg-file | plaintext-file | none

Shell := bash | zsh | fish | powershell
```

---

## BEHAVIOR: resolve
Constraint: required

Given a spec declaring `Deployment: kubectl-style-cli`, a translator reads
this template to determine defaults, constraints, and valid overrides
before generating any code or build configuration.

INPUTS:
```
template:  TemplateTable
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
- template is the kubectl-style-cli template (Template-For = "kubectl-style-cli")
- spec_meta contains at least Deployment, Verification, Safety-Level

STEPS:
1. Verify Template-For = "kubectl-style-cli"; on mismatch → error, halt.
2. Merge preset layers in order: vendor → system → user → project (last writer wins).
3. For each constraint=required key K: if not resolved → errors += violation.
4. For each constraint=default key K: apply preset value if present, else template default.
5. For each constraint=forbidden key K: if present in spec_meta or any preset → errors += violation.
6. For each constraint=supported key K: apply if declared in spec_meta or preset; skip silently if absent.
7. Apply LANGUAGE precedence: project preset > user preset > system preset > template default.
8. Validate cross-key constraints (BINARY-TYPE vs LANGUAGE, PLATFORM vs PACKAGE-FORMAT,
   CLI-ARG-STYLE consistency with subcommand structure).
   On violation → errors += constraint description.
9. If errors non-empty → return errors (reject, do not return resolved).
   Else → return resolved.

POSTCONDITIONS:
- resolved contains an effective value for every required key
- for each key K with constraint=required: resolved[K] is set, else errors += violation
- for each key K with constraint=default: resolved[K] = preset[K] if present,
  else resolved[K] = template default value for K
- for each key K with constraint=forbidden: if spec_meta contains K,
  errors += "Key <K> is forbidden for Deployment: kubectl-style-cli"
- resolved["LANGUAGE"] follows precedence:
    project preset > user preset > system preset > template default

ERRORS:
- ERR_TEMPLATE_MISMATCH if Template-For ≠ "kubectl-style-cli"
- ERR_REQUIRED_MISSING if a required key has no resolved value
- ERR_FORBIDDEN_SET if a forbidden key appears in spec_meta or preset
- ERR_CONSTRAINT_VIOLATION for cross-key invariants

---

## BEHAVIOR/INTERNAL: precedence-resolution
Constraint: required

Defines how conflicting values across layers are resolved for any key.

INPUTS:
```
key:       string
template:  TemplateTable
spec_meta: Map<string, string>
presets:   Map<string, Map<string, string>>  // layer name → values
```

OUTPUTS:
```
value:    string      // effective value (may be empty for absent supported keys)
warning:  string      // set if spec META overrides a default/required key
error:    string      // set if spec META declares a forbidden key
```

STEPS:
1. Start with template defaults as the base map.
2. Merge /usr/share/pcd/presets/ values (vendor defaults); later entries override earlier.
3. Merge /etc/pcd/presets/ values (system admin); overrides vendor defaults.
4. Merge ~/.config/pcd/presets/ values (user); overrides system.
5. Merge <project-dir>/.pcd/ values (project-local); overrides user.
6. For each key in spec META:
   - if constraint=supported → apply; return value.
   - if constraint=required or default → return value, warning = "Spec overrides template default for <K>. Ensure this is intentional."
   - if constraint=forbidden → return empty value, error = "Key <K> is forbidden in kubectl-style-cli specs."

POSTCONDITIONS:
- value reflects the highest-precedence layer that declared a value for key
- warning is non-empty only when spec META overrode a required or default key
- error is non-empty only when spec META declared a forbidden key

ERRORS:
- ERR_SPEC_OVERRIDE_FORBIDDEN if spec META declares a forbidden key

Resolution order (last writer wins):
  1. template default
  2. /usr/share/pcd/presets/    (vendor default)
  3. /etc/pcd/presets/          (system administrator)
  4. ~/.config/pcd/presets/     (user)
  5. <project-dir>/.pcd/        (project-local, committed to git)
  6. spec META explicit override        (only permitted for constraint=supported keys)

---

## TEMPLATE-TABLE

### Versioning and identity

| Key | Value | Constraint | Notes |
|-----|-------|------------|-------|
| VERSION | MAJOR.MINOR.PATCH | required | Semantic versioning. Spec author increments on every meaningful change. |
| SPEC-SCHEMA | MAJOR.MINOR.PATCH | required | Version of the Post-Coding spec schema this file was written against. |
| AUTHOR | name <email> | required | At least one Author: line required. Repeating key; multiple authors permitted. |
| LICENSE | SPDX identifier | required | Must be a valid SPDX license identifier or compound expression. |

### Language and binary

| Key | Value | Constraint | Notes |
|-----|-------|------------|-------|
| LANGUAGE | Go | default | Default target language. The Kubernetes ecosystem is Go-native. |
| LANGUAGE-ALTERNATIVES | Rust | supported | Valid alternative for memory-safety-critical CLIs. |
| BINARY-TYPE | static | default | Single static binary. Portability across Linux distributions without runtime deps. |
| BINARY-TYPE | dynamic | forbidden | Forbidden: defeats the portability promise that makes K8s CLIs shippable. |
| BINARY-COUNT | 1 | required | Exactly one binary with a subcommand tree. Multi-binary tools require separate specs. |
| RUNTIME-DEPS | none | required | No runtime dependencies permitted. All dependencies linked statically. |

### CLI argument structure — the kubectl convention

| Key | Value | Constraint | Notes |
|-----|-------|------------|-------|
| CLI-ARG-STYLE | subcommand-flag | required | Hierarchical subcommands with POSIX-style flags. `<bin> verb [resource] [--flag value]`. The convention established by kubectl and followed by helm, flux, argocd, fleet, tilt. |
| CLI-ARG-STYLE | key=value | forbidden | Forbidden: not the Kubernetes ecosystem convention. If your tool fits key=value style, use `cli-tool` instead. |
| CLI-SUBCOMMAND-DEPTH-MAX | 3 | required | Maximum nesting: `<bin> group subcommand [args]`. E.g. `kubectl config set-context`. Deeper nesting degrades discoverability. |
| CLI-VERB-FIRST | true | required | Top-level subcommand is a verb or a resource-verb ("get", "create", "new", "doctor"). Command reads like a sentence. |
| CLI-FLAG-STYLE | POSIX | required | Long flags `--namespace`, short flags `-n`. Short flags are single-character; long flags are kebab-case. |
| CLI-FLAG-ASSIGNMENT | both | required | Accept both `--flag=value` and `--flag value`. Scripts tend to prefer `--flag=value`; humans `--flag value`. |
| CLI-SHORT-FLAGS-STACK | true | required | Short boolean flags stack: `-xvf` = `-x -v -f`. Standard POSIX behaviour. |
| CLI-SUBCOMMAND-SEPARATOR | space | required | Subcommands separated by spaces, not colons or dashes. `rda services add`, not `rda:services:add` or `rda-services-add`. |

### Help and discoverability

| Key | Value | Constraint | Notes |
|-----|-------|------------|-------|
| HELP-SYSTEM | built-in | required | `-h` and `--help` available on every command and subcommand level. Root `<bin>` with no args prints help, not error. |
| HELP-EXAMPLES | required | required | Each subcommand's help must include at least one usage example. Discoverability is the primary job of help text in a subcommand CLI. |
| HELP-USAGE-FIRST-LINE | true | required | First line of any help output is `Usage: <bin> ...` on stderr. |
| COMPLETIONS | bash | required | `<bin> completion bash` emits a bash completion script. |
| COMPLETIONS | zsh | required | `<bin> completion zsh` emits a zsh completion script. |
| COMPLETIONS | fish | required | `<bin> completion fish` emits a fish completion script. |
| COMPLETIONS | powershell | supported | Optional. Emit on Windows platforms. |

### Output and streaming

| Key | Value | Constraint | Notes |
|-----|-------|------------|-------|
| OUTPUT-FORMAT-DEFAULT | human | required | Default output is human-readable: tables, indented key:value, or brief messages. Never machine-formatted unless explicitly requested. |
| OUTPUT-FORMAT | json | required | `-o json` / `--output json` produces machine-parseable JSON on stdout. |
| OUTPUT-FORMAT | yaml | required | `-o yaml` / `--output yaml` produces machine-parseable YAML on stdout. K8s-native format. |
| OUTPUT-FORMAT | jsonpath | supported | `-o jsonpath=<expr>` applies a JSONPath expression to the JSON output. Used by scripts. |
| OUTPUT-FORMAT | wide | supported | `-o wide` for tabular output with extra columns. |
| OUTPUT-FORMAT | name | supported | `-o name` returns only the resource name(s). Used by shell pipelines. |
| OUTPUT-QUIET-FLAG | -q | required | `-q` / `--quiet` suppresses all non-essential output. Exit code remains the signal of success/failure. |
| OUTPUT-VERBOSE-FLAG | -v | required | `-v` / `--verbose` enables debug-level logging to stderr. Usable as counted flag (`-vvv`) for higher verbosity levels. |
| STREAM-OUTPUT | stdout | required | Normal output on stdout. Machine-readable formats always on stdout. |
| STREAM-DIAGNOSTICS | stderr | required | Errors, warnings, progress indicators, interactive prompts all on stderr. |
| STREAM-TTY-DETECTION | required | required | Detect whether stdout is a TTY. Disable color codes, progress bars, and interactive prompts when redirected to a pipe or file. |

### Exit codes

| Key | Value | Constraint | Notes |
|-----|-------|------------|-------|
| EXIT-CODE-OK | 0 | required | Success. |
| EXIT-CODE-ERROR | 1 | required | Logical error: operation failed, precondition violated, resource not found. |
| EXIT-CODE-INVOCATION | 2 | required | Invocation error: unknown flag, missing required arg, malformed input. |
| EXIT-CODE-RESERVED | 3-125 | supported | Tool-specific additional exit codes in this range. Document in man page. |

### Signals

| Key | Value | Constraint | Notes |
|-----|-------|------------|-------|
| SIGNAL-HANDLING | SIGTERM | required | Clean exit on SIGTERM. Roll back partial work when feasible; never leave filesystem or remote state in a half-written form. |
| SIGNAL-HANDLING | SIGINT | required | Clean exit on SIGINT (Ctrl-C). Same guarantees as SIGTERM. |
| SIGNAL-HANDLING | SIGPIPE | required | Ignore SIGPIPE when streaming to stdout (e.g. piped into `head`); do not panic. |

### Configuration

| Key | Value | Constraint | Notes |
|-----|-------|------------|-------|
| CONFIG-FILES | xdg | required | Config lives in XDG-compliant locations: `$XDG_CONFIG_HOME/<name>/config.yaml` (fallback `~/.config/<name>/config.yaml`). System default at `/etc/<name>/config.yaml`. |
| CONFIG-FILES-OVERRIDE-FLAG | --config | required | `--config <path>` flag overrides the default config file location. |
| CONFIG-FILES-OVERRIDE-ENV | <NAME>_CONFIG | required | `<NAME>_CONFIG=<path>` env var overrides the default config file location. Flag takes precedence over env var. |
| CONFIG-FORMAT | yaml | required | Config files are YAML. K8s ecosystem convention. |
| CONFIG-ENV-VARS | supported | supported | Behaviour may be overridden by env vars prefixed with `<NAME>_`. Each documented in the man page. |
| CONFIG-ENV-VARS-OVERRIDE | required | required | Precedence: CLI flag > env var > config file > built-in default. |

### Kubernetes integration (where applicable)

| Key | Value | Constraint | Notes |
|-----|-------|------------|-------|
| KUBECONFIG-INTEGRATION | supported | supported | If the tool talks to Kubernetes, honor the `KUBECONFIG` env var and the standard `~/.kube/config` fallback. |
| K8S-CONTEXT-FLAG | --context | supported | `--context <name>` flag. Respect the kubectl `--context` convention. |
| K8S-NAMESPACE-FLAG | --namespace / -n | supported | Standard K8s flag pair. |
| K8S-ALL-NAMESPACES-FLAG | --all-namespaces / -A | supported | Scope any list/get operation to all namespaces. Standard K8s convention. |

### Network and authentication

| Key | Value | Constraint | Notes |
|-----|-------|------------|-------|
| NETWORK-CALLS | supported | supported | Network calls are expected: API servers, registries, IdPs, git remotes. Each endpoint documented in the spec. |
| TLS-VERIFICATION | strict | required | TLS certificates validated against the system trust store by default. |
| TLS-SKIP-VERIFY-FLAG | --insecure-skip-tls-verify | forbidden | Forbidden by default. Tools that genuinely need this flag must declare it `supported` explicitly and document the security impact. |
| AUTH-STORAGE | os-keyring | required | Secrets (tokens, passwords, SSH keys) stored in the OS keyring (Keychain on macOS, Secret Service on Linux, Credential Manager on Windows). |
| AUTH-STORAGE | xdg-file | supported | Non-secret auth material (config, public identifiers, refresh hints) may live in XDG config paths. Must have mode 0600 if it contains session data. |
| AUTH-STORAGE | plaintext-file | forbidden | Secrets are never stored in plaintext files, regardless of path. |
| AUTH-METHOD | sso-oauth-device | supported | OAuth 2.0 device authorization grant. Preferred for interactive CLIs. |
| AUTH-METHOD | sso-oauth-pkce | supported | OAuth 2.0 authorization code flow with PKCE. Requires a local browser launch. |
| AUTH-METHOD | token-file | supported | Bearer token read from a file, never from a flag (flags leak in shell history). |
| AUTH-METHOD | token-flag | forbidden | Forbidden: flags appear in shell history, process listings, and logs. |

### File system boundaries

| Key | Value | Constraint | Notes |
|-----|-------|------------|-------|
| FILE-MODIFICATION | project-scope | supported | May write files inside a declared project directory (its own scope). |
| FILE-MODIFICATION | config-scope | supported | May read/write its own config under `~/.config/<name>/`. |
| FILE-MODIFICATION | input-files | forbidden | Never modifies its own input files unless the command is explicitly a mutating one. |
| FILE-MODIFICATION | system-paths | forbidden | Never writes outside home and explicit project directories. |
| IDEMPOTENT | per-command | required | Each BEHAVIOR block declares its idempotency explicitly in STEPS. Mutating commands must be idempotent whenever feasible. |

### Packaging and distribution

| Key | Value | Constraint | Notes |
|-----|-------|------------|-------|
| PACKAGE-FORMAT | RPM | required | Linux RPM package. OBS build target. |
| PACKAGE-FORMAT | DEB | required | Linux DEB package. OBS build target. |
| PACKAGE-FORMAT | OCI | supported | OCI container image. Required for CI pipeline integration scenarios. |
| PACKAGE-FORMAT | PKG | supported | macOS installer package. Required if macOS platform is declared. |
| PACKAGE-FORMAT | MSI | supported | Windows installer. Required if Windows platform is declared. |
| PACKAGE-FORMAT | binary | supported | Raw binary distribution for platforms without package manager integration. |

### Installation and supply chain

| Key | Value | Constraint | Notes |
|-----|-------|------------|-------|
| INSTALL-METHOD | OBS | required | Primary distribution channel: SUSE Open Build Service. Signed packages, auditable. |
| INSTALL-METHOD | curl | forbidden | `curl \| sh` installers are forbidden. Supply chain security requirement. |
| INSTALL-METHOD | homebrew | supported | Homebrew formula for macOS/Linux. Documented in README if supported. |
| INSTALL-METHOD | krew | supported | kubectl plugin manager. Applicable if the tool is a kubectl plugin (name starts with `kubectl-`). |
| INSTALL-METHOD | helm-plugin | supported | Helm plugin directory install. Applicable if the tool is a Helm plugin. |

### Base images (OCI builds only)

| Key | Value | Constraint | Notes |
|-----|-------|------------|-------|
| OCI-BASE-IMAGE | registry.suse.com/bci/bci-micro:latest | default | Default base for static binary runtime. Minimal footprint. |
| OCI-BASE-IMAGE-BUILDER | registry.suse.com/bci/golang:latest | default | Build stage base for Go. Contains Go toolchain on BCI. |
| OCI-BASE-IMAGE-FROM | suse.com | required | Base images must come from `registry.suse.com/*` (BCI) or `dp.apps.rancher.io/*` (AppCo). |
| OCI-BASE-IMAGE-UNQUALIFIED | docker.io | forbidden | Unqualified Docker Hub names and unpinned tags are forbidden. Supply chain security requirement. |

### Platform coverage

| Key | Value | Constraint | Notes |
|-----|-------|------------|-------|
| PLATFORM | Linux | required | Primary target. |
| PLATFORM | macOS | supported | Optional. If declared, PKG or binary output format required. |
| PLATFORM | Windows | supported | Optional. If declared, MSI or binary output format required. |

### Documentation

| Key | Value | Constraint | Notes |
|-----|-------|------------|-------|
| MAN-PAGES | section-1-root | required | Root `<name>(1)` man page covering the tool overview, common options, exit codes, signals. |
| MAN-PAGES | section-1-subcommands | required | One man page per top-level subcommand: `<name>-<verb>(1)`. E.g. `rda-new(1)`, `rda-doctor(1)`. |
| MAN-PAGES | section-1-deep-subcommands | supported | Optional man pages for nested subcommands: `<name>-<group>-<verb>(1)`. Recommended when nesting is used heavily. |
| DOC-README | required | required | README.md documenting installation via OBS, quick start, link to man pages. |

---

## PRECONDITIONS

- This template is applied only when spec META declares Deployment: kubectl-style-cli
- Preset files must be valid TOML or YAML
- If PLATFORM includes macOS, PACKAGE-FORMAT must include PKG or binary
- If PLATFORM includes Windows, PACKAGE-FORMAT must include MSI or binary
- LANGUAGE value in resolved output must be one of: Go, Rust
- If BINARY-TYPE is dynamic, resolution rejects (BINARY-TYPE=dynamic is forbidden here)
- At least one top-level subcommand must be declared in the spec's BEHAVIOR sections
- If CLI-SUBCOMMAND-DEPTH-MAX is exceeded by any declared BEHAVIOR, resolution rejects

---

## POSTCONDITIONS

- Every spec using Deployment: kubectl-style-cli is governed by this template
- A spec may not declare LANGUAGE directly in META unless using Deployment: manual
- Resolved LANGUAGE is always Go or Rust
- curl is never an accepted install method, regardless of preset override
- Unqualified Docker Hub images are never accepted as base image, regardless of preset override
- Plaintext credential storage is never accepted, regardless of preset override
- Every subcommand has an associated man page deliverable

---

## INVARIANTS

- [observable]  constraint=forbidden rows cannot be overridden at any preset layer
- [observable]  constraint=required rows must resolve to a value; missing value is an error
- [observable]  LANGUAGE resolution always produces exactly one value
- [observable]  every subcommand declared in the spec has a corresponding `<name>-<verb>(1)` man page in deliverables
- [observable]  `<bin> <verb> --help` exits 0 and prints usage on every subcommand without additional args requirement
- [observable]  `<bin> --help` (root) exits 0 and prints the command tree
- [observable]  `<bin> completion bash`, `<bin> completion zsh`, `<bin> completion fish` each produce a valid completion script on stdout and exit 0
- [observable]  `<bin> -o json` on any read command produces valid JSON on stdout
- [observable]  TTY detection disables color codes when stdout is not a TTY
- [observable]  a spec declaring Deployment: kubectl-style-cli inherits all required constraints whether or not the spec author is aware of them
- [observable]  template version is recorded in every audit bundle that references it
- [observable]  every generated artifact embeds the SHA256 of the spec file it was produced from
- [implementation]  secrets stored in the OS keyring are never written to filesystem paths during normal operation
- [implementation]  TLS verification is never disabled by default; --insecure-skip-tls-verify is opt-in per-spec and per-invocation

---

## EXAMPLES

EXAMPLE: minimal_spec_resolution
GIVEN:
  spec META contains:
    Deployment: kubectl-style-cli
    Verification: none
    Safety-Level: QM
  no preset files present (template defaults only)
WHEN:
  resolved = resolve(template, spec_meta, preset={})
THEN:
  resolved["LANGUAGE"] = "Go"
  resolved["BINARY-TYPE"] = "static"
  resolved["CLI-ARG-STYLE"] = "subcommand-flag"
  resolved["CLI-FLAG-STYLE"] = "POSIX"
  resolved["OUTPUT-FORMAT-DEFAULT"] = "human"
  resolved["EXIT-CODE-OK"] = "0"
  resolved["INSTALL-METHOD"] = "OBS"
  errors = []
  warnings = []

EXAMPLE: org_preset_overrides_language_to_rust
GIVEN:
  spec META contains:
    Deployment: kubectl-style-cli
    Verification: none
    Safety-Level: QM
  /etc/pcd/presets/org.toml contains:
    [templates.kubectl-style-cli]
    default_language = "Rust"
WHEN:
  resolved = resolve(template, spec_meta, preset={LANGUAGE: "Rust"})
THEN:
  resolved["LANGUAGE"] = "Rust"
  errors = []
  warnings = []

EXAMPLE: forbidden_key_value_style_rejected
GIVEN:
  spec META contains:
    Deployment: kubectl-style-cli
    CLI-ARG-STYLE: key=value
WHEN:
  resolved = resolve(template, spec_meta, preset={})
THEN:
  errors contains:
    "Key CLI-ARG-STYLE=key=value is forbidden for Deployment: kubectl-style-cli. Use Deployment: cli-tool for key=value CLIs."
  resolved is not produced (errors non-empty → reject)

EXAMPLE: forbidden_curl_install_rejected
GIVEN:
  spec META contains:
    Deployment: kubectl-style-cli
    INSTALL-METHOD: curl
WHEN:
  resolved = resolve(template, spec_meta, preset={})
THEN:
  errors contains:
    "Key INSTALL-METHOD=curl is forbidden for Deployment: kubectl-style-cli. Supply chain security requirement."
  resolved is not produced

EXAMPLE: forbidden_tls_skip_rejected
GIVEN:
  spec META contains:
    Deployment: kubectl-style-cli
    TLS-SKIP-VERIFY-FLAG: --insecure-skip-tls-verify
WHEN:
  resolved = resolve(template, spec_meta, preset={})
THEN:
  errors contains:
    "Key TLS-SKIP-VERIFY-FLAG=--insecure-skip-tls-verify is forbidden by default. If your tool genuinely requires this capability, declare it explicitly in spec META with a rationale."
  resolved is not produced

EXAMPLE: macos_platform_requires_pkg_or_binary
GIVEN:
  spec META contains:
    Deployment: kubectl-style-cli
    Verification: none
    Safety-Level: QM
  preset declares PLATFORM includes macOS
  preset does not declare PACKAGE-FORMAT = PKG or binary
WHEN:
  resolved = resolve(template, spec_meta, preset={PLATFORM: "macOS"})
THEN:
  errors contains:
    "PLATFORM macOS requires PACKAGE-FORMAT: PKG or PACKAGE-FORMAT: binary"
  resolved is not produced

EXAMPLE: unqualified_base_image_rejected
GIVEN:
  OCI build configured with:
    OCI-BASE-IMAGE: docker.io/library/alpine:latest
WHEN:
  translator processes spec
THEN:
  errors contains:
    "OCI-BASE-IMAGE must originate from registry.suse.com/* (BCI) or dp.apps.rancher.io/* (AppCo). Unqualified Docker Hub names are forbidden."

EXAMPLE: subcommand_depth_exceeds_max
GIVEN:
  spec declares a BEHAVIOR: `<bin> config context credentials rotate key`
  (four levels of subcommand nesting beyond the root binary)
WHEN:
  resolved = resolve(template, spec_meta, preset={})
THEN:
  errors contains:
    "CLI-SUBCOMMAND-DEPTH-MAX=3 exceeded by subcommand 'config context credentials rotate key' (depth 4). Flatten the command hierarchy or split into multiple top-level commands."
  resolved is not produced

---

## DELIVERABLES

Defines the files a translator must produce for each PACKAGE-FORMAT
declared as `required` or `supported` in the TEMPLATE-TABLE. A translator
must produce all deliverables for every `required` PACKAGE-FORMAT. For
`supported` PACKAGE-FORMATs, deliverables are produced only if that
format is active in the resolved preset.

The prompt to the translator must not enumerate these files —
the translator derives them from this section.

### Delivery Order

Deliverables must be produced in the following order:
1. Core implementation files (source, go.mod / Cargo.toml, Makefile, README.md, LICENSE)
2. Help text and completion scripts (validated by ad-hoc test)
3. Man pages (root + per-subcommand)
4. Required packaging artifacts (RPM, DEB) in table order
5. Supported packaging artifacts if preset active (OCI, PKG, MSI, binary)
6. TRANSLATION_REPORT.md last, after all other files are written and verified

### Deliverables Table

| PACKAGE-FORMAT | Constraint | Required Deliverable Files | Notes |
|---|---|---|---|
| source | required | `main.go` or `cmd/<n>/main.go`, plus `internal/` packages per subcommand. `go.mod`. For Rust: `src/main.rs`, `Cargo.toml`. | Subcommand-per-package layout strongly recommended. Translator documents chosen structure in translation report. |
| build | required | `Makefile` | Must include: `build`, `test`, `install`, `clean`, `man`, `completions` targets. `build` must set `CGO_ENABLED=0` for Go. `man` target: pandoc-based generation for each `.1.md` source. `completions` target: invokes the binary to produce bash, zsh, fish completion scripts. |
| docs | required | `README.md` | Must document: installation via OBS (zypper, apt, dnf), installation via Homebrew if supported, quick start, link to man pages, subcommand tree overview. Must not document curl-based installation. |
| man-root | required | `<n>.1.md`, `<n>.1` | Root man page. Markdown source converted to troff via `pandoc`. |
| man-subcommand | required | `<n>-<verb>.1.md`, `<n>-<verb>.1` | One man page per top-level subcommand declared in the spec's BEHAVIOR blocks. |
| completions | required | Generated at runtime by the binary itself | `<bin> completion bash` etc. No file deliverable; verified at build time by running the binary and asserting non-empty output + exit 0. |
| license | required | `LICENSE` | SPDX identifier from spec META + authoritative URL to the full license text. Never reproduce the full license text. |
| RPM | required | `<n>.spec` | OBS RPM spec file. Name, Version, License (SPDX), Summary, BuildRequires (must include pandoc), %build, %install, %files sections. %files must include man pages and bash completion at `/etc/bash_completion.d/<n>`. |
| DEB | required | `debian/control`, `debian/changelog`, `debian/rules`, `debian/copyright` | Standard Debian source package layout. `debian/copyright` must use DEP-5 machine-readable format with SPDX license identifier. `Build-Depends` must include pandoc. |
| OCI | supported | `Containerfile` | OCI-compliant container build file. Named `Containerfile`. Multi-stage build required. Builder stage: `FROM registry.suse.com/bci/golang:latest AS builder` for Go. Final stage: `FROM registry.suse.com/bci/bci-micro:latest` (not scratch — CLI tools may need a shell for debugging). Must not expose ports. |
| PKG | supported | `<n>.pkgbuild` | macOS installer package descriptor. Required when PLATFORM includes macOS and binary not chosen. |
| MSI | supported | `<n>.wxs` | WiX Toolset source for Windows installer. Required when PLATFORM includes Windows and binary not chosen. |
| binary | supported | none | Raw binary only. No packaging descriptor required. |
| report | required | `TRANSLATION_REPORT.md` | AI translator self-evaluation. Must include: language resolution rationale, delivery mode, template constraints compliance table, subcommand tree produced, ambiguities, deviations, per-example confidence levels, help system verification, completion script verification, man page enumeration. Written last after all other files verified on disk. |
| spec-hash | required | embedded in all artifacts | SHA256 of the spec file embedded as in cli-tool template: source file header comments, `TRANSLATION_REPORT.md` `Spec-SHA256:` field, binary `--version` / `version` subcommand output, RPM `.spec` comment, DEB `control` `X-PCD-Spec-SHA256:` field, `Containerfile` `LABEL pcd.spec.sha256=`, `Makefile` `SPEC_SHA256` variable. Computed once before any output is written. |

### Naming Convention

`<n>` in the above table refers to the component name as declared in the
specification title (first `#` heading). It must be:
- lowercase
- hyphen-separated (no underscores)
- no version suffix in the filename itself

---

## DEPLOYMENT

Runtime: this file is a template specification, not executable code.
It is read by pcd-lint (for template resolution validation) and by
AI translators (for code generation context).

Location in preset hierarchy:
  /usr/share/pcd/templates/kubectl-style-cli.template.md

Versioning:
  Template version is declared in META (Version: field).
  Specs reference the template by name (Deployment: kubectl-style-cli).
  Audit bundles record the template version used at generation time.
  Breaking changes to a template increment the minor version.
  Additions of supported rows are non-breaking.
  Changes to required or forbidden rows are breaking.
  Current version: 0.1.0

---

## EXECUTION

The translator must read this section before generating any code.
It specifies the exact delivery phases, resume logic, compile gate,
help/completion verification, and man page expansion for kubectl-style-cli
components. Follow it exactly.

### Input files

The translator receives in the working directory:
- `kubectl-style-cli.template.md` — this deployment template
- `<spec-name>.md` — the component specification

If the spec's DEPENDENCIES section references hints files, they are also
present. Read them before writing `go.mod`/`Cargo.toml` or any code that
uses those libraries.

### Resume logic

Before writing any file, list the output directory.
If a listed deliverable already exists and is non-empty, skip it — treat
it as complete and move to the next missing file. Report which files were
found and which are being produced.

### Delivery phases

Produce files in this exact order. Complete each phase before starting
the next. Do not produce `TRANSLATION_REPORT.md` until Phase 6 is done.

**Phase 1 — Core implementation**
- Source files implementing the subcommand tree. Recommended layout for Go:
  - `main.go` — entry point, root command definition
  - `cmd/<subcommand>/` — one package per top-level subcommand
  - `internal/` — shared helpers, not exposed
- `go.mod` (Go) or `Cargo.toml` (Rust) — direct dependencies only
- Each subcommand must implement: INPUTS parsing, PRECONDITIONS validation,
  STEPS execution, ERRORS mapping to non-zero exit codes

**Phase 2 — Help, completions, shared infrastructure**
- Root help text validated: `<bin>` with no args, `<bin> --help`, `<bin> -h` all exit 0 and print the command tree
- Completion commands implemented: `<bin> completion bash`, `<bin> completion zsh`, `<bin> completion fish`
- Each subcommand help text includes at least one usage example

**Phase 3 — Man pages**
- `<n>.1.md` — root man page
- For every top-level subcommand declared in a BEHAVIOR block: `<n>-<verb>.1.md`
- All `.1.md` files converted to troff `.1` files via `pandoc` in the Makefile

**Phase 4 — Build and packaging**
- `Makefile` with all targets
- `<n>.spec` (RPM spec)
- `debian/control`, `debian/changelog`, `debian/rules`, `debian/copyright`
- `Containerfile` (if OCI is active in preset)
- `<n>.pkgbuild` (if PKG is active)
- `<n>.wxs` (if MSI is active)
- `LICENSE`

**Phase 5 — Test infrastructure**
- `independent_tests/INDEPENDENT_TESTS.go` (or equivalent for Rust)
- `translation_report/translation-workflow.pikchr`

**Phase 6 — Documentation**
- `README.md`

**Phase 7 — Compile gate and binary verification** (see below)

**Phase 8 — Report (last)**
- `TRANSLATION_REPORT.md`

### Compile gate

Execute after Phase 6 and before Phase 8. If your environment cannot
execute shell commands, document this explicitly under the heading
"Phase 7 — Compile gate not executed" in TRANSLATION_REPORT.md and
state why. Do not silently omit this phase.

**Step 1 — Dependency resolution**

For Go: `go mod tidy`
For Rust: `cargo fetch`

**Step 2 — Compilation**

For Go: `go build ./...`
For Rust: `cargo build --release`

If compilation fails, fix only the identified errors and re-run.
Repeat until compilation succeeds or all reasonable fixes are exhausted.

**Step 3 — Binary verification**

With the compiled binary available:
- `./<n>` (no args) exits 0 and prints help
- `./<n> --help` exits 0 and prints help
- `./<n> -h` exits 0 and prints help
- `./<n> version` exits 0 and prints the version including `spec:<hash>`
- `./<n> completion bash` exits 0 and produces non-empty output
- `./<n> completion zsh` exits 0 and produces non-empty output
- `./<n> completion fish` exits 0 and produces non-empty output
- For every top-level subcommand `<verb>` declared in the spec:
  `./<n> <verb> --help` exits 0 and prints help for that subcommand

Record pass/fail for each verification in TRANSLATION_REPORT.md.

**Step 4 — Record result**

Once all steps pass, do not modify any source files further.
Proceed immediately to Phase 8.

### go.mod / Cargo.toml rules

- Declare only direct dependencies (those your code imports directly)
- For Go: do NOT hand-write indirect dependencies (resolved by `go mod tidy`)
- Do NOT fabricate pseudo-versions, commit hashes, or crate versions
- If hints files are present: use the verified versions they provide
- If no hints file: flag the dependency in TRANSLATION_REPORT.md as
  requiring manual version verification before building

### Recommended CLI framework

For Go: `github.com/spf13/cobra` for command structure, `github.com/spf13/viper`
for configuration layering (XDG + env + flag). These are the de-facto standards
across the Kubernetes ecosystem and will produce idiomatic code that integrates
with existing kubectl-style tooling conventions.

For Rust: `clap` (v4+) for command structure, `config` crate for layered config.

The translator may choose differently if a hints file justifies it, but must
document the deviation in TRANSLATION_REPORT.md.
