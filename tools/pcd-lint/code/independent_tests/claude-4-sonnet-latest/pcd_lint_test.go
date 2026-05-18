// generated from spec: pcd-lint.md sha256:293541ab62274835c61de50947f6283748831c4681cf3f02c4be2f8e942d28a9
// tests by: claude-4-sonnet-latest

package pcdlinttest

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// binaryPath is the path to the pcd-lint binary under test.
// Tests resolve it relative to the test file's location.
var binaryPath string

func TestMain(m *testing.M) {
	// Attempt to locate or build the binary.
	// The binary should be at ../../pcd-lint (two levels up from independent_tests/<llm>/).
	abs, err := filepath.Abs("../../pcd-lint")
	if err != nil {
		fmt.Fprintf(os.Stderr, "cannot resolve binary path: %v\n", err)
		os.Exit(1)
	}
	binaryPath = abs

	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		// Try to build it
		cmd := exec.Command("go", "build", "-o", binaryPath, "../../cmd/pcd-lint/")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "failed to build pcd-lint: %v\n", err)
			os.Exit(1)
		}
	}

	os.Exit(m.Run())
}

// runBinary runs pcd-lint with the given arguments and returns stdout, stderr, exit code.
func runBinary(t *testing.T, args ...string) (stdout, stderr string, exitCode int) {
	t.Helper()
	cmd := exec.Command(binaryPath, args...)
	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf
	err := cmd.Run()
	exitCode = 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			t.Fatalf("unexpected error running binary: %v", err)
		}
	}
	return outBuf.String(), errBuf.String(), exitCode
}

// writeTempSpec writes content to a temp .md file and returns its path.
func writeTempSpec(t *testing.T, name, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write fixture %s: %v", name, err)
	}
	return path
}

// minimalValidSpec returns a structurally complete valid spec string.
// All required sections and META fields are present.
func minimalValidSpec() string {
	return `# test-spec

## META
Deployment:   cli-tool
Version:      0.1.0
Spec-Schema:  0.1.0
Author:       Jane Example <jane@example.org>
License:      Apache-2.0
Verification: none
Safety-Level: QM

## TYPES

SomeType := string where non-empty

## BEHAVIOR: do-something
Constraint: required

STEPS:
1. Do the thing.

## PRECONDITIONS
- input must be valid

## POSTCONDITIONS
- output is produced

## INVARIANTS
- [observable] tool never modifies input files

## EXAMPLES

### EXAMPLE: happy-path
GIVEN:
  a valid input
WHEN:
  do-something is called
THEN:
  result = Ok
`
}

// ---------------------------------------------------------------------------
// EXAMPLE: valid_minimal_spec
// ---------------------------------------------------------------------------

func TestValidMinimalSpec(t *testing.T) {
	path := writeTempSpec(t, "spec.md", minimalValidSpec())
	stdout, stderr, code := runBinary(t, path)
	if code != 0 {
		t.Errorf("expected exit 0, got %d; stderr=%q", code, stderr)
	}
	if stderr != "" {
		t.Errorf("expected empty stderr, got: %q", stderr)
	}
	base := filepath.Base(path)
	want := fmt.Sprintf("✓ %s: valid", base)
	if !strings.Contains(stdout, want) {
		t.Errorf("expected stdout to contain %q, got: %q", want, stdout)
	}
}

// ---------------------------------------------------------------------------
// EXAMPLE: multiple_authors_valid
// ---------------------------------------------------------------------------

func TestMultipleAuthorsValid(t *testing.T) {
	spec := `# test-spec

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

SomeType := string where non-empty

## BEHAVIOR: do-something

STEPS:
1. Do the thing.

## PRECONDITIONS
- input must be valid

## POSTCONDITIONS
- output is produced

## INVARIANTS
- [observable] tool never modifies input files

## EXAMPLES

### EXAMPLE: happy-path
GIVEN:
  a valid input
WHEN:
  do-something is called
THEN:
  result = Ok
`
	path := writeTempSpec(t, "spec.md", spec)
	stdout, stderr, code := runBinary(t, path)
	if code != 0 {
		t.Errorf("expected exit 0, got %d; stderr=%q", code, stderr)
	}
	if stderr != "" {
		t.Errorf("expected empty stderr, got: %q", stderr)
	}
	base := filepath.Base(path)
	if !strings.Contains(stdout, fmt.Sprintf("✓ %s: valid", base)) {
		t.Errorf("expected valid summary in stdout, got: %q", stdout)
	}
}

// ---------------------------------------------------------------------------
// EXAMPLE: invalid_spdx_license
// ---------------------------------------------------------------------------

func TestInvalidSpdxLicense(t *testing.T) {
	spec := strings.ReplaceAll(minimalValidSpec(), "License:      Apache-2.0", "License:      MIT License")
	path := writeTempSpec(t, "spec.md", spec)
	_, stderr, code := runBinary(t, path)
	if code != 1 {
		t.Errorf("expected exit 1, got %d", code)
	}
	if !strings.Contains(stderr, "License 'MIT License' is not a valid SPDX identifier") {
		t.Errorf("expected SPDX error in stderr, got: %q", stderr)
	}
	if !strings.Contains(stderr, "https://spdx.org/licenses/") {
		t.Errorf("expected spdx.org URL in stderr, got: %q", stderr)
	}
}

// ---------------------------------------------------------------------------
// EXAMPLE: invalid_version_format
// ---------------------------------------------------------------------------

func TestInvalidVersionFormat(t *testing.T) {
	spec := strings.ReplaceAll(minimalValidSpec(), "Version:      0.1.0", "Version:      1.0")
	path := writeTempSpec(t, "spec.md", spec)
	_, stderr, code := runBinary(t, path)
	if code != 1 {
		t.Errorf("expected exit 1, got %d", code)
	}
	if !strings.Contains(stderr, "Version '1.0' is not valid semantic versioning") {
		t.Errorf("expected version error in stderr, got: %q", stderr)
	}
}

// ---------------------------------------------------------------------------
// EXAMPLE: missing_author
// ---------------------------------------------------------------------------

func TestMissingAuthor(t *testing.T) {
	spec := strings.ReplaceAll(minimalValidSpec(), "Author:       Jane Example <jane@example.org>\n", "")
	path := writeTempSpec(t, "spec.md", spec)
	_, stderr, code := runBinary(t, path)
	if code != 1 {
		t.Errorf("expected exit 1, got %d", code)
	}
	if !strings.Contains(stderr, "Missing required META field: Author (at least one Author: line required)") {
		t.Errorf("expected missing author error, got: %q", stderr)
	}
}

// ---------------------------------------------------------------------------
// EXAMPLE: missing_section — INVARIANTS
// ---------------------------------------------------------------------------

func TestMissingSection(t *testing.T) {
	// Remove the INVARIANTS section
	spec := `# test-spec

## META
Deployment:   cli-tool
Version:      0.1.0
Spec-Schema:  0.1.0
Author:       Jane Example <jane@example.org>
License:      Apache-2.0
Verification: none
Safety-Level: QM

## TYPES

## BEHAVIOR: do-something

STEPS:
1. Do the thing.

## PRECONDITIONS
- input must be valid

## POSTCONDITIONS
- output is produced

## EXAMPLES

### EXAMPLE: happy-path
GIVEN:
  a valid input
WHEN:
  do-something is called
THEN:
  result = Ok
`
	path := writeTempSpec(t, "spec.md", spec)
	stdout, stderr, code := runBinary(t, path)
	if code != 1 {
		t.Errorf("expected exit 1, got %d", code)
	}
	if !strings.Contains(stderr, "Missing required section: ## INVARIANTS") {
		t.Errorf("expected missing INVARIANTS in stderr, got: %q", stderr)
	}
	base := filepath.Base(path)
	if !strings.Contains(stdout, fmt.Sprintf("✗ %s: 1 error(s), 0 warning(s)", base)) {
		t.Errorf("expected error summary in stdout, got: %q", stdout)
	}
}

