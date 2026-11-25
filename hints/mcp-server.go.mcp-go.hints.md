# Hints: mcp-server.go.mcp-go

Template:  mcp-server
Language:  Go
Library:   github.com/mark3labs/mcp-go
Version:   v1.6.0
Release:   1

`Version` is the version of the library. `Release` counts the changes of
this file since `Version` was last set; a new library version removes the
line.

---

## Verified dependency string

```
require github.com/mark3labs/mcp-go v1.6.0
```

Use exactly this string in `go.mod`. Do not invent a different version.
Run `go mod tidy` after writing `go.mod` to resolve indirect dependencies.

v1.6.0 has `go 1.25.5` in its `go.mod`. The build needs Go 1.25.5 or
later, and the `go` line of the server's `go.mod` says `go 1.25.5` or
later.

---

## Asset embedding — Go implementation

The spec requires all assets (templates, hints, prompts) to be compiled
into the binary using the language's native asset embedding mechanism.
In Go this is `embed.FS` via `//go:embed` directives.

### Directory layout at build time

The `//go:embed` directive resolves paths relative to the `.go` file
that contains it. Place the directives in `internal/store/store.go`.
The staged asset directories must sit at:

```
internal/store/
  store.go              <- contains //go:embed directives
  assets/
    templates/          <- staged from repo-root/templates/
    hints/              <- staged from repo-root/hints/
    prompts/            <- staged from repo-root/prompts/
```

The `assets/` directory and its subdirectories are build artefacts.
They must be in `.gitignore` and must not be committed.

### embed.FS declarations in store.go

```go
import "embed"

//go:embed assets/templates/*
var embeddedTemplates embed.FS

//go:embed assets/hints/*
var embeddedHints embed.FS

//go:embed assets/prompts/*
var embeddedPrompts embed.FS
```

`embed.FS` preserves the directory prefix. A file staged at
`assets/templates/cli-tool.template.md` is accessed as:
`embeddedTemplates.ReadFile("assets/templates/cli-tool.template.md")`.
Use `fs.WalkDir` to enumerate files and strip the prefix when deriving keys.

### Key derivation

```go
// templates: "assets/templates/cli-tool.template.md" -> "cli-tool"
// hints:     "assets/hints/python-tool.hints.md"     -> "python-tool"
// prompts:   "assets/prompts/interview-prompt.md"    -> "interview"
// prompts:   "assets/prompts/translator.md"          -> "translator"

func assetKey(path, suffix string) string {
    base := filepath.Base(path)
    base = strings.TrimSuffix(base, suffix)
    base = strings.TrimSuffix(base, ".md")
    return base
}
// Call as: assetKey(path, ".template"), assetKey(path, ".hints"), assetKey(path, "")
```

### Loading embedded assets at startup

```go
import iofs "io/fs"

func loadEmbedded(fsys embed.FS, dir, suffix string) (map[string]string, error) {
    result := make(map[string]string)
    err := iofs.WalkDir(fsys, dir, func(path string, d iofs.DirEntry, err error) error {
        if err != nil || d.IsDir() || !strings.HasSuffix(path, ".md") {
            return err
        }
        data, err := fsys.ReadFile(path)
        if err != nil {
            return err
        }
        result[assetKey(path, suffix)] = string(data)
        return nil
    })
    return result, err
}
```

Call once at startup, before applying filesystem overlays:

```go
templates, _ := loadEmbedded(embeddedTemplates, "assets/templates", ".template")
hints,     _ := loadEmbedded(embeddedHints,     "assets/hints",     ".hints")
prompts,   _ := loadEmbedded(embeddedPrompts,   "assets/prompts",   "")
```

Then apply filesystem overlay directories in last-wins order, using
`os.ReadFile` / `filepath.WalkDir` on each of:
  `/usr/share/pcd/{templates,hints,prompts}/`
  `/etc/pcd/{templates,hints,prompts}/`
  `~/.config/pcd/{templates,hints,prompts}/`
  `./.pcd/{templates,hints,prompts}/`

Missing directories are silently skipped.

---

## Vendoring: `vendor/` goes into a tarball of its own

`vendor/` is never committed, and it is never part of the source tarball.
It is packaged in a separate tarball with this name:

```
%{name}-%{version}-vendor.tar.xz
```

The Makefile makes `vendor/` and the tarball. The RPM spec names the
tarball as `Source1` and unpacks it into the source tree. The two sections
below show both. In them the package has the name of the binary, so the
Makefile writes the same file name as `$(BINARY)-$(VERSION)-vendor.tar.xz`.

Wrong: the dependencies inside the source tarball.

```spec
# The source tarball contains vendor/
Source0:  %{name}-%{version}.tar.xz

%prep
%autosetup
```

Wrong: a `build` target that needs `vendor/` and does not make it. It
fails in a fresh clone.

```makefile
build:
	CGO_ENABLED=0 go build -mod=vendor -o $(BINARY) .
```

---

## Makefile — asset staging and vendoring

```makefile
BINARY  := mcp-server-pcd
VERSION := 0.1.0
STORE   := internal/store/assets

.PHONY: all build embed-assets vendor-tarball clean

all: build

# Stage assets from repo root into the embed location.
# Must run before go build.
embed-assets:
	mkdir -p $(STORE)/templates $(STORE)/hints $(STORE)/prompts
	cp ../../templates/*.template.md $(STORE)/templates/
	cp ../../hints/*.hints.md        $(STORE)/hints/
	cp ../../prompts/*.md            $(STORE)/prompts/

# vendor/ is a directory, not a phony target. This rule runs when vendor/
# is missing or older than go.mod or go.sum.
vendor: go.mod go.sum
	go mod vendor
	touch vendor

build: embed-assets vendor
	CGO_ENABLED=0 go build \
	    -mod=vendor \
	    -ldflags="-X main.serverVersion=$(VERSION)" \
	    -o $(BINARY) \
	    .

# Generate vendor tarball for OBS/RPM builds — run once per release.
vendor-tarball: vendor
	tar -cJf $(BINARY)-$(VERSION)-vendor.tar.xz vendor/

clean:
	rm -rf $(STORE) $(BINARY) vendor
```

A `test` target depends on `vendor` in the same way.

`.gitignore` additions:

```
internal/store/assets/
vendor/
*-vendor.tar.xz
mcp-server-pcd
```

---

## RPM spec — vendoring and asset staging

```spec
Name:     mcp-server-pcd
Version:  0.1.0
Release:  1%{?dist}
Source0:  %{name}-%{version}.tar.xz
Source1:  %{name}-%{version}-vendor.tar.xz

# Source1 generated by: go mod vendor &&
#   tar -cJf mcp-server-pcd-0.1.0-vendor.tar.xz vendor/

BuildRequires:  golang >= 1.25
Recommends:     pcd-templates

%prep
%autosetup -a 1
# -a 1 unpacks Source1 (vendor/) alongside Source0.

%build
make embed-assets
CGO_ENABLED=0 go build \
    -mod=vendor \
    -ldflags="-X main.serverVersion=%{version}" \
    -o %{name} \
    .
```

`make embed-assets` in `%build` populates `internal/store/assets/`
before `go build` reads it via `//go:embed`. The asset directories
are not in the source tarball — they are created during the build.

`%build` calls `go build -mod=vendor` itself and not `make build`. A
`%check` section calls `go test -mod=vendor` in the same way and not
`make test`. The `vendor` rule of the Makefile compares file times, and a
package build has no network: it must never start `go mod vendor`.

---

## Debian package — vendoring

The Debian source package has the vendor tarball as a component
tarball. Its file name is the Debian form of the same tarball:

```
<source>_<version>.orig-vendor.tar.xz
```

Make it from the tarball of `make vendor-tarball` by a copy with that
name; the content stays the same. The source package uses the format
`3.0 (quilt)`, so the deliverables have one more file:

```
debian/source/format:  3.0 (quilt)
```

dpkg-source unpacks a component tarball `orig-vendor` into `vendor/` of
the source tree. A tarball with the single top directory `vendor/`
gives the same tree. `debian/rules` then builds with `-mod=vendor`, as
`%build` of the RPM spec does, and never runs `go mod vendor`.

Wrong: `debian/rules` with `GOFLAGS = -mod=vendor` and no component
tarball. The build tree has no `vendor/`, and the build fails.

---

## Key API shapes (v1.6.0)

### Server creation

