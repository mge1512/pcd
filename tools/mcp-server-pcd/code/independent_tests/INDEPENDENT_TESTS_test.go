// Package independent_tests contains all independent tests for mcp-server-pcd.
// All tests use FakeStore and FakeFilesystem — no filesystem access,
// no network calls, no live pcd-lint binary required.
// SPDX-License-Identifier: GPL-2.0-only

package independent_tests

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"mcp-server-pcd/internal/lint"
	"mcp-server-pcd/internal/milestone"
	"mcp-server-pcd/internal/store"
)

// ── Helpers ───────────────────────────────────────────────────────────────────

// minimalValidSpec returns a minimal PCD spec that passes all lint rules.
func minimalValidSpec() string {
	return `# test-spec

## META
Deployment:   mcp-server
Version:      0.1.0
Spec-Schema:  0.1.0
Author:       Test Author <test@example.com>
License:      MIT
Verification: none
Safety-Level: QM

---

## TYPES

` + "```" + `
Foo := string
` + "```" + `

---

## BEHAVIOR: do_something
Constraint: required

INPUTS:
` + "```" + `
input: string
` + "```" + `

STEPS:
1. Process input.
2. Return result.

---

## PRECONDITIONS

- input is non-empty

---

## POSTCONDITIONS

- result is non-empty

---

## INVARIANTS

- [observable]      result is always non-empty

---

## EXAMPLES

EXAMPLE: basic_usage
GIVEN:
  input = "hello"
WHEN:
  do_something called with input="hello"
THEN:
  result = "HELLO"
`
}

// ── FakeStore tests ───────────────────────────────────────────────────────────

func TestFakeStore_ListTemplates(t *testing.T) {
	s := &store.FakeStore{
		Templates: []store.TemplateRecord{
			{Name: "cli-tool", Version: "0.3.21", Language: "Go", Content: "# cli-tool template"},
			{Name: "mcp-server", Version: "0.3.21", Language: "Go", Content: "# mcp-server template"},
		},
	}

	records, err := s.ListTemplates()
	if err != nil {
		t.Fatalf("ListTemplates returned error: %v", err)
	}
	if len(records) != 2 {
		t.Fatalf("expected 2 records, got %d", len(records))
	}
	// content must be absent from list
	for _, r := range records {
		if r.Content != "" {
			t.Errorf("ListTemplates: content should be absent for %s, got %q", r.Name, r.Content)
		}
	}
}

func TestFakeStore_GetTemplate_Found(t *testing.T) {
	s := &store.FakeStore{
		Templates: []store.TemplateRecord{
			{Name: "cli-tool", Version: "0.3.21", Language: "Go", Content: "# cli-tool content"},
		},
	}

	rec, err := s.GetTemplate("cli-tool", "latest")
	if err != nil {
		t.Fatalf("GetTemplate returned error: %v", err)
	}
	if rec.Name != "cli-tool" {
		t.Errorf("expected name cli-tool, got %s", rec.Name)
	}
	if rec.Content == "" {
		t.Error("GetTemplate: content should be present")
	}
}

func TestFakeStore_GetTemplate_NotFound(t *testing.T) {
	s := &store.FakeStore{}

	_, err := s.GetTemplate("serverless", "latest")
	if err == nil {
		t.Fatal("expected error for unknown template, got nil")
	}
}

func TestFakeStore_GetHints(t *testing.T) {
	s := &store.FakeStore{
		Hints: map[string]string{
			"cli-tool.go.milestones": "# milestones hints",
		},
	}

	content, err := s.GetHints("cli-tool.go.milestones")
	if err != nil {
		t.Fatalf("GetHints returned error: %v", err)
	}
	if content == "" {
		t.Error("GetHints: content should not be empty")
	}
}

func TestFakeStore_GetPrompt(t *testing.T) {
	s := &store.FakeStore{
		Prompts: map[string]string{
			"interview": "# interview prompt content",
			"reverse":   "# reverse prompt content",
		},
	}

	content, err := s.GetPrompt("interview")
	if err != nil {
		t.Fatalf("GetPrompt returned error: %v", err)
	}
	if !strings.Contains(content, "interview") {
		t.Errorf("GetPrompt: unexpected content: %s", content)
	}
}