// ---------------------------------------------------------------------------
// EXAMPLE: unknown_deployment_template
// ---------------------------------------------------------------------------

func TestUnknownDeploymentTemplate(t *testing.T) {
	spec := strings.ReplaceAll(minimalValidSpec(), "Deployment:   cli-tool", "Deployment:   serverless")
	path := writeTempSpec(t, "spec.md", spec)
	_, stderr, code := runBinary(t, path)
	if code != 1 {
		t.Errorf("expected exit 1, got %d", code)
	}
	if !strings.Contains(stderr, "Unknown deployment template: 'serverless'") {
		t.Errorf("expected unknown deployment error, got: %q", stderr)
	}
	if !strings.Contains(stderr, "pcd-lint list-templates") {
		t.Errorf("expected list-templates hint in stderr, got: %q", stderr)
	}
}

// ---------------------------------------------------------------------------
// EXAMPLE: deprecated_target_field_permissive
// ---------------------------------------------------------------------------

func TestDeprecatedTargetFieldPermissive(t *testing.T) {
	spec := `# test-spec

## META
Deployment:   backend-service
Target:       Go
Version:      0.1.0
Spec-Schema:  0.1.0
Author:       Jane Example <jane@example.org>
License:      Apache-2.0
Verification: none
Safety-Level: QM

## TYPES

SomeType := string where non-empty

## BEHAVIOR: do-something

STEPS:
1. Do the thing.

## PRECONDITIONS
- input must be valid

## POSTCONDITIONS
- output is produced

## INVARIANTS
- [observable] tool never modifies input files

## EXAMPLES

### EXAMPLE: happy-path
GIVEN:
  a valid input
WHEN:
  do-something is called
THEN:
  result = Ok
`
	path := writeTempSpec(t, "spec.md", spec)
	stdout, stderr, code := runBinary(t, path)
	if code != 0 {
		t.Errorf("expected exit 0, got %d", code)
	}
	if !strings.Contains(stderr, "META field 'Target' is deprecated since v0.3.0") {
		t.Errorf("expected Target deprecated warning, got: %q", stderr)
	}
	base := filepath.Base(path)
	if !strings.Contains(stdout, fmt.Sprintf("✓ %s: valid (1 warning(s))", base)) {
		t.Errorf("expected warning summary in stdout, got: %q", stdout)
	}
}

// ---------------------------------------------------------------------------
// EXAMPLE: deprecated_target_field_strict
// ---------------------------------------------------------------------------

func TestDeprecatedTargetFieldStrict(t *testing.T) {
	spec := `# test-spec

## META
Deployment:   backend-service
Target:       Go
Version:      0.1.0
Spec-Schema:  0.1.0
Author:       Jane Example <jane@example.org>
License:      Apache-2.0
Verification: none
Safety-Level: QM

## TYPES

SomeType := string where non-empty

## BEHAVIOR: do-something

STEPS:
1. Do the thing.

## PRECONDITIONS
- input must be valid

## POSTCONDITIONS
- output is produced

## INVARIANTS
- [observable] tool never modifies input files

## EXAMPLES

### EXAMPLE: happy-path
GIVEN:
  a valid input
WHEN:
  do-something is called
THEN:
  result = Ok
`
	path := writeTempSpec(t, "spec.md", spec)
	stdout, stderr, code := runBinary(t, "strict=true", path)
	if code != 1 {
		t.Errorf("expected exit 1 in strict mode, got %d", code)
	}
	if !strings.Contains(stderr, "META field 'Target' is deprecated since v0.3.0") {
		t.Errorf("expected Target deprecated warning, got: %q", stderr)
	}
	base := filepath.Base(path)
	if !strings.Contains(stdout, fmt.Sprintf("✗ %s: 0 error(s), 1 warning(s) [strict mode]", base)) {
		t.Errorf("expected strict mode summary in stdout, got: %q", stdout)
	}
}

// ---------------------------------------------------------------------------
// EXAMPLE: enhance_existing_missing_language
// ---------------------------------------------------------------------------

func TestEnhanceExistingMissingLanguage(t *testing.T) {
	spec := `# test-spec

## META
Deployment:   enhance-existing
Version:      0.1.0
Spec-Schema:  0.1.0
Author:       Jane Example <jane@example.org>
License:      Apache-2.0
Verification: none
Safety-Level: QM

## TYPES

SomeType := string where non-empty

## BEHAVIOR: do-something

STEPS:
1. Do the thing.

## PRECONDITIONS
- input must be valid

## POSTCONDITIONS
- output is produced

## INVARIANTS
- [observable] tool never modifies input files

## EXAMPLES

### EXAMPLE: happy-path
GIVEN:
  a valid input
WHEN:
  do-something is called
THEN:
  result = Ok
`
	path := writeTempSpec(t, "spec.md", spec)
	_, stderr, code := runBinary(t, path)
	if code != 1 {
		t.Errorf("expected exit 1, got %d", code)
	}
	if !strings.Contains(stderr, "Deployment 'enhance-existing' requires META field 'Language'") {
		t.Errorf("expected missing Language error, got: %q", stderr)
	}
}

// ---------------------------------------------------------------------------
// EXAMPLE: empty_given_block_permissive
// ---------------------------------------------------------------------------

func TestEmptyGivenBlockPermissive(t *testing.T) {
	spec := `# test-spec

## META
Deployment:   cli-tool
Version:      0.1.0
Spec-Schema:  0.1.0
Author:       Jane Example <jane@example.org>
License:      Apache-2.0
Verification: none
Safety-Level: QM

## TYPES

SomeType := string where non-empty

## BEHAVIOR: do-something

STEPS:
1. Do the thing.

## PRECONDITIONS
- input must be valid

## POSTCONDITIONS
- output is produced

## INVARIANTS
- [observable] tool never modifies input files

## EXAMPLES

### EXAMPLE: foo
GIVEN:

WHEN:
  result = foo()
THEN:
  result = Ok
`
	path := writeTempSpec(t, "spec.md", spec)
	stdout, stderr, code := runBinary(t, path)
	if code != 0 {
		t.Errorf("expected exit 0, got %d", code)
	}
	if !strings.Contains(stderr, "Example 'foo' has empty GIVEN block") {
		t.Errorf("expected empty GIVEN warning, got: %q", stderr)
	}
	base := filepath.Base(path)
	if !strings.Contains(stdout, fmt.Sprintf("✓ %s: valid (1 warning(s))", base)) {
		t.Errorf("expected warning summary, got: %q", stdout)
	}
}

// ---------------------------------------------------------------------------
// EXAMPLE: multiple_errors
// ---------------------------------------------------------------------------

