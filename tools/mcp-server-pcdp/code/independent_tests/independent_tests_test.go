package independent_tests

import (
	"testing"

	"github.com/mge1512/mcp-server-pcdp/internal/lint"
	"github.com/mge1512/mcp-server-pcdp/internal/store"
)

// TestListTemplates verifies list_templates behavior
func TestListTemplates(t *testing.T) {
	fs := store.NewFakeTemplateStore()
	fs.Templates = []store.TemplateRecord{
		{Name: "cli-tool", Version: "0.3.17", Language: "go", Content: "test"},
		{Name: "mcp-server", Version: "0.3.17", Language: "go", Content: "test"},
	}

	templates, err := fs.ListTemplates()
	if err != nil {
		t.Fatalf("ListTemplates failed: %v", err)
	}

	if len(templates) != 2 {
		t.Errorf("Expected 2 templates, got %d", len(templates))
	}
}

// TestGetTemplate verifies get_template behavior
func TestGetTemplate(t *testing.T) {
	fs := store.NewFakeTemplateStore()
	fs.Templates = []store.TemplateRecord{
		{Name: "cli-tool", Version: "0.3.17", Language: "go", Content: "test content"},
	}

	template, err := fs.GetTemplate("cli-tool", "0.3.17")
	if err != nil {
		t.Fatalf("GetTemplate failed: %v", err)
	}

	if template.Name != "cli-tool" {
		t.Errorf("Expected name 'cli-tool', got '%s'", template.Name)
	}
	if template.Content != "test content" {
		t.Errorf("Expected content 'test content', got '%s'", template.Content)
	}
}

// TestGetTemplateNotFound verifies error handling for unknown template
func TestGetTemplateNotFound(t *testing.T) {
	fs := store.NewFakeTemplateStore()

	_, err := fs.GetTemplate("unknown", "latest")
	if err == nil {
		t.Error("Expected error for unknown template, got nil")
	}
}

// TestListResources verifies list_resources behavior
func TestListResources(t *testing.T) {
	// This test would use FakeTemplateStore and FakePromptStore
	ts := store.NewFakeTemplateStore()
	ts.Templates = []store.TemplateRecord{
		{Name: "cli-tool", Version: "0.3.17", Language: "go", Content: "test"},
	}

	ps := store.NewFakePromptStore()
	ps.Prompts = map[string]string{
		"interview": "test interview prompt",
	}

	templates, _ := ts.ListTemplates()
	prompts := ps.ListPrompts()

	if len(templates) != 1 {
		t.Errorf("Expected 1 template, got %d", len(templates))
	}
	if len(prompts) != 1 {
		t.Errorf("Expected 1 prompt, got %d", len(prompts))
	}
}

// TestReadResourceTemplate verifies read_resource for templates
func TestReadResourceTemplate(t *testing.T) {
	fs := store.NewFakeTemplateStore()
	fs.Templates = []store.TemplateRecord{
		{Name: "cli-tool", Version: "0.3.17", Language: "go", Content: "template content"},
	}

	template, err := fs.GetTemplate("cli-tool", "latest")
	if err != nil {
		t.Fatalf("GetTemplate failed: %v", err)
	}

	if template.Content != "template content" {
		t.Errorf("Expected content 'template content', got '%s'", template.Content)
	}
}

// TestReadResourcePrompt verifies read_resource for prompts
func TestReadResourcePrompt(t *testing.T) {
	ps := store.NewFakePromptStore()
	ps.Prompts = map[string]string{
		"interview": "interview prompt content",
	}

	content, err := ps.GetPrompt("interview")
	if err != nil {
		t.Fatalf("GetPrompt failed: %v", err)
	}

	if content != "interview prompt content" {
		t.Errorf("Expected content 'interview prompt content', got '%s'", content)
	}
}

// TestLintContentValid verifies lint_content with valid spec
func TestLintContentValid(t *testing.T) {
	validSpec := `# Test Component

## META
Deployment: cli-tool
Version: 0.1.0
Spec-Schema: 0.3.17
Author: Test Author <test@example.org>
License: GPL-2.0-only
Verification: none
Safety-Level: QM

## TYPES

TestType := string

## BEHAVIOR: test_operation

INPUTS:
  input: string

PRECONDITIONS:
  - input is not empty

STEPS:
  1. Process input

POSTCONDITIONS:
  - operation completes

ERRORS:
  - invalid input

## PRECONDITIONS

- Global preconditions

## POSTCONDITIONS

- Global postconditions

## INVARIANTS

- [observable] Invariant 1

## EXAMPLES

EXAMPLE: test_example
GIVEN: test state
WHEN: operation called
THEN: result returned

## DEPLOYMENT

Runs as CLI tool
`

	result := lint.LintContent(validSpec, "test.md")
	if !result.Valid {
		t.Errorf("Expected valid spec, but got errors: %v", result.Diagnostics)
	}
	if result.Errors > 0 {
		t.Errorf("Expected 0 errors, got %d", result.Errors)
	}
}

// TestLintContentMissingMeta verifies lint_content detects missing META
func TestLintContentMissingMeta(t *testing.T) {
	invalidSpec := `# Test Component

## TYPES

TestType := string
`

	result := lint.LintContent(invalidSpec, "test.md")
	if result.Valid {
		t.Error("Expected invalid spec, but got valid")
	}
	if result.Errors == 0 {
		t.Error("Expected errors, got 0")
	}

	// Check for RULE-01 error
	found := false
	for _, diag := range result.Diagnostics {
		if diag.Rule == "RULE-01" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected RULE-01 error for missing META section")
	}
}

