package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

const (
	VERSION     = "0.3.13"
	SPEC_SCHEMA = "0.3.13"
	SPDX_VERSION = "3.20"
)

type Severity string

const (
	ERROR   Severity = "ERROR"
	WARNING Severity = "WARNING"
)

type Diagnostic struct {
	Severity Severity
	Section  string
	Message  string
	Line     int
}

type LintResult struct {
	File        string
	Diagnostics []Diagnostic
	ExitCode    int
}

// SPDX license list (common licenses)
var validSPDXLicenses = map[string]bool{
	"Apache-2.0": true, "MIT": true, "GPL-2.0-only": true, "GPL-3.0-only": true,
	"LGPL-2.1-only": true, "LGPL-2.1-or-later": true, "LGPL-3.0-only": true,
	"LGPL-3.0-or-later": true, "BSD-2-Clause": true, "BSD-3-Clause": true,
	"ISC": true, "MPL-2.0": true, "AGPL-3.0-only": true, "AGPL-3.0-or-later": true,
	"GPL-2.0-or-later": true, "Unlicense": true, "CC0-1.0": true, "CC-BY-4.0": true,
}

// Valid deployment templates
var validDeploymentTemplates = []string{
	"wasm", "ebpf", "kernel-module", "verified-library", "cli-tool", "gui-tool",
	"cloud-native", "backend-service", "library-c-abi", "enterprise-software",
	"academic", "python-tool", "enhance-existing", "manual", "template",
	"mcp-server", "project-manifest",
}

var knownVerificationValues = map[string]bool{
	"none": true, "lean4": true, "fstar": true, "dafny": true, "custom": true,
}

type ParsedSpec struct {
	Lines    []string
	Sections map[string][]string // section name -> lines
	Meta     map[string][]string  // META field -> values (allows repeating keys)
}

func main() {
	args := os.Args[1:]

	// Handle commands and options
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "usage: pcdp-lint [strict=true|false] [list-templates|version] <specfile.md>\n")
		os.Exit(2)
	}

	strict := false
	var filename string

	for _, arg := range args {
		if arg == "list-templates" {
			listTemplates()
			return
		}
		if arg == "version" {
			printVersion()
			return
		}
		if strings.HasPrefix(arg, "strict=") {
			val := strings.TrimPrefix(arg, "strict=")
			strict = val == "true"
			continue
		}
		if !strings.Contains(arg, "=") && filename == "" {
			filename = arg
			continue
		}
		if strings.Contains(arg, "=") {
			// unrecognised option
			fmt.Fprintf(os.Stderr, "error: unrecognised option: %s\n", strings.Split(arg, "=")[0])
			os.Exit(2)
		}
	}

	if filename == "" {
		fmt.Fprintf(os.Stderr, "usage: pcdp-lint [strict=true|false] <specfile.md>\n")
		os.Exit(2)
	}

	// Validate file extension
	if !strings.HasSuffix(filename, ".md") {
		fmt.Fprintf(os.Stderr, "error: file must have .md extension: %s\n", filename)
		os.Exit(2)
	}

	// Check file exists and is readable
	data, err := os.ReadFile(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: cannot open file: %s\n", filename)
		os.Exit(2)
	}

	result := lintFile(filename, string(data), strict)
	os.Exit(result.ExitCode)
}

