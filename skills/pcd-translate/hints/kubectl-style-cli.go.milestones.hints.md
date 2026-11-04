# kubectl-style-cli.go.milestones.hints

Scaffold patterns for `kubectl-style-cli` deployment template, Go target.
Read this file before the scaffold milestone (`Scaffold: true`) so the
package layout established in pass 1 is stable across all subsequent
milestone passes.

These hints were extracted from the first end-to-end production translation
of a non-self-specified kubectl-style-cli component (the `rda` CLI for
SUSE Rancher Developer Access, idefxH/rda-cli).

---

## Package layout (scaffold pass)

```
code/
├── main.go                         # 12 lines, calls cmd.Execute()
├── go.mod                          # module github.com/<owner>/<binary>
├── Makefile                        # build/test/install/clean/man/completions
├── README.md                       # OBS install + quick start
├── LICENSE
├── Containerfile                   # multi-stage, BCI-only bases
├── <binary>.spec                   # OBS RPM spec
├── debian/                         # control, changelog, rules, copyright
├── man/                            # one .1.md per top-level subcommand
├── cmd/                            # one .go per top-level subcommand
│   ├── root.go                     # cobra root command, flag wiring
│   ├── <verb>.go                   # one file per top-level verb
│   └── <group>_<verb>.go           # one file per nested verb
├── internal/                       # one package per spec INTERFACE + concern
│   ├── types/                      # shared types from spec TYPES section
│   ├── errs/                       # named error sentinels
│   ├── exitcode/                   # 0/1/2 mapping
│   ├── output/                     # human/json/yaml renderers
│   ├── version/                    # binary version + spec hash + template
│   ├── <interface-name>/           # one package per spec INTERFACE
│   │   ├── interface.go            # the interface declaration
│   │   ├── production.go           # the production implementation
│   │   └── fake.go                 # the test-double implementation
│   ├── <behavior-internal>/        # one package per BEHAVIOR/INTERNAL
│   │   └── <name>.go
│   └── projectfile/                # if the spec mentions a project metadata file
└── independent_tests/              # one file per milestone
    └── milestone_NNN_test.go
```

Why this shape:

- One file per top-level subcommand under `cmd/` keeps each command's
  flag surface and execution body next to its tests.
- Nested subcommands use snake-cased filenames to read naturally
  (`templates_list.go` for `<binary> templates list`).
- One package per INTERFACE under `internal/` so production and test
  double can share unexported helpers without exposing them.
- `internal/types` holds the spec's TYPES section verbatim. Other
  packages import these types rather than redeclaring them.
- `internal/errs` centralises every `ERR_*` sentinel from the spec's
  ERRORS sections so error identity is comparable via `errors.Is`.
- `independent_tests/` lives at the code root (not under `internal/`)
  so tests use only the public-via-internal-friend interfaces and
  declared test doubles. New tests file per milestone keeps the
  per-pass blast radius easy to read in git diffs.

---

## Stub contract (scaffold pass)

Every BEHAVIOR's stub function must:

1. Compile.
2. Return the **typed zero value** for its declared output type.
3. For output types that serialise to a JSON object, return an
   **initialised empty value** — never a nil pointer or nil map.
   - `(SomeStruct{}, ErrNotImplemented)` — not `(nil, ...)`.
   - `(make([]SomeType, 0), ErrNotImplemented)` — not `(nil, ...)`.
   - `(map[string]string{}, ErrNotImplemented)` — not `(nil, ...)`.
4. Return `errs.ErrNotImplemented` so subsequent milestones can replace
   the stub by editing only that function body.

For BEHAVIORs whose acceptance criteria already exercise the function
in the scaffold milestone (typically `<binary> --version`, `<binary>
--help`, `<binary> completion <shell>`), implement them fully in the
scaffold pass — they are part of the compile gate's observable contract.

---

## CLI wiring (cobra)

Use `github.com/spf13/cobra` for the command tree.

`cmd/root.go` declares the root command with persistent flags shared by
all subcommands:

- `--config <path>` — overrides extension/XDG config discovery
- `--output / -o` — default human, accepts json | yaml (and whatever
  formats the spec declares as `OUTPUT-FORMAT`)
- `--verbose / -v` — counted flag, enables debug logging to stderr
- `--quiet / -q` — suppresses non-essential output

`cmd/root.go.Execute()` returns the integer exit code (0/1/2). `main.go`
does `os.Exit(cmd.Execute())`. This keeps `main.go` minimal and tests
of the command tree directly callable from Go test functions.

Each subcommand file declares its `cobra.Command{}` and registers itself
to the root or to its parent group via `init()` or an explicit factory
function. Either pattern works; pick one and apply consistently.

Group commands (e.g. `<binary> templates` as a parent of `list` and
`show`) live in their own file (`templates.go`) declaring the parent
command, with `templates_list.go` and `templates_show.go` adding their
children.

---

## Configuration layering

The kubectl-style-cli template requires precedence: flag > env > config
file > built-in default.

In v1 implementations of rda we found that `github.com/spf13/viper` was
heavier than necessary for this narrow use case. Two acceptable
implementations:

1. **viper** — bind flags via `viper.BindPFlag`, env via
   `viper.AutomaticEnv()` + prefix, config file via `viper.SetConfigFile`.
   Use this when the component is expected to grow into reload, multi-format,
   or environment-binding heavy use cases.

2. **internal/configresolve** — a small, explicit resolver that reads the
   `--config` flag value, then the `<NAME>_CONFIG` env var, then the
   extension state if any, then the XDG default. Each layer is a function
   call returning `(*Config, error)`. The first non-nil result wins.
   Use this when the component is small, the config schema is fixed,
   and binary size or audit clarity matter.

Both honour the precedence and can be swapped without changing the
calling sites in `cmd/`.

---

## Output rendering

`internal/output` exposes a single function `Render(format, value, w io.Writer) error`
that dispatches on format:

- `human` — type-specific table or indented key:value walker
- `json` — `json.NewEncoder(w).Encode(value)` with `SetIndent("", "  ")`
- `yaml` — `yaml.NewEncoder(w).Encode(value)`

Every command that takes `--output` calls this function with the same
value type for all three formats. The human renderer can be type-specific
(per-package), but the structured renderers (json/yaml) must be generic.

For TTY detection (color, progress bars), use `golang.org/x/term.IsTerminal(int(os.Stdout.Fd()))`.

---

## Spec hash embedding

The translator's spec-hash invariant requires every artifact to embed
the SHA256 of the spec file. Concrete locations for kubectl-style-cli + Go:

- Top of every `.go` file:
  `// generated from spec: <specname>.md sha256:<hash>`
- `Makefile`: `SPEC_SHA256 := <hash>` variable, used in `-ldflags`
- `<binary>.spec` (RPM): `# pcd-spec-sha256: <hash>` comment
- `debian/control`: `X-PCD-Spec-SHA256: <hash>` field
- `Containerfile`: `LABEL pcd.spec.sha256="<hash>"`
- Binary version output: `<binary> version` includes `spec:<hash>`
- `TRANSLATION_REPORT.md` header: `Spec-SHA256: <hash>`

Compute the hash once at the start of the translation run via
`sha256sum <specname>.md`. Pass it through the build via a Makefile
variable + `-ldflags="-X internal/version.SpecSHA256=<hash>"` so the
binary surfaces it at runtime without re-reading the spec file.

---

## Test discipline

`independent_tests/` MUST use only the declared test-double
implementations from the spec's INTERFACES section — never the
production implementation, never live external services.

Per-milestone test files (`milestone_NNN_test.go`) are recommended:

- Each milestone's test file adds tests for its Included BEHAVIORs.
- Shared test fixtures (e.g. a real on-disk `file://` git bundle for
  load-opinion-bundle) live in helpers in the latest milestone's file
  and are imported by earlier and later tests.

Acceptance criteria from the spec's MILESTONE section should be
mirrored by named tests where feasible — the per-EXAMPLE confidence
table in the translation report needs to point at named test functions
to claim High confidence.

