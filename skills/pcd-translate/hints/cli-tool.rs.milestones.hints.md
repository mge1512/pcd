# Hints: cli-tool scaffold-first milestones — Rust implementation

Template:  cli-tool
Language:  Rust
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
files, never add new types, never restructure modules. If M0 is
correct, every later milestone is a focused fill-in exercise.

---

## File layout principle

Split source into logically cohesive modules so that each milestone
touches at most two or three files. Rust modules map naturally to files.

Minimum required files for any cli-tool:

```
src/
├── main.rs          — entry point, argument parsing, main()
├── types.rs         — all struct and enum definitions
├── interfaces.rs    — all trait definitions + production and test-double
│                      implementations
└── lib.rs           — (optional) re-exports for test access
Cargo.toml           — dependencies; see below
```

Larger components should split by domain concern following the spec's
MILESTONE groupings: BEHAVIORs in the same milestone belong in the
same module file.

---

## Cargo.toml — required dependencies

```toml
[package]
name    = "component-name"
version = "0.1.0"
edition = "2021"

[dependencies]
serde      = { version = "1", features = ["derive"] }
serde_json = "1"

[profile.release]
strip = true      # smaller binary
```

For static linking (required by cli-tool template):
```toml
[target.x86_64-unknown-linux-gnu]
rustflags = ["-C", "target-feature=+crt-static"]
```

Or add to `.cargo/config.toml`:
```toml
[target.x86_64-unknown-linux-gnu]
rustflags = ["-C", "target-feature=+crt-static"]
```

This links the Rust standard library statically against glibc.
No additional rustup target required — x86_64-unknown-linux-gnu is
the default host target on Linux.

---

## M0 stub convention

Every stub function must:

1. Have the **correct signature** matching what its caller expects,
   including correct lifetime annotations where required.

2. Return the **correct zero value** for its declared output type.
   Use `Default::default()` where the type implements Default, or
   construct an empty-but-valid value explicitly:

   ```rust
   // CORRECT — empty but valid scope
   ScopeWrapper {
       attributes: std::collections::HashMap::new(),
       elements: Vec::new(),
   }

   // WRONG — None serialises to null in JSON
   None::<ScopeWrapper<MyRecord>>
   ```

3. Be **silent at normal verbosity**. Use a debug macro:
   ```rust
   macro_rules! debug_log {
       ($msg:expr) => {
           if std::env::var("APP_DEBUG").as_deref() == Ok("1") {
               eprintln!("DEBUG: {}", $msg);
           }
       };
   }
   ```
   Replace APP_DEBUG with the component's own env var name.

4. **Compile cleanly** — no unused imports, no dead code warnings
   unless suppressed with `#[allow(dead_code)]` on the stub.

For stub functions that return `Result`:
```rust
fn collect_something(&self) -> Result<MyScope, Box<dyn std::error::Error>> {
    debug_log!("collect_something: not yet implemented");
    Ok(MyScope::default())
}
```

---

## OSCommandRunner — must NOT be a stub

Implement it in full in M0. This is the most critical M0 requirement:

```rust
use std::process::Command;

pub struct OSCommandRunner;

impl CommandRunner for OSCommandRunner {
    fn run(&self, cmd: &str, args: &[&str]) -> Result<(String, String), Box<dyn std::error::Error>> {
        let output = Command::new(cmd)
            .args(args)
            .env("PATH", "/sbin:/bin:/usr/bin:/usr/sbin")
            .output()?;
        let stdout = String::from_utf8_lossy(&output.stdout).into_owned();
        let stderr = String::from_utf8_lossy(&output.stderr).into_owned();
        if output.status.success() {
            Ok((stdout, stderr))
        } else {
            Err(format!("{} failed: {}", cmd, stderr).into())
        }
    }
}
```

Note: Rust's `Command::env()` sets the variable only for the child
process — it does not affect the parent. This is cleaner than the Go
pattern and requires no cleanup.

A stub returning `Ok(("".into(), "".into()))` causes every
command-dependent module to silently produce empty output.

---

## JSON serialisation with serde

All types that appear in JSON output must derive `Serialize` and
`Deserialize`. Use `rename` or `rename_all` to produce underscore_style
keys matching the spec:

```rust
use serde::{Deserialize, Serialize};

#[derive(Serialize, Deserialize, Debug, Default, Clone)]
pub struct MyMeta {
    pub format_version: u32,
    pub collected_at:   String,
    // Field names already snake_case — no rename needed
}
```

