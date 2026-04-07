// pcd-lint — Post-Coding Development specification linter
// License: GPL-2.0-only
// See https://spdx.org/licenses/GPL-2.0-only.html

package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/pcd-tools/pcd-lint/internal/lint"
)

func main() {
	args := os.Args[1:]

	if len(args) == 0 {
		usage()
		os.Exit(2)
	}

	// Handle bare-word commands
	if args[0] == "list-templates" {
		lint.CmdListTemplates()
		return
	}
	if args[0] == "version" {
		lint.CmdVersion()
		return
	}

	// Parse key=value options
	strict := false
	var fileArg string

	for _, a := range args {
		if strings.Contains(a, "=") {
			parts := strings.SplitN(a, "=", 2)
			key := parts[0]
			val := parts[1]
			switch key {
			case "strict":
				if val == "true" {
					strict = true
				} else if val == "false" {
					strict = false
				} else {
					fmt.Fprintf(os.Stderr, "error: unrecognised option: %s\n", key)
					os.Exit(2)
				}
			default:
				fmt.Fprintf(os.Stderr, "error: unrecognised option: %s\n", key)
				os.Exit(2)
			}
		} else {
			if fileArg != "" {
				fmt.Fprintln(os.Stderr, "error: too many file arguments")
				os.Exit(2)
			}
			fileArg = a
		}
	}

	if fileArg == "" {
		usage()
		os.Exit(2)
	}

	// Check .md extension
	if !strings.HasSuffix(fileArg, ".md") {
		fmt.Fprintf(os.Stderr, "error: file must have .md extension: %s\n", fileArg)
		os.Exit(2)
	}

	// Check file exists and is readable
	if _, err := os.Stat(fileArg); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "error: cannot open file: %s\n", fileArg)
		os.Exit(2)
	}

	result := lint.LintSpec(fileArg, strict)

	// Write diagnostics to stderr
	for _, d := range result.Diagnostics {
		fmt.Fprintln(os.Stderr, lint.FormatDiagnostic(d, fileArg))
	}

	// Write summary to stdout
	fmt.Println(lint.FormatSummary(result, strict))

	os.Exit(result.ExitCode)
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage: pcd-lint [strict=true] <specfile.md>")
	fmt.Fprintln(os.Stderr, "       pcd-lint list-templates")
	fmt.Fprintln(os.Stderr, "       pcd-lint version")
}
