// Package independent_tests contains all independent tests for mcp-server-pcd.
//
// All tests use FakeStore and FakeFilesystem — no filesystem access,
// no network calls, no live pcd-lint binary required.
//
// Test doubles used:
//   - store.FakeStore (configurable: Templates, Hints, Prompts)
//   - milestone.FakeFilesystem (configurable: Files, ReadErr, WriteErr)
//
// Test files in this package:
//   - INDEPENDENT_TESTS_test.go — all test functions
//
// Covered behaviors:
//   - list_templates (EXAMPLE: list_templates_returns_names)
//   - get_template (EXAMPLE: get_template_cli_tool, get_template_unknown)
//   - read_resource (EXAMPLE: read_resource_invalid_uri, read_resource_* URI parsing)
//   - lint_content (EXAMPLE: lint_content_valid_spec, lint_content_missing_invariants,
//     lint_content_milestone_scaffold_not_first, lint_content_two_scaffold_milestones,
//     lint_content_bad_extension, lint_content_matches_cli)
//   - set_milestone_status (EXAMPLE: set_milestone_active, set_milestone_active_conflict,
//     set_milestone_released)
//   - transport argument parsing (stdio, http, listen=)
//
// SPDX-License-Identifier: GPL-2.0-only
package independent_tests