For fields where the Go/JSON name differs from a natural Rust name:
```rust
#[derive(Serialize, Deserialize)]
pub struct MyRecord {
    #[serde(rename = "vendor_id")]
    pub vendor_id: String,
}
```

For structs with `_attributes` and `_elements` keys (Machinery pattern):
```rust
#[derive(Serialize, Deserialize, Debug, Default)]
pub struct ScopeWrapper<T> {
    #[serde(rename = "_attributes")]
    pub attributes: std::collections::HashMap<String, serde_json::Value>,
    #[serde(rename = "_elements")]
    pub elements: Vec<T>,
}
```

Without explicit field names or rename_all, serde uses the Rust field
name as-is. Since Rust already uses snake_case, this usually works —
but verify against the spec's JSON schema.

---

## ScopeWrapper pattern

The spec's `ScopeWrapper<T>` maps directly to a Rust generic struct:

```rust
#[derive(Serialize, Deserialize, Debug, Clone)]
pub struct ScopeWrapper<T> {
    #[serde(rename = "_attributes")]
    pub attributes: std::collections::HashMap<String, serde_json::Value>,
    #[serde(rename = "_elements")]
    pub elements: Vec<T>,
}

impl<T> Default for ScopeWrapper<T> {
    fn default() -> Self {
        Self {
            attributes: std::collections::HashMap::new(),
            elements:   Vec::new(),
        }
    }
}
```

Always initialise to `ScopeWrapper::default()` in stubs, never to
`None`. JSON serialisation of `None` produces `null`; an empty
ScopeWrapper produces `{"_attributes":{},"_elements":[]}`.

---

## Trait definitions for interfaces

The spec's interface declarations map to Rust traits:

```rust
pub trait Filesystem: Send + Sync {
    fn read_file(&self, path: &str) -> Result<String, std::io::Error>;
    fn exists(&self, path: &str) -> bool;
    fn is_executable(&self, path: &str) -> bool;
    fn glob(&self, pattern: &str) -> Result<Vec<String>, Box<dyn std::error::Error>>;
}

pub trait CommandRunner: Send + Sync {
    fn run(&self, cmd: &str, args: &[&str])
        -> Result<(String, String), Box<dyn std::error::Error>>;
}
```

Add `Send + Sync` bounds when traits are used across threads. For
single-threaded tools this is optional but good practice.

---

## Test doubles

Rust test doubles use `cfg(test)` or a separate module:

```rust
#[cfg(test)]
pub struct FakeCommandRunner {
    pub responses: std::collections::HashMap<String, (String, String)>,
}

#[cfg(test)]
impl CommandRunner for FakeCommandRunner {
    fn run(&self, cmd: &str, _args: &[&str])
        -> Result<(String, String), Box<dyn std::error::Error>>
    {
        self.responses
            .get(cmd)
            .cloned()
            .map(Ok)
            .unwrap_or_else(|| Ok((String::new(), String::new())))
    }
}
```

Or use `mockall` crate if the preset activates it. Without mockall,
manual test doubles as above are sufficient.

---

## Renderer trait stubs for M0

```rust
pub trait Renderer {
    fn header(&self, manifest: &Manifest) -> String;
    fn toc(&self, sections: &[String]) -> String;
    fn section(&self, title: &str, level: u8, content: &str) -> String;
    fn footer(&self) -> String;
    fn escape(&self, raw: &str) -> String;
}

pub struct HtmlRenderer;
impl Renderer for HtmlRenderer {
    fn header(&self, _: &Manifest) -> String  { String::new() }
    fn toc(&self, _: &[String]) -> String     { String::new() }
    fn section(&self, _: &str, _: u8, _: &str) -> String { String::new() }
    fn footer(&self) -> String                { String::new() }
    fn escape(&self, raw: &str) -> String     { raw.to_owned() }
}
// Repeat for TexRenderer, DocBookRenderer, MarkdownRenderer, JsonRenderer
```

Do not write a single generic render function matching on types.
Write one typed function per scope (see component-specific hints).
In M0 all render functions return `String::new()`.

---

## Static binary — glibc static linking

The cli-tool template requires `BINARY-TYPE: static`.

```bash
# No extra target needed. Set in .cargo/config.toml:
# [target.x86_64-unknown-linux-gnu]
# rustflags = ["-C", "target-feature=+crt-static"]

# Build static binary:
cargo build --release

# Verify:
file target/release/<binary>
# must output "statically linked"
```

