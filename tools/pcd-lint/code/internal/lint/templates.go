// generated from spec: pcd-lint.md sha256:293541ab62274835c61de50947f6283748831c4681cf3f02c4be2f8e942d28a9

package lint

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// knownTemplates is the ordered list of all known DeploymentTemplate values.
// Order follows the spec's DeploymentTemplate type definition.
var knownTemplates = []string{
	"wasm",
	"ebpf",
	"kernel-module",
	"verified-library",
	"cli-tool",
	"gui-tool",
	"cloud-native",
	"backend-service",
	"library-c-abi",
	"enterprise-software",
	"academic",
	"python-tool",
	"enhance-existing",
	"manual",
	"template",
	"mcp-server",
	"project-manifest",
}

// specialAnnotations defines fixed annotations for special template values.
var specialAnnotations = map[string]string{
	"enhance-existing": "(declare Language: in META)",
	"manual":           "(declare Target: in META)",
	"template":         "(template definition file, not translatable)",
	"project-manifest": "(architect artifact, no code generated)",
}

// templateSearchDirs returns the list of existing template directories,
// in ascending precedence order (last-wins).
func templateSearchDirs() []string {
	home, _ := os.UserHomeDir()
	candidates := []string{
		"/usr/share/pcd/templates",
		"/etc/pcd/templates",
		filepath.Join(home, ".config", "pcd", "templates"),
		"./.pcd/templates",
	}
	var result []string
	for _, dir := range candidates {
		if info, err := os.Stat(dir); err == nil && info.IsDir() {
			result = append(result, dir)
		}
	}
	return result
}

// findTemplateFile returns the path to the companion template file for the given name.
// Returns "" if not found in any search directory.
func findTemplateFile(name string) string {
	dirs := templateSearchDirs()
	filename := name + ".template.md"
	last := ""
	for _, dir := range dirs {
		candidate := filepath.Join(dir, filename)
		if _, err := os.Stat(candidate); err == nil {
			last = candidate
		}
	}
	return last
}

// readDefaultLanguage reads the default language from a template file's TEMPLATE-TABLE section.
// Returns "" if not found or the file cannot be read.
func readDefaultLanguage(path string) string {
	f, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer f.Close()

	inTable := false
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()

		if strings.Contains(line, "## TEMPLATE-TABLE") {
			inTable = true
			continue
		}
		if strings.HasPrefix(line, "## ") && inTable {
			break
		}
		if !inTable {
			continue
		}

		// Look for table rows: | LANGUAGE | <value> | default | ...
		if strings.Contains(line, "| LANGUAGE |") || strings.Contains(line, "| LANGUAGE|") {
			// Parse: | Key | Value | Constraint | Notes |
			parts := strings.Split(line, "|")
			if len(parts) >= 4 {
				key := strings.TrimSpace(parts[1])
				value := strings.TrimSpace(parts[2])
				constraint := strings.TrimSpace(parts[3])
				if key == "LANGUAGE" && constraint == "default" && value != "" {
					return value
				}
			}
		}
	}
	return ""
}

// ListTemplates prints all known deployment templates with their resolved default
// target language annotations, then exits 0.
func ListTemplates() {
	for _, t := range knownTemplates {
		var annotation string

		// Special values use fixed annotations
		if fixed, ok := specialAnnotations[t]; ok {
			annotation = fixed
		} else {
			// Try to find companion template file
			templatePath := findTemplateFile(t)
			if templatePath == "" {
				annotation = "(template file not found)"
			} else {
				lang := readDefaultLanguage(templatePath)
				if lang == "" {
					annotation = "(installed)"
				} else {
					annotation = lang
				}
			}
		}

		fmt.Printf("%s  →  %s\n", t, annotation)
	}
}
