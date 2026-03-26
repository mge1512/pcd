package lint

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/pcdp/mcp-server-pcdp/internal/store"
)

// LintContent validates a PCDP specification and returns diagnostics
func LintContent(content string, filename string) *store.LintResult {
	result := &store.LintResult{
		Diagnostics: []store.Diagnostic{},
	}

	lines := strings.Split(content, "\n")

	// RULE-01: Check for required META section
	if !hasSection(content, "## META") {
		result.Diagnostics = append(result.Diagnostics, store.Diagnostic{
			Severity: "error",
			Line:     1,
			Section:  "META",
			Message:  "Missing required META section",
			Rule:     "RULE-01",
		})
		result.Errors++
	}

	// RULE-02: Check for required TYPES section
	if !hasSection(content, "## TYPES") {
		result.Diagnostics = append(result.Diagnostics, store.Diagnostic{
			Severity: "error",
			Line:     findSectionLine(lines, "## META") + 5,
			Section:  "structure",
			Message:  "Missing required TYPES section",
			Rule:     "RULE-02",
		})
		result.Errors++
	}

	// RULE-03: Check for required BEHAVIOR section
	if !hasSection(content, "## BEHAVIOR") {
		result.Diagnostics = append(result.Diagnostics, store.Diagnostic{
			Severity: "error",
			Line:     findSectionLine(lines, "## TYPES") + 5,
			Section:  "structure",
			Message:  "Missing required BEHAVIOR section",
			Rule:     "RULE-03",
		})
		result.Errors++
	}

	// RULE-04: Check for required PRECONDITIONS section
	if !hasSection(content, "## PRECONDITIONS") {
		result.Diagnostics = append(result.Diagnostics, store.Diagnostic{
			Severity: "warning",
			Line:     findSectionLine(lines, "## BEHAVIOR") + 10,
			Section:  "structure",
			Message:  "Missing PRECONDITIONS section",
			Rule:     "RULE-04",
		})
		result.Warnings++
	}

	// RULE-05: Check for required POSTCONDITIONS section
	if !hasSection(content, "## POSTCONDITIONS") {
		result.Diagnostics = append(result.Diagnostics, store.Diagnostic{
			Severity: "warning",
			Line:     findSectionLine(lines, "## PRECONDITIONS") + 5,
			Section:  "structure",
			Message:  "Missing POSTCONDITIONS section",
			Rule:     "RULE-05",
		})
		result.Warnings++
	}

	// RULE-06: Check for required INVARIANTS section
	if !hasSection(content, "## INVARIANTS") {
		result.Diagnostics = append(result.Diagnostics, store.Diagnostic{
			Severity: "error",
			Line:     findSectionLine(lines, "## POSTCONDITIONS") + 5,
			Section:  "structure",
			Message:  "Missing required INVARIANTS section",
			Rule:     "RULE-06",
		})
		result.Errors++
	}

	// RULE-07: Check for required EXAMPLES section
	if !hasSection(content, "## EXAMPLES") {
		result.Diagnostics = append(result.Diagnostics, store.Diagnostic{
			Severity: "warning",
			Line:     findSectionLine(lines, "## INVARIANTS") + 5,
			Section:  "structure",
			Message:  "Missing EXAMPLES section",
			Rule:     "RULE-07",
		})
		result.Warnings++
	}

	// RULE-08: Check for required DEPLOYMENT section
	if !hasSection(content, "## DEPLOYMENT") {
		result.Diagnostics = append(result.Diagnostics, store.Diagnostic{
			Severity: "warning",
			Line:     findSectionLine(lines, "## EXAMPLES") + 5,
			Section:  "structure",
			Message:  "Missing DEPLOYMENT section",
			Rule:     "RULE-08",
		})
		result.Warnings++
	}

	// RULE-09: Validate META section fields
	if hasSection(content, "## META") {
		metaSection := extractSection(content, "## META")
		requiredFields := []string{"Deployment:", "Version:", "Spec-Schema:", "Author:", "License:"}
		for _, field := range requiredFields {
			if !strings.Contains(metaSection, field) {
				result.Diagnostics = append(result.Diagnostics, store.Diagnostic{
					Severity: "error",
					Line:     findSectionLine(lines, "## META"),
					Section:  "META",
					Message:  fmt.Sprintf("Missing required field: %s", field),
					Rule:     "RULE-09",
				})
				result.Errors++
			}
		}
	}

	// RULE-10: Check BEHAVIOR blocks have required subsections
	behaviorBlocks := extractBehaviorBlocks(content)
	for _, block := range behaviorBlocks {
		if !strings.Contains(block, "INPUTS:") {
			result.Diagnostics = append(result.Diagnostics, store.Diagnostic{
				Severity: "error",
				Line:     findLineInContent(lines, block),
				Section:  "BEHAVIOR",
				Message:  "BEHAVIOR block missing INPUTS subsection",
				Rule:     "RULE-10",
			})
			result.Errors++
		}
		if !strings.Contains(block, "STEPS:") {
			result.Diagnostics = append(result.Diagnostics, store.Diagnostic{
				Severity: "error",
				Line:     findLineInContent(lines, block),
				Section:  "BEHAVIOR",
				Message:  "BEHAVIOR block missing STEPS subsection",
				Rule:     "RULE-10",
			})
			result.Errors++
		}
		if !strings.Contains(block, "POSTCONDITIONS:") {
			result.Diagnostics = append(result.Diagnostics, store.Diagnostic{
				Severity: "error",
				Line:     findLineInContent(lines, block),
				Section:  "BEHAVIOR",
				Message:  "BEHAVIOR block missing POSTCONDITIONS subsection",
				Rule:     "RULE-10",
			})
			result.Errors++
		}
		if !strings.Contains(block, "ERRORS:") {
			result.Diagnostics = append(result.Diagnostics, store.Diagnostic{
				Severity: "error",
				Line:     findLineInContent(lines, block),
				Section:  "BEHAVIOR",
				Message:  "BEHAVIOR block missing ERRORS subsection",
				Rule:     "RULE-10",
			})
			result.Errors++
		}
	}

	// RULE-11: Check for INVARIANT annotations
	if hasSection(content, "## INVARIANTS") {
		invariants := extractSection(content, "## INVARIANTS")
		lines := strings.Split(invariants, "\n")
		for i, line := range lines {
			if strings.HasPrefix(strings.TrimSpace(line), "-") {
				if !strings.Contains(line, "[observable]") && !strings.Contains(line, "[implementation]") {
					result.Diagnostics = append(result.Diagnostics, store.Diagnostic{
						Severity: "warning",
						Line:     findSectionLine(strings.Split(content, "\n"), "## INVARIANTS") + i,
						Section:  "INVARIANTS",
						Message:  "INVARIANT missing [observable] or [implementation] annotation",
						Rule:     "RULE-11",
					})
					result.Warnings++
				}
			}
		}
	}

	// RULE-12: Check EXAMPLES structure
	if hasSection(content, "## EXAMPLES") {
		examples := extractSection(content, "## EXAMPLES")
		if !strings.Contains(examples, "GIVEN:") {
			result.Diagnostics = append(result.Diagnostics, store.Diagnostic{
				Severity: "warning",
				Line:     findSectionLine(lines, "## EXAMPLES"),
				Section:  "EXAMPLES",
				Message:  "EXAMPLES should contain GIVEN clauses",
				Rule:     "RULE-12",
			})
			result.Warnings++
		}
		if !strings.Contains(examples, "WHEN:") {
			result.Diagnostics = append(result.Diagnostics, store.Diagnostic{
				Severity: "warning",
				Line:     findSectionLine(lines, "## EXAMPLES"),
				Section:  "EXAMPLES",
				Message:  "EXAMPLES should contain WHEN clauses",
				Rule:     "RULE-12",
			})
			result.Warnings++
		}
		if !strings.Contains(examples, "THEN:") {
			result.Diagnostics = append(result.Diagnostics, store.Diagnostic{
				Severity: "warning",
				Line:     findSectionLine(lines, "## EXAMPLES"),
				Section:  "EXAMPLES",
				Message:  "EXAMPLES should contain THEN clauses",
				Rule:     "RULE-12",
			})
			result.Warnings++
		}
	}

	// RULE-13: Check for valid semantic versioning in Version field
	if hasSection(content, "## META") {
		metaSection := extractSection(content, "## META")
		versionMatch := regexp.MustCompile(`Version:\s+(\S+)`).FindStringSubmatch(metaSection)
		if len(versionMatch) > 1 {
			version := versionMatch[1]
			if !isValidSemver(version) {
				result.Diagnostics = append(result.Diagnostics, store.Diagnostic{
					Severity: "error",
					Line:     findSectionLine(lines, "## META"),
					Section:  "META",
					Message:  fmt.Sprintf("Invalid semantic version: %s", version),
					Rule:     "RULE-13",
				})
				result.Errors++
			}
		}
	}

	// RULE-14: Check for valid Spec-Schema version
	if hasSection(content, "## META") {
		metaSection := extractSection(content, "## META")
		schemaMatch := regexp.MustCompile(`Spec-Schema:\s+(\S+)`).FindStringSubmatch(metaSection)
		if len(schemaMatch) > 1 {
			schema := schemaMatch[1]
			if !isValidSemver(schema) {
				result.Diagnostics = append(result.Diagnostics, store.Diagnostic{
					Severity: "error",
					Line:     findSectionLine(lines, "## META"),
					Section:  "META",
					Message:  fmt.Sprintf("Invalid Spec-Schema version: %s", schema),
					Rule:     "RULE-14",
				})
				result.Errors++
			}
		}
	}

	result.Valid = result.Errors == 0
	return result
}

