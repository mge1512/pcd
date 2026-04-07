% MCP-SERVER-PCD(1) mcp-server-pcd 0.2.0
% Matthias G. Eckermann
% April 2026

# NAME

mcp-server-pcd - PCD Model Context Protocol server

# SYNOPSIS

**mcp-server-pcd** *stdio*

**mcp-server-pcd** *http* [*listen=host:port*]

# DESCRIPTION

**mcp-server-pcd** is a Model Context Protocol (MCP) server for the
Post-Coding Development (PCD) workflow. It provides MCP tools and resources
for working with PCD deployment templates, hints files, prompts, and spec
linting.

The binary is self-contained: all templates, hints, and prompts are embedded
at build time. No runtime dependencies are required. Install **pcd-templates**
to enable site-local asset overrides.

# TRANSPORTS

**stdio**
:   Read JSON-RPC 2.0 messages from stdin, write responses to stdout.
    Used by mcphost, Claude Desktop, VS Code, and similar CLI-based MCP hosts.
    The server is launched as a subprocess by the host.

**http**
:   Serve MCP Streamable HTTP transport on the listen address (default:
    127.0.0.1:8080). Used by web-based MCP hosts and remote access scenarios.
    Endpoint: POST /mcp

# OPTIONS

**listen=**_host:port_
:   Bind address for HTTP transport. Default: 127.0.0.1:8080.
    Only valid with the **http** transport.

# MCP TOOLS

**list_templates**
:   List all installed PCD deployment templates. Returns name, version, and
    language for each entry; content is omitted. Use **get_template** to
    retrieve full content.

**get_template** name=_NAME_ [version=_VERSION_]
:   Get a PCD deployment template by name. Returns full Markdown content.
    version defaults to "latest".

**list_resources**
:   List all PCD resources (templates, hints, prompts) as resource URIs.
    URI format: pcd://<type>/<name>

**read_resource** uri=_URI_
:   Read a PCD resource by URI. URI format: pcd://<type>/<name>.
    Types: templates, hints, prompts.

**lint_content** content=_TEXT_ filename=_FILENAME_
:   Validate a PCD specification given as a string. Applies RULE-01 through
    RULE-17. filename must have .md extension.

**lint_file** path=_PATH_
:   Validate a PCD specification file on disk. Applies RULE-01 through RULE-17.
    path must be an absolute path to a .md file.

**get_schema_version**
:   Return the PCD Spec-Schema version this binary was built against.

**set_milestone_status** spec_path=_PATH_ milestone_name=_NAME_ new_status=_STATUS_
:   Set the Status: field of a named MILESTONE section in a spec file on disk.
    Valid status values: pending, active, failed, released.
    Only one milestone may be active at a time.

# ASSET OVERLAY SEARCH PATH

The binary embeds all assets at build time. At startup, it additionally
checks the following directories for overlay assets (last-wins, ascending
precedence):

1. /usr/share/pcd/{templates,hints,prompts}/  (pcd-templates package)
2. /etc/pcd/{templates,hints,prompts}/        (system administrator)
3. ~/.config/pcd/{templates,hints,prompts}/   (user)
4. ./.pcd/{templates,hints,prompts}/          (project-local)

Directories that do not exist are silently skipped.

# SIGNALS

**SIGTERM**, **SIGINT**
:   Graceful shutdown: stop accepting new connections, drain in-flight
    requests (10-second timeout), exit 0.

# EXIT STATUS

**0**
:   Clean shutdown.

**1**
:   HTTP bind failure or stdio transport error.

**2**
:   Invalid command-line arguments.

# EXAMPLES

Start in stdio mode (for mcphost):

    mcp-server-pcd stdio

Start as HTTP service on default address:

    mcp-server-pcd http

Start as HTTP service on custom address:

    mcp-server-pcd http listen=0.0.0.0:9000

mcphost configuration (stdio):

    mcpServers:
      pcd:
        command: mcp-server-pcd
        args: [stdio]

mcphost configuration (http, running as service):

    mcpServers:
      pcd:
        url: http://127.0.0.1:8080/mcp

# FILES

**/usr/bin/mcp-server-pcd**
:   Server binary.

**/usr/lib/systemd/system/mcp-server-pcd.service**
:   Systemd service unit for HTTP transport.

**/usr/share/pcd/**
:   Site-local asset overlay directory (pcd-templates package).

# SEE ALSO

pcd-lint(1)

# LICENSE

GPL-2.0-only — https://spdx.org/licenses/GPL-2.0-only.html
