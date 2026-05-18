// tests by: mistral-large-2512
// generated from spec: pcd-lint.md sha256:293541ab62274835c61de50947f6283748831c4681cf3f02c4be2f8e942d28a9

package mistral_large_2512

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestMain ensures the pcd-lint binary is built before running tests.
func TestMain(m *testing.M) {
	// Build the pcd-lint binary if it doesn't exist.
	if _, err := os.Stat("../pcd-lint"); os.IsNotExist(err) {
		cmd := exec.Command("go", "build", "-o", "../pcd-lint", "../cmd/pcd-lint/main.go")
		if err := cmd.Run(); err != nil {
			fmt.Printf("Failed to build pcd-lint: %v\n", err)
			os.Exit(1)
		}
	}
	os.Exit(m.Run())
}

// runPcdLint runs the pcd-lint binary with the given args and returns stdout, stderr, and exit code.
func runPcdLint(t *testing.T, args ...string) (string, string, int) {
	cmd := exec.Command("../pcd-lint", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			t.Fatalf("Failed to run pcd-lint: %v", err)
		}
	}
	return stdout.String(), stderr.String(), exitCode
}

// writeFixture writes a spec fixture to a temporary file and returns the file path.
func writeFixture(t *testing.T, name, content string) string {
	dir := "testdata"
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("Failed to create testdata directory: %v", err)
	}
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write fixture %s: %v", name, err)
	}
	return path
}

// TestValidMinimalSpec tests a valid minimal spec (EXAMPLE: valid_minimal_spec).
func TestValidMinimalSpec(t *testing.T) {
	fixture := `# pcd-lint minimal spec

## META
Deployment:   cli-tool
Version:      0.1.0
Spec-Schema:  0.1.0
Author:       Jane Example <jane@example.org>
License:      Apache-2.0
Verification: none
Safety-Level: QM

---

## TYPES

---

## BEHAVIOR: lint

---

## PRECONDITIONS

---

## POSTCONDITIONS

---

## INVARIANTS

---

## EXAMPLES

### EXAMPLE: valid_minimal_spec
GIVEN:
  a valid spec
WHEN:
  pcd-lint is run
THEN:
  exit_code = 0
`
	path := writeFixture(t, "valid_minimal_spec.md", fixture)
	stdout, stderr, exitCode := runPcdLint(t, path)

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}
	if !strings.Contains(stdout, "✓ valid_minimal_spec.md: valid") {
		t.Errorf("Expected stdout to contain '✓ valid_minimal_spec.md: valid', got: %s", stdout)
	}
	if stderr != "" {
		t.Errorf("Expected empty stderr, got: %s", stderr)
	}
}

// TestMultipleAuthorsValid tests a spec with multiple authors (EXAMPLE: multiple_authors_valid).
func TestMultipleAuthorsValid(t *testing.T) {
	fixture := `# pcd-lint multiple authors

## META
Deployment:   cli-tool
Version:      0.1.0
Spec-Schema:  0.1.0
Author:       Jane Example <jane@example.org>
Author:       John Example <john@example.org>
License:      Apache-2.0
Verification: none
Safety-Level: QM

---

## TYPES

---

## BEHAVIOR: lint

---

## PRECONDITIONS

---

## POSTCONDITIONS

---

## INVARIANTS

---

## EXAMPLES

### EXAMPLE: multiple_authors
GIVEN:
  a valid spec with multiple authors
WHEN:
  pcd-lint is run
THEN:
  exit_code = 0
`
	path := writeFixture(t, "multiple_authors_valid.md", fixture)
	stdout, stderr, exitCode := runPcdLint(t, path)

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}
	if !strings.Contains(stdout, "✓ multiple_authors_valid.md: valid") {
		t.Errorf("Expected stdout to contain '✓ multiple_authors_valid.md: valid', got: %s", stdout)
	}
	if stderr != "" {
		t.Errorf("Expected empty stderr, got: %s", stderr)
	}
}

