


# mcp-server-pcd

## META
Deployment:        mcp-server
Version:           0.3.2
Spec-Schema:       0.3.21
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
// Format: "<template>.<language>.<library>" (library-specific hints)
//      or "<template>.<language>.milestones" (scaffold-first hints)
//      or "<component>.implementation"       (component-specific hints)
// Examples:
//   "cloud-native.go.go-libvirt"
//   "cli-tool.go.milestones"
//   "sitar.implementation"

ResourceURI := string
// Format: "pcd://<type>/<n>"
// Types: templates, prompts, hints
// Examples:
//   pcd://templates/cli-tool
//   pcd://prompts/interview
//   pcd://prompts/reverse
//   pcd://prompts/translator
//   pcd://hints/cloud-native.go.go-libvirt
//   pcd://hints/cli-tool.go.milestones

Diagnostic := {
  severity:  "error" | "warning"
  line:      integer       // 1-based; 1 for file-level diagnostics
  section:   string        // e.g. "META", "BEHAVIOR", "structure"
  message:   string
  rule:      string        // e.g. "RULE-01", "RULE-08", "RULE-17"
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

MilestoneStatus := "pending" | "active" | "failed" | "released"
// Pipeline state for ## MILESTONE sections.
// Transitions: pending → active → released (pass) or failed (fail).
// Exactly one milestone may be active at any time (RULE-15).
// Status is managed by the agent pipeline, not by the spec author.

SetMilestoneResult := {
  spec_path:        string          // path of the modified spec file
  milestone_name:   string          // e.g. "0.1.0"
  previous_status:  MilestoneStatus
  new_status:       MilestoneStatus
}

SpecHashResult := {
  spec_path:      string    // path of the spec file
  spec_hash:      string    // SHA256 hex digest of the spec file
  report_hash:    string    // Spec-SHA256 value from TRANSLATION_REPORT.md, or "" if absent
  match:          boolean   // true if spec_hash = report_hash
  status:         string    // "current" | "stale" | "no-report" | "no-hash-in-report"
}

ChangeImpactRecommendation := "full-regeneration" | "incremental"

ChangeImpactResult := {
  recommendation:  ChangeImpactRecommendation
  primary_factor:  string          // one-sentence reason for the recommendation
  structural_impact: string        // high | medium | low
  blast_radius:    string          // description of affected scope
  scaffold_affected: boolean       // true if scaffold milestone is in scope
  released_milestone_affected: boolean  // true if any released milestone is in scope
  consistency_risk: string         // high | medium | low
  if_incremental:  string          // what to change (empty if full-regeneration)
  if_regeneration: string          // what to preserve (empty if incremental)
  reasoning:       string          // full assessment narrative
}

```

---

## INTERFACES

```
Filesystem {
  // Used by lint_file and set_milestone_status.
  required-methods:
    ReadFile(path)              -> (content: string, error)
    WriteFile(path, content)    -> error
  implementations-required:
    production:  OSFilesystem
    test-double: FakeFilesystem {
      configurable: Files map[string]string,
                    ReadErr map[string]error,
                    WriteErr map[string]error
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
```
none
```

PRECONDITIONS:
- AssetStore is initialised

STEPS:
1. Call AssetStore.ListTemplates(); on error → MCP error -32603.
2. For each template record, omit the content field from the response.
3. Return the list of records.
   MECHANISM: content is excluded from list responses to keep response size small;
   callers use get_template to fetch content.

POSTCONDITIONS:
- response contains one entry per installed template
- each entry has name, version, language; content is absent

ERRORS:
- MCP error -32603 on AssetStore failure

---

## BEHAVIOR: get_template
Constraint: required

INPUTS:
```
name:    TemplateName
version: TemplateVersion   // default: "latest"
```

PRECONDITIONS:
- name is non-empty

STEPS:
1. Call AssetStore.GetTemplate(name, version); on not found →
   MCP error -32602 with message "unknown template: {name}".
2. Return the full TemplateRecord including content.

POSTCONDITIONS:
- response contains name, version, language, and full Markdown content

ERRORS:
- MCP error -32602 if template name is unknown

---

## BEHAVIOR: list_resources
Constraint: required

INPUTS:
```
none
```

STEPS:
1. Call AssetStore.ListTemplates(), AssetStore.ListHintsKeys(),
   AssetStore.ListPrompts() in any order; on error → MCP error -32603.
2. Construct ResourceRecord list:
   - templates: uri = "pcd://templates/{name}"
   - hints:     uri = "pcd://hints/{key}"
   - prompts:   uri = "pcd://prompts/{name}"
3. Return the combined list.

POSTCONDITIONS:
- all installed templates, hints files, and prompts are represented
- each entry has uri and name; content is absent

ERRORS:
- MCP error -32603 on AssetStore failure

---

## BEHAVIOR: read_resource
Constraint: required

INPUTS:
```
uri: ResourceURI
```

PRECONDITIONS:
- uri matches pattern "pcd://<type>/<n>"

STEPS:
1. Parse uri; if format does not match "pcd://<type>/<n>" →
   MCP error -32602 with message "invalid resource URI: {uri}".
2. Dispatch by type:
   - "templates": call AssetStore.GetTemplate(n, "latest")
   - "hints":     call AssetStore.GetHints(n)
   - "prompts":   call AssetStore.GetPrompt(n)
   On unknown type → MCP error -32602.
3. On not found → MCP error -32602 with message "resource not found: {uri}".
4. Return ResourceRecord with uri, name, and full content.

POSTCONDITIONS:
- response contains uri, name, and full Markdown content

ERRORS:
- MCP error -32602 for invalid URI format
- MCP error -32602 for unknown resource type
- MCP error -32602 for resource not found

---

## BEHAVIOR: lint_content
Constraint: required

Validates a PCD specification given as a string. Applies all rules
RULE-01 through RULE-17, identical to the pcd-lint CLI.

INPUTS:
```
content:  string    // full spec Markdown text
filename: string    // used for diagnostics; must have .md extension
```

PRECONDITIONS:
- content is non-empty
- filename has .md extension

STEPS:
1. If filename does not end in ".md" →
   MCP error -32602 with message "filename must have .md extension: {filename}".
2. Run the embedded lint engine on content with filename.
   The lint engine applies RULE-01 through RULE-17 in order.
   All rules run regardless of earlier failures.
3. Return LintResult.

POSTCONDITIONS:
- result.valid = true iff result.errors = 0
- result.diagnostics contains one entry per rule violation
- result is identical to pcd-lint CLI output for the same input

ERRORS:
- MCP error -32602 if filename does not have .md extension

---

## BEHAVIOR: lint_file
Constraint: required

INPUTS:
```
path: string    // absolute path to spec file on disk
```

STEPS:
1. Read file via Filesystem.ReadFile(path); on error →
   MCP error -32602 with message "cannot open file: {path}".
2. Extract filename = basename(path).
3. Run lint_content(content, filename).

POSTCONDITIONS:
- Same as lint_content postconditions.

ERRORS:
- MCP error -32602 if file not found or not readable

---

## BEHAVIOR: get_schema_version
Constraint: required

INPUTS:
```
none
```

STEPS:
1. Return the Spec-Schema version this binary was built against.

POSTCONDITIONS:
- response is a semantic version string (e.g. "0.3.21")

ERRORS:
none

---

## BEHAVIOR: set_milestone_status
Constraint: required

Sets the `Status:` field of a named MILESTONE section in a spec file on disk.
Used by the agent pipeline to advance the milestone cursor without human
intervention.

INPUTS:
```
spec_path:        string          // absolute path to the spec .md file
milestone_name:   string          // exact MILESTONE label, e.g. "0.1.0"
new_status:       MilestoneStatus // pending | active | failed | released
```

PRECONDITIONS:
- spec_path exists and is readable and writable
- spec_path has .md extension
- milestone_name matches an existing ## MILESTONE: section in the file
- new_status is a valid MilestoneStatus value

STEPS:
1. Read spec_path from disk via Filesystem.ReadFile; on error →
   MCP error -32602 with message "cannot open file: {spec_path}".
2. Locate the `## MILESTONE: {milestone_name}` section; on not found →
   MCP error -32602 with message
   "MILESTONE '{milestone_name}' not found in {spec_path}".
3. If new_status = "active": scan all other MILESTONE sections in the file.
   If any other section already has `Status: active` →
   MCP error -32602 with message
   "Cannot set MILESTONE '{milestone_name}' to active: MILESTONE '{other}'
    is already active. Set it to released or failed first."
4. Record previous_status (current Status: value, or "pending" if absent).
5. Replace or insert the `Status: {value}` line within the located MILESTONE section.
   MECHANISM: the Status: line must be the first non-blank line after the
   ## MILESTONE: header line. If no Status: line is present, insert one.
   Do not modify any other content in the file.
6. Write the modified content back to spec_path via Filesystem.WriteFile;
   on error → MCP error -32603.
7. Return SetMilestoneResult.

POSTCONDITIONS:
- spec_path on disk has exactly the Status: value changed for the named milestone
- All other content in the file is byte-for-byte identical to the input
- If new_status = "active", no other milestone in the file has Status: active
- result.previous_status reflects the status before this call
- result.new_status = new_status

ERRORS:
- MCP error -32602 if spec_path not found or not readable
- MCP error -32602 if milestone_name not found in spec
- MCP error -32602 if new_status = "active" and another milestone is already active
- MCP error -32603 if write to spec_path fails

---

## BEHAVIOR: assess_change_impact
Constraint: required

Analyses a specification change and recommends the most appropriate
translation strategy: full regeneration from scratch or incremental
update of the existing generated code.

Given the change description and optionally the existing spec and code,
the tool applies the PCD regeneration strategy framework (see whitepaper
A.19) to estimate structural impact, blast radius, scaffold involvement,
and consistency risk.

INPUTS:
```
change_description: string   // unified diff of the spec, or plain-language
                             // description of what changed — required
old_spec:           string   // full spec content before the change — optional
                             // but recommended; improves blast radius analysis
new_spec:           string   // full spec content after the change — optional
existing_code:      string   // generated implementation — optional; if provided,
                             // the tool identifies affected files and functions
```

PRECONDITIONS:
- change_description is non-empty
- At least change_description must be provided; all other inputs are optional

STEPS:
1. Parse change_description to identify which spec sections are affected
   (TYPES, INTERFACES, INVARIANTS, BEHAVIOR, EXAMPLES, MILESTONE, META).
2. If old_spec or new_spec provided: extract the full set of changed elements
   with their section types.
3. Evaluate structural impact:
   a. If TYPES, INTERFACES, or INVARIANTS are affected → structural_impact = "high"
   b. If only BEHAVIOR STEPS or EXAMPLES are affected → structural_impact = "low"
      or "medium" depending on count
   c. If only META is affected → structural_impact = "none"
4. Evaluate scaffold involvement:
   a. If any MILESTONE with Scaffold: true is in the changed scope →
      scaffold_affected = true
   b. Otherwise → scaffold_affected = false
5. Evaluate released milestone involvement:
   a. If any MILESTONE with Status: released is in the changed scope →
      released_milestone_affected = true
   b. Otherwise → released_milestone_affected = false
6. Estimate blast radius:
   a. If existing_code provided: count files and functions referencing changed elements
   b. If not provided: estimate from spec cross-references
   c. Classify: "1–2 BEHAVIORs" | "3–5 BEHAVIORs" | "5+ BEHAVIORs or shared types"
7. Assess consistency risk based on codebase provenance if inferable from inputs.
8. Apply decision rules:
   - If structural_impact = "high" OR scaffold_affected = true OR
     released_milestone_affected = true → recommendation = "full-regeneration"
   - If structural_impact = "low" AND blast_radius ≤ 2 BEHAVIORs AND
     scaffold_affected = false AND released_milestone_affected = false →
     recommendation = "incremental"
   - Otherwise → recommendation = "full-regeneration" (conservative default)
9. Compose reasoning narrative and populate if_incremental or if_regeneration
   fields as appropriate.
10. Return ChangeImpactResult.

POSTCONDITIONS:
- result.recommendation is always set
- result.primary_factor states the single most important deciding factor
- result.reasoning provides a complete narrative suitable for the audit bundle
- if result.recommendation = "incremental": result.if_incremental lists
  the specific files/functions/sections to change
- if result.recommendation = "full-regeneration": result.if_regeneration
  lists decisions from the existing code worth preserving as translator notes

ERRORS:
- MCP error -32602 if change_description is empty

---

## BEHAVIOR: verify_spec_hash
Constraint: required

Computes the SHA256 of a spec file and compares it to the `Spec-SHA256:`
field recorded in the most recent `TRANSLATION_REPORT.md` adjacent to the
spec. Reports whether the generated artifacts are current with the spec.

INPUTS:
```
spec_path: string   // path to the spec .md file — required
```

PRECONDITIONS:
- spec_path is non-empty
- spec_path points to a readable file with .md extension

STEPS:
1. Compute SHA256 of the file at spec_path. Store as spec_hash.
2. Look for TRANSLATION_REPORT.md in the same directory as spec_path,
   then in a `code/` subdirectory of the spec's parent directory.
   If not found: return SpecHashResult with status = "no-report".
3. If found: search for a line matching `Spec-SHA256: <hex>` in the report.
   If not found: return SpecHashResult with status = "no-hash-in-report".
4. Extract report_hash from the `Spec-SHA256:` line.
5. If spec_hash = report_hash:
     return SpecHashResult with match = true, status = "current"
   Else:
     return SpecHashResult with match = false, status = "stale"

POSTCONDITIONS:
- result.spec_hash is always the current SHA256 of the spec file
- result.status is one of: "current" | "stale" | "no-report" | "no-hash-in-report"
- result.match is true only when status = "current"

ERRORS:
- MCP error -32602 if spec_path is empty or file not readable
- MCP error -32602 if spec_path does not end in .md

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

- Server never modifies any file on disk except via set_milestone_status.
- Server never makes outbound network calls.
- Server never reads environment variables for behaviour control.
- lint_content and lint_file produce identical output to pcd-lint CLI
  for identical input, including RULE-01 through RULE-17.
- assess_change_impact applies the decision rules from whitepaper A.19
  deterministically — same inputs always produce the same recommendation.
- All MCP responses are valid JSON-RPC 2.0.

---

## INVARIANTS

- [observable]      stdio transport: stdout contains only MCP JSON-RPC messages
- [observable]      lint_content result is identical to pcd-lint CLI on same input
                    for RULE-01 through RULE-17
- [observable]      set_milestone_status never modifies any content in the spec
                    other than the Status: line of the named milestone
- [observable]      set_milestone_status with new_status=active fails if any other
                    milestone already has Status: active in the same file
- [observable]      server is idempotent: same request always returns same response
- [observable]      server never exits with code other than 0, 1, or 2
- [observable]      assess_change_impact with structural_impact=high always
                    returns recommendation=full-regeneration
- [observable]      verify_spec_hash with matching hashes always returns
                    status="current" and match=true
- [implementation]  rule execution order: RULE-01 through RULE-17, same as pcd-lint
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

### EXAMPLE: list_templates_returns_names
GIVEN:
  AssetStore contains cli-tool@0.3.21 and mcp-server@0.3.21
WHEN:
  tool list_templates called with no arguments
THEN:
  response contains two entries
  each entry has name, version, language
  content field is absent from each entry

### EXAMPLE: get_template_cli_tool
GIVEN:
  AssetStore contains cli-tool@0.3.21
WHEN:
  tool get_template called with name="cli-tool" version="latest"
THEN:
  response contains name="cli-tool" version="0.3.21"
  response contains full template Markdown in content field

### EXAMPLE: get_template_unknown
GIVEN:
  AssetStore does not contain "serverless"
WHEN:
  tool get_template called with name="serverless"
THEN:
  MCP error -32602 returned
  message contains "unknown template: serverless"

### EXAMPLE: read_resource_interview_prompt
GIVEN:
  AssetStore contains prompt named "interview"
WHEN:
  tool read_resource called with uri="pcd://prompts/interview"
THEN:
  response contains uri="pcd://prompts/interview"
  response contains full interview prompt Markdown in content field

### EXAMPLE: read_resource_reverse_prompt
GIVEN:
  AssetStore contains prompt named "reverse"
WHEN:
  tool read_resource called with uri="pcd://prompts/reverse"
THEN:
  response contains uri="pcd://prompts/reverse"
  response contains full reverse prompt Markdown in content field

### EXAMPLE: read_resource_milestones_hints
GIVEN:
  AssetStore contains hints file with key "cli-tool.go.milestones"
WHEN:
  tool read_resource called with uri="pcd://hints/cli-tool.go.milestones"
THEN:
  response contains uri="pcd://hints/cli-tool.go.milestones"
  response contains full hints file Markdown in content field

### EXAMPLE: read_resource_invalid_uri
GIVEN:
  any server state
WHEN:
  tool read_resource called with uri="http://example.com/bad"
THEN:
  MCP error -32602 returned
  message contains "invalid resource URI"

### EXAMPLE: lint_content_valid_spec
GIVEN:
  content is a valid PCD spec with all required sections
WHEN:
  tool lint_content called with that content and filename="myspec.md"
THEN:
  result.valid = true
  result.errors = 0
  result.diagnostics is empty

### EXAMPLE: lint_content_missing_invariants
GIVEN:
  content is a PCD spec missing the INVARIANTS section
WHEN:
  tool lint_content called with that content
THEN:
  result.valid = false
  result.errors >= 1
  one diagnostic has rule="RULE-01" severity="error"
  diagnostic message contains "INVARIANTS"

### EXAMPLE: lint_content_milestone_scaffold_not_first
GIVEN:
  content is a PCD spec with two MILESTONE sections:
    ## MILESTONE: 0.1.0  (no Scaffold field)
    ## MILESTONE: 0.0.0  Scaffold: true
  the scaffold milestone appears second in document order
WHEN:
  tool lint_content called with that content and filename="myspec.md"
THEN:
  result.valid = false
  result.errors >= 1
  one diagnostic has rule="RULE-17"
  diagnostic message contains "must appear first"

### EXAMPLE: lint_content_two_scaffold_milestones
GIVEN:
  content is a PCD spec with two MILESTONE sections both having Scaffold: true
WHEN:
  tool lint_content called with that content and filename="myspec.md"
THEN:
  result.valid = false
  result.errors >= 1
  one diagnostic has rule="RULE-17"
  diagnostic message contains "more than one MILESTONE has Scaffold: true"

### EXAMPLE: lint_content_bad_extension
GIVEN:
  any valid spec content
WHEN:
  tool lint_content called with filename="myspec.txt"
THEN:
  MCP error -32602 returned
  message contains ".md extension"

### EXAMPLE: lint_file_not_found
GIVEN:
  path "/tmp/missing.md" does not exist
WHEN:
  tool lint_file called with path="/tmp/missing.md"
THEN:
  MCP error -32602 returned
  message contains "cannot open file"

### EXAMPLE: lint_content_matches_cli
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

### EXAMPLE: stdio_startup
GIVEN:
  server started with bare word "stdio"
WHEN:
  MCP initialize request sent on stdin
THEN:
  valid MCP initialize response written to stdout
  server capabilities include tools and resources

### EXAMPLE: http_startup
GIVEN:
  server started with bare word "http"
WHEN:
  HTTP POST to /mcp with MCP initialize request
THEN:
  valid MCP initialize response returned
  HTTP status 200

### EXAMPLE: http_bind_failure
GIVEN:
  port 8080 is already in use
WHEN:
  server started with "http" and listen=127.0.0.1:8080
THEN:
  error message written to stderr
  server exits with code 1

### EXAMPLE: standalone_no_pcd_templates
GIVEN:
  pcd-templates package is not installed
  no overlay directories exist on the system
WHEN:
  server started and list_templates called
THEN:
  response contains all templates shipped with the binary
  server does not error or exit

### EXAMPLE: verify_spec_hash_stale
GIVEN:
  spec_path = "tools/calc-interest/spec/calc-interest.md"
  SHA256 of spec file = "abc123...def456"
  TRANSLATION_REPORT.md contains: "Spec-SHA256: 000000...111111"
WHEN:
  tool verify_spec_hash called with spec_path
THEN:
  result.spec_hash = "abc123...def456"
  result.report_hash = "000000...111111"
  result.match = false
  result.status = "stale"

### EXAMPLE: verify_spec_hash_current
GIVEN:
  spec_path = "tools/pcd-lint/spec/pcd-lint.md"
  SHA256 of spec file = "aabbcc...ddeeff"
  TRANSLATION_REPORT.md contains: "Spec-SHA256: aabbcc...ddeeff"
WHEN:
  tool verify_spec_hash called with spec_path
THEN:
  result.match = true
  result.status = "current"

### EXAMPLE: assess_change_type_modification
GIVEN:
  change_description = "Changed Money type: added currency field (string, ISO 4217)"
  old_spec contains:
    ## TYPES
    Money := decimal where precision = 2, value >= 0
  new_spec contains:
    ## TYPES
    Money := { amount: decimal where precision = 2, value >= 0
               currency: string where matches ISO-4217 }
WHEN:
  tool assess_change_impact called with those inputs
THEN:
  result.recommendation = "full-regeneration"
  result.structural_impact = "high"
  result.scaffold_affected = false
  result.primary_factor contains "TYPES"
  result.reasoning explains that Money is used across multiple BEHAVIORs

### EXAMPLE: assess_change_steps_isolated
GIVEN:
  change_description = "BEHAVIOR: compute-interest STEPS: added rounding
    to nearest cent in step 1"
  The spec has 4 BEHAVIORs; only compute-interest uses Money output
WHEN:
  tool assess_change_impact called with those inputs
THEN:
  result.recommendation = "incremental"
  result.structural_impact = "low"
  result.blast_radius contains "1"
  result.if_incremental identifies compute-interest implementation function

### EXAMPLE: assess_change_scaffold
GIVEN:
  change_description = "Added new package internal/cache to M0 scaffold milestone"
WHEN:
  tool assess_change_impact called with that input
THEN:
  result.recommendation = "full-regeneration"
  result.scaffold_affected = true
  result.primary_factor contains "scaffold"

### EXAMPLE: set_milestone_active
GIVEN:
  spec file "/tmp/sitar.md" contains:
    ## MILESTONE: 0.1.0
    Status: pending
    ## MILESTONE: 0.2.0
    Status: pending
WHEN:
  tool set_milestone_status called with
    spec_path="/tmp/sitar.md" milestone_name="0.1.0" new_status="active"
THEN:
  file "/tmp/sitar.md" has Status: active under ## MILESTONE: 0.1.0
  file "/tmp/sitar.md" has Status: pending under ## MILESTONE: 0.2.0
  result.previous_status = "pending"
  result.new_status = "active"
  no other content in the file is changed

### EXAMPLE: set_milestone_active_conflict
GIVEN:
  spec file "/tmp/sitar.md" contains:
    ## MILESTONE: 0.1.0
    Status: active
    ## MILESTONE: 0.2.0
    Status: pending
WHEN:
  tool set_milestone_status called with
    spec_path="/tmp/sitar.md" milestone_name="0.2.0" new_status="active"
THEN:
  MCP error -32602 returned
  message contains "MILESTONE '0.1.0' is already active"
  file "/tmp/sitar.md" is not modified

### EXAMPLE: set_milestone_released
GIVEN:
  spec file "/tmp/sitar.md" contains:
    ## MILESTONE: 0.1.0
    Status: active
WHEN:
  tool set_milestone_status called with
    spec_path="/tmp/sitar.md" milestone_name="0.1.0" new_status="released"
THEN:
  file "/tmp/sitar.md" has Status: released under ## MILESTONE: 0.1.0
  result.previous_status = "active"
  result.new_status = "released"

### EXAMPLE: list_resources_includes_workflow_prompts
GIVEN:
  AssetStore is initialised from the embedded build
WHEN:
  tool list_resources called with no arguments
THEN:
  response contains pcd://prompts/translator
  response contains pcd://prompts/interview
  response contains pcd://prompts/reverse
  response contains pcd://prompts/change-impact
  response contains pcd://prompts/reviewer
  response contains pcd://prompts/tiebreaker

### EXAMPLE: read_resource_change_impact_prompt
GIVEN:
  AssetStore contains the change-impact prompt
WHEN:
  tool read_resource called with uri="pcd://prompts/change-impact"
THEN:
  response.content begins with "# PCD Change Impact Assessment Prompt"
  response.mime_type = "text/markdown"

### EXAMPLE: read_resource_reviewer_prompt
GIVEN:
  AssetStore contains the reviewer prompt
WHEN:
  tool read_resource called with uri="pcd://prompts/reviewer"
THEN:
  response.content begins with "# PCD Translation Review Prompt"
  response.mime_type = "text/markdown"

### EXAMPLE: read_resource_tiebreaker_prompt
GIVEN:
  AssetStore contains the tiebreaker prompt
WHEN:
  tool read_resource called with uri="pcd://prompts/tiebreaker"
THEN:
  response.content begins with "# PCD Translation Tie-Breaker Prompt"
  response.mime_type = "text/markdown"

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
      // e.g. "cli-tool.go.milestones.hints.md" -> key "cli-tool.go.milestones"
      // e.g. "sitar.implementation.hints.md"   -> key "sitar.implementation"
      // e.g. "python-tool.hints.md"            -> key "python-tool"

  - type: prompts
    source: repo-root/prompts/*.md
    filter: exclude README-*.md
    key-derivation: filename stem before ".md", with special mapping:
      - prompt.md            → "translator"
      - interview-prompt.md  → "interview"
      - reverse-prompt.md    → "reverse"
    known-keys:
      - translator      // spec → code translation prompt
      - interview       // guided spec authoring → spec
      - reverse         // reverse-engineer code → spec
      - change-impact   // assess spec change impact → recommendation
      - reviewer        // single-translation review → REVIEW_REPORT.md
      - tiebreaker      // dual-translation divergence → TIEBREAK_REPORT.md
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
  files: main.go, internal/lint/*.go, internal/store/*.go, internal/milestone/*.go
  notes: >
    Split into packages: main (transport wiring), internal/lint
    (rule engine applying RULE-01 through RULE-17, shared with pcd-lint),
    internal/store (unified AssetStore — templates, hints, prompts),
    internal/milestone (set_milestone_status file editing logic),
    internal/changeimpact (assess_change_impact analysis logic).
    Reuse lint rule logic from pcd-lint if available as a library;
    otherwise inline. The lint engine must implement all 17 rules.

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
    Multi-stage build. Builder FROM registry.suse.com/bci/golang:latest.
    Final stage FROM scratch.
    EXPOSE 8080. ENTRYPOINT defaults to http transport.

COMPONENT: service-unit
  files: mcp-server-pcd.service
  notes: systemd unit for http transport; socket activation optional.

COMPONENT: license
  files: LICENSE
  notes: GPL-2.0-only — SPDX identifier and URL only, do not reproduce full text.

COMPONENT: tests
  files: independent_tests/INDEPENDENT_TESTS.go
  notes: >
    All tests use FakeStore, FakeFilesystem.
    No filesystem access. No network calls. No live pcd-lint binary.
    Must include TestLintMatchesCLI (verifies lint_content invariant).
    Must include TestSetMilestoneStatus (verifies file editing invariants).
    FakeFilesystem must support both ReadFile and WriteFile.

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