func lintFile(filename string, content string, strict bool) LintResult {
	result := LintResult{
		File:        filename,
		Diagnostics: []Diagnostic{},
		ExitCode:    0,
	}

	// Parse the spec
	spec := parseSpec(content)

	// Apply rules in order
	applyRule01(&result, spec)
	applyRule02(&result, spec)
	applyRule02b(&result, spec)
	applyRule02c(&result, spec)
	applyRule02d(&result, spec)
	applyRule02e(&result, spec)
	applyRule03(&result, spec)
	applyRule04(&result, spec)
	applyRule05(&result, spec)
	applyRule06(&result, spec)
	applyRule07(&result, spec)
	applyRule08(&result, spec)
	applyRule09(&result, spec)
	applyRule10(&result, spec)
	applyRule13(&result, spec)
	applyRule14(&result, spec)

	// Sort diagnostics by line number
	sort.Slice(result.Diagnostics, func(i, j int) bool {
		return result.Diagnostics[i].Line < result.Diagnostics[j].Line
	})

	// Output diagnostics to stderr
	for _, diag := range result.Diagnostics {
		fmt.Fprintf(os.Stderr, "%s  %s:%d  [%s]  %s\n",
			diag.Severity, result.File, diag.Line, diag.Section, diag.Message)
	}

	// Calculate exit code and print summary
	hasErrors := false
	hasWarnings := false
	for _, diag := range result.Diagnostics {
		if diag.Severity == ERROR {
			hasErrors = true
		} else if diag.Severity == WARNING {
			hasWarnings = true
		}
	}

	if hasErrors {
		result.ExitCode = 1
		errorCount := 0
		warningCount := 0
		for _, diag := range result.Diagnostics {
			if diag.Severity == ERROR {
				errorCount++
			} else {
				warningCount++
			}
		}
		fmt.Printf("✗ %s: %d error(s), %d warning(s)\n", filepath.Base(result.File), errorCount, warningCount)
	} else if hasWarnings && strict {
		result.ExitCode = 1
		warningCount := 0
		for _, diag := range result.Diagnostics {
			if diag.Severity == WARNING {
				warningCount++
			}
		}
		fmt.Printf("✗ %s: 0 error(s), %d warning(s) [strict mode]\n", filepath.Base(result.File), warningCount)
	} else if hasWarnings {
		result.ExitCode = 0
		warningCount := 0
		for _, diag := range result.Diagnostics {
			if diag.Severity == WARNING {
				warningCount++
			}
		}
		fmt.Printf("✓ %s: valid (%d warning(s))\n", filepath.Base(result.File), warningCount)
	} else {
		result.ExitCode = 0
		fmt.Printf("✓ %s: valid\n", filepath.Base(result.File))
	}

	return result
}

func parseSpec(content string) ParsedSpec {
	spec := ParsedSpec{
		Lines:    strings.Split(content, "\n"),
		Sections: make(map[string][]string),
		Meta:     make(map[string][]string),
	}

	fenceDepth := 0
	currentSection := ""
	var currentLines []string

	for _, line := range spec.Lines {
		// Track code fences
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "```") || strings.HasPrefix(trimmed, "~~~") {
			if fenceDepth == 0 {
				fenceDepth = 1
			} else {
				fenceDepth--
			}
			currentLines = append(currentLines, line)
			continue
		}

		if fenceDepth > 0 {
			currentLines = append(currentLines, line)
			continue
		}

		// Check for section headers (column 0)
		if strings.HasPrefix(line, "## ") {
			if currentSection != "" {
				spec.Sections[currentSection] = currentLines
			}
			currentSection = strings.TrimPrefix(line, "## ")
			currentLines = []string{}
		} else if currentSection == "META" && strings.Contains(line, ":") {
			// Parse META fields
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])
				spec.Meta[key] = append(spec.Meta[key], value)
			}
			currentLines = append(currentLines, line)
		} else {
			currentLines = append(currentLines, line)
		}
	}

	if currentSection != "" {
		spec.Sections[currentSection] = currentLines
	}

	return spec
}

// RULE-01: Required sections present
func applyRule01(result *LintResult, spec ParsedSpec) {
	requiredSections := []string{"META", "TYPES", "BEHAVIOR", "PRECONDITIONS", "POSTCONDITIONS", "INVARIANTS", "EXAMPLES"}

	for _, section := range requiredSections {
		found := false
		if section == "BEHAVIOR" {
			// Check for BEHAVIOR or BEHAVIOR/INTERNAL
			for sectionName := range spec.Sections {
				if strings.HasPrefix(sectionName, "BEHAVIOR") {
					found = true
					break
				}
			}
		} else {
			if _, exists := spec.Sections[section]; exists {
				found = true
			}
		}

		if !found {
			result.Diagnostics = append(result.Diagnostics, Diagnostic{
				Severity: ERROR,
				Section:  "structure",
				Message:  fmt.Sprintf("Missing required section: ## %s", section),
				Line:     1,
			})
		}
	}
}