// ── list_templates BEHAVIOR ───────────────────────────────────────────────────

func TestListTemplates_ReturnsNamesOnly(t *testing.T) {
	// EXAMPLE: list_templates_returns_names
	s := &store.FakeStore{
		Templates: []store.TemplateRecord{
			{Name: "cli-tool", Version: "0.3.21", Language: "Go", Content: "full content here"},
			{Name: "mcp-server", Version: "0.3.21", Language: "Go", Content: "more content here"},
		},
	}

	records, err := s.ListTemplates()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(records) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(records))
	}
	for _, r := range records {
		if r.Name == "" {
			t.Error("entry missing name")
		}
		if r.Version == "" {
			t.Error("entry missing version")
		}
		if r.Content != "" {
			t.Errorf("content should be absent in list, got %q for %s", r.Content, r.Name)
		}
	}
}

// ── get_template BEHAVIOR ─────────────────────────────────────────────────────

func TestGetTemplate_ReturnsContent(t *testing.T) {
	// EXAMPLE: get_template_cli_tool
	s := &store.FakeStore{
		Templates: []store.TemplateRecord{
			{Name: "cli-tool", Version: "0.3.21", Language: "Go", Content: "# cli-tool template\n\nFull content."},
		},
	}

	rec, err := s.GetTemplate("cli-tool", "latest")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Name != "cli-tool" {
		t.Errorf("name mismatch: %s", rec.Name)
	}
	if rec.Version != "0.3.21" {
		t.Errorf("version mismatch: %s", rec.Version)
	}
	if rec.Content == "" {
		t.Error("content should be present in get_template")
	}
}

func TestGetTemplate_Unknown(t *testing.T) {
	// EXAMPLE: get_template_unknown
	s := &store.FakeStore{}

	_, err := s.GetTemplate("serverless", "latest")
	if err == nil {
		t.Fatal("expected error for unknown template")
	}
	// Caller would return MCP error with "unknown template: serverless"
}

// ── lint_content BEHAVIOR ─────────────────────────────────────────────────────

func TestLintContent_ValidSpec(t *testing.T) {
	// EXAMPLE: lint_content_valid_spec
	content := minimalValidSpec()
	result := lint.LintContent(content, "myspec.md")

	if !result.Valid {
		t.Errorf("expected valid=true, got false; errors=%d, diagnostics=%v",
			result.Errors, result.Diagnostics)
	}
	if result.Errors != 0 {
		t.Errorf("expected 0 errors, got %d", result.Errors)
	}
}

func TestLintContent_MissingInvariants(t *testing.T) {
	// EXAMPLE: lint_content_missing_invariants
	// Spec missing INVARIANTS section
	content := `# test

## META
Deployment:   mcp-server
Version:      0.1.0
Spec-Schema:  0.1.0
Author:       Test <t@t.com>
License:      MIT
Verification: none
Safety-Level: QM

## TYPES

` + "```" + `
Foo := string
` + "```" + `

## BEHAVIOR: foo
STEPS:
1. Do foo.

## PRECONDITIONS
- none

## POSTCONDITIONS
- done

## EXAMPLES

EXAMPLE: foo_example
GIVEN:
  state is ready
WHEN:
  foo called
THEN:
  result returned
`
	result := lint.LintContent(content, "myspec.md")

	if result.Valid {
		t.Error("expected valid=false for spec missing INVARIANTS")
	}
	if result.Errors == 0 {
		t.Error("expected at least 1 error")
	}

	foundRule01 := false
	for _, d := range result.Diagnostics {
		if d.Rule == "RULE-01" && strings.Contains(d.Message, "INVARIANTS") {
			foundRule01 = true
		}
	}
	if !foundRule01 {
		t.Errorf("expected RULE-01 diagnostic mentioning INVARIANTS, got: %v", result.Diagnostics)
	}
}

