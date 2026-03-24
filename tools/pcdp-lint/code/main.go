package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
)

// Types
type Severity int

const (
	Error Severity = iota
	Warning
)

func (s Severity) String() string {
	switch s {
	case Error:
		return "ERROR"
	case Warning:
		return "WARNING"
	default:
		return "UNKNOWN"
	}
}

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

type MetaField struct {
	Key   string
	Value string
}

// Known deployment templates
var deploymentTemplates = []string{
	"wasm", "ebpf", "kernel-module", "verified-library",
	"cli-tool", "gui-tool", "cloud-native", "backend-service",
	"library-c-abi", "enterprise-software", "academic",
	"python-tool", "enhance-existing", "manual", "template",
	"mcp-server", "project-manifest",
}

// Known verification values
var knownVerificationValues = []string{
	"none", "lean4", "fstar", "dafny", "custom",
}

// SPDX license identifiers (simplified list for implementation)
var spdxLicenses = map[string]bool{
	"Apache-2.0": true, "MIT": true, "GPL-2.0-only": true, "GPL-2.0-or-later": true,
	"GPL-3.0-only": true, "GPL-3.0-or-later": true, "LGPL-2.1-only": true,
	"LGPL-2.1-or-later": true, "LGPL-3.0-only": true, "LGPL-3.0-or-later": true,
	"BSD-2-Clause": true, "BSD-3-Clause": true, "ISC": true, "CC-BY-4.0": true,
	"CC0-1.0": true, "Unlicense": true, "0BSD": true, "MPL-2.0": true,
}

func main() {
	args := os.Args[1:]
	
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "Usage: pcdp-lint [strict=true] <specfile.md>\n")
		fmt.Fprintf(os.Stderr, "       pcdp-lint list-templates\n")
		os.Exit(2)
	}

	// Handle list-templates command
	if len(args) == 1 && args[0] == "list-templates" {
		listTemplates()
		return
	}

	// Parse key=value arguments
	strict := false
	var filename string
	
	for _, arg := range args {
		if strings.Contains(arg, "=") {
			parts := strings.SplitN(arg, "=", 2)
			key := parts[0]
			value := parts[1]
			
			switch key {
			case "strict":
				if value == "true" {
					strict = true
				} else if value == "false" {
					strict = false
				} else {
					fmt.Fprintf(os.Stderr, "error: invalid value for strict: %s\n", value)
					os.Exit(2)
				}
			default:
				fmt.Fprintf(os.Stderr, "error: unrecognised option: %s\n", key)
				os.Exit(2)
			}
		} else {
			if filename != "" {
				fmt.Fprintf(os.Stderr, "error: multiple files specified\n")
				os.Exit(2)
			}
			filename = arg
		}
	}

	if filename == "" {
		fmt.Fprintf(os.Stderr, "Usage: pcdp-lint [strict=true] <specfile.md>\n")
		os.Exit(2)
	}

	result := lint(filename, strict)
	
	// Output diagnostics to stderr
	for _, diag := range result.Diagnostics {
		fmt.Fprintf(os.Stderr, "%s  %s:%d  [%s]  %s\n", 
			diag.Severity, result.File, diag.Line, diag.Section, diag.Message)
	}
	
	// Output summary to stdout
	errorCount := 0
	warningCount := 0
	for _, diag := range result.Diagnostics {
		if diag.Severity == Error {
			errorCount++
		} else {
			warningCount++
		}
	}
	
	if errorCount == 0 && warningCount == 0 {
		fmt.Printf("✓ %s: valid\n", result.File)
	} else if errorCount == 0 && !strict {
		fmt.Printf("✓ %s: valid (%d warning(s))\n", result.File, warningCount)
	} else if strict && errorCount == 0 && warningCount > 0 {
		fmt.Printf("✗ %s: %d error(s), %d warning(s) [strict mode]\n", 
			result.File, errorCount, warningCount)
	} else {
		if strict && warningCount > 0 {
			fmt.Printf("✗ %s: %d error(s), %d warning(s) [strict mode]\n", 
				result.File, errorCount, warningCount)
		} else {
			fmt.Printf("✗ %s: %d error(s), %d warning(s)\n", 
				result.File, errorCount, warningCount)
		}
	}
	
	os.Exit(result.ExitCode)
}

