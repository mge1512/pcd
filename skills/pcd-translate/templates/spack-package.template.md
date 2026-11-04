# spack-package

## META

Deployment:   spack-package
Version:      0.1.0
Spec-Schema:  0.3.22
Author:       Matthias G. Eckermann <pcd@mailbox.org>
License:      GPL-2.0-only
Verification: none
Safety-Level: QM

---

## SPACK-VERSION

This template was derived from Spack source code version **1.1.1**
(tarball `spack-1.1.1.tar.gz`, released 2025).

Primary source files analyzed:

| File (relative to spack-1.1.1/) | Purpose |
|---|---|
| `lib/spack/spack/directives.py` | Authoritative directive signatures: `version`, `variant`, `depends_on`, `conflicts`, `patch`, `provides`, `license`, `requires`, `maintainers` |
| `lib/spack/spack/audit.py` | Complete `spack audit` check inventory (all `@package_directives`, `@package_properties`, `@package_attributes`, `@package_https_directives` checks) |
| `lib/spack/spack/cmd/audit.py` | CLI interface to audit groups |
| `lib/spack/docs/build_systems/cmakepackage.rst` | CMakePackage phases, `cmake_args`, `define_from_variant` usage |
| `share/spack/templates/mock-repository/package.pyt` | Official Spack new-package skeleton |

When Spack is updated, re-examine `directives.py` for new directive arguments,
and `audit.py` for new `@package_directives` / `@package_properties` check
functions. Bump the `SPACK-VERSION` block and the `Version:` field in META.

---

## PURPOSE

A `spack-package` spec describes a single Spack package recipe (`package.py`).
The primary deliverable is a Python class inheriting from a Spack build system
base class (e.g. `CMakePackage`, `AutotoolsPackage`, `Package`) that declares
versions, variants, dependencies, and build logic in Spack's declarative DSL.

Spack packages are **specifications**, not general-purpose code. The translator
must not invent dependency names, version strings, variant values, or checksum
digests. Every such value must originate from the source repository's build
files (e.g. `CMakeLists.txt`, `configure.ac`, `meson.build`) or from explicit
spec instructions.

---

## TYPES

```
PackageName       := lowercase string matching [a-z0-9][a-z0-9_-]*
                     // Spack convention: all lowercase, hyphens preferred over underscores

ClassName         := PascalCase derived from PackageName
                     // e.g. "my-lib" -> "MyLib", "fxdiv" -> "Fxdiv"

BuildSystem       := CMakePackage | AutotoolsPackage | MesonPackage |
                     MakefilePackage | PythonPackage | Package
                     // Package = generic, use when no standard build system fits

VersionString     := string matching [0-9][0-9a-z._-]*
                     // e.g. "1.0", "2.3.1", "0.9.0-rc1"

SHA256Digest      := hex string of exactly 64 characters
                     // Only sha256 is accepted; md5/sha1/sha224 are forbidden

DepType           := "build" | "link" | "run" | "test" | tuple thereof
                     // default when omitted: ("build", "link")
                     // cmake dep type is always "build"

VariantName       := lowercase string matching [a-z][a-z0-9_-]*
                     // must not be in RESERVED_NAMES: arch, compiler, etc.

VariantDefault    := bool | non-empty-string
                     // bool for boolean variants; string for multi-valued
                     // empty string and None are forbidden (unparsable from CLI)

WhenSpec          := Spack spec string constraining when a directive applies
                     // e.g. "@2.0:", "+cuda", "platform=linux"
                     // for transitive deps use "^depname+variant" not "depname+variant"

SPDXIdentifier    := string  // e.g. "MIT", "Apache-2.0", "GPL-2.0-only"

PatchDigest       := SHA256Digest  // patches must use sha256

FetchSource       := UrlFetch | GitFetch
UrlFetch          := { url: string, sha256: SHA256Digest }
GitFetch          := { git: string, commit: string }
                     // tag or branch allowed but commit preferred for reproducibility
```

---

## INTERFACES

```
Primary deliverable:
  package.py          Spack package recipe; location in repository:
                        var/spack/repos/<reponame>/packages/<pkgname>/package.py
                      or in standalone repo:
                        packages/<pkgname>/package.py

Validation commands (must all exit 0 before spec is considered correct):
  spack style --fix <pkg>       syntax and style (flake8 + isort)
  spack audit packages <pkg>    directive correctness (see INVARIANTS)
  spack spec <pkg>              parse + concretize with default variants
  spack install --fake <pkg>    dry-run install (dependency resolution only)

```

