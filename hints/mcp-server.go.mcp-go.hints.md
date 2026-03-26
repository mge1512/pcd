# Hints: mcp-server.go.mcp-go

Template:  mcp-server
Language:  Go
Library:   github.com/mark3labs/mcp-go
Version:   v0.46.0

---

## Verified dependency string

```
require github.com/mark3labs/mcp-go v0.46.0
```

Use exactly this string in `go.mod`. Do not invent a different version.
Run `go mod tidy` after writing `go.mod` to resolve indirect dependencies.

---

## Key API shapes (v0.46.0)

### Server creation

```go
import "github.com/mark3labs/mcp-go/server"
import "github.com/mark3labs/mcp-go/mcp"

s := server.NewMCPServer(
    "mcp-server-pcdp",
    "0.1.0",
    server.WithToolCapabilities(true),
    server.WithResourceCapabilities(true, true), // subscribe, listChanged
)
```

### Adding a tool

```go
s.AddTool(
    mcp.NewTool("lint_content",
        mcp.WithDescription("Validate a PCDP spec given as a string"),
        mcp.WithString("content",
            mcp.Required(),
            mcp.Description("Full Markdown text of the spec"),
        ),
        mcp.WithString("filename",
            mcp.Description("Used for diagnostic references (default: spec.md)"),
        ),
    ),
    lintContentHandler,
)
```

### Tool handler signature

```go
func lintContentHandler(
    ctx context.Context,
    req mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
    content, _ := req.Params.Arguments["content"].(string)
    filename, _ := req.Params.Arguments["filename"].(string)
    // ...
    return mcp.NewToolResultText(resultJSON), nil
}
```

### Returning errors from a tool

```go
// MCP-level error (wrong params, not-found, etc.)
return nil, fmt.Errorf("unknown template: %s", name)

// Or use structured error for known codes:
return mcp.NewToolResultError("unknown template: " + name), nil
// Note: NewToolResultError marks isError=true in the result,
// which is the MCP-idiomatic way for tool-level errors.
// Reserve returning a Go error for transport/protocol failures.
```

### Adding a resource

```go
s.AddResource(
    mcp.NewResource(
        "pcdp://templates/cli-tool",
        "cli-tool deployment template",
        mcp.WithResourceDescription("Deployment template for CLI tools"),
        mcp.WithMIMEType("text/markdown"),
    ),
    cliToolResourceHandler,
)
```

### Resource handler signature

```go
func cliToolResourceHandler(
    ctx context.Context,
    req mcp.ReadResourceRequest,
) ([]mcp.ResourceContents, error) {
    return []mcp.ResourceContents{
        mcp.TextResourceContents{
            URI:      req.Params.URI,
            MIMEType: "text/markdown",
            Text:     templateContent,
        },
    }, nil
}
```

### Resource templates (dynamic URIs)

For dynamic resources like `pcdp://hints/{key}`, use resource templates:

```go
s.AddResourceTemplate(
    mcp.NewResourceTemplate(
        "pcdp://hints/{key}",
        "PCDP hints file",
        mcp.WithTemplateDescription("Library hints for PCDP translation"),
        mcp.WithTemplateMIMEType("text/markdown"),
    ),
    hintsResourceHandler,
)
```

### stdio transport

```go
if err := server.ServeStdio(s); err != nil {
    fmt.Fprintf(os.Stderr, "stdio error: %v\n", err)
    os.Exit(1)
}
```

### Streamable HTTP transport

```go
httpServer := server.NewStreamableHTTPServer(s)
addr := "127.0.0.1:8080" // override with listen= arg
if err := httpServer.Start(addr); err != nil {
    fmt.Fprintf(os.Stderr, "http bind error: %v\n", err)
    os.Exit(1)
}
```

### Graceful shutdown (http transport)

```go
ctx, stop := signal.NotifyContext(context.Background(),
    syscall.SIGTERM, syscall.SIGINT)
defer stop()

go func() {
    if err := httpServer.Start(addr); err != nil &&
        !errors.Is(err, http.ErrServerClosed) {
        fmt.Fprintf(os.Stderr, "http error: %v\n", err)
        os.Exit(1)
    }
}()

<-ctx.Done()
shutdownCtx, cancel := context.WithTimeout(
    context.Background(), 10*time.Second)
defer cancel()
httpServer.Shutdown(shutdownCtx)
```

---

## Known gotchas

- `mcp.NewToolResultError` sets `isError: true` inside a successful
  tool result — this is the correct MCP idiom for domain errors
  (unknown template, lint failure, etc.). Do NOT return a Go `error`
  for these cases; that signals a transport/protocol failure.

- Resource handlers must return `[]mcp.ResourceContents`, not a single
  item. Always wrap in a slice.

- `WithResourceCapabilities(subscribe bool, listChanged bool)` —
  set both to `true` to advertise full resource support to clients.

- The HTTP endpoint is `/mcp` by default in `NewStreamableHTTPServer`.
  Do not change this without a good reason — MCP clients expect `/mcp`.

- `server.ServeStdio` blocks until stdin is closed. Do not call it in
  a goroutine unless you handle the return value.

- For the transport selector: check `os.Args` for bare words "stdio"
  and "http" before creating the server, not inside a handler.