func listTemplates() {
	templateAnnotations := map[string]string{
		"wasm":                "Go",
		"ebpf":               "C",
		"kernel-module":      "C",
		"verified-library":   "C",
		"cli-tool":           "Go",
		"gui-tool":           "Go",
		"cloud-native":       "Go",
		"backend-service":    "Go",
		"library-c-abi":      "C",
		"enterprise-software": "Go",
		"academic":           "Go",
		"python-tool":        "Python",
		"enhance-existing":   "(declare Language: in META)",
		"manual":             "(declare Target: in META)",
		"template":           "(template definition file, not translatable)",
		"mcp-server":         "Go",
		"project-manifest":   "(architect artifact, no code generated)",
	}
	
	for _, template := range deploymentTemplates {
		annotation := templateAnnotations[template]
		if annotation == "" {
			annotation = "(template file not found)"
		}
		fmt.Printf("%s  →  %s\n", template, annotation)
	}
	
	os.Exit(0)
}

func lint(filename string, strict bool) LintResult {
	result := LintResult{
		File:        filename,
		Diagnostics: []Diagnostic{},
		ExitCode:    0,
	}

	// Check file extension
	if !strings.HasSuffix(filename, ".md") {
		fmt.Fprintf(os.Stderr, "error: file must have .md extension: %s\n", filename)
		os.Exit(2)
	}

	// Try to open and read file
	file, err := os.Open(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: cannot open file: %s\n", filename)
		os.Exit(2)
	}
	defer file.Close()

	lines := []string{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "error: cannot read file: %s\n", filename)
		os.Exit(2)
	}

	// Apply all rules
	result.Diagnostics = append(result.Diagnostics, rule01RequiredSections(lines)...)
	result.Diagnostics = append(result.Diagnostics, rule02MetaFields(lines)...)
	result.Diagnostics = append(result.Diagnostics, rule03DeploymentTemplate(lines)...)
	result.Diagnostics = append(result.Diagnostics, rule04DeprecatedFields(lines)...)
	result.Diagnostics = append(result.Diagnostics, rule05VerificationField(lines)...)
	result.Diagnostics = append(result.Diagnostics, rule06ExamplesStructure(lines)...)
	result.Diagnostics = append(result.Diagnostics, rule07ExamplesContent(lines)...)
	result.Diagnostics = append(result.Diagnostics, rule08BehaviorSteps(lines)...)
	result.Diagnostics = append(result.Diagnostics, rule09InvariantTags(lines)...)
	result.Diagnostics = append(result.Diagnostics, rule10NegativePathExamples(lines)...)
	result.Diagnostics = append(result.Diagnostics, rule11ToolchainConstraints(lines)...)
	result.Diagnostics = append(result.Diagnostics, rule12CrossSectionConsistency(lines)...)
	result.Diagnostics = append(result.Diagnostics, rule13BehaviorConstraints(lines)...)

	// Sort diagnostics by line number
	sort.Slice(result.Diagnostics, func(i, j int) bool {
		return result.Diagnostics[i].Line < result.Diagnostics[j].Line
	})

	// Compute exit code
	hasError := false
	hasWarning := false
	for _, diag := range result.Diagnostics {
		if diag.Severity == Error {
			hasError = true
		} else {
			hasWarning = true
		}
	}

	if hasError || (strict && hasWarning) {
		result.ExitCode = 1
	} else {
		result.ExitCode = 0
	}

	return result
}

func rule01RequiredSections(lines []string) []Diagnostic {
	diagnostics := []Diagnostic{}
	requiredSections := []string{
		"## META", "## TYPES", "## BEHAVIOR", "## PRECONDITIONS",
		"## POSTCONDITIONS", "## INVARIANTS", "## EXAMPLES",
	}

	foundSections := make(map[string]bool)
	
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		for _, section := range requiredSections {
			if trimmed == section {
				foundSections[section] = true
			}
		}
		// Also check for BEHAVIOR variants
		if strings.HasPrefix(trimmed, "## BEHAVIOR:") || strings.HasPrefix(trimmed, "## BEHAVIOR/INTERNAL:") {
			foundSections["## BEHAVIOR"] = true
		}
	}

	for _, section := range requiredSections {
		if !foundSections[section] {
			diagnostics = append(diagnostics, Diagnostic{
				Severity: Error,
				Section:  "structure",
				Message:  fmt.Sprintf("Missing required section: %s", section),
				Line:     1,
			})
		}
	}

	return diagnostics
}