func TestMultipleErrors(t *testing.T) {
	// Missing ## INVARIANTS, ## EXAMPLES, and Deployment field
	spec := `# test-spec

## META
Version:      0.1.0
Spec-Schema:  0.1.0
Author:       Jane Example <jane@example.org>
License:      Apache-2.0
Verification: none
Safety-Level: QM

## TYPES

## BEHAVIOR: do-something

STEPS:
1. Do the thing.

## PRECONDITIONS
- input valid

## POSTCONDITIONS
- output produced
`
	path := writeTempSpec(t, "spec.md", spec)
	stdout, stderr, code := runBinary(t, path)
	if code != 1 {
		t.Errorf("expected exit 1, got %d", code)
	}
	for _, msg := range []string{
		"Missing required section: ## INVARIANTS",
		"Missing required section: ## EXAMPLES",
		"Missing required META field: Deployment",
	} {
		if !strings.Contains(stderr, msg) {
			t.Errorf("expected %q in stderr, got: %q", msg, stderr)
		}
	}
	base := filepath.Base(path)
	if !strings.Contains(stdout, fmt.Sprintf("✗ %s: 3 error(s), 0 warning(s)", base)) {
		t.Errorf("expected 3-error summary in stdout, got: %q", stdout)
	}
}

// ---------------------------------------------------------------------------
// EXAMPLE: file_not_found
// ---------------------------------------------------------------------------

func TestFileNotFound(t *testing.T) {
	stdout, stderr, code := runBinary(t, "missing.md")
	if code != 2 {
		t.Errorf("expected exit 2, got %d", code)
	}
	if !strings.Contains(stderr, "error: cannot open file: missing.md") {
		t.Errorf("expected file-not-found error, got: %q", stderr)
	}
	if stdout != "" {
		t.Errorf("expected empty stdout, got: %q", stdout)
	}
}

// ---------------------------------------------------------------------------
// EXAMPLE: unrecognised_option
// ---------------------------------------------------------------------------

func TestUnrecognisedOption(t *testing.T) {
	path := writeTempSpec(t, "spec.md", minimalValidSpec())
	stdout, stderr, code := runBinary(t, "verbose=yes", path)
	if code != 2 {
		t.Errorf("expected exit 2, got %d", code)
	}
	if !strings.Contains(stderr, "error: unrecognised option: verbose") {
		t.Errorf("expected unrecognised option error, got: %q", stderr)
	}
	if stdout != "" {
		t.Errorf("expected empty stdout, got: %q", stdout)
	}
}

// ---------------------------------------------------------------------------
// EXAMPLE: behavior_internal_recognised
// ---------------------------------------------------------------------------

func TestBehaviorInternalRecognised(t *testing.T) {
	spec := `# test-spec

## META
Deployment:   cli-tool
Version:      0.1.0
Spec-Schema:  0.1.0
Author:       Jane Example <jane@example.org>
License:      Apache-2.0
Verification: none
Safety-Level: QM

## TYPES

SomeType := string where non-empty

## BEHAVIOR: lint

STEPS:
1. Do lint.

## BEHAVIOR/INTERNAL: precedence-resolution

STEPS:
1. Resolve precedence.

## PRECONDITIONS
- file must exist

## POSTCONDITIONS
- diagnostics emitted

## INVARIANTS
- [observable] tool is idempotent

## EXAMPLES

### EXAMPLE: happy-path
GIVEN:
  a valid input
WHEN:
  lint is called
THEN:
  exit_code = 0
`
	path := writeTempSpec(t, "spec.md", spec)
	stdout, stderr, code := runBinary(t, path)
	if code != 0 {
		t.Errorf("expected exit 0, got %d; stderr=%q", code, stderr)
	}
	if stderr != "" {
		t.Errorf("expected empty stderr, got: %q", stderr)
	}
	base := filepath.Base(path)
	if !strings.Contains(stdout, fmt.Sprintf("✓ %s: valid", base)) {
		t.Errorf("expected valid summary, got: %q", stdout)
	}
}

// ---------------------------------------------------------------------------
// EXAMPLE: behavior_internal_unknown_variant
// ---------------------------------------------------------------------------

func TestBehaviorInternalUnknownVariant(t *testing.T) {
	spec := `# test-spec

## META
Deployment:   cli-tool
Version:      0.1.0
Spec-Schema:  0.1.0
Author:       Jane Example <jane@example.org>
License:      Apache-2.0
Verification: none
Safety-Level: QM

## TYPES

SomeType := string where non-empty

## BEHAVIOR/PRIVATE: foo

STEPS:
1. Do foo.

## PRECONDITIONS
- input valid

## POSTCONDITIONS
- output produced

## INVARIANTS
- [observable] tool is idempotent

## EXAMPLES

### EXAMPLE: happy-path
GIVEN:
  a valid input
WHEN:
  foo is called
THEN:
  result = Ok
`
	path := writeTempSpec(t, "spec.md", spec)
	_, stderr, code := runBinary(t, path)
	if code != 1 {
		t.Errorf("expected exit 1, got %d", code)
	}
	if !strings.Contains(stderr, "Missing required section: ## BEHAVIOR") {
		t.Errorf("expected missing BEHAVIOR error, got: %q", stderr)
	}
}

// ---------------------------------------------------------------------------
// EXAMPLE: list_templates
// ---------------------------------------------------------------------------

func TestListTemplates(t *testing.T) {
	stdout, stderr, code := runBinary(t, "list-templates")
	if code != 0 {
		t.Errorf("expected exit 0, got %d", code)
	}
	if stderr != "" {
		t.Errorf("expected empty stderr, got: %q", stderr)
	}
	lines := strings.Split(strings.TrimRight(stdout, "\n"), "\n")
	if len(lines) != 17 {
		t.Errorf("expected exactly 17 lines, got %d: %q", len(lines), stdout)
	}
	for _, line := range lines {
		if !strings.Contains(line, "→") {
			t.Errorf("expected '→' in line: %q", line)
		}
	}
}

// ---------------------------------------------------------------------------
// EXAMPLE: non_md_extension
// ---------------------------------------------------------------------------

func TestNonMdExtension(t *testing.T) {
	// Write a file with .txt extension
	dir := t.TempDir()
	path := filepath.Join(dir, "myspec.txt")
	if err := os.WriteFile(path, []byte("content"), 0644); err != nil {
		t.Fatal(err)
	}
	stdout, stderr, code := runBinary(t, path)
	if code != 2 {
		t.Errorf("expected exit 2, got %d", code)
	}
	// The spec says: "error: file must have .md extension: myspec.txt"
	// It should include the basename at minimum
	if !strings.Contains(stderr, "error: file must have .md extension:") {
		t.Errorf("expected .md extension error, got: %q", stderr)
	}
	if stdout != "" {
		t.Errorf("expected empty stdout, got: %q", stdout)
	}
}

// ---------------------------------------------------------------------------
// EXAMPLE: multi_pass_example_valid
// ---------------------------------------------------------------------------

func TestMultiPassExampleValid(t *testing.T) {
	spec := `# test-spec

## META
Deployment:   cli-tool
Version:      0.1.0
Spec-Schema:  0.1.0
Author:       Jane Example <jane@example.org>
License:      Apache-2.0
Verification: none
Safety-Level: QM

## TYPES

SomeType := string where non-empty

## BEHAVIOR: reconcile

STEPS:
1. Validate; on failure → return Err(INVALID)
2. Apply changes.

## PRECONDITIONS
- input valid

## POSTCONDITIONS
- state updated

## INVARIANTS
- [observable] tool is idempotent

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

### EXAMPLE: reconcile_error
GIVEN:
  invalid inputs
WHEN:
  reconcile is called
THEN:
  result = Err(INVALID)
`
	path := writeTempSpec(t, "spec.md", spec)
	stdout, stderr, code := runBinary(t, path)
	if code != 0 {
		t.Errorf("expected exit 0, got %d; stderr=%q", code, stderr)
	}
	if stderr != "" {
		t.Errorf("expected empty stderr, got: %q", stderr)
	}
	base := filepath.Base(path)
	if !strings.Contains(stdout, fmt.Sprintf("✓ %s: valid", base)) {
		t.Errorf("expected valid summary, got: %q", stdout)
	}
}