```go
import "github.com/mark3labs/mcp-go/server"
import "github.com/mark3labs/mcp-go/mcp"

s := server.NewMCPServer(
    "mcp-server-pcd",
    "0.1.0",
    server.WithToolCapabilities(true),
    server.WithResourceCapabilities(true, true),
)
```

### Adding a tool

```go
s.AddTool(
    mcp.NewTool("lint_content",
        mcp.WithDescription("Validate a PCD spec given as a string"),
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
    content := req.GetString("content", "")
    filename := req.GetString("filename", "spec.md")
    // ...
    return mcp.NewToolResultText(resultJSON), nil
}
```

### Returning errors from a tool

```go
// Domain error (wrong params, not-found, etc.) — use NewToolResultError.
return mcp.NewToolResultError("unknown template: " + name), nil
// NewToolResultError sets isError=true — the MCP-idiomatic way.
// Reserve returning a Go error for transport/protocol failures only.
```

### Adding a resource

```go
s.AddResource(
    mcp.NewResource(
        "pcd://templates/cli-tool",
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

```go
s.AddResourceTemplate(
    mcp.NewResourceTemplate(
        "pcd://hints/{key}",
        "PCD hints file",
        mcp.WithTemplateDescription("Library hints for PCD translation"),
        mcp.WithTemplateMIMEType("text/markdown"),
    ),
    hintsResourceHandler,
)
```

Register resource templates for all three dynamic URI patterns:
`pcd://templates/{name}`, `pcd://prompts/{name}`, `pcd://hints/{key}`.

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
addr := "127.0.0.1:8080"
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
  tool result — this is the correct MCP idiom for domain errors.
  Do NOT return a Go `error` for these cases.

- Resource handlers must return `[]mcp.ResourceContents`, not a single
  item. Always wrap in a slice.

- `WithResourceCapabilities(subscribe bool, listChanged bool)` —
  set both to `true` to advertise full resource support to clients.

- The HTTP endpoint is `/mcp` by default in `NewStreamableHTTPServer`.
  Do not change this.

- `server.ServeStdio` blocks until stdin is closed. Do not call it in
  a goroutine unless you handle the return value.

- For the transport selector: check `os.Args` for bare words "stdio"
  and "http" before creating the server, not inside a handler.

- `embed.FS` paths always use forward slashes even on Windows.
  Use `path.Join` (not `filepath.Join`) when constructing embed paths.

- `//go:embed` requires the target directory to exist and be non-empty
  at compile time. If `internal/store/assets/` is missing, `go build`
  fails immediately. Run `make embed-assets` first.

- Do not commit `internal/store/assets/` to the repository.
  The directory is a build artefact, recreated by `make embed-assets`.

- `req.Params.Arguments` has the type `any` and cannot be indexed. Read
  arguments with `req.GetString(key, default)`, `req.RequireString(key)`
  or `req.GetArguments()`.

- A tool filter (`server.WithToolFilter`) runs for `tools/list` and for
  every `tools/call`. A tool that the filter removes cannot be called.
  Do no network request in a filter. To change the tool list of one
  request, use `hooks.AddAfterListTools` with `server.WithHooks`: it
  runs for `tools/list` only.

- `server.WithRecovery()` turns a panic in a tool handler into a Go error,
  so the client gets a JSON-RPC error. It writes nothing to standard
  error. When the template asks for an error log entry, recover in the
  handler and write the line there:

  ```go
  func guarded(h server.ToolHandlerFunc) server.ToolHandlerFunc {
      return func(ctx context.Context, req mcp.CallToolRequest) (res *mcp.CallToolResult, err error) {
          defer func() {
              if p := recover(); p != nil {
                  fmt.Fprintf(os.Stderr, "%s internal error: tool %s: %v\n",
                      time.Now().UTC().Format(time.RFC3339), req.Params.Name, p)
                  res, err = nil, errors.New("internal error of the server")
              }
          }()
          return h(ctx, req)
      }
  }
  ```

- The Go package `regexp` refuses a repeat count above 1000.
  `regexp.MustCompile("^[!-~]{1,4096}$")` panics when the program starts,
  with `invalid repeat count`. Check such a length with `len` and the
  characters with a pattern without a count, for example `^[!-~]+$`.