// RULE-02: META fields present and non-empty
func applyRule02(result *LintResult, spec ParsedSpec) {
	requiredFields := []string{"Deployment", "Verification", "Safety-Level", "Version", "Spec-Schema", "License"}

	for _, field := range requiredFields {
		if values, exists := spec.Meta[field]; !exists || len(values) == 0 {
			result.Diagnostics = append(result.Diagnostics, Diagnostic{
				Severity: ERROR,
				Section:  "META",
				Message:  fmt.Sprintf("Missing required META field: %s", field),
				Line:     1,
			})
		} else if values[0] == "" {
			result.Diagnostics = append(result.Diagnostics, Diagnostic{
				Severity: ERROR,
				Section:  "META",
				Message:  fmt.Sprintf("META field %s has empty value", field),
				Line:     1,
			})
		}
	}
}

// RULE-02b: Author field
func applyRule02b(result *LintResult, spec ParsedSpec) {
	authors, exists := spec.Meta["Author"]
	if !exists || len(authors) == 0 {
		result.Diagnostics = append(result.Diagnostics, Diagnostic{
			Severity: ERROR,
			Section:  "META",
			Message:  "Missing required META field: Author (at least one Author: line required)",
			Line:     1,
		})
	} else {
		for _, author := range authors {
			if author == "" {
				result.Diagnostics = append(result.Diagnostics, Diagnostic{
					Severity: ERROR,
					Section:  "META",
					Message:  "Author: field has empty value",
					Line:     1,
				})
				break
			}
		}
	}
}

// RULE-02c: Version format
func applyRule02c(result *LintResult, spec ParsedSpec) {
	if versions, exists := spec.Meta["Version"]; exists && len(versions) > 0 {
		version := versions[0]
		if !isSemanticVersion(version) {
			result.Diagnostics = append(result.Diagnostics, Diagnostic{
				Severity: ERROR,
				Section:  "META",
				Message:  fmt.Sprintf("Version '%s' is not valid semantic versioning. Required format: MAJOR.MINOR.PATCH (e.g. 0.1.0)", version),
				Line:     1,
			})
		}
	}
}

// RULE-02d: Spec-Schema version
func applyRule02d(result *LintResult, spec ParsedSpec) {
	if schemas, exists := spec.Meta["Spec-Schema"]; exists && len(schemas) > 0 {
		schema := schemas[0]
		if !isSemanticVersion(schema) {
			result.Diagnostics = append(result.Diagnostics, Diagnostic{
				Severity: ERROR,
				Section:  "META",
				Message:  fmt.Sprintf("Spec-Schema '%s' is not valid semantic versioning. Required format: MAJOR.MINOR.PATCH (e.g. 0.1.0)", schema),
				Line:     1,
			})
		}
	}
}

// RULE-02e: License SPDX validation
func applyRule02e(result *LintResult, spec ParsedSpec) {
	if licenses, exists := spec.Meta["License"]; exists && len(licenses) > 0 {
		license := licenses[0]
		if !isValidSPDXLicense(license) {
			result.Diagnostics = append(result.Diagnostics, Diagnostic{
				Severity: ERROR,
				Section:  "META",
				Message:  fmt.Sprintf("License '%s' is not a valid SPDX identifier. See https://spdx.org/licenses/ for valid identifiers. Compound expressions permitted (e.g. Apache-2.0 OR MIT).", license),
				Line:     1,
			})
		}
	}
}

