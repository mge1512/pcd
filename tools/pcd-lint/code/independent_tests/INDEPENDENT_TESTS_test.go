// Package independent_tests provides self-contained tests for pcd-lint.
// All tests run without any live external service.
// Tests are organised by EXAMPLE from the pcd-lint specification.
//
// SPDX-License-Identifier: GPL-2.0-only

package independent_tests

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pcd-tools/pcd-lint/internal/lint"
)

// Re-export types for convenience in test functions
type Severity = lint.Severity
type LintResult = lint.LintResult
type Diagnostic = lint.Diagnostic

const SevError = lint.SevError
const SevWarning = lint.SevWarning

func lintSpec(path string, strict bool) lint.LintResult {
	return lint.LintSpec(path, strict)
}

func formatSummary(result lint.LintResult, strict bool) string {
	return lint.FormatSummary(result, strict)
}

func isValidSPDX(expr string) bool {
	return lint.IsValidSPDX(expr)
}

var knownTemplates = lint.KnownTemplates

// ── Helpers ───────────────────────────────────────────────────────────────────

func writeSpec(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "spec.md")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("writeSpec: %v", err)
	}
	return path
}

func minimalValidSpec(metaOverrides string) string {
	return `# test-component

## META
Deployment:   cli-tool
Version:      0.1.0
Spec-Schema:  0.1.0
Author:       Jane Example <jane@example.org>
License:      Apache-2.0
Verification: none
Safety-Level: QM
` + metaOverrides + `

## TYPES

` + "```" + `
Foo := string
` + "```" + `

## BEHAVIOR: do-something
Constraint: required

INPUTS:
` + "```" + `
x: string
` + "```" + `

STEPS:
1. Process x.
2. Return result.

## PRECONDITIONS
- x is non-empty

## POSTCONDITIONS
- result is produced

## INVARIANTS
- [observable] tool is idempotent

## EXAMPLES

EXAMPLE: happy_path
GIVEN:
  x = "hello"
WHEN:
  result = do-something(x)
THEN:
  result = "HELLO"
`
}

// ── EXAMPLE: valid_minimal_spec ────────────────────────────────────────────────

func TestValidMinimalSpec(t *testing.T) {
	path := writeSpec(t, minimalValidSpec(""))
	result := lintSpec(path, false)

	if result.ExitCode != 0 {
		t.Errorf("expected exit_code=0, got %d", result.ExitCode)
		for _, d := range result.Diagnostics {
			t.Logf("diag: %+v", d)
		}
	}
	if len(result.Diagnostics) != 0 {
		t.Errorf("expected no diagnostics, got %d", len(result.Diagnostics))
		for _, d := range result.Diagnostics {
			t.Logf("diag: %+v", d)
		}
	}
}

// ── EXAMPLE: multiple_authors_valid ───────────────────────────────────────────

func TestMultipleAuthorsValid(t *testing.T) {
	spec := `# test-component

## META
Deployment:   cli-tool
Version:      0.1.0
Spec-Schema:  0.1.0
Author:       Jane Example <jane@example.org>
Author:       John Example <john@example.org>
License:      Apache-2.0
Verification: none
Safety-Level: QM

## TYPES

` + "```" + `
Foo := string
` + "```" + `

## BEHAVIOR: do-something

STEPS:
1. Process input.

## PRECONDITIONS
- input is valid

## POSTCONDITIONS
- output produced

## INVARIANTS
- [observable] idempotent

## EXAMPLES

EXAMPLE: happy_path
GIVEN:
  input = "x"
WHEN:
  result = do-something(input)
THEN:
  result = "X"
`
	path := writeSpec(t, spec)
	result := lintSpec(path, false)

	if result.ExitCode != 0 {
		t.Errorf("expected exit_code=0, got %d", result.ExitCode)
		for _, d := range result.Diagnostics {
			t.Logf("diag: %+v", d)
		}
	}
}

// ── EXAMPLE: invalid_spdx_license ─────────────────────────────────────────────

func TestInvalidSPDXLicense(t *testing.T) {
	spec := strings.Replace(minimalValidSpec(""), "License:      Apache-2.0", "License:      MIT License", 1)
	path := writeSpec(t, spec)
	result := lintSpec(path, false)

	if result.ExitCode != 1 {
		t.Errorf("expected exit_code=1, got %d", result.ExitCode)
	}
	found := false
	for _, d := range result.Diagnostics {
		if d.Severity == SevError && strings.Contains(d.Message, "MIT License") && strings.Contains(d.Message, "not a valid SPDX identifier") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected Error about 'MIT License' not being valid SPDX")
		for _, d := range result.Diagnostics {
			t.Logf("diag: %+v", d)
		}
	}
}

// ── EXAMPLE: invalid_version_format ───────────────────────────────────────────