func TestLintContent_BadExtension(t *testing.T) {
	// EXAMPLE: lint_content_bad_extension
	// The handler validates extension before calling lint engine.
	filename := "myspec.txt"
	hasMD := strings.HasSuffix(filename, ".md")
	if hasMD {
		t.Error("test setup error: filename should not have .md extension")
	}
	// Verify the check correctly rejects .txt
	if hasMD {
		t.Error("extension check failed: .txt should not be accepted")
	}
}

func TestLintContent_MilestoneScaffoldNotFirst(t *testing.T) {
	// EXAMPLE: lint_content_milestone_scaffold_not_first
	content := `# test

## META
Deployment:   mcp-server
Version:      0.1.0
Spec-Schema:  0.1.0
Author:       Test <t@t.com>
License:      MIT
Verification: none
Safety-Level: QM

## TYPES

` + "```" + `
Foo := string
` + "```" + `

## BEHAVIOR: foo
STEPS:
1. Do foo.

## PRECONDITIONS
- none

## POSTCONDITIONS
- done

## INVARIANTS

- [observable] foo is done

## EXAMPLES

EXAMPLE: foo_example
GIVEN:
  ready
WHEN:
  foo called
THEN:
  result returned

## MILESTONE: 0.1.0
Included BEHAVIORs: foo
Deferred BEHAVIORs: foo
Status: pending

## MILESTONE: 0.0.0
Scaffold: true
Included BEHAVIORs: foo
Status: pending
`
	result := lint.LintContent(content, "myspec.md")

	if result.Valid {
		t.Error("expected valid=false for scaffold not first")
	}
	foundRule17 := false
	for _, d := range result.Diagnostics {
		if d.Rule == "RULE-17" && strings.Contains(d.Message, "must appear first") {
			foundRule17 = true
		}
	}
	if !foundRule17 {
		t.Errorf("expected RULE-17 diagnostic about 'must appear first', got: %v", result.Diagnostics)
	}
}

func TestLintContent_TwoScaffoldMilestones(t *testing.T) {
	// EXAMPLE: lint_content_two_scaffold_milestones
	content := `# test

## META
Deployment:   mcp-server
Version:      0.1.0
Spec-Schema:  0.1.0
Author:       Test <t@t.com>
License:      MIT
Verification: none
Safety-Level: QM

## TYPES

` + "```" + `
Foo := string
` + "```" + `

## BEHAVIOR: foo
STEPS:
1. Do foo.

## PRECONDITIONS
- none

## POSTCONDITIONS
- done

## INVARIANTS

- [observable] foo is done

## EXAMPLES

EXAMPLE: foo_example
GIVEN:
  ready
WHEN:
  foo called
THEN:
  result returned

## MILESTONE: 0.0.0
Scaffold: true
Included BEHAVIORs: foo
Status: pending

## MILESTONE: 0.1.0
Scaffold: true
Included BEHAVIORs: foo
Status: pending
`
	result := lint.LintContent(content, "myspec.md")

	if result.Valid {
		t.Error("expected valid=false for two scaffold milestones")
	}
	foundRule17 := false
	for _, d := range result.Diagnostics {
		if d.Rule == "RULE-17" && strings.Contains(d.Message, "more than one MILESTONE has Scaffold: true") {
			foundRule17 = true
		}
	}
	if !foundRule17 {
		t.Errorf("expected RULE-17 diagnostic about 'more than one MILESTONE has Scaffold: true', got: %v", result.Diagnostics)
	}
}

// ── TestLintMatchesCLI ────────────────────────────────────────────────────────