// TestInvalidSpdxLicense tests an invalid SPDX license (EXAMPLE: invalid_spdx_license).
func TestInvalidSpdxLicense(t *testing.T) {
	fixture := `# pcd-lint invalid SPDX

## META
Deployment:   cli-tool
Version:      0.1.0
Spec-Schema:  0.1.0
Author:       Jane Example <jane@example.org>
License:      MIT License
Verification: none
Safety-Level: QM

---

## TYPES

---

## BEHAVIOR: lint

---

## PRECONDITIONS

---

## POSTCONDITIONS

---

## INVARIANTS

---

## EXAMPLES
`
	path := writeFixture(t, "invalid_spdx_license.md", fixture)
	stdout, stderr, exitCode := runPcdLint(t, path)

	if exitCode != 1 {
		t.Errorf("Expected exit code 1, got %d", exitCode)
	}
	if !strings.Contains(stderr, "License 'MIT License' is not a valid SPDX identifier") {
		t.Errorf("Expected stderr to contain SPDX error, got: %s", stderr)
	}
	if !strings.Contains(stdout, "✗ invalid_spdx_license.md: 1 error(s), 0 warning(s)") {
		t.Errorf("Expected stdout to contain error summary, got: %s", stdout)
	}
}

// TestInvalidVersionFormat tests an invalid version format (EXAMPLE: invalid_version_format).
func TestInvalidVersionFormat(t *testing.T) {
	fixture := `# pcd-lint invalid version

## META
Deployment:   cli-tool
Version:      1.0
Spec-Schema:  0.1.0
Author:       Jane Example <jane@example.org>
License:      Apache-2.0
Verification: none
Safety-Level: QM

---

## TYPES

---

## BEHAVIOR: lint

---

## PRECONDITIONS

---

## POSTCONDITIONS

---

## INVARIANTS

---

## EXAMPLES
`
	path := writeFixture(t, "invalid_version_format.md", fixture)
	stdout, stderr, exitCode := runPcdLint(t, path)

	if exitCode != 1 {
		t.Errorf("Expected exit code 1, got %d", exitCode)
	}
	if !strings.Contains(stderr, "Version '1.0' is not valid semantic versioning") {
		t.Errorf("Expected stderr to contain version error, got: %s", stderr)
	}
	if !strings.Contains(stdout, "✗ invalid_version_format.md: 1 error(s), 0 warning(s)") {
		t.Errorf("Expected stdout to contain error summary, got: %s", stdout)
	}
}

// TestMissingAuthor tests a missing Author field (EXAMPLE: missing_author).
func TestMissingAuthor(t *testing.T) {
	fixture := `# pcd-lint missing author

## META
Deployment:   cli-tool
Version:      0.1.0
Spec-Schema:  0.1.0
License:      Apache-2.0
Verification: none
Safety-Level: QM

---

## TYPES

---

## BEHAVIOR: lint

---

## PRECONDITIONS

---

## POSTCONDITIONS

---

## INVARIANTS

---

## EXAMPLES
`
	path := writeFixture(t, "missing_author.md", fixture)
	stdout, stderr, exitCode := runPcdLint(t, path)

	if exitCode != 1 {
		t.Errorf("Expected exit code 1, got %d", exitCode)
	}
	if !strings.Contains(stderr, "Missing required META field: Author (at least one Author: line required)") {
		t.Errorf("Expected stderr to contain missing author error, got: %s", stderr)
	}
	if !strings.Contains(stdout, "✗ missing_author.md: 1 error(s), 0 warning(s)") {
		t.Errorf("Expected stdout to contain error summary, got: %s", stdout)
	}
}

// TestMissingSection tests a missing required section (EXAMPLE: missing_section).
func TestMissingSection(t *testing.T) {
	fixture := `# pcd-lint missing INVARIANTS

## META
Deployment:   cli-tool
Version:      0.1.0
Spec-Schema:  0.1.0
Author:       Jane Example <jane@example.org>
License:      Apache-2.0
Verification: none
Safety-Level: QM

---

## TYPES

---

## BEHAVIOR: lint

---

## PRECONDITIONS

---

## POSTCONDITIONS

---

## EXAMPLES
`
	path := writeFixture(t, "missing_section.md", fixture)
	stdout, stderr, exitCode := runPcdLint(t, path)

	if exitCode != 1 {
		t.Errorf("Expected exit code 1, got %d", exitCode)
	}
	if !strings.Contains(stderr, "Missing required section: ## INVARIANTS") {
		t.Errorf("Expected stderr to contain missing section error, got: %s", stderr)
	}
	if !strings.Contains(stdout, "✗ missing_section.md: 1 error(s), 0 warning(s)") {
		t.Errorf("Expected stdout to contain error summary, got: %s", stdout)
	}
}