func TestInvalidVersionFormat(t *testing.T) {
	spec := strings.Replace(minimalValidSpec(""), "Version:      0.1.0", "Version:      1.0", 1)
	path := writeSpec(t, spec)
	result := lintSpec(path, false)

	if result.ExitCode != 1 {
		t.Errorf("expected exit_code=1, got %d", result.ExitCode)
	}
	found := false
	for _, d := range result.Diagnostics {
		if d.Severity == SevError && strings.Contains(d.Message, "Version '1.0'") && strings.Contains(d.Message, "not valid semantic versioning") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected Error about Version '1.0' not valid semantic versioning")
		for _, d := range result.Diagnostics {
			t.Logf("diag: %+v", d)
		}
	}
}

// ── EXAMPLE: missing_author ────────────────────────────────────────────────────

func TestMissingAuthor(t *testing.T) {
	spec := strings.Replace(minimalValidSpec(""), "Author:       Jane Example <jane@example.org>\n", "", 1)
	path := writeSpec(t, spec)
	result := lintSpec(path, false)

	if result.ExitCode != 1 {
		t.Errorf("expected exit_code=1, got %d", result.ExitCode)
	}
	found := false
	for _, d := range result.Diagnostics {
		if d.Severity == SevError && strings.Contains(d.Message, "Missing required META field: Author") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected Error about missing Author field")
		for _, d := range result.Diagnostics {
			t.Logf("diag: %+v", d)
		}
	}
}

// ── EXAMPLE: missing_section ──────────────────────────────────────────────────

func TestMissingSection(t *testing.T) {
	spec := minimalValidSpec("")
	spec = strings.Replace(spec, "\n## INVARIANTS\n- [observable] tool is idempotent\n", "\n", 1)
	path := writeSpec(t, spec)
	result := lintSpec(path, false)

	if result.ExitCode != 1 {
		t.Errorf("expected exit_code=1, got %d", result.ExitCode)
	}
	found := false
	for _, d := range result.Diagnostics {
		if d.Severity == SevError && strings.Contains(d.Message, "Missing required section: ## INVARIANTS") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected Error about missing ## INVARIANTS section")
		for _, d := range result.Diagnostics {
			t.Logf("diag: %+v", d)
		}
	}

	summary := formatSummary(result, false)
	if !strings.Contains(summary, "error(s)") {
		t.Errorf("expected summary to contain 'error(s)', got: %s", summary)
	}
}

// ── EXAMPLE: unknown_deployment_template ──────────────────────────────────────

func TestUnknownDeploymentTemplate(t *testing.T) {
	spec := strings.Replace(minimalValidSpec(""), "Deployment:   cli-tool", "Deployment:   serverless", 1)
	path := writeSpec(t, spec)
	result := lintSpec(path, false)

	if result.ExitCode != 1 {
		t.Errorf("expected exit_code=1, got %d", result.ExitCode)
	}
	found := false
	for _, d := range result.Diagnostics {
		if d.Severity == SevError && strings.Contains(d.Message, "Unknown deployment template: 'serverless'") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected Error about unknown deployment template 'serverless'")
		for _, d := range result.Diagnostics {
			t.Logf("diag: %+v", d)
		}
	}
}

// ── EXAMPLE: deprecated_target_field_permissive ────────────────────────────────

func TestDeprecatedTargetFieldPermissive(t *testing.T) {
	spec := strings.Replace(minimalValidSpec(""), "Deployment:   cli-tool", "Deployment:   backend-service\nTarget: Go", 1)
	path := writeSpec(t, spec)
	result := lintSpec(path, false)

	if result.ExitCode != 0 {
		t.Errorf("expected exit_code=0 (permissive), got %d", result.ExitCode)
	}
	found := false
	for _, d := range result.Diagnostics {
		if d.Severity == SevWarning && strings.Contains(d.Message, "deprecated since v0.3.0") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected Warning about deprecated Target field")
		for _, d := range result.Diagnostics {
			t.Logf("diag: %+v", d)
		}
	}
	summary := formatSummary(result, false)
	if !strings.HasPrefix(summary, "✓") {
		t.Errorf("expected summary to start with ✓, got: %s", summary)
	}
	if !strings.Contains(summary, "warning(s)") {
		t.Errorf("expected summary to contain 'warning(s)', got: %s", summary)
	}
}

// ── EXAMPLE: deprecated_target_field_strict ───────────────────────────────────

func TestDeprecatedTargetFieldStrict(t *testing.T) {
	spec := strings.Replace(minimalValidSpec(""), "Deployment:   cli-tool", "Deployment:   backend-service\nTarget: Go", 1)
	path := writeSpec(t, spec)
	result := lintSpec(path, true)

	if result.ExitCode != 1 {
		t.Errorf("expected exit_code=1 (strict), got %d", result.ExitCode)
	}
	summary := formatSummary(result, true)
	if !strings.Contains(summary, "[strict mode]") {
		t.Errorf("expected summary to contain '[strict mode]', got: %s", summary)
	}
	if !strings.HasPrefix(summary, "✗") {
		t.Errorf("expected summary to start with ✗, got: %s", summary)
	}
}

// ── EXAMPLE: enhance_existing_missing_language ────────────────────────────────