Set in Makefile:
```makefile
build:
	cargo build --release

install:
	install -m 755 target/release/$(BINARY) $(DESTDIR)/usr/bin/
```

In the RPM spec `%build` section:
```
cargo build --release
```

---

## M0 compile gate commands

```bash
cargo build                  # debug build; must exit 0
cargo build --release          # static release build (glibc, crt-static)
file target/release/<binary>
  # must output "statically linked"
./<binary> version           # must print version and exit 0
./<binary> help              # must print usage and exit 0
./<binary> format=bad_value  # must exit 2 (invocation error)
cargo test                   # all unit tests must pass
```

Note: `cargo build` is significantly slower than `go build` on first
compile due to dependency compilation. Subsequent builds with `--incremental`
(default) are faster. Allow 2-5 minutes for M0 first build.

---

## format="" and format="all" normalisation

```rust
if config.format.as_deref() == Some("all") {
    config.format = None;  // or Some("".into()) depending on your type
}
```

Then in render:
```rust
match config.format.as_deref() {
    None | Some("") => { /* render all formats */ }
    Some(fmt)       => { /* render single format */ }
}
```

---

## Output path construction

Use `std::path::PathBuf`:

```rust
use std::path::PathBuf;

let mut outpath = PathBuf::from(&config.outdir);
outpath.push(format!("{}-{}{}", binary_name, hostname, ext));

// Create parent directory:
if let Some(parent) = outpath.parent() {
    std::fs::create_dir_all(parent)?;
}
```

Never concatenate paths with string `+` or `format!`. `PathBuf`
handles trailing slashes, relative paths, and OS separators correctly.

---

## Privilege check placement

Place in the collection function, not in `main()`:

```rust
fn collect(config: &Config) -> Result<Manifest, Box<dyn std::error::Error>> {
    if unsafe { libc::geteuid() } != 0 {
        eprintln!("Please run as root.");
        std::process::exit(1);
    }
    // ...
}
```

Or use the `nix` crate:
```rust
use nix::unistd::Uid;
if !Uid::effective().is_root() {
    eprintln!("Please run as root.");
    std::process::exit(1);
}
```

Add to Cargo.toml if using nix: `nix = { version = "0.27", features = ["user"] }`

---

## Signal handling

```rust
use std::sync::atomic::{AtomicBool, Ordering};
use std::sync::Arc;

let running = Arc::new(AtomicBool::new(true));
let r = running.clone();
ctrlc::set_handler(move || {
    r.store(false, Ordering::SeqCst);
    std::process::exit(0);
}).expect("Error setting signal handler");
```

Or more simply, using the `signal-hook` crate:
```toml
signal-hook = "0.3"
```

```rust
signal_hook::flag::register(signal_hook::consts::SIGTERM, Arc::clone(&term))?;
signal_hook::flag::register(signal_hook::consts::SIGINT,  Arc::clone(&term))?;
```

For M0 stubs, Rust's default SIGTERM/SIGINT behaviour (process exit)
is acceptable. Explicit signal handling can be added in a later milestone.

---

## Error handling pattern

Use `Box<dyn std::error::Error>` for collection functions that call
external commands. This avoids complex error type definitions in M0:

```rust
fn collect_cpu(fs: &dyn Filesystem, cr: &dyn CommandRunner)
    -> Result<CpuScope, Box<dyn std::error::Error>>
{
    // stub
    Ok(CpuScope::default())
}
```

In later milestones, replace with a proper error enum if needed.
For M0, `Box<dyn Error>` compiles cleanly and defers the error
type decision.

---

## Compile-time warnings to suppress in M0

M0 will generate many dead code and unused variable warnings.
Suppress at the crate level in `main.rs` or `lib.rs`:

```rust
#![allow(dead_code)]
#![allow(unused_variables)]
#![allow(unused_imports)]
```

Remove these attributes progressively as milestones are implemented.
By M7 all suppressions should be gone.

---

## Compile time expectation

| Pass | Expected time (first build) | Expected time (incremental) |
|---|---|---|
| M0 (debug)   | 2–5 min | 10–30 sec |
| M0 (release) | 3–8 min | 30–90 sec |
| M1–M7        | 10–30 sec incremental | same |

The first `cargo build` compiles all dependencies. Subsequent builds
with `--incremental` recompile only changed files. This is much slower
than Go on first build but comparable on incremental.
ENDOFFILE