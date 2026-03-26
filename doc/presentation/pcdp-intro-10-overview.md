# What

## Human Intent, Machine Implementation

::::columns
:::: {.column width=60%}

Domain experts write **specifications**.

AI generates **all** implementation code.

Engineers never write implementation code directly.

\medskip

\textit{Not vibe coding:} if the output is wrong,\
you fix the **spec** --- never the code.

::::
:::: {.column width=40%}

![](pcdp-logo-green.png){height=5cm}

::::
::::

---

## This is not "AI-assisted coding"

+---------------------+---------------+---------------+-------------------+
|                     | Traditional   | Vibe Coding   | **PCDP**          |
+=====================+===============+===============+===================+
| Human writes        | code          | code + prompts| **specs only**    |
+---------------------+---------------+---------------+-------------------+
| AI role             | none          | suggests      | **translates**    |
+---------------------+---------------+---------------+-------------------+
| Primary artifact    | source code   | source code   | **specification** |
+---------------------+---------------+---------------+-------------------+
| Target language     | developer     | developer     | **template**      |
+---------------------+---------------+---------------+-------------------+
| Regulated domains   | manual audit  | prohibited    | **enabled**       |
+---------------------+---------------+---------------+-------------------+

\bigskip

In vibe coding, if the output is wrong, you edit the code.\
In PCDP, **you never touch the code** --- you fix the spec and regenerate.

---

## The specification

::::columns
:::: {.column width=50%}

```markdown
## BEHAVIOR: transfer
INPUTS:
  from: Account
  amount: Amount
STEPS:
  1. Validate preconditions.
  2. Debit from.balance.
  3. Credit to.balance.
POSTCONDITIONS:
  - SUM(balances) unchanged
EXAMPLES:
  EXAMPLE: success
  GIVEN: balance = 100
  WHEN:  transfer(30)
  THEN:  balance = 70
```

::::
:::: {.column width=50%}

No programming language.

No target platform.

No toolchain knowledge required.

The **template** decides all of that.

::::
::::

# Why

## 1. Separation of concerns

::::columns
:::: {.column width=50%}

The spec says **what** and **where**.

The template decides **how**:

- programming language
- packaging format
- toolchain conventions
- deployment target

::::
:::: {.column width=50%}

Change language across the whole project?

**Change one preset file.**

The spec is untouched.

A spec written today is valid\ 
in 2045 --- regardless of\ 
what language is in fashion.

::::
::::

---

## 2. Long-term maintainability

::::columns
:::: {.column width=55%}

**Specs are more stable than code.**

Code accumulates technical debt.

Specs describe intent --- intent rarely changes.

\bigskip

When requirements change:

- update the spec
- regenerate the code
- no manual refactoring

::::
:::: {.column width=45%}

**The spec is the documentation.**

Not a comment that drifts from the code.

Not a wiki page nobody updates.

The running system is always\ 
derived from the spec.

::::
::::

---

## 3. Domain experts author directly

::::columns
:::: {.column width=50%}

**Today**

![](pcdp-workflow-today.png)

Every handoff loses information.\
Requirements get simplified.\
Misunderstandings go undetected.

::::
:::: {.column width=50%}

**With PCDP**

![](pcdp-workflow-pcdp.png)

The cardiologist specifies\
the device behaviour directly.

No interpreter in the middle.

::::
::::

---

## And: formal verification is available

The LLM translation is probabilistic --- we do not claim otherwise.

\bigskip

For components that **require** stronger guarantees:

\bigskip

::::columns
:::: {.column width=60%}

Lean 4, F*, or Dafny can be added\ 
as an **optional verification layer** ---\ 
without changing the specification format.

The spec stays the same.\ 
The verification path is a template choice.

::::
:::: {.column width=40%}

- memory safety by construction
- formal proofs of invariants
- state machine correctness
- ISO 26262 / DO-178C evidence

::::
::::

---

## Proof: pcdp-lint

`pcdp-lint` --- the specification validator --- was\ 
**specified and generated using PCDP itself.**

Zero hand-written implementation code.

\bigskip

Tested across multiple AI models,\ 
three continents, one result:

\bigskip

\begin{center}
\Large Every model resolved Go from the template\\
\large without being told.
\end{center}

---

# How

## The workflow

![](pcdp-workflow.png){height=3cm}

1. Domain expert writes a spec (or AI interviews them)
2. `pcdp-lint` validates structure
3. Deployment template resolves the language
4. LLM translates spec \textrightarrow{} code + audit bundle

---

## Language is never your problem

![](pcdp-resolution.png){height=3cm}

The spec declares **what** and **where** (deployment context).

The template resolves **language, packaging, conventions**.

A spec written today is valid if the organisation\
switches from Go to Rust in 2045 --- **no spec change needed**.

---

## The audit bundle

::::columns
:::: {.column width=50%}

Every translation produces:

- specification (human-reviewable)
- generated source code
- packaging artifacts (RPM, DEB, OCI)
- independent test suite
- translation report
- workflow diagram (Pikchr)
- metadata + hashes

::::
:::: {.column width=50%}

Designed for:

- ISO 26262 (automotive)
- DO-178C (aviation)
- IEC 62304 (medical)
- Common Criteria EAL4+
- EU Cyber Resilience Act

::::
::::

---

## Getting started

::::columns
:::: {.column width=50%}

**Write a spec (no format knowledge needed):**

```bash
# AI interviews the domain expert:
ollama run llama3.2 \
  "$(cat prompts/interview-prompt.md)"
```

Then validate:

```bash
pcdp-lint myspec.md
```

::::
:::: {.column width=50%}

**Translate to code:**

```bash
# mcphost with the translator prompt
# reads template + spec,
# produces code + audit bundle
```

\bigskip

Everything in the repo:\
**github.com/mge1512/pcdp**

::::
::::
