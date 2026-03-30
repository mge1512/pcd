# mcp-server-pcd

## META
Deployment:        mcp-server
Version:           0.1.0
Spec-Schema:       0.3.19
Author:            Matthias G. Eckermann <pcd@mailbox.org>
License:           GPL-2.0-only
Verification:      none
Safety-Level:      QM

---

## TYPES

```
TemplateName := string
// Known values: cli-tool, mcp-server, cloud-native, verified-library,
// library-c-abi, python-tool, project-manifest, and any user-installed
// template. Unknown names return TemplateNotFound.

TemplateVersion := string
// Semantic version (MAJOR.MINOR.PATCH) or "latest".
// "latest" resolves to the highest installed version.

HintsKey := string
// Format: "<template>.<language>.<library>"
// Example: "cloud-native.go.go-libvirt"

ResourceURI := string
// Format: "pcd://<type>/<n>"
// Types: templates, prompts, hints
// Examples:
//   pcd://templates/cli-tool
//   pcd://prompts/interview
//   pcd://prompts/translator
//   pcd://hints/cloud-native.go.go-libvirt

Diagnostic := {
  severity:  "error" | "warning"
  line:      integer       // 1-based; 1 for file-level diagnostics
  section:   string        // e.g. "META", "BEHAVIOR", "structure"
  message:   string
  rule:      string        // e.g. "RULE-01", "RULE-08"
}

LintResult := {
  valid:        boolean
  errors:       integer
  warnings:     integer
  diagnostics:  Diagnostic[]
}

TemplateRecord := {
  name:     TemplateName
  version:  TemplateVersion
  language: string         // default target language
  content:  string         // full template Markdown text
}

ResourceRecord := {
  uri:      ResourceURI
  name:     string
  content:  string
}
```

---

## INTERFACES

```
Filesystem {
  // Optional interface — used only when lint_file is called.
  // Not required for lint_content, resource serving, or tool listing.
  required-methods:
    ReadFile(path) -> (content: string, error)
  implementations-required:
    production:  OSFilesystem
    test-double: FakeFilesystem {
      configurable: Files map[string]string, ReadErr map[string]error
    }
}

AssetStore {
  // Provides templates, hints, and prompt content to all MCP tools.
  //
  // ASSET EMBEDDING (build-time):
  //   All assets (templates, hints, prompts) are compiled into the
  //   binary at build time using the language's native asset embedding
  //   mechanism. The binary is self-contained and fully functional
  //   with no pcd-templates package installed at runtime.
  //
  //   The build system stages assets into a location the compiler can
  //   embed from, relative to the store implementation. The exact
  //   mechanism is implementation-defined — see the language hints file.
  //
  //   Staged asset directories are build artefacts only:
  //     - populated by the build system before compilation
  //     - listed in .gitignore
  //     - removed by the clean target
  //     - never committed to the repository
  //     - not required at runtime
  //
  // RUNTIME OVERLAYS (optional, site-local):
  //   After serving embedded assets, the implementation checks
  //   filesystem overlay directories and applies them with last-wins
  //   precedence. All overlay directories are optional — their absence
  //   is not an error.
  //
  //   Overlay search order (ascending precedence):
  //     /usr/share/pcd/templates|hints|prompts/
  //     /etc/pcd/templates|hints|prompts/
  //     ~/.config/pcd/templates|hints|prompts/
  //     ./.pcd/templates|hints|prompts/
  //
  //   A filesystem entry with the same key as an embedded entry
  //   replaces it. New keys in the filesystem are added alongside
  //   embedded entries.

  required-methods:
    ListTemplates()              -> ([]TemplateRecord, error)
    GetTemplate(name, version)   -> (TemplateRecord, error)
    GetHints(key: string)        -> (content: string, error)
    ListHintsKeys()              -> ([]string, error)
    GetPrompt(name: string)      -> (content: string, error)
    ListPrompts()                -> ([]string, error)

  implementations-required:
    production:  EmbeddedLayeredStore {
      // Serves all embedded assets as the base layer.
      // Applies filesystem overlays on top at startup.
      // Returns NotFound only if the key is absent from both
      // embedded assets and all overlay directories.
    }
    test-double: FakeStore {
      configurable:
        Templates []TemplateRecord
        Hints     map[string]string
        Prompts   map[string]string
      // In-memory only. No filesystem access. No embedded assets.
    }
}
```

---

## BEHAVIOR: list_templates
Constraint: required