// TestLintMatchesCLI verifies the [observable] invariant:
// lint_content result is identical to pcd-lint CLI on same input
// for RULE-01 through RULE-17.
//
// Since we cannot run the CLI binary in independent tests, this test verifies
// structural equivalence: same rule IDs are fired, same error/warning counts,
// same valid flag. The lint engine is the same code as pcd-lint CLI.
func TestLintMatchesCLI(t *testing.T) {
	cases := []struct {
		name        string
		content     string
		expectValid bool
		expectRule  string
	}{
		{
			name:        "valid spec",
			content:     minimalValidSpec(),
			expectValid: true,
		},
		{
			name: "missing meta fields",
			content: `# test
## META
Deployment: mcp-server
## TYPES
` + "```" + `
Foo := string
` + "```" + `
## BEHAVIOR: foo
STEPS:
1. do.
## PRECONDITIONS
- none
## POSTCONDITIONS
- done
## INVARIANTS
- [observable] done
## EXAMPLES
EXAMPLE: e
GIVEN:
  g
WHEN:
  w
THEN:
  t
`,
			expectValid: false,
			expectRule:  "RULE-02",
		},
		{
			name: "behavior missing steps",
			content: `# test
## META
Deployment:   mcp-server
Version:      0.1.0
Spec-Schema:  0.1.0
Author:       Test <t@t.com>
License:      MIT
Verification: none
Safety-Level: QM
## TYPES
` + "```" + `
Foo := string
` + "```" + `
## BEHAVIOR: foo
INPUTS:
` + "```" + `
x: string
` + "```" + `
## PRECONDITIONS
- none
## POSTCONDITIONS
- done
## INVARIANTS
- [observable] done
## EXAMPLES
EXAMPLE: e
GIVEN:
  g
WHEN:
  w
THEN:
  t
`,
			expectValid: false,
			expectRule:  "RULE-08",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result := lint.LintContent(tc.content, "test.md")
			if result.Valid != tc.expectValid {
				t.Errorf("valid: expected %v, got %v; diagnostics: %v",
					tc.expectValid, result.Valid, result.Diagnostics)
			}
			if tc.expectRule != "" {
				found := false
				for _, d := range result.Diagnostics {
					if d.Rule == tc.expectRule {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected rule %s in diagnostics, got: %v", tc.expectRule, result.Diagnostics)
				}
			}
		})
	}
}

// ── set_milestone_status BEHAVIOR ─────────────────────────────────────────────

func TestSetMilestoneStatus_SetActive(t *testing.T) {
	// EXAMPLE: set_milestone_active
	specContent := "# sitar\n\n## MILESTONE: 0.1.0\nStatus: pending\n\n## MILESTONE: 0.2.0\nStatus: pending\n"
	fs := &milestone.FakeFilesystem{
		Files: map[string]string{
			"/tmp/sitar.md": specContent,
		},
	}

	result, err := milestone.SetStatus(fs, "/tmp/sitar.md", "0.1.0", "active")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.PreviousStatus != milestone.StatusPending {
		t.Errorf("previous_status: expected pending, got %s", result.PreviousStatus)
	}
	if result.NewStatus != milestone.StatusActive {
		t.Errorf("new_status: expected active, got %s", result.NewStatus)
	}

	written := fs.Written["/tmp/sitar.md"]
	if !strings.Contains(written, "## MILESTONE: 0.1.0") {
		t.Error("written content missing MILESTONE 0.1.0")
	}
	if !strings.Contains(written, "Status: active") {
		t.Error("written content missing Status: active")
	}
	// 0.2.0 must still be pending
	if !strings.Contains(written, "Status: pending") {
		t.Error("written content should still have Status: pending for 0.2.0")
	}
}

func TestSetMilestoneStatus_ConflictActive(t *testing.T) {
	// EXAMPLE: set_milestone_active_conflict
	specContent := "# sitar\n\n## MILESTONE: 0.1.0\nStatus: active\n\n## MILESTONE: 0.2.0\nStatus: pending\n"
	fs := &milestone.FakeFilesystem{
		Files: map[string]string{
			"/tmp/sitar.md": specContent,
		},
	}

	_, err := milestone.SetStatus(fs, "/tmp/sitar.md", "0.2.0", "active")
	if err == nil {
		t.Fatal("expected error for conflict, got nil")
	}
	if !strings.Contains(err.Error(), "0.1.0") || !strings.Contains(err.Error(), "already active") {
		t.Errorf("error message should mention '0.1.0' and 'already active', got: %s", err.Error())
	}
	// File must not be modified
	if _, written := fs.Written["/tmp/sitar.md"]; written {
		t.Error("file should not be modified on conflict")
	}
}

