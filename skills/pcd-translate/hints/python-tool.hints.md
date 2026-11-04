# hints/python-tool.hints.md
# Hints-For: python-tool
# Version: 0.1.0

Practical hints for LLMs translating PCD specs into Python tools.
Style: brief explanation + code snippet. No padding.

---

## flake8 + black coexistence

black and flake8 disagree on two rules by default. Always add a
`.flake8` or `[tool.flake8]` section to suppress them:

```ini
# .flake8
[flake8]
max-line-length = 88
extend-ignore = E203, W503
```

`E203` — whitespace before `:` in slices (black formats this way).
`W503` — line break before binary operator (black prefers this).
Without these, `flake8` will flag valid black output on every run.

Do NOT add `E501` to ignore — let black enforce line length at 88.
If a line exceeds 88 after black, it is a string literal or comment
that needs manual wrapping.

---

## mypy strict — most frequent errors and fixes

**Missing return type on a function:**
```python
# error: Function is missing a return type annotation
def build_parser():  ...

# fix:
def build_parser() -> argparse.ArgumentParser: ...
```

**Optional not handled:**
```python
# error: Item "None" of "str | None" has no attribute "upper"
value = os.environ.get("KEY")
print(value.upper())

# fix:
value = os.environ.get("KEY")
if value is not None:
    print(value.upper())
```

**Dict / list without type parameter:**
```python
# error: Need type annotation for "items"
items = {}

# fix:
items: dict[str, int] = {}
```

**Any in strict mode:**
```python
# error: Returning Any from function declared to return "str"
def get_name(data: dict[str, Any]) -> str:
    return data["name"]  # Any

# fix: cast or narrow
from typing import cast
def get_name(data: dict[str, Any]) -> str:
    return cast(str, data["name"])
```

**argparse.Namespace in strict mode** — Namespace attributes are untyped.
Create a typed dataclass or TypedDict and convert immediately in `main()`:

```python
from dataclasses import dataclass

@dataclass
class Args:
    verbose: bool
    input: str

def parse_args() -> Args:
    parser = build_parser()
    raw = parser.parse_args()
    return Args(verbose=raw.verbose, input=raw.input)
```

Then pass `Args` (not `Namespace`) to `run()` — mypy can check it fully.

---

## argparse patterns

**Subcommands:**
```python
subparsers = parser.add_subparsers(dest="command", required=True)

run_parser = subparsers.add_parser("run", help="Run the tool")
run_parser.add_argument("--input", required=True)

check_parser = subparsers.add_parser("check", help="Validate only")
check_parser.add_argument("--strict", action="store_true")
```

Dispatch in `main()`:
```python
if args.command == "run":
    run(args)
elif args.command == "check":
    check(args)
```

**Mutually exclusive group:**
```python
group = parser.add_mutually_exclusive_group()
group.add_argument("--verbose", action="store_true")
group.add_argument("--quiet", action="store_true")
```

**Required=True on a group** (at least one must be given):
```python
group = parser.add_mutually_exclusive_group(required=True)
```

**Path argument with existence check:**
```python
import pathlib

parser.add_argument(
    "--config",
    type=pathlib.Path,
    default=pathlib.Path("/etc/{TOOL_NAME}/config.toml"),
)
```
Check existence in business logic, not in argparse — better error messages.

---

## logging — configuration patterns

**Basic setup in `main()` — always to stderr:**
```python
logging.basicConfig(
    level=logging.DEBUG if args.verbose else logging.INFO,
    format="%(levelname)s %(name)s: %(message)s",
    stream=sys.stderr,
)
```

**Per-module logger — always use `__name__`:**
```python
logger = logging.getLogger(__name__)
logger.debug("processing file=%r", path)
```
Never use `logging.debug()` directly in modules — it writes to the root
logger and bypasses per-module level control.

**Suppress noisy third-party loggers:**
```python
logging.getLogger("urllib3").setLevel(logging.WARNING)
logging.getLogger("requests").setLevel(logging.WARNING)
```

**File handler alongside stderr:**
```python
fh = logging.FileHandler("/var/log/{TOOL_NAME}.log")
fh.setLevel(logging.DEBUG)
fh.setFormatter(logging.Formatter("%(asctime)s %(levelname)s %(name)s: %(message)s"))
logging.getLogger().addHandler(fh)
```
Add after `basicConfig()`, not before.

---

## pyproject.toml — common mistakes

**hatchling does not find the package:**
The `packages` path must match the actual directory under `src/`:
```toml
[tool.hatch.build.targets.wheel]
packages = ["src/{PACKAGE_NAME}"]
```
If TOOL_NAME is `my-tool`, PACKAGE_NAME must be `my_tool` (underscore).
hatchling does NOT auto-convert hyphens.

**`uv build` picks up wrong files:**
Add explicit excludes if tests or packaging dirs bleed into the wheel:
```toml
[tool.hatch.build]
exclude = ["tests/", "packaging/"]
```

**version not found at runtime:**
```python
# wrong — hardcoded, drifts from pyproject.toml
__version__ = "0.1.0"

# right — reads from package metadata at runtime
from importlib.metadata import version, PackageNotFoundError
try:
    __version__ = version("{TOOL_NAME}")
except PackageNotFoundError:
    __version__ = "0.0.0+unknown"
```

**dev dependencies in wrong table:**
`[dependency-groups]` (PEP 735, uv native) is correct for dev deps.
Do NOT put dev deps in `[project.optional-dependencies]` — that leaks
them into the published wheel metadata.