---

## DEPENDENCIES

```
Runtime (inherited from BuildSystem base class — do not re-declare):
  cmake@3.15:          type=build    (CMakePackage provides this automatically)
  c                    type=build    (declared via build_system("c") or inherits)
  cxx                  type=build    (declared via build_system("cxx") if needed)

Translator must declare:
  All library dependencies found in CMakeLists.txt find_package() calls,
  configure.ac PKG_CHECK_MODULES / AC_CHECK_LIB calls, or meson.build
  dependency() calls, with correct type= (build / link / run / test).

Do-not-fabricate rule:
  If a dependency name is not verifiable in the source repository metadata,
  it must not be declared. Use a comment "# TODO: verify <dep>" instead.
```

---

## BEHAVIOR: version

Declare one or more released versions with fetch source and checksum.

```
INPUTS:
  ver:        VersionString        version label
  url:        string               tarball URL (for UrlFetch)
  sha256:     SHA256Digest         checksum of tarball (for UrlFetch)
  git:        string               repository URL (for GitFetch)
  commit:     string               full commit hash (for GitFetch)
  preferred:  bool (optional)      mark as preferred version
  deprecated: bool (optional)      mark as deprecated

PRECONDITIONS:
  - ver is unique within this package
  - exactly one of (url, git) is provided per version
  - if url is provided then sha256 is provided
  - if git is provided then commit, tag, or branch is provided
  - sha256 is exactly 64 hex characters (no md5, sha1, sha224, sha512 accepted)

STEPS:
  1. For each released version found in source repository or upstream release page:
     a. Record version string exactly as published
     b. Locate the canonical tarball URL
     c. Compute or obtain the sha256 digest of the tarball
     d. Declare version(ver, sha256=digest, url=url)
  2. Mark the recommended stable release with preferred=True if multiple
     major release series are present
  3. For EOL versions: add deprecated=True, do not remove them

POSTCONDITIONS:
  - at least one version is declared
  - every version with a URL has a sha256 digest
  - no version uses md5, sha1, or sha224 checksums

ERRORS:
  - version string does not follow semver-like pattern: emit comment
    "# NOTE: upstream version string '<ver>' is non-standard"
  - sha256 cannot be obtained: do not declare the version; add TODO comment
```

---

## BEHAVIOR: variant

Declare a user-selectable build option.

```
INPUTS:
  name:        VariantName
  default:     VariantDefault
  description: string (non-empty)
  values:      tuple of strings (optional, for multi-valued variants)
  multi:       bool (optional, default False)
  when:        WhenSpec (optional)
  sticky:      bool (optional, default False)

PRECONDITIONS:
  - name is not in Spack RESERVED_NAMES
  - default is bool or non-empty string
  - description is non-empty (required by spack audit)
  - if values is provided, default must be a member of values
  - name matches [a-z][a-z0-9_-]*

STEPS:
  1. For each configurable option found in source build files:
     a. Choose a lowercase hyphen-separated name
     b. Determine the correct default (match upstream cmake option default)
     c. Write a concise description (imperative form: "Enable X support")
     d. For options with a fixed set of values, declare values=(...)
     e. Add when= if the option is only meaningful for certain versions or
        other variant combinations

POSTCONDITIONS:
  - every variant has a non-empty description
  - every variant default is parsable from CLI (not None, not empty string)
  - variant names do not shadow Spack reserved names

ERRORS:
  - upstream option has no clear default: use default=False for bool variants
  - upstream option name cannot be mapped to valid VariantName:
    transliterate to closest lowercase-hyphen form, add comment
```

---

## BEHAVIOR: depends_on

Declare a dependency on another Spack package.