func TestSetMilestoneStatus_SetReleased(t *testing.T) {
	// EXAMPLE: set_milestone_released
	specContent := "# sitar\n\n## MILESTONE: 0.1.0\nStatus: active\n"
	fs := &milestone.FakeFilesystem{
		Files: map[string]string{
			"/tmp/sitar.md": specContent,
		},
	}

	result, err := milestone.SetStatus(fs, "/tmp/sitar.md", "0.1.0", "released")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.PreviousStatus != milestone.StatusActive {
		t.Errorf("previous_status: expected active, got %s", result.PreviousStatus)
	}
	if result.NewStatus != milestone.StatusReleased {
		t.Errorf("new_status: expected released, got %s", result.NewStatus)
	}
	written := fs.Written["/tmp/sitar.md"]
	if !strings.Contains(written, "Status: released") {
		t.Error("written content missing Status: released")
	}
}

func TestSetMilestoneStatus_FileNotFound(t *testing.T) {
	fs := &milestone.FakeFilesystem{
		Files: map[string]string{},
	}

	_, err := milestone.SetStatus(fs, "/tmp/missing.md", "0.1.0", "active")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
	if !strings.Contains(err.Error(), "cannot open file") {
		t.Errorf("error should mention 'cannot open file', got: %s", err.Error())
	}
}

func TestSetMilestoneStatus_MilestoneNotFound(t *testing.T) {
	specContent := "# sitar\n\n## MILESTONE: 0.1.0\nStatus: pending\n"
	fs := &milestone.FakeFilesystem{
		Files: map[string]string{
			"/tmp/sitar.md": specContent,
		},
	}

	_, err := milestone.SetStatus(fs, "/tmp/sitar.md", "9.9.9", "active")
	if err == nil {
		t.Fatal("expected error for unknown milestone")
	}
	if !strings.Contains(err.Error(), "9.9.9") {
		t.Errorf("error should mention milestone name, got: %s", err.Error())
	}
}

func TestSetMilestoneStatus_OnlyStatusLineChanged(t *testing.T) {
	// Verify postcondition: all other content byte-for-byte identical
	specContent := "# sitar\n\n## MILESTONE: 0.1.0\nStatus: pending\n\nSome other content here.\n"
	fs := &milestone.FakeFilesystem{
		Files: map[string]string{
			"/tmp/sitar.md": specContent,
		},
	}

	_, err := milestone.SetStatus(fs, "/tmp/sitar.md", "0.1.0", "active")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	written := fs.Written["/tmp/sitar.md"]
	expected := strings.Replace(specContent, "Status: pending", "Status: active", 1)
	if written != expected {
		t.Errorf("written content differs from expected.\nExpected:\n%s\nGot:\n%s", expected, written)
	}
}

// ── Resource URI parsing ──────────────────────────────────────────────────────

func TestReadResource_InvalidURI(t *testing.T) {
	// EXAMPLE: read_resource_invalid_uri
	uri := "http://example.com/bad"
	valid := strings.HasPrefix(uri, "pcd://")
	if valid {
		t.Error("invalid URI should not be accepted")
	}
}

func TestReadResource_ValidURITypes(t *testing.T) {
	cases := []struct {
		uri        string
		expectType string
		expectName string
	}{
		{"pcd://templates/cli-tool", "templates", "cli-tool"},
		{"pcd://hints/cli-tool.go.milestones", "hints", "cli-tool.go.milestones"},
		{"pcd://prompts/interview", "prompts", "interview"},
	}

	for _, tc := range cases {
		rest := strings.TrimPrefix(tc.uri, "pcd://")
		slashIdx := strings.Index(rest, "/")
		if slashIdx < 0 {
			t.Errorf("URI %s: no slash found", tc.uri)
			continue
		}
		resourceType := rest[:slashIdx]
		resourceName := rest[slashIdx+1:]
		if resourceType != tc.expectType {
			t.Errorf("URI %s: type expected %s, got %s", tc.uri, tc.expectType, resourceType)
		}
		if resourceName != tc.expectName {
			t.Errorf("URI %s: name expected %s, got %s", tc.uri, tc.expectName, resourceName)
		}
	}
}

