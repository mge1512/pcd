// generated from spec: pcd-lint.md sha256:293541ab62274835c61de50947f6283748831c4681cf3f02c4be2f8e942d28a9

// Package lint implements the pcd-lint specification validation logic.
// It applies RULE-01 through RULE-21 to specification files.
package lint

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/mge1512/pcd/tools/pcd-lint/internal/spdx"
)

// Severity of a diagnostic.
type Severity string

const (
	SeverityError   Severity = "ERROR"
	SeverityWarning Severity = "WARNING"
)

// Diagnostic represents a single lint finding.
type Diagnostic struct {
	Severity Severity
	Section  string
	Message  string
	Line     int
}

// LintResult is the result of linting a file.
type LintResult struct {
	File        string
	Diagnostics []Diagnostic
	ExitCode    int
}

// Options controls lint behaviour.
type Options struct {
	Strict      bool
	CheckReport bool
}

// knownDeployments is the set of valid DeploymentTemplate values.
var knownDeployments = map[string]bool{
	"wasm":              true,
	"ebpf":              true,
	"kernel-module":     true,
	"verified-library":  true,
	"cli-tool":          true,
	"gui-tool":          true,
	"cloud-native":      true,
	"backend-service":   true,
	"library-c-abi":     true,
	"enterprise-software": true,
	"academic":          true,
	"python-tool":       true,
	"enhance-existing":  true,
	"manual":            true,
	"template":          true,
	"mcp-server":        true,
	"project-manifest":  true,
}

// knownVerificationValues is the set of known Verification field values.
var knownVerificationValues = map[string]bool{
	"none":   true,
	"lean4":  true,
	"fstar":  true,
	"dafny":  true,
	"custom": true,
}

// specFile holds the parsed content of a specification file.
type specFile struct {
	path     string
	lines    []string
	sections []section
	meta     map[string][]metaField // key → values (for repeating keys like Author)
	// Parsed elements
	behaviorNames []string // all BEHAVIOR and BEHAVIOR/INTERNAL names
	milestones    []milestone
	exampleBlocks []exampleBlock
	includes      []includesRef
}

type section struct {
	name      string
	startLine int // 1-based
	endLine   int // 1-based (exclusive); 0 means end of file
}

type metaField struct {
	key   string
	value string
	line  int
}

type milestone struct {
	name     string
	startLine int
	fields   map[string]milestoneField
	scaffold string // "true", "false", or ""
	status   string
	included []string
	deferred []string
}

type milestoneField struct {
	value string
	line  int
}

type exampleBlock struct {
	name      string
	startLine int
	hasGiven  bool
	hasWhen   bool
	hasThen   bool
	givenEmpty bool
	whenEmpty  bool
	thenEmpty  bool
	whenWithoutThen bool // WHEN: not immediately followed by THEN:
}

type includesRef struct {
	value string
	line  int
}

// Lint validates the given file and returns a LintResult.
func Lint(path string, opts Options) LintResult {
	result := LintResult{File: path}

	// STEP 1: Verify .md extension
	if !strings.HasSuffix(path, ".md") {
		fmt.Fprintf(os.Stderr, "error: file must have .md extension: %s\n", path)
		result.ExitCode = 2
		return result
	}

	// STEP 2: Open and read file
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: cannot open file: %s\n", path)
		result.ExitCode = 2
		return result
	}

	lines := splitLines(string(data))
	sf := parseSpecFile(path, lines)

	var diags []Diagnostic

	// STEP 3: Apply RULE-01 through RULE-21
	diags = append(diags, applyRule01(sf)...)
	diags = append(diags, applyRule02(sf)...)
	diags = append(diags, applyRule02b(sf)...)
	diags = append(diags, applyRule02c(sf)...)
	diags = append(diags, applyRule02d(sf)...)
	diags = append(diags, applyRule02e(sf)...)
	diags = append(diags, applyRule03(sf)...)
	diags = append(diags, applyRule04(sf)...)
	diags = append(diags, applyRule05(sf)...)
	diags = append(diags, applyRule06(sf)...)
	diags = append(diags, applyRule07(sf)...)
	diags = append(diags, applyRule08(sf)...)
	diags = append(diags, applyRule09(sf)...)
	diags = append(diags, applyRule10(sf)...)
	diags = append(diags, applyRule11(sf)...)
	diags = append(diags, applyRule12(sf)...)
	diags = append(diags, applyRule13(sf)...)
	diags = append(diags, applyRule14(sf)...)
	diags = append(diags, applyRule15(sf)...)
	diags = append(diags, applyRule16(sf)...)
	diags = append(diags, applyRule17(sf)...)
	if opts.CheckReport {
		diags = append(diags, applyRule18(sf)...)
	}
	if len(sf.includes) > 0 {
		diags = append(diags, applyRule19(sf)...)
		diags = append(diags, applyRule20(sf)...)
		diags = append(diags, applyRule21(sf)...)
	}

	// STEP 4: Sort diagnostics by line number (monotonically non-decreasing)
	sort.SliceStable(diags, func(i, j int) bool {
		return diags[i].Line < diags[j].Line
	})

	result.Diagnostics = diags

	// STEP 6: Compute exit code
	hasError := false
	hasWarning := false
	for _, d := range diags {
		if d.Severity == SeverityError {
			hasError = true
		} else if d.Severity == SeverityWarning {
			hasWarning = true
		}
	}
	if hasError || (opts.Strict && hasWarning) {
		result.ExitCode = 1
	} else {
		result.ExitCode = 0
	}

	return result
}

// splitLines splits content into lines (without trailing newlines).
func splitLines(content string) []string {
	lines := strings.Split(content, "\n")
	// Remove trailing empty line from Split if content ends with \n
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	return lines
}

