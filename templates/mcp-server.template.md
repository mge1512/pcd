

# mcp-server.template

## META
Deployment:  template
Version:     0.3.13
Spec-Schema: 0.3.13
Author:      Matthias G. Eckermann <pcdp@mailbox.org>
License:     CC-BY-4.0
Verification: none
Safety-Level: QM
Template-For: mcp-server

---

## TYPES

```
Language := Go | Python | Rust
// Go is the default. Python and Rust are supported alternatives.
// No other languages permitted for MCP servers in v1.

GoFramework := mcp-go | go-sdk
// mcp-go:  github.com/mark3labs/mcp-go
//          Community SDK. Supports stdio, Streamable HTTP, SSE, in-process.
//          Default choice: covers all required transports.
// go-sdk:  github.com/modelcontextprotocol/go-sdk
//          Official SDK, maintained with Google. Supports stdio and command
//          transports. HTTP transport support in progress.
//          Choose when: spec compliance priority or official SDK required.
// The framework is selected via preset; not declared in the spec itself.

Transport := stdio | streamable-http
// Both transports are required for all mcp-server components.
// stdio:           Used by mcphost, Claude Desktop, VS Code, and similar
//                  CLI-based MCP hosts. The server reads from stdin and
//                  writes to stdout. Launched as a subprocess by the host.
// streamable-http: Used by web-based MCP hosts, remote access, and
//                  multi-client scenarios. HTTP POST for requests,
//                  Server-Sent Events (SSE) for streaming responses.
//                  Endpoint: /mcp (default), configurable via listen=

MCPCapability := tools | resources | prompts | sampling
// tools:     Functions the LLM can call. Primary capability for most servers.
// resources: Read-only data sources the LLM can access (files, DB records).
// prompts:   Reusable prompt templates the LLM can request.
// sampling:  Server-initiated LLM sampling (advanced, rarely needed).

ToolName := string where matches "^[a-z][a-z0-9_-]*$"
// Snake_case or hyphen-separated. No spaces. No uppercase.
// Must be unique within the server.

ErrorCode := integer
// MCP error codes follow JSON-RPC 2.0 conventions:
// -32700: Parse error
// -32600: Invalid request
// -32601: Method not found
// -32602: Invalid params
// -32603: Internal error
// Application-specific errors: -32000 to -32099

ListenAddress := string
// Default: 127.0.0.1:8080 for streamable-http transport
// Override via listen= key=value argument
```

---

## TEMPLATE-TABLE