func TestEnhanceExistingMissingLanguage(t *testing.T) {
	spec := strings.Replace(minimalValidSpec(""), "Deployment:   cli-tool", "Deployment:   enhance-existing", 1)
	path := writeSpec(t, spec)
	result := lintSpec(path, false)

	if result.ExitCode != 1 {
		t.Errorf("expected exit_code=1, got %d", result.ExitCode)
	}
	found := false
	for _, d := range result.Diagnostics {
		if d.Severity == SevError && strings.Contains(d.Message, "requires META field 'Language'") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected Error about missing Language field for enhance-existing")
		for _, d := range result.Diagnostics {
			t.Logf("diag: %+v", d)
		}
	}
}

// ── EXAMPLE: empty_given_block_permissive ─────────────────────────────────────

func TestEmptyGivenBlockPermissive(t *testing.T) {
	spec := `# test-component

## META
Deployment:   cli-tool
Version:      0.1.0
Spec-Schema:  0.1.0
Author:       Jane Example <jane@example.org>
License:      Apache-2.0
Verification: none
Safety-Level: QM

## TYPES

` + "```" + `
Foo := string
` + "```" + `

## BEHAVIOR: do-something

STEPS:
1. Do it.

## PRECONDITIONS
- valid

## POSTCONDITIONS
- done

## INVARIANTS
- [observable] idempotent

## EXAMPLES

EXAMPLE: foo
GIVEN:

WHEN:
  result = foo()
THEN:
  result = Ok
`
	path := writeSpec(t, spec)
	result := lintSpec(path, false)

	if result.ExitCode != 0 {
		t.Errorf("expected exit_code=0 (permissive), got %d", result.ExitCode)
	}
	found := false
	for _, d := range result.Diagnostics {
		if d.Severity == SevWarning && strings.Contains(d.Message, "empty GIVEN block") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected Warning about empty GIVEN block")
		for _, d := range result.Diagnostics {
			t.Logf("diag: %+v", d)
		}
	}
	summary := formatSummary(result, false)
	if !strings.HasPrefix(summary, "✓") {
		t.Errorf("expected ✓ summary, got: %s", summary)
	}
}

// ── EXAMPLE: multiple_errors ──────────────────────────────────────────────────

func TestMultipleErrors(t *testing.T) {
	spec := `# test-component

## META
Version:      0.1.0
Spec-Schema:  0.1.0
Author:       Jane Example <jane@example.org>
License:      Apache-2.0
Verification: none
Safety-Level: QM

## TYPES

` + "```" + `
Foo := string
` + "```" + `

## BEHAVIOR: do-something

STEPS:
1. Do it.

## PRECONDITIONS
- valid

## POSTCONDITIONS
- done
`
	path := writeSpec(t, spec)
	result := lintSpec(path, false)

	if result.ExitCode != 1 {
		t.Errorf("expected exit_code=1, got %d", result.ExitCode)
	}

	msgs := map[string]bool{}
	for _, d := range result.Diagnostics {
		msgs[d.Message] = true
	}

	expectedMsgs := []string{
		"Missing required section: ## INVARIANTS",
		"Missing required section: ## EXAMPLES",
		"Missing required META field: Deployment",
	}
	for _, em := range expectedMsgs {
		found := false
		for msg := range msgs {
			if strings.Contains(msg, em) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected diagnostic containing: %q", em)
		}
	}

	summary := formatSummary(result, false)
	if !strings.HasPrefix(summary, "✗") {
		t.Errorf("expected ✗ summary, got: %s", summary)
	}
}

// ── EXAMPLE: file_not_found ───────────────────────────────────────────────────

func TestFileNotFound(t *testing.T) {
	path := "/tmp/pcd-lint-test-nonexistent-file-xyz.md"
	os.Remove(path)

	_, err := os.Stat(path)
	if !os.IsNotExist(err) {
		t.Skip("test file unexpectedly exists")
	}
	if err == nil {
		t.Errorf("expected file to not exist")
	}
}

// ── EXAMPLE: non_md_extension ─────────────────────────────────────────────────

func TestNonMdExtension(t *testing.T) {
	path := "myspec.txt"
	hasMd := strings.HasSuffix(path, ".md")
	if hasMd {
		t.Errorf("expected .txt to not have .md suffix")
	}
}

// ── EXAMPLE: behavior_internal_recognised ─────────────────────────────────────

func TestBehaviorInternalRecognised(t *testing.T) {
	spec := `# test-component

## META
Deployment:   cli-tool
Version:      0.1.0
Spec-Schema:  0.1.0
Author:       Jane Example <jane@example.org>
License:      Apache-2.0
Verification: none
Safety-Level: QM

## TYPES

` + "```" + `
Foo := string
` + "```" + `

## BEHAVIOR: lint

STEPS:
1. Validate file.
2. Return result.

## BEHAVIOR/INTERNAL: precedence-resolution

STEPS:
1. Merge layers.
2. Return resolved.

## PRECONDITIONS
- file exists

## POSTCONDITIONS
- result produced

## INVARIANTS
- [observable] idempotent

## EXAMPLES

EXAMPLE: happy_path
GIVEN:
  file is valid
WHEN:
  result = lint(file)
THEN:
  exit_code = 0
`
	path := writeSpec(t, spec)
	result := lintSpec(path, false)

	if result.ExitCode != 0 {
		t.Errorf("expected exit_code=0, got %d", result.ExitCode)
		for _, d := range result.Diagnostics {
			t.Logf("diag: %+v", d)
		}
	}
}