// ---------------------------------------------------------------------------
// EXAMPLE: behavior_missing_steps
// ---------------------------------------------------------------------------

func TestBehaviorMissingSteps(t *testing.T) {
	spec := `# test-spec

## META
Deployment:   cli-tool
Version:      0.1.0
Spec-Schema:  0.1.0
Author:       Jane Example <jane@example.org>
License:      Apache-2.0
Verification: none
Safety-Level: QM

## TYPES

SomeType := string where non-empty

## BEHAVIOR: do-something
Constraint: required

PRECONDITIONS:
  - input is valid
POSTCONDITIONS:
  - output is produced

## PRECONDITIONS
- input valid

## POSTCONDITIONS
- output produced

## INVARIANTS
- [observable] tool is idempotent

## EXAMPLES

### EXAMPLE: happy-path
GIVEN:
  a valid input
WHEN:
  do-something is called
THEN:
  result = Ok
`
	path := writeTempSpec(t, "spec.md", spec)
	_, stderr, code := runBinary(t, path)
	if code != 1 {
		t.Errorf("expected exit 1, got %d", code)
	}
	if !strings.Contains(stderr, "missing required STEPS: block") {
		t.Errorf("expected missing STEPS error, got: %q", stderr)
	}
}

// ---------------------------------------------------------------------------
// EXAMPLE: invariant_missing_tag_warning
// ---------------------------------------------------------------------------

func TestInvariantMissingTagWarning(t *testing.T) {
	spec := `# test-spec

## META
Deployment:   cli-tool
Version:      0.1.0
Spec-Schema:  0.1.0
Author:       Jane Example <jane@example.org>
License:      Apache-2.0
Verification: none
Safety-Level: QM

## TYPES

SomeType := string where non-empty

## BEHAVIOR: do-something

STEPS:
1. Do the thing.

## PRECONDITIONS
- input valid

## POSTCONDITIONS
- output produced

## INVARIANTS
- tool never modifies input files
- exit_code = 2 on invocation errors

## EXAMPLES

### EXAMPLE: happy-path
GIVEN:
  a valid input
WHEN:
  do-something is called
THEN:
  result = Ok
`
	path := writeTempSpec(t, "spec.md", spec)
	stdout, stderr, code := runBinary(t, path)
	if code != 0 {
		t.Errorf("expected exit 0, got %d", code)
	}
	// Two warnings for two untagged invariant entries
	if !strings.Contains(stderr, "missing tag") {
		t.Errorf("expected 'missing tag' warning, got: %q", stderr)
	}
	base := filepath.Base(path)
	if !strings.Contains(stdout, fmt.Sprintf("✓ %s: valid (2 warning(s))", base)) {
		t.Errorf("expected 2-warning summary, got: %q", stdout)
	}
}

// ---------------------------------------------------------------------------
// EXAMPLE: invariant_missing_tag_strict
// ---------------------------------------------------------------------------

func TestInvariantMissingTagStrict(t *testing.T) {
	spec := `# test-spec

## META
Deployment:   cli-tool
Version:      0.1.0
Spec-Schema:  0.1.0
Author:       Jane Example <jane@example.org>
License:      Apache-2.0
Verification: none
Safety-Level: QM

## TYPES

SomeType := string where non-empty

## BEHAVIOR: do-something

STEPS:
1. Do the thing.

## PRECONDITIONS
- input valid

## POSTCONDITIONS
- output produced

## INVARIANTS
- tool never modifies input files
- exit_code = 2 on invocation errors

## EXAMPLES

### EXAMPLE: happy-path
GIVEN:
  a valid input
WHEN:
  do-something is called
THEN:
  result = Ok
`
	path := writeTempSpec(t, "spec.md", spec)
	stdout, _, code := runBinary(t, "strict=true", path)
	if code != 1 {
		t.Errorf("expected exit 1 in strict mode, got %d", code)
	}
	if !strings.Contains(stdout, "[strict mode]") {
		t.Errorf("expected [strict mode] in stdout, got: %q", stdout)
	}
}

// ---------------------------------------------------------------------------
// EXAMPLE: behavior_error_exits_no_negative_example
// ---------------------------------------------------------------------------

func TestBehaviorErrorExitsNoNegativeExample(t *testing.T) {
	spec := `# test-spec

## META
Deployment:   cli-tool
Version:      0.1.0
Spec-Schema:  0.1.0
Author:       Jane Example <jane@example.org>
License:      Apache-2.0
Verification: none
Safety-Level: QM

## TYPES

SomeType := string where non-empty

## BEHAVIOR: transfer

STEPS:
1. Validate inputs; on failure → return Err(INVALID)
2. Transfer funds.

## PRECONDITIONS
- inputs valid

## POSTCONDITIONS
- funds transferred

## INVARIANTS
- [observable] tool is idempotent

## EXAMPLES

### EXAMPLE: successful_transfer
GIVEN:
  valid inputs
WHEN:
  transfer(a, b, 10)
THEN:
  result = Ok
`
	path := writeTempSpec(t, "spec.md", spec)
	_, stderr, code := runBinary(t, path)
	if code != 1 {
		t.Errorf("expected exit 1, got %d", code)
	}
	if !strings.Contains(stderr, "has error exits in STEPS but no negative-path EXAMPLE") {
		t.Errorf("expected negative-path error, got: %q", stderr)
	}
}

// ---------------------------------------------------------------------------
// EXAMPLE: behavior_error_exits_with_negative_example
// ---------------------------------------------------------------------------

func TestBehaviorErrorExitsWithNegativeExample(t *testing.T) {
	spec := `# test-spec

## META
Deployment:   cli-tool
Version:      0.1.0
Spec-Schema:  0.1.0
Author:       Jane Example <jane@example.org>
License:      Apache-2.0
Verification: none
Safety-Level: QM

## TYPES

SomeType := string where non-empty

## BEHAVIOR: transfer

STEPS:
1. Validate inputs; on failure → return Err(INVALID)
2. Transfer funds.

## PRECONDITIONS
- inputs valid

## POSTCONDITIONS
- funds transferred

## INVARIANTS
- [observable] tool is idempotent

## EXAMPLES

### EXAMPLE: successful_transfer
GIVEN:
  valid inputs
WHEN:
  transfer(a, b, 10)
THEN:
  result = Ok

### EXAMPLE: transfer_invalid_input
GIVEN:
  amount = -1
WHEN:
  transfer(a, b, -1)
THEN:
  result = Err(INVALID)
`
	path := writeTempSpec(t, "spec.md", spec)
	_, stderr, code := runBinary(t, path)
	if code != 0 {
		t.Errorf("expected exit 0, got %d; stderr=%q", code, stderr)
	}
	if stderr != "" {
		t.Errorf("expected empty stderr, got: %q", stderr)
	}
}

// ---------------------------------------------------------------------------
// EXAMPLE: behavior_constraint_invalid_value
// ---------------------------------------------------------------------------

func TestBehaviorConstraintInvalidValue(t *testing.T) {
	spec := `# test-spec

## META
Deployment:   cli-tool
Version:      0.1.0
Spec-Schema:  0.1.0
Author:       Jane Example <jane@example.org>
License:      Apache-2.0
Verification: none
Safety-Level: QM

## TYPES

SomeType := string where non-empty

## BEHAVIOR: some-op
Constraint: optional

STEPS:
1. Do the thing.

## PRECONDITIONS
- input valid

## POSTCONDITIONS
- output produced

## INVARIANTS
- [observable] tool is idempotent

## EXAMPLES

### EXAMPLE: happy-path
GIVEN:
  a valid input
WHEN:
  some-op is called
THEN:
  result = Ok
`
	path := writeTempSpec(t, "spec.md", spec)
	_, stderr, code := runBinary(t, path)
	if code != 1 {
		t.Errorf("expected exit 1, got %d", code)
	}
	if !strings.Contains(stderr, "invalid Constraint: value 'optional'") {
		t.Errorf("expected invalid constraint error, got: %q", stderr)
	}
	if !strings.Contains(stderr, "Valid values: required, supported, forbidden") {
		t.Errorf("expected valid values hint, got: %q", stderr)
	}
}

