# prompts/ - derived copies

The files in this directory are copies, written by `make skills` from the
procedure files of the two skills. The canonical files are:

| Copy | Canonical file |
|---|---|
| `interview-prompt.md` | `skills/pcd-spec/modes/interview-procedure.md` |
| `reverse-prompt.md` | `skills/pcd-spec/modes/reverse-procedure.md` |
| `change-impact.md` | `skills/pcd-spec/modes/change-impact-procedure.md` |
| `prompt.md` | `skills/pcd-translate/modes/translation-procedure.md` |
| `reviewer.md` | `skills/pcd-translate/modes/review-procedure.md` |
| `security-reviewer.md` | `skills/pcd-translate/modes/security-review-procedure.md` |
| `tiebreaker.md` | `skills/pcd-translate/modes/tiebreak-procedure.md` |

The copies exist so that the deployment templates, the asset build of
`mcp-server-pcd` and the bundled translation inputs under `tools/*/spec/`
keep their paths and their hashes. Edit the canonical file, run
`make skills`, and commit both; `make check-skills` fails when a copy is
stale. The mapping is `MAP` in this directory. See `CONTRIBUTING.md`,
"Derived Skill Content".
