# Deployment Templates

A **deployment template** declares everything a PCD translator needs to know
that is *not* in the specification: the target programming language, the
required deliverables, the packaging artefacts, the build and verification
recipe, and the conventions that the generated implementation must follow.

The specification answers *what* the software does. The template answers
*how it ships*. The same specification produces a Go binary today and a
Rust binary in 2045 by changing only the template, not the spec.

---

## Selecting a Template

| Template | Default language | Alternatives | Use when... |
|---|---|---|---|
| [`cli-tool`](cli-tool.template.md) | Go | Rust, C, C++, C# | A single-binary command-line tool with `--flag` style arguments |
| [`kubectl-style-cli`](kubectl-style-cli.template.md) | Go | Rust | A multi-verb CLI in the `kubectl` or `git` style: `<bin> verb resource --flag` with shell completions |
| [`mcp-server`](mcp-server.template.md) | Go | Python, Rust | A Model Context Protocol server with stdio and/or streamable-HTTP transports |
| [`backend-service`](backend-service.template.md) | Go | Rust | A 12-factor application running as a daemon under systemd or in a container |
| [`cloud-native`](cloud-native.template.md) | Go | — | A Kubernetes operator or controller; produces CRDs, RBAC manifests, and Helm chart |
| [`gui-tool`](gui-tool.template.md) | OS-dependent | Qt6, Tauri, Flutter | A desktop GUI application; default language depends on host platform |
| [`cockpit-module`](cockpit-module.template.md) | HTML + JS + CSS | — | A plugin for the Cockpit web administration interface (`cockpit-project.org`) |
| [`python-tool`](python-tool.template.md) | Python | — | A Python tool, automation script, or data pipeline. QM safety level only — not for safety-critical components |
| [`library-c-abi`](library-c-abi.template.md) | C | Rust (via `cbindgen`) | A general-purpose C-ABI shared library with stable ABI |
| [`verified-library`](verified-library.template.md) | C | Rust | A safety- or security-critical C-ABI library requiring formal verification (ASIL-B/C/D, DAL-A/B, EAL4+/EUCC) |
| [`spack-package`](spack-package.template.md) | Python (Spack DSL) | — | A Spack package recipe (`package.py`) for HPC and scientific software distribution |
| [`project-manifest`](project-manifest.template.md) | N/A | — | An architect artefact that defines a multi-component system and produces a project-level audit bundle. No code is generated. |

If your component does not match any of these, use the spec's
`Deployment: manual` mode and declare every field explicitly. This is the
documented fallback — not a workaround.

---

## What a Template Controls

A template fixes the answers to questions that should not vary per component:

**Target language and valid alternatives.** Every template declares one
default language. Most permit a small set of alternatives (Go cli-tools can
become Rust cli-tools, but not Java ones). Some templates permit none —
`cloud-native` is Go-only because the Kubernetes ecosystem effectively
mandates Go, and `python-tool` is Python-only because the template name
*is* the language constraint.

**Required deliverables.** Every template lists the files that must be
produced beyond source code: RPM `.spec` files, Debian packaging metadata,
man pages (section 1 for tools, section 3 for libraries), `Containerfile`
where applicable, `README.md`, `LICENSE`, and the `TRANSLATION_REPORT.md`
audit record. A translation that omits any required deliverable is
incomplete by definition.

**Build and verification recipe.** Each template's `EXECUTION` section
gives the translator an ordered phase list: when to generate the source
files, when to run the compile gate, what commands to run, and what
"success" means. For `cli-tool` this includes a `version` subcommand that
must print the spec hash; for `cloud-native` it includes Helm chart
validation; for `python-tool` it includes `mypy --strict` and `pytest`.

**Constraints that translators must respect.** Some are positive
(`pandoc` is a required build dependency for any template that ships man
pages); some are negative (`curl` is never an accepted install method;
`key=value` CLI argument style is forbidden for `python-tool`). The
template enforces these regardless of which LLM does the translation.

---

## The Preset Hierarchy

Templates declare defaults. Organisations, users, and projects can override
those defaults without modifying any template. The resolution order is:

1. Template default (this directory)
2. System preset — `/etc/pcd/presets/`
3. User preset — `~/.config/pcd/presets/`
4. Project preset — `.pcd/presets/`

Later entries override earlier ones. An organisation that standardises on
Rust for CLI tools ships a system preset that sets
`default_language = "rust"` for `cli-tool`. A single project that needs C++
overrides that with a project preset. No template file is touched, no
specification is rewritten — the change applies only where intended, and
nowhere else.

This is the same layering model that `systemd` uses for unit presets, and
for the same reason: it lets a thousand projects diverge cleanly from a
single set of upstream defaults.

---

## Adding a New Template

A new template is itself a PCD specification with `Deployment: template`
in its META section. The reference templates in this directory are working
examples — `cli-tool.template.md` is a good starting point for any
language-resolving template; `cockpit-module.template.md` is a good
starting point for any template that does not target a compiled language.

See [`CONTRIBUTING.md`](../CONTRIBUTING.md) for the process. A new template
should be added when an existing one cannot be made to fit by adjusting
presets — not as a way to capture a single-project deviation.

---

## Notes

The `gui-tool` template chooses its default language based on the host
operating system, because there is no universally correct answer: GTK on
Linux argues for C, the macOS frameworks argue for Swift, and Windows
argues for C#. Where this is unacceptable, declare a preset that pins the
language explicitly.

The `verified-library` template does not permit `Safety-Level: QM`. If
your library does not require ASIL or equivalent certification, use
`library-c-abi` instead. The two are deliberately separate templates so
the constraint cannot be relaxed by accident.

The `spack-package` template has no `EXECUTION` phase that compiles
generated code: the deliverable is a declarative Python class that Spack
itself executes. Validation happens via `spack audit`, which the template
maps to PCD `INVARIANTS`.