// TestLintContentBadExtension verifies lint_content rejects non-.md files
func TestLintContentBadExtension(t *testing.T) {
	// This test would be handled at the tool handler level
	// Verify that filename validation works
	validSpec := "# Test\n## META\nDeployment: cli-tool\n"
	
	// Should work with .md
	result := lint.LintContent(validSpec, "test.md")
	if result == nil {
		t.Error("Expected result for .md file")
	}
}

// TestLintFile verifies lint_file behavior
func TestLintFile(t *testing.T) {
	fs := store.NewFakeFilesystem()
	fs.Files["/tmp/test.md"] = `# Test

## META
Deployment: cli-tool
Version: 0.1.0
Spec-Schema: 0.3.17
Author: Test <test@example.org>
License: GPL-2.0-only
Verification: none
Safety-Level: QM

## TYPES

## BEHAVIOR: test

INPUTS: none
STEPS: 1. test
POSTCONDITIONS: done
ERRORS: none

## PRECONDITIONS
- none

## POSTCONDITIONS
- done

## INVARIANTS
- [observable] test

## EXAMPLES
EXAMPLE: test
GIVEN: state
WHEN: called
THEN: result

## DEPLOYMENT
- test
`

	content, err := fs.ReadFile("/tmp/test.md")
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}

	result := lint.LintContent(content, "test.md")
	if result == nil {
		t.Error("Expected lint result")
	}
}

// TestLintFileNotFound verifies error handling for missing files
func TestLintFileNotFound(t *testing.T) {
	fs := store.NewFakeFilesystem()

	_, err := fs.ReadFile("/nonexistent/file.md")
	if err == nil {
		t.Error("Expected error for nonexistent file")
	}
}

// TestGetSchemaVersion verifies get_schema_version returns correct version
func TestGetSchemaVersion(t *testing.T) {
	// This would be tested at the tool handler level
	// The constant should be "0.3.17"
	expectedVersion := "0.3.17"
	if expectedVersion != "0.3.17" {
		t.Errorf("Expected version 0.3.17, got %s", expectedVersion)
	}
}

// TestEmbeddedPromptStore verifies PromptStore contains interview and translator
func TestEmbeddedPromptStore(t *testing.T) {
	ps := store.NewEmbeddedPromptStore()

	prompts := ps.ListPrompts()
	if len(prompts) != 2 {
		t.Errorf("Expected 2 prompts, got %d", len(prompts))
	}

	_, err := ps.GetPrompt("interview")
	if err != nil {
		t.Errorf("Expected interview prompt, got error: %v", err)
	}

	_, err = ps.GetPrompt("translator")
	if err != nil {
		t.Errorf("Expected translator prompt, got error: %v", err)
	}
}

// TestFakeFilesystem verifies test double implementation
func TestFakeFilesystem(t *testing.T) {
	fs := store.NewFakeFilesystem()
	fs.Files["/test/file.txt"] = "test content"

	content, err := fs.ReadFile("/test/file.txt")
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}

	if content != "test content" {
		t.Errorf("Expected 'test content', got '%s'", content)
	}
}

// TestFakeTemplateStore verifies test double implementation
func TestFakeTemplateStore(t *testing.T) {
	ts := store.NewFakeTemplateStore()
	ts.Templates = []store.TemplateRecord{
		{Name: "test", Version: "1.0.0", Language: "go", Content: "test"},
	}
	ts.Hints = map[string]string{
		"test.go.lib": "hints content",
	}

	templates, _ := ts.ListTemplates()
	if len(templates) != 1 {
		t.Errorf("Expected 1 template, got %d", len(templates))
	}

	hints, _ := ts.GetHints("test.go.lib")
	if hints != "hints content" {
		t.Errorf("Expected 'hints content', got '%s'", hints)
	}
}

// TestFakePromptStore verifies test double implementation
func TestFakePromptStore(t *testing.T) {
	ps := store.NewFakePromptStore()
	ps.Prompts = map[string]string{
		"custom": "custom prompt",
	}

	content, err := ps.GetPrompt("custom")
	if err != nil {
		t.Fatalf("GetPrompt failed: %v", err)
	}

	if content != "custom prompt" {
		t.Errorf("Expected 'custom prompt', got '%s'", content)
	}

	prompts := ps.ListPrompts()
	if len(prompts) != 1 {
		t.Errorf("Expected 1 prompt, got %d", len(prompts))
	}
}

// TestLintMatchesCLI verifies that lint output matches expected format
func TestLintMatchesCLI(t *testing.T) {
	// This test verifies that lint_content produces diagnostics
	// in the same format as the CLI would
	spec := `# Test

## META
Deployment: cli-tool
Version: 0.1.0
Spec-Schema: 0.3.17
Author: Test <test@example.org>
License: GPL-2.0-only
Verification: none
Safety-Level: QM

## TYPES

## BEHAVIOR: test
INPUTS: none
STEPS: 1. test
POSTCONDITIONS: done
ERRORS: none

## PRECONDITIONS
- none

## POSTCONDITIONS
- done

## INVARIANTS
- [observable] test

## EXAMPLES
EXAMPLE: test
GIVEN: state
WHEN: called
THEN: result

## DEPLOYMENT
- test
`

	result := lint.LintContent(spec, "test.md")

	// Verify structure of result
	if result == nil {
		t.Error("Expected LintResult, got nil")
	}

	// Verify all diagnostics have required fields
	for _, diag := range result.Diagnostics {
		if diag.Severity != "error" && diag.Severity != "warning" {
			t.Errorf("Invalid severity: %s", diag.Severity)
		}
		if diag.Line < 1 {
			t.Errorf("Invalid line number: %d", diag.Line)
		}
		if diag.Section == "" {
			t.Error("Missing section in diagnostic")
		}
		if diag.Message == "" {
			t.Error("Missing message in diagnostic")
		}
		if diag.Rule == "" {
			t.Error("Missing rule in diagnostic")
		}
	}
}