**license field syntax changed in PEP 639 (Python 3.12+):**
```toml
# old style (still works, use for 3.11 compat)
license = { text = "GPL-2.0-only" }

# new style (PEP 639, requires packaging >= 24)
license = "GPL-2.0-only"
```
Stick with `{ text = "..." }` until distro packaging tools catch up.

---

## uv — when things go wrong

**`uv sync` fails: no solution found**
Usually a Python version conflict. Check:
```sh
uv python list
uv sync --python 3.11
```

**Editable install not reflected:**
```sh
uv sync --reinstall-package {TOOL_NAME}
```

**Lock file out of date after editing pyproject.toml:**
```sh
uv lock
uv sync
```

**uv not available in RPM/DEB build environment:**
Do not use `uv` in `%build` or `dh` rules. Use stdlib `python3 -m build`
there. uv is a developer tool, not a build system dependency.

**`uv run` vs installed entry point:**
`uv run {TOOL_NAME}` uses the venv entry point.
`uv run python -m {PACKAGE_NAME}` uses `__main__.py`.
Both must work — test both before packaging.

---

## src/ layout — import errors

**`ModuleNotFoundError` when running tests directly:**
```sh
# wrong
python tests/test_core.py

# right
uv run pytest tests/
```
pytest adds `src/` to `sys.path` via the editable install. Running
scripts directly does not — always use `uv run pytest`.

**`ImportError` in RPM %check or CI:**
Ensure the package is installed (editable or wheel) before running pytest:
```sh
pip install -e . --no-deps
pytest tests/
```

**Circular imports with src/ layout:**
If `cli.py` imports from `core.py` and `core.py` imports from `cli.py`,
you have a design problem — not a src/ problem. Move shared types to a
`models.py` or `types.py` module that neither imports from the other.

**`__init__.py` should be nearly empty:**
Only `__version__` and public re-exports belong there. Do not put
business logic in `__init__.py` — it runs on every import.

---

## RPM noarch — python3_sitelib vs python3_sitearch

Use `python3_sitelib` for pure-Python packages (no C extensions).
Use `python3_sitearch` only if the package contains compiled `.so` files.

A python-tool built with `uv build` / `python3 -m build --wheel` and
declared `noarch` will always install to `python3_sitelib`. If you
accidentally use `python3_sitearch`, the files will not be found on
the target system.

```spec
# correct for pure Python
BuildArch:  noarch

%files
%{python3_sitelib}/{PACKAGE_NAME}/
%{python3_sitelib}/{PACKAGE_NAME}-%{version}.dist-info/
```

**`%{version}` vs `{VERSION}` in spec:**
Inside a spec file, always use RPM macros: `%{version}`, `%{name}`.
Literal `{VERSION}` is a PCD template variable — it is substituted
before the spec is used, then RPM macros take over.

**pip install --no-deps in %install:**
Always use `--no-deps`. Runtime dependencies are declared as RPM
`Requires:` and managed by zypper/dnf — not by pip inside the build root.

```spec
%install
pip install --no-deps --root %{buildroot} --prefix %{_prefix} \
    dist/%{name}-%{version}-py3-none-any.whl
```

---

## DEB dh-python — common pitfalls

**`dh_python3` not finding the package:**
Ensure `debian/rules` calls the python3 sequence:
```makefile
%:
	dh $@ --with python3
```

**Entry point not installed as executable:**
dh-python installs wheel entry points automatically when using
`pybuild`. If using a custom `%install`, copy the script manually:
```makefile
override_dh_auto_install:
	pip install --no-deps --root debian/{TOOL_NAME} \
	    dist/{TOOL_NAME}-*.whl
```

**`${python3:Depends}` is empty:**
Only populated when `dh_python3` runs during build. If missing,
add `python3-all` to `Build-Depends` and ensure `dh --with python3`.

**debian/copyright must list all .py SPDX headers:**
Use `licensecheck` or `scancode` to verify. The SPDX headers in every
`.py` file make this straightforward — tools can auto-generate
`debian/copyright` from them.

---

## OCI — pip inside container, layer caching

**Layer order matters for cache efficiency:**
```dockerfile
# Install deps first (changes rarely)
COPY dist/{TOOL_NAME}-{VERSION}-py3-none-any.whl .
RUN pip3 install --no-cache-dir {TOOL_NAME}-{VERSION}-py3-none-any.whl

# Copy source last (changes often) -- not needed if installing from wheel
```

When installing from a pre-built wheel, the entire app is in one layer.
This is correct — do not split it.

**`--no-cache-dir` is mandatory in Containerfile:**
Without it, pip writes its HTTP cache into the image layer, bloating
the image. Cache has no value in an OCI image.

**Do not install uv inside the container:**
uv is a developer tool. The container only needs pip3 (or no pip at all
if you copy the wheel contents manually with `pip install --target`).

**openSUSE Leap python package name:**
The package is `python311`, not `python3.11` or `python3`.
Verify against the current Leap release before pinning:
```dockerfile
RUN zypper --non-interactive install --no-recommends python311
```

**ENTRYPOINT vs CMD:**
```dockerfile
# correct -- tool is the entrypoint, arguments via CMD
ENTRYPOINT ["{TOOL_NAME}"]
CMD ["--help"]

# wrong -- wrapping in sh defeats exec signal handling
ENTRYPOINT ["sh", "-c", "{TOOL_NAME}"]
```
Always use exec form `["..."]`, never shell form `"..."` — shell form
means SIGTERM goes to sh, not to your process.