// parseSpecFile parses the spec file into a structured representation.
// It respects code-fence tracking (BEHAVIOR/INTERNAL: code-fence-tracking).
func parseSpecFile(path string, lines []string) *specFile {
	sf := &specFile{
		path:  path,
		lines: lines,
		meta:  make(map[string][]metaField),
	}

	// Parse sections, META, behaviors, milestones, examples
	fenceDepth := 0
	inMeta := false
	inExamples := false
	inMilestone := false
	inBehavior := false
	_ = inBehavior
	currentMilestone := (*milestone)(nil)
	currentBehavior := ""
	_ = currentBehavior
	behaviorSet := map[string]bool{}

	type sectionInfo struct {
		name      string
		startLine int
	}
	var currentSection *sectionInfo
	var openSections []sectionInfo

	for i, rawLine := range lines {
		lineNum := i + 1 // 1-based

		// Code-fence tracking (BEHAVIOR/INTERNAL: code-fence-tracking)
		trimmed := strings.TrimSpace(rawLine)
		if strings.HasPrefix(trimmed, "```") || strings.HasPrefix(trimmed, "~~~") {
			if fenceDepth == 0 {
				fenceDepth = 1
			} else {
				fenceDepth--
			}
			continue
		}
		if fenceDepth > 0 {
			continue
		}

		// Only structural markers at column 0 (no leading whitespace)
		// Exception: fence detection uses TrimSpace (done above)

		// Detect section headers: "## SOMETHING" at column 0
		if strings.HasPrefix(rawLine, "## ") {
			// Close previous section
			if currentSection != nil {
				sf.sections = append(sf.sections, section{
					name:      currentSection.name,
					startLine: currentSection.startLine,
					endLine:   lineNum,
				})
				openSections = append(openSections, *currentSection)
			}
			sectionName := rawLine[3:] // after "## "
			sectionName = strings.TrimSpace(sectionName)
			currentSection = &sectionInfo{name: rawLine, startLine: lineNum}

			// Track META section
			if sectionName == "META" {
				inMeta = true
				inExamples = false
				inMilestone = false
				inBehavior = false
			} else if sectionName == "EXAMPLES" {
				inMeta = false
				inExamples = true
				inMilestone = false
				inBehavior = false
			} else if strings.HasPrefix(sectionName, "MILESTONE:") {
				inMeta = false
				inExamples = false
				inMilestone = true
				inBehavior = false
				// Save previous milestone
				if currentMilestone != nil {
					sf.milestones = append(sf.milestones, *currentMilestone)
				}
				mName := strings.TrimSpace(strings.TrimPrefix(sectionName, "MILESTONE:"))
				currentMilestone = &milestone{
					name:      mName,
					startLine: lineNum,
					fields:    make(map[string]milestoneField),
				}
			} else if strings.HasPrefix(sectionName, "BEHAVIOR:") || strings.HasPrefix(sectionName, "BEHAVIOR/INTERNAL:") {
				inMeta = false
				inExamples = false
				inMilestone = false
				inBehavior = true
				// Extract behavior name
				var bName string
				if strings.HasPrefix(sectionName, "BEHAVIOR/INTERNAL:") {
					bName = strings.TrimSpace(strings.TrimPrefix(sectionName, "BEHAVIOR/INTERNAL:"))
				} else {
					bName = strings.TrimSpace(strings.TrimPrefix(sectionName, "BEHAVIOR:"))
				}
				if !behaviorSet[bName] {
					behaviorSet[bName] = true
					sf.behaviorNames = append(sf.behaviorNames, bName)
				}
				currentBehavior = bName
			} else {
				inMeta = false
				inMilestone = false
				inBehavior = false
				// Keep inExamples only if we're still within EXAMPLES section
				// Actually any ## heading closes examples
				inExamples = false
			}
			continue
		}

		// Parse META fields (key: value at column 0, within META section)
		if inMeta && strings.Contains(rawLine, ":") && !strings.HasPrefix(rawLine, " ") && !strings.HasPrefix(rawLine, "\t") && !strings.HasPrefix(rawLine, "#") && !strings.HasPrefix(rawLine, "-") {
			colonIdx := strings.Index(rawLine, ":")
			key := strings.TrimSpace(rawLine[:colonIdx])
			value := strings.TrimSpace(rawLine[colonIdx+1:])
			if key != "" {
				sf.meta[key] = append(sf.meta[key], metaField{key: key, value: value, line: lineNum})
			}
			// Track Includes:
			if key == "Includes" {
				sf.includes = append(sf.includes, includesRef{value: value, line: lineNum})
			}
		}

		// Parse MILESTONE fields
		if inMilestone && currentMilestone != nil {
			if strings.HasPrefix(rawLine, "Status:") {
				val := strings.TrimSpace(strings.TrimPrefix(rawLine, "Status:"))
				currentMilestone.status = val
				currentMilestone.fields["Status"] = milestoneField{value: val, line: lineNum}
			} else if strings.HasPrefix(rawLine, "Scaffold:") {
				val := strings.TrimSpace(strings.TrimPrefix(rawLine, "Scaffold:"))
				currentMilestone.scaffold = val
				currentMilestone.fields["Scaffold"] = milestoneField{value: val, line: lineNum}
			} else if strings.HasPrefix(rawLine, "Included BEHAVIORs:") {
				val := strings.TrimSpace(strings.TrimPrefix(rawLine, "Included BEHAVIORs:"))
				currentMilestone.fields["Included BEHAVIORs"] = milestoneField{value: val, line: lineNum}
				currentMilestone.included = parseBehaviorList(val)
			} else if strings.HasPrefix(rawLine, "Deferred BEHAVIORs:") {
				val := strings.TrimSpace(strings.TrimPrefix(rawLine, "Deferred BEHAVIORs:"))
				currentMilestone.fields["Deferred BEHAVIORs"] = milestoneField{value: val, line: lineNum}
				currentMilestone.deferred = parseBehaviorList(val)
			} else if strings.HasPrefix(rawLine, "Acceptance criteria:") {
				currentMilestone.fields["Acceptance criteria"] = milestoneField{value: "present", line: lineNum}
			}
		}

		// Parse EXAMPLES section
		if inExamples {
			// ### EXAMPLE: name at column 0
			if strings.HasPrefix(rawLine, "### EXAMPLE: ") {
				name := strings.TrimSpace(strings.TrimPrefix(rawLine, "### EXAMPLE: "))
				sf.exampleBlocks = append(sf.exampleBlocks, exampleBlock{
					name:      name,
					startLine: lineNum,
				})
			}
		}
	}

	// Close last section
	if currentSection != nil {
		sf.sections = append(sf.sections, section{
			name:      currentSection.name,
			startLine: currentSection.startLine,
			endLine:   len(lines) + 1,
		})
	}
	// Save last milestone
	if currentMilestone != nil {
		sf.milestones = append(sf.milestones, *currentMilestone)
	}

	// Parse example blocks (GIVEN/WHEN/THEN structure) with fence tracking
	sf.exampleBlocks = parseExampleBlocks(lines, sf.exampleBlocks)

	return sf
}

// parseBehaviorList splits a comma-separated list of behavior names.
func parseBehaviorList(val string) []string {
	if val == "" {
		return nil
	}
	parts := strings.Split(val, ",")
	var result []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}

// hasSection returns true if the spec contains a section matching the given prefix.
func (sf *specFile) hasSection(prefix string) bool {
	for i, rawLine := range sf.lines {
		if fenceDepthAt(sf.lines, i) > 0 {
			continue
		}
		if strings.HasPrefix(rawLine, prefix) {
			return true
		}
	}
	return false
}

// hasSectionExact returns true if the spec contains a section with the exact header.
func (sf *specFile) hasSectionExact(header string) bool {
	for i, rawLine := range sf.lines {
		if fenceDepthAt(sf.lines, i) > 0 {
			continue
		}
		if rawLine == header {
			return true
		}
	}
	return false
}

// fenceDepthAt computes the fence depth at a given line index (0-based).
// Used for checking fence status at specific lines.
func fenceDepthAt(lines []string, targetIdx int) int {
	depth := 0
	for i := 0; i < targetIdx; i++ {
		trimmed := strings.TrimSpace(lines[i])
		if strings.HasPrefix(trimmed, "```") || strings.HasPrefix(trimmed, "~~~") {
			if depth == 0 {
				depth = 1
			} else {
				depth--
			}
		}
	}
	return depth
}

// metaValue returns the first value for a META key, or "" if not present.
func (sf *specFile) metaValue(key string) string {
	vals := sf.meta[key]
	if len(vals) == 0 {
		return ""
	}
	return vals[0].value
}

// metaLine returns the line number of the first META key, or 1 if not present.
func (sf *specFile) metaLine(key string) int {
	vals := sf.meta[key]
	if len(vals) == 0 {
		return 1
	}
	return vals[0].line
}

// metaPresent returns true if the META field is present (even with empty value).
func (sf *specFile) metaPresent(key string) bool {
	_, ok := sf.meta[key]
	return ok
}

// findSectionLine returns the line number of a section header, or 1 if not found.
func (sf *specFile) findSectionLine(prefix string) int {
	for i, rawLine := range sf.lines {
		if fenceDepthAt(sf.lines, i) > 0 {
			continue
		}
		if strings.HasPrefix(rawLine, prefix) {
			return i + 1
		}
	}
	return 1
}

// ---------------------------------------------------------------------------
// RULE-01: Required sections present
// ---------------------------------------------------------------------------

func applyRule01(sf *specFile) []Diagnostic {
	var diags []Diagnostic
	required := []string{
		"## META",
		"## TYPES",
		"## BEHAVIOR",      // matched by prefix (BEHAVIOR: or BEHAVIOR/INTERNAL:)
		"## PRECONDITIONS",
		"## POSTCONDITIONS",
		"## INVARIANTS",
		"## EXAMPLES",
	}
	for _, sec := range required {
		if sec == "## BEHAVIOR" {
			// Satisfied by presence of "## BEHAVIOR:" or "## BEHAVIOR/INTERNAL:"
			if !sf.hasSection("## BEHAVIOR: ") && !sf.hasSection("## BEHAVIOR/INTERNAL: ") {
				diags = append(diags, Diagnostic{
					Severity: SeverityError,
					Section:  "structure",
					Line:     1,
					Message:  fmt.Sprintf("Missing required section: ## BEHAVIOR"),
				})
			}
		} else {
			if !sf.hasSectionExact(sec) {
				diags = append(diags, Diagnostic{
					Severity: SeverityError,
					Section:  "structure",
					Line:     1,
					Message:  fmt.Sprintf("Missing required section: %s", sec),
				})
			}
		}
	}
	return diags
}

// ---------------------------------------------------------------------------
// RULE-02: META fields present and non-empty
// ---------------------------------------------------------------------------

func applyRule02(sf *specFile) []Diagnostic {
	var diags []Diagnostic
	required := []string{"Deployment", "Verification", "Safety-Level", "Version", "Spec-Schema", "License"}
	for _, field := range required {
		if !sf.metaPresent(field) {
			diags = append(diags, Diagnostic{
				Severity: SeverityError,
				Section:  "META",
				Line:     sf.findSectionLine("## META"),
				Message:  fmt.Sprintf("Missing required META field: %s", field),
			})
		} else if sf.metaValue(field) == "" {
			diags = append(diags, Diagnostic{
				Severity: SeverityError,
				Section:  "META",
				Line:     sf.metaLine(field),
				Message:  fmt.Sprintf("META field %s has empty value", field),
			})
		}
	}
	return diags
}

// ---------------------------------------------------------------------------
// RULE-02b: Author field
// ---------------------------------------------------------------------------