INPUTS:
  none

PRECONDITIONS:
  - AssetStore is reachable

STEPS:
  1. Call AssetStore.ListTemplates().
     On failure: return MCP error -32603 with store error message.
  2. Return the list as a JSON array of TemplateRecord objects,
     omitting the content field.
     MECHANISM: content is omitted from list results to keep
     response size small; callers use get_template to fetch content.

POSTCONDITIONS:
  - Response contains one entry per installed template.
  - Each entry has name, version, language fields.
  - content field is absent from each entry.

ERRORS:
  - -32603  AssetStore unreachable or read error

---

## BEHAVIOR: get_template
Constraint: required

INPUTS:
  name:     TemplateName
  version:  TemplateVersion   // default: "latest"

PRECONDITIONS:
  - name is a non-empty string
  - version is "latest" or a valid semantic version

STEPS:
  1. Call AssetStore.GetTemplate(name, version).
     On TemplateNotFound: return MCP error -32602,
       message "unknown template: {name}".
     On VersionNotFound: return MCP error -32602,
       message "version {version} not found for template {name}".
     On other error: return MCP error -32603.
  2. Return the full TemplateRecord including content.

POSTCONDITIONS:
  - Response contains name, version, language, and full content.

ERRORS:
  - -32602  Unknown template name or version
  - -32603  Store read error

---

## BEHAVIOR: list_resources
Constraint: required

INPUTS:
  none

PRECONDITIONS:
  - AssetStore is reachable

STEPS:
  1. Call AssetStore.ListTemplates() to enumerate template URIs.
     On failure: record error, continue with empty template list.
  2. Call AssetStore.ListPrompts() to enumerate prompt URIs.
  3. Call AssetStore.ListHintsKeys() to enumerate hints URIs.
  4. Assemble ResourceRecord list with URIs in format:
       pcd://templates/<n>
       pcd://prompts/<n>
       pcd://hints/<key>
  5. Return list.

POSTCONDITIONS:
  - All embedded templates, prompts, and hints are represented.
  - Filesystem overlay entries are included if present.
  - Each entry has uri and name. content is absent.

ERRORS:
  - Partial results returned if AssetStore fails for one asset type.

---

## BEHAVIOR: read_resource
Constraint: required

INPUTS:
  uri:  ResourceURI

PRECONDITIONS:
  - uri is a non-empty string
  - uri matches "pcd://<type>/<n>" where type is one of:
    templates, prompts, hints

STEPS:
  1. Parse uri into type and name components.
     On parse failure: return MCP error -32602,
       message "invalid resource URI: {uri}".
  2. Dispatch by type:
     - "templates": call AssetStore.GetTemplate(name, "latest").
       On TemplateNotFound: return MCP error -32602,
         message "resource not found: {uri}".
     - "prompts": call AssetStore.GetPrompt(name).
       On not found: return MCP error -32602,
         message "resource not found: {uri}".
     - "hints": call AssetStore.GetHints(name).
       On not found: return MCP error -32602,
         message "resource not found: {uri}".
     - unknown type: return MCP error -32602,
         message "unknown resource type: {type}".
  3. Return ResourceRecord with uri, name, and content.

POSTCONDITIONS:
  - Response contains full content of the requested resource.

ERRORS:
  - -32602  URI parse error, unknown type, or resource not found
  - -32603  Store read error

---

## BEHAVIOR: lint_content
Constraint: required

INPUTS:
  content:   string    // full Markdown text of the spec
  filename:  string    // used for diagnostic line references; default "spec.md"

PRECONDITIONS:
  - content is a non-empty string
  - filename has .md extension or is empty (default applied)

STEPS:
  1. If filename is empty, set filename to "spec.md".
  2. If filename does not end in ".md": return MCP error -32602,
       message "filename must have .md extension".
  3. Run all pcd-lint rules (RULE-01 through RULE-14) against content.
     MECHANISM: identical rule set and logic to the pcd-lint CLI;
     no network calls; no filesystem access; pure in-memory validation.
  4. Assemble LintResult with valid, errors, warnings, diagnostics.
  5. Return LintResult as JSON.

POSTCONDITIONS:
  - valid = true iff errors = 0.
  - Every Diagnostic has severity, line, section, message, rule.
  - Line numbers are 1-based relative to content.
  - Result is identical to running: pcd-lint <file> on the same content.

ERRORS:
  - -32602  filename missing .md extension

---