// TestUnknownDeploymentTemplate tests an unknown deployment template (EXAMPLE: unknown_deployment_template).
func TestUnknownDeploymentTemplate(t *testing.T) {
	fixture := `# pcd-lint unknown deployment

## META
Deployment:   serverless
Version:      0.1.0
Spec-Schema:  0.1.0
Author:       Jane Example <jane@example.org>
License:      Apache-2.0
Verification: none
Safety-Level: QM

---

## TYPES

---

## BEHAVIOR: lint

---

## PRECONDITIONS

---

## POSTCONDITIONS

---

## INVARIANTS

---

## EXAMPLES
`
	path := writeFixture(t, "unknown_deployment_template.md", fixture)
	stdout, stderr, exitCode := runPcdLint(t, path)

	if exitCode != 1 {
		t.Errorf("Expected exit code 1, got %d", exitCode)
	}
	if !strings.Contains(stderr, "Unknown deployment template: 'serverless'") {
		t.Errorf("Expected stderr to contain unknown deployment error, got: %s", stderr)
	}
	if !strings.Contains(stdout, "✗ unknown_deployment_template.md: 1 error(s), 0 warning(s)") {
		t.Errorf("Expected stdout to contain error summary, got: %s", stdout)
	}
}

// TestDeprecatedTargetFieldPermissive tests a deprecated Target field in permissive mode (EXAMPLE: deprecated_target_field_permissive).
func TestDeprecatedTargetFieldPermissive(t *testing.T) {
	fixture := `# pcd-lint deprecated Target

## META
Deployment:   backend-service
Target:       Go
Version:      0.1.0
Spec-Schema:  0.1.0
Author:       Jane Example <jane@example.org>
License:      Apache-2.0
Verification: none
Safety-Level: QM

---

## TYPES

---

## BEHAVIOR: lint

---

## PRECONDITIONS

---

## POSTCONDITIONS

---

## INVARIANTS

---

## EXAMPLES
`
	path := writeFixture(t, "deprecated_target_field.md", fixture)
	stdout, stderr, exitCode := runPcdLint(t, path)

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}
	if !strings.Contains(stderr, "META field 'Target' is deprecated since v0.3.0") {
		t.Errorf("Expected stderr to contain deprecated Target warning, got: %s", stderr)
	}
	if !strings.Contains(stdout, "✓ deprecated_target_field.md: valid (1 warning(s))") {
		t.Errorf("Expected stdout to contain warning summary, got: %s", stdout)
	}
}

// TestDeprecatedTargetFieldStrict tests a deprecated Target field in strict mode (EXAMPLE: deprecated_target_field_strict).
func TestDeprecatedTargetFieldStrict(t *testing.T) {
	fixture := `# pcd-lint deprecated Target

## META
Deployment:   backend-service
Target:       Go
Version:      0.1.0
Spec-Schema:  0.1.0
Author:       Jane Example <jane@example.org>
License:      Apache-2.0
Verification: none
Safety-Level: QM

---

## TYPES

---

## BEHAVIOR: lint

---

## PRECONDITIONS

---

## POSTCONDITIONS

---

## INVARIANTS

---

## EXAMPLES
`
	path := writeFixture(t, "deprecated_target_field.md", fixture)
	stdout, stderr, exitCode := runPcdLint(t, "strict=true", path)

	if exitCode != 1 {
		t.Errorf("Expected exit code 1, got %d", exitCode)
	}
	if !strings.Contains(stderr, "META field 'Target' is deprecated since v0.3.0") {
		t.Errorf("Expected stderr to contain deprecated Target warning, got: %s", stderr)
	}
	if !strings.Contains(stdout, "✗ deprecated_target_field.md: 0 error(s), 1 warning(s) [strict mode]") {
		t.Errorf("Expected stdout to contain strict mode error summary, got: %s", stdout)
	}
}