| Key | Value | Constraint | Notes |
|-----|-------|------------|-------|
| VERSION | MAJOR.MINOR.PATCH | required | |
| SPEC-SCHEMA | MAJOR.MINOR.PATCH | required | |
| AUTHOR | name <email> | required | Repeating field. |
| LICENSE | SPDX identifier | required | |
| LANGUAGE | Go | default | Primary language for MCP servers. |
| LANGUAGE-ALTERNATIVES | Python | supported | Via preset override. Use official MCP Python SDK. |
| LANGUAGE-ALTERNATIVES | Rust | supported | Via preset override. Use official MCP Rust SDK. |
| GO-FRAMEWORK | mcp-go | default | github.com/mark3labs/mcp-go. Covers both required transports. |
| GO-FRAMEWORK-ALTERNATIVE | go-sdk | supported | github.com/modelcontextprotocol/go-sdk. HTTP transport in progress; choose for spec compliance. |
| TRANSPORT | stdio | required | Always implement. Primary transport for CLI-based hosts. |
| TRANSPORT | streamable-http | required | Always implement. Required for web and remote access. |
| BINARY-TYPE | static | required | Single static binary. No runtime dependencies. CGO_ENABLED=0 for Go. |
| BINARY-COUNT | 1 | required | One binary serves both transports, selected by invocation mode. |
| CLI-ARG-STYLE | key=value | required | Consistent with pcdp conventions. |
| TRANSPORT-SELECTION | bare-word | required | Transport selected by bare word: `mcp-server-{n} stdio` or `mcp-server-{n} http` |
| LISTEN-DEFAULT | 127.0.0.1:8080 | default | Default bind address for streamable-http. Override: listen=host:port |
| MCP-CAPABILITY | tools | required | All MCP servers must implement at least tools capability. |
| MCP-CAPABILITY | resources | supported | Optional. Declare in spec BEHAVIOR sections if used. |
| MCP-CAPABILITY | prompts | supported | Optional. Declare in spec BEHAVIOR sections if used. |
| NETWORK-CALLS | context-dependent | supported | MCP servers exist to provide data — network calls are permitted and expected. Declare external dependencies in DEPLOYMENT section. |
| CONFIG-ENV-VARS | forbidden | forbidden | Configuration via key=value arguments or preset files only. Not environment variables. |
| ERROR-HANDLING | JSON-RPC 2.0 | required | All errors must be returned as JSON-RPC error responses, not panics. |
| SIGNAL-HANDLING | SIGTERM | required | Clean shutdown: drain in-flight requests, close connections, exit 0. |
| SIGNAL-HANDLING | SIGINT | required | Same as SIGTERM. |
| IDEMPOTENT-TOOLS | recommended | supported | Tools should be idempotent where possible. Document side effects in tool descriptions. |
| OUTPUT-FORMAT | RPM | required | OBS RPM package. |
| OUTPUT-FORMAT | DEB | required | OBS DEB package. |
| OUTPUT-FORMAT | OCI | required | Containerfile for OCI image. MCP servers are commonly deployed as containers. |
| OUTPUT-FORMAT | binary | supported | Raw binary for direct installation. |
| INSTALL-METHOD | OBS | required | Primary distribution via OBS. |
| INSTALL-METHOD | curl | forbidden | Supply chain security requirement. |
| PLATFORM | Linux | required | Primary platform. |
| PLATFORM | macOS | supported | Supported for development use. |
| PLATFORM | Windows | supported | Supported via stdio transport only in v1. |

---

## DELIVERABLES

Defines the files a translator must produce. All required OUTPUT-FORMATs
must be produced. Supported OUTPUT-FORMATs are produced if active in preset.

### Delivery Order

1. Core implementation files
2. Required packaging artifacts (RPM, DEB, OCI)
3. Supported artifacts if preset active
4. TRANSLATION_REPORT.md last, after all files verified and compiled

### Deliverables Table

| OUTPUT-FORMAT | Constraint | Required Deliverable Files | Notes |
|---|---|---|---|
| source | required | `main.go`, `go.mod` | Single file preferred under ~800 lines. Tool handlers may be split into `tools.go`, `resources.go` etc. for larger servers. |
| build | required | `Makefile` | Targets: `build`, `test`, `install`, `clean`. CGO_ENABLED=0. |
| docs | required | `README.md` | Must document: installation (zypper/apt/dnf), invocation (stdio and http modes), tool list with descriptions, configuration. |
| license | required | `LICENSE` | Full license text or reference to SPDX identifier with authoritative URL. |
| RPM | required | `{n}.spec` | OBS RPM spec. Must include systemd service unit for http transport mode. |
| DEB | required | `debian/control`, `debian/changelog`, `debian/rules`, `debian/copyright` | DEP-5 copyright. Must include systemd service unit. |
| OCI | required | `Containerfile` | Multi-stage build. Builder FROM registry.suse.com/bci/golang:latest. Final stage FROM scratch. Expose port 8080 for http transport. ENTRYPOINT default to http mode. Never use unqualified image names. |
| report | required | `TRANSLATION_REPORT.md` | Must include: framework choice rationale, transport implementation notes, tool list with descriptions, compilation result. |

### Systemd Service Unit

For http transport mode, a systemd service unit is required in both RPM
and DEB packaging:

```ini
[Unit]
Description={component name} MCP server
After=network.target

[Service]
Type=simple
ExecStart=/usr/bin/mcp-server-{n} http listen=127.0.0.1:8080
Restart=on-failure
NoNewPrivileges=true
ProtectSystem=strict
ProtectHome=true

[Install]
WantedBy=multi-user.target
```

### Deliverable Content Requirements