## BEHAVIOR: lint_file
Constraint: required

INPUTS:
  path:    string    // absolute or relative filesystem path to a .md file

PRECONDITIONS:
  - path is a non-empty string
  - Filesystem is accessible

STEPS:
  1. If path does not end in ".md": return MCP error -32602,
       message "file must have .md extension: {path}".
  2. Call Filesystem.ReadFile(path).
     On file-not-found: return MCP error -32602,
       message "cannot open file: {path}".
     On read error: return MCP error -32603,
       message "read error: {path}: {error}".
  3. Run lint_content(content, basename(path)).
  4. Return LintResult.

POSTCONDITIONS:
  - Same as lint_content postconditions.
  - Filesystem is never modified.

ERRORS:
  - -32602  Missing .md extension, file not found
  - -32603  Filesystem read error

---

## BEHAVIOR: get_schema_version
Constraint: required

INPUTS:
  none

STEPS:
  1. Return the Spec-Schema version this server was built against
     as a plain string.
     MECHANISM: value is compiled in as a constant; no runtime lookup.

POSTCONDITIONS:
  - Response is a semantic version string, e.g. "0.3.19".

ERRORS:
  none

---

## BEHAVIOR: http-transport
Constraint: required

INPUTS:
  listen:  ListenAddress    // default: 127.0.0.1:8080

STEPS:
  1. If listen= argument is absent, use 127.0.0.1:8080.
  2. Bind HTTP listener on listen address.
     On bind failure: write error to stderr and exit 1.
  3. Serve MCP Streamable HTTP transport on /mcp endpoint.
  4. On SIGTERM or SIGINT: stop accepting new connections,
     complete in-flight requests, then exit 0.
     MECHANISM: graceful shutdown with 10-second drain timeout.

POSTCONDITIONS:
  - All tools and resources are accessible via HTTP POST to /mcp.
  - Server exits 0 on clean shutdown.

ERRORS:
  - exit 1  bind failure

---

## BEHAVIOR: stdio-transport
Constraint: required

INPUTS:
  none    // transport is selected by bare-word argument "stdio"

STEPS:
  1. Serve MCP stdio transport: read JSON-RPC from stdin,
     write responses to stdout.
  2. Diagnostics and startup messages to stderr only.
  3. On EOF on stdin or SIGTERM/SIGINT: flush pending writes and exit 0.

POSTCONDITIONS:
  - stdout contains only valid MCP JSON-RPC messages.
  - Server exits 0 on clean shutdown.

ERRORS:
  none

---

## PRECONDITIONS

- Invocation provides exactly one transport selector:
    bare word "stdio" → stdio transport
    bare word "http"  → http transport
    (default if neither: stdio)
- If both are given: write error to stderr and exit 2.
- If listen= is given without "http": write warning to stderr, ignore.

---

## POSTCONDITIONS

- Server never modifies any file on disk.
- Server never makes outbound network calls.
- Server never reads environment variables for behaviour control.
- lint_content and lint_file produce identical output to pcd-lint CLI
  for identical input.
- All MCP responses are valid JSON-RPC 2.0.

---

## INVARIANTS

- [observable]      stdio transport: stdout contains only MCP JSON-RPC messages
- [observable]      lint_content result is identical to pcd-lint CLI on same input
- [observable]      server is idempotent: same request always returns same response
- [observable]      server never exits with code other than 0, 1, or 2
- [implementation]  rule execution order: RULE-01 through RULE-14, same as pcd-lint
- [implementation]  resource URIs follow pcd://<type>/<n> scheme exactly
- [implementation]  all assets (templates, hints, prompts) are embedded into the
                    binary at build time using a single unified asset embedding
                    mechanism — no distinction in method between asset types
- [implementation]  staged asset directories used during build are not committed
                    to the repository and not required at runtime
- [observable]      GetTemplate, GetHints, and GetPrompt succeed for all assets
                    shipped with the binary, regardless of runtime filesystem state
- [observable]      the binary is fully functional with no pcd-templates package
                    installed and no overlay directories present on the system

---

## EXAMPLES

EXAMPLE: list_templates_returns_names
GIVEN:
  AssetStore contains cli-tool@0.3.19 and mcp-server@0.3.19
WHEN:
  tool list_templates called with no arguments
THEN:
  response contains two entries
  each entry has name, version, language
  content field is absent from each entry

EXAMPLE: get_template_cli_tool
GIVEN:
  AssetStore contains cli-tool@0.3.19