func applyRule02b(sf *specFile) []Diagnostic {
	var diags []Diagnostic
	authors := sf.meta["Author"]
	if len(authors) == 0 {
		diags = append(diags, Diagnostic{
			Severity: SeverityError,
			Section:  "META",
			Line:     sf.findSectionLine("## META"),
			Message:  "Missing required META field: Author (at least one Author: line required)",
		})
	} else {
		for _, a := range authors {
			if a.value == "" {
				diags = append(diags, Diagnostic{
					Severity: SeverityError,
					Section:  "META",
					Line:     a.line,
					Message:  "Author: field has empty value",
				})
			}
		}
	}
	return diags
}

// ---------------------------------------------------------------------------
// RULE-02c: Version format
// ---------------------------------------------------------------------------

func applyRule02c(sf *specFile) []Diagnostic {
	var diags []Diagnostic
	v := sf.metaValue("Version")
	if v != "" && !isSemanticVersion(v) {
		diags = append(diags, Diagnostic{
			Severity: SeverityError,
			Section:  "META",
			Line:     sf.metaLine("Version"),
			Message: fmt.Sprintf(
				"Version '%s' is not valid semantic versioning. Required format: MAJOR.MINOR.PATCH (e.g. 0.1.0)", v),
		})
	}
	return diags
}

// ---------------------------------------------------------------------------
// RULE-02d: Spec-Schema version
// ---------------------------------------------------------------------------

func applyRule02d(sf *specFile) []Diagnostic {
	var diags []Diagnostic
	s := sf.metaValue("Spec-Schema")
	if s != "" && !isSemanticVersion(s) {
		diags = append(diags, Diagnostic{
			Severity: SeverityError,
			Section:  "META",
			Line:     sf.metaLine("Spec-Schema"),
			Message: fmt.Sprintf(
				"Spec-Schema '%s' is not valid semantic versioning. Required format: MAJOR.MINOR.PATCH (e.g. 0.1.0)", s),
		})
	}
	return diags
}

// ---------------------------------------------------------------------------
// RULE-02e: License SPDX validation
// ---------------------------------------------------------------------------

func applyRule02e(sf *specFile) []Diagnostic {
	var diags []Diagnostic
	l := sf.metaValue("License")
	if l != "" && !spdx.IsValid(l) {
		diags = append(diags, Diagnostic{
			Severity: SeverityError,
			Section:  "META",
			Line:     sf.metaLine("License"),
			Message: fmt.Sprintf(
				"License '%s' is not a valid SPDX identifier. See https://spdx.org/licenses/ for valid identifiers. Compound expressions permitted (e.g. Apache-2.0 OR MIT).", l),
		})
	}
	return diags
}

// ---------------------------------------------------------------------------
// RULE-03: Deployment template resolves
// ---------------------------------------------------------------------------

func applyRule03(sf *specFile) []Diagnostic {
	var diags []Diagnostic
	d := sf.metaValue("Deployment")
	if d == "" {
		return diags // RULE-02 already reported it
	}

	if d == "crypto-library" {
		diags = append(diags, Diagnostic{
			Severity: SeverityError,
			Section:  "META",
			Line:     1,
			Message: "Deployment 'crypto-library' was retired in 0.3.6. Use 'verified-library' instead. " +
				"verified-library covers all safety- and security-critical C-ABI libraries including cryptographic primitives.",
		})
		return diags
	}

	if !knownDeployments[d] {
		diags = append(diags, Diagnostic{
			Severity: SeverityError,
			Section:  "META",
			Line:     sf.metaLine("Deployment"),
			Message: fmt.Sprintf(
				"Unknown deployment template: '%s'. Run 'pcd-lint list-templates' to see valid values.", d),
		})
		return diags
	}

	if d == "enhance-existing" {
		if !sf.metaPresent("Language") {
			diags = append(diags, Diagnostic{
				Severity: SeverityError,
				Section:  "META",
				Line:     sf.metaLine("Deployment"),
				Message:  "Deployment 'enhance-existing' requires META field 'Language'",
			})
		} else if sf.metaValue("Language") == "" {
			diags = append(diags, Diagnostic{
				Severity: SeverityError,
				Section:  "META",
				Line:     sf.metaLine("Language"),
				Message:  "META field 'Language' has empty value",
			})
		}
	}

	if d == "manual" {
		if !sf.metaPresent("Target") {
			diags = append(diags, Diagnostic{
				Severity: SeverityError,
				Section:  "META",
				Line:     sf.metaLine("Deployment"),
				Message:  "Deployment 'manual' requires META field 'Target' (no template available for language resolution)",
			})
		}
	}

	if d == "python-tool" {
		sl := sf.metaValue("Safety-Level")
		if sl != "QM" {
			diags = append(diags, Diagnostic{
				Severity: SeverityError,
				Section:  "META",
				Line:     sf.metaLine("Safety-Level"),
				Message: "Deployment 'python-tool' requires Safety-Level: QM. " +
					"Python is not suitable for safety-critical components.",
			})
		}
		v := sf.metaValue("Verification")
		if v != "none" {
			diags = append(diags, Diagnostic{
				Severity: SeverityError,
				Section:  "META",
				Line:     sf.metaLine("Verification"),
				Message: "Deployment 'python-tool' requires Verification: none. " +
					"No formal verification path exists for Python.",
			})
		}
	}

	if d == "verified-library" {
		sl := sf.metaValue("Safety-Level")
		if sl == "QM" {
			diags = append(diags, Diagnostic{
				Severity: SeverityWarning,
				Section:  "META",
				Line:     sf.metaLine("Safety-Level"),
				Message: "Deployment 'verified-library' with Safety-Level: QM is unusual. " +
					"verified-library is intended for safety- or security-critical components. " +
					"Consider using library-c-abi for general-purpose libraries.",
			})
		}
	}

	return diags
}

// ---------------------------------------------------------------------------
// RULE-04: Deprecated META fields
// ---------------------------------------------------------------------------

func applyRule04(sf *specFile) []Diagnostic {
	var diags []Diagnostic
	d := sf.metaValue("Deployment")

	if sf.metaPresent("Target") && d != "manual" {
		diags = append(diags, Diagnostic{
			Severity: SeverityWarning,
			Section:  "META",
			Line:     sf.metaLine("Target"),
			Message: "META field 'Target' is deprecated since v0.3.0. " +
				"Target language is derived from the deployment template. " +
				"Remove 'Target', or switch to Deployment: manual if explicit language control is required.",
		})
	}

	if sf.metaPresent("Domain") {
		diags = append(diags, Diagnostic{
			Severity: SeverityWarning,
			Section:  "META",
			Line:     sf.metaLine("Domain"),
			Message: "META field 'Domain' is deprecated since v0.3.0. " +
				"Use 'Deployment' instead.",
		})
	}

	return diags
}

// ---------------------------------------------------------------------------
// RULE-05: Verification field value
// ---------------------------------------------------------------------------

func applyRule05(sf *specFile) []Diagnostic {
	var diags []Diagnostic
	v := sf.metaValue("Verification")
	if v != "" && !knownVerificationValues[v] {
		diags = append(diags, Diagnostic{
			Severity: SeverityWarning,
			Section:  "META",
			Line:     sf.metaLine("Verification"),
			Message: fmt.Sprintf(
				"Unknown verification value: '%s'. Known values: none, lean4, fstar, dafny, custom. "+
					"Custom verification backends are permitted; verify the value is intentional.", v),
		})
	}
	return diags
}

// ---------------------------------------------------------------------------
// RULE-06: EXAMPLES section structure
// ---------------------------------------------------------------------------

