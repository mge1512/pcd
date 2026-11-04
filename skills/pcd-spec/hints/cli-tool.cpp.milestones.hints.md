# Hints: cli-tool scaffold-first milestones - C++ implementation

Template:  cli-tool
Language:  C++ (C++17)
Topic:     scaffold-first milestone pattern

These hints apply to any cli-tool specification that uses the scaffold-first
milestone pattern (Scaffold: true on M0). They are advisory; they cannot override
spec invariants. Read the spec and the deployment template first, then these
hints, before writing any code.

This is the C++ sibling of `cli-tool.go.milestones.hints.md` and
`cli-tool.rs.milestones.hints.md`. The structure mirrors them; the idioms are C++.

---

## The scaffold-first rule

M0 produces a complete, compilable skeleton. The only M0 acceptance criterion is a
clean configure + build (and the bare-word version/help/format gate below). No
real logic in M0. Later milestones fill in function bodies; they do not create new
files, add new types, or restructure the source tree. If M0 is right, every later
pass is a focused fill-in.

---

## File layout and build system (CMake)

Use CMake (the cli-tool template's C++ build system). Lay the source out so each
milestone touches at most two or three files; the spec's MILESTONE groupings are
the guide.

```
CMakeLists.txt          - project, C++17, find_package, target, install
cmake/                  - Find modules only if a dependency ships none
src/
  main.cpp              - entry point, dispatch wiring only
  cli.{hpp,cpp}         - dispatch, key=value parsing, global contract
  types.hpp             - all data-model structs and enums
  command_runner.{hpp,cpp} - OSCommandRunner (implemented in full at M0)
  <one pair per BEHAVIOR group>.{hpp,cpp}
include/                - public headers if any are installed
tests/                  - black-box tests (invoke the built binary)
```

Minimal `CMakeLists.txt` shape:

```cmake
cmake_minimum_required(VERSION 3.20)   # 3.28 on SLE 15 SP7, 3.31 on SLE 16; 3.20 is a safe floor
project(<tool> CXX)

set(CMAKE_CXX_STANDARD 17)
set(CMAKE_CXX_STANDARD_REQUIRED ON)
set(CMAKE_CXX_EXTENSIONS OFF)

# Dependencies are DYNAMICALLY linked distro libraries (see the project decisions file).
find_package(PkgConfig REQUIRED)
# Prefer pkg_check_modules(... REQUIRED IMPORTED_TARGET ...) for the distro libs:
# their CMake configs can be fragile across service packs (some are Meson-generated
# and omit the package-config helper), whereas the .pc files are stable. The project
# decisions file names the specific libraries and the exact modules to request.

add_executable(<tool>
  src/main.cpp src/cli.cpp src/command_runner.cpp
  # ... one .cpp per behaviour group
)
target_include_directories(<tool> PRIVATE src include)
# target_link_libraries(<tool> PRIVATE ...) per component hints.

include(GNUInstallDirs)
install(TARGETS <tool> RUNTIME DESTINATION ${CMAKE_INSTALL_BINDIR})
```

A thin top-level `Makefile` wraps CMake and provides the conventional targets. It
is REQUIRED (not optional) and MUST define at least these phony targets: `build`
(configure + `cmake --build`, then copy the binary to the project root), `test`
(build, then compile and run the black-box tests), `man` (render the man page),
`clean`, and `dist`. The `dist` target produces the release source tarball:
`<tool>-$(VERSION).tar.gz` containing a single top-level directory
`<tool>-$(VERSION)/`, where `$(VERSION)` is read from a top-level `VERSION` file (so
it is identical to the RPM spec `Version:` and the embedded binary version), with
build artefacts (`build/`, VCS dirs) excluded. Use the name `dist` (autotools
heritage) deliberately, NOT `package`, which collides with CMake CPack's own
`package` target. The OUTPUT contract is fixed: a `<tool>-$(VERSION).tar.gz` whose
sole top-level entry is `<tool>-$(VERSION)/`, version sourced from the `VERSION`
file, so `rpmbuild`'s default `%setup`/`%autosetup` (which cd's into
`%{name}-%{version}/`) succeeds. The exact recipe may differ (a git-archive or a
file-list copy are both fine).

---

## Compiler selection (SLE 15 vs SLE 16)

The default system compiler differs by service pack, and this matters for C++17:

- **SLE 15 SP7:** the default `gcc-c++` is GCC 7, which is too old for clean C++17.
  Install and use the side-by-side `gcc15` / `gcc15-c++` packages (GCC 15.2) and
  point the build at them, e.g. `CXX=g++-15 cmake ...` or a CMake toolchain file
  setting `CMAKE_CXX_COMPILER=g++-15`. Do not build with the default GCC 7.