WHEN:
  tool get_template called with name="cli-tool" version="latest"
THEN:
  response contains name="cli-tool" version="0.3.19"
  response contains full template Markdown in content field

EXAMPLE: get_template_unknown
GIVEN:
  AssetStore does not contain "serverless"
WHEN:
  tool get_template called with name="serverless"
THEN:
  MCP error -32602 returned
  message contains "unknown template: serverless"

EXAMPLE: read_resource_interview_prompt
GIVEN:
  AssetStore contains prompt named "interview"
WHEN:
  tool read_resource called with uri="pcd://prompts/interview"
THEN:
  response contains uri="pcd://prompts/interview"
  response contains full interview prompt Markdown in content field

EXAMPLE: read_resource_invalid_uri
GIVEN:
  any server state
WHEN:
  tool read_resource called with uri="http://example.com/bad"
THEN:
  MCP error -32602 returned
  message contains "invalid resource URI"

EXAMPLE: lint_content_valid_spec
GIVEN:
  content is a valid PCD spec with all required sections
WHEN:
  tool lint_content called with that content and filename="myspec.md"
THEN:
  result.valid = true
  result.errors = 0
  result.diagnostics is empty

EXAMPLE: lint_content_missing_invariants
GIVEN:
  content is a PCD spec missing the INVARIANTS section
WHEN:
  tool lint_content called with that content
THEN:
  result.valid = false
  result.errors >= 1
  one diagnostic has rule="RULE-01" severity="error"
  diagnostic message contains "INVARIANTS"

EXAMPLE: lint_content_bad_extension
GIVEN:
  any valid spec content
WHEN:
  tool lint_content called with filename="myspec.txt"
THEN:
  MCP error -32602 returned
  message contains ".md extension"

EXAMPLE: lint_file_not_found
GIVEN:
  path "/tmp/missing.md" does not exist
WHEN:
  tool lint_file called with path="/tmp/missing.md"
THEN:
  MCP error -32602 returned
  message contains "cannot open file"

EXAMPLE: lint_content_matches_cli
GIVEN:
  content is any PCD spec
  the same content is saved to disk as "test.md"
WHEN:
  tool lint_content called with that content
  AND pcd-lint CLI run on "test.md"
THEN:
  result.valid matches CLI exit code (0 = valid, 1 = invalid)
  result.diagnostics matches CLI diagnostic output
  MECHANISM: [observable] invariant — validated by TestLintMatchesCLI

EXAMPLE: stdio_startup
GIVEN:
  server started with bare word "stdio"
WHEN:
  MCP initialize request sent on stdin
THEN:
  valid MCP initialize response written to stdout
  server capabilities include tools and resources

EXAMPLE: http_startup
GIVEN:
  server started with bare word "http"
WHEN:
  HTTP POST to /mcp with MCP initialize request
THEN:
  valid MCP initialize response returned
  HTTP status 200

EXAMPLE: http_bind_failure
GIVEN:
  port 8080 is already in use
WHEN:
  server started with "http" and listen=127.0.0.1:8080
THEN:
  error message written to stderr
  server exits with code 1

EXAMPLE: standalone_no_pcd_templates
GIVEN:
  pcd-templates package is not installed
  no overlay directories exist on the system
WHEN:
  server started and list_templates called
THEN:
  response contains all templates shipped with the binary
  server does not error or exit

---

## TOOLCHAIN-CONSTRAINTS

```
EMBED-ASSETS:
  // All asset types (templates, hints, prompts) are compiled into
  // the binary using the language's native asset embedding mechanism.
  // No distinction in embedding method between asset types.
  //
  // BUILD-TIME LAYOUT:
  //   The build system must stage all assets into a location the
  //   compiler can embed from, relative to the store implementation.
  //   The exact relative paths are implementation-defined — see the
  //   language hints file.
  //
  //   Source of truth in the repository:
  //     repo-root/templates/    (all *.template.md files)
  //     repo-root/hints/        (all *.hints.md files)
  //     repo-root/prompts/      (all *.md prompt files)
  //
  //   The build system copies or links these to the staging location
  //   before compilation. Staged directories are build artefacts only.

  - type: templates
    source: repo-root/templates/*.template.md
    key-derivation: filename stem before ".template.md"
      // e.g. "cli-tool.template.md" -> key "cli-tool"

  - type: hints
    source: repo-root/hints/*.hints.md
    key-derivation: filename stem before ".hints.md"
      // e.g. "python-tool.hints.md" -> key "python-tool"

  - type: prompts
    source: repo-root/prompts/*.md
    key-derivation: filename stem before ".md"
      // e.g. "interview-prompt.md" -> key "interview"
      // e.g. "translator.md"       -> key "translator"
```