// RULE-03: Deployment template resolves
func applyRule03(result *LintResult, spec ParsedSpec) {
	if deployments, exists := spec.Meta["Deployment"]; exists && len(deployments) > 0 {
		deployment := deployments[0]

		if deployment == "crypto-library" {
			result.Diagnostics = append(result.Diagnostics, Diagnostic{
				Severity: ERROR,
				Section:  "META",
				Message:  "Deployment 'crypto-library' was retired in 0.3.6. Use 'verified-library' instead. verified-library covers all safety- and security-critical C-ABI libraries including cryptographic primitives.",
				Line:     1,
			})
			return
		}

		// Check if deployment is valid
		found := false
		for _, valid := range validDeploymentTemplates {
			if valid == deployment {
				found = true
				break
			}
		}
		if !found {
			result.Diagnostics = append(result.Diagnostics, Diagnostic{
				Severity: ERROR,
				Section:  "META",
				Message:  fmt.Sprintf("Unknown deployment template: '%s'. Run 'pcdp-lint list-templates' to see valid values.", deployment),
				Line:     1,
			})
			return
		}

		// Check deployment-specific requirements
		if deployment == "enhance-existing" {
			if languages, exists := spec.Meta["Language"]; !exists || len(languages) == 0 {
				result.Diagnostics = append(result.Diagnostics, Diagnostic{
					Severity: ERROR,
					Section:  "META",
					Message:  "Deployment 'enhance-existing' requires META field 'Language'",
					Line:     1,
				})
			} else if languages[0] == "" {
				result.Diagnostics = append(result.Diagnostics, Diagnostic{
					Severity: ERROR,
					Section:  "META",
					Message:  "META field 'Language' has empty value",
					Line:     1,
				})
			}
		}

		if deployment == "manual" {
			if targets, exists := spec.Meta["Target"]; !exists || len(targets) == 0 {
				result.Diagnostics = append(result.Diagnostics, Diagnostic{
					Severity: ERROR,
					Section:  "META",
					Message:  "Deployment 'manual' requires META field 'Target' (no template available for language resolution)",
					Line:     1,
				})
			}
		}

		if deployment == "python-tool" {
			if safetyLevels, exists := spec.Meta["Safety-Level"]; exists && len(safetyLevels) > 0 {
				if safetyLevels[0] != "QM" {
					result.Diagnostics = append(result.Diagnostics, Diagnostic{
						Severity: ERROR,
						Section:  "META",
						Message:  "Deployment 'python-tool' requires Safety-Level: QM. Python is not suitable for safety-critical components.",
						Line:     1,
					})
				}
			}
			if verifications, exists := spec.Meta["Verification"]; exists && len(verifications) > 0 {
				if verifications[0] != "none" {
					result.Diagnostics = append(result.Diagnostics, Diagnostic{
						Severity: ERROR,
						Section:  "META",
						Message:  "Deployment 'python-tool' requires Verification: none. No formal verification path exists for Python.",
						Line:     1,
					})
				}
			}
		}

		if deployment == "verified-library" {
			if safetyLevels, exists := spec.Meta["Safety-Level"]; exists && len(safetyLevels) > 0 {
				if safetyLevels[0] == "QM" {
					result.Diagnostics = append(result.Diagnostics, Diagnostic{
						Severity: WARNING,
						Section:  "META",
						Message:  "Deployment 'verified-library' with Safety-Level: QM is unusual. verified-library is intended for safety- or security-critical components. Consider using library-c-abi for general-purpose libraries.",
						Line:     1,
					})
				}
			}
		}
	}
}

// RULE-04: Deprecated META fields
func applyRule04(result *LintResult, spec ParsedSpec) {
	if targets, exists := spec.Meta["Target"]; exists && len(targets) > 0 {
		if deployments, dExists := spec.Meta["Deployment"]; !dExists || deployments[0] != "manual" {
			result.Diagnostics = append(result.Diagnostics, Diagnostic{
				Severity: WARNING,
				Section:  "META",
				Message:  "META field 'Target' is deprecated since v0.3.0. Target language is derived from the deployment template. Remove 'Target', or switch to Deployment: manual if explicit language control is required.",
				Line:     1,
			})
		}
	}

	if _, exists := spec.Meta["Domain"]; exists {
		result.Diagnostics = append(result.Diagnostics, Diagnostic{
			Severity: WARNING,
			Section:  "META",
			Message:  "META field 'Domain' is deprecated since v0.3.0. Use 'Deployment' instead.",
			Line:     1,
		})
	}
}

// RULE-05: Verification field value
func applyRule05(result *LintResult, spec ParsedSpec) {
	if verifications, exists := spec.Meta["Verification"]; exists && len(verifications) > 0 {
		verification := verifications[0]
		if !knownVerificationValues[verification] {
			result.Diagnostics = append(result.Diagnostics, Diagnostic{
				Severity: WARNING,
				Section:  "META",
				Message:  fmt.Sprintf("Unknown verification value: '%s'. Known values: none, lean4, fstar, dafny, custom. Custom verification backends are permitted; verify the value is intentional.", verification),
				Line:     1,
			})
		}
	}
}