// ── EXAMPLE: behavior_internal_unknown_variant ────────────────────────────────

func TestBehaviorInternalUnknownVariant(t *testing.T) {
	spec := `# test-component

## META
Deployment:   cli-tool
Version:      0.1.0
Spec-Schema:  0.1.0
Author:       Jane Example <jane@example.org>
License:      Apache-2.0
Verification: none
Safety-Level: QM

## TYPES

` + "```" + `
Foo := string
` + "```" + `

## BEHAVIOR/PRIVATE: foo

STEPS:
1. Do foo.

## PRECONDITIONS
- valid

## POSTCONDITIONS
- done

## INVARIANTS
- [observable] idempotent

## EXAMPLES

EXAMPLE: test
GIVEN:
  foo condition
WHEN:
  foo()
THEN:
  result = ok
`
	path := writeSpec(t, spec)
	result := lintSpec(path, false)

	if result.ExitCode != 1 {
		t.Errorf("expected exit_code=1, got %d", result.ExitCode)
	}
	found := false
	for _, d := range result.Diagnostics {
		if d.Severity == SevError && strings.Contains(d.Message, "Missing required section: ## BEHAVIOR") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected Error about missing ## BEHAVIOR section")
		for _, d := range result.Diagnostics {
			t.Logf("diag: %+v", d)
		}
	}
}

// ── EXAMPLE: multi_pass_example_valid ─────────────────────────────────────────

func TestMultiPassExampleValid(t *testing.T) {
	spec := `# test-component

## META
Deployment:   cli-tool
Version:      0.1.0
Spec-Schema:  0.1.0
Author:       Jane Example <jane@example.org>
License:      Apache-2.0
Verification: none
Safety-Level: QM

## TYPES

` + "```" + `
Foo := string
` + "```" + `

## BEHAVIOR: reconcile

STEPS:
1. Check state.
2. On mismatch, requeue after delay.

## PRECONDITIONS
- domain exists

## POSTCONDITIONS
- state reconciled

## INVARIANTS
- [observable] idempotent

## EXAMPLES

EXAMPLE: reconcile_graceful_stop
GIVEN:
  VM "testvm-01", spec.desiredState = Stopped
  Domain is Running
WHEN:  reconcile runs (pass 1)
THEN:
  domain.Shutdown() is called
  result = RequeueAfter(10s)
WHEN:  reconcile runs (pass 2); domain is Shutoff
THEN:
  status.phase = Stopped
  result = RequeueAfter(60s)
`
	path := writeSpec(t, spec)
	result := lintSpec(path, false)

	if result.ExitCode != 0 {
		t.Errorf("expected exit_code=0 for multi-pass example, got %d", result.ExitCode)
		for _, d := range result.Diagnostics {
			t.Logf("diag: %+v", d)
		}
	}
}

// ── EXAMPLE: behavior_missing_steps ───────────────────────────────────────────

func TestBehaviorMissingSteps(t *testing.T) {
	spec := `# test-component

## META
Deployment:   cli-tool
Version:      0.1.0
Spec-Schema:  0.1.0
Author:       Jane Example <jane@example.org>
License:      Apache-2.0
Verification: none
Safety-Level: QM

## TYPES

` + "```" + `
Foo := string
` + "```" + `

## BEHAVIOR: do-something
PRECONDITIONS:
  - input is valid
POSTCONDITIONS:
  - output is produced

## PRECONDITIONS
- input is valid

## POSTCONDITIONS
- output produced

## INVARIANTS
- [observable] idempotent

## EXAMPLES

EXAMPLE: test
GIVEN:
  input valid
WHEN:
  do-something(input)
THEN:
  result = ok
`
	path := writeSpec(t, spec)
	result := lintSpec(path, false)

	if result.ExitCode != 1 {
		t.Errorf("expected exit_code=1, got %d", result.ExitCode)
	}
	found := false
	for _, d := range result.Diagnostics {
		if d.Severity == SevError && strings.Contains(d.Message, "missing required STEPS: block") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected Error about missing STEPS: block")
		for _, d := range result.Diagnostics {
			t.Logf("diag: %+v", d)
		}
	}
}

// ── EXAMPLE: invariant_missing_tag_warning ────────────────────────────────────

func TestInvariantMissingTagWarning(t *testing.T) {
	spec := `# test-component

## META
Deployment:   cli-tool
Version:      0.1.0
Spec-Schema:  0.1.0
Author:       Jane Example <jane@example.org>
License:      Apache-2.0
Verification: none
Safety-Level: QM

## TYPES

` + "```" + `
Foo := string
` + "```" + `

## BEHAVIOR: do-something

STEPS:
1. Do it.

## PRECONDITIONS
- valid

## POSTCONDITIONS
- done

## INVARIANTS
- tool never modifies input files
- exit_code = 2 on invocation errors

## EXAMPLES

EXAMPLE: test
GIVEN:
  input valid
WHEN:
  do-something(input)
THEN:
  result = ok
`
	path := writeSpec(t, spec)
	result := lintSpec(path, false)

	if result.ExitCode != 0 {
		t.Errorf("expected exit_code=0 (permissive), got %d", result.ExitCode)
	}
	warnCount := 0
	for _, d := range result.Diagnostics {
		if d.Severity == SevWarning && strings.Contains(d.Message, "missing tag") {
			warnCount++
		}
	}
	if warnCount < 2 {
		t.Errorf("expected at least 2 warnings about missing tags, got %d", warnCount)
		for _, d := range result.Diagnostics {
			t.Logf("diag: %+v", d)
		}
	}
	summary := formatSummary(result, false)
	if !strings.HasPrefix(summary, "✓") {
		t.Errorf("expected ✓ summary, got: %s", summary)
	}
}

