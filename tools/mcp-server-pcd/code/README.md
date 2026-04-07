# mcp-server-pcd

PCD Model Context Protocol server — templates, hints, lint, and milestone tools.

## Overview

`mcp-server-pcd` is an MCP server for the [Post-Coding Development (PCD)](https://github.com/mge1512/pcd) workflow. It provides MCP tools and resources for:

- Listing and retrieving PCD deployment templates
- Listing and reading hints files and prompts
- Linting PCD specification files (RULE-01 through RULE-17, identical to `pcd-lint`)
- Setting milestone status in spec files

The binary is **self-contained**: all templates, hints, and prompts are embedded at build time. No runtime dependencies are required.

## Installation

### openSUSE / SUSE Linux Enterprise

```sh
zypper install mcp-server-pcd
```

### Fedora / RHEL

```sh
dnf install mcp-server-pcd
```

### Debian / Ubuntu

```sh
apt install mcp-server-pcd
```

### From source

```sh
make embed-assets
CGO_ENABLED=0 go build -o mcp-server-pcd .
```

> **Note:** `make embed-assets` must run before `go build`. It stages templates,
> hints, and prompts into `internal/store/assets/` for embedding.

## Invocation

### stdio transport (for mcphost, Claude Desktop, VS Code)

```sh
mcp-server-pcd stdio
```

### HTTP transport (for web-based hosts, remote access)

```sh
mcp-server-pcd http
mcp-server-pcd http listen=0.0.0.0:9000
```

Default listen address: `127.0.0.1:8080`

## MCP Host Configuration

### mcphost (stdio)

```yaml
mcpServers:
  pcd:
    command: mcp-server-pcd
    args: [stdio]
```

### mcphost (http, running as service)

```yaml
mcpServers:
  pcd:
    url: http://127.0.0.1:8080/mcp
```

## MCP Tools

| Tool | Description |
|------|-------------|
| `list_templates` | List all installed PCD deployment templates (name, version, language; no content) |
| `get_template` | Get a template by name; returns full Markdown content |
| `list_resources` | List all resources (templates, hints, prompts) as `pcd://` URIs |
| `read_resource` | Read a resource by URI (`pcd://templates/…`, `pcd://hints/…`, `pcd://prompts/…`) |
| `lint_content` | Validate a PCD spec string; applies RULE-01 through RULE-17 |
| `lint_file` | Validate a PCD spec file on disk |
| `get_schema_version` | Return the Spec-Schema version this binary was built against |
| `set_milestone_status` | Set a MILESTONE `Status:` field in a spec file |

### Tool: lint_content

```json
{
  "content": "# my-spec\n\n## META\n...",
  "filename": "my-spec.md"
}
```

Returns:

```json
{
  "valid": true,
  "errors": 0,
  "warnings": 0,
  "diagnostics": []
}
```

### Tool: set_milestone_status

```json
{
  "spec_path": "/home/user/project/my-spec.md",
  "milestone_name": "0.1.0",
  "new_status": "active"
}
```

Valid status values: `pending`, `active`, `failed`, `released`.
Only one milestone may be `active` at a time.

## Asset Overlay Search Path

The binary embeds all assets at build time. At startup it additionally checks
these directories for overlay assets (last-wins, ascending precedence):

| Precedence | Directory | Provider |
|------------|-----------|----------|
| 1 (lowest) | `/usr/share/pcd/{templates,hints,prompts}/` | pcd-templates |
| 2 | `/etc/pcd/{templates,hints,prompts}/` | system admin |
| 3 | `~/.config/pcd/{templates,hints,prompts}/` | user |
| 4 (highest) | `./.pcd/{templates,hints,prompts}/` | project-local |

Missing directories are silently skipped.

Install `pcd-templates` to enable the system-level overlay:

```sh
zypper install pcd-templates   # openSUSE/SLES
dnf install pcd-templates      # Fedora/RHEL
apt install pcd-templates      # Debian/Ubuntu
```

## Running as a systemd Service

```sh
systemctl enable --now mcp-server-pcd
```

The service unit binds to `127.0.0.1:8080` by default. Edit
`/usr/lib/systemd/system/mcp-server-pcd.service` to change the address.

## Build System

```makefile
make embed-assets    # stage assets into internal/store/assets/
make build           # build binary (runs embed-assets first)
make test            # run independent tests
make clean           # remove staged assets and binary
make man             # generate man page (requires pandoc)
make vendor-tarball  # create vendor tarball for OBS packaging
```

## License

GPL-2.0-only — https://spdx.org/licenses/GPL-2.0-only.html