func applyRule06(sf *specFile) []Diagnostic {
	var diags []Diagnostic

	// Find the EXAMPLES section
	if !sf.hasSectionExact("## EXAMPLES") {
		return diags // RULE-01 will report missing section
	}

	examplesLine := sf.findSectionLine("## EXAMPLES")

	// Check for flat form "EXAMPLE: " (not "### EXAMPLE: ") inside EXAMPLES section
	// Check for wrong heading level "## EXAMPLE: " or "#### EXAMPLE: "
	inExamples := false
	fenceDepth := 0
	for i, rawLine := range sf.lines {
		lineNum := i + 1

		trimmed := strings.TrimSpace(rawLine)
		if strings.HasPrefix(trimmed, "```") || strings.HasPrefix(trimmed, "~~~") {
			if fenceDepth == 0 {
				fenceDepth = 1
			} else {
				fenceDepth--
			}
			continue
		}
		if fenceDepth > 0 {
			continue
		}

		if rawLine == "## EXAMPLES" {
			inExamples = true
			continue
		}
		if strings.HasPrefix(rawLine, "## ") && rawLine != "## EXAMPLES" && inExamples {
			inExamples = false
		}

		if !inExamples {
			continue
		}

		// Check for flat form: "EXAMPLE: " at column 0 (no heading markers)
		if strings.HasPrefix(rawLine, "EXAMPLE: ") {
			diags = append(diags, Diagnostic{
				Severity: SeverityError,
				Section:  "EXAMPLES",
				Line:     lineNum,
				Message: "Example header must be '### EXAMPLE: <name>'. " +
					"Flat 'EXAMPLE: <name>' is no longer accepted (changed in v0.4.0). " +
					"See: doc/spec-composition.md or RULE-06.",
			})
		}

		// Check for wrong heading level: "## EXAMPLE: " or "#### EXAMPLE: "
		if strings.HasPrefix(rawLine, "## EXAMPLE: ") {
			name := strings.TrimSpace(strings.TrimPrefix(rawLine, "## EXAMPLE: "))
			diags = append(diags, Diagnostic{
				Severity: SeverityError,
				Section:  "EXAMPLES",
				Line:     lineNum,
				Message: fmt.Sprintf(
					"Example header must be at heading level 3 (three pound signs). Found heading level 2 for '%s'.", name),
			})
		}
		if strings.HasPrefix(rawLine, "#### EXAMPLE: ") {
			name := strings.TrimSpace(strings.TrimPrefix(rawLine, "#### EXAMPLE: "))
			diags = append(diags, Diagnostic{
				Severity: SeverityError,
				Section:  "EXAMPLES",
				Line:     lineNum,
				Message: fmt.Sprintf(
					"Example header must be at heading level 3 (three pound signs). Found heading level 4 for '%s'.", name),
			})
		}
		if strings.HasPrefix(rawLine, "##### EXAMPLE: ") {
			name := strings.TrimSpace(strings.TrimPrefix(rawLine, "##### EXAMPLE: "))
			diags = append(diags, Diagnostic{
				Severity: SeverityError,
				Section:  "EXAMPLES",
				Line:     lineNum,
				Message: fmt.Sprintf(
					"Example header must be at heading level 3 (three pound signs). Found heading level 5 for '%s'.", name),
			})
		}
	}

	// Check that at least one example block exists
	if len(sf.exampleBlocks) == 0 {
		diags = append(diags, Diagnostic{
			Severity: SeverityError,
			Section:  "EXAMPLES",
			Line:     examplesLine,
			Message: "EXAMPLES section contains no example blocks. " +
				"Each example requires ### EXAMPLE: heading and GIVEN:, WHEN:, THEN: markers.",
		})
		return diags
	}

	// Check each example block for GIVEN/WHEN/THEN
	for _, ex := range sf.exampleBlocks {
		if !ex.hasGiven {
			diags = append(diags, Diagnostic{
				Severity: SeverityError,
				Section:  "EXAMPLES",
				Line:     ex.startLine,
				Message:  fmt.Sprintf("Example '%s' missing GIVEN: marker", ex.name),
			})
		}
		if !ex.hasWhen {
			diags = append(diags, Diagnostic{
				Severity: SeverityError,
				Section:  "EXAMPLES",
				Line:     ex.startLine,
				Message:  fmt.Sprintf("Example '%s' missing WHEN: marker", ex.name),
			})
		}
		if !ex.hasThen {
			diags = append(diags, Diagnostic{
				Severity: SeverityError,
				Section:  "EXAMPLES",
				Line:     ex.startLine,
				Message:  fmt.Sprintf("Example '%s' missing THEN: marker", ex.name),
			})
		}
		if ex.whenWithoutThen {
			diags = append(diags, Diagnostic{
				Severity: SeverityError,
				Section:  "EXAMPLES",
				Line:     ex.startLine,
				Message:  fmt.Sprintf("Example '%s' has WHEN: without a matching THEN:", ex.name),
			})
		}
	}

	return diags
}

// ---------------------------------------------------------------------------
// RULE-07: EXAMPLES minimum content
// ---------------------------------------------------------------------------

func applyRule07(sf *specFile) []Diagnostic {
	var diags []Diagnostic
	for _, ex := range sf.exampleBlocks {
		if ex.givenEmpty {
			diags = append(diags, Diagnostic{
				Severity: SeverityWarning,
				Section:  "EXAMPLES",
				Line:     ex.startLine,
				Message:  fmt.Sprintf("Example '%s' has empty GIVEN block", ex.name),
			})
		}
		if ex.whenEmpty {
			diags = append(diags, Diagnostic{
				Severity: SeverityWarning,
				Section:  "EXAMPLES",
				Line:     ex.startLine,
				Message:  fmt.Sprintf("Example '%s' has empty WHEN block", ex.name),
			})
		}
		if ex.thenEmpty {
			diags = append(diags, Diagnostic{
				Severity: SeverityWarning,
				Section:  "EXAMPLES",
				Line:     ex.startLine,
				Message:  fmt.Sprintf("Example '%s' has empty THEN block", ex.name),
			})
		}
	}
	return diags
}

// ---------------------------------------------------------------------------
// RULE-08: BEHAVIOR blocks must contain STEPS
// ---------------------------------------------------------------------------

func applyRule08(sf *specFile) []Diagnostic {
	var diags []Diagnostic

	// Find each BEHAVIOR/BEHAVIOR/INTERNAL section and check for STEPS:
	behaviorSections := extractBehaviorSections(sf)
	for _, bs := range behaviorSections {
		hasSteps := false
		fenceDepth := 0
		for _, rawLine := range bs.lines {
			trimmed := strings.TrimSpace(rawLine)
			if strings.HasPrefix(trimmed, "```") || strings.HasPrefix(trimmed, "~~~") {
				if fenceDepth == 0 {
					fenceDepth = 1
				} else {
					fenceDepth--
				}
				continue
			}
			if fenceDepth > 0 {
				continue
			}
			if rawLine == "STEPS:" {
				hasSteps = true
				break
			}
		}
		if !hasSteps {
			diags = append(diags, Diagnostic{
				Severity: SeverityError,
				Section:  bs.sectionHeader,
				Line:     bs.startLine,
				Message: fmt.Sprintf(
					"BEHAVIOR '%s' is missing required STEPS: block. Every BEHAVIOR must include ordered, imperative STEPS.", bs.name),
			})
		}
	}
	return diags
}

// ---------------------------------------------------------------------------
// RULE-09: INVARIANTS entries should carry observable/implementation tags
// ---------------------------------------------------------------------------

func applyRule09(sf *specFile) []Diagnostic {
	var diags []Diagnostic

	if !sf.hasSectionExact("## INVARIANTS") {
		return diags
	}

	inInvariants := false
	fenceDepth := 0
	for i, rawLine := range sf.lines {
		lineNum := i + 1
		trimmed := strings.TrimSpace(rawLine)
		if strings.HasPrefix(trimmed, "```") || strings.HasPrefix(trimmed, "~~~") {
			if fenceDepth == 0 {
				fenceDepth = 1
			} else {
				fenceDepth--
			}
			continue
		}
		if fenceDepth > 0 {
			continue
		}

		if rawLine == "## INVARIANTS" {
			inInvariants = true
			continue
		}
		if strings.HasPrefix(rawLine, "## ") && inInvariants {
			break
		}
		if !inInvariants {
			continue
		}

		// An entry line: non-empty, non-heading, not a separator (---)
		if rawLine == "" || strings.HasPrefix(rawLine, "#") || rawLine == "---" {
			continue
		}
		// Check if it's a list entry line
		if !strings.HasPrefix(rawLine, "-") {
			continue
		}

		// Check for [observable] or [implementation] tag
		if !strings.HasPrefix(rawLine, "- [observable]") && !strings.HasPrefix(rawLine, "- [implementation]") {
			diags = append(diags, Diagnostic{
				Severity: SeverityWarning,
				Section:  "INVARIANTS",
				Line:     lineNum,
				Message: "Invariant entry missing tag. " +
					"Prefix with [observable] or [implementation] for audit utility.",
			})
		}
	}
	return diags
}

// ---------------------------------------------------------------------------
// RULE-10: Negative-path EXAMPLE required for BEHAVIOR with error exits
// ---------------------------------------------------------------------------