A 0.0.0 scaffold pass should produce 8–10 tests covering the
fully-implemented stubs (typically `version`, `completion`, and any
input-validation paths that don't depend on external state). Each
subsequent milestone adds 4–10 more tests as BEHAVIORs are promoted
from stubs to real bodies.

---

## Common deviations to flag

When translating a kubectl-style-cli + Go spec, expect to encounter
these recurring decisions where the spec is silent or ambiguous:

1. **`--target-dir` semantics for scaffolding commands.** Specs often
   say "use this dir, else compute from cwd+name". Whether `--target-dir`
   is the *parent* directory or the *project root* is not usually
   declared. Pick **parent semantics** (`<target-dir>/<name>`); it
   matches how `git clone --target-dir` and similar tools behave, and
   is the more natural user-facing meaning.

2. **Case-insensitive substring matching for "the library chart".**
   Specs that reference "the dependency matching the library chart
   convention" rarely declare the convention explicitly. A safe v1
   heuristic is `strings.Contains(strings.ToLower(name), "library")`.
   Document the heuristic in the translation report and flag it for
   formalisation in the next spec revision (typically by adding an
   authoritative name field to the bundle manifest).

3. **YAML round-trip loses comments.** `yaml.v3` on generic maps does
   not preserve comments. If the spec asks for a comment to be
   inserted (e.g. "mark this dependency as optional"), document the
   gap and leave the entry unmodified rather than invent a workaround.
   Future revisions can adopt the `yaml.v3` Node API or a
   comment-preserving parser like `go-yaml-edit`.

4. **OCI registry pulls require auth wiring.** If the spec's HelmClient
   references `oci://` URIs but the auth path is not declared, return
   `ErrNotImplemented` for `oci://` and support `file://` only. Document
   the gap; future revisions add registry credentials sourced from the
   appropriate component (Rancher Desktop extension, environment, etc.).

5. **Stale-cache detection requires persistent state.** Loaders that
   should detect "cached HEAD differs from intended ref" need a state
   file. If the spec doesn't declare one, defer the check; treat
   `refresh=true` as the explicit opt-in. Document the gap.

---

## Translation report content for kubectl-style-cli

Every kubectl-style-cli pass should include in `TRANSLATION_REPORT.md`:

- Pass history table: pass number, milestone, spec-SHA, spec META version,
  date, status. Earlier passes are frozen at the SHA active at the time.
- Per-pass: active milestone, included/deferred BEHAVIORs, STEPS applied
  (per BEHAVIOR, file, notes), interface methods promoted from stubs to
  real implementations.
- Compile gate result: `go mod tidy`, `go build ./...`, `go test
  ./independent_tests/...`, the binary size.
- Acceptance criteria result: each criterion + observed output + pass/fail.
- Per-EXAMPLE confidence update: from previous to new, with named test
  evidence.
- Rules that could not be implemented exactly as written: numbered list
  of every spec text that was deviated from, with rationale and a
  follow-up flag.
- New direct dependencies promoted from indirect.
- Dead code left in place deliberately (with rationale).

This template makes the audit bundle trivially reviewable across
milestones because every pass has the same shape.

---

## Tooling versions known to work (v1)

| Tool | Version | Notes |
|------|---------|-------|
| Go | 1.22+ | go.mod can declare 1.22; 1.23 is the spec recommendation |
| github.com/spf13/cobra | v1.8.0+ | for the command tree |
| github.com/spf13/viper | v1.18.0 (optional) | only if using viper for config |
| helm.sh/helm/v3 | v3.14.0+ | brings ~30 MB of K8s API types — accept this; binary lands ~50 MB |
| github.com/go-git/go-git/v5 | v5.11.0+ | static-link git operations |
| gopkg.in/yaml.v3 | v3.0.1+ | direct dep regardless of viper choice |
| pandoc | 2.x+ | required at build time for man pages |

---

This file is CC-BY-4.0. Update it as the kubectl-style-cli template
itself evolves. Each major scaffold-pattern change should bump a
section heading version comment so downstream translations know which
hints version they are reading.