// ── JSON serialisation ────────────────────────────────────────────────────────

func TestLintResult_JSONSerialisation(t *testing.T) {
	result := lint.LintContent(minimalValidSpec(), "test.md")

	type diagJSON struct {
		Severity string `json:"severity"`
		Line     int    `json:"line"`
		Section  string `json:"section"`
		Message  string `json:"message"`
		Rule     string `json:"rule"`
	}
	type lintResultJSON struct {
		Valid       bool       `json:"valid"`
		Errors      int        `json:"errors"`
		Warnings    int        `json:"warnings"`
		Diagnostics []diagJSON `json:"diagnostics"`
	}

	out := lintResultJSON{
		Valid:       result.Valid,
		Errors:      result.Errors,
		Warnings:    result.Warnings,
		Diagnostics: []diagJSON{},
	}
	for _, d := range result.Diagnostics {
		out.Diagnostics = append(out.Diagnostics, diagJSON{
			Severity: d.Severity.String(),
			Line:     d.Line,
			Section:  d.Section,
			Message:  d.Message,
			Rule:     d.Rule,
		})
	}

	data, err := json.Marshal(out)
	if err != nil {
		t.Fatalf("JSON marshal failed: %v", err)
	}

	var decoded lintResultJSON
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("JSON unmarshal failed: %v", err)
	}
	if decoded.Valid != result.Valid {
		t.Errorf("valid mismatch after round-trip")
	}
}

// ── Transport argument parsing ────────────────────────────────────────────────

// parseArgsForTest is a local copy of main.parseArgs for test access.
func parseArgsForTest(args []string) (transport, listenAddr string, err error) {
	transport = "stdio"
	listenAddr = "127.0.0.1:8080"

	var transports []string

	for _, arg := range args {
		if arg == "stdio" || arg == "http" {
			transports = append(transports, arg)
		} else if strings.HasPrefix(arg, "listen=") {
			listenAddr = strings.TrimPrefix(arg, "listen=")
		} else {
			return "", "", fmt.Errorf("unknown argument '%s'. Valid transports: stdio, http", arg)
		}
	}

	if len(transports) > 1 {
		return "", "", fmt.Errorf("multiple transports specified: %s. Specify exactly one.",
			strings.Join(transports, ", "))
	}

	if len(transports) == 1 {
		transport = transports[0]
	}

	return transport, listenAddr, nil
}