- **SLE 16.0:** the default `gcc-c++` is already GCC 15, so no special selection is
  needed.

State this in the build instructions and the RPM spec: on SLE 15, `BuildRequires:
gcc15-c++` and build with g++-15; on SLE 16, the default toolchain suffices.

---

## Build-system selection and external API feature detection

- **Build system: prefer CMake.** Meson is an acceptable option where the
  project context already uses it; autotools (automake) only in exceptional
  cases (e.g. integrating into an existing autotools tree). Record the choice
  in `TRANSLATION_REPORT.md`.
- **Feature-detect external library API versions in the build system, never by
  guessing header macros.** The build system already knows each library's
  version: CMake captures it from `pkg_check_modules`, Meson from
  `dependency()`. When the detected version crosses an API boundary, define a
  project compile definition (CMake:
  `target_compile_definitions(<target> PRIVATE <PROJ>_<FEATURE>)`) and branch
  in code with `#if defined(<PROJ>_<FEATURE>)`. This keeps the version logic
  in one place, makes both branches compilable, and survives library updates.
  Record the detected version and which branch was compiled in
  `TRANSLATION_REPORT.md`.
- If an operation has no stable API across the supported library versions,
  raise that in the report as a spec or hints concern; do not paper over it
  with runtime guesses.

---

## Dynamic linking (NOT static)

This C++ tool links its dependencies DYNAMICALLY against the distribution's
supported shared libraries. Do not attempt a static binary, and do not vendor or
pin the dependencies: building against each service pack's own shared libraries
via OBS is the supply-chain-correct approach, and static linking would force
vendoring (a liability under a signed-supply-chain posture). The per-SP package
then links the right soname. The project decisions file names the specific
libraries this tool links and any per-SP soname or API differences among them.

(This is the deliberate difference from the Go sibling, which builds a single
static `CGO_ENABLED=0` binary. C++ here is dynamic by design.)

---

## SLE BuildRequires naming (devel vs runtime packages)

- `BuildRequires:` names development packages, which on SLE 15 AND 16 follow
  the `<name>-devel` convention with NO `lib` prefix: the jsoncpp devel
  package is `jsoncpp-devel` (NOT `libjsoncpp-devel`), the yaml-cpp devel
  package is `yaml-cpp-devel` (NOT `libyaml-cpp-devel`).
- The `lib<name><soname>` packages with soname digits (e.g. `libyaml-cpp0_8`)
  are the RUNTIME shared-library packages; they never belong in
  `BuildRequires:`. The RPM shared-library dependency generator derives the
  runtime `Requires:` from the linked sonames automatically.
- A build that fails to resolve a `lib<name>-devel` BuildRequires is almost
  always this naming defect, not a missing repository.

---

## M0 stub convention

Every stub must:

1. Have the correct signature its caller expects (correct parameter and return
   types, `const` correctness, reference vs value as the callers need). Wrong
   signatures break later milestones at compile time.
2. Return a correct empty-but-valid value. For a scope, return a default-constructed
   wrapper whose containers are empty, never a null pointer:

   ```cpp
   // CORRECT - empty but valid; serialises to {"_attributes":{},"_elements":[]}
   ScopeWrapper<MyRecord> scope;   // members default-construct to empty
   return scope;

   // WRONG - a null/optional-empty that serialises to null and breaks consumers
   ```

3. Be silent at normal verbosity. Gate any stub trace behind a debug env var:

   ```cpp
   inline void debug_log(const std::string& msg) {
       const char* d = std::getenv("<TOOL>_DEBUG");
       if (d && std::string(d) == "1")
           std::cerr << "DEBUG: " << msg << "\n";
   }
   ```

4. Compile cleanly with `-Wall -Wextra`; suppress unavoidable M0 unused-parameter
   warnings narrowly (e.g. `[[maybe_unused]]`), not by disabling the warning
   globally.

---

## OSCommandRunner - must NOT be a stub

A real cli-tool drives a few external commands. Even when most integration is via
linked libraries, some operations have no library API and are done by executing a
command. Implement the runner in full at M0 (a stub that returns empty output
makes every command-dependent path silently empty while the build passes - the
most common scaffold-first failure):

```cpp
// command_runner.hpp
struct CommandResult { std::string out; std::string err; int code; };

class CommandRunner {
public:
    virtual ~CommandRunner() = default;
    virtual CommandResult run(const std::string& cmd,
                              const std::vector<std::string>& args) const = 0;
};

class OSCommandRunner : public CommandRunner {
public:
    CommandResult run(const std::string& cmd,
                      const std::vector<std::string>& args) const override;
};
```