```
INPUTS:
  spec:    Spack spec string  e.g. "hdf5@1.8:", "mpi", "python@3.8:"
  type:    DepType            (default: ("build", "link"))
  when:    WhenSpec           (optional)
  patches: list (optional)    patches to apply to the dependency

PRECONDITIONS:
  - dependency package name exists in the Spack repository
    (verify with: spack info <depname>)
  - version constraint, if given, is satisfiable by at least one
    known version of the dependency
  - virtual packages (mpi, blas, lapack, etc.) must not have variants
    in the spec string
  - self-referential when= conditions are forbidden:
    bad:  depends_on("foo@1.0", when="^foo+bar")
    good: depends_on("foo@1.0", when="+myvariant")
  - when= references transitive deps with ^ prefix:
    bad:  when="depname+variant"
    good: when="^depname+variant"

STEPS:
  1. For each find_package / PKG_CHECK_MODULES / dependency() call in source:
     a. Map to Spack package name (consult spack list or builtin packages)
     b. Extract minimum version requirement if expressed in build files
     c. Assign correct type: build-only tools get type="build";
        linked libraries get type=("build","link") (default);
        runtime-only deps get type="run"
     d. Add when= if dependency is conditional on a variant or version
  2. Do not declare cmake as a dependency for CMakePackage (inherited)
  3. Do not declare c/cxx/fortran as depends_on; use build_system() directive

POSTCONDITIONS:
  - no unknown package names in depends_on directives
  - no virtual deps with variants in spec
  - no self-referential when= conditions
  - version constraints satisfied by known repo versions

ERRORS:
  - dependency name not found in spack repository:
    add comment "# TODO: package '<depname>' not yet in Spack"
    do not declare the depends_on
  - version constraint unsatisfiable: loosen constraint or add TODO comment
```

---

## BEHAVIOR: conflicts

Declare mutually exclusive variant or platform combinations.

```
INPUTS:
  conflict_spec:  Spack spec string  the conflicting condition
  when:           WhenSpec (optional) trigger condition
  msg:            string (optional)   human-readable explanation

PRECONDITIONS:
  - all variant names in conflict_spec exist in this package
  - all variant values are valid for their respective variants
  - msg is provided when the conflict reason is not self-evident

STEPS:
  1. For each CMAKE_CONFLICT / mutually exclusive option in source:
     a. Express as conflicts("+varA", when="+varB", msg="...")
  2. For platform restrictions express as:
     conflicts("+cuda", when="platform=darwin", msg="CUDA not supported on macOS")

POSTCONDITIONS:
  - conflict_spec references only declared variants
  - when= references only declared variants or known platform/os/target values

ERRORS:
  - unknown variant in conflict: do not declare, add TODO comment
```

---

## BEHAVIOR: cmake_args

Map variants and configuration to CMake -D flags (CMakePackage only).

```
INPUTS:
  self.spec    current concrete spec (variants, version, dependencies)

PRECONDITIONS:
  - method is only present when BuildSystem is CMakePackage
  - define_from_variant() is used for every boolean variant that maps
    directly to a CMake option
  - raw string flags ("-DFOO=bar") are used only when no variant exists
    for the option

STEPS:
  1. For each cmake option in CMakeLists.txt that has a corresponding variant:
     a. Use self.define_from_variant("CMAKE_OPTION_NAME", "variant-name")
  2. For cmake options with fixed values (not user-selectable):
     a. Use self.define("CMAKE_OPTION_NAME", value)
  3. For options depending on dependency presence:
     a. Use self.define("ENABLE_FOO", self.spec.satisfies("+foo"))
  4. Return the list of args

POSTCONDITIONS:
  - every variant declared in BEHAVIOR: variant has a corresponding
    define_from_variant() call, or an explicit comment explaining why it
    does not map to a cmake flag
  - no hard-coded paths or system-specific values in args
  - CMAKE_INSTALL_PREFIX is never set here (Spack sets it automatically)

ERRORS:
  - cmake option name not found in CMakeLists.txt: add TODO comment
```

---

## BEHAVIOR: patch

Apply a source patch with checksum verification.

```
INPUTS:
  filename_or_url:  string          local filename or HTTPS URL of patch
  sha256:           PatchDigest     sha256 of the patch file
  when:             WhenSpec        version or variant condition
  working_dir:      string          relative path to apply patch in (optional)
  level:            int             patch -p level (default 1)

PRECONDITIONS:
  - sha256 is provided for every patch (required by spack audit)
  - local patches are placed in the package directory alongside package.py
  - URL patches use https (required by spack audit)
  - when= is specified to scope patches to affected versions

STEPS:
  1. For each upstream bugfix or portability patch required:
     a. Obtain patch file and compute sha256
     b. Prefer upstream commit URLs over local copies
     c. Scope with when="@X.Y" to the affected version range
  2. For portability fixes not upstream: place .patch file in package dir

POSTCONDITIONS:
  - every patch has a sha256 checksum
  - patch URLs use https not http

ERRORS:
  - sha256 cannot be computed: do not declare patch, add TODO comment
```

---

## BEHAVIOR: metadata

Declare package-level metadata: homepage, description, license, maintainers.