// TestEnhanceExistingMissingLanguage tests missing Language field for enhance-existing (EXAMPLE: enhance_existing_missing_language).
func TestEnhanceExistingMissingLanguage(t *testing.T) {
	fixture := `# pcd-lint enhance-existing missing Language

## META
Deployment:   enhance-existing
Version:      0.1.0
Spec-Schema:  0.1.0
Author:       Jane Example <jane@example.org>
License:      Apache-2.0
Verification: none
Safety-Level: QM

---

## TYPES

---

## BEHAVIOR: lint

---

## PRECONDITIONS

---

## POSTCONDITIONS

---

## INVARIANTS

---

## EXAMPLES
`
	path := writeFixture(t, "enhance_existing_missing_language.md", fixture)
	stdout, stderr, exitCode := runPcdLint(t, path)

	if exitCode != 1 {
		t.Errorf("Expected exit code 1, got %d", exitCode)
	}
	if !strings.Contains(stderr, "Deployment 'enhance-existing' requires META field 'Language'") {
		t.Errorf("Expected stderr to contain missing Language error, got: %s", stderr)
	}
	if !strings.Contains(stdout, "✗ enhance_existing_missing_language.md: 1 error(s), 0 warning(s)") {
		t.Errorf("Expected stdout to contain error summary, got: %s", stdout)
	}
}

// TestEmptyGivenBlockPermissive tests an empty GIVEN block (EXAMPLE: empty_given_block_permissive).
func TestEmptyGivenBlockPermissive(t *testing.T) {
	fixture := `# pcd-lint empty GIVEN

## META
Deployment:   cli-tool
Version:      0.1.0
Spec-Schema:  0.1.0
Author:       Jane Example <jane@example.org>
License:      Apache-2.0
Verification: none
Safety-Level: QM

---

## TYPES

---

## BEHAVIOR: lint

---

## PRECONDITIONS

---

## POSTCONDITIONS

---

## INVARIANTS

---

## EXAMPLES

### EXAMPLE: foo
GIVEN:

WHEN:
  result = foo()
THEN:
  result = Ok
`
	path := writeFixture(t, "empty_given_block.md", fixture)
	stdout, stderr, exitCode := runPcdLint(t, path)

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}
	if !strings.Contains(stderr, "Example 'foo' has empty GIVEN block") {
		t.Errorf("Expected stderr to contain empty GIVEN warning, got: %s", stderr)
	}
	if !strings.Contains(stdout, "✓ empty_given_block.md: valid (1 warning(s))") {
		t.Errorf("Expected stdout to contain warning summary, got: %s", stdout)
	}
}

// TestMultipleErrors tests multiple errors in a single spec (EXAMPLE: multiple_errors).
func TestMultipleErrors(t *testing.T) {
	fixture := `# pcd-lint multiple errors

## META
Version:      0.1.0
Spec-Schema:  0.1.0
Author:       Jane Example <jane@example.org>
License:      Apache-2.0
Verification: none
Safety-Level: QM

---

## TYPES

---

## BEHAVIOR: lint

---

## PRECONDITIONS

---

## POSTCONDITIONS
`
	path := writeFixture(t, "multiple_errors.md", fixture)
	stdout, stderr, exitCode := runPcdLint(t, path)

	if exitCode != 1 {
		t.Errorf("Expected exit code 1, got %d", exitCode)
	}
	errors := []string{
		"Missing required section: ## INVARIANTS",
		"Missing required section: ## EXAMPLES",
		"Missing required META field: Deployment",
	}
	for _, err := range errors {
		if !strings.Contains(stderr, err) {
			t.Errorf("Expected stderr to contain '%s', got: %s", err, stderr)
		}
	}
	if !strings.Contains(stdout, "✗ multiple_errors.md: 3 error(s), 0 warning(s)") {
		t.Errorf("Expected stdout to contain error summary, got: %s", stdout)
	}
}

// TestFileNotFound tests a non-existent file (EXAMPLE: file_not_found).
func TestFileNotFound(t *testing.T) {
	stdout, stderr, exitCode := runPcdLint(t, "missing.md")

	if exitCode != 2 {
		t.Errorf("Expected exit code 2, got %d", exitCode)
	}
	if !strings.Contains(stderr, "error: cannot open file: missing.md") {
		t.Errorf("Expected stderr to contain file not found error, got: %s", stderr)
	}
	if stdout != "" {
		t.Errorf("Expected empty stdout, got: %s", stdout)
	}
}

