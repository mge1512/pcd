// generated from spec: pcd-lint.md sha256:293541ab62274835c61de50947f6283748831c4681cf3f02c4be2f8e942d28a9

// pcd-lint validates specification files against the structural rules
// defined in the pcd-lint specification.
//
// Usage:
//
//	pcd-lint <specfile.md>
//	pcd-lint strict=true <specfile.md>
//	pcd-lint check-report=true <specfile.md>
//	pcd-lint strict=true check-report=true <specfile.md>
//	pcd-lint list-templates
//	pcd-lint version
package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/mge1512/pcd/tools/pcd-lint/internal/lint"
	"github.com/mge1512/pcd/tools/pcd-lint/internal/spdx"
)

const (
	// Version is the pcd-lint tool version.
	Version = "0.4.0"
	// SpecSchema is the spec schema version this implementation targets.
	SpecSchema = "0.4.0"
	// SpecHash is the SHA256 of the spec file this was generated from.
	SpecHash = "293541ab62274835c61de50947f6283748831c4681cf3f02c4be2f8e942d28a9"
)

func main() {
	args := os.Args[1:]

	if len(args) == 0 {
		printUsage()
		os.Exit(2)
	}

	// Handle bare-word commands
	if args[0] == "list-templates" {
		lint.ListTemplates()
		os.Exit(0)
	}

	if args[0] == "version" {
		fmt.Printf("pcd-lint %s (schema %s) spdx/%s spec:%s\n",
			Version, SpecSchema, spdx.Version, SpecHash)
		os.Exit(0)
	}

	// Parse key=value options
	opts := lint.Options{}
	var fileArg string

	for _, arg := range args {
		if strings.Contains(arg, "=") {
			parts := strings.SplitN(arg, "=", 2)
			key := parts[0]
			value := parts[1]
			switch key {
			case "strict":
				if value == "true" {
					opts.Strict = true
				} else if value == "false" {
					opts.Strict = false
				} else {
					fmt.Fprintf(os.Stderr, "error: unrecognised option: %s\n", key)
					os.Exit(2)
				}
			case "check-report":
				if value == "true" {
					opts.CheckReport = true
				} else if value == "false" {
					opts.CheckReport = false
				} else {
					fmt.Fprintf(os.Stderr, "error: unrecognised option: %s\n", key)
					os.Exit(2)
				}
			default:
				fmt.Fprintf(os.Stderr, "error: unrecognised option: %s\n", key)
				os.Exit(2)
			}
		} else {
			// Bare word — should be the file argument
			fileArg = arg
		}
	}

	if fileArg == "" {
		printUsage()
		os.Exit(2)
	}

	// Run lint
	result := lint.Lint(fileArg, opts)

	// If exit code is 2 (invocation error), lint already wrote to stderr and we exit.
	if result.ExitCode == 2 {
		os.Exit(2)
	}

	// Write diagnostics to stderr (STEP 5)
	for _, d := range result.Diagnostics {
		fmt.Fprintln(os.Stderr, lint.FormatDiagnostic(fileArg, d))
	}

	// Write summary to stdout (STEP 7)
	fmt.Fprintln(os.Stdout, lint.FormatSummary(fileArg, result, opts))

	os.Exit(result.ExitCode)
}

func printUsage() {
	fmt.Fprintf(os.Stderr, "Usage: pcd-lint [strict=true] [check-report=true] <specfile.md>\n")
	fmt.Fprintf(os.Stderr, "       pcd-lint list-templates\n")
	fmt.Fprintf(os.Stderr, "       pcd-lint version\n")
}