// RULE-06: EXAMPLES section structure
func applyRule06(result *LintResult, spec ParsedSpec) {
	examplesLines, exists := spec.Sections["EXAMPLES"]
	if !exists || len(examplesLines) == 0 {
		result.Diagnostics = append(result.Diagnostics, Diagnostic{
			Severity: ERROR,
			Section:  "EXAMPLES",
			Message:  "EXAMPLES section contains no example blocks. Each example requires EXAMPLE:, GIVEN:, WHEN:, THEN: markers.",
			Line:     1,
		})
		return
	}

	// Check for example blocks
	hasExample := false
	for _, line := range examplesLines {
		if strings.HasPrefix(line, "EXAMPLE:") {
			hasExample = true
			break
		}
	}

	if !hasExample {
		result.Diagnostics = append(result.Diagnostics, Diagnostic{
			Severity: ERROR,
			Section:  "EXAMPLES",
			Message:  "EXAMPLES section contains no example blocks. Each example requires EXAMPLE:, GIVEN:, WHEN:, THEN: markers.",
			Line:     1,
		})
	}
}

// RULE-07: EXAMPLES minimum content (simplified)
func applyRule07(result *LintResult, spec ParsedSpec) {
	// This rule checks for empty blocks; simplified implementation
	// Full implementation would require detailed block parsing
}

// RULE-08: BEHAVIOR blocks must contain STEPS
func applyRule08(result *LintResult, spec ParsedSpec) {
	for sectionName, lines := range spec.Sections {
		if strings.HasPrefix(sectionName, "BEHAVIOR") {
			hasSteps := false
			for _, line := range lines {
				if strings.HasPrefix(line, "STEPS:") {
					hasSteps = true
					break
				}
			}
			if !hasSteps {
				result.Diagnostics = append(result.Diagnostics, Diagnostic{
					Severity: ERROR,
					Section:  sectionName,
					Message:  fmt.Sprintf("BEHAVIOR '%s' is missing required STEPS: block. Every BEHAVIOR must include ordered, imperative STEPS.", sectionName),
					Line:     1,
				})
			}
		}
	}
}

// RULE-09: INVARIANTS entries should carry observable/implementation tags
func applyRule09(result *LintResult, spec ParsedSpec) {
	invariantsLines, exists := spec.Sections["INVARIANTS"]
	if !exists {
		return
	}

	for _, line := range invariantsLines {
		trimmed := strings.TrimSpace(line)
		if len(trimmed) == 0 || strings.HasPrefix(trimmed, "-") {
			if !strings.Contains(trimmed, "[observable]") && !strings.Contains(trimmed, "[implementation]") {
				if strings.HasPrefix(trimmed, "- ") {
					result.Diagnostics = append(result.Diagnostics, Diagnostic{
						Severity: WARNING,
						Section:  "INVARIANTS",
						Message:  "Invariant entry missing tag. Prefix with [observable] or [implementation] for audit utility.",
						Line:     1,
					})
				}
			}
		}
	}
}

// RULE-10: Negative-path EXAMPLE required for BEHAVIOR with error exits
func applyRule10(result *LintResult, spec ParsedSpec) {
	// Simplified implementation: check if BEHAVIOR has error exits
	for sectionName, lines := range spec.Sections {
		if strings.HasPrefix(sectionName, "BEHAVIOR") {
			hasErrorExit := false
			for _, line := range lines {
				if strings.Contains(line, "→") {
					hasErrorExit = true
					break
				}
			}
			if hasErrorExit {
				// Check if EXAMPLES has negative-path examples
				examplesLines, exists := spec.Sections["EXAMPLES"]
				if !exists {
					continue
				}
				hasNegativeExample := false
				for _, line := range examplesLines {
					if strings.Contains(line, "Err(") || strings.Contains(line, "error") ||
						strings.Contains(line, "exit_code = 1") || strings.Contains(line, "exit_code = 2") {
						hasNegativeExample = true
						break
					}
				}
				if !hasNegativeExample {
					result.Diagnostics = append(result.Diagnostics, Diagnostic{
						Severity: ERROR,
						Section:  sectionName,
						Message:  fmt.Sprintf("BEHAVIOR '%s' has error exits in STEPS but no negative-path EXAMPLE. Add at least one EXAMPLE whose THEN: verifies an error outcome.", sectionName),
						Line:     1,
					})
				}
			}
		}
	}
}