func applyRule10(sf *specFile) []Diagnostic {
	var diags []Diagnostic

	behaviorSections := extractBehaviorSections(sf)
	for _, bs := range behaviorSections {
		hasErrorExits := false
		fenceDepth := 0
		inSteps := false
		for _, rawLine := range bs.lines {
			trimmed := strings.TrimSpace(rawLine)
			if strings.HasPrefix(trimmed, "```") || strings.HasPrefix(trimmed, "~~~") {
				if fenceDepth == 0 {
					fenceDepth = 1
				} else {
					fenceDepth--
				}
				continue
			}
			if fenceDepth > 0 {
				continue
			}
			if rawLine == "STEPS:" {
				inSteps = true
				continue
			}
			// STEPS block ends at next non-indented non-empty structural line
			if inSteps && rawLine != "" && !strings.HasPrefix(rawLine, " ") && !strings.HasPrefix(rawLine, "\t") &&
				rawLine != "STEPS:" && strings.HasPrefix(rawLine, "##") {
				inSteps = false
			}
			if inSteps && strings.Contains(rawLine, "→") {
				// Check for error exit notation: "on failure →", "→ exit N", "→ return Err", "→ Err("
				if strings.Contains(rawLine, "on failure →") ||
					strings.Contains(rawLine, "→ exit ") ||
					strings.Contains(rawLine, "→ return Err") ||
					strings.Contains(rawLine, "→ Err(") {
					hasErrorExits = true
				}
			}
		}

		if hasErrorExits {
			negativeFound := findNegativeExample(sf, bs.name)
			if !negativeFound {
				diags = append(diags, Diagnostic{
					Severity: SeverityError,
					Section:  bs.sectionHeader,
					Line:     bs.startLine,
					Message: fmt.Sprintf(
						"BEHAVIOR '%s' has error exits in STEPS but no negative-path EXAMPLE. "+
							"Add at least one EXAMPLE whose THEN: verifies an error outcome.", bs.name),
				})
			}
		}
	}
	return diags
}

// ---------------------------------------------------------------------------
// RULE-11: TOOLCHAIN-CONSTRAINTS section structure
// ---------------------------------------------------------------------------

func applyRule11(sf *specFile) []Diagnostic {
	var diags []Diagnostic

	if !sf.hasSection("## TOOLCHAIN-CONSTRAINTS") {
		return diags
	}

	inSection := false
	fenceDepth := 0
	for i, rawLine := range sf.lines {
		lineNum := i + 1
		trimmed := strings.TrimSpace(rawLine)
		if strings.HasPrefix(trimmed, "```") || strings.HasPrefix(trimmed, "~~~") {
			if fenceDepth == 0 {
				fenceDepth = 1
			} else {
				fenceDepth--
			}
			continue
		}
		if fenceDepth > 0 {
			continue
		}

		if strings.HasPrefix(rawLine, "## TOOLCHAIN-CONSTRAINTS") {
			inSection = true
			continue
		}
		if strings.HasPrefix(rawLine, "## ") && inSection {
			break
		}
		if !inSection {
			continue
		}
		if rawLine == "" || rawLine == "---" {
			continue
		}

		// Check for constraint values other than "required" or "forbidden"
		lower := strings.ToLower(rawLine)
		if strings.Contains(rawLine, ":") {
			// Extract value after colon
			colonIdx := strings.Index(rawLine, ":")
			if colonIdx >= 0 {
				val := strings.TrimSpace(rawLine[colonIdx+1:])
				if val != "" && val != "required" && val != "forbidden" {
					_ = lineNum
					// Only warn if the value looks like a constraint declaration
					if strings.HasPrefix(strings.TrimSpace(rawLine), "-") {
						_ = lower
						// Check if it's a constraint-like pattern
					}
				}
			}
		}
	}
	return diags
}

// ---------------------------------------------------------------------------
// RULE-12: Cross-section consistency (partial)
// ---------------------------------------------------------------------------

func applyRule12(sf *specFile) []Diagnostic {
	var diags []Diagnostic
	// 12b: Type name consistency — check if type names from TYPES are redefined in BEHAVIOR
	typeNames := collectTypeNames(sf)
	behaviorSections := extractBehaviorSections(sf)
	for _, bs := range behaviorSections {
		for _, rawLine := range bs.lines {
			for _, typeName := range typeNames {
				if strings.HasPrefix(rawLine, typeName+" :=") {
					diags = append(diags, Diagnostic{
						Severity: SeverityError,
						Section:  bs.sectionHeader,
						Line:     bs.startLine,
						Message: fmt.Sprintf(
							"Type '%s' declared in TYPES is redefined in BEHAVIOR. Types must be declared in TYPES only.", typeName),
					})
				}
			}
		}
	}
	return diags
}

// ---------------------------------------------------------------------------
// RULE-13: Constraint: field value on BEHAVIOR headers
// ---------------------------------------------------------------------------

func applyRule13(sf *specFile) []Diagnostic {
	var diags []Diagnostic
	validConstraints := map[string]bool{
		"required":  true,
		"supported": true,
		"forbidden": true,
	}

	behaviorSections := extractBehaviorSections(sf)
	for _, bs := range behaviorSections {
		fenceDepth := 0
		passedSteps := false
		for _, rawLine := range bs.lines {
			trimmed := strings.TrimSpace(rawLine)
			if strings.HasPrefix(trimmed, "```") || strings.HasPrefix(trimmed, "~~~") {
				if fenceDepth == 0 {
					fenceDepth = 1
				} else {
					fenceDepth--
				}
				continue
			}
			if fenceDepth > 0 {
				continue
			}

			if rawLine == "STEPS:" {
				passedSteps = true
			}

			// Only check Constraint: before STEPS: (header-level field)
			// and only if it's at column 0 (not indented, not in a subsection)
			if !passedSteps && strings.HasPrefix(rawLine, "Constraint:") && !strings.HasPrefix(rawLine, "#") {
				val := strings.TrimSpace(strings.TrimPrefix(rawLine, "Constraint:"))
				if !validConstraints[val] {
					diags = append(diags, Diagnostic{
						Severity: SeverityError,
						Section:  bs.sectionHeader,
						Line:     bs.startLine,
						Message: fmt.Sprintf(
							"BEHAVIOR '%s' has invalid Constraint: value '%s'. Valid values: required, supported, forbidden.", bs.name, val),
					})
				}
				if val == "forbidden" {
					// Check for reason: annotation
					hasReason := false
					for _, rl := range bs.lines {
						if strings.HasPrefix(rl, "  reason:") {
							hasReason = true
							break
						}
					}
					if !hasReason {
						diags = append(diags, Diagnostic{
							Severity: SeverityWarning,
							Section:  bs.sectionHeader,
							Line:     bs.startLine,
							Message: fmt.Sprintf(
								"BEHAVIOR '%s' is Constraint: forbidden but has no reason: annotation.", bs.name),
						})
					}
				}
			}
		}
	}
	return diags
}

// ---------------------------------------------------------------------------
// RULE-14: EXECUTION section required in deployment templates
// ---------------------------------------------------------------------------

func applyRule14(sf *specFile) []Diagnostic {
	var diags []Diagnostic

	d := sf.metaValue("Deployment")
	if d != "template" {
		return diags
	}

	// Check for EXECUTION: none in META
	if sf.metaValue("EXECUTION") == "none" {
		return diags
	}

	if !sf.hasSection("## EXECUTION") {
		diags = append(diags, Diagnostic{
			Severity: SeverityWarning,
			Section:  "structure",
			Line:     1,
			Message: "Deployment template is missing ## EXECUTION section. " +
				"Translators cannot determine delivery phases without it. " +
				"Add ## EXECUTION or declare 'EXECUTION: none' in META if this template intentionally has no execution recipe.",
		})
		return diags
	}

	// Check EXECUTION section content
	execContent := sf.extractSectionContent("## EXECUTION")
	if !strings.Contains(execContent, "### Delivery phases") {
		diags = append(diags, Diagnostic{
			Severity: SeverityWarning,
			Section:  "EXECUTION",
			Line:     sf.findSectionLine("## EXECUTION"),
			Message:  "## EXECUTION section has no '### Delivery phases' subsection.",
		})
	}
	if !strings.Contains(execContent, "### Compile gate") && !strings.Contains(execContent, "COMPILE-GATE: none") {
		diags = append(diags, Diagnostic{
			Severity: SeverityWarning,
			Section:  "EXECUTION",
			Line:     sf.findSectionLine("## EXECUTION"),
			Message: "## EXECUTION section has no '### Compile gate' subsection " +
				"and does not declare 'COMPILE-GATE: none'. " +
				"Translators will not know how to verify compilation.",
		})
	}
	if !strings.Contains(execContent, "### Resume logic") {
		diags = append(diags, Diagnostic{
			Severity: SeverityWarning,
			Section:  "EXECUTION",
			Line:     sf.findSectionLine("## EXECUTION"),
			Message:  "## EXECUTION section has no '### Resume logic' subsection.",
		})
	}

	return diags
}

// ---------------------------------------------------------------------------
// RULE-15: MILESTONE section structure and single-active constraint
// ---------------------------------------------------------------------------

