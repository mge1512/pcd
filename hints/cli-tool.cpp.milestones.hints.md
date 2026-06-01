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
project(zypper-declarative CXX)

set(CMAKE_CXX_STANDARD 17)
set(CMAKE_CXX_STANDARD_REQUIRED ON)
set(CMAKE_CXX_EXTENSIONS OFF)

# Dependencies are DYNAMICALLY linked distro libraries (see component hints).
find_package(PkgConfig REQUIRED)
# pkg_check_modules(... ) / find_package(...) per the component decisions file.

add_executable(zypper-declarative
  src/main.cpp src/cli.cpp src/command_runner.cpp
  # ... one .cpp per behaviour group
)
target_include_directories(zypper-declarative PRIVATE src include)
# target_link_libraries(zypper-declarative PRIVATE ...) per component hints.

include(GNUInstallDirs)
install(TARGETS zypper-declarative RUNTIME DESTINATION ${CMAKE_INSTALL_BINDIR})
```

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

## Dynamic linking (NOT static)

This C++ tool links its dependencies DYNAMICALLY against the distribution's
supported shared libraries. Do not attempt a static binary, and do not vendor or
pin the dependencies: building against each service pack's own libzypp, libsnapper,
jsoncpp, and yaml-cpp via OBS is the supply-chain-correct approach, and static
linking would force vendoring (a liability under a signed-supply-chain posture).
The per-SP package then links the right soname.

(This is the deliberate difference from the Go sibling, which builds a single
static `CGO_ENABLED=0` binary. C++ here is dynamic by design.)

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
       const char* d = std::getenv("ZYPPER_DECLARATIVE_DEBUG");
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

Model the spec's data model as plain structs. The `_attributes`/`_elements`
Machinery idiom maps to a templated wrapper:

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

file build/zypper-declarative          # an ELF dynamically linked executable
ldd build/zypper-declarative           # shows libzypp, libsnapper, etc. (dynamic)
./build/zypper-declarative version     # prints version + spec hash, exit 0
./build/zypper-declarative help        # prints usage, exit 0
./build/zypper-declarative --version   # tolerated alias, identical output, exit 0
./build/zypper-declarative format=bad_value ; test $? -eq 2   # invocation error
```

Note the gate uses bare-word `version` and `help` (the canonical global commands),
with `--version` accepted as a tolerated alias. There is no static-binary check
here (this build is dynamic by design); instead confirm with `ldd` that the
expected shared libraries are linked.

Operations requiring root or a live SUSE host (real package, snapshot, and
transaction work) are deferred to on-target human verification; only the
unprivileged paths above are verifiable during translation.

---

## Warnings and standard

Build with `-Wall -Wextra` and treat the data-model and parsing code as warning-
clean. Use C++17 features (`std::optional`, `std::filesystem`, structured
bindings) freely; do not require C++20. This keeps the source portable across the
GCC 15 toolchain on both service packs without depending on C++20-only library
facilities.