func rule02MetaFields(lines []string) []Diagnostic {
	diagnostics := []Diagnostic{}
	requiredFields := []string{
		"Deployment", "Verification", "Safety-Level",
		"Version", "Spec-Schema", "License",
	}

	metaFields := extractMetaFields(lines)
	foundFields := make(map[string]bool)
	authorFound := false

	for _, field := range metaFields {
		foundFields[field.Key] = true
		if field.Key == "Author" {
			authorFound = true
		}
		
		// Check for empty values
		if strings.TrimSpace(field.Value) == "" {
			diagnostics = append(diagnostics, Diagnostic{
				Severity: Error,
				Section:  "META",
				Message:  fmt.Sprintf("META field %s has empty value", field.Key),
				Line:     getMetaFieldLine(lines, field.Key),
			})
		}
	}

	// Check required fields
	for _, field := range requiredFields {
		if !foundFields[field] {
			diagnostics = append(diagnostics, Diagnostic{
				Severity: Error,
				Section:  "META",
				Message:  fmt.Sprintf("Missing required META field: %s", field),
				Line:     getMetaSectionLine(lines),
			})
		}
	}

	// Check Author field specifically
	if !authorFound {
		diagnostics = append(diagnostics, Diagnostic{
			Severity: Error,
			Section:  "META",
			Message:  "Missing required META field: Author (at least one Author: line required)",
			Line:     getMetaSectionLine(lines),
		})
	}

	// Version format validation
	if version := getMetaFieldValue(metaFields, "Version"); version != "" {
		if !isValidSemanticVersion(version) {
			diagnostics = append(diagnostics, Diagnostic{
				Severity: Error,
				Section:  "META",
				Message:  fmt.Sprintf("Version '%s' is not valid semantic versioning. Required format: MAJOR.MINOR.PATCH (e.g. 0.1.0)", version),
				Line:     getMetaFieldLine(lines, "Version"),
			})
		}
	}

	// Spec-Schema format validation
	if specSchema := getMetaFieldValue(metaFields, "Spec-Schema"); specSchema != "" {
		if !isValidSemanticVersion(specSchema) {
			diagnostics = append(diagnostics, Diagnostic{
				Severity: Error,
				Section:  "META",
				Message:  fmt.Sprintf("Spec-Schema '%s' is not valid semantic versioning. Required format: MAJOR.MINOR.PATCH (e.g. 0.1.0)", specSchema),
				Line:     getMetaFieldLine(lines, "Spec-Schema"),
			})
		}
	}

	// License SPDX validation
	if license := getMetaFieldValue(metaFields, "License"); license != "" {
		if !isValidSPDXLicense(license) {
			diagnostics = append(diagnostics, Diagnostic{
				Severity: Error,
				Section:  "META",
				Message:  fmt.Sprintf("License '%s' is not a valid SPDX identifier. See https://spdx.org/licenses/ for valid identifiers. Compound expressions permitted (e.g. Apache-2.0 OR MIT).", license),
				Line:     getMetaFieldLine(lines, "License"),
			})
		}
	}

	return diagnostics
}