// ── EXAMPLE: invariant_missing_tag_strict ─────────────────────────────────────

func TestInvariantMissingTagStrict(t *testing.T) {
	spec := `# test-component

## META
Deployment:   cli-tool
Version:      0.1.0
Spec-Schema:  0.1.0
Author:       Jane Example <jane@example.org>
License:      Apache-2.0
Verification: none
Safety-Level: QM

## TYPES

` + "```" + `
Foo := string
` + "```" + `

## BEHAVIOR: do-something

STEPS:
1. Do it.

## PRECONDITIONS
- valid

## POSTCONDITIONS
- done

## INVARIANTS
- tool never modifies input files
- exit_code = 2 on invocation errors

## EXAMPLES

EXAMPLE: test
GIVEN:
  input valid
WHEN:
  do-something(input)
THEN:
  result = ok
`
	path := writeSpec(t, spec)
	result := lintSpec(path, true)

	if result.ExitCode != 1 {
		t.Errorf("expected exit_code=1 (strict), got %d", result.ExitCode)
	}
	summary := formatSummary(result, true)
	if !strings.Contains(summary, "[strict mode]") {
		t.Errorf("expected '[strict mode]' in summary, got: %s", summary)
	}
}

// ── EXAMPLE: behavior_error_exits_no_negative_example ─────────────────────────

func TestBehaviorErrorExitsNoNegativeExample(t *testing.T) {
	spec := `# test-component

## META
Deployment:   cli-tool
Version:      0.1.0
Spec-Schema:  0.1.0
Author:       Jane Example <jane@example.org>
License:      Apache-2.0
Verification: none
Safety-Level: QM

## TYPES

` + "```" + `
Foo := string
` + "```" + `

## BEHAVIOR: transfer

STEPS:
1. Validate inputs; on failure → return Err(INVALID)
2. Transfer funds.

## PRECONDITIONS
- inputs valid

## POSTCONDITIONS
- transfer complete

## INVARIANTS
- [observable] idempotent

## EXAMPLES

EXAMPLE: successful_transfer
GIVEN:
  valid inputs
WHEN:
  transfer(a, b, 10)
THEN:
  result = Ok
`
	path := writeSpec(t, spec)
	result := lintSpec(path, false)

	if result.ExitCode != 1 {
		t.Errorf("expected exit_code=1, got %d", result.ExitCode)
	}
	found := false
	for _, d := range result.Diagnostics {
		if d.Severity == SevError && strings.Contains(d.Message, "error exits in STEPS but no negative-path EXAMPLE") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected Error about missing negative-path EXAMPLE")
		for _, d := range result.Diagnostics {
			t.Logf("diag: %+v", d)
		}
	}
}

// ── EXAMPLE: behavior_error_exits_with_negative_example ───────────────────────

func TestBehaviorErrorExitsWithNegativeExample(t *testing.T) {
	spec := `# test-component

## META
Deployment:   cli-tool
Version:      0.1.0
Spec-Schema:  0.1.0
Author:       Jane Example <jane@example.org>
License:      Apache-2.0
Verification: none
Safety-Level: QM

## TYPES

` + "```" + `
Foo := string
` + "```" + `

## BEHAVIOR: transfer

STEPS:
1. Validate inputs; on failure → return Err(INVALID)
2. Transfer funds.

## PRECONDITIONS
- inputs valid

## POSTCONDITIONS
- transfer complete

## INVARIANTS
- [observable] idempotent

## EXAMPLES

EXAMPLE: successful_transfer
GIVEN:
  valid inputs
WHEN:
  transfer(a, b, 10)
THEN:
  result = Ok

EXAMPLE: transfer_invalid_input
GIVEN:
  amount = -1
WHEN:
  transfer(a, b, -1)
THEN:
  result = Err(INVALID)
`
	path := writeSpec(t, spec)
	result := lintSpec(path, false)

	if result.ExitCode != 0 {
		t.Errorf("expected exit_code=0, got %d", result.ExitCode)
		for _, d := range result.Diagnostics {
			t.Logf("diag: %+v", d)
		}
	}
}

// ── EXAMPLE: behavior_constraint_invalid_value ────────────────────────────────

