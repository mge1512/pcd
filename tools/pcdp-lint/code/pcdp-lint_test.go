package main

import (
	"strings"
	"testing"
)

// Test: valid minimal spec
func TestValidMinimalSpec(t *testing.T) {
	spec := `# pcdp-lint

## META
Deployment:   cli-tool
Version:      0.1.0
Spec-Schema:  0.1.0
Author:       Jane Example <jane@example.org>
License:      Apache-2.0
Verification: none
Safety-Level: QM

## TYPES
SpecFile := path

## BEHAVIOR: lint
Constraint: required

STEPS:
1. Validate file

## PRECONDITIONS
- file exists

## POSTCONDITIONS
- file is validated

## INVARIANTS
- [observable] tool is idempotent

## EXAMPLES

EXAMPLE: valid_minimal_spec
GIVEN:
  file contains all required sections
WHEN:
  lint is run
THEN:
  exit_code = 0
`
	result := lintFile("test.md", spec, false)
	if result.ExitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", result.ExitCode)
	}
	if len(result.Diagnostics) > 0 {
		t.Errorf("Expected no diagnostics, got %d", len(result.Diagnostics))
		for _, d := range result.Diagnostics {
			t.Logf("Diagnostic: %s - %s", d.Severity, d.Message)
		}
	}
}

// Test: missing required section
func TestMissingRequiredSection(t *testing.T) {
	spec := `# pcdp-lint

## META
Deployment: cli-tool

## TYPES
Type := string

## BEHAVIOR: test
STEPS:
1. Test

## PRECONDITIONS
- none

## POSTCONDITIONS
- none

## EXAMPLES
EXAMPLE: test
GIVEN: input
WHEN: run
THEN: output
`
	result := lintFile("test.md", spec, false)
	if result.ExitCode == 0 {
		t.Errorf("Expected non-zero exit code for missing INVARIANTS, got %d", result.ExitCode)
	}
	hasError := false
	for _, d := range result.Diagnostics {
		if d.Severity == ERROR && strings.Contains(d.Message, "INVARIANTS") {
			hasError = true
			break
		}
	}
	if !hasError {
		t.Error("Expected error for missing INVARIANTS section")
	}
}

// Test: invalid SPDX license
func TestInvalidSPDXLicense(t *testing.T) {
	spec := `# test

## META
Deployment:   cli-tool
Version:      0.1.0
Spec-Schema:  0.1.0
Author:       Test Author <test@example.org>
License:      MIT License
Verification: none
Safety-Level: QM

## TYPES
Type := string

## BEHAVIOR: test
STEPS:
1. Test

## PRECONDITIONS
- none

## POSTCONDITIONS
- none

## INVARIANTS
- [observable] test

## EXAMPLES
EXAMPLE: test
GIVEN: input
WHEN: run
THEN: output
`
	result := lintFile("test.md", spec, false)
	if result.ExitCode == 0 {
		t.Errorf("Expected non-zero exit code for invalid license, got %d", result.ExitCode)
	}
	hasError := false
	for _, d := range result.Diagnostics {
		if d.Severity == ERROR && strings.Contains(d.Message, "License") {
			hasError = true
			break
		}
	}
	if !hasError {
		t.Error("Expected error for invalid SPDX license")
	}
}

// Test: invalid version format
func TestInvalidVersionFormat(t *testing.T) {
	spec := `# test

## META
Deployment:   cli-tool
Version:      1.0
Spec-Schema:  0.1.0
Author:       Test Author <test@example.org>
License:      Apache-2.0
Verification: none
Safety-Level: QM

## TYPES
Type := string

## BEHAVIOR: test
STEPS:
1. Test

## PRECONDITIONS
- none

## POSTCONDITIONS
- none

## INVARIANTS
- [observable] test

## EXAMPLES
EXAMPLE: test
GIVEN: input
WHEN: run
THEN: output
`
	result := lintFile("test.md", spec, false)
	if result.ExitCode == 0 {
		t.Errorf("Expected non-zero exit code for invalid version, got %d", result.ExitCode)
	}
	hasError := false
	for _, d := range result.Diagnostics {
		if d.Severity == ERROR && strings.Contains(d.Message, "Version") {
			hasError = true
			break
		}
	}
	if !hasError {
		t.Error("Expected error for invalid version format")
	}
}