func rule03DeploymentTemplate(lines []string) []Diagnostic {
	diagnostics := []Diagnostic{}
	metaFields := extractMetaFields(lines)
	
	deployment := getMetaFieldValue(metaFields, "Deployment")
	if deployment == "" {
		return diagnostics // Already handled by rule02
	}

	// Check for retired crypto-library
	if deployment == "crypto-library" {
		diagnostics = append(diagnostics, Diagnostic{
			Severity: Error,
			Section:  "META",
			Message:  "Deployment 'crypto-library' was retired in 0.3.6. Use 'verified-library' instead. verified-library covers all safety- and security-critical C-ABI libraries including cryptographic primitives.",
			Line:     getMetaFieldLine(lines, "Deployment"),
		})
		return diagnostics
	}

	// Check if deployment template is known
	validTemplate := false
	for _, template := range deploymentTemplates {
		if deployment == template {
			validTemplate = true
			break
		}
	}
	
	if !validTemplate {
		diagnostics = append(diagnostics, Diagnostic{
			Severity: Error,
			Section:  "META",
			Message:  fmt.Sprintf("Unknown deployment template: '%s'. Run 'pcdp-lint list-templates' to see valid values.", deployment),
			Line:     getMetaFieldLine(lines, "Deployment"),
		})
		return diagnostics
	}

	// Special validation for enhance-existing
	if deployment == "enhance-existing" {
		language := getMetaFieldValue(metaFields, "Language")
		if language == "" {
			diagnostics = append(diagnostics, Diagnostic{
				Severity: Error,
				Section:  "META",
				Message:  "Deployment 'enhance-existing' requires META field 'Language'",
				Line:     getMetaFieldLine(lines, "Deployment"),
			})
		}
	}

	// Special validation for manual
	if deployment == "manual" {
		target := getMetaFieldValue(metaFields, "Target")
		if target == "" {
			diagnostics = append(diagnostics, Diagnostic{
				Severity: Error,
				Section:  "META",
				Message:  "Deployment 'manual' requires META field 'Target' (no template available for language resolution)",
				Line:     getMetaFieldLine(lines, "Deployment"),
			})
		}
	}

	// Special validation for python-tool
	if deployment == "python-tool" {
		safetyLevel := getMetaFieldValue(metaFields, "Safety-Level")
		if safetyLevel != "QM" {
			diagnostics = append(diagnostics, Diagnostic{
				Severity: Error,
				Section:  "META",
				Message:  "Deployment 'python-tool' requires Safety-Level: QM. Python is not suitable for safety-critical components.",
				Line:     getMetaFieldLine(lines, "Safety-Level"),
			})
		}
		
		verification := getMetaFieldValue(metaFields, "Verification")
		if verification != "none" {
			diagnostics = append(diagnostics, Diagnostic{
				Severity: Error,
				Section:  "META",
				Message:  "Deployment 'python-tool' requires Verification: none. No formal verification path exists for Python.",
				Line:     getMetaFieldLine(lines, "Verification"),
			})
		}
	}

	// Special validation for verified-library
	if deployment == "verified-library" {
		safetyLevel := getMetaFieldValue(metaFields, "Safety-Level")
		if safetyLevel == "QM" {
			diagnostics = append(diagnostics, Diagnostic{
				Severity: Warning,
				Section:  "META",
				Message:  "Deployment 'verified-library' with Safety-Level: QM is unusual. verified-library is intended for safety- or security-critical components. Consider using library-c-abi for general-purpose libraries.",
				Line:     getMetaFieldLine(lines, "Safety-Level"),
			})
		}
	}

	return diagnostics
}

func rule04DeprecatedFields(lines []string) []Diagnostic {
	diagnostics := []Diagnostic{}
	metaFields := extractMetaFields(lines)
	deployment := getMetaFieldValue(metaFields, "Deployment")

	// Check for deprecated Target field
	if target := getMetaFieldValue(metaFields, "Target"); target != "" && deployment != "manual" {
		diagnostics = append(diagnostics, Diagnostic{
			Severity: Warning,
			Section:  "META",
			Message:  "META field 'Target' is deprecated since v0.3.0. Target language is derived from the deployment template. Remove 'Target', or switch to Deployment: manual if explicit language control is required.",
			Line:     getMetaFieldLine(lines, "Target"),
		})
	}

	// Check for deprecated Domain field
	if domain := getMetaFieldValue(metaFields, "Domain"); domain != "" {
		diagnostics = append(diagnostics, Diagnostic{
			Severity: Warning,
			Section:  "META",
			Message:  "META field 'Domain' is deprecated since v0.3.0. Use 'Deployment' instead.",
			Line:     getMetaFieldLine(lines, "Domain"),
		})
	}

	return diagnostics
}

func rule05VerificationField(lines []string) []Diagnostic {
	diagnostics := []Diagnostic{}
	metaFields := extractMetaFields(lines)
	
	verification := getMetaFieldValue(metaFields, "Verification")
	if verification == "" {
		return diagnostics // Already handled by rule02
	}

	validVerification := false
	for _, valid := range knownVerificationValues {
		if verification == valid {
			validVerification = true
			break
		}
	}

	if !validVerification {
		diagnostics = append(diagnostics, Diagnostic{
			Severity: Warning,
			Section:  "META",
			Message:  fmt.Sprintf("Unknown verification value: '%s'. Known values: none, lean4, fstar, dafny, custom. Custom verification backends are permitted; verify the value is intentional.", verification),
			Line:     getMetaFieldLine(lines, "Verification"),
		})
	}

	return diagnostics
}