**main.go must implement both transports:**
```go
// Invocation:
//   mcp-server-{n} stdio          ← serve over stdin/stdout
//   mcp-server-{n} http           ← serve over HTTP (default: 127.0.0.1:8080)
//   mcp-server-{n} http listen=host:port
func main() {
    // Parse transport from first argument (bare word: stdio | http)
    // Parse key=value options
    // Initialise server with capabilities declared in spec
    // Run selected transport
}
```

**go.mod framework selection:**

For mcp-go (default):
```
require github.com/mark3labs/mcp-go vX.Y.Z
```

For go-sdk (alternative):
```
require github.com/modelcontextprotocol/go-sdk vX.Y.Z
```

**Containerfile:**
```dockerfile
FROM registry.suse.com/bci/golang:latest AS builder
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o mcp-server-{n} .

FROM scratch
COPY --from=builder /build/mcp-server-{n} /usr/bin/mcp-server-{n}
EXPOSE 8080
ENTRYPOINT ["/usr/bin/mcp-server-{n}", "http"]
```

Builder stage must use `registry.suse.com/bci/golang:latest`.
Never use unqualified names such as `golang:1.24` or `docker.io/golang`.
Final stage must be `FROM scratch` for static binaries.
Do not include a package manager in the final image.

---

## BEHAVIOR: stdio-transport
Constraint: required

INPUTS:
```
(reads JSON-RPC 2.0 messages from stdin)
```

OUTPUTS:
```
(writes JSON-RPC 2.0 messages to stdout)
```

PRECONDITIONS:
- stdin is connected to an MCP-compatible host
- Server capabilities are declared in spec BEHAVIOR sections

STEPS:
1. Register all tools declared in spec BEHAVIOR sections.
2. Enter read loop: read one JSON-RPC 2.0 message from stdin.
3. On parse error → write JSON-RPC error response to stdout; continue loop.
4. Dispatch message to registered handler.
   MECHANISM: handle each request in its own goroutine so slow handlers
   do not block stdin reads.
5. Write JSON-RPC response to stdout (synchronise writes; no interleaving).
6. On EOF from stdin → drain in-flight goroutines, exit 0.
7. On SIGTERM/SIGINT → drain in-flight goroutines, exit 0.

POSTCONDITIONS:
- Server responds to all MCP protocol messages
- Errors are returned as JSON-RPC error responses, never as panics
- On EOF from stdin: clean shutdown, exit 0
- On SIGTERM/SIGINT: drain in-flight requests, exit 0

---

## BEHAVIOR: http-transport
Constraint: required

INPUTS:
```
listen: string   // host:port to bind; default 127.0.0.1:8080
```

OUTPUTS:
```
(serves HTTP POST /mcp for requests)
(serves GET /mcp with SSE for streaming)
```

PRECONDITIONS:
- listen address is available and bindable
- Server capabilities declared in spec BEHAVIOR sections

STEPS:
1. Register all tools declared in spec BEHAVIOR sections.
2. Bind HTTP listener on listen address; on bind error → exit 1 with diagnostic.
3. Register POST /mcp handler (JSON-RPC 2.0 request/response).
4. Register GET /mcp handler (Server-Sent Events for streaming responses).
5. Accept connections.
   On POST /mcp: parse JSON-RPC body; on parse error → HTTP 400 + JSON-RPC error body.
   Dispatch to handler; return HTTP 200 + JSON-RPC response body (errors included in body per spec).
   On GET /mcp: upgrade to SSE; stream responses until handler completes or client disconnects.
6. On SIGTERM/SIGINT → stop accepting new connections, drain in-flight requests,
   close listener, exit 0.
   MECHANISM: use context cancellation propagated from signal handler.

POSTCONDITIONS:
- Server accepts connections on listen address
- POST /mcp handles JSON-RPC 2.0 requests
- GET /mcp streams Server-Sent Events for streaming responses
- On SIGTERM/SIGINT: drain in-flight requests, close listener, exit 0
- Errors returned as HTTP 4xx/5xx with JSON-RPC error body

---

## PRECONDITIONS

- Both stdio and streamable-http transports must be implemented
- At least one MCP tool must be declared in the spec BEHAVIOR sections
- All tool names must match ToolName type (lowercase, no spaces)
- External service dependencies must be declared in DEPLOYMENT section
- Server must not read configuration from environment variables

---

## POSTCONDITIONS