```
INPUTS:
  homepage:     string (HTTPS URL)
  url:          string (HTTPS URL to tarball, or set per-version)
  git:          string (HTTPS URL to git repo, if git-based)
  license:      SPDXIdentifier
  maintainers:  list of GitHub usernames (optional)

PRECONDITIONS:
  - class has a docstring (required by spack audit)
  - docstring does not contain FIXME, boilerplate, or "example.com"
  - homepage uses https not http
  - license is a valid SPDX identifier

STEPS:
  1. Set homepage to the canonical project URL (https)
  2. Set url to a representative tarball URL, or set per version() if URLs
     vary in structure across versions
  3. Set git if the package is git-only or git-primary
  4. Add license("SPDX-ID") directive
  5. Write a one-sentence docstring summarising what the package does
  6. Optionally add maintainers(["github-user"]) for active maintainers

POSTCONDITIONS:
  - class has a non-empty docstring
  - docstring contains no FIXME, "remove this boilerplate", "FIXME: Put",
    "FIXME: Add", or "example.com" strings
  - homepage is present and uses https
  - license directive is present with valid SPDX identifier

ERRORS:
  - SPDX identifier unclear from source: use "Unknown" and add TODO comment
  - homepage not findable: use git URL as fallback
```

---

## INVARIANTS

All invariants below correspond directly to `spack audit` checks implemented
in `lib/spack/spack/audit.py` (Spack 1.1.1). The check function name is noted
for traceability.

### Supply chain and checksums

```
[invariant]  Every version() with a URL has sha256=; no md5=, sha1=, sha224=
             is ever used.
             // audit.py: _ensure_all_packages_use_sha256_checksums
             // Rationale: sha256 is the only hash accepted by spack audit;
             //   md5 and sha1 are cryptographically broken; sha224 is non-standard.
             //   Relevant for FIPS environments and supply-chain security.

[invariant]  Every patch() has sha256=; patch URLs use https not http.
             // audit.py: _ensure_all_packages_use_sha256_checksums,
             //           _check_patch_urls, _linting_package_file

[invariant]  All URLs in version(), patch(), and homepage use https.
             // audit.py: _linting_package_file
```

### Structural correctness

```
[invariant]  The package class has a docstring.
             // audit.py: _ensure_docstring_and_no_fixme

[invariant]  The docstring contains none of: "remove this boilerplate",
             "FIXME: Put", "FIXME: Add", "example.com".
             // audit.py: _ensure_docstring_and_no_fixme

[invariant]  The package class name is PascalCase derived from the package
             name; the package name (directory) is all lowercase.
             // audit.py: _ensure_all_package_names_are_lowercase

[invariant]  No reserved Spack attribute names are used as class attributes.
             // audit.py: _search_for_reserved_attributes_names_in_packages

[invariant]  No deprecated Spack methods (setup_*_environment moving to
             builder classes) are present in the package class.
             // audit.py: _search_for_deprecated_package_methods,
             //           _ensure_env_methods_are_ported_to_builders
```

### Variant correctness

```
[invariant]  Every variant() has a non-empty description= argument.
             // audit.py: _ensure_variants_have_descriptions

[invariant]  Every variant default is parsable from the CLI:
             must be bool or non-empty string, not None or "".
             // audit.py: _ensure_variant_defaults_are_parsable
             // directives.py: variant() raises DirectiveError on empty default

[invariant]  Variant names do not appear in Spack's RESERVED_NAMES list.
             // directives.py: variant() raises DirectiveError on reserved name
```

### Dependency correctness

```
[invariant]  Every package name in depends_on() exists in the Spack repository.
             No dependency names are invented or hallucinated.
             // audit.py: _issues_in_depends_on_directive

[invariant]  Version constraints in depends_on() are satisfiable by at least
             one known version of the dependency in the Spack repository.
             // audit.py: _version_constraints_are_satisfiable_by_some_version_in_repo

[invariant]  Virtual packages (mpi, blas, lapack, scalapack, etc.) must not
             have variants in the depends_on spec string.
             // audit.py: _issues_in_depends_on_directive (check_virtual_with_variants)

[invariant]  depends_on() does not create self-referential conditions:
             depends_on("foo@X", when="^foo+bar") is forbidden.
             // audit.py: _issues_in_depends_on_directive (problematic_edges check)

[invariant]  Variant names used in depends_on() when= or spec arguments
             must exist in the referenced package and accept the given values.
             // audit.py: _unknown_variants_in_directives,
             //           _issues_in_depends_on_directive (variant prevalidation)
```