Implement `OSCommandRunner::run` with `fork`/`execvp` and pipes (capturing stdout
and stderr separately and the exit status), setting a fixed `PATH`
(`/sbin:/bin:/usr/bin:/usr/sbin`) in the child only. Do not use `std::system`
(it goes through a shell and does not separate streams or give a clean exit code).
A non-zero exit is returned in `code`, not thrown, because some tools report
"differences found" with a non-zero exit that the caller must interpret as data,
not failure.

---

## Data-model types and the ScopeWrapper pattern

Model the spec's data model as plain structs. A spec that wraps each scope as an
`_attributes` map plus an `_elements` list (with underscore_style serialised keys)
maps to a templated wrapper:

```cpp
template <class T>
struct ScopeWrapper {
    std::map<std::string, std::string> attributes;  // "_attributes"
    std::vector<T> elements;                          // "_elements"
};
```

Serialisation field names must be the spec's underscore_style keys, regardless of
the C++ member names; the JSON/YAML component hints specify how each library
emits `_attributes` and `_elements`. Initialise every scope empty-but-valid in
stubs (an empty wrapper serialises to `{"_attributes":{},"_elements":[]}`, a null
does not).

Where the spec distinguishes an ABSENT scope from a PRESENT-but-empty one
(unmanaged vs reconcile-to-empty), represent absence with `std::optional<Scope>`:
`std::nullopt` = absent, a present `ScopeWrapper` with empty `elements` =
present-empty. Do not collapse the two.

---

## Interfaces and test doubles

Where the spec declares a seam (the command runner, a clock, a filesystem
reader), express it as an abstract base class with a production implementation and
a test double, so unit tests need no live system:

```cpp
class FakeCommandRunner : public CommandRunner {
public:
    std::map<std::string, CommandResult> responses;
    CommandResult run(const std::string& cmd,
                      const std::vector<std::string>&) const override {
        auto it = responses.find(cmd);
        return it != responses.end() ? it->second : CommandResult{"", "", 0};
    }
};
```

The black-box acceptance tests (tests/) invoke the built binary itself and assert
on stdout, stderr, and exit code; they do not link the internals.

The test harness's OWN command-runner (the helper the black-box tests use to launch
the binary and capture its output) MUST be hermetic and robust; this is a required
contract, not left to the author's discretion:
- Capture stdout and stderr PER INVOCATION into unique temporary files (`mkstemp`)
  or a per-test temporary directory, and clean them up. NEVER write to shared fixed
  paths such as `/tmp/out` and `/tmp/err`: shared paths race and cross-contaminate
  across tests, break re-runs, and make the suite non-parallelizable.
- Do NOT launch the binary through a shell (`std::system`/`popen` with a command
  string): that exposes a shell-quoting and injection surface for any argument
  containing spaces or special characters. Use `posix_spawn`, or `fork`+`execvp`
  with an argv vector, passing arguments as a vector, not a concatenated string.
- If capturing via pipes rather than temp files, drain stdout AND stderr
  CONCURRENTLY (poll/select or two threads); a sequential read of one then the other
  deadlocks when the binary fills the other pipe's buffer (a real failure mode on a
  command with large output). Temp files avoid this deadlock and are the simpler
  choice; either is acceptable as long as it is per-invocation and shell-free.
- Run the binary as the build user, never via `sudo`; an interactive sudo prompt
  hangs the suite. If the tool reads privileged system state, a test should use the
  tool's own non-fatal mode for unreadable sources where one exists (the project
  decisions file gives the specific option), so a protected source does not abort a
  test.

A "simplified" command-runner that uses `std::system` with shared `/tmp/out` and
`/tmp/err` is NOT acceptable even though it compiles; it is the exact shortcut to
avoid.

### Every test must assert the EXAMPLE's actual outcome (no hollow tests)

A test generated for an `### EXAMPLE:` MUST assert that example's THEN conditions,
the specific stdout/stderr content, exit code, file or scope state the example
declares, NOT merely that the binary exited 0. A test whose only assertion is
`exit_code == 0` is FORBIDDEN: it is "green theater", it passes even when the
behaviour under test is completely broken, and it is worse than no test because it
manufactures false confidence. Do NOT emit "fallback" or "auto-generated generic"
test bodies; that pattern (a single exit-code check standing in for a real
behavioural assertion) is exactly what must not happen, and a suite where most
tests are such stubs is a failed test-authoring run, not a passing one.

If an EXAMPLE needs a FIXTURE to be testable, BUILD the fixture; do not fall back
to a bare run. The example's GIVEN tells you what to construct: a manifest file (a
specific JSON/YAML document), an applied-record baseline, a synthetic root or `/etc`
subtree, or specific filesystem objects (a symlink, a special file, a ghost). Create
them in a per-test temporary directory, point the tool at them (via the relevant
`*-path` / root option the spec defines), run the verb, and assert the THEN. Offline
modes (comparing a manifest against a captured state dump, where the spec provides
them) let many behaviours be asserted without a live system; prefer them for
fixture-based tests.