---

## DEPENDENCIES

- name: github.com/mark3labs/mcp-go
  purpose: MCP stdio and streamable-http transport implementation
  do-not-fabricate: true
  version: v0.46.0
  notes: >
    Use v0.46.0 exactly. See hints/mcp-server.go.mcp-go.hints.md
    for API shapes and known gotchas.

---

## DELIVERABLES

COMPONENT: implementation
  files: main.go, internal/lint/*.go, internal/store/*.go
  notes: >
    Split into packages: main (transport wiring), internal/lint
    (rule engine, shared with pcd-lint), internal/store (unified
    AssetStore — templates, hints, prompts). Reuse lint rule logic
    from pcd-lint if available as a library; otherwise inline.

COMPONENT: module
  files: go.mod
  notes: Direct dependencies only. go mod tidy required before building.

COMPONENT: build
  files: Makefile
  notes: >
    Must include an embed-assets target that stages templates, hints,
    and prompts into the build-time embedding location before go build.
    Must include a clean target that removes staged directories.
    See hints/mcp-server.go.mcp-go.hints.md for exact paths and commands.

COMPONENT: packaging
  files: mcp-server-pcd.spec, debian/control, debian/changelog,
         debian/rules, debian/copyright
  notes: >
    Source0: mcp-server-pcd-{version}.tar.xz
    Source1: mcp-server-pcd-{version}-vendor.tar.xz
    Vendor tarball generated by: go mod vendor &&
      tar -cJf mcp-server-pcd-{version}-vendor.tar.xz vendor/
    RPM %build must run make embed-assets before go build.
    Requires: (none). Recommends: pcd-templates.

COMPONENT: container
  files: Containerfile
  notes: >
    Multi-stage build. Final stage FROM scratch.
    EXPOSE 8080. ENTRYPOINT defaults to http transport.

COMPONENT: service-unit
  files: mcp-server-pcd.service
  notes: systemd unit for http transport; socket activation optional.

COMPONENT: license
  files: LICENSE

COMPONENT: tests
  files: independent_tests/INDEPENDENT_TESTS.go
  notes: >
    All tests use FakeStore, FakeFilesystem.
    No filesystem access. No network calls. No live pcd-lint binary.
    Must include TestLintMatchesCLI (verifies lint_content invariant).
    FakeStore replaces FakeTemplateStore and FakePromptStore.

COMPONENT: documentation
  files: README.md

COMPONENT: report
  files: TRANSLATION_REPORT.md, translation_report/translation-workflow.pikchr

---

## DEPLOYMENT

Runtime: Linux service (http transport) or subprocess (stdio transport).

Install locations:
  Binary:        /usr/bin/mcp-server-pcd
  System config: /etc/pcd/
  Service unit:  /usr/lib/systemd/system/mcp-server-pcd.service

The binary embeds all templates, hints, and prompts at build time.
It is self-contained and functional without pcd-templates installed.

Install pcd-templates to enable site-local asset overrides:
  Templates:  /usr/share/pcd/templates/   (pcd-templates)
  Hints:      /usr/share/pcd/hints/        (pcd-templates)
  Prompts:    /usr/share/pcd/prompts/      (pcd-templates)

Runtime dependency:
  Requires:    (none — binary is self-contained)
  Recommends:  pcd-templates

Asset overlay search path (last-wins, ascending precedence):
  1. /usr/share/pcd/templates|hints|prompts/   (pcd-templates)
  2. /etc/pcd/templates|hints|prompts/          (system administrator)
  3. ~/.config/pcd/templates|hints|prompts/     (user)
  4. ./.pcd/templates|hints|prompts/            (project-local)
  Directories that do not exist are silently skipped.

mcphost config (stdio):
```yaml
mcpServers:
  pcd:
    command: mcp-server-pcd
    args: [stdio]
```

mcphost config (http, if running as service):
```yaml
mcpServers:
  pcd:
    url: http://127.0.0.1:8080/mcp
```

Version in preset hierarchy:
  Specs reference the server by connecting to it — no version declaration
  in the spec itself. Server advertises schema version via get_schema_version.