// Helper functions

func hasSection(content, sectionName string) bool {
	return strings.Contains(content, sectionName)
}

func extractSection(content, sectionName string) string {
	parts := strings.SplitN(content, sectionName, 2)
	if len(parts) < 2 {
		return ""
	}
	section := parts[1]
	// Find next section marker
	nextSection := strings.Index(section, "\n## ")
	if nextSection > 0 {
		section = section[:nextSection]
	}
	return section
}

func findSectionLine(lines []string, sectionName string) int {
	for i, line := range lines {
		if strings.Contains(line, sectionName) {
			return i + 1 // 1-based line numbers
		}
	}
	return 1
}

func findLineInContent(lines []string, content string) int {
	// Find first line of content in lines
	contentFirstLine := strings.Split(content, "\n")[0]
	for i, line := range lines {
		if strings.Contains(line, contentFirstLine) {
			return i + 1
		}
	}
	return 1
}

func extractBehaviorBlocks(content string) []string {
	var blocks []string
	parts := strings.Split(content, "## BEHAVIOR:")
	for i := 1; i < len(parts); i++ {
		block := "## BEHAVIOR:" + parts[i]
		// Extract until next section
		if idx := strings.Index(block, "\n## "); idx > 0 {
			block = block[:idx]
		}
		blocks = append(blocks, block)
	}
	return blocks
}

func isValidSemver(version string) bool {
	// Simple semver check: MAJOR.MINOR.PATCH
	pattern := regexp.MustCompile(`^\d+\.\d+\.\d+$`)
	return pattern.MatchString(version)
}