// Test: missing author
func TestMissingAuthor(t *testing.T) {
	spec := `# test

## META
Deployment:   cli-tool
Version:      0.1.0
Spec-Schema:  0.1.0
License:      Apache-2.0
Verification: none
Safety-Level: QM

## TYPES
Type := string

## BEHAVIOR: test
STEPS:
1. Test

## PRECONDITIONS
- none

## POSTCONDITIONS
- none

## INVARIANTS
- [observable] test

## EXAMPLES
EXAMPLE: test
GIVEN: input
WHEN: run
THEN: output
`
	result := lintFile("test.md", spec, false)
	if result.ExitCode == 0 {
		t.Errorf("Expected non-zero exit code for missing author, got %d", result.ExitCode)
	}
	hasError := false
	for _, d := range result.Diagnostics {
		if d.Severity == ERROR && strings.Contains(d.Message, "Author") {
			hasError = true
			break
		}
	}
	if !hasError {
		t.Error("Expected error for missing author")
	}
}

// Test: unknown deployment template
func TestUnknownDeploymentTemplate(t *testing.T) {
	spec := `# test

## META
Deployment:   serverless
Version:      0.1.0
Spec-Schema:  0.1.0
Author:       Test Author <test@example.org>
License:      Apache-2.0
Verification: none
Safety-Level: QM

## TYPES
Type := string

## BEHAVIOR: test
STEPS:
1. Test

## PRECONDITIONS
- none

## POSTCONDITIONS
- none

## INVARIANTS
- [observable] test

## EXAMPLES
EXAMPLE: test
GIVEN: input
WHEN: run
THEN: output
`
	result := lintFile("test.md", spec, false)
	if result.ExitCode == 0 {
		t.Errorf("Expected non-zero exit code for unknown deployment, got %d", result.ExitCode)
	}
	hasError := false
	for _, d := range result.Diagnostics {
		if d.Severity == ERROR && strings.Contains(d.Message, "Unknown deployment") {
			hasError = true
			break
		}
	}
	if !hasError {
		t.Error("Expected error for unknown deployment template")
	}
}

// Test: deprecated Target field
func TestDeprecatedTargetField(t *testing.T) {
	spec := `# test

## META
Deployment:   backend-service
Version:      0.1.0
Spec-Schema:  0.1.0
Author:       Test Author <test@example.org>
License:      Apache-2.0
Target:       Go
Verification: none
Safety-Level: QM

## TYPES
Type := string

## BEHAVIOR: test
STEPS:
1. Test

## PRECONDITIONS
- none

## POSTCONDITIONS
- none

## INVARIANTS
- [observable] test

## EXAMPLES
EXAMPLE: test
GIVEN: input
WHEN: run
THEN: output
`
	result := lintFile("test.md", spec, false)
	if result.ExitCode != 0 {
		// Should have warning but still exit 0
		t.Logf("Exit code: %d", result.ExitCode)
	}
	hasWarning := false
	for _, d := range result.Diagnostics {
		if d.Severity == WARNING && strings.Contains(d.Message, "Target") {
			hasWarning = true
			break
		}
	}
	if !hasWarning {
		t.Error("Expected warning for deprecated Target field")
	}
}

// Test: behavior missing STEPS
func TestBehaviorMissingSTEPS(t *testing.T) {
	spec := `# test

## META
Deployment:   cli-tool
Version:      0.1.0
Spec-Schema:  0.1.0
Author:       Test Author <test@example.org>
License:      Apache-2.0
Verification: none
Safety-Level: QM

## TYPES
Type := string

## BEHAVIOR: test
PRECONDITIONS:
- none

## PRECONDITIONS
- none

## POSTCONDITIONS
- none

## INVARIANTS
- [observable] test

## EXAMPLES
EXAMPLE: test
GIVEN: input
WHEN: run
THEN: output
`
	result := lintFile("test.md", spec, false)
	if result.ExitCode == 0 {
		t.Errorf("Expected non-zero exit code for missing STEPS, got %d", result.ExitCode)
	}
	hasError := false
	for _, d := range result.Diagnostics {
		if d.Severity == ERROR && strings.Contains(d.Message, "STEPS") {
			hasError = true
			break
		}
	}
	if !hasError {
		t.Error("Expected error for missing STEPS in BEHAVIOR")
	}
}