func applyRule15(sf *specFile) []Diagnostic {
	var diags []Diagnostic

	if len(sf.milestones) == 0 {
		return diags
	}

	activeMilestones := 0
	for _, m := range sf.milestones {
		// Check Included BEHAVIORs
		if _, ok := m.fields["Included BEHAVIORs"]; !ok {
			diags = append(diags, Diagnostic{
				Severity: SeverityError,
				Section:  fmt.Sprintf("MILESTONE: %s", m.name),
				Line:     m.startLine,
				Message:  fmt.Sprintf("MILESTONE '%s' is missing required 'Included BEHAVIORs:' field.", m.name),
			})
		}

		// Check Deferred BEHAVIORs (required unless Scaffold: true)
		_, hasDeferred := m.fields["Deferred BEHAVIORs"]
		if !hasDeferred && m.scaffold != "true" {
			diags = append(diags, Diagnostic{
				Severity: SeverityError,
				Section:  fmt.Sprintf("MILESTONE: %s", m.name),
				Line:     m.startLine,
				Message: fmt.Sprintf("MILESTONE '%s' is missing required 'Deferred BEHAVIORs:' field. "+
					"Omit this field only when Scaffold: true (scaffold milestones have no deferred BEHAVIORs by definition).", m.name),
			})
		}

		// Check Acceptance criteria (warning)
		if _, ok := m.fields["Acceptance criteria"]; !ok {
			diags = append(diags, Diagnostic{
				Severity: SeverityWarning,
				Section:  fmt.Sprintf("MILESTONE: %s", m.name),
				Line:     m.startLine,
				Message: fmt.Sprintf("MILESTONE '%s' has no 'Acceptance criteria:' field. "+
					"Translators and agents cannot verify completion.", m.name),
			})
		}

		// Check Status
		if _, ok := m.fields["Status"]; !ok {
			diags = append(diags, Diagnostic{
				Severity: SeverityWarning,
				Section:  fmt.Sprintf("MILESTONE: %s", m.name),
				Line:     m.startLine,
				Message: fmt.Sprintf("MILESTONE '%s' has no Status: field. "+
					"Expected: pending | active | failed | released.", m.name),
			})
		} else {
			validStatuses := map[string]bool{
				"pending": true, "active": true, "failed": true, "released": true,
			}
			if !validStatuses[m.status] {
				diags = append(diags, Diagnostic{
					Severity: SeverityError,
					Section:  fmt.Sprintf("MILESTONE: %s", m.name),
					Line:     m.startLine,
					Message: fmt.Sprintf("MILESTONE '%s' has invalid Status: value '%s'. "+
						"Valid values: pending, active, failed, released.", m.name, m.status),
				})
			}
		}

		// Check Scaffold value if present
		if sf_, ok := m.fields["Scaffold"]; ok {
			if sf_.value != "true" && sf_.value != "false" {
				diags = append(diags, Diagnostic{
					Severity: SeverityError,
					Section:  fmt.Sprintf("MILESTONE: %s", m.name),
					Line:     m.startLine,
					Message: fmt.Sprintf("MILESTONE '%s' has invalid Scaffold: value '%s'. "+
						"Valid values: true, false.", m.name, sf_.value),
				})
			}
		}

		if m.status == "active" {
			activeMilestones++
		}
	}

	if activeMilestones > 1 {
		diags = append(diags, Diagnostic{
			Severity: SeverityError,
			Section:  "structure",
			Line:     1,
			Message: "More than one MILESTONE has Status: active. " +
				"Exactly one milestone may be active at a time.",
		})
	}

	return diags
}

// ---------------------------------------------------------------------------
// RULE-16: MILESTONE BEHAVIOR names exist in spec
// ---------------------------------------------------------------------------

func applyRule16(sf *specFile) []Diagnostic {
	var diags []Diagnostic

	if len(sf.milestones) == 0 {
		return diags
	}

	behaviorSet := map[string]bool{}
	for _, name := range sf.behaviorNames {
		behaviorSet[name] = true
	}

	for _, m := range sf.milestones {
		for _, name := range m.included {
			if !behaviorSet[name] {
				diags = append(diags, Diagnostic{
					Severity: SeverityError,
					Section:  fmt.Sprintf("MILESTONE: %s", m.name),
					Line:     m.startLine,
					Message: fmt.Sprintf(
						"MILESTONE '%s' lists BEHAVIOR '%s' under Included BEHAVIORs but no such BEHAVIOR exists in the spec.",
						m.name, name),
				})
			}
		}
		for _, name := range m.deferred {
			if !behaviorSet[name] {
				diags = append(diags, Diagnostic{
					Severity: SeverityError,
					Section:  fmt.Sprintf("MILESTONE: %s", m.name),
					Line:     m.startLine,
					Message: fmt.Sprintf(
						"MILESTONE '%s' lists BEHAVIOR '%s' under Deferred BEHAVIORs but no such BEHAVIOR exists in the spec.",
						m.name, name),
				})
			}
		}
	}
	return diags
}

// ---------------------------------------------------------------------------
// RULE-17: Scaffold milestone ordering and uniqueness
// ---------------------------------------------------------------------------

func applyRule17(sf *specFile) []Diagnostic {
	var diags []Diagnostic

	if len(sf.milestones) == 0 {
		return diags
	}

	var scaffoldMilestones []milestone
	for _, m := range sf.milestones {
		if m.scaffold == "true" {
			scaffoldMilestones = append(scaffoldMilestones, m)
		}
	}

	if len(scaffoldMilestones) > 1 {
		diags = append(diags, Diagnostic{
			Severity: SeverityError,
			Section:  "structure",
			Line:     1,
			Message: "More than one MILESTONE has Scaffold: true. " +
				"At most one scaffold milestone is permitted per spec.",
		})
	}

	if len(scaffoldMilestones) == 1 {
		sm := scaffoldMilestones[0]
		first := sf.milestones[0]
		if sm.name != first.name {
			diags = append(diags, Diagnostic{
				Severity: SeverityError,
				Section:  fmt.Sprintf("MILESTONE: %s", sm.name),
				Line:     sm.startLine,
				Message: fmt.Sprintf(
					"Scaffold milestone '%s' must appear first in the spec "+
						"(lowest version number / earliest in document order). "+
						"Later milestones depend on the scaffold foundation.", sm.name),
			})
		}
	}

	return diags
}

// ---------------------------------------------------------------------------
// RULE-18: Spec hash presence in TRANSLATION_REPORT
// ---------------------------------------------------------------------------

func applyRule18(sf *specFile) []Diagnostic {
	var diags []Diagnostic

	// Look for TRANSLATION_REPORT.md adjacent to the spec
	dir := filepath.Dir(sf.path)
	reportPath := filepath.Join(dir, "TRANSLATION_REPORT.md")

	// Also check code/ subdirectory
	altReportPath := filepath.Join(dir, "code", "TRANSLATION_REPORT.md")

	var reportData []byte
	var err error
	reportData, err = os.ReadFile(reportPath)
	if err != nil {
		reportData, err = os.ReadFile(altReportPath)
		if err != nil {
			return diags // No report found; nothing to check
		}
	}

	reportContent := string(reportData)

	// Check for Spec-SHA256: field
	hasHash := false
	lines := splitLines(reportContent)
	for _, l := range lines {
		if strings.HasPrefix(l, "Spec-SHA256:") {
			val := strings.TrimSpace(strings.TrimPrefix(l, "Spec-SHA256:"))
			// Check it's a 64-char hex string
			if len(val) == 64 && isHex(val) {
				hasHash = true
				break
			}
		}
	}

	if !hasHash {
		diags = append(diags, Diagnostic{
			Severity: SeverityWarning,
			Section:  "report",
			Line:     1,
			Message: "TRANSLATION_REPORT.md is missing Spec-SHA256: field. " +
				"Every translation run must record the SHA256 of the spec it was produced from. " +
				"Run: sha256sum <specname>.md",
		})
	}

	return diags
}

// ---------------------------------------------------------------------------
// RULE-19: Includes path resolves
// ---------------------------------------------------------------------------

func applyRule19(sf *specFile) []Diagnostic {
	var diags []Diagnostic
	dir := filepath.Dir(sf.path)
	for _, inc := range sf.includes {
		resolved := filepath.Join(dir, inc.value)
		if _, err := os.Stat(resolved); err != nil {
			diags = append(diags, Diagnostic{
				Severity: SeverityError,
				Section:  "META",
				Line:     inc.line,
				Message:  fmt.Sprintf("Includes path does not resolve: %s", inc.value),
			})
		}
	}
	return diags
}

// ---------------------------------------------------------------------------
// RULE-20: Merged spec has no name collisions
// ---------------------------------------------------------------------------