func TestBehaviorConstraintInvalidValue(t *testing.T) {
	spec := `# test-component

## META
Deployment:   cli-tool
Version:      0.1.0
Spec-Schema:  0.1.0
Author:       Jane Example <jane@example.org>
License:      Apache-2.0
Verification: none
Safety-Level: QM

## TYPES

` + "```" + `
Foo := string
` + "```" + `

## BEHAVIOR: some-op
Constraint: optional

STEPS:
1. Do something.

## PRECONDITIONS
- valid

## POSTCONDITIONS
- done

## INVARIANTS
- [observable] idempotent

## EXAMPLES

EXAMPLE: test
GIVEN:
  valid
WHEN:
  some-op()
THEN:
  result = ok
`
	path := writeSpec(t, spec)
	result := lintSpec(path, false)

	if result.ExitCode != 1 {
		t.Errorf("expected exit_code=1, got %d", result.ExitCode)
	}
	found := false
	for _, d := range result.Diagnostics {
		if d.Severity == SevError &&
			strings.Contains(d.Message, "invalid Constraint: value 'optional'") &&
			strings.Contains(d.Message, "Valid values: required, supported, forbidden") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected Error about invalid Constraint value 'optional'")
		for _, d := range result.Diagnostics {
			t.Logf("diag: %+v", d)
		}
	}
}

// ── EXAMPLE: behavior_constraint_forbidden_no_reason ──────────────────────────

func TestBehaviorConstraintForbiddenNoReason(t *testing.T) {
	spec := `# test-component

## META
Deployment:   cli-tool
Version:      0.1.0
Spec-Schema:  0.1.0
Author:       Jane Example <jane@example.org>
License:      Apache-2.0
Verification: none
Safety-Level: QM

## TYPES

` + "```" + `
Foo := string
` + "```" + `

## BEHAVIOR: legacy-mode
Constraint: forbidden

STEPS:
1. Legacy mode operation.

## PRECONDITIONS
- valid

## POSTCONDITIONS
- done

## INVARIANTS
- [observable] idempotent

## EXAMPLES

EXAMPLE: test
GIVEN:
  valid
WHEN:
  legacy-mode()
THEN:
  result = ok
`
	path := writeSpec(t, spec)
	result := lintSpec(path, false)

	if result.ExitCode != 0 {
		t.Errorf("expected exit_code=0 (warning only), got %d", result.ExitCode)
	}
	found := false
	for _, d := range result.Diagnostics {
		if d.Severity == SevWarning && strings.Contains(d.Message, "Constraint: forbidden but has no reason: annotation") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected Warning about forbidden constraint without reason")
		for _, d := range result.Diagnostics {
			t.Logf("diag: %+v", d)
		}
	}
}

// ── EXAMPLE: behavior_constraint_absent_defaults_required ─────────────────────

func TestBehaviorConstraintAbsentDefaultsRequired(t *testing.T) {
	path := writeSpec(t, minimalValidSpec(""))
	result := lintSpec(path, false)

	if result.ExitCode != 0 {
		t.Errorf("expected exit_code=0, got %d", result.ExitCode)
		for _, d := range result.Diagnostics {
			t.Logf("diag: %+v", d)
		}
	}
}

// ── EXAMPLE: fenced_block_markers_ignored ─────────────────────────────────────

func TestFencedBlockMarkersIgnored(t *testing.T) {
	spec := `# test-component

## META
Deployment:   cli-tool
Version:      0.1.0
Spec-Schema:  0.1.0
Author:       Jane Example <jane@example.org>
License:      Apache-2.0
Verification: none
Safety-Level: QM

## TYPES

` + "```" + `
Foo := string
` + "```" + `

## BEHAVIOR: do-something

STEPS:
1. Do it.

## PRECONDITIONS
- valid

## POSTCONDITIONS
- done

## INVARIANTS
- [observable] idempotent

## EXAMPLES

EXAMPLE: outer
GIVEN:
  some condition
WHEN:
` + "```" + `
EXAMPLE: fake
WHEN: something
THEN: something
` + "```" + `
THEN:
  result = Ok
`
	path := writeSpec(t, spec)
	result := lintSpec(path, false)

	if result.ExitCode != 0 {
		t.Errorf("expected exit_code=0 (fenced markers ignored), got %d", result.ExitCode)
		for _, d := range result.Diagnostics {
			t.Logf("diag: %+v", d)
		}
	}
}

// ── EXAMPLE: milestone_valid_scaffold_first ────────────────────────────────────

func TestMilestoneValidScaffoldFirst(t *testing.T) {
	spec := `# test-component

## META
Deployment:   cli-tool
Version:      0.1.0
Spec-Schema:  0.1.0
Author:       Jane Example <jane@example.org>
License:      Apache-2.0
Verification: none
Safety-Level: QM

## TYPES

` + "```" + `
Foo := string
` + "```" + `

## BEHAVIOR: collect

STEPS:
1. Collect data.

## BEHAVIOR: render

STEPS:
1. Render output.

## BEHAVIOR: main

STEPS:
1. Entry point.

## PRECONDITIONS
- valid

## POSTCONDITIONS
- done

## INVARIANTS
- [observable] idempotent

## EXAMPLES

EXAMPLE: test
GIVEN:
  valid
WHEN:
  collect()
THEN:
  result = ok

## MILESTONE: 0.0.0
Status: released
Scaffold: true
Included BEHAVIORs: collect, render, main
Acceptance criteria:
  ./tool --version | grep -q "^tool "

## MILESTONE: 0.1.0
Status: active
Included BEHAVIORs: collect
Deferred BEHAVIORs: render, main
Acceptance criteria:
  ./tool collect | jq '.result | length > 0'
`
	path := writeSpec(t, spec)
	result := lintSpec(path, false)

	if result.ExitCode != 0 {
		t.Errorf("expected exit_code=0, got %d", result.ExitCode)
		for _, d := range result.Diagnostics {
			t.Logf("diag: %+v", d)
		}
	}
}