If a behaviour GENUINELY cannot be exercised at test time, because it requires a
live transactional root, real snapshot creation, or root privilege the build user
lacks, the test MUST be explicitly marked deferred/skipped and counted as such (the
report distinguishes tested from deferred), NOT emitted as a passing exit-0 stub. An
honest "skipped: needs a live transactional target" is correct; a fake-green stub is
not. Mark such a case clearly (e.g. print a SKIP line and do not count it as a
pass), so the suite's green status reflects only behaviours actually verified.

---

## Clean up after yourself (test-author and translator)

Both roles MUST leave no stray artifacts behind. A run that litters the build host
or the source tree with temporary files, scratch directories, or half-written
outputs is not a clean run, even if it compiles and passes.

- The TEST-AUTHOR: every test creates its fixtures (temp files, scratch dirs,
  manifests, synthetic roots) in a per-test temporary location and DELETES them
  when the test finishes, including on the failure path (use RAII or an explicit
  cleanup at every return). Capture temp files (`mkstemp`) are `unlink`ed after use.
  No fixture is written to a fixed shared path or left in `/tmp`, the source tree,
  or the working directory after the suite exits. The suite must be re-runnable any
  number of times with no residue and no dependence on leftovers from a prior run.
- The TRANSLATOR: do not leave scratch files, generated intermediates, editor
  backups, or `build/` contents in the source tree as deliverables. Build output
  goes under the build directory and is excluded from `make dist` (the tarball
  contains only sources, not artifacts). Temporary files created during generation
  or self-checks are removed before the run is considered done. The committed tree
  is the sources plus the generated code, nothing extraneous.

The test of a clean run: after it finishes, `git status` shows only intended files,
and a second identical run starts from the same clean state and behaves identically.

---

## Output path construction

Use `std::filesystem`, never string concatenation:

```cpp
#include <filesystem>
namespace fs = std::filesystem;

fs::path outpath = fs::path(outdir) / (binary_name + "-" + hostname + ext);
std::error_code ec;
fs::create_directories(outpath.parent_path(), ec);   // before opening
```

`std::filesystem` handles separators and relative paths; concatenation does not.

---

## Privilege check placement

Place privilege checks in the operation that needs them, not in `main()`, so that
version, help, and invocation-error paths work unprivileged:

```cpp
if (::geteuid() != 0) {
    std::cerr << "this operation requires root\n";
    return 2;   // map to the spec's exit code for the verb
}
```

(Prefer returning a mapped exit code over `exit()` deep in a call; only the verb
layer decides the process exit status.)

---

## Signal handling

Install handlers in `main()` (safe at M0). Keep the handler async-signal-safe: set
a flag, or for the simple "clean exit" case use `_exit`:

```cpp
#include <csignal>
static volatile std::sig_atomic_t g_stop = 0;
extern "C" void on_signal(int) { g_stop = 1; }
// in main():
std::signal(SIGTERM, on_signal);
std::signal(SIGINT,  on_signal);
```

For a tool whose only requirement is "clean exit on SIGTERM/SIGINT", the default
termination is acceptable at M0; explicit handling for an interruptible long
operation (e.g. discarding a transaction) is added in the milestone that
implements that operation.

---

## M0 compile gate commands

```bash
# On SLE 15: select g++-15 first (CXX=g++-15). On SLE 16: default is GCC 15.
cmake -S . -B build -DCMAKE_BUILD_TYPE=Release
cmake --build build -j

file build/<tool>          # an ELF dynamically linked executable
ldd build/<tool>           # shows the project's linked distro shared libs (dynamic)
./build/<tool> version     # prints version + spec hash, exit 0
./build/<tool> help        # prints usage, exit 0
./build/<tool> --version   # tolerated alias, identical output, exit 0
./build/<tool> <bad-option> ; test $? -eq 2   # invocation error (exit 2)
```

Note the gate uses bare-word `version` and `help` (the canonical global commands),
with `--version` accepted as a tolerated alias. There is no static-binary check
here (this build is dynamic by design); instead confirm with `ldd` that the
expected shared libraries are linked.

Operations that MUTATE the system or genuinely require root are deferred to
on-target human verification; the unprivileged paths above, and any READ-ONLY
system queries, are verifiable during translation and must not be deferred. Do not
treat an empty result from a read-only query you chose not to run as a real absence
of data.

---

## Warnings and standard

Build with `-Wall -Wextra` and treat the data-model and parsing code as warning-
clean. Use C++17 features (`std::optional`, `std::filesystem`, structured
bindings) freely; do not require C++20. This keeps the source portable across the
GCC 15 toolchain on both service packs without depending on C++20-only library
facilities.