- Binary serves both transports from the same binary
- Transport selected by bare word argument (stdio | http)
- All tools declared in spec BEHAVIOR sections are registered and callable
- All errors are JSON-RPC 2.0 error responses
- Server handles concurrent tool calls safely (Go goroutines)
- Systemd service units are included in RPM and DEB packaging

---

## INVARIANTS

- [observable]      both transports always implemented — never one without the other
- [observable]      tool names are lowercase and match ^[a-z][a-z0-9_-]*$
- [observable]      all errors are JSON-RPC 2.0 formatted — no panics reach the client
- [implementation]  CGO_ENABLED=0 for Go — static binary, no libc dependency
- [observable]      configuration via key=value only — no environment variables
- [observable]      TRANSLATION_REPORT.md documents the Go framework choice
- [observable]      template version recorded in every audit bundle

---

## EXAMPLES

EXAMPLE: minimal_tool_server_stdio
GIVEN:
  spec declares one BEHAVIOR: tool named "greet"
  invocation: mcp-server-{n} stdio
  MCP host sends: {"method": "tools/call", "params": {"name": "greet", "arguments": {"name": "World"}}}
WHEN:
  server processes the tool call
THEN:
  server responds with JSON-RPC result containing greeting text
  response written to stdout
  exit_code = 0 on EOF

EXAMPLE: minimal_tool_server_http
GIVEN:
  spec declares one BEHAVIOR: tool named "greet"
  invocation: mcp-server-{n} http listen=127.0.0.1:9000
WHEN:
  MCP host sends POST /mcp with tool call JSON-RPC body
THEN:
  server returns 200 with JSON-RPC result body
  server remains running, accepts further requests

EXAMPLE: unknown_transport_argument
GIVEN:
  invocation: mcp-server-{n} websocket
WHEN:
  server starts
THEN:
  stderr = "error: unknown transport 'websocket'. Valid: stdio, http"
  exit_code = 2

EXAMPLE: tool_error_handling
GIVEN:
  invocation: mcp-server-{n} stdio
  MCP host calls a tool that encounters an internal error
WHEN:
  tool handler returns an error
THEN:
  server returns JSON-RPC error response (not a panic)
  stderr contains error log entry with timestamp
  server continues running and accepts further requests

EXAMPLE: graceful_shutdown_http
GIVEN:
  server running in http mode with active connection
  SIGTERM signal received
WHEN:
  signal handler fires
THEN:
  server stops accepting new connections
  in-flight requests complete (up to 30 second drain timeout)
  server exits 0

EXAMPLE: framework_mcp_go_default
GIVEN:
  no preset override for GO-FRAMEWORK
  translator reads TEMPLATE-TABLE
WHEN:
  translator selects Go framework
THEN:
  go.mod contains: require github.com/mark3labs/mcp-go
  TRANSLATION_REPORT.md documents: "GO-FRAMEWORK: mcp-go (template default)"
  both stdio and http transports implemented using mcp-go API

EXAMPLE: framework_go_sdk_override
GIVEN:
  preset declares: GO-FRAMEWORK = go-sdk
  translator reads resolved preset
WHEN:
  translator selects Go framework
THEN:
  go.mod contains: require github.com/modelcontextprotocol/go-sdk
  TRANSLATION_REPORT.md documents: "GO-FRAMEWORK: go-sdk (preset override)"
  translator notes in report if http transport required workaround

---

## EXECUTION

The translator must read this section before generating any code.
It specifies the exact delivery phases, resume logic, and compile gate
for mcp-server components. Follow it exactly.

### Input files

The translator receives in the working directory:
- `mcp-server.template.md` — this deployment template
- `<spec-name>.md` — the component specification

If the spec's DEPENDENCIES section references hints files, they are also
present. Read them before writing `go.mod` or any code that uses them.

### Resume logic

Before writing any file, list the output directory.
If a listed deliverable already exists and is non-empty, skip it — treat
it as complete and move to the next missing file. Report which files were
found and which are being produced.

### Delivery phases

Produce files in this exact order. Complete each phase before starting
the next. Do not produce `TRANSLATION_REPORT.md` until Phase 5 is done.