// ── EXAMPLE: milestone_scaffold_not_first ─────────────────────────────────────

func TestMilestoneScaffoldNotFirst(t *testing.T) {
	spec := `# test-component

## META
Deployment:   cli-tool
Version:      0.1.0
Spec-Schema:  0.1.0
Author:       Jane Example <jane@example.org>
License:      Apache-2.0
Verification: none
Safety-Level: QM

## TYPES

` + "```" + `
Foo := string
` + "```" + `

## BEHAVIOR: collect

STEPS:
1. Collect data.

## BEHAVIOR: render

STEPS:
1. Render output.

## PRECONDITIONS
- valid

## POSTCONDITIONS
- done

## INVARIANTS
- [observable] idempotent

## EXAMPLES

EXAMPLE: test
GIVEN:
  valid
WHEN:
  collect()
THEN:
  result = ok

## MILESTONE: 0.1.0
Status: released
Included BEHAVIORs: collect
Deferred BEHAVIORs: render

## MILESTONE: 0.2.0
Status: active
Scaffold: true
Included BEHAVIORs: collect, render
Deferred BEHAVIORs: render
`
	path := writeSpec(t, spec)
	result := lintSpec(path, false)

	if result.ExitCode != 1 {
		t.Errorf("expected exit_code=1, got %d", result.ExitCode)
	}
	found := false
	for _, d := range result.Diagnostics {
		if d.Severity == SevError && strings.Contains(d.Message, "must appear first") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected Error about scaffold milestone not being first")
		for _, d := range result.Diagnostics {
			t.Logf("diag: %+v", d)
		}
	}
}

// ── EXAMPLE: milestone_two_scaffold_rejected ──────────────────────────────────

func TestMilestoneTwoScaffoldRejected(t *testing.T) {
	spec := `# test-component

## META
Deployment:   cli-tool
Version:      0.1.0
Spec-Schema:  0.1.0
Author:       Jane Example <jane@example.org>
License:      Apache-2.0
Verification: none
Safety-Level: QM

## TYPES

` + "```" + `
Foo := string
` + "```" + `

## BEHAVIOR: collect

STEPS:
1. Collect data.

## BEHAVIOR: render

STEPS:
1. Render output.

## PRECONDITIONS
- valid

## POSTCONDITIONS
- done

## INVARIANTS
- [observable] idempotent

## EXAMPLES

EXAMPLE: test
GIVEN:
  valid
WHEN:
  collect()
THEN:
  result = ok

## MILESTONE: 0.1.0
Status: active
Scaffold: true
Included BEHAVIORs: collect
Deferred BEHAVIORs: render

## MILESTONE: 0.2.0
Status: pending
Scaffold: true
Included BEHAVIORs: collect, render
Deferred BEHAVIORs: render
`
	path := writeSpec(t, spec)
	result := lintSpec(path, false)

	if result.ExitCode != 1 {
		t.Errorf("expected exit_code=1, got %d", result.ExitCode)
	}
	found := false
	for _, d := range result.Diagnostics {
		if d.Severity == SevError && strings.Contains(d.Message, "More than one MILESTONE has Scaffold: true") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected Error about two scaffold milestones")
		for _, d := range result.Diagnostics {
			t.Logf("diag: %+v", d)
		}
	}
}

// ── EXAMPLE: milestone_two_active_rejected ────────────────────────────────────

func TestMilestoneTwoActiveRejected(t *testing.T) {
	spec := `# test-component

## META
Deployment:   cli-tool
Version:      0.1.0
Spec-Schema:  0.1.0
Author:       Jane Example <jane@example.org>
License:      Apache-2.0
Verification: none
Safety-Level: QM

## TYPES

` + "```" + `
Foo := string
` + "```" + `

## BEHAVIOR: collect

STEPS:
1. Collect data.

## BEHAVIOR: render

STEPS:
1. Render output.

## PRECONDITIONS
- valid

## POSTCONDITIONS
- done

## INVARIANTS
- [observable] idempotent

## EXAMPLES

EXAMPLE: test
GIVEN:
  valid
WHEN:
  collect()
THEN:
  result = ok

## MILESTONE: 0.1.0
Status: active
Included BEHAVIORs: collect
Deferred BEHAVIORs: render

## MILESTONE: 0.2.0
Status: active
Included BEHAVIORs: collect, render
Deferred BEHAVIORs: render
`
	path := writeSpec(t, spec)
	result := lintSpec(path, false)

	if result.ExitCode != 1 {
		t.Errorf("expected exit_code=1, got %d", result.ExitCode)
	}
	found := false
	for _, d := range result.Diagnostics {
		if d.Severity == SevError && strings.Contains(d.Message, "More than one MILESTONE has Status: active") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected Error about two active milestones")
		for _, d := range result.Diagnostics {
			t.Logf("diag: %+v", d)
		}
	}
}