func rule06ExamplesStructure(lines []string) []Diagnostic {
	diagnostics := []Diagnostic{}
	
	exampleBlocks := extractExampleBlocks(lines)
	if len(exampleBlocks) == 0 {
		diagnostics = append(diagnostics, Diagnostic{
			Severity: Error,
			Section:  "EXAMPLES",
			Message:  "EXAMPLES section contains no example blocks. Each example requires EXAMPLE:, GIVEN:, WHEN:, THEN: markers.",
			Line:     getSectionLine(lines, "## EXAMPLES"),
		})
		return diagnostics
	}

	for _, block := range exampleBlocks {
		if !block.HasGiven {
			diagnostics = append(diagnostics, Diagnostic{
				Severity: Error,
				Section:  "EXAMPLES",
				Message:  fmt.Sprintf("Example '%s' missing GIVEN: marker", block.Name),
				Line:     block.StartLine,
			})
		}
		
		if len(block.WhenThenPairs) == 0 {
			diagnostics = append(diagnostics, Diagnostic{
				Severity: Error,
				Section:  "EXAMPLES",
				Message:  fmt.Sprintf("Example '%s' missing WHEN: marker", block.Name),
				Line:     block.StartLine,
			})
			diagnostics = append(diagnostics, Diagnostic{
				Severity: Error,
				Section:  "EXAMPLES",
				Message:  fmt.Sprintf("Example '%s' missing THEN: marker", block.Name),
				Line:     block.StartLine,
			})
		} else {
			// Check WHEN/THEN pairing
			for _, pair := range block.WhenThenPairs {
				if !pair.HasThen {
					diagnostics = append(diagnostics, Diagnostic{
						Severity: Error,
						Section:  "EXAMPLES",
						Message:  fmt.Sprintf("Example '%s' has WHEN: without a matching THEN:", block.Name),
						Line:     pair.WhenLine,
					})
				}
			}
		}
	}

	return diagnostics
}

func rule07ExamplesContent(lines []string) []Diagnostic {
	diagnostics := []Diagnostic{}
	
	exampleBlocks := extractExampleBlocks(lines)
	
	for _, block := range exampleBlocks {
		if block.HasGiven && len(strings.TrimSpace(block.GivenContent)) == 0 {
			diagnostics = append(diagnostics, Diagnostic{
				Severity: Warning,
				Section:  "EXAMPLES",
				Message:  fmt.Sprintf("Example '%s' has empty GIVEN block", block.Name),
				Line:     block.GivenLine,
			})
		}
		
		for _, pair := range block.WhenThenPairs {
			if len(strings.TrimSpace(pair.WhenContent)) == 0 {
				diagnostics = append(diagnostics, Diagnostic{
					Severity: Warning,
					Section:  "EXAMPLES",
					Message:  fmt.Sprintf("Example '%s' has empty WHEN block", block.Name),
					Line:     pair.WhenLine,
				})
			}
			if pair.HasThen && len(strings.TrimSpace(pair.ThenContent)) == 0 {
				diagnostics = append(diagnostics, Diagnostic{
					Severity: Warning,
					Section:  "EXAMPLES",
					Message:  fmt.Sprintf("Example '%s' has empty THEN block", block.Name),
					Line:     pair.ThenLine,
				})
			}
		}
	}

	return diagnostics
}

func rule08BehaviorSteps(lines []string) []Diagnostic {
	diagnostics := []Diagnostic{}
	
	behaviorSections := extractBehaviorSections(lines)
	
	for _, behavior := range behaviorSections {
		hasSteps := false
		for i := behavior.StartLine; i <= behavior.EndLine && i < len(lines); i++ {
			if strings.HasPrefix(strings.TrimSpace(lines[i]), "STEPS:") {
				hasSteps = true
				break
			}
		}
		
		if !hasSteps {
			diagnostics = append(diagnostics, Diagnostic{
				Severity: Error,
				Section:  behavior.Name,
				Message:  fmt.Sprintf("BEHAVIOR '%s' is missing required STEPS: block. Every BEHAVIOR must include ordered, imperative STEPS.", behavior.Name),
				Line:     behavior.StartLine + 1,
			})
		}
	}

	return diagnostics
}