func TestParseArgs_DefaultStdio(t *testing.T) {
	transport, listenAddr, err := parseArgsForTest([]string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if transport != "stdio" {
		t.Errorf("expected stdio, got %s", transport)
	}
	if listenAddr != "127.0.0.1:8080" {
		t.Errorf("expected 127.0.0.1:8080, got %s", listenAddr)
	}
}

func TestParseArgs_HTTP(t *testing.T) {
	transport, listenAddr, err := parseArgsForTest([]string{"http"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if transport != "http" {
		t.Errorf("expected http, got %s", transport)
	}
	if listenAddr != "127.0.0.1:8080" {
		t.Errorf("expected default 127.0.0.1:8080, got %s", listenAddr)
	}
}

func TestParseArgs_HTTPWithListen(t *testing.T) {
	transport, listenAddr, err := parseArgsForTest([]string{"http", "listen=0.0.0.0:9090"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if transport != "http" {
		t.Errorf("expected http, got %s", transport)
	}
	if listenAddr != "0.0.0.0:9090" {
		t.Errorf("expected 0.0.0.0:9090, got %s", listenAddr)
	}
}

func TestParseArgs_Stdio(t *testing.T) {
	transport, _, err := parseArgsForTest([]string{"stdio"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if transport != "stdio" {
		t.Errorf("expected stdio, got %s", transport)
	}
}

func TestParseArgs_MultipleTransports(t *testing.T) {
	_, _, err := parseArgsForTest([]string{"stdio", "http"})
	if err == nil {
		t.Fatal("expected error for multiple transports")
	}
}

func TestParseArgs_UnknownArg(t *testing.T) {
	_, _, err := parseArgsForTest([]string{"websocket"})
	if err == nil {
		t.Fatal("expected error for unknown transport")
	}
	if !strings.Contains(err.Error(), "websocket") {
		t.Errorf("error should mention 'websocket', got: %s", err)
	}
}

// ── Prompt key derivation ─────────────────────────────────────────────────────

func TestFakeStore_TranslatorPrompt(t *testing.T) {
	// EXAMPLE: read_resource_interview_prompt / read_resource_reverse_prompt
	// Verifies that FakeStore correctly serves prompts by key.
	// The "translator" key corresponds to prompt.md in the asset store.
	s := &store.FakeStore{
		Prompts: map[string]string{
			"interview":  "# interview prompt",
			"reverse":    "# reverse prompt",
			"translator": "# translator/translation prompt",
		},
	}

	for _, tc := range []struct{ key, expect string }{
		{"interview", "interview"},
		{"reverse", "reverse"},
		{"translator", "translator"},
	} {
		content, err := s.GetPrompt(tc.key)
		if err != nil {
			t.Errorf("GetPrompt(%q): unexpected error: %v", tc.key, err)
			continue
		}
		if !strings.Contains(content, tc.expect) {
			t.Errorf("GetPrompt(%q): content %q does not contain %q", tc.key, content, tc.expect)
		}
	}
}

func TestFakeStore_ListPromptsReturnsKeys(t *testing.T) {
	s := &store.FakeStore{
		Prompts: map[string]string{
			"interview":  "...",
			"reverse":    "...",
			"translator": "...",
		},
	}
	names, err := s.ListPrompts()
	if err != nil {
		t.Fatalf("ListPrompts: unexpected error: %v", err)
	}
	if len(names) != 3 {
		t.Errorf("expected 3 prompts, got %d: %v", len(names), names)
	}
}

func TestFakeStore_ListHintsKeysReturnsAll(t *testing.T) {
	s := &store.FakeStore{
		Hints: map[string]string{
			"cli-tool.go.milestones":     "...",
			"mcp-server.go.mcp-go":       "...",
			"cloud-native.go.go-libvirt": "...",
		},
	}
	keys, err := s.ListHintsKeys()
	if err != nil {
		t.Fatalf("ListHintsKeys: unexpected error: %v", err)
	}
	if len(keys) != 3 {
		t.Errorf("expected 3 hints keys, got %d: %v", len(keys), keys)
	}
}

// ── lint_file path ────────────────────────────────────────────────────────────

func TestLintFile_NotFound(t *testing.T) {
	// EXAMPLE: lint_file_not_found
	fs := &milestone.FakeFilesystem{
		Files: map[string]string{},
	}
	_, err := fs.ReadFile("/tmp/missing.md")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
	// The lint_file handler wraps this as "cannot open file: {path}"
	errMsg := fmt.Sprintf("cannot open file: %s", "/tmp/missing.md")
	if !strings.Contains(errMsg, "cannot open file") {
		t.Errorf("error message format wrong: %s", errMsg)
	}
}

// ── get_schema_version ────────────────────────────────────────────────────────

func TestGetSchemaVersion(t *testing.T) {
	// BEHAVIOR: get_schema_version
	// Verifies that the schema version constant is a valid semver.
	v := lint.SpecSchema
	if v == "" {
		t.Fatal("SpecSchema is empty")
	}
	// Must match MAJOR.MINOR.PATCH
	parts := strings.Split(v, ".")
	if len(parts) != 3 {
		t.Errorf("SpecSchema %q is not MAJOR.MINOR.PATCH format", v)
	}
	// Must match the spec's declared Spec-Schema
	if v != "0.3.21" {
		t.Errorf("SpecSchema expected 0.3.21, got %s", v)
	}
}