// ── EXAMPLE: milestone_unknown_behavior_name ──────────────────────────────────

func TestMilestoneUnknownBehaviorName(t *testing.T) {
	spec := `# test-component

## META
Deployment:   cli-tool
Version:      0.1.0
Spec-Schema:  0.1.0
Author:       Jane Example <jane@example.org>
License:      Apache-2.0
Verification: none
Safety-Level: QM

## TYPES

` + "```" + `
Foo := string
` + "```" + `

## BEHAVIOR: collect

STEPS:
1. Collect data.

## BEHAVIOR: render

STEPS:
1. Render output.

## PRECONDITIONS
- valid

## POSTCONDITIONS
- done

## INVARIANTS
- [observable] idempotent

## EXAMPLES

EXAMPLE: test
GIVEN:
  valid
WHEN:
  collect()
THEN:
  result = ok

## MILESTONE: 0.1.0
Status: active
Included BEHAVIORs: collect, nonexistent-behavior
Deferred BEHAVIORs: render
`
	path := writeSpec(t, spec)
	result := lintSpec(path, false)

	if result.ExitCode != 1 {
		t.Errorf("expected exit_code=1, got %d", result.ExitCode)
	}
	found := false
	for _, d := range result.Diagnostics {
		if d.Severity == SevError && strings.Contains(d.Message, "nonexistent-behavior") && strings.Contains(d.Message, "Included BEHAVIORs") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected Error about nonexistent-behavior in Included BEHAVIORs")
		for _, d := range result.Diagnostics {
			t.Logf("diag: %+v", d)
		}
	}
}

// ── SPDX validation unit tests ────────────────────────────────────────────────

func TestSPDXValidIdentifiers(t *testing.T) {
	valid := []string{
		"Apache-2.0", "MIT", "GPL-2.0-only", "GPL-3.0-or-later",
		"Apache-2.0 OR MIT", "LGPL-2.1-or-later",
		"GPL-2.0-only AND MIT",
	}
	for _, id := range valid {
		if !isValidSPDX(id) {
			t.Errorf("expected %q to be valid SPDX", id)
		}
	}
}

func TestSPDXInvalidIdentifiers(t *testing.T) {
	invalid := []string{
		"MIT License", "GPL", "Apache", "BSD", "ISC License",
		"not-a-license",
	}
	for _, id := range invalid {
		if isValidSPDX(id) {
			t.Errorf("expected %q to be invalid SPDX", id)
		}
	}
}

// ── Diagnostic ordering test ──────────────────────────────────────────────────

func TestDiagnosticOrdering(t *testing.T) {
	path := writeSpec(t, minimalValidSpec(""))
	result := lintSpec(path, false)

	for i := 1; i < len(result.Diagnostics); i++ {
		if result.Diagnostics[i].Line < result.Diagnostics[i-1].Line {
			t.Errorf("diagnostics not sorted: line %d before line %d",
				result.Diagnostics[i-1].Line, result.Diagnostics[i].Line)
		}
	}
}

// ── list-templates unit test ──────────────────────────────────────────────────

func TestKnownTemplatesCount(t *testing.T) {
	if len(knownTemplates) != 17 {
		t.Errorf("expected 17 known templates, got %d", len(knownTemplates))
	}
}

// ── Format tests ──────────────────────────────────────────────────────────────

func TestFormatSummaryValid(t *testing.T) {
	result := lint.LintResult{File: "spec.md", ExitCode: 0}
	s := formatSummary(result, false)
	if s != "✓ spec.md: valid" {
		t.Errorf("unexpected summary: %q", s)
	}
}

func TestFormatSummaryValidWithWarnings(t *testing.T) {
	result := lint.LintResult{
		File:        "spec.md",
		ExitCode:    0,
		Diagnostics: []lint.Diagnostic{{Severity: SevWarning}},
	}
	s := formatSummary(result, false)
	if s != "✓ spec.md: valid (1 warning(s))" {
		t.Errorf("unexpected summary: %q", s)
	}
}

func TestFormatSummaryError(t *testing.T) {
	result := lint.LintResult{
		File:        "spec.md",
		ExitCode:    1,
		Diagnostics: []lint.Diagnostic{{Severity: SevError}},
	}
	s := formatSummary(result, false)
	if s != "✗ spec.md: 1 error(s), 0 warning(s)" {
		t.Errorf("unexpected summary: %q", s)
	}
}

func TestFormatSummaryStrictMode(t *testing.T) {
	result := lint.LintResult{
		File:        "spec.md",
		ExitCode:    1,
		Diagnostics: []lint.Diagnostic{{Severity: SevWarning}},
	}
	s := formatSummary(result, true)
	if s != "✗ spec.md: 0 error(s), 1 warning(s) [strict mode]" {
		t.Errorf("unexpected summary: %q", s)
	}
}