// Test: multiple authors valid
func TestMultipleAuthorsValid(t *testing.T) {
	spec := `# test

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
Type := string

## BEHAVIOR: test
STEPS:
1. Test

## PRECONDITIONS
- none

## POSTCONDITIONS
- none

## INVARIANTS
- [observable] test

## EXAMPLES
EXAMPLE: test
GIVEN: input
WHEN: run
THEN: output
`
	result := lintFile("test.md", spec, false)
	if result.ExitCode != 0 {
		t.Errorf("Expected exit code 0 for multiple authors, got %d", result.ExitCode)
	}
}

// Test: compound SPDX license expression
func TestCompoundSPDXLicense(t *testing.T) {
	spec := `# test

## META
Deployment:   cli-tool
Version:      0.1.0
Spec-Schema:  0.1.0
Author:       Test Author <test@example.org>
License:      Apache-2.0 OR MIT
Verification: none
Safety-Level: QM

## TYPES
Type := string

## BEHAVIOR: test
STEPS:
1. Test

## PRECONDITIONS
- none

## POSTCONDITIONS
- none

## INVARIANTS
- [observable] test

## EXAMPLES
EXAMPLE: test
GIVEN: input
WHEN: run
THEN: output
`
	result := lintFile("test.md", spec, false)
	hasLicenseError := false
	for _, d := range result.Diagnostics {
		if d.Severity == ERROR && strings.Contains(d.Message, "License") {
			hasLicenseError = true
			break
		}
	}
	if hasLicenseError {
		t.Error("Expected no error for compound SPDX license")
	}
}

// Test: strict mode with warnings
func TestStrictModeWithWarnings(t *testing.T) {
	spec := `# test

## META
Deployment:   backend-service
Version:      0.1.0
Spec-Schema:  0.1.0
Author:       Test Author <test@example.org>
License:      Apache-2.0
Target:       Go
Verification: none
Safety-Level: QM

## TYPES
Type := string

## BEHAVIOR: test
STEPS:
1. Test

## PRECONDITIONS
- none

## POSTCONDITIONS
- none

## INVARIANTS
- [observable] test

## EXAMPLES
EXAMPLE: test
GIVEN: input
WHEN: run
THEN: output
`
	result := lintFile("test.md", spec, true)
	if result.ExitCode == 0 {
		t.Errorf("Expected non-zero exit code in strict mode with warnings, got %d", result.ExitCode)
	}
}

// Test: fenced code blocks are ignored
func TestFencedCodeBlocksIgnored(t *testing.T) {
	spec := `# test

## META
Deployment:   cli-tool
Version:      0.1.0
Spec-Schema:  0.1.0
Author:       Test Author <test@example.org>
License:      Apache-2.0
Verification: none
Safety-Level: QM

## TYPES
Type := string

## BEHAVIOR: test
STEPS:
1. Test

## PRECONDITIONS
- none

## POSTCONDITIONS
- none

## INVARIANTS
- [observable] test

## EXAMPLES

EXAMPLE: test
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
	result := lintFile("test.md", spec, false)
	if result.ExitCode != 0 {
		t.Errorf("Expected exit code 0 (fenced blocks ignored), got %d", result.ExitCode)
		for _, d := range result.Diagnostics {
			t.Logf("Diagnostic: %s - %s", d.Severity, d.Message)
		}
	}
}

// Test: semantic version validation
func TestSemanticVersionValidation(t *testing.T) {
	tests := []struct {
		version string
		valid   bool
	}{
		{"0.1.0", true},
		{"1.2.3", true},
		{"10.20.30", true},
		{"0.1", false},
		{"1.0.0.0", false},
		{"1.0.0-rc1", false},
		{"v1.0.0", false},
	}

	for _, test := range tests {
		result := isSemanticVersion(test.version)
		if result != test.valid {
			t.Errorf("isSemanticVersion('%s') = %v, want %v", test.version, result, test.valid)
		}
	}
}

// Test: SPDX license validation
func TestSPDXLicenseValidation(t *testing.T) {
	tests := []struct {
		license string
		valid   bool
	}{
		{"Apache-2.0", true},
		{"MIT", true},
		{"GPL-2.0-only", true},
		{"Apache-2.0 OR MIT", true},
		{"MIT License", false},
		{"Apache 2.0", false},
		{"GPL-2", false},
	}

	for _, test := range tests {
		result := isValidSPDXLicense(test.license)
		if result != test.valid {
			t.Errorf("isValidSPDXLicense('%s') = %v, want %v", test.license, result, test.valid)
		}
	}
}
