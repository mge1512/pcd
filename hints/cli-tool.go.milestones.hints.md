# Hints: cli-tool scaffold-first milestones — Go implementation

Template:  cli-tool
Language:  Go
Topic:     scaffold-first milestone pattern

These hints apply to any cli-tool specification that uses the
scaffold-first milestone pattern (Scaffold: true on M0).
They are advisory. They cannot override spec invariants.
Read the spec and deployment template first.
Read these hints before writing any code.

---

## The scaffold-first rule

M0 is different from all other milestones. Its sole purpose is to
produce a complete, compilable skeleton. The only acceptance criterion
is a clean compile. No real logic is implemented in M0.

All subsequent milestones fill in stub bodies. They never create new
files, never add new types, never restructure packages. If M0 is
correct, every later milestone is a focused fill-in exercise.

---

## File layout principle

The scaffold milestone must create all files the completed implementation
will ever need. Split source into logically cohesive files so that each
milestone touches at most two or three files. This keeps context window
pressure low in later passes.

Minimum required files for any cli-tool:

```
main.go        — entry point, argument parsing, main()
types.go       — all type and struct definitions
interfaces.go  — all interface definitions + production and test-double
                 implementations
main_test.go   — unit tests using test doubles only
```

Larger components should additionally split by domain concern:
one file per major BEHAVIOR group (e.g. collect_storage.go,
a dedicated render file). The spec's MILESTONE groupings are a guide:
BEHAVIORs that appear together in one milestone belong in one file.

---

## M0 stub convention

Every stub function must:

1. Have the **correct signature** matching what its caller expects.
   Wrong signatures cause compile failures in later milestones.

2. Return the **correct zero value** for its declared output type.
   Never return nil where a non-nil zero value is semantically required.

   For collection functions that populate output scopes:
   ```go
   // CORRECT — empty but valid; JSON produces {"_attributes":{},"_elements":[]}
   scope := &MyScope{
       Attributes: map[string]interface{}{},
       Elements:   []MyRecord{},
   }

   // WRONG — JSON produces null; breaks consumers expecting an object
   var scope *MyScope = nil
   ```

3. Be **silent at normal verbosity**. Stubs must not print
   "not yet implemented" during normal runs. Use a debug-only log gate:
   ```go
   func debugLog(msg string) {
       if os.Getenv("APP_DEBUG") == "1" {
           fmt.Fprintf(os.Stderr, "DEBUG: %s\n", msg)
       }
   }
   ```
   Call `debugLog("behaviorName: not yet implemented")` from stubs if
   desired. Replace APP_DEBUG with the component's own env var name.

4. **Compile cleanly** with no unused imports.

---

## OSCommandRunner.Run — must NOT be a stub

This is the most critical M0 requirement for any cli-tool that invokes
external commands. Implement it in full in M0:

```go
func (r *OSCommandRunner) Run(cmd string, args []string) (string, string, error) {
    oldPath := os.Getenv("PATH")
    os.Setenv("PATH", "/sbin:/bin:/usr/bin:/usr/sbin")
    defer os.Setenv("PATH", oldPath)

    c := exec.Command(cmd, args...)
    var stdout, stderr bytes.Buffer
    c.Stdout = &stdout
    c.Stderr = &stderr
    err := c.Run()
    return stdout.String(), stderr.String(), err
}
```

A stub returning ("", "", nil) for all inputs causes every
command-dependent module to silently produce empty output while the
compile gate passes. This is the single most common failure mode in
scaffold-first translations.

Required imports: os, os/exec, bytes.

---

## JSON struct tags

Go struct fields must be exported (capitalised) for JSON marshalling,
but the JSON output must use the field names declared in the spec
(typically underscore_style). Use struct tags:

```go
type MyMeta struct {
    FormatVersion int    `json:"format_version"`
    CollectedAt   string `json:"collected_at"`
}
```

Without struct tags, FormatVersion marshals as "FormatVersion" —
wrong for any schema that uses underscore_style keys.

Every struct that appears in JSON output needs explicit json: tags
on every field. Missing tags are a silent correctness bug.

---

## ScopeWrapper pattern

When the spec uses ScopeWrapper<T> or the _attributes/_elements pattern,
implement in Go as:

```go
// Option A — generics (Go 1.21+, preferred)
type ScopeWrapper[T any] struct {
    Attributes map[string]interface{} `json:"_attributes"`
    Elements   []T                    `json:"_elements"`
}

// Option B — concrete struct per scope (always works)
type MyScope struct {
    Attributes map[string]interface{} `json:"_attributes"`
    Elements   []MyRecord             `json:"_elements"`
}
```

Either option is acceptable. Initialise every scope to an
empty-but-valid state in stubs:
```go
scope := MyScope{
    Attributes: map[string]interface{}{},
    Elements:   []MyRecord{},
}
```

---

## Renderer stubs for M0

When the spec declares a Renderer interface with multiple format
implementations, define all implementations as stubs in M0:

```go
type HTMLRenderer struct{}
func (r *HTMLRenderer) Header(/*...*/) string        { return "" }
func (r *HTMLRenderer) TOC(/*...*/) string           { return "" }
func (r *HTMLRenderer) Section(/*...*/) string       { return "" }
func (r *HTMLRenderer) Footer() string               { return "" }
func (r *HTMLRenderer) Escape(raw string) string     { return raw }
```

Repeat for every declared renderer type. Do not omit any.
In the implementation milestone, Header() and Footer() must return
non-empty format-valid strings. The empty stub is only valid in M0.

Do not write a single generic render function with a type switch
covering only some cases. Write one typed function per scope.
A type switch that covers only known cases silently drops unknown ones.

---

## Static binary

The cli-tool template requires BINARY-TYPE: static. Set CGO_ENABLED=0
in every build target:

```makefile
build:
	CGO_ENABLED=0 go build -o $(BINARY_NAME) .
```

Also in the RPM spec %build section and in the Containerfile builder stage.

---

## M0 compile gate commands

```bash
go mod tidy
go build ./...
file ./<binary>              # must output "statically linked"
./<binary> version           # must print version and exit 0
./<binary> help              # must print usage and exit 0
./<binary> format=bad_value  # must exit 2 (invocation error)
```

For components requiring elevated privileges, collection tests are
deferred to human verification. Only the non-privileged paths above
are verifiable during translation.

---

## format="" and format="all" normalisation

When a spec declares format="" and format="all" as equivalent,
normalise in prepareConfig() immediately after parsing:

```go
if config.Format == "all" {
    config.Format = ""
}
```

Then in render():
```go
if config.Format == "" {
    // render all active formats
}
```

---

## Output path construction

Always use filepath.Join, never string concatenation:

```go
outpath = filepath.Join(config.Outdir, binaryName+"-"+hostname+ext)
```

Create the output directory before opening any file:
```go
if err := os.MkdirAll(filepath.Dir(outpath), 0755); err != nil {
    fmt.Fprintf(os.Stderr, "error creating directory: %v\n", err)
    os.Exit(1)
}
```

---

## Privilege check placement

Place privilege checks in the collection function, not in main().
This ensures help, version, and invocation-error paths work without
elevated privileges:

```go
func collect(config *Config) (*Manifest, error) {
    if os.Geteuid() != 0 {
        fmt.Fprintf(os.Stderr, "Please run as root.\n")
        os.Exit(1)
    }
    // ...
}
```

---

## Signal handling

Implement in main() — safe to do in M0:

```go
sigChan := make(chan os.Signal, 1)
signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)
go func() {
    <-sigChan
    os.Exit(0)
}()
```

Required imports: os/signal, syscall.