// TestUnrecognisedOption tests an unrecognised key=value argument (EXAMPLE: unrecognised_option).
func TestUnrecognisedOption(t *testing.T) {
	fixture := `# pcd-lint valid spec

## META
Deployment:   cli-tool
Version:      0.1.0
Spec-Schema:  0.1.0
Author:       Jane Example <jane@example.org>
License:      Apache-2.0
Verification: none
Safety-Level: QM

---

## TYPES

---

## BEHAVIOR: lint

---

## PRECONDITIONS

---

## POSTCONDITIONS

---

## INVARIANTS

---

## EXAMPLES
`
	path := writeFixture(t, "valid_spec.md", fixture)
	stdout, stderr, exitCode := runPcdLint(t, "verbose=yes", path)

	if exitCode != 2 {
		t.Errorf("Expected exit code 2, got %d", exitCode)
	}
	if !strings.Contains(stderr, "error: unrecognised option: verbose") {
		t.Errorf("Expected stderr to contain unrecognised option error, got: %s", stderr)
	}
	if stdout != "" {
		t.Errorf("Expected empty stdout, got: %s", stdout)
	}
}

// TestBehaviorInternalRecognised tests BEHAVIOR/INTERNAL recognition (EXAMPLE: behavior_internal_recognised).
func TestBehaviorInternalRecognised(t *testing.T) {
	fixture := `# pcd-lint BEHAVIOR/INTERNAL

## META
Deployment:   cli-tool
Version:      0.1.0
Spec-Schema:  0.1.0
Author:       Jane Example <jane@example.org>
License:      Apache-2.0
Verification: none
Safety-Level: QM

---

## TYPES

---

## BEHAVIOR: lint

---

## BEHAVIOR/INTERNAL: precedence-resolution

---

## PRECONDITIONS

---

## POSTCONDITIONS

---

## INVARIANTS

---

## EXAMPLES
`
	path := writeFixture(t, "behavior_internal_recognised.md", fixture)
	stdout, stderr, exitCode := runPcdLint(t, path)

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}
	if !strings.Contains(stdout, "✓ behavior_internal_recognised.md: valid") {
		t.Errorf("Expected stdout to contain valid summary, got: %s", stdout)
	}
	if stderr != "" {
		t.Errorf("Expected empty stderr, got: %s", stderr)
	}
}

// TestListTemplates tests the list-templates command (EXAMPLE: list_templates).
func TestListTemplates(t *testing.T) {
	stdout, stderr, exitCode := runPcdLint(t, "list-templates")

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}
	if stderr != "" {
		t.Errorf("Expected empty stderr, got: %s", stderr)
	}
	lines := strings.Split(strings.TrimSpace(stdout), "\n")
	if len(lines) != 17 {
		t.Errorf("Expected 17 lines in list-templates output, got %d", len(lines))
	}
	for _, line := range lines {
		if !strings.Contains(line, "→") {
			t.Errorf("Expected line to contain '→', got: %s", line)
		}
	}
}

// TestNonMdExtension tests a file with a non-.md extension (EXAMPLE: non_md_extension).
func TestNonMdExtension(t *testing.T) {
	path := writeFixture(t, "myspec.txt", "# Not a spec")
	stdout, stderr, exitCode := runPcdLint(t, path)

	if exitCode != 2 {
		t.Errorf("Expected exit code 2, got %d", exitCode)
	}
	if !strings.Contains(stderr, "error: file must have .md extension: myspec.txt") {
		t.Errorf("Expected stderr to contain extension error, got: %s", stderr)
	}
	if stdout != "" {
		t.Errorf("Expected empty stdout, got: %s", stdout)
	}
}

// TestMultiPassExampleValid tests a valid multi-pass EXAMPLE (EXAMPLE: multi_pass_example_valid).
func TestMultiPassExampleValid(t *testing.T) {
	fixture := `# pcd-lint multi-pass EXAMPLE

## META
Deployment:   cli-tool
Version:      0.1.0
Spec-Schema:  0.1.0
Author:       Jane Example <jane@example.org>
License:      Apache-2.0
Verification: none
Safety-Level: QM

---

## TYPES

---

## BEHAVIOR: reconcile
STEPS:
  1. on failure → exit 1

---

## PRECONDITIONS

---

## POSTCONDITIONS

---

## INVARIANTS

---

## EXAMPLES

### EXAMPLE: reconcile_graceful_stop
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
	path := writeFixture(t, "multi_pass_example_valid.md", fixture)
	stdout, stderr, exitCode := runPcdLint(t, path)

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}
	if !strings.Contains(stdout, "✓ multi_pass_example_valid.md: valid") {
		t.Errorf("Expected stdout to contain valid summary, got: %s", stdout)
	}
	if stderr != "" {
		t.Errorf("Expected empty stderr, got: %s", stderr)
	}
}