func applyRule20(sf *specFile) []Diagnostic {
	var diags []Diagnostic
	// Parse included specs and check for collisions
	dir := filepath.Dir(sf.path)
	for _, inc := range sf.includes {
		includedPath := filepath.Join(dir, inc.value)
		data, err := os.ReadFile(includedPath)
		if err != nil {
			continue // RULE-19 already reported
		}
		includedLines := splitLines(string(data))
		includedSF := parseSpecFile(includedPath, includedLines)

		// Check TYPE collisions
		hostTypes := collectTypeNames(sf)
		includedTypes := collectTypeNames(includedSF)
		hostTypeSet := map[string]bool{}
		for _, t := range hostTypes {
			hostTypeSet[t] = true
		}
		for _, t := range includedTypes {
			if hostTypeSet[t] {
				diags = append(diags, Diagnostic{
					Severity: SeverityError,
					Section:  "META",
					Line:     inc.line,
					Message: fmt.Sprintf(
						"Name collision after merge: TYPE %s appears in both %s and %s",
						t, filepath.Base(includedPath), filepath.Base(sf.path)),
				})
			}
		}

		// Check EXAMPLE collisions
		hostExamples := map[string]bool{}
		for _, ex := range sf.exampleBlocks {
			hostExamples[ex.name] = true
		}
		for _, ex := range includedSF.exampleBlocks {
			if hostExamples[ex.name] {
				diags = append(diags, Diagnostic{
					Severity: SeverityError,
					Section:  "META",
					Line:     inc.line,
					Message: fmt.Sprintf(
						"Name collision after merge: EXAMPLE %s appears in both %s and %s",
						ex.name, filepath.Base(includedPath), filepath.Base(sf.path)),
				})
			}
		}

		// Check BEHAVIOR collisions
		hostBehaviors := map[string]bool{}
		for _, b := range sf.behaviorNames {
			hostBehaviors[b] = true
		}
		for _, b := range includedSF.behaviorNames {
			if hostBehaviors[b] {
				diags = append(diags, Diagnostic{
					Severity: SeverityError,
					Section:  "META",
					Line:     inc.line,
					Message: fmt.Sprintf(
						"Name collision after merge: BEHAVIOR %s appears in both %s and %s",
						b, filepath.Base(includedPath), filepath.Base(sf.path)),
				})
			}
		}
	}
	return diags
}

// ---------------------------------------------------------------------------
// RULE-21: Inclusion graph is acyclic and well-formed
// ---------------------------------------------------------------------------

func applyRule21(sf *specFile) []Diagnostic {
	var diags []Diagnostic

	// DFS to detect cycles
	visited := map[string]bool{}
	inStack := map[string]bool{}
	var stack []string

	var dfs func(path string) bool
	dfs = func(currentPath string) bool {
		abs, err := filepath.Abs(currentPath)
		if err != nil {
			return false
		}
		if inStack[abs] {
			// Cycle detected
			cycleStart := -1
			for i, p := range stack {
				if p == abs {
					cycleStart = i
					break
				}
			}
			cyclePath := append(stack[cycleStart:], abs)
			bases := make([]string, len(cyclePath))
			for i, p := range cyclePath {
				bases[i] = filepath.Base(p)
			}
			diags = append(diags, Diagnostic{
				Severity: SeverityError,
				Section:  "META",
				Line:     1,
				Message:  fmt.Sprintf("Inclusion cycle: %s", strings.Join(bases, " → ")),
			})
			return true
		}
		if visited[abs] {
			return false
		}
		visited[abs] = true
		inStack[abs] = true
		stack = append(stack, abs)

		// Read and check this file's includes
		data, err := os.ReadFile(currentPath)
		if err != nil {
			stack = stack[:len(stack)-1]
			inStack[abs] = false
			return false
		}
		lines := splitLines(string(data))
		childSF := parseSpecFile(currentPath, lines)

		// Check for MILESTONE and DEPLOYMENT sections (only for included files, not the host)
		if currentPath != sf.path {
			for _, line := range lines {
				if fenceDepthAt(lines, indexOf(lines, line)) > 0 {
					continue
				}
				if strings.HasPrefix(line, "## MILESTONE:") {
					diags = append(diags, Diagnostic{
						Severity: SeverityError,
						Section:  "structure",
						Line:     1,
						Message:  fmt.Sprintf("Included spec must not declare MILESTONE section: %s", filepath.Base(currentPath)),
					})
					break
				}
			}
			for i, line := range lines {
				if fenceDepthAt(lines, i) > 0 {
					continue
				}
				if line == "## DEPLOYMENT" || strings.HasPrefix(line, "## DEPLOYMENT ") {
					diags = append(diags, Diagnostic{
						Severity: SeverityError,
						Section:  "structure",
						Line:     1,
						Message:  fmt.Sprintf("Included spec must not declare DEPLOYMENT section: %s", filepath.Base(currentPath)),
					})
					break
				}
			}
		}

		dir := filepath.Dir(currentPath)
		for _, inc := range childSF.includes {
			childPath := filepath.Join(dir, inc.value)
			dfs(childPath)
		}

		stack = stack[:len(stack)-1]
		inStack[abs] = false
		return false
	}

	dfs(sf.path)
	return diags
}

// indexOf returns the index of s in lines, or 0 if not found.
func indexOf(lines []string, s string) int {
	for i, l := range lines {
		if l == s {
			return i
		}
	}
	return 0
}

// ---------------------------------------------------------------------------
// Helper types and functions
// ---------------------------------------------------------------------------

type behaviorSection struct {
	name          string
	sectionHeader string
	startLine     int
	lines         []string
}

// extractBehaviorSections extracts the content of each BEHAVIOR/BEHAVIOR/INTERNAL section.
func extractBehaviorSections(sf *specFile) []behaviorSection {
	var result []behaviorSection
	var current *behaviorSection
	fenceDepth := 0

	for i, rawLine := range sf.lines {
		lineNum := i + 1
		trimmed := strings.TrimSpace(rawLine)

		if strings.HasPrefix(trimmed, "```") || strings.HasPrefix(trimmed, "~~~") {
			if fenceDepth == 0 {
				fenceDepth = 1
			} else {
				fenceDepth--
			}
			if current != nil {
				current.lines = append(current.lines, rawLine)
			}
			continue
		}

		if fenceDepth > 0 {
			if current != nil {
				current.lines = append(current.lines, rawLine)
			}
			continue
		}

		// Check for BEHAVIOR or BEHAVIOR/INTERNAL section at column 0
		if strings.HasPrefix(rawLine, "## BEHAVIOR: ") || strings.HasPrefix(rawLine, "## BEHAVIOR/INTERNAL: ") {
			if current != nil {
				result = append(result, *current)
			}
			var name, header string
			if strings.HasPrefix(rawLine, "## BEHAVIOR/INTERNAL: ") {
				name = strings.TrimSpace(strings.TrimPrefix(rawLine, "## BEHAVIOR/INTERNAL: "))
				header = fmt.Sprintf("BEHAVIOR/INTERNAL: %s", name)
			} else {
				name = strings.TrimSpace(strings.TrimPrefix(rawLine, "## BEHAVIOR: "))
				header = fmt.Sprintf("BEHAVIOR: %s", name)
			}
			current = &behaviorSection{
				name:          name,
				sectionHeader: header,
				startLine:     lineNum,
			}
			continue
		}

		// Any other ## heading closes the current behavior section
		if strings.HasPrefix(rawLine, "## ") && current != nil {
			result = append(result, *current)
			current = nil
			continue
		}

		if current != nil {
			current.lines = append(current.lines, rawLine)
		}
	}
	if current != nil {
		result = append(result, *current)
	}
	return result
}

// parseExampleBlocks parses the GIVEN/WHEN/THEN structure for each example block.
func parseExampleBlocks(lines []string, blocks []exampleBlock) []exampleBlock {
	if len(blocks) == 0 {
		return blocks
	}

	// For each block, find its content range and parse
	for bi := range blocks {
		startIdx := blocks[bi].startLine - 1 // 0-based index of "### EXAMPLE: name" line
		var endIdx int
		if bi+1 < len(blocks) {
			endIdx = blocks[bi+1].startLine - 1
		} else {
			endIdx = len(lines)
		}

		// Parse within startIdx..endIdx
		blockLines := lines[startIdx:endIdx]
		parseExampleBlock(&blocks[bi], blockLines)
	}
	return blocks
}

func parseExampleBlock(ex *exampleBlock, lines []string) {
	type state int
	const (
		stateStart state = iota
		stateGiven
		stateWhen
		stateThen
	)

	cur := stateStart
	givenContent := false
	whenContent := false
	thenContent := false
	lastWhenIdx := -1
	lastWhenHasThen := false
	whenCount := 0
	thenCount := 0

	fenceDepth := 0

	for i, rawLine := range lines {
		trimmed := strings.TrimSpace(rawLine)

		// Fence tracking
		if strings.HasPrefix(trimmed, "```") || strings.HasPrefix(trimmed, "~~~") {
			if fenceDepth == 0 {
				fenceDepth = 1
			} else {
				fenceDepth--
			}
			// Fence markers themselves count as content within the block
			switch cur {
			case stateGiven:
				givenContent = true
			case stateWhen:
				whenContent = true
			case stateThen:
				thenContent = true
			}
			continue
		}
		if fenceDepth > 0 {
			// Lines inside fences count as content
			if trimmed != "" {
				switch cur {
				case stateGiven:
					givenContent = true
				case stateWhen:
					whenContent = true
				case stateThen:
					thenContent = true
				}
			}
			continue
		}

		if i == 0 {
			// The "### EXAMPLE: name" line itself
			continue
		}

		if strings.HasPrefix(rawLine, "GIVEN:") {
			cur = stateGiven
			// Content on the GIVEN: line itself counts
			inlineGiven := strings.TrimSpace(strings.TrimPrefix(rawLine, "GIVEN:"))
			givenContent = inlineGiven != ""
			continue
		}
		if strings.HasPrefix(rawLine, "WHEN:") {
			if cur == stateWhen && !lastWhenHasThen {
				ex.whenWithoutThen = true
			}
			cur = stateWhen
			// Content on the WHEN: line itself counts as block content
			inlineContent := strings.TrimSpace(strings.TrimPrefix(rawLine, "WHEN:"))
			whenContent = inlineContent != ""
			lastWhenIdx = i
			lastWhenHasThen = false
			whenCount++
			ex.hasWhen = true
			continue
		}
		if strings.HasPrefix(rawLine, "THEN:") {
			cur = stateThen
			// Content on the THEN: line itself counts
			inlineThen := strings.TrimSpace(strings.TrimPrefix(rawLine, "THEN:"))
			thenContent = inlineThen != ""
			lastWhenHasThen = true
			thenCount++
			ex.hasThen = true
			continue
		}

		switch cur {
		case stateGiven:
			if trimmed != "" {
				givenContent = true
			}
		case stateWhen:
			if trimmed != "" {
				whenContent = true
			}
		case stateThen:
			if trimmed != "" {
				thenContent = true
			}
		}
	}

	// Final WHEN without THEN
	if cur == stateWhen && lastWhenIdx >= 0 && !lastWhenHasThen {
		ex.whenWithoutThen = true
	}

	// Check GIVEN exists
	for _, rawLine := range lines[1:] {
		if strings.HasPrefix(rawLine, "GIVEN:") {
			ex.hasGiven = true
			break
		}
	}

	// Set empty flags
	if ex.hasGiven && !givenContent {
		ex.givenEmpty = true
	}
	if ex.hasWhen && !whenContent {
		ex.whenEmpty = true
	}
	if ex.hasThen && !thenContent {
		ex.thenEmpty = true
	}

	_ = whenCount
	_ = thenCount
}

// findNegativeExample checks if there's a negative-path example for the given behavior.
func findNegativeExample(sf *specFile, behaviorName string) bool {
	negativePatterns := []string{
		"Err(", "error", "exit_code = 1", "exit_code = 2",
		"stderr contains", "exit 1", "exit 2",
	}

	// For specs with a single BEHAVIOR, all EXAMPLES reference it.
	// For multi-BEHAVIOR specs, association is by name matching.
	singleBehavior := len(sf.behaviorNames) == 1

	inExamples := false
	inExample := false
	inThen := false
	currentExampleName := ""
	referencesThisBehavior := false
	fenceDepth := 0

	for i, rawLine := range sf.lines {
		trimmed := strings.TrimSpace(rawLine)
		if strings.HasPrefix(trimmed, "```") || strings.HasPrefix(trimmed, "~~~") {
			if fenceDepth == 0 {
				fenceDepth = 1
			} else {
				fenceDepth--
			}
			continue
		}
		if fenceDepth > 0 {
			continue
		}

		_ = i

		if rawLine == "## EXAMPLES" {
			inExamples = true
			continue
		}
		if strings.HasPrefix(rawLine, "## ") && inExamples {
			inExamples = false
		}
		if !inExamples {
			continue
		}

		if strings.HasPrefix(rawLine, "### EXAMPLE: ") {
			inExample = true
			inThen = false
			currentExampleName = strings.TrimSpace(strings.TrimPrefix(rawLine, "### EXAMPLE: "))
			referencesThisBehavior = singleBehavior ||
				strings.Contains(strings.ToLower(currentExampleName), strings.ToLower(behaviorName))
			continue
		}

		if strings.HasPrefix(rawLine, "WHEN:") {
			if !referencesThisBehavior {
				// Check if WHEN line references the behavior
				whenContent := strings.TrimSpace(strings.TrimPrefix(rawLine, "WHEN:"))
				if strings.Contains(strings.ToLower(whenContent), strings.ToLower(behaviorName)) {
					referencesThisBehavior = true
				}
			}
			inThen = false
		}

		if rawLine == "THEN:" || strings.HasPrefix(rawLine, "THEN:") {
			inThen = true
		}

		if inExample && inThen && referencesThisBehavior {
			for _, pat := range negativePatterns {
				if strings.Contains(rawLine, pat) {
					return true
				}
			}
		}
	}
	return false
}

// collectTypeNames collects all type names from the TYPES section.
func collectTypeNames(sf *specFile) []string {
	var names []string
	inTypes := false
	fenceDepth := 0
	for i, rawLine := range sf.lines {
		trimmed := strings.TrimSpace(rawLine)
		if strings.HasPrefix(trimmed, "```") || strings.HasPrefix(trimmed, "~~~") {
			if fenceDepth == 0 {
				fenceDepth = 1
			} else {
				fenceDepth--
			}
			continue
		}
		if fenceDepth > 0 {
			continue
		}

		_ = i

		if rawLine == "## TYPES" {
			inTypes = true
			continue
		}
		if strings.HasPrefix(rawLine, "## ") && inTypes {
			break
		}
		if !inTypes {
			continue
		}

		// Lines matching "^<TypeName> :="
		if idx := strings.Index(rawLine, " :="); idx > 0 {
			name := rawLine[:idx]
			// Type names are typically PascalCase or simple identifiers
			if !strings.Contains(name, " ") && name != "" {
				names = append(names, name)
			}
		}
	}
	return names
}

// extractSectionContent returns the text content of a section (between its header and next ## heading).
func (sf *specFile) extractSectionContent(headerPrefix string) string {
	var sb strings.Builder
	inSection := false
	for _, rawLine := range sf.lines {
		if strings.HasPrefix(rawLine, headerPrefix) {
			inSection = true
			continue
		}
		if strings.HasPrefix(rawLine, "## ") && inSection {
			break
		}
		if inSection {
			sb.WriteString(rawLine)
			sb.WriteByte('\n')
		}
	}
	return sb.String()
}

// isSemanticVersion checks if a string matches "^[0-9]+\.[0-9]+\.[0-9]+$".
func isSemanticVersion(v string) bool {
	parts := strings.Split(v, ".")
	if len(parts) != 3 {
		return false
	}
	for _, p := range parts {
		if p == "" {
			return false
		}
		for _, c := range p {
			if c < '0' || c > '9' {
				return false
			}
		}
	}
	return true
}

// isHex checks if a string is all lowercase hex characters.
func isHex(s string) bool {
	for _, c := range s {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
			return false
		}
	}
	return true
}

// FormatDiagnostic formats a diagnostic line per the spec's format:
// {SEVERITY}  {file}:{line}  [{section}]  {message}
func FormatDiagnostic(file string, d Diagnostic) string {
	return fmt.Sprintf("%-7s  %s:%d  [%s]  %s",
		string(d.Severity), filepath.Base(file), d.Line, d.Section, d.Message)
}

// FormatSummary formats the summary line per the spec's format.
func FormatSummary(file string, result LintResult, opts Options) string {
	base := filepath.Base(file)
	errorCount := 0
	warningCount := 0
	for _, d := range result.Diagnostics {
		if d.Severity == SeverityError {
			errorCount++
		} else if d.Severity == SeverityWarning {
			warningCount++
		}
	}

	if result.ExitCode == 0 {
		if warningCount == 0 {
			return fmt.Sprintf("✓ %s: valid", base)
		}
		return fmt.Sprintf("✓ %s: valid (%d warning(s))", base, warningCount)
	}
	// exit code 1
	if opts.Strict {
		return fmt.Sprintf("✗ %s: %d error(s), %d warning(s) [strict mode]", base, errorCount, warningCount)
	}
	return fmt.Sprintf("✗ %s: %d error(s), %d warning(s)", base, errorCount, warningCount)
}
