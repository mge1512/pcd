# The specification

## Spec sections

::::columns
:::: {.column width=45%}

![](pcdp-spec-sections.png){height=6cm}

::::
:::: {.column width=55%}

`pcdp-lint` validates structure\
before any translation begins.

\bigskip

**Key constraints:**

- BEHAVIOR must include `STEPS:`
- INVARIANTS carry `[observable]`\
  or `[implementation]` tags
- EXAMPLES must cover error paths
- No programming language in spec

::::
::::

---

## BEHAVIOR in depth

::::columns
:::: {.column width=52%}

```markdown
## BEHAVIOR: shutdown
Constraint: required
INPUTS:
  vm: VirtualMachine
PRECONDITIONS:
  - vm.state = Running
STEPS:
  1. Send shutdown signal.
  2. Record shutdown-start time.
     MECHANISM: do NOT block;
     requeue immediately.
  3. On re-entry: check elapsed.
     If > timeout: Destroy(vm).
POSTCONDITIONS:
  - vm.state = Stopped
```

::::
:::: {.column width=48%}

`STEPS` = the algorithm,\
not just the contract.

\bigskip

`MECHANISM:` = where the\
*how* matters for correctness.

\bigskip

`Constraint:` field:

- `required` --- always implement
- `supported` --- preset activates it
- `forbidden` --- never implement

::::
::::

---

## INTERFACES: testable abstractions

```markdown
## INTERFACES

Transport {
  required-methods:
    Connect(ctx, spec) -> (Session, error)
  implementations-required:
    production:  RealTransport
    test-double: FakeTransport {
      configurable: ConnectErr, Sessions
    }
}
```

\bigskip

Independent tests use **only** declared test doubles.\
`go test ./independent_tests/` runs without any live service.

---

# Templates and translation

## Deployment templates

::::columns
:::: {.column width=50%}

Templates encode **everything** the spec does not:

- target language + alternatives
- binary type, packaging formats
- CLI conventions
- security constraints
- TYPE-BINDINGS
- delivery phases (EXECUTION)
- compile gate

::::
:::: {.column width=50%}

![](pcdp-resolution.png){height=3cm}

\bigskip

One preset file changes\
language for the whole project.

The spec is untouched.

::::
::::

---

## TYPE-BINDINGS: no translator discretion

Template excerpt:

```markdown
## TYPE-BINDINGS
| Spec type  | LANGUAGE=Go           |
|------------|-----------------------|
| Duration   | metav1.Duration       |
| Timestamp  | metav1.Time           |
| Condition  | metav1.Condition      |
| List<T>    | []T                   |
```

\bigskip

Without TYPE-BINDINGS: one translator produces `string`,\
another produces `metav1.Duration` --- incompatible at the API level.

\bigskip

With TYPE-BINDINGS: **deterministic, language-specific types** --- applied\
mechanically, not by translator judgement.

---

## The translation phases

::::columns
:::: {.column width=40%}

![](pcdp-translation-phases.png){height=5.5cm}

::::
:::: {.column width=60%}

Defined in the template's\
`## EXECUTION` section ---\
**not** in the prompt.

\bigskip

**Resume logic:** partial runs\
restart from the first missing file.

\bigskip

**Compile gate** is language-specific:

```bash
go mod tidy   # direct deps only
              # no fabricated versions
go build ./...
```

\bigskip

`TRANSLATION_REPORT.md` is\
**always last** --- its presence\
signals a complete run.

::::
::::

---

# Tooling

## pcdp-lint

::::columns
:::: {.column width=55%}

**What it validates:**

- required sections present
- META fields + SPDX license
- deployment template resolves
- BEHAVIOR has STEPS
- INVARIANTS have tags
- EXAMPLES cover error paths
- Constraint: field values
- cross-section consistency

::::
:::: {.column width=45%}

**Self-hosting:**

`pcdp-lint` was specified and\
generated using PCDP itself.

\bigskip

Running `pcdp-lint` against\
its own spec passes --- the\
paradigm validates itself.

\bigskip

```bash
pcdp-lint myspec.md
pcdp-lint strict=true myspec.md
pcdp-lint list-templates
```

::::
::::

---

## Translation confidence

Every `TRANSLATION_REPORT.md` includes:

\bigskip

+----------------+------------+-------------------------------------+--------------------+
| EXAMPLE        | Confidence | Verification method                 | Unverified claims  |
+================+============+=====================================+====================+
| vm-start       | High       | `TestVMStart` --- FakeSession         | libvirt UUID format|
+----------------+------------+-------------------------------------+--------------------+
| graceful-stop  | Medium     | `TestGracefulStop` pass 1 only      | timeout path       |
+----------------+------------+-------------------------------------+--------------------+
| host-ready     | Low        | none                                | entire sequence    |
+----------------+------------+-------------------------------------+--------------------+

\bigskip

- **High** = named test in `independent_tests/` passes without live services
- **Medium** = partial test coverage
- **Low** = reasoning only --- no test

\bigskip

Unverified claims must be listed **explicitly**. Never silently omitted.

---

## Large projects

::::columns
:::: {.column width=55%}

**Architect defines interfaces first:**

```markdown
## INTERFACE
EXPORTS:
  TYPES:  Account, TransferResult
  BEHAVIOR: transfer
  INVARIANTS:
    - GLOBAL: balance >= 0
```

\bigskip

Component specs declare imports:

```markdown
Imports:
  - account-svc: ./account.md
                 #INTERFACE@>=1.2.0
```

::::
:::: {.column width=45%}

**Rules:**

- interfaces before implementation
- build order = topological sort
- no circular dependencies
- component fits in one LLM\
  context window

\bigskip

**pcdp-lint v2** (planned):\
validates full project graph ---\
imports, circular deps,\
interface versioning.

::::
::::

---

## mcp-server-pcdp

::::columns
:::: {.column width=45%}

![](pcdp-mcp-server.png){height=4cm}

::::
:::: {.column width=55%}

Any MCP-capable host connects directly:

- **mcphost** (CLI)
- **VS Code** with Copilot
- **Claude Desktop**
- custom agents

\bigskip

Preset layering server-side:\
vendor \textrightarrow{} `/etc/pcdp-server/` \textrightarrow{} team \textrightarrow{} user

\bigskip

The LLM is the wizard.\
The server is the data layer.

::::
::::