func rule09InvariantTags(lines []string) []Diagnostic {
	diagnostics := []Diagnostic{}
	
	invariantsStart := getSectionLine(lines, "## INVARIANTS")
	if invariantsStart == -1 {
		return diagnostics
	}
	
	invariantsEnd := getNextSectionLine(lines, invariantsStart)
	if invariantsEnd == -1 {
		invariantsEnd = len(lines)
	}
	
	for i := invariantsStart + 1; i < invariantsEnd; i++ {
		line := strings.TrimSpace(lines[i])
		if line == "" || strings.HasPrefix(line, "##") || strings.HasPrefix(line, "---") {
			continue
		}
		
		// Check if it's an entry line (starts with -)
		if strings.HasPrefix(line, "-") {
			if !strings.HasPrefix(line, "- [observable]") && !strings.HasPrefix(line, "- [implementation]") {
				diagnostics = append(diagnostics, Diagnostic{
					Severity: Warning,
					Section:  "INVARIANTS",
					Message:  "Invariant entry missing tag. Prefix with [observable] or [implementation] for audit utility.",
					Line:     i + 1,
				})
			}
		}
	}

	return diagnostics
}

func rule10NegativePathExamples(lines []string) []Diagnostic {
	diagnostics := []Diagnostic{}
	
	behaviorSections := extractBehaviorSections(lines)
	exampleBlocks := extractExampleBlocks(lines)
	
	for _, behavior := range behaviorSections {
		hasErrorExits := false
		
		// Check for error exits in STEPS
		for i := behavior.StartLine; i <= behavior.EndLine && i < len(lines); i++ {
			line := lines[i]
			if strings.Contains(line, "→") || strings.Contains(line, "error") || 
			   strings.Contains(line, "exit") || strings.Contains(line, "failure") {
				hasErrorExits = true
				break
			}
		}
		
		if hasErrorExits {
			// Look for negative-path examples
			hasNegativeExample := false
			
			for _, block := range exampleBlocks {
				for _, pair := range block.WhenThenPairs {
					thenContent := strings.ToLower(pair.ThenContent)
					if strings.Contains(thenContent, "err(") ||
					   strings.Contains(thenContent, "error") ||
					   strings.Contains(thenContent, "exit_code = 1") ||
					   strings.Contains(thenContent, "exit_code = 2") ||
					   strings.Contains(thenContent, "stderr contains") {
						hasNegativeExample = true
						break
					}
				}
				if hasNegativeExample {
					break
				}
			}
			
			if !hasNegativeExample {
				diagnostics = append(diagnostics, Diagnostic{
					Severity: Error,
					Section:  behavior.Name,
					Message:  fmt.Sprintf("BEHAVIOR '%s' has error exits in STEPS but no negative-path EXAMPLE. Add at least one EXAMPLE whose THEN: verifies an error outcome.", behavior.Name),
					Line:     behavior.StartLine + 1,
				})
			}
		}
	}

	return diagnostics
}

func rule11ToolchainConstraints(lines []string) []Diagnostic {
	diagnostics := []Diagnostic{}
	
	toolchainStart := getSectionLine(lines, "## TOOLCHAIN-CONSTRAINTS")
	if toolchainStart == -1 {
		return diagnostics // Section is optional
	}
	
	toolchainEnd := getNextSectionLine(lines, toolchainStart)
	if toolchainEnd == -1 {
		toolchainEnd = len(lines)
	}
	
	for i := toolchainStart + 1; i < toolchainEnd; i++ {
		line := strings.TrimSpace(lines[i])
		if line == "" || strings.HasPrefix(line, "##") {
			continue
		}
		
		// Look for constraint values other than required/forbidden
		if strings.Contains(line, ":") && !strings.Contains(line, "required") && 
		   !strings.Contains(line, "forbidden") && strings.Contains(line, "constraint") {
			diagnostics = append(diagnostics, Diagnostic{
				Severity: Warning,
				Section:  "TOOLCHAIN-CONSTRAINTS",
				Message:  "TOOLCHAIN-CONSTRAINTS entry uses unknown constraint value. Valid values: required, forbidden.",
				Line:     i + 1,
			})
		}
	}

	return diagnostics
}