### Directive when= correctness

```
[invariant]  when= arguments in depends_on(), variant(), provides(), patch(),
             resource() must not use bare package names as conditions.
             Transitive dependency conditions must use ^ prefix:
               correct: when="^depname+variant"
               wrong:   when="depname+variant"
             // audit.py: _named_specs_in_when_arguments

[invariant]  All variant names referenced in conflicts(), provides(),
             resource() directives exist in this package.
             // audit.py: _unknown_variants_in_directives
```

---

## EXAMPLES

### EXAMPLE: minimal_cmake_package

A minimal correct CMake package with one version and one boolean variant.

```
GIVEN:
  pkgname    = "fxdiv"
  upstream   = github.com/Maratyszcza/FXdiv
  buildsys   = CMake
  variants   = [tests: bool default=False, benchmarks: bool default=False,
                inline-assembly: bool default=False]
  deps       = [cmake@3.5: build, c build, cxx build]
  license    = MIT

WHEN:
  translator generates package.py

THEN:
  from spack_repo.builtin.build_systems.cmake import CMakePackage
  from spack.package import *

  class Fxdiv(CMakePackage):
      """Fixed-point division library for C/C++."""

      homepage = "https://github.com/Maratyszcza/FXdiv"
      git      = "https://github.com/Maratyszcza/FXdiv.git"

      license("MIT")
      maintainers("mge1512")

      version("main", branch="main")
      version("2021-09-13", commit="b408327ac2a526c03c8b9974a4d7c4fb8d4c7fd2",
              sha256="...")

      variant("inline-assembly", default=False,
              description="Enable inline assembly optimizations")
      variant("tests",      default=False, description="Build test suite")
      variant("benchmarks", default=False, description="Build benchmarks")

      depends_on("cmake@3.5:", type="build")
      depends_on("c",          type="build")
      depends_on("cxx",        type="build")

      def cmake_args(self):
          return [
              self.define_from_variant("FXDIV_USE_INLINE_ASSEMBLY", "inline-assembly"),
              self.define_from_variant("FXDIV_BUILD_TESTS",         "tests"),
              self.define_from_variant("FXDIV_BUILD_BENCHMARKS",    "benchmarks"),
          ]

NOTE:
  - spack audit passes: docstring present, sha256 used, variants have
    descriptions, no unknown deps, variant defaults parsable
  - define_from_variant covers all three variants (no untranslated options)
  - cmake dependency not re-declared (CMakePackage inherits it)
```

### EXAMPLE: conditional_dependency

A dependency that is conditional on a variant.

```
GIVEN:
  package has variant("hdf5", default=False, description="Enable HDF5 I/O support")
  HDF5 is only a dependency when +hdf5

WHEN:
  translator generates depends_on for HDF5

THEN:
  variant("hdf5", default=False, description="Enable HDF5 I/O support")
  depends_on("hdf5@1.8:", when="+hdf5")

NOT:
  depends_on("hdf5@1.8:", when="hdf5=True")   # wrong: use +variant not key=value
  depends_on("hdf5@1.8:", when="hdf5+True")   # wrong: no such syntax
```

### EXAMPLE: virtual_dependency

An MPI-parallel package requiring MPI.

```
GIVEN:
  package links against MPI

WHEN:
  translator declares MPI dependency

THEN:
  depends_on("mpi")

NOT:
  depends_on("mpi+cuda")    # wrong: virtual packages cannot have variants
  depends_on("openmpi")     # wrong: pin to virtual not concrete provider
```

### EXAMPLE: audit_failure_unknown_dep

A hallucinated dependency that spack audit would catch.

```
GIVEN:
  translator invents a dependency "libfast" not present in Spack repository

WHEN:
  spack audit packages mypkg

THEN:
  error: "mypkg: unknown package 'libfast' in 'depends_on' directive"

CORRECT BEHAVIOR:
  Do not declare depends_on("libfast").
  Instead add comment:
    # TODO: 'libfast' not found in Spack repository — verify dep name
```

### EXAMPLE: audit_failure_missing_description

A variant without description that spack audit would catch.

```
GIVEN:
  variant("cuda", default=False)   # no description

WHEN:
  spack audit packages mypkg

THEN:
  error: "Variant 'cuda' in package 'mypkg' is missing a description"

CORRECT BEHAVIOR:
  variant("cuda", default=False, description="Enable CUDA GPU acceleration")
```