// ---------------------------------------------------------------------------
// EXAMPLE: behavior_constraint_forbidden_no_reason
// ---------------------------------------------------------------------------

func TestBehaviorConstraintForbiddenNoReason(t *testing.T) {
	spec := `# test-spec

## META
Deployment:   cli-tool
Version:      0.1.0
Spec-Schema:  0.1.0
Author:       Jane Example <jane@example.org>
License:      Apache-2.0
Verification: none
Safety-Level: QM

## TYPES

SomeType := string where non-empty

## BEHAVIOR: do-something

STEPS:
1. Do the thing.

## BEHAVIOR: legacy-mode
Constraint: forbidden

STEPS:
1. Legacy step.

## PRECONDITIONS
- input valid

## POSTCONDITIONS
- output produced

## INVARIANTS
- [observable] tool is idempotent

## EXAMPLES

### EXAMPLE: happy-path
GIVEN:
  a valid input
WHEN:
  do-something is called
THEN:
  result = Ok
`
	path := writeTempSpec(t, "spec.md", spec)
	stdout, stderr, code := runBinary(t, path)
	if code != 0 {
		t.Errorf("expected exit 0, got %d", code)
	}
	if !strings.Contains(stderr, "Constraint: forbidden but has no reason: annotation") {
		t.Errorf("expected forbidden-no-reason warning, got: %q", stderr)
	}
	base := filepath.Base(path)
	// At minimum one warning
	if strings.Contains(stdout, "✗") {
		t.Errorf("expected success (exit 0), stdout indicates failure: %q", stdout)
	}
	_ = base
}

// ---------------------------------------------------------------------------
// EXAMPLE: behavior_constraint_absent_defaults_required
// ---------------------------------------------------------------------------

func TestBehaviorConstraintAbsentDefaultsRequired(t *testing.T) {
	// Absence of Constraint: field is valid — no diagnostic emitted
	path := writeTempSpec(t, "spec.md", minimalValidSpec())
	_, stderr, code := runBinary(t, path)
	if code != 0 {
		t.Errorf("expected exit 0, got %d", code)
	}
	if strings.Contains(stderr, "Constraint") {
		t.Errorf("expected no Constraint diagnostic, got: %q", stderr)
	}
}

// ---------------------------------------------------------------------------
// EXAMPLE: fenced_block_markers_ignored
// ---------------------------------------------------------------------------