func rule12CrossSectionConsistency(lines []string) []Diagnostic {
	diagnostics := []Diagnostic{}
	
	// This is a simplified implementation of rule 12
	// Full implementation would require more sophisticated parsing
	
	return diagnostics
}

func rule13BehaviorConstraints(lines []string) []Diagnostic {
	diagnostics := []Diagnostic{}
	validConstraints := []string{"required", "supported", "forbidden"}
	
	behaviorSections := extractBehaviorSections(lines)
	
	for _, behavior := range behaviorSections {
		for i := behavior.StartLine; i <= behavior.EndLine && i < len(lines); i++ {
			line := strings.TrimSpace(lines[i])
			if strings.HasPrefix(line, "Constraint:") {
				parts := strings.SplitN(line, ":", 2)
				if len(parts) == 2 {
					constraint := strings.TrimSpace(parts[1])
					
					validConstraint := false
					for _, valid := range validConstraints {
						if constraint == valid {
							validConstraint = true
							break
						}
					}
					
					if !validConstraint {
						diagnostics = append(diagnostics, Diagnostic{
							Severity: Error,
							Section:  behavior.Name,
							Message:  fmt.Sprintf("BEHAVIOR '%s' has invalid Constraint: value '%s'. Valid values: required, supported, forbidden.", behavior.Name, constraint),
							Line:     i + 1,
						})
					}
					
					if constraint == "forbidden" {
						// Check for reason annotation
						hasReason := false
						for j := i + 1; j <= behavior.EndLine && j < len(lines); j++ {
							if strings.HasPrefix(strings.TrimSpace(lines[j]), "reason:") {
								hasReason = true
								break
							}
						}
						
						if !hasReason {
							diagnostics = append(diagnostics, Diagnostic{
								Severity: Warning,
								Section:  behavior.Name,
								Message:  fmt.Sprintf("BEHAVIOR '%s' is Constraint: forbidden but has no reason: annotation.", behavior.Name),
								Line:     i + 1,
							})
						}
					}
				}
			}
		}
	}

	return diagnostics
}

// Helper functions

func extractMetaFields(lines []string) []MetaField {
	fields := []MetaField{}
	inMeta := false
	
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "## META" {
			inMeta = true
			continue
		}
		if strings.HasPrefix(trimmed, "##") && trimmed != "## META" {
			inMeta = false
		}
		
		if inMeta && strings.Contains(line, ":") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])
				fields = append(fields, MetaField{Key: key, Value: value})
			}
		}
	}
	
	return fields
}

func getMetaFieldValue(fields []MetaField, key string) string {
	for _, field := range fields {
		if field.Key == key {
			return field.Value
		}
	}
	return ""
}

func getMetaFieldLine(lines []string, key string) int {
	inMeta := false
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "## META" {
			inMeta = true
			continue
		}
		if strings.HasPrefix(trimmed, "##") && trimmed != "## META" {
			inMeta = false
		}
		
		if inMeta && strings.HasPrefix(strings.TrimSpace(line), key+":") {
			return i + 1
		}
	}
	return 1
}

func getMetaSectionLine(lines []string) int {
	for i, line := range lines {
		if strings.TrimSpace(line) == "## META" {
			return i + 1
		}
	}
	return 1
}

func getSectionLine(lines []string, section string) int {
	for i, line := range lines {
		if strings.TrimSpace(line) == section {
			return i
		}
	}
	return -1
}

func getNextSectionLine(lines []string, start int) int {
	for i := start + 1; i < len(lines); i++ {
		if strings.HasPrefix(strings.TrimSpace(lines[i]), "##") {
			return i
		}
	}
	return -1
}

func isValidSemanticVersion(version string) bool {
	pattern := `^[0-9]+\.[0-9]+\.[0-9]+$`
	matched, _ := regexp.MatchString(pattern, version)
	return matched
}

func isValidSPDXLicense(license string) bool {
	// Simplified validation - check if it's in our known list or contains OR/AND
	if spdxLicenses[license] {
		return true
	}
	
	// Handle compound expressions
	if strings.Contains(license, " OR ") || strings.Contains(license, " AND ") {
		// For simplicity, assume compound expressions are valid
		return true
	}
	
	return false
}

// Types for parsing structures

type ExampleBlock struct {
	Name          string
	StartLine     int
	HasGiven      bool
	GivenLine     int
	GivenContent  string
	WhenThenPairs []WhenThenPair
}

