# TEST_REPORT.md

## Spec-SHA256: 293541ab62274835c61de50947f6283748831c4681cf3f02c4be2f8e942d28a9
Spec-SHA256 (host): 293541ab62274835c61de50947f6283748831c4681cf3f02c4be2f8e942d28a9
Included-Specs: |
  | Path | SHA256 |
  |------|--------|

## LLM-Name: mistral-large-2512
## Mode: test-author
## Deployment-Template: cli-tool.template.md v0.3.26
## Preset-Resolution: none (system defaults only)
## Hints-Files-Read: none
## Test-Compile-Gate: pass

## Target Language: Go

## Tests Produced
| Test Function | Covers EXAMPLE/BEHAVIOR/INVARIANT |
|---------------|------------------------------------|
| TestValidMinimalSpec | EXAMPLE: valid_minimal_spec |
| TestMultipleAuthorsValid | EXAMPLE: multiple_authors_valid |
| TestInvalidSpdxLicense | EXAMPLE: invalid_spdx_license |
| TestInvalidVersionFormat | EXAMPLE: invalid_version_format |
| TestMissingAuthor | EXAMPLE: missing_author |
| TestMissingSection | EXAMPLE: missing_section |
| TestUnknownDeploymentTemplate | EXAMPLE: unknown_deployment_template |
| TestDeprecatedTargetFieldPermissive | EXAMPLE: deprecated_target_field_permissive |
| TestDeprecatedTargetFieldStrict | EXAMPLE: deprecated_target_field_strict |
| TestEnhanceExistingMissingLanguage | EXAMPLE: enhance_existing_missing_language |
| TestEmptyGivenBlockPermissive | EXAMPLE: empty_given_block_permissive |
| TestMultipleErrors | EXAMPLE: multiple_errors |
| TestFileNotFound | EXAMPLE: file_not_found |
| TestUnrecognisedOption | EXAMPLE: unrecognised_option |
| TestBehaviorInternalRecognised | EXAMPLE: behavior_internal_recognised |
| TestListTemplates | BEHAVIOR: list-templates |
| TestNonMdExtension | EXAMPLE: non_md_extension |
| TestMultiPassExampleValid | EXAMPLE: multi_pass_example_valid |

## Specification Ambiguities
None encountered.

## Test Methodology
- All tests invoke `pcd-lint` as a black-box CLI tool using `exec.Command`.
- Fixtures are complete spec files (`.md`) covering all required sections.
- Assertions are made on stdout, stderr, and exit code.
- No internal Go functions or mocks are used; the test harness is purely black-box.