### EXAMPLE: audit_failure_bad_checksum

A version using md5 instead of sha256.

```
GIVEN:
  version("1.0", md5="d41d8cd98f00b204e9800998ecf8427e")

WHEN:
  spack audit packages mypkg

THEN:
  error: "Package 'mypkg' does not use sha256 checksum"

CORRECT BEHAVIOR:
  version("1.0", sha256="e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855")
```

### EXAMPLE: when_spec_transitive_wrong

Incorrect use of when= for a transitive dependency condition.

```
GIVEN:
  package conditionally depends on "hdf5" when another dependency "petsc"
  is built with HDF5 support

WHEN:
  translator writes when= condition

THEN — WRONG:
  depends_on("hdf5", when="petsc+hdf5")   # bare name, not transitive form

THEN — CORRECT:
  depends_on("hdf5", when="^petsc+hdf5")  # ^ prefix for transitive dep
```

---

## MILESTONES

Milestone-based delivery is optional for single-file deliverables. Use milestones
when the package has more than 15 variants or complex conditional dependency graphs
that benefit from phased validation.

### MILESTONE: 0.0.0 — Scaffold
```
Scaffold: true
Constraint: required-for-release

Deliverable: package.py that imports correctly and defines the class.
  Load check: spack python -c "import importlib; importlib.import_module('...package')"
  No variants, no dependencies required — stubs acceptable.

Acceptance:
  spack style <pkg>       exits 0
  spack spec <pkg>        exits 0  (concretizes with defaults)
```

### MILESTONE: 0.1.0 — Versions + metadata
```
Constraint: required-for-release

Deliverable: All released versions declared with sha256; docstring present;
  homepage set; license declared.

Acceptance:
  spack audit packages <pkg>    exits 0 (no checksum or docstring errors)
  spack install --fake <pkg>    exits 0 (at least one version installs dry-run)
```

### MILESTONE: 0.2.0 — Variants + dependencies
```
Constraint: required-for-release

Deliverable: All variants declared with descriptions; all library dependencies
  declared with correct types; cmake_args covers all variants.

Acceptance:
  spack audit packages <pkg>    exits 0 (no variant or dep errors)
  spack spec <pkg> +<variant>   exits 0 for each boolean variant
  spack install --fake <pkg>    exits 0
```

### MILESTONE: 0.3.0 — Full validation
```
Constraint: required-for-release

Deliverable: All audit checks pass; patches declared if needed.

Acceptance:
  spack audit packages <pkg>         exits 0
  spack audit packages --all <pkg>   exits 0
  spack install <pkg>                exits 0 (real install in clean env)
```

---

## HINTS

```
// Spack-specific anti-patterns to avoid:

- Do not use setup_run_environment() or setup_build_environment() in the
  Package class body; port to CMakeBuilder subclass or use use_cmake_prefix_path.
  (spack audit catches this: _ensure_env_methods_are_ported_to_builders)

- Do not set CMAKE_INSTALL_PREFIX in cmake_args; Spack sets it automatically.

- For packages that provide multiple virtual packages, declare all of them:
    provides("blas")
    provides("lapack")
  Virtual package names must be in Spack's virtual package registry.

- The "c" and "cxx" dependencies are compiler language declarations, not
  package names. Use depends_on("c", type="build") — this is correct Spack 1.x
  syntax. Do not use depends_on("gcc") or depends_on("g++").

- version() with git= and commit= is preferred over branch= for reproducibility.
  If using branch=, also provide a sha256 via sha256= on a specific tag.

- OpenMP is not a Spack package; it is handled via compiler flags.
  Do not declare depends_on("openmp"). Use:
    variant("openmp", default=False, description="Enable OpenMP parallelism")
    and in cmake_args: self.define_from_variant("ENABLE_OPENMP", "openmp")

- For CUDA support, inherit from CudaPackage in addition to CMakePackage:
    class MyPkg(CMakePackage, CudaPackage):
  CudaPackage provides the cuda_arch variant automatically.

- For ROCm/HIP support, inherit from ROCmPackage:
    class MyPkg(CMakePackage, ROCmPackage):

- patch() must always include sha256=; never use patch() without it.
  (spack audit: _ensure_all_packages_use_sha256_checksums checks patch resources)

- The "spack create" command can scaffold a package from a URL. Use it as a
  starting point, then apply this spec to add variants and audit compliance.
```