**Phase 1 — Core implementation**
- `main.go` — implements both stdio and http transports; transport selected
  by bare-word argument (`stdio` | `http`)
- `go.mod` — direct dependencies only; see Compile gate below

**Phase 2 — Build and packaging**
- `Makefile`
- `<n>.spec` (RPM spec; must include systemd service unit for http transport)
- `debian/control`, `debian/changelog`, `debian/rules`, `debian/copyright`
  (must include systemd service unit for http transport)
- `Containerfile` — multi-stage; builder FROM registry.suse.com/bci/golang:latest;
  final stage FROM scratch; EXPOSE 8080; ENTRYPOINT default to http transport
- `LICENSE`

**Phase 3 — Test infrastructure**
- `independent_tests/INDEPENDENT_TESTS.go`
- `translation_report/translation-workflow.pikchr`

**Phase 4 — Documentation**
- `README.md` — must document stdio invocation, http invocation with
  `listen=` option, and all exposed MCP tools

**Phase 5 — Compile gate** (see below)

**Phase 6 — Report (last)**
- `TRANSLATION_REPORT.md`

### Compile gate

Execute after Phase 4 and before Phase 6. If your environment cannot
execute shell commands, document this explicitly under the heading
"Phase 5 — Compile gate not executed" in TRANSLATION_REPORT.md and
state why. Do not silently omit this phase.

**Step 1 — Framework selection**

Determine the Go MCP framework from the resolved preset:
- Default: `github.com/mark3labs/mcp-go` (supports stdio + HTTP natively)
- Alternative: `github.com/modelcontextprotocol/go-sdk` (if preset declares it;
  note HTTP transport status in TRANSLATION_REPORT.md)

Document the framework choice and rationale in TRANSLATION_REPORT.md.

**Step 2 — Dependency resolution**

Run: `go mod tidy`

Do not hand-write indirect dependencies.
If `go mod tidy` cannot be run, note in TRANSLATION_REPORT.md.

**Step 3 — Compilation**

Run: `go build ./...`

If compilation fails, fix only the identified errors and re-run.
Do not rewrite unaffected files.

**Step 4 — Record result**

Record pass/fail in TRANSLATION_REPORT.md.
Once all steps pass, do not modify source files further.
Proceed immediately to Phase 6.

### Dual-transport requirement

Both stdio and http transports must be implemented in the same binary.
Transport selection is by bare-word argument only (`stdio` | `http`).
The default listen address for http is `127.0.0.1:8080`.
Override via `listen=host:port` key=value argument.

### go.mod rules

- Declare only direct dependencies
- Do NOT hand-write indirect dependencies (resolved by `go mod tidy`)
- Do NOT fabricate pseudo-versions or commit hashes for untagged modules


---

## DEPLOYMENT

Runtime: long-running server process, single static binary

Invocation:
  mcp-server-{n} stdio
  mcp-server-{n} http
  mcp-server-{n} http listen=host:port

Transport selection (bare word, first argument):
  stdio    stdin/stdout JSON-RPC 2.0
  http     Streamable HTTP on listen address

Key=value options:
  listen=host:port    HTTP listen address (default: 127.0.0.1:8080)

Output streams:
  stdout: JSON-RPC 2.0 protocol messages (stdio transport)
          or unused (http transport — do not mix with protocol messages)
  stderr: Server logs, errors, startup messages

Naming convention:
  Binary name: mcp-server-{n}
  where {n} is the use-case identifier, lowercase, hyphen-separated
  Examples: mcp-server-pcdp, mcp-server-github, mcp-server-postgres

Installation:
  OBS package: mcp-server-{n}
  Available for: openSUSE Leap, SUSE Linux Enterprise, Fedora, Debian/Ubuntu
  No curl-based installation.
  Systemd service for http mode included in package.

Platform:
  Linux (primary, both transports)
  macOS (supported, both transports)
  Windows (supported, stdio transport only in v1)

Go Framework Selection Guide:
  mcp-go (default):  Quick setup, both transports, active community,
                     slightly more opinionated API
  go-sdk (alt):      Official spec compliance, Google collaboration,
                     stdio well-supported; check HTTP transport status
                     before selecting for production http use

Location: /usr/share/pcdp/templates/mcp-server.template.md