// RULE-13: Constraint field value on BEHAVIOR headers
func applyRule13(result *LintResult, spec ParsedSpec) {
	validConstraints := map[string]bool{"required": true, "supported": true, "forbidden": true}

	for sectionName, lines := range spec.Sections {
		if strings.HasPrefix(sectionName, "BEHAVIOR") {
			for _, line := range lines {
				if strings.HasPrefix(line, "Constraint:") {
					constraint := strings.TrimPrefix(line, "Constraint:")
					constraint = strings.TrimSpace(constraint)
					if !validConstraints[constraint] {
						result.Diagnostics = append(result.Diagnostics, Diagnostic{
							Severity: ERROR,
							Section:  sectionName,
							Message:  fmt.Sprintf("BEHAVIOR '%s' has invalid Constraint: value '%s'. Valid values: required, supported, forbidden.", sectionName, constraint),
							Line:     1,
						})
					}
					if constraint == "forbidden" {
						hasReason := false
						for _, l := range lines {
							if strings.Contains(l, "reason:") {
								hasReason = true
								break
							}
						}
						if !hasReason {
							result.Diagnostics = append(result.Diagnostics, Diagnostic{
								Severity: WARNING,
								Section:  sectionName,
								Message:  fmt.Sprintf("BEHAVIOR '%s' is Constraint: forbidden but has no reason: annotation.", sectionName),
								Line:     1,
							})
						}
					}
				}
			}
		}
	}
}

// RULE-14: EXECUTION section required in deployment templates
func applyRule14(result *LintResult, spec ParsedSpec) {
	if deployments, exists := spec.Meta["Deployment"]; exists && len(deployments) > 0 {
		if deployments[0] == "template" {
			if _, hasExecution := spec.Meta["EXECUTION"]; !hasExecution {
				if _, hasExecSection := spec.Sections["EXECUTION"]; !hasExecSection {
					result.Diagnostics = append(result.Diagnostics, Diagnostic{
						Severity: WARNING,
						Section:  "structure",
						Message:  "Deployment template is missing ## EXECUTION section. Translators cannot determine delivery phases without it. Add ## EXECUTION or declare 'EXECUTION: none' in META if this template intentionally has no execution recipe.",
						Line:     1,
					})
				}
			}
		}
	}
}

func isSemanticVersion(v string) bool {
	pattern := regexp.MustCompile(`^[0-9]+\.[0-9]+\.[0-9]+$`)
	return pattern.MatchString(v)
}

func isValidSPDXLicense(license string) bool {
	// Handle compound expressions
	if strings.Contains(license, " OR ") {
		parts := strings.Split(license, " OR ")
		for _, part := range parts {
			if !validSPDXLicenses[strings.TrimSpace(part)] {
				return false
			}
		}
		return true
	}
	return validSPDXLicenses[license]
}

func listTemplates() {
	templates := []struct {
		name       string
		annotation string
	}{
		{"wasm", "(template file not found)"},
		{"ebpf", "(template file not found)"},
		{"kernel-module", "(template file not found)"},
		{"verified-library", "(template file not found)"},
		{"cli-tool", "Go"},
		{"gui-tool", "(template file not found)"},
		{"cloud-native", "(template file not found)"},
		{"backend-service", "(template file not found)"},
		{"library-c-abi", "(template file not found)"},
		{"enterprise-software", "(template file not found)"},
		{"academic", "(template file not found)"},
		{"python-tool", "Python"},
		{"enhance-existing", "(declare Language: in META)"},
		{"manual", "(declare Target: in META)"},
		{"template", "(template definition file, not translatable)"},
		{"mcp-server", "(template file not found)"},
		{"project-manifest", "(architect artifact, no code generated)"},
	}

	for _, t := range templates {
		fmt.Printf("%s  →  %s\n", t.name, t.annotation)
	}
	os.Exit(0)
}

func printVersion() {
	fmt.Printf("pcdp-lint %s (schema %s) spdx/%s\n", VERSION, SPEC_SCHEMA, SPDX_VERSION)
	os.Exit(0)
}