func TestFencedBlockMarkersIgnored(t *testing.T) {
	spec := `# test-spec

## META
Deployment:   cli-tool
Version:      0.1.0
Spec-Schema:  0.1.0
Author:       Jane Example <jane@example.org>
License:      Apache-2.0
Verification: none
Safety-Level: QM

## TYPES

SomeType := string where non-empty

## BEHAVIOR: do-something

STEPS:
1. Do the thing.

## PRECONDITIONS
- input valid

## POSTCONDITIONS
- output produced

## INVARIANTS
- [observable] tool is idempotent

## EXAMPLES

### EXAMPLE: outer
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
	path := writeTempSpec(t, "spec.md", spec)
	stdout, stderr, code := runBinary(t, path)
	if code != 0 {
		t.Errorf("expected exit 0, got %d; stderr=%q", code, stderr)
	}
	if stderr != "" {
		t.Errorf("expected empty stderr, got: %q", stderr)
	}
	base := filepath.Base(path)
	if !strings.Contains(stdout, fmt.Sprintf("✓ %s: valid", base)) {
		t.Errorf("expected valid summary, got: %q", stdout)
	}
}

// ---------------------------------------------------------------------------
// INVARIANT: idempotent — running twice produces identical output
// ---------------------------------------------------------------------------

func TestIdempotent(t *testing.T) {
	path := writeTempSpec(t, "spec.md", minimalValidSpec())
	out1, err1, code1 := runBinary(t, path)
	out2, err2, code2 := runBinary(t, path)
	if code1 != code2 {
		t.Errorf("idempotent: exit codes differ: %d vs %d", code1, code2)
	}
	if out1 != out2 {
		t.Errorf("idempotent: stdout differs:\n%q\nvs\n%q", out1, out2)
	}
	if err1 != err2 {
		t.Errorf("idempotent: stderr differs:\n%q\nvs\n%q", err1, err2)
	}
}

// ---------------------------------------------------------------------------
// INVARIANT: exit_code=2 is invocation error only
// ---------------------------------------------------------------------------

func TestExitCode2IsInvocationOnly(t *testing.T) {
	// Bad extension triggers invocation error
	dir := t.TempDir()
	p := filepath.Join(dir, "x.txt")
	_ = os.WriteFile(p, []byte("x"), 0644)
	_, _, code := runBinary(t, p)
	if code != 2 {
		t.Errorf("expected exit 2 for .txt file, got %d", code)
	}

	// Missing file triggers invocation error
	_, _, code2 := runBinary(t, "nonexistent.md")
	if code2 != 2 {
		t.Errorf("expected exit 2 for missing file, got %d", code2)
	}
}

// ---------------------------------------------------------------------------
// INVARIANT: warnings alone never exit 1 unless strict=true
// ---------------------------------------------------------------------------

func TestWarningsAloneNoExit1WithoutStrict(t *testing.T) {
	// Invariant entries without tags → 2 warnings
	spec := `# test-spec

## META
Deployment:   cli-tool
Version:      0.1.0
Spec-Schema:  0.1.0
Author:       Jane Example <jane@example.org>
License:      Apache-2.0
Verification: none
Safety-Level: QM

## TYPES

SomeType := string where non-empty

## BEHAVIOR: do-something

STEPS:
1. Do the thing.

## PRECONDITIONS
- input valid

## POSTCONDITIONS
- output produced

## INVARIANTS
- untagged entry one
- untagged entry two

## EXAMPLES

### EXAMPLE: happy-path
GIVEN:
  a valid input
WHEN:
  do-something is called
THEN:
  result = Ok
`
	path := writeTempSpec(t, "spec.md", spec)
	_, _, code := runBinary(t, path)
	if code != 0 {
		t.Errorf("expected exit 0 for warnings-only, got %d", code)
	}
}

// ---------------------------------------------------------------------------
// INVARIANT: stderr receives diagnostics; stdout receives summary
// ---------------------------------------------------------------------------

func TestStreamSeparation(t *testing.T) {
	// A spec with an error: diagnostic on stderr, summary on stdout
	spec := strings.ReplaceAll(minimalValidSpec(), "Version:      0.1.0", "Version:      bad")
	path := writeTempSpec(t, "spec.md", spec)
	stdout, stderr, _ := runBinary(t, path)

	// Diagnostic (ERROR) must be on stderr
	if !strings.Contains(stderr, "ERROR") {
		t.Errorf("expected ERROR on stderr, got: %q", stderr)
	}
	// Summary (✗) must be on stdout
	if !strings.Contains(stdout, "✗") {
		t.Errorf("expected ✗ summary on stdout, got: %q", stdout)
	}
	// Summary must NOT be on stderr
	if strings.Contains(stderr, "✗") {
		t.Errorf("summary must not appear on stderr, got: %q", stderr)
	}
	// Diagnostic must NOT be on stdout
	if strings.Contains(stdout, "ERROR") {
		t.Errorf("diagnostic must not appear on stdout, got: %q", stdout)
	}
}

// ---------------------------------------------------------------------------
// INVARIANT: diagnostic line numbers are monotonically non-decreasing
// ---------------------------------------------------------------------------

func TestDiagnosticLineNumbersMonotonic(t *testing.T) {
	// Spec with multiple errors at different lines
	spec := `# test-spec

## META
Deployment:   cli-tool
Version:      bad-version
Spec-Schema:  also-bad
Author:       Jane Example <jane@example.org>
License:      Not-A-Real-SPDX-ID-XYZ
Verification: none
Safety-Level: QM

## TYPES

## BEHAVIOR: do-something

STEPS:
1. Do the thing.

## PRECONDITIONS
- ok

## POSTCONDITIONS
- ok

## INVARIANTS
- [observable] ok

## EXAMPLES

### EXAMPLE: happy-path
GIVEN:
  input ok
WHEN:
  do-something called
THEN:
  result = Ok
`
	path := writeTempSpec(t, "spec.md", spec)
	_, stderr, _ := runBinary(t, path)
	// Parse line numbers from stderr lines
	lines := strings.Split(strings.TrimSpace(stderr), "\n")
	prevLine := -1
	for _, l := range lines {
		if l == "" {
			continue
		}
		// Format: SEVERITY  file:LINE  [section]  message
		// Extract the line number part
		parts := strings.Fields(l)
		if len(parts) < 2 {
			continue
		}
		// parts[1] should be "file:LINE"
		colonIdx := strings.LastIndex(parts[1], ":")
		if colonIdx < 0 {
			continue
		}
		var lineNum int
		fmt.Sscanf(parts[1][colonIdx+1:], "%d", &lineNum)
		if lineNum > 0 && lineNum < prevLine {
			t.Errorf("diagnostic line numbers not monotonic: %d after %d in line: %q", lineNum, prevLine, l)
		}
		if lineNum > 0 {
			prevLine = lineNum
		}
	}
}

// ---------------------------------------------------------------------------
// EXAMPLE: missing required file argument (no list-templates)
// ---------------------------------------------------------------------------

func TestMissingFileArgument(t *testing.T) {
	stdout, stderr, code := runBinary(t)
	if code != 2 {
		t.Errorf("expected exit 2 for missing file arg, got %d", code)
	}
	// Usage line written to stderr
	if stdout != "" {
		t.Errorf("expected empty stdout, got: %q", stdout)
	}
	_ = stderr // usage line present on stderr
}

// ---------------------------------------------------------------------------
// EXAMPLE: crypto-library deprecated deployment
// ---------------------------------------------------------------------------

func TestCryptoLibraryDeprecated(t *testing.T) {
	spec := strings.ReplaceAll(minimalValidSpec(), "Deployment:   cli-tool", "Deployment:   crypto-library")
	path := writeTempSpec(t, "spec.md", spec)
	_, stderr, code := runBinary(t, path)
	if code != 1 {
		t.Errorf("expected exit 1, got %d", code)
	}
	if !strings.Contains(stderr, "retired") && !strings.Contains(stderr, "crypto-library") {
		t.Errorf("expected crypto-library retired message, got: %q", stderr)
	}
}

// ---------------------------------------------------------------------------
// EXAMPLE: python-tool requires Safety-Level QM
// ---------------------------------------------------------------------------

func TestPythonToolRequiresQM(t *testing.T) {
	spec := `# test-spec

## META
Deployment:   python-tool
Version:      0.1.0
Spec-Schema:  0.1.0
Author:       Jane Example <jane@example.org>
License:      Apache-2.0
Verification: none
Safety-Level: ASIL-B

## TYPES

SomeType := string where non-empty

## BEHAVIOR: do-something

STEPS:
1. Do the thing.

## PRECONDITIONS
- input valid

## POSTCONDITIONS
- output produced

## INVARIANTS
- [observable] tool is idempotent

## EXAMPLES

### EXAMPLE: happy-path
GIVEN:
  a valid input
WHEN:
  do-something is called
THEN:
  result = Ok
`
	path := writeTempSpec(t, "spec.md", spec)
	_, stderr, code := runBinary(t, path)
	if code != 1 {
		t.Errorf("expected exit 1, got %d", code)
	}
	if !strings.Contains(stderr, "python-tool") && !strings.Contains(stderr, "Safety-Level") {
		t.Errorf("expected python-tool safety-level error, got: %q", stderr)
	}
}

// ---------------------------------------------------------------------------
// RULE-05: Unknown verification value → Warning
// ---------------------------------------------------------------------------

func TestUnknownVerificationValue(t *testing.T) {
	spec := strings.ReplaceAll(minimalValidSpec(), "Verification: none", "Verification: custom-engine")
	path := writeTempSpec(t, "spec.md", spec)
	_, stderr, code := runBinary(t, path)
	if code != 0 {
		t.Errorf("expected exit 0 for unknown verification (warning), got %d", code)
	}
	if !strings.Contains(stderr, "Unknown verification value") {
		t.Errorf("expected unknown verification warning, got: %q", stderr)
	}
}

// ---------------------------------------------------------------------------
// RULE-06: EXAMPLES section no example blocks → Error
// ---------------------------------------------------------------------------

func TestExamplesNoBlocks(t *testing.T) {
	spec := `# test-spec

## META
Deployment:   cli-tool
Version:      0.1.0
Spec-Schema:  0.1.0
Author:       Jane Example <jane@example.org>
License:      Apache-2.0
Verification: none
Safety-Level: QM

## TYPES

SomeType := string where non-empty

## BEHAVIOR: do-something

STEPS:
1. Do the thing.

## PRECONDITIONS
- input valid

## POSTCONDITIONS
- output produced

## INVARIANTS
- [observable] tool is idempotent

## EXAMPLES

No example blocks here, just some text.
`
	path := writeTempSpec(t, "spec.md", spec)
	_, stderr, code := runBinary(t, path)
	if code != 1 {
		t.Errorf("expected exit 1, got %d", code)
	}
	if !strings.Contains(stderr, "EXAMPLES section contains no example blocks") {
		t.Errorf("expected no-examples error, got: %q", stderr)
	}
}

// ---------------------------------------------------------------------------
// RULE-06: Flat EXAMPLE: header form rejected (v0.4.0)
// ---------------------------------------------------------------------------

func TestFlatExampleHeaderRejected(t *testing.T) {
	spec := `# test-spec

## META
Deployment:   cli-tool
Version:      0.1.0
Spec-Schema:  0.1.0
Author:       Jane Example <jane@example.org>
License:      Apache-2.0
Verification: none
Safety-Level: QM

## TYPES

SomeType := string where non-empty

## BEHAVIOR: do-something

STEPS:
1. Do the thing.

## PRECONDITIONS
- input valid

## POSTCONDITIONS
- output produced

## INVARIANTS
- [observable] tool is idempotent

## EXAMPLES

EXAMPLE: foo
GIVEN:
  some condition
WHEN:
  something
THEN:
  result = Ok
`
	path := writeTempSpec(t, "spec.md", spec)
	_, stderr, code := runBinary(t, path)
	if code != 1 {
		t.Errorf("expected exit 1, got %d", code)
	}
	if !strings.Contains(stderr, "Flat 'EXAMPLE: <name>' is no longer accepted") {
		t.Errorf("expected flat-form error, got: %q", stderr)
	}
}

// ---------------------------------------------------------------------------
// RULE-06: Wrong heading level (## EXAMPLE: or #### EXAMPLE:) → Error
// ---------------------------------------------------------------------------

func TestWrongExampleHeadingLevel(t *testing.T) {
	spec := `# test-spec

## META
Deployment:   cli-tool
Version:      0.1.0
Spec-Schema:  0.1.0
Author:       Jane Example <jane@example.org>
License:      Apache-2.0
Verification: none
Safety-Level: QM

## TYPES

SomeType := string where non-empty

## BEHAVIOR: do-something

STEPS:
1. Do the thing.

## PRECONDITIONS
- input valid

## POSTCONDITIONS
- output produced

## INVARIANTS
- [observable] tool is idempotent

## EXAMPLES

#### EXAMPLE: foo
GIVEN:
  some condition
WHEN:
  something
THEN:
  result = Ok
`
	path := writeTempSpec(t, "spec.md", spec)
	_, stderr, code := runBinary(t, path)
	if code != 1 {
		t.Errorf("expected exit 1, got %d", code)
	}
	if !strings.Contains(stderr, "heading level") {
		t.Errorf("expected heading level error, got: %q", stderr)
	}
}

// ---------------------------------------------------------------------------
// RULE-06: ### EXAMPLE: accepted (correct form)
// ---------------------------------------------------------------------------

func TestCorrectExampleHeadingAccepted(t *testing.T) {
	spec := `# test-spec

## META
Deployment:   cli-tool
Version:      0.1.0
Spec-Schema:  0.1.0
Author:       Jane Example <jane@example.org>
License:      Apache-2.0
Verification: none
Safety-Level: QM

## TYPES

SomeType := string where non-empty

## BEHAVIOR: do-something

STEPS:
1. Do the thing.

## PRECONDITIONS
- input valid

## POSTCONDITIONS
- output produced

## INVARIANTS
- [observable] tool is idempotent

## EXAMPLES

### EXAMPLE: foo
GIVEN:
  some condition
WHEN:
  something
THEN:
  result = Ok
`
	path := writeTempSpec(t, "spec.md", spec)
	_, stderr, code := runBinary(t, path)
	if code != 0 {
		t.Errorf("expected exit 0, got %d; stderr=%q", code, stderr)
	}
	if strings.Contains(stderr, "heading") {
		t.Errorf("unexpected heading error, got: %q", stderr)
	}
}

// ---------------------------------------------------------------------------
// RULE-15: MILESTONE section structure — valid scaffold first
// ---------------------------------------------------------------------------

func TestMilestoneValidScaffoldFirst(t *testing.T) {
	spec := `# test-spec

## META
Deployment:   cli-tool
Version:      0.1.0
Spec-Schema:  0.1.0
Author:       Jane Example <jane@example.org>
License:      Apache-2.0
Verification: none
Safety-Level: QM

## TYPES

SomeType := string where non-empty

## BEHAVIOR: collect

STEPS:
1. Collect data.

## BEHAVIOR: render

STEPS:
1. Render output.

## BEHAVIOR: main

STEPS:
1. Main entry.

## PRECONDITIONS
- input valid

## POSTCONDITIONS
- output produced

## INVARIANTS
- [observable] tool is idempotent

## EXAMPLES

### EXAMPLE: happy-path
GIVEN:
  a valid input
WHEN:
  collect is called
THEN:
  result = Ok

## MILESTONE: 0.0.0
Status: released
Scaffold: true
Included BEHAVIORs: collect, render, main
Acceptance criteria:
  ./tool version | grep -q "^tool "

## MILESTONE: 0.1.0
Status: active
Included BEHAVIORs: collect
Deferred BEHAVIORs: render, main
Acceptance criteria:
  ./tool collect | jq '.result | length > 0'
`
	path := writeTempSpec(t, "spec.md", spec)
	_, stderr, code := runBinary(t, path)
	if code != 0 {
		t.Errorf("expected exit 0, got %d; stderr=%q", code, stderr)
	}
	if stderr != "" {
		t.Errorf("expected empty stderr, got: %q", stderr)
	}
}

// ---------------------------------------------------------------------------
// EXAMPLE: milestone_scaffold_not_first
// ---------------------------------------------------------------------------

func TestMilestoneScaffoldNotFirst(t *testing.T) {
	spec := `# test-spec

## META
Deployment:   cli-tool
Version:      0.1.0
Spec-Schema:  0.1.0
Author:       Jane Example <jane@example.org>
License:      Apache-2.0
Verification: none
Safety-Level: QM

## TYPES

SomeType := string where non-empty

## BEHAVIOR: collect

STEPS:
1. Collect data.

## BEHAVIOR: render

STEPS:
1. Render output.

## PRECONDITIONS
- input valid

## POSTCONDITIONS
- output produced

## INVARIANTS
- [observable] tool is idempotent

## EXAMPLES

### EXAMPLE: happy-path
GIVEN:
  a valid input
WHEN:
  collect is called
THEN:
  result = Ok

## MILESTONE: 0.1.0
Status: released
Included BEHAVIORs: collect
Deferred BEHAVIORs: render

## MILESTONE: 0.2.0
Status: active
Scaffold: true
Included BEHAVIORs: collect, render
Deferred BEHAVIORs:
`
	path := writeTempSpec(t, "spec.md", spec)
	_, stderr, code := runBinary(t, path)
	if code != 1 {
		t.Errorf("expected exit 1, got %d", code)
	}
	if !strings.Contains(stderr, "must appear first") {
		t.Errorf("expected scaffold-not-first error, got: %q", stderr)
	}
}

// ---------------------------------------------------------------------------
// EXAMPLE: milestone_two_scaffold_rejected
// ---------------------------------------------------------------------------

func TestMilestoneTwoScaffoldRejected(t *testing.T) {
	spec := `# test-spec

## META
Deployment:   cli-tool
Version:      0.1.0
Spec-Schema:  0.1.0
Author:       Jane Example <jane@example.org>
License:      Apache-2.0
Verification: none
Safety-Level: QM

## TYPES

SomeType := string where non-empty

## BEHAVIOR: collect

STEPS:
1. Collect data.

## BEHAVIOR: render

STEPS:
1. Render output.

## PRECONDITIONS
- input valid

## POSTCONDITIONS
- output produced

## INVARIANTS
- [observable] tool is idempotent

## EXAMPLES

### EXAMPLE: happy-path
GIVEN:
  a valid input
WHEN:
  collect is called
THEN:
  result = Ok

## MILESTONE: 0.0.0
Status: released
Scaffold: true
Included BEHAVIORs: collect, render

## MILESTONE: 0.1.0
Status: active
Scaffold: true
Included BEHAVIORs: collect, render
Deferred BEHAVIORs:
`
	path := writeTempSpec(t, "spec.md", spec)
	_, stderr, code := runBinary(t, path)
	if code != 1 {
		t.Errorf("expected exit 1, got %d", code)
	}
	if !strings.Contains(stderr, "More than one MILESTONE has Scaffold: true") {
		t.Errorf("expected two-scaffold error, got: %q", stderr)
	}
}

// ---------------------------------------------------------------------------
// EXAMPLE: milestone_two_active_rejected
// ---------------------------------------------------------------------------

func TestMilestoneTwoActiveRejected(t *testing.T) {
	spec := `# test-spec

## META
Deployment:   cli-tool
Version:      0.1.0
Spec-Schema:  0.1.0
Author:       Jane Example <jane@example.org>
License:      Apache-2.0
Verification: none
Safety-Level: QM

## TYPES

SomeType := string where non-empty

## BEHAVIOR: collect

STEPS:
1. Collect data.

## BEHAVIOR: render

STEPS:
1. Render output.

## PRECONDITIONS
- input valid

## POSTCONDITIONS
- output produced

## INVARIANTS
- [observable] tool is idempotent

## EXAMPLES

### EXAMPLE: happy-path
GIVEN:
  a valid input
WHEN:
  collect is called
THEN:
  result = Ok

## MILESTONE: 0.1.0
Status: active
Included BEHAVIORs: collect
Deferred BEHAVIORs: render

## MILESTONE: 0.2.0
Status: active
Included BEHAVIORs: render
Deferred BEHAVIORs:
`
	path := writeTempSpec(t, "spec.md", spec)
	_, stderr, code := runBinary(t, path)
	if code != 1 {
		t.Errorf("expected exit 1, got %d", code)
	}
	if !strings.Contains(stderr, "More than one MILESTONE has Status: active") {
		t.Errorf("expected two-active error, got: %q", stderr)
	}
}

// ---------------------------------------------------------------------------
// EXAMPLE: milestone_unknown_behavior_name
// ---------------------------------------------------------------------------

func TestMilestoneUnknownBehaviorName(t *testing.T) {
	spec := `# test-spec

## META
Deployment:   cli-tool
Version:      0.1.0
Spec-Schema:  0.1.0
Author:       Jane Example <jane@example.org>
License:      Apache-2.0
Verification: none
Safety-Level: QM

## TYPES

SomeType := string where non-empty

## BEHAVIOR: collect

STEPS:
1. Collect data.

## BEHAVIOR: render

STEPS:
1. Render output.

## PRECONDITIONS
- input valid

## POSTCONDITIONS
- output produced

## INVARIANTS
- [observable] tool is idempotent

## EXAMPLES

### EXAMPLE: happy-path
GIVEN:
  a valid input
WHEN:
  collect is called
THEN:
  result = Ok

## MILESTONE: 0.1.0
Status: active
Included BEHAVIORs: collect, nonexistent-behavior
Deferred BEHAVIORs: render
Acceptance criteria:
  ./tool collect
`
	path := writeTempSpec(t, "spec.md", spec)
	_, stderr, code := runBinary(t, path)
	if code != 1 {
		t.Errorf("expected exit 1, got %d", code)
	}
	if !strings.Contains(stderr, "nonexistent-behavior") {
		t.Errorf("expected unknown behavior name in error, got: %q", stderr)
	}
}

// ---------------------------------------------------------------------------
// RULE-04: deprecated Domain field → Warning
// ---------------------------------------------------------------------------

func TestDeprecatedDomainField(t *testing.T) {
	spec := `# test-spec

## META
Deployment:   cli-tool
Version:      0.1.0
Spec-Schema:  0.1.0
Author:       Jane Example <jane@example.org>
License:      Apache-2.0
Verification: none
Safety-Level: QM
Domain:       tools

## TYPES

SomeType := string where non-empty

## BEHAVIOR: do-something

STEPS:
1. Do the thing.

## PRECONDITIONS
- input valid

## POSTCONDITIONS
- output produced

## INVARIANTS
- [observable] tool is idempotent

## EXAMPLES

### EXAMPLE: happy-path
GIVEN:
  a valid input
WHEN:
  do-something is called
THEN:
  result = Ok
`
	path := writeTempSpec(t, "spec.md", spec)
	_, stderr, code := runBinary(t, path)
	if code != 0 {
		t.Errorf("expected exit 0, got %d", code)
	}
	if !strings.Contains(stderr, "deprecated") || !strings.Contains(stderr, "Domain") {
		t.Errorf("expected Domain deprecated warning, got: %q", stderr)
	}
}

// ---------------------------------------------------------------------------
// RULE-03: verified-library with Safety-Level QM → Warning
// ---------------------------------------------------------------------------

func TestVerifiedLibraryQMWarning(t *testing.T) {
	spec := `# test-spec

## META
Deployment:   verified-library
Version:      0.1.0
Spec-Schema:  0.1.0
Author:       Jane Example <jane@example.org>
License:      Apache-2.0
Verification: none
Safety-Level: QM

## TYPES

SomeType := string where non-empty

## BEHAVIOR: do-something

STEPS:
1. Do the thing.

## PRECONDITIONS
- input valid

## POSTCONDITIONS
- output produced

## INVARIANTS
- [observable] tool is idempotent

## EXAMPLES

### EXAMPLE: happy-path
GIVEN:
  a valid input
WHEN:
  do-something is called
THEN:
  result = Ok
`
	path := writeTempSpec(t, "spec.md", spec)
	_, stderr, code := runBinary(t, path)
	if code != 0 {
		t.Errorf("expected exit 0 for warning, got %d", code)
	}
	if !strings.Contains(stderr, "verified-library") {
		t.Errorf("expected verified-library warning, got: %q", stderr)
	}
}

// ---------------------------------------------------------------------------
// Diagnostic format check: ERROR  file:line  [section]  message
// ---------------------------------------------------------------------------

func TestDiagnosticFormat(t *testing.T) {
	spec := strings.ReplaceAll(minimalValidSpec(), "Version:      0.1.0", "Version:      bad")
	path := writeTempSpec(t, "spec.md", spec)
	_, stderr, _ := runBinary(t, path)
	// Must match: "ERROR  filename:LINE  [section]  message"
	lines := strings.Split(strings.TrimSpace(stderr), "\n")
	found := false
	for _, l := range lines {
		if strings.HasPrefix(l, "ERROR") {
			// Check basic structure: ERROR, file:line, [section], message
			parts := strings.Fields(l)
			if len(parts) < 4 {
				t.Errorf("diagnostic line too short: %q", l)
			}
			if parts[0] != "ERROR" {
				t.Errorf("expected ERROR severity, got: %q", parts[0])
			}
			// parts[1] should be file:line
			if !strings.Contains(parts[1], ":") {
				t.Errorf("expected file:line in parts[1], got: %q", parts[1])
			}
			// parts[2] should be [section]
			if !strings.HasPrefix(parts[2], "[") || !strings.HasSuffix(parts[2], "]") {
				t.Errorf("expected [section] in parts[2], got: %q", parts[2])
			}
			found = true
		}
	}
	if !found {
		t.Errorf("no ERROR diagnostic found in stderr: %q", stderr)
	}
}

// ---------------------------------------------------------------------------
// INVARIANT: errors always produce exit_code >= 1
// ---------------------------------------------------------------------------

func TestErrorsAlwaysExitNonZero(t *testing.T) {
	// A spec missing multiple required sections
	spec := `# minimal

## META
Deployment: cli-tool
Version: 0.1.0
Spec-Schema: 0.1.0
Author: Test <t@t.org>
License: Apache-2.0
Verification: none
Safety-Level: QM
`
	path := writeTempSpec(t, "spec.md", spec)
	_, _, code := runBinary(t, path)
	if code == 0 {
		t.Errorf("expected non-zero exit for spec with errors, got 0")
	}
}

// ---------------------------------------------------------------------------
// INVARIANT: exit_code=0 never when Error diagnostics present
// ---------------------------------------------------------------------------

func TestNoExit0WhenErrorPresent(t *testing.T) {
	// Invalid SPDX → Error → must not exit 0 regardless of strict
	spec := strings.ReplaceAll(minimalValidSpec(), "License:      Apache-2.0", "License:      TOTALLY-INVALID-XYZ")
	path := writeTempSpec(t, "spec.md", spec)
	_, _, code := runBinary(t, path)
	if code == 0 {
		t.Errorf("expected non-zero exit when Error present, got 0")
	}
}