type WhenThenPair struct {
	WhenLine    int
	WhenContent string
	HasThen     bool
	ThenLine    int
	ThenContent string
}

type BehaviorSection struct {
	Name      string
	StartLine int
	EndLine   int
}

func extractExampleBlocks(lines []string) []ExampleBlock {
	blocks := []ExampleBlock{}
	
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "EXAMPLE:") {
			name := strings.TrimSpace(strings.TrimPrefix(trimmed, "EXAMPLE:"))
			block := ExampleBlock{
				Name:          name,
				StartLine:     i + 1,
				WhenThenPairs: []WhenThenPair{},
			}
			
			// Parse the block content
			j := i + 1
			var currentPair *WhenThenPair
			
			for j < len(lines) {
				blockLine := strings.TrimSpace(lines[j])
				
				// Stop at next example or section
				if strings.HasPrefix(blockLine, "EXAMPLE:") || strings.HasPrefix(blockLine, "##") {
					break
				}
				
				if strings.HasPrefix(blockLine, "GIVEN:") {
					block.HasGiven = true
					block.GivenLine = j + 1
					// Collect GIVEN content until WHEN
					j++
					givenContent := ""
					for j < len(lines) {
						nextLine := strings.TrimSpace(lines[j])
						if strings.HasPrefix(nextLine, "WHEN:") || strings.HasPrefix(nextLine, "EXAMPLE:") || strings.HasPrefix(nextLine, "##") {
							break
						}
						givenContent += lines[j] + "\n"
						j++
					}
					block.GivenContent = givenContent
					continue
				}
				
				if strings.HasPrefix(blockLine, "WHEN:") {
					if currentPair != nil {
						block.WhenThenPairs = append(block.WhenThenPairs, *currentPair)
					}
					currentPair = &WhenThenPair{
						WhenLine: j + 1,
					}
					// Collect WHEN content until THEN
					j++
					whenContent := ""
					for j < len(lines) {
						nextLine := strings.TrimSpace(lines[j])
						if strings.HasPrefix(nextLine, "THEN:") || strings.HasPrefix(nextLine, "WHEN:") || strings.HasPrefix(nextLine, "EXAMPLE:") || strings.HasPrefix(nextLine, "##") {
							break
						}
						whenContent += lines[j] + "\n"
						j++
					}
					currentPair.WhenContent = whenContent
					continue
				}
				
				if strings.HasPrefix(blockLine, "THEN:") && currentPair != nil {
					currentPair.HasThen = true
					currentPair.ThenLine = j + 1
					// Collect THEN content until next WHEN or end
					j++
					thenContent := ""
					for j < len(lines) {
						nextLine := strings.TrimSpace(lines[j])
						if strings.HasPrefix(nextLine, "WHEN:") || strings.HasPrefix(nextLine, "EXAMPLE:") || strings.HasPrefix(nextLine, "##") {
							break
						}
						thenContent += lines[j] + "\n"
						j++
					}
					currentPair.ThenContent = thenContent
					continue
				}
				
				j++
			}
			
			if currentPair != nil {
				block.WhenThenPairs = append(block.WhenThenPairs, *currentPair)
			}
			
			blocks = append(blocks, block)
		}
	}
	
	return blocks
}

func extractBehaviorSections(lines []string) []BehaviorSection {
	sections := []BehaviorSection{}
	
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "## BEHAVIOR:") || strings.HasPrefix(trimmed, "## BEHAVIOR/INTERNAL:") {
			name := ""
			if strings.HasPrefix(trimmed, "## BEHAVIOR:") {
				name = strings.TrimSpace(strings.TrimPrefix(trimmed, "## BEHAVIOR:"))
			} else {
				name = strings.TrimSpace(strings.TrimPrefix(trimmed, "## BEHAVIOR/INTERNAL:"))
			}
			
			// Find end of section
			endLine := len(lines) - 1
			for j := i + 1; j < len(lines); j++ {
				if strings.HasPrefix(strings.TrimSpace(lines[j]), "##") {
					endLine = j - 1
					break
				}
			}
			
			sections = append(sections, BehaviorSection{
				Name:      name,
				StartLine: i,
				EndLine:   endLine,
			})
		}
	}
	
	return sections
}