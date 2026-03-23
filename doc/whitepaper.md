# Post-Coding Development Paradigm
## Human Intent, Machine Implementation

**Status:** Draft  
**Version:** 0.3.8  
**Author:** Matthias G. Eckermann <pcdp@mailbox.org>  
**Date:** 2026-03-19

---

## Executive Summary

The **Post-Coding Development Paradigm** fundamentally changes how software is created: **domain experts write specifications in structured natural language; AI generates verified implementations**. Engineers never write implementation code. Instead, they author precise specifications in Markdown describing what a system should do—data types, behaviors, invariants, state machines, deployment context. An AI translator converts these specifications into type-safe, memory-safe implementations, optionally through formal verification in proven meta-languages (Lean 4, F*, Dafny).

**This is not "AI-assisted coding"** where developers write code with AI suggestions. This is **post-coding development** where domain experts write specifications and AI generates all implementation code. The human role shifts from programming to architectural specification.

The paradigm enables AI-augmented development in **safety-critical and regulated domains** (automotive, aviation, medical devices, finance) that currently prohibit AI code generation due to auditability requirements. Specifications remain human-reviewable; optional formal proofs provide certification evidence. The verifiable meta-language layer is **optional**—teams can choose direct code generation for rapid iteration or formal verification for high-assurance requirements.

**The project is open source and technology-agnostic.** The semantic core (when used) is designed to be pluggable. Rather than inventing proprietary formal systems, we leverage mature, proven verification technologies (Lean 4, F*, Dafny) as optional meta-languages. The stable specification format and intermediate representation ensure teams can choose verification technologies that match their expertise and regulatory requirements—or skip formal verification entirely for lower-risk components.

**Target language is not a specification concern.** A key design principle introduced in v0.3.0: the target programming language is a *function of the deployment context*, not a free variable that the spec author must decide. Deployment templates encode this mapping, separating architectural intent (the specification) from implementation mechanics (the generated code).

Deliverables include specification schemas, deployment templates, translator prototypes, pluggable IR formats, reference backends (C, Rust, Go, WASM), CI patterns producing audit bundles, and migration guidance for gradual adoption. The `pcdp-lint` validator is developed as the paradigm's own reference implementation—eating its own dog food from the first artifact.

### Value Proposition at a Glance

| Category | Traditional Coding | AI-Assisted Coding ("Vibe Coding") | Post-Coding Development |
|----------|-------------------|-------------------------------------|------------------------|
| Human writes | Implementation code | Code + prompts, iterates with AI | Specifications (never code) |
| AI role | None | Suggests/completes code | Translates specs → verified implementation |
| Primary artifact | Source code | Source code (AI-influenced) | Specification (code is generated) |
| Target language chosen by | Developer | Developer | Deployment template |
| Safety guarantees | Depends on language/testing | Same as traditional | Type-safe + memory-safe by construction (when using meta-language) |
| Auditability | Review code | Review AI-influenced code (opaque) | Review specifications + proofs |
| Regulatory compliance | Expensive manual audits | Prohibited (can't audit AI suggestions) | Enabled (auditable specs + optional proofs) |
| Domain expert role | Consults, doesn't code | Consults, doesn't code | **Authors specifications directly** |
| Maintainability | Code rot | Code rot + AI drift | Specifications remain stable |

**Bottom line:** Domain experts with architectural capacity write specifications describing system behavior and deployment context. AI generates all implementation code—either directly or through formal verification. This enables AI development in safety-critical domains while shifting engineering effort from coding to precise specification.

---

## 1. Introduction

AI has made code synthesis cheap, but unstructured generation is brittle and unsuitable for regulated environments. Traditional formal methods provide guarantees but require specialized expertise. The Post-Coding Development Paradigm bridges this gap through a fundamental shift: **humans write what the system should do (specifications); AI writes how to do it (implementation)**.

The paradigm is built on four core principles:

**1. Specifications, not code, as primary artifacts**  
Domain experts author structured Markdown specifications describing system behavior, data types, invariants, state machines, and deployment context (backend, embedded, kernel-driver, etc.). These specifications are written in natural language with tables—no programming required, no formal syntax required. Engineers with domain expertise and architectural capacity can write valid specifications without knowing the target programming language or meta-language.

**2. AI translates specifications to implementations**  
An AI translator converts specifications to executable code. Two paths are available:
- **Direct path:** Specification → target language (Go, C, Rust, etc.) for rapid iteration
- **Verified path:** Specification → meta-language (Lean 4, F*, Dafny) → target language for formal guarantees

The meta-language layer is **optional**. Teams choose based on risk, regulatory requirements, and timeline.

**3. Multiple LLMs ensure translation quality (optional)**  
For critical components, 2-3 independent AI translators generate separate implementations from the same specification. Cross-validation catches translation errors and specification ambiguities. For lower-risk components, single-LLM translation with testing is sufficient.

**4. Use proven meta-languages, don't invent new ones**  
Rather than creating proprietary formal systems, the paradigm leverages mature verification technologies (Lean 4 for theorem proving, F* for effect tracking, Dafny for SMT-based verification) as pluggable meta-languages. This provides immediate credibility, active communities, and allows teams to choose technologies they trust—or skip formal verification entirely.

**This is not "vibe coding."** In AI-assisted development, programmers write code and iterate with AI suggestions. In post-coding development, domain experts write specifications and never touch implementation code. If the generated code is incorrect, engineers refine the specification—not the code—and regenerate entirely.

**Target domains** include safety-critical systems (automotive, aviation, medical devices, industrial control), regulated environments (finance, government, healthcare), and any context where formal correctness, auditability, or certification is required. The workflow also benefits general engineering by separating intent (specifications) from implementation (generated code).

---

## 2. Goals

- **Enable AI development in safety-critical domains:** Solve the regulatory problem—current AI code generation cannot be used in automotive (ISO 26262), aviation (DO-178C), medical devices (IEC 62304), or other safety-critical domains because AI-generated code cannot be audited. By making specifications the reviewable artifact and providing optional formal proofs, enable AI-augmented development in these markets.

- **Shift engineering role from coding to specification:** Transform the job of software engineers from writing implementation code to authoring precise, architectural specifications. Domain experts with system knowledge become the primary authors.

- **Eliminate translation ambiguity:** Address the critical weakness—informal natural language specifications create ambiguity when translating to formal code. Use constrained specification format with formal notation, required sections, and executable examples to reduce translation variance by 80-90%.

- **Remove target language as a specification concern:** Target language selection is a deployment-time decision, not an authoring decision. Deployment templates encode the mapping from deployment context to target language, freeing spec authors from implementation choices.

- **Provide flexible verification paths:** Support both direct code generation (for rapid iteration) and formal verification through meta-languages (for high-assurance requirements). Teams choose the verification level appropriate to their risk and regulatory context.

- **Type and memory safety by construction:** When using the verified path, entire classes of bugs (null pointer dereferences, buffer overflows, use-after-free, data races) are prevented at compile-time through the meta-language's type system, not detected at runtime through testing.

- **Incremental adoption:** Enable pilots on small, high-value components (crypto primitives, state machines, drivers) and expand without rewriting existing systems.

- **Open and replaceable:** Publish specification schemas and tooling as open source. The meta-language is pluggable—teams choose Lean 4, F*, Dafny, or develop custom intermediate formats. The specification format and IR remain stable regardless of backend choice.

- **Auditability and certification:** Produce human-reviewable specifications and optional formal proofs suitable for regulatory audits and certification processes.

---

## 3. Tenets

- **Specifications are first-class artifacts:** Markdown specifications are the canonical source of truth, not implementation code. Specifications include deployment context and architectural decisions, not just functional requirements.

- **Domain experts write specifications:** The target user is a domain expert with architectural capacity—someone who understands the system's purpose, deployment environment, and safety/security requirements. They write specifications in structured natural language, not code.

- **AI translates, humans validate:** AI converts specifications to implementations (either directly or through meta-languages). Humans validate specifications, review proofs (if generated), and gate deployment. AI never autonomously deploys code.

- **Target language is not a human decision:** The spec author declares *what* and *where* (deployment context). The *target language* is derived automatically from the deployment template. This keeps specifications technology-agnostic and stable over time.

- **Verification is optional and pluggable:** Teams choose their verification path:
  - **No formal verification:** Spec → Go/C/Rust directly (fastest, lowest assurance)
  - **Formal verification:** Spec → Lean 4/F*/Dafny → Go/C/Rust (slower, highest assurance)
  - **Hybrid:** Formal verification for critical paths, direct generation for non-critical code

- **Use proven technologies, don't invent new ones:** Leverage existing mature meta-languages (Lean 4, F*, Dafny) rather than creating proprietary formal systems. This provides immediate ecosystem support and allows replacement as better technologies emerge.

- **Auditable outputs:** Generated code is intentionally simple and traceable to specifications. When formal verification is used, proofs and audit bundles provide certification evidence.

- **Incrementalism:** Adopt component by component. Specifications can describe interfaces to existing hand-written code. Mixed codebases (generated + manual) are explicitly supported.

---

## 4. State of the Business

Industry shows convergence of specification-driven development, AI code generation, and formal verification. Tool vendors explore spec-to-code pipelines; research demonstrates proof-carrying code extraction; enterprises adopt AI coding assistants but hesitate where auditability and certification matter.

**Critical gap:** Safety-critical and regulated industries (automotive, aviation, medical devices, finance, government) **cannot use** current AI code generation tools. Regulatory frameworks (ISO 26262, DO-178C, IEC 62304, Common Criteria) require auditable development processes. AI-suggested code is considered "opaque" and fails auditability requirements. This represents a massive market (automotive software alone is $50B+ annually) where AI productivity gains are currently prohibited.

Our positioning emphasizes three differentiators:

1. **Auditable specifications, not opaque code:** Regulatory audits review specifications (human-readable) and optional formal proofs (machine-checkable), not AI-generated code. This solves the certification blocker.

2. **Flexible verification:** Optional meta-language layer allows teams to choose verification level. Safety-critical components use formal verification; supporting infrastructure uses direct generation.

3. **Pluggable architecture:** Stable specification format and IR support multiple meta-languages and direct code generation. No vendor lock-in; teams choose verification technology or skip it entirely.

A concise competitor comparison is provided in Appendix A.3.

---

## 5. Lessons Learned

- **AI needs structured input:** Freeform prompts produce brittle, non-reproducible outputs. Structured specifications with explicit deployment context enable reliable translation.

- **Translation ambiguity is the critical weakness:** The gap between informal natural language and formal code is where hallucinations occur. Constrained specification formats with formal notation and executable examples reduce translation variance dramatically.

- **Meta-language compatibility matters:** If using formal verification, the meta-language must have sufficient LLM training data for reliable translation. Languages with small communities or unique syntax (e.g., ATS2) fail AI translation even if theoretically powerful.

- **Verification must be optional:** Requiring formal verification for all code blocks adoption. Teams need the flexibility to verify critical paths and generate non-critical code directly.

- **Domain context is essential:** Specifications without deployment context (backend vs. embedded vs. kernel-driver) are ambiguous. The paradigm requires architectural information, not just functional requirements.

- **Target language must not be a spec author's concern:** Early versions of this paradigm required spec authors to declare a target language. This was identified as an anti-pattern—it pulls authors back into implementation thinking. Target language is now derived entirely from deployment templates.

- **Executable examples serve as test oracle:** Specifications that include concrete input/output examples enable validation of translations—generated code must pass all examples or translation is rejected.

- **Openness builds trust:** Open specification formats, reproducible builds, and transparent verification toolchains are decisive for enterprise and regulatory adoption.

- **Small pilots validate approach:** Targeted pilots on well-defined components (crypto primitives, state machines) create momentum and reduce organizational risk.

- **Eat your own dog food:** The `pcdp-lint` validator is being developed using this paradigm itself. A specification that cannot describe its own tooling has not yet proven its expressiveness.

---

## 6. Strategic Priorities

**Pilot and validate**  
- Run pilots on well-scoped components (crypto primitives, register accessors, state machines, protocol implementations).
- Use `pcdp-lint` as the first reference implementation developed under the paradigm.
- Measure defect reduction, specification review time, audit effort, and certification cost reduction.
- For highest-assurance components, employ dual-LLM verification where independent translators cross-validate via formal equivalence.

**Specification format and tooling**  
- Define stable Markdown specification schema with required sections (Data Types, Behavior, Invariants, State Machine, Deployment Context, Security).
- Build linters and validators for specification quality.
- Provide deployment templates covering the standard deployment contexts (see Appendix A.11).
- Develop IDE support (VS Code extension, specification completions).

**Deployment template system**  
- Define the initial set of deployment templates with language defaults and constraints.
- Implement systemd-style preset layering (system defaults → org presets → user presets → project presets).
- Provide documented override mechanism for cases where template defaults are insufficient.

**Flexible verification paths**  
- Implement both direct code generation (spec → Go/C/Rust) and formal verification (spec → meta-language → Go/C/Rust) paths.
- Document when to use each approach (risk-based decision framework).
- Provide migration path: start with direct generation, add formal verification to critical components as needed.

**Pluggable meta-language support**  
- Define stable intermediate representation (IR) format that is meta-language agnostic.
- Provide reference implementations for multiple meta-languages (Lean 4 as initial reference, F* and Dafny adapters).
- Document meta-language adapter interface so teams can integrate preferred verification technology or custom IR formats.
- The IR is the contract; the meta-language is interchangeable.

**Tooling and CI integration**  
- Build CI steps producing audit bundles (specification + IR + generated code + proofs + metadata).
- Integrate with PR gating (specification changes trigger regeneration and verification).
- Provide diff tools showing specification changes and their impact on generated code.
- Support mixed codebases (specifications for new components, manual code for legacy).

**Regulatory and certification support**  
- Engage with safety standards bodies (ISO 26262, DO-178C, IEC 62304) to validate audit bundle format.
- Develop certification guidance documentation.
- Partner with certification authorities for pilot programs.
- Create traceability matrix template (specification requirements → proofs → generated code).

**Adoption and community**  
- Seed with demos, specification templates, and training materials.
- Build community hub for sharing specification patterns.
- Engage safety-critical engineering teams (automotive, aviation, medical device) for early pilots.
- Publish case studies showing certification cost/time reduction.

**Metrics and risk management**  
- Track: adoption velocity, defect density, specification review time, certification cost reduction, audit effort.
- Measure: time-to-verification (spec → deployed code), specification quality (ambiguities caught), translation reliability.
- Maintain fallback: hand-written implementations for edge cases, human override for generated code, graceful degradation if translation fails.

---

# Appendix

## A.1 Constrained Specification Language

### The Translation Ambiguity Problem

**Critical weakness identified:** The gap between informal natural language specifications and formal code (Lean 4, F*, etc.) is where AI hallucination risk is highest. Free-form Markdown like "the transfer function should move money from one account to another and make sure accounts don't go negative" leaves massive room for misinterpretation.

**Solution:** Use **constrained specification language** with formal semantics embedded in Markdown structure. This reduces translation ambiguity by 80-90% compared to free-form natural language.

### Design Principles

**1. Required sections with machine validation:**
- Specifications must include: META, TYPES, BEHAVIOR, PRECONDITIONS, POSTCONDITIONS, INVARIANTS, EXAMPLES
- Schema validator rejects incomplete specifications before translation begins
- Missing or malformed sections caught before LLM involvement

**2. Formal notation for critical properties:**
- Preconditions/postconditions use mathematical notation or controlled subset
- Example: `balance >= 0` not "balance should be positive"
- Invariants use quantifiers: `∀ a: Account. a.balance >= 0`
- Eliminates natural language ambiguity for safety-critical properties

**3. Executable examples as test oracle:**
- Specifications include concrete GIVEN/WHEN/THEN examples
- Generated code must pass all examples or translation is rejected
- Examples serve as regression tests and specification validation
- If translation produces code that fails examples, regenerate immediately

**4. Controlled vocabulary:**
- Consistent keywords (PRECONDITIONS not "Requirements", "Needs", or "Must have")
- Domain-specific templates enforce terminology (automotive template uses ISO 26262 terms, crypto template uses cryptographic standard terms)
- Reduces variance in LLM interpretation

**5. Deployment context explicitly required — target language derived, not declared:**
- META section mandates: Deployment template, Verification Level, Safety Level
- Target language is **not** declared by the spec author; it is derived from the deployment template
- Ambiguity about execution environment eliminated
- LLM has full context for appropriate code generation decisions

### Constrained Format Example

**Free-form specification (ambiguous, risky):**
```markdown
# Money Transfer

The system should allow transferring money between accounts.
Make sure the source account has enough money before transferring.
Don't let account balances go negative.
Keep track of the total money in the system.
```

**Problems with free-form:**
- What happens if source = destination?
- What does "enough money" mean exactly?
- Should failed transfers modify anything?
- "Keep track" is vague—log it? Assert it? Prove it?
- No examples to validate understanding

**Constrained specification (precise, validated):**
```markdown
# Account Transfer

## META
Deployment: backend-service
Verification: lean4
Safety-Level: financial-integrity-critical
Standard: None (internal system)

## TYPES
Account := {
  id: u64 where id > 0,
  balance: i64 where balance >= 0
}

Amount := i64 where amount > 0

TransferResult := Ok | Err(ErrorCode)

ErrorCode := INSUFFICIENT_FUNDS | SAME_ACCOUNT | INVALID_AMOUNT

## BEHAVIOR: transfer
INPUTS:
  from: Account
  to: Account
  amount: Amount

PRECONDITIONS:
  - from.balance >= amount
  - from.id ≠ to.id
  - amount > 0

POSTCONDITIONS:
  - from.balance' = from.balance - amount
  - to.balance' = to.balance + amount
  - ∀ other: Account. other ∉ {from, to} ⟹ other.balance' = other.balance

SIDE-EFFECTS:
  - Creates transfer_log entry with timestamp, from.id, to.id, amount, status

INVARIANTS (GLOBAL):
  - ∀ a: Account. a.balance >= 0
  - Σ(all_balances)' = Σ(all_balances)  // conservation of money

INVARIANTS (LOCAL):
  - Result = Ok ⟹ from.balance' = from.balance - amount
  - Result = Err(_) ⟹ from.balance' = from.balance ∧ to.balance' = to.balance

ERRORS:
  - INSUFFICIENT_FUNDS when from.balance < amount
  - SAME_ACCOUNT when from.id = to.id
  - INVALID_AMOUNT when amount <= 0

## EXAMPLES

EXAMPLE: successful_transfer
GIVEN:
  account_a = Account { id: 1, balance: 100 }
  account_b = Account { id: 2, balance: 50 }
WHEN:
  result = transfer(account_a, account_b, 30)
THEN:
  result = Ok
  account_a.balance = 70
  account_b.balance = 80
  Σ(balances) = 150  // conservation holds

EXAMPLE: insufficient_funds
GIVEN:
  account_a = Account { id: 1, balance: 20 }
  account_b = Account { id: 2, balance: 50 }
WHEN:
  result = transfer(account_a, account_b, 30)
THEN:
  result = Err(INSUFFICIENT_FUNDS)
  account_a.balance = 20  // unchanged
  account_b.balance = 50  // unchanged

EXAMPLE: same_account_rejection
GIVEN:
  account_a = Account { id: 1, balance: 100 }
WHEN:
  result = transfer(account_a, account_a, 30)
THEN:
  result = Err(SAME_ACCOUNT)
  account_a.balance = 100  // unchanged

## DEPLOYMENT
Runtime: Backend REST API endpoint /api/v1/transfer
Database: PostgreSQL, requires SERIALIZABLE transaction
Concurrency: Multiple instances, optimistic locking on account.balance
Monitoring: Prometheus metrics on transfer_success, transfer_failure
Logging: All attempts logged with user_id, from_id, to_id, amount, result
```

### Why This Works

**1. Machine-parsable before translation:**
```
$ pcdp-lint account_transfer.md

✓ All required sections present (META, TYPES, BEHAVIOR, PRECONDITIONS, POSTCONDITIONS, INVARIANTS, EXAMPLES)
✓ META complete: Deployment, Verification, Safety-Level
✓ Deployment template 'backend-service' resolved → Go (system default)
✓ TYPES well-formed: 3 types defined, all constraints valid
✓ PRECONDITIONS: 3 conditions, all reference declared variables
✓ POSTCONDITIONS: 3 conditions, all use valid notation
✓ INVARIANTS: 2 global, 2 local, all well-formed
✓ EXAMPLES: 3 examples, all have GIVEN/WHEN/THEN structure
✓ DEPLOYMENT section present

Ready for translation.
```

**2. Examples validate translation:**
```
$ spec-validate account_transfer.md account_transfer.lean

Running example: successful_transfer
  ✓ Result matches expected: Ok
  ✓ account_a.balance = 70
  ✓ account_b.balance = 80
  ✓ Conservation holds: 150 = 150

Running example: insufficient_funds
  ✓ Result matches expected: Err(INSUFFICIENT_FUNDS)
  ✓ Balances unchanged

Running example: same_account_rejection
  ✓ Result matches expected: Err(SAME_ACCOUNT)
  ✓ Balance unchanged

All examples passed. Translation validated.
```

**3. Type-checker catches formal errors:**
```
$ lean account_transfer.lean

Compiling account_transfer.lean...
✓ All types well-formed
✓ All proof obligations discharged
✓ No type errors

Ready for code generation.
```

### Specification Linting and Validation

**Pre-translation validation (catch problems before LLM):**

```bash
#!/bin/bash
# pcdp-lint.sh - Validate specification structure

spec_file=$1

# Check required sections
required="META TYPES BEHAVIOR PRECONDITIONS POSTCONDITIONS INVARIANTS EXAMPLES"
for section in $required; do
  if ! grep -q "^## $section" "$spec_file"; then
    echo "ERROR: Missing required section: $section"
    exit 1
  fi
done

# Validate META fields (v0.3.0: Target removed, Deployment required)
required_meta="Deployment Verification Safety-Level"
for field in $required_meta; do
  if ! grep -q "^$field:" "$spec_file"; then
    echo "ERROR: META missing required field: $field"
    exit 1
  fi
done

# Resolve deployment template
deployment=$(grep "^Deployment:" "$spec_file" | awk '{print $2}')
spec-template-resolve "$deployment" || {
  echo "ERROR: Unknown deployment template: $deployment"
  echo "       Use 'pcdp-lint --list-templates' to see available templates"
  exit 1
}

# Check EXAMPLES structure
if ! grep -q "^GIVEN:" "$spec_file"; then
  echo "ERROR: EXAMPLES must include GIVEN/WHEN/THEN structure"
  exit 1
fi

echo "✓ Specification structure valid"
```

**Post-translation validation (verify translation correctness):**

```bash
#!/bin/bash
# spec-validate.sh - Validate translation against examples

spec_file=$1
lean_file=$2

# Extract examples from specification
examples=$(sed -n '/^EXAMPLE:/,/^$/p' "$spec_file")

# Run test suite generated from examples
lake test --examples || {
  echo "ERROR: Translation failed example validation"
  echo "Generated code does not satisfy specification examples"
  exit 1
}

echo "✓ All examples passed - translation validated"
```

### Domain-Specific Templates

**Automotive Safety Function Template (ISO 26262):**
```markdown
# {{FunctionName}}

## META
Deployment: automotive-embedded
Verification: lean4
Safety-Level: {{ASIL-A|ASIL-B|ASIL-C|ASIL-D|QM}}
Standard: ISO-26262-2018
ASIL-Decomposition: {{if applicable}}

## SAFETY-REQUIREMENTS (from ISO 26262)
- SR-{{ID}}: {{requirement text}}
- FR-{{ID}}: {{functional requirement}}
- TR-{{ID}}: {{technical requirement}}

## FAILURE-MODES
| Failure Mode | Detection | Mitigation | ASIL |
|--------------|-----------|------------|------|
| {{mode}} | {{method}} | {{action}} | {{level}} |

## TYPES
{{same as general template}}

## BEHAVIOR: {{function}}
{{same as general template, but references SR/FR/TR}}

## INVARIANTS (SAFETY)
- {{safety invariant tied to SR-ID}}
- {{timing constraint if hard real-time}}
- {{resource bounds if safety-critical}}

## EXAMPLES
{{test cases covering normal operation and failure modes}}

## DEPLOYMENT
Runtime: {{RTOS name, version}}
Timing: {{WCET requirement}}
Resources: {{stack, heap limits}}
Diagnostics: {{DTC codes, monitoring}}
```

**Cryptographic Primitive Template:**
```markdown
# {{AlgorithmName}}

## META
Deployment: verified-library
Verification: lean4
Safety-Level: security-critical
Standard: {{FIPS-140-3|NIST-SP-800-XXX|RFC-XXXX}}

## SECURITY-PROPERTIES
- Confidentiality: {{Yes|No|N/A}}
- Integrity: {{Yes|No|N/A}}
- Authentication: {{Yes|No|N/A}}
- Constant-Time: {{Required|Not-Required}}
- Side-Channel-Resistance: {{Required|Not-Required}}

## TYPES
{{key, nonce, plaintext, ciphertext with size constraints}}

## BEHAVIOR: {{encrypt|decrypt|sign|verify}}
PRECONDITIONS:
  - {{key length valid per standard}}
  - {{nonce unique per key (if applicable)}}
POSTCONDITIONS:
  - {{correctness property}}
  - {{no key material leaked}}
INVARIANTS (SECURITY):
  - Constant-time execution (no secret-dependent branching)
  - Key zeroized after use
  - {{standard-specific requirements}}

## EXAMPLES
{{test vectors from standard (e.g., NIST CAVP)}}

## DEPLOYMENT
Runtime: {{backend|embedded|HSM}}
Key-Storage: {{method}}
Zeroization: {{when and how}}
Side-Channel-Countermeasures: {{list}}
```

### BEHAVIOR Sections

A specification may contain multiple BEHAVIOR sections, each describing a distinct operation. This is the standard pattern for CLI tools with multiple commands, or any component with more than one public entry point.

A `BEHAVIOR/INTERNAL` section describes implementation logic that is not directly user-facing — internal rules, algorithms, or sub-procedures invoked by a `BEHAVIOR`. Translators use `BEHAVIOR/INTERNAL` sections to generate private functions or methods. `pcdp-lint` validates `BEHAVIOR/INTERNAL` sections with the same structural rules as `BEHAVIOR` sections.

Example:
```markdown
## BEHAVIOR: lint
...user-facing operation...

## BEHAVIOR/INTERNAL: precedence-resolution
...internal algorithm called by lint...
```

### Translation Validation Workflow

```
1. Domain expert writes constrained specification
     ↓
2. pcdp-lint validates structure and completeness
   pcdp-lint resolves deployment template → target language
     ↓ (if invalid → reject with specific errors)
3. Human reviews specification (peer review)
     ↓
4. LLM translates specification → Lean 4/F*/Dafny (or directly to target language)
     ↓
5. spec-validate runs examples against generated code
     ↓ (if examples fail → regenerate translation)
6. Meta-language type-checker validates formal correctness
     ↓ (if type errors → regenerate or manual fix)
7. Optional: Dual-LLM for safety-critical (ASIL-C/D, DAL-A/B)
     ↓
8. Generate target code (Go, C, Rust, etc.) - deterministic, from template
     ↓
9. Property-based tests (runtime validation)
     ↓
10. Audit bundle generation (spec + IR + code + proofs + metadata)
```

### Benefits of Constrained Format

| Aspect | Free-Form Markdown | Constrained Specification |
|--------|-------------------|---------------------------|
| **Ambiguity** | High - natural language interpretation varies | Low - formal notation and structure reduce variance |
| **Validation** | None until after translation | Pre-translation linting catches errors early |
| **Examples** | Optional, often missing | Required GIVEN/WHEN/THEN tests validate translation |
| **Translation variance** | High - different LLMs produce very different results | Low - constrained input → consistent output |
| **Audit trail** | Informal, hard to trace requirements | Formal sections map to regulatory requirements |
| **Learning curve** | Easy to start, hard to write well | Steeper initial learning, but templates help |
| **Error detection** | Late (after deployment) | Early (before translation) |
| **Target language** | Must be decided by author | Derived from deployment template |

---

## A.2 Specification Format Details

### Core Specification Sections

Every specification must include:

**1. Metadata (v0.3.6 format)**
```markdown
# Component Name
- **Deployment:**   cli-tool | backend-service | wasm | ebpf | kernel-module |
                    verified-library | library-c-abi | python-tool |
                    cloud-native | enterprise-software |
                    gui-tool | academic | enhance-existing | manual | template
- **Version:**      MAJOR.MINOR.PATCH (semantic versioning)
- **Spec-Schema:**  MAJOR.MINOR.PATCH (schema version this spec was written against)
- **Author:**       Name <email>  (repeating field; multiple authors permitted)
- **License:**      SPDX identifier (e.g. Apache-2.0, MIT, GPL-2.0-only, CC-BY-4.0)
- **Verification:** none | lean4 | fstar | dafny | custom
- **Safety Level:** QM | ASIL-A | ASIL-B | ASIL-C | ASIL-D | DAL-E | DAL-D | ...
```

Note: `Target` (language) field removed from META in v0.3.0. Target language is derived from the deployment template. `Domain` field removed in v0.3.0; use `Deployment` instead. See Appendix A.11.

**2. Data Types**
```markdown
## Data Types
| Name | Type | Constraints | Safety Properties |
|------|------|-------------|-------------------|
| account_id | uuid | non-null, immutable | |
| balance | decimal(18,2) | >= 0 | ASIL-B (safety-critical) |
| transfer_id | uuid | unique, immutable | idempotency key |
```

**3. Behavior**
```markdown
## Behavior: transfer_funds
**Context:** Backend service, PostgreSQL transaction
**Inputs:** from_account_id, to_account_id, amount, transfer_id
**Outputs:** Result<TransferReceipt, TransferError>
**Preconditions:**
- from_account exists and is active
- to_account exists and is active
- amount > 0
- balance(from_account) >= amount
- transfer_id not previously used

**Postconditions:**
- balance(from_account)_after = balance(from_account)_before - amount
- balance(to_account)_after = balance(to_account)_before + amount
- sum(all_balances)_after = sum(all_balances)_before
- transfer_record created with transfer_id, timestamp, status=completed

**Error Conditions:**
- ERR_INSUFFICIENT_FUNDS if balance(from_account) < amount
- ERR_ACCOUNT_NOT_FOUND if account doesn't exist
- ERR_DUPLICATE_TRANSFER if transfer_id already used
- ERR_ACCOUNT_FROZEN if account is not active
```

**4. Invariants**
```markdown
## Invariants
- **Conservation of money:** sum(all_account_balances) is constant across all operations
- **Non-negative balances:** all account balances >= 0 at all times
- **Idempotency:** repeated calls with same transfer_id produce same result
- **Atomicity:** transfer completes entirely or not at all (no partial state)
```

**5. State Machine (if applicable)**
```markdown
## State Machine
| State | Valid Transitions | Guards |
|-------|------------------|---------|
| Pending | → Completed, → Failed | |
| Completed | (terminal) | |
| Failed | → Pending (retry) | retry_count < 3 |
```

**6. Deployment Context**
```markdown
## Deployment Context
- **Runtime:** Backend service (REST API endpoint)
- **Database:** PostgreSQL with ACID transactions
- **Concurrency:** Multiple instances, optimistic locking on account balance
- **Monitoring:** Emit metrics on success/failure, log all transfers
- **Error Handling:** Return structured errors, never panic
```

**7. Security Properties**
```markdown
## Security
- **Authentication:** Requires valid OAuth2 bearer token
- **Authorization:** User must own from_account or have transfer permission
- **Audit:** Log all transfer attempts (success and failure) with user_id, timestamp
- **Sensitive Data:** Mask account_id in logs, never log balances
```

### Optional Sections

**Performance Requirements**
```markdown
## Performance
- **Latency:** p99 < 100ms for transfer completion
- **Throughput:** Support 1000 transfers/second per instance
```

**Testing Requirements**
```markdown
## Testing
- Property test: conservation of money holds across random transfer sequences
- Property test: concurrent transfers to same account maintain consistency
- Unit test: all error conditions produce correct error codes
```

---

## A.2 Complete Example: Account Transfer (Specification → Lean 4 → Go)

**Note:** This example shows the optional formal verification path (using Lean 4 as meta-language). Teams can also choose direct code generation (Specification → Go) for faster iteration. Target language (Go) is derived from the `backend-service` deployment template, not declared by the spec author.

### Specification (Human-Authored Markdown)

```markdown
# Account Transfer State Machine

**Deployment:** backend-service
**Verification:** Lean 4 (formal proofs required for financial safety)
**Safety Level:** Financial integrity critical

## Data Types
| Name | Type | Constraints |
|------|------|-------------|
| account_id | u64 | non-zero |
| balance | i64 | >= 0 (invariant) |
| amount | i64 | > 0 (must transfer positive amount) |

## Behavior: transfer
**Inputs:** from_account_id, to_account_id, amount
**Outputs:** Result (success or error code)
**Preconditions:**
- balance(from_account) >= amount
- from_account != to_account

**Postconditions:**
- balance(from_account)_new = balance(from_account)_old - amount
- balance(to_account)_new = balance(to_account)_old + amount
- total_system_balance unchanged

## Invariants
- All balances >= 0 (enforced by type system)
- Conservation: sum(all_balances) constant
- Atomicity: both updates occur or neither occurs

## Errors
- ERR_INSUFFICIENT_FUNDS (1)
- ERR_INVALID_AMOUNT (2)
- ERR_SAME_ACCOUNT (3)

## Deployment Context
- Backend service using PostgreSQL transactions
- Concurrent access via optimistic locking
- Idempotent (can be retried safely)
```

### Lean 4 Meta-Language (AI-Generated)

```lean
-- transfer.lean
-- Provenance: spec=account_transfer.md, generated=2026-03-17
-- Meta-language: Lean 4 (verification layer)

import Std.Data.Nat

-- Account type with non-negative balance invariant
structure Account where
  id : Nat
  balance : Nat  -- Nat ensures >= 0 by construction
  h_id_nonzero : id ≠ 0

-- Error codes matching specification
inductive TransferError where
  | insufficientFunds
  | invalidAmount
  | sameAccount

-- Transfer result type
def TransferResult := Except TransferError Unit

-- System state: collection of accounts
structure SystemState where
  accounts : List Account
  h_unique_ids : accounts.map (·.id) |>.Nodup

-- Helper: get account by ID
def getAccount (s : SystemState) (id : Nat) : Option Account :=
  s.accounts.find? (·.id == id)

-- Core transfer function with formal preconditions
def transfer (s : SystemState) (from_id to_id : Nat) (amount : Nat) :
    TransferResult × SystemState :=
  if from_id == to_id then
    (.error .sameAccount, s)
  else if amount == 0 then
    (.error .invalidAmount, s)
  else
    match getAccount s from_id, getAccount s to_id with
    | some from_acct, some to_acct =>
        if from_acct.balance < amount then
          (.error .insufficientFunds, s)
        else
          let s' := updateBalance s from_id (from_acct.balance - amount)
          let s'' := updateBalance s' to_id (to_acct.balance + amount)
          (.ok (), s'')
    | _, _ => (.error .invalidAmount, s)

-- Formal property: conservation of total balance
theorem transfer_preserves_total_balance (s : SystemState) (from to : Nat) (amt : Nat) :
    let (result, s') := transfer s from to amt
    totalBalance s' = totalBalance s := by
  sorry  -- Proof to be completed by verification engineer
```

### Generated Go Code (Deployment Artifact)

```go
// transfer.go
// Provenance: spec=account_transfer.md, meta-lang=lean4, generated=2026-03-17
// Deployment template: backend-service → Go

package transfer

import "errors"

var (
	ErrInsufficientFunds = errors.New("insufficient funds")
	ErrInvalidAmount     = errors.New("invalid amount")
	ErrSameAccount       = errors.New("same account")
	ErrAccountNotFound   = errors.New("account not found")
)

type Account struct {
	ID      uint64
	Balance int64
}

// Transfer moves amount from one account to another.
// Preconditions (verified via Lean 4):
//   - from.ID != to.ID
//   - amount > 0
//   - from.Balance >= amount
// Postconditions (proven):
//   - total balance unchanged
//   - all balances >= 0
func Transfer(from, to *Account, amount int64) error {
	if from.ID == to.ID {
		return ErrSameAccount
	}
	if amount <= 0 {
		return ErrInvalidAmount
	}
	if from.Balance < amount {
		return ErrInsufficientFunds
	}
	from.Balance -= amount
	to.Balance += amount
	return nil
}
```

### Audit Bundle (Certification Artifact)

```
audit_bundle/
├── specification/
│   └── account_transfer.md          # Human-reviewable specification
├── meta_language/                   # Optional: verified path only
│   └── transfer.lean                # Lean 4 formal model with proofs
├── generated/
│   └── transfer.go                  # Deployment code (Go, from backend-service template)
├── proofs/                          # Optional: verified path only
│   ├── conservation.proof           # Proof of balance conservation
│   └── non_negative.proof           # Proof of non-negative balances
├── translation_report/
│   └── TRANSLATION_REPORT.md        # AI translator self-evaluation
└── metadata.json                    # Traceability, hashes, versions
```

The `translation_report/` directory is a required artifact for any translated
component. In regulated domain contexts (Common Criteria, ISO 26262), the
translation report provides the human-reviewable record of what the AI
translator decided and why — the closest equivalent to a compiler log for
the AI translation step. It must be produced by the translator, written to
disk, and included in the audit bundle before submission for review.

**metadata.json:**
```json
{
  "specification": {
    "file": "account_transfer.md",
    "version": "1.0",
    "author": "domain-expert@company.com",
    "safety_level": "financial_integrity_critical"
  },
  "deployment_template": {
    "name": "backend-service",
    "resolved_target": "go",
    "preset_source": "/etc/pcdp/presets/org.toml"
  },
  "translation": {
    "translator": "spec2lean-v1.0",
    "model": "claude-sonnet-4-20250514",
    "timestamp": "2026-03-17T10:00:00Z",
    "verification_mode": "formal"
  },
  "meta_language": {
    "type": "lean4",
    "version": "4.12.0",
    "proof_obligations": 2,
    "proofs_discharged": 2,
    "proofs_manual": 0
  },
  "code_generation": {
    "target_language": "go",
    "go_version": "1.24",
    "generated_lines": 38
  },
  "traceability": {
    "spec_hash": "sha256:abc123...",
    "lean_hash": "sha256:def456...",
    "code_hash": "sha256:789xyz..."
  }
}
```

---

## A.3 Competitor Landscape

| Name | Scope | Strength | Limitation | Relevance to Post-Coding Paradigm |
|------|-------|----------|------------|-----------------------------------|
| AI Coding Assistants (e.g. Copilot, Cursor) | Code completion | Widely adopted; boosts productivity | **Humans still write code**; no formal guarantees; **prohibited in safety-critical domains** | Different paradigm - still requires programming |
| Specification-first OSS toolkits | Spec → code mapping | Accessible; CI-friendly | Often lacks verification; **humans still write/edit code** | Similar goal, lacks formal verification option |
| F* / Coq code extraction | Verified components | Strong formal proofs; proven extraction | **Humans write F*/Coq code** (steep learning curve); not AI-mediated | Inspiration for meta-language layer, but requires formal methods expertise |
| TLA+ / Model checkers | System modeling | Excellent for protocol verification | Not for implementation; **humans write TLA+ specs** | Complementary for high-level design validation |
| Lean 4 / Dafny / Isabelle | Theorem proving | Mathematical guarantees; mature tooling | **Humans write formal proofs**; steep learning curve | We use as meta-language (AI generates, humans don't write) |
| Low-code/No-code platforms | Application development | Non-programmers can build apps | Domain-specific; no formal verification; limited to CRUD/workflow apps | Different market (business apps vs. systems software) |

**How Post-Coding Development differs:**

1. **Humans never write code** - not even with AI assistance. Specifications only.
2. **AI generates all implementation** - including formal proofs (when meta-language path used).
3. **Enables safety-critical development** - solves regulatory prohibition on AI-generated code through auditable specifications + formal proofs.
4. **Flexible verification** - optional meta-language layer (can skip for low-risk components).
5. **Domain expert authorship** - specifications written by people with system knowledge, not formal methods experts.

---

## A.4 Step-by-Step Adoption in Existing Projects

**Phase 0: Preparation**
- Identify candidate components (small, high-risk, well-understood domain).
- Train team on specification writing (not programming, not formal methods).
- Select deployment templates applicable to the project domain.
- Configure org-level presets in `/etc/pcdp/presets/`.

**Phase 1: Specification alongside existing code**
- Write Markdown specifications for 1-2 existing components.
- Keep current manual implementation unchanged.
- Run specification validators in CI (check completeness, consistency).
- Goal: Validate that domain experts can write adequate specifications.

**Phase 2: Generate and compare (direct path first)**
- Use AI translator: Specification → target language via deployment template (skip formal verification initially).
- Run existing test suites against generated code.
- Compare behavior with manual implementation.
- Goal: Validate translation quality without verification overhead.

**Phase 3: Add formal verification for critical paths**
- For safety-critical functions, switch to verified path: Specification → Lean 4 → target language.
- Run formal verification, discharge proof obligations.
- Compare with direct generation path.
- Goal: Measure verification cost/benefit tradeoff.

**Phase 4: Shadow deployment**
- Deploy generated code in non-critical paths or test environments.
- Monitor behavior, performance, and failure modes.
- Keep manual implementation as fallback.
- Goal: Build confidence in generated code reliability.

**Phase 5: Replace and certify**
- Replace manual implementation with generated code in production.
- Use audit bundles for certification (if regulated domain).
- Add property-based tests, invariant monitoring.
- Goal: Achieve regulatory approval, measure cost reduction.

**Phase 6: Expand and iterate**
- Add more components using specification-first approach.
- Create organization-specific specification templates.
- Integrate specification reviews into PR gating.
- Train new engineers in specification writing (not coding).

**Fallback mechanisms:**
- Maintain manual implementations for edge cases where specification is unclear.
- Human override capability: mark components as "MANUAL_IMPL" to skip generation.
- Graceful degradation: if translation fails, use previous generated version or manual fallback.
- Governance: Establish review board for approving meta-language changes and certifying backends.

---

## A.5 Governance, Licensing, and Next Steps

### Licensing

The project uses differentiated licensing that mirrors the Linux ecosystem model:

**Specification documents and deployment templates: `CC-BY-4.0`**
The specification format, constrained Markdown schema, and deployment templates
are licensed under Creative Commons Attribution 4.0 International. Anyone may
read, implement, adapt, and build upon the specifications — including proprietary
translators and commercial tools — provided attribution is given. This maximises
adoption and allows regulated-industry organisations to build certified
closed-source translators without license conflict.

**Reference implementation (`pcdp-lint`): `GPL-2.0-only`**
The `pcdp-lint` validator and any other reference tools are licensed under
GNU General Public License version 2 only. This follows the Linux kernel model:
companies that ship or modify `pcdp-lint` must contribute their changes back.
This forces collaboration on the validation toolchain — the compliance layer
that must remain community-controlled and vendor-neutral for the paradigm to
be trustworthy in regulated markets. No single company can fork `pcdp-lint`,
add proprietary validation rules, and use it to lock in customers.

The strategic rationale: the GPL-2.0-only reference implementation was a
significant factor in Linux's success — it forced competitors to work together
on the platform layer. The same dynamic applies here: interoperability on
the validator is more valuable than any single company's proprietary advantage.

**License compatibility note:**
- CC-BY-4.0 specifications may be implemented by GPL-2.0-only, Apache-2.0,
  proprietary, or any other licensed tools without conflict.
- GPL-2.0-only `pcdp-lint` may validate Apache-2.0, CC-BY-4.0, and other
  licensed specifications without conflict.
- Organisations developing proprietary translators should use the
  CC-BY-4.0 specification documents as their normative reference.

**Security:**
- Publish security disclosure policy.
- Require reproducible builds for safety-critical modules.
- Audit bundle format must be stable and versioned.

**Intellectual Property:**
- Specification format is open standard (anyone can implement translators).
- Reference translators are open source.
- Organizations may develop proprietary translators or meta-language adapters.

**Next Steps:**
1. Finalize specification schema (stable v1.0 format), incorporating v0.3.0 META changes.
2. Implement deployment template system with systemd-style preset layering.
3. Implement reference translator with both paths:
   - Direct: Specification → target language (via deployment template)
   - Verified: Specification → Lean 4 → target language
4. Build `pcdp-lint` as reference implementation under the paradigm itself.
5. Develop domain-specific templates (automotive, finance, embedded, crypto).
6. Run pilot with safety-critical domain (automotive or medical device).
7. Engage regulatory bodies (ISO, FAA, FDA) for certification pathway validation.
8. Publish specification format as open standard.
9. Create community hub for sharing specification patterns.
10. Iterate based on pilot feedback and expand meta-language adapter ecosystem.

---

## A.6 When to Use This Paradigm (Decision Framework)

### Decision Tree

```
Does your project involve...
├─ Safety-critical or regulated requirements (automotive, aviation, medical)?
│  ├─ YES → Strong fit. Solves regulatory prohibition on AI code generation.
│  └─ NO → Continue below
│
├─ High-assurance requirements (crypto, finance, security)?
│  ├─ YES → Good fit. Formal verification path provides mathematical guarantees.
│  └─ NO → Continue below
│
├─ Complex state machines or protocols hard to implement correctly?
│  ├─ YES → Good fit. Specifications clarify intent; verification catches errors.
│  └─ NO → Continue below
│
├─ Need clear separation between "what" (requirements) and "how" (implementation)?
│  ├─ YES → Medium fit. Specifications serve as living documentation.
│  └─ NO → Stick with traditional development or AI-assisted coding.
```

### When to Choose Formal Verification Path vs. Direct Generation

| Your Situation | Use Verified Path (Spec → Meta-Lang → Code) | Use Direct Path (Spec → Code) |
|----------------|----------------------------------------------|-------------------------------|
| Safety-critical (ISO 26262, DO-178C) | **Required** - formal proofs needed for certification | Not sufficient for certification |
| Financial transactions | **Recommended** - prove conservation, atomicity | Acceptable with extensive testing |
| Crypto implementations | **Recommended** - prove constant-time, correctness | Risky - subtle bugs have security impact |
| Device drivers | **Optional** - verify memory safety, state correctness | Acceptable for non-critical drivers |
| Web backend services | **Optional** - verify business logic invariants | Typical choice - faster iteration |
| Internal tooling | Not needed | **Recommended** - fastest path |

---

## A.7 De-Risking AI Dependency

### The Core Risk

AI translators can hallucinate invalid intermediate representations, introduce subtle bugs, or degrade over time as models change. Unlike traditional compilers, LLM behavior is probabilistic and not deterministic.

### Mitigation Strategies

#### 1. Multi-Stage Verification Pipeline

```
Specification (human)
  ↓ [pcdp-lint validates + deployment template resolves target language]
  ↓ [AI translates]
Intermediate Representation (machine-checkable)
  ↓ [Meta-language type-checks] ← VERIFICATION CHECKPOINT
Proof Obligations
  ↓ [Theorem prover discharges] ← VERIFICATION CHECKPOINT
Generated Code (target language from template)
  ↓ [Property tests] ← VERIFICATION CHECKPOINT
Audit Bundle
```

**Key insight:** AI only handles Spec→IR translation. Every downstream step is deterministic and verifiable.

#### 2. Translation Confidence Scoring

Track and surface translator confidence:

```markdown
## Translation Report
- Specification complexity: Medium (2 state machines, 8 invariants, 1 deployment context)
- Deployment template resolved: backend-service → Go
- IR validation: PASS (well-formed, complete)
- Meta-language check: PASS (Lean 4 type-checked, 0 errors)
- Novelty flag: ⚠️ Pattern "optimistic locking in embedded context" not common in training
- Confidence: 82% (suggest manual review of concurrency handling)
```

#### 3. Pinned Translator Versions + Reproducibility

- **Pin AI model versions** in CI configuration (e.g., `translator: claude-sonnet-4-20250514`).
- **Lock translation cache:** Store generated IR in version control; regenerate only when specification changes.
- **Reproducible builds:** Audit bundles include translator version, model identifier, timestamp, and specification hash.

#### 4. Ensemble Translation for Critical Paths

For highest-assurance components (e.g., cryptographic primitives, safety functions):

```
Specification → [Translator A] → IR_A + Tests_A
             → [Translator B] → IR_B + Tests_B
             → [Translator C] → IR_C + Tests_C

Cross-Validation:
- Type-check all IRs independently
- Run Tests_A against IR_B and IR_C
- Differential testing: All IRs produce identical outputs for same inputs
```

#### 5. Fallback to Human Override

Always allow engineers to:
- **Hand-write IR** when translator fails or produces low-confidence output.
- **Annotate specifications** with translation hints.
- **Override generated code** with `MANUAL_IMPL` marker that bypasses regeneration.

#### 6. Continuous Validation in CI

Every commit runs:
- Specification validation: complete, required sections present
- Deployment template resolution: template known, target language resolved
- Schema check: specification conforms to schema version
- IR validation: generated IR well-formed
- Meta-language check: IR type-checks (if verified path)
- Property tests: generated artifacts satisfy postconditions
- Regression tests: behavior matches previous version
- Audit bundle generation

#### 7. Community-Driven Pattern Library

Build open repository of:
- **Verified spec→IR patterns** for common idioms (state machines, transactions, crypto operations).
- **Known-bad translations** that humans caught (negative examples for training/testing).
- **Test cases** for translator validation.
- **Domain templates** per deployment context.

#### 8. Dual-LLM Verification for Critical Components

For highest-assurance modules, employ independent translation and formal cross-validation.

**See Appendix A.10 for full technical implementation details.**

### Summary: Defense in Depth

| Layer | Mechanism | What It Catches |
|-------|-----------|----------------|
| Template resolution | Deployment template system | Missing or ambiguous target language decisions |
| Translation | Schema validation, confidence scoring | Malformed IR, incomplete specs |
| Verification (meta-language) | Type-checking, proof discharge | Semantically invalid IR, property violations |
| Generation | Deterministic code generation, property tests | Code violating postconditions |
| Deployment | Differential testing, regression tests | Behavioral changes, invariant violations |
| Escape hatch | Human override, MANUAL_IMPL | Cases where AI fundamentally fails |

---

## A.8 Why Lean 4 as One Possible Meta-Language

### Selection Criteria

When choosing a meta-language for the verified path, we evaluated:

1. **LLM training data presence:** Can current AI models reliably generate correct code in this language?
2. **Verification power:** Does it support dependent types, refinement types, or proof obligations?
3. **Community & tooling:** Active community, good documentation, mature toolchains?
4. **Code extraction:** Can it emit readable, auditable target code?
5. **Learning curve (for humans who might read it):** Accessible syntax for audit purposes?

### Why Lean 4 is a Strong Candidate

| Criterion | Lean 4 | ATS2 (rejected after testing) | F* | Dafny |
|-----------|--------|-------------------------------|-----|-------|
| **LLM compatibility** | **Excellent** | Poor - niche syntax | Good | Good |
| **Verification power** | **Excellent** - dependent types | **Excellent** - linear types | **Excellent** - refinement types | Good - SMT-based |
| **Community** | **Growing rapidly** | Small, stagnant | Medium, research-focused | Medium, Microsoft-backed |
| **Code extraction** | Good - `lean --target=c` | Good - emits C directly | **Excellent** | Limited - primarily .NET |
| **Learning curve** | Medium | **Steep** | Steep | **Low** |
| **Modern tooling** | **Excellent** - LSP, lake | Dated | Good | Good |

### Lessons from the ATS2 Experiment

We initially considered ATS2 for its powerful linear type system. Multiple LLMs struggled to generate syntactically correct ATS2 code — the syntax is underrepresented in training data. This validated the pluggable architecture and the lesson: **LLM compatibility is non-negotiable** for an AI-native paradigm.

### Alternative Meta-Languages and Custom Formats

- **F\*:** SMT-based verification, effect tracking, proven in production at Microsoft.
- **Dafny:** Accessible syntax, .NET integration, lower barrier to human audit.
- **Refinement-type systems:** Rust ecosystem integration (Flux, Prusti).
- **Coq:** Maximum proof power for academic/research contexts.
- **Custom intermediate formats:** Domain-specific IRs for particular compliance regimes.

---

## A.9 Stakeholder Comparison

### High-Level Comparison

| Dimension | Traditional Coding | AI-Assisted Coding | Post-Coding Development |
|-----------|-------------------|-------------------|------------------------|
| **Human Role** | Write implementation code | Write code + iterate with AI | Write specifications (no coding) |
| **Primary Artifact** | Source code | Source code (AI-influenced) | Specification |
| **Target Language** | Developer decides | Developer decides | Deployment template decides |
| **Memory Safety** | Depends on language | Same as language choice | Guaranteed by meta-language (verified path) |
| **Auditability** | Hard for complex code | Harder (AI suggestions opaque) | **Easy: specs + proofs** |
| **Regulatory Compliance** | Expensive manual process | **Prohibited in safety-critical** | **Enabled: auditable specs + formal proofs** |
| **Long-Term Maintainability** | Code rot | Code rot + AI drift | **Specifications remain stable** |

### Engineering Leadership Perspective

| Concern | Traditional | AI-Assisted | Post-Coding Development |
|---------|-------------|-------------|------------------------|
| Hiring difficulty | High | Medium | **Low (domain experts, not programmers)** |
| Onboarding time | Long | Medium | **Short (learn spec format, review specs)** |
| Bus factor | High | Medium | **Low (knowledge explicit in specs)** |
| Technical debt | High | High | **Low (specs stable, code regenerated)** |

---

## A.10 Technical Deep-Dive: Dual-LLM Verification

### Overview

When using AI to translate specifications to meta-language implementations, there is inherent risk of translation errors and hallucinations. A dual-LLM approach with cross-validation significantly improves reliability by having two independent translators compete and validate each other against a common specification.

### Core Methodology

1. **LLM-1:** Generates `ir_1.lean` + `tests_1.lean` from `spec.md`
2. **LLM-2:** Generates `ir_2.lean` + `tests_2.lean` from the same `spec.md`
3. **Cross-validation:** Run each test suite against both IR implementations

**Validation matrix:**
- `tests_1` against `ir_1` (LLM-1 self-check)
- `tests_1` against `ir_2` (LLM-1 tests validate LLM-2 work)
- `tests_2` against `ir_1` (LLM-2 tests validate LLM-1 work)
- `tests_2` against `ir_2` (LLM-2 self-check)

If all four pass, confidence is high. If cross-tests fail, either the specification is ambiguous or one LLM hallucinated.

### Verification Strategies

#### Strategy 1: Property-Based Testing

Extract key properties from specification and verify both implementations satisfy them.

#### Strategy 2: Differential Testing

Compare outputs directly for identical inputs from specification.

#### Strategy 3: Formal Equivalence Proofs (Highest Assurance)

Prove mathematical equivalence of two IR implementations in the meta-language.

#### Strategy 4: Specification-Driven Validation

Define formal specification structure and verify both implementations conform.

---

## A.11 Deployment Templates and Target Language Resolution

*Added in v0.3.0*

### Motivation

Early versions of this paradigm required spec authors to declare a `Target` language in the META section. This was identified as an **anti-pattern**: it pulls the specification author back into implementation thinking, couples specifications to technology choices that change over time, and creates inconsistency when an organisation standardises on a different language than the template author assumed.

The key insight: **target language is not a free variable—it is a function of deployment context.** Once you declare `Deployment: ebpf`, the target language space collapses to one option (restricted C). The spec author never needed to decide anything.

### Design

Target language selection follows this resolution order:

1. The deployment template declares a **default** target language and optionally a set of **valid alternatives**
2. Organisation, user, and project presets may **override** the default within the valid set
3. The spec author may **explicitly override** in the META section, but only if the deployment template permits it — and doing so is a conscious, documented deviation

This keeps specifications clean and stable. A spec written today remains valid if the organisation changes its Go default to Rust in 2029 — the spec does not change, only the preset does.

### Deployment Template Reference

| Template Name | Default Language | Valid Alternatives | Fixed Constraints | Notes |
|---|---|---|---|---|
| `wasm` | Rust | — | WASM-compatible subset | No alternatives; WASM toolchain requires Rust |
| `ebpf` | Restricted C | — | eBPF verifier rules: no unbounded loops, no floating point, stack limits | No alternatives; kernel eBPF verifier is C-specific |
| `kernel-module` | C | — | Kernel coding style, no userspace libs | Optional Lean 4 verified path strongly recommended |
| `verified-library` | C | Rust | Constant-time, side-channel resistance, formal verification recommended, FIPS/CC/ASIL/DAL compliance | Replaces `crypto-library`. Covers all safety- and security-critical C-ABI libraries. Formal verification via Lean 4/F*/Dafny strongly recommended. QM safety level not permitted. |
| `cli-tool` | Go | Rust, C, C++, C# | Single static binary preferred | Platform-independent default. C# targets Windows. |
| `gui-tool` | *OS-dependent* | — | See platform slot below | No universal default |
| `cloud-native` | Go | — | Kubernetes/CNCF conventions, OCI-compatible | Reflects ecosystem consensus |
| `backend-service` | Go | Rust | 12-factor app conventions | |
| `library-c-abi` | C | Rust (via cbindgen) | Stable ABI, C-compatible headers | As of CMake 4.3, CPS file required for CMake ecosystem consumers. `.cps` is a required deliverable. |
| `python-tool` | Python | — | QM safety level only, Verification: none mandatory | No formal verification path. For tooling, automation, data pipelines. Not suitable for safety-critical components. `pyproject.toml` required deliverable. |
| `enterprise-software` | Java | Kotlin | JVM ecosystem assumed | |
| `academic` | Fortran | C, Julia | Math/Physics/HPC context | |
| `enhance-existing` | *Must declare* | *Must declare* | Must match existing codebase | See below |
| `manual` | *Must declare all* | — | No template assistance | Fully explicit fallback |
| `project-manifest` | N/A | — | Multi-component project definition | Architect artifact. No code generated; produces project audit bundle. v0.3.9 target. |
| `mcp-server` | Go | Rust | MCP protocol (stdio / HTTP+SSE), tool registration, error conventions | For MCP server components. v0.3.9 target. |

**GUI tool platform slots:**

| Platform | Default Language |
|---|---|
| Linux | C (GTK) or Go (with CGo) |
| Windows | C# |
| macOS | Swift |
| Cross-platform | Go (with platform-native binding layer) |

**enhance-existing requirements:**

When using `enhance-existing`, the spec author must declare the existing language:

```markdown
## META
Deployment: enhance-existing
Language: COBOL
```

Valid values include any language with a functioning compiler/interpreter: COBOL, Fortran, PHP, Python, Perl, Ruby, Java, C, C++, Go, Rust, and others. The toolchain must be able to generate code compatible with the existing codebase. The pcdp-lint tool will warn if no translator backend exists for the declared language.

### Preset Layering (systemd-style)

Presets follow a layered override model identical in principle to systemd's unit file loading. Later layers override earlier ones; the first match for any given setting wins in reverse order.

```
/usr/share/pcdp/templates/        # shipped deployment template definitions (read-only)
/usr/share/pcdp/presets/          # shipped vendor/community presets (read-only)
/etc/pcdp/presets/                # system administrator overrides
~/.config/pcdp/presets/           # user-level overrides
<project-dir>/.pcdp/presets/      # project-local overrides (committed to git)
```

### Template Search Path

Deployment template files (`*.template.md`) are located at runtime using
a compile-time variable `TEMPLATE_DIR`. This ensures no runtime path
discovery and no environment variable magic — the path is baked in at
build time, consistent with supply chain security requirements.

```
TEMPLATE_DIR (compile-time default, read-only)     /usr/share/pcdp/templates/
/etc/pcdp/templates/                         system administrator additions
~/.config/pcdp/templates/                    user additions
<project-dir>/.pcdp/templates/               project-local additions
```

Later entries take precedence. A template file found in a later path
overrides one with the same name in an earlier path, allowing
organisations and projects to ship custom or overriding templates
without modifying the system-level installation.

The `TEMPLATE_DIR` default for Linux OBS packages is
`/usr/share/pcdp/templates/`. Platform defaults for macOS
and Windows are deferred to v2.

For OBS packaging, `TEMPLATE_DIR` must be set at build time:
```
%build
make build TEMPLATE_DIR=/usr/share/pcdp/templates/
```

Example preset file (`/etc/pcdp/presets/suse.toml`):

```toml
[templates.cli-tool]
default_language = "go"
# Organisation standardises on Go for CLI tools; Rust available for explicit opt-in

[templates.backend-service]
default_language = "go"

[templates.verified-library]
default_language = "c"
verification = "lean4"  # org mandates formal verification for all safety/security-critical libraries

[templates.kernel-module]
default_language = "c"
verification = "lean4"  # org mandates formal verification for kernel code

[templates.python-tool]
# No language override needed — Python is the only option
# Reminder: python-tool is QM only, not for safety-critical components
```

Example project-local override (`.pcdp/presets/project.toml`):

```toml
[templates.cli-tool]
default_language = "rust"
# This project uses Rust for the CLI layer; overrides org default of Go
```

### pcdp-lint as Reference Implementation

`pcdp-lint` is the first component to be specified and generated under this paradigm. This serves two purposes:

1. **Empirical validation:** If the paradigm cannot describe its own tooling unambiguously, the template design is incomplete. Any gap discovered during `pcdp-lint` specification authoring feeds back directly into template and schema improvements.

2. **Demonstrable credibility:** A specification system that can specify itself is a stronger argument than any hand-picked external example.

The `pcdp-lint` specification uses:

```markdown
## META
Deployment: cli-tool
Verification: none
Safety-Level: QM
```

Target language resolves to Go (system default for `cli-tool`), unless overridden by org preset. The spec author — in this case, the project itself — does not need to decide or declare a target language.

The `pcdp-lint` specification will be developed as the next working artifact, with the template schema defined above as its validation schema.

### Open Questions for v0.4.0

The following questions were deferred to keep v0.3.0 focused:

1. **Interface compatibility for enhance-existing:** When enhancing existing code, should the spec declare the interface type (extend-module, replace-function, add-endpoint, ffi-wrapper)? Deferred — start with `Language:` declaration only and add interface typing when the first real enhance-existing use case is attempted.

2. **Verification level in templates:** Should templates specify a default verification level (none / lean4 / fstar) or leave this always to the spec author? Current position: verification level stays with the spec author as an explicit architectural decision. Templates may recommend but not mandate.

3. **Template versioning:** How are breaking changes to deployment templates handled when a spec was authored against an older template version? Requires a template version pinning mechanism, likely in the audit bundle metadata.

---

## A.12 Related Work and Industry Landscape

### Overview

The Post-Coding Development Paradigm combines several established ideas in a novel way. Each individual ingredient has precedent; the combination does not exist as a productised, accessible, regulated-domain-ready system.

### Closest Existing Approaches

**OpenAPI / AsyncAPI Specifications**
Structured, machine-readable specifications describing APIs — lintable, diffable, and code-generatable from a schema. The closest analogy to our constrained Markdown format in terms of tooling philosophy. Key differences: OpenAPI describes interfaces only, not full component behaviour, invariants, or state machines. There is no formal verification layer, no deployment template concept, and no pathway to regulated-domain certification.

**Behaviour-Driven Development — Gherkin / Cucumber**
The GIVEN/WHEN/THEN structure used in the EXAMPLES section of this paradigm is directly borrowed from the BDD tradition, which has been in production use since 2008. Gherkin drives test generation, not code generation, and has no formal verification layer. It is a spiritual predecessor to the EXAMPLES section, not a competitor.

**TLA+ and Alloy**
Formal specification languages used in industry for system design. Amazon Web Services has used TLA+ extensively for distributed systems (DynamoDB, S3). Alloy is used in security protocol design. Both provide mathematical rigour over system behaviour. Key differences: humans write TLA+ and Alloy directly — these are programming languages for specifications, not natural language. There is no AI translation layer, no deployment template concept, and no pathway from specification to deployable code.

**F* and HACL\***
Microsoft Research's F* has been used to produce formally verified C code for cryptographic primitives. The HACL* library (used in Firefox, the Linux kernel, and WireGuard) was produced this way. This is the closest existing work to our verified path. Key difference: humans write F* directly. The paradigm's contribution is placing AI as the translator so domain experts — not formal methods specialists — author the primary artifact.

**Dafny**
Microsoft Research's Dafny compiles verified code to multiple target languages. It is accessible enough that some engineers use it without a formal methods background. Key difference: Dafny is still a programming language. Authors write Dafny, not structured natural language. It is a candidate meta-language within this paradigm, not a competing paradigm.

**Low-Code / No-Code Platforms**
OutSystems, Mendix, Appian and similar platforms allow non-programmers to build applications through visual editors. Key differences: domain-specific (primarily CRUD and workflow applications), no formal verification, not applicable to systems programming, embedded, or safety-critical contexts.

**LLM-Based Code Generation**
Generates code from freeform natural language prompts. Widely adopted for productivity gains in non-regulated contexts. Key differences: no structured specification format (freeform prompts), no formal verification, no audit bundle, explicitly prohibited in safety-critical and regulated domains by ISO 26262, DO-178C, IEC 62304, and Common Criteria frameworks. This is "vibe coding" — the paradigm this work explicitly positions against.

**Correct-by-Construction Synthesis (Research)**
Academic research into automatically synthesising programs from formal specifications exists under labels including "program synthesis", "correct-by-construction development", and "specification-carrying code." Active research groups at MIT, CMU, and several European universities. Key differences: research prototypes, not productised; require formal specification languages (not natural language); no deployment template concept; no pathway to regulated-domain certification.

### Comparative Summary

| Approach | Human writes | AI layer | Formal verification | Regulated domains | Deployment templates |
|---|---|---|---|---|---|
| OpenAPI / AsyncAPI | Structured schema | Code-gen tools | No | No | No |
| Gherkin / BDD | Natural language tests | No | No | No | No |
| TLA+ / Alloy | Formal spec language | No | Yes (model checking) | Partial | No |
| F* / HACL* | F* code | No | Yes (dependent types) | Yes (crypto) | No |
| Dafny | Dafny code | No | Yes (SMT) | No | No |
| LLM code generation | Prompts + code | Yes | No | **Prohibited** | No |
| Program synthesis (research) | Formal spec language | Partial | Yes | No | No |
| **Post-Coding Development** | **Structured natural language** | **Yes** | **Optional, pluggable** | **Yes (primary target)** | **Yes** |

### What Is Genuinely Novel

The combination that does not exist elsewhere:

1. **Natural language as the primary artifact** — not a programming language, not a formal language, not a visual editor. Structured Markdown that domain experts can write without programming training.

2. **Deployment templates as a first-class concept** — target language is not a human decision; it is derived from deployment context. Specifications are technology-agnostic and stable across language ecosystem changes.

3. **Formal verification as optional and pluggable** — teams choose verification level based on risk and regulatory context. The same specification format works for a quick CLI tool (no verification) and a safety-critical automotive function (Lean 4 verified path).

4. **Regulated-domain certification as a design goal** — audit bundles, traceability matrices, and formal proofs are first-class outputs, not afterthoughts. The paradigm is designed to satisfy ISO 26262, DO-178C, IEC 62304, and Common Criteria requirements.

5. **Self-hosting from the first artifact** — the `pcdp-lint` validator is developed using its own specification format. This is not a theoretical claim; it is an empirical test run from the beginning.

### Academic Framing

For readers from a formal methods or programming languages background: this paradigm sits at the intersection of *specification-driven development*, *correct-by-construction synthesis*, and *AI-mediated program translation*. The novelty is not in any individual technique but in the system design that makes these accessible to domain experts in regulated industries — and in the deployment template abstraction that decouples specification authoring from implementation technology choices.

---

## A.13 Recommended Translator Prompt

The following prompt is the standard wrapper for invoking an AI translator
against a Post-Coding specification. It is technology-agnostic and should
work with any capable LLM. It is versioned alongside the schema.

**Prompt version:** 0.3.3

### System Prompt Template

Set once per invocation, before the user prompt. Replace parameters
with concrete values — do not send angle-bracket placeholders to the model.

```
Your working directory is <target-dir>/
The following input files are present there:
  - <deployment-template>.template.md  — the deployment template
  - <spec-name>.md                     — the component specification to implement
Do NOT output file contents to the terminal.
Do NOT write the input files to disk again.
Write ALL to be produced files to disk using the filesystem write tool.
After writing each file, confirm the filename written.
Also write the translation report as TRANSLATION_REPORT.md into that directory.
```

**Parameters:**
- `<target-dir>` — working directory on the target filesystem
- `<deployment-template>` — exact filename of the deployment template (e.g. `cli-tool.template_0_3_3`)
- `<spec-name>` — exact filename of the specification (e.g. `pcdp-lint_0_3_3`)

### User Prompt Template

```
I am providing two files in your working directory:

1. <deployment-template>.template.md — the deployment template, which defines
   conventions, constraints, defaults, and required deliverables for this
   type of component under the Post-Coding Development Paradigm.

2. <spec-name>.md — the component specification, written in the
   Post-Coding Development Paradigm format.

Before writing any files, briefly state your plan: which files you
will produce, in which order, and why.

Your task:
Implement the component in full, exactly as specified. Do not add
features not described in the specification. Do not omit any
specified behaviour.

Derive the target language from the deployment template:
the template declares the default language and valid alternatives.
Use the default unless a project preset overrides it — if you deviate
from the default, state why explicitly.

Produce all deliverables defined in the deployment template's
DELIVERABLES section, in the order specified there. Do not enumerate
these files yourself — read them from the template's DELIVERABLES section.

The deployment template describes the target runtime environment of
the generated artifact, not the environment where this prompt is
being evaluated. Do not make language or toolchain decisions based
on what is available in your current execution environment.

After writing all files, verify each file exists on disk by listing
the directory. Report any discrepancies between intended and actual
deliverables.

Do not ask clarifying questions. If the specification is ambiguous,
make the most conservative interpretation, implement it, and note
the ambiguity in the translation report.
```

### Prompt Design Rationale

**System prompt carries environment facts; user prompt carries task instructions.**
Separating these allows the system prompt to be set once per environment
and the user prompt to be reused across specifications.

**Concrete filenames in the system prompt, not placeholders.**
Smaller models do not reliably resolve angle-bracket placeholders to
actual files. The system prompt must name the files explicitly with
their role labels ("the deployment template", "the component specification").

**"Before writing any files, briefly state your plan"** — the reasoning
step before generation. Models that plan before writing produce more
complete and consistent deliverable sets than those that generate immediately.

**"Verify each file exists on disk after writing"** — closes the delivery
loop. The translator confirms what actually landed, not just what it
intended to write.

**"Derive the target language from the deployment template"** — the
paradigm's core claim under test. The LLM must read the template and
resolve the language, not be told explicitly.

**"The deployment template describes the target runtime environment...
not the environment where this prompt is being evaluated"** — prevents
the LLM from making language choices based on sandbox constraints.

**"Do not ask clarifying questions"** — forces ambiguities into the
translation report. This is the feedback mechanism for spec improvement.

---

## A.14 Empirical Testing: pcdp-lint

In March 2026, the `pcdp-lint` specification and `cli-tool` deployment template
were submitted to multiple LLMs across different environments as an empirical
test of the paradigm. This appendix documents all test runs and findings.

LLM identities are anonymized. The test methodology and findings are what
matter — not which commercial product produced them. Labels reflect
capability class and environment rather than vendor.

### Test Configuration

- **Specification:** `pcdp-lint.md`
- **Template:** `cli-tool.template.md`
- **Prompt:** A.13 prompt, refined iteratively across runs
- **Environments tested:** Browser, API + mcphost + filesystem MCP, local Ollama

### Models Tested

| Label | Capability Class | Location | Infrastructure |
|---|---|---|---|
| LLM-A | Frontier, IDE-integrated | US cloud | IDE plugin |
| LLM-B | Frontier, browser-based | US cloud | Browser |
| LLM-C | Frontier, browser-based | US cloud | Browser |
| LLM-C | Frontier, API-accessible | US cloud | API + mcphost |
| LLM-C | Frontier, API + extended reasoning | US cloud | API + mcphost |
| LLM-D | Frontier, API-accessible | US cloud | API + mcphost |
| LLM-E | 120B open-weight | Regional EU provider | API + mcphost |
| LLM-F | 30B open-weight coder | Local hardware | Ollama + mcphost |

### Universal Finding: Language Resolution

**Every model tested resolved Go as the target language by reading the
deployment template, without being told explicitly.** All cited the
template's `LANGUAGE | Go | default` entry as the source of their decision.
This is the paradigm's core claim — and it held across all tested models,
environments, and prompt versions.

### Deliverable Completeness by Run

| | LLM-A | LLM-B | LLM-C browser | LLM-C mcphost v1 | LLM-C mcphost v2 | LLM-D | LLM-E best | LLM-F |
|---|---|---|---|---|---|---|---|---|
| main.go / all 7 rules | ~6/7 | 7/7 | 7/7 | 7/7 | 7/7 | scaffold | 7/7 | TBD |
| RPM spec | No | No | Yes | Yes | Yes | Yes | Yes | TBD |
| DEB complete | No | No | workaround | workaround | **Yes proper** | Yes | Yes | TBD |
| Containerfile | ? | ? | No | No | No | No | **Yes** | TBD |
| Makefile | ? | ? | Yes | Yes | Yes | ? | Yes | TBD |
| README with OBS | No | No | ? | ? | **Yes** | ? | ? | TBD |
| LICENSE | No | No | No | No | No | No | Yes | TBD |
| Report to disk | No | No | No | No | **Yes** | **Yes** | Partial | TBD |
| Template constraints table | No | No | No | No | **Yes** | No | No | TBD |
| Confidence calibration | Good 85% | Poor 100% | Good 94% | Poor 100% | Good 90-95% | **Excellent** | Poor 100% | TBD |

### Key Findings

**Delivery mode determines deliverable completeness.**
Browser-based runs produced source code only. MCP filesystem runs produced
complete packaging artifacts. The system prompt ("Write ALL files to disk,
do NOT output to terminal") was essential for reliable filesystem delivery.

**Extended reasoning improves output quality.**
The LLM-C run with extended reasoning (v2 mcphost run) produced:
- Correct `debian/` subdirectory structure without workaround
- Filesystem verification after writing each file
- Template constraints compliance table in the translation report
- OBS-aware README with correct `zypper`/`apt`/`dnf` installation commands

These behaviours were inferred from the template and spec without explicit
prompting, suggesting extended reasoning enables deeper cross-referencing
between the two input documents.

**LLM-D: most honest translation report.**
The only model to report a scaffold implementation with honest low confidence
scores (10% for unimplemented rules). Also the only model to identify the
AST/parser complexity as a genuine spec gap:

> "A full, production-ready implementation of all 15+ validation rules would
> be substantially more complex, requiring a proper Markdown parser and
> Abstract Syntax Tree (AST) to accurately identify line numbers and section
> boundaries for all edge cases."

This finding led to the parsing approach note added to the spec in v0.3.3.
Per Option B (confirmed): the spec describes semantics, not parsing strategy.
EXAMPLES are the acceptance test regardless of internal implementation.

**LLM-E: digital sovereignty proof of concept.**
A 120B open-weight model hosted at a regional European provider successfully
translated the specification, producing Go source, RPM spec, complete Debian
package layout, Containerfile, and macOS pkgbuild skeleton — the most
complete deliverable set of any single run. Key observations:
- Run quality was sensitive to prompt phrasing and file staging
- Extended reasoning produced better output than immediate generation
- Demonstrates the paradigm works without US cloud infrastructure
- Relevant for data sovereignty, regulated industry, and air-gapped deployment

**Prompt evolution across runs.**
The system prompt was refined iteratively through the test series:

| Version | Key addition | Trigger |
|---|---|---|
| v1 | Working directory | Files written to unexpected locations |
| v2 | "Write ALL to disk, do NOT output to terminal" | Models defaulting to inline output |
| v3 | "Do NOT write input files to disk again" | Models re-writing prompt files |
| v4 | Concrete filenames with role labels | Smaller models not resolving placeholders |
| v5 | "Before writing, state your plan" | Incomplete deliverable sets |
| v6 | "Verify files on disk after writing" | LLM-C extended reasoning self-verified unprompted |

**Translation report format evolution.**
The translation report emerged as a critical audit artifact through testing:
- Early runs: no report, or report only in terminal
- Mid runs: report written to disk but incomplete
- Best runs: full report with template constraints compliance table,
  per-example confidence levels, parsing approach, deviation documentation

The `TRANSLATION_REPORT.md` is now a required deliverable in the template
DELIVERABLES table as of v0.3.3.

### Convergent Spec Gaps (found by multiple models)

| Gap | Models | Fix |
|---|---|---|
| Line=1 for missing sections | All | RULE-01: explicit in v0.3.2 |
| list-templates needs companion template files | All | DEPLOYMENT note in v0.3.2 |
| SPDX list not provided | All | Build-time dependency; embedded at build time |
| Source filename convention unclear | Multiple | DELIVERABLES table in v0.3.3 |
| Delivery order not specified | Multiple | DELIVERABLES order in v0.3.3 |

### Divergent Spec Gaps (found by single model)

| Gap | Model | Fix |
|---|---|---|
| list_templates EXAMPLE missing header | LLM-C browser | Fixed in v0.3.2 |
| .md extension not enforced with exit 2 | LLM-A | PRECONDITIONS in v0.3.2 |
| Empty GIVEN block boundary ambiguous | LLM-C browser | RULE-07 tightened in v0.3.2 |
| strict mode not threaded to summary | LLM-A | Summary marked normative in v0.3.2 |
| AST requirement for robust parsing | LLM-D | Parsing note added in v0.3.3 |
| Signal handling skipped silently | All | Signal handling note in v0.3.3 |
| TRANSLATION_REPORT.md not required | All | Required deliverable in v0.3.3 |

### Infrastructure Lessons

**mcphost version matters.**
The `MultiContent is deprecated` warnings indicate an older SDK version.
Tool call reliability and large payload handling improve with SDK updates.
Recommend tracking mcphost version in the audit bundle metadata.

**max_tokens must be set high.**
32000 minimum for a complete pcdp-lint implementation including packaging.
Lower values cause JSON truncation in tool calls, producing silent failures.

**Filesystem MCP config must allow subdirectory creation.**
Models will attempt to create `debian/` subdirectory. MCP config must
permit this or the model will produce workarounds (flat files + setup script).

**Local model performance.**
LLM-F (30B, CPU-only) was extremely slow. For production use, local model
inference requires GPU acceleration. CPU-only inference is feasible for
testing and validation only. A 120B model (LLM-E) on GPU-accelerated
infrastructure at a regional EU provider performed comparably to frontier
models on deliverable completeness.

### Overall Assessment

The paradigm held across every environment and model tested. The universal
language resolution finding — eight runs, three continents, one correct
answer, all derived from the template — validates the core design claim.
Deliverable completeness and translation report quality varied by model
capability and delivery environment, but the specification structure
provided sufficient guidance for all tested models to produce a working
implementation. Spec gaps found during testing were small, targeted, and
directly fixable — exactly the behaviour expected from a first empirical
validation cycle.

---

## A.15 License Compliance and Software Composition Analysis

### The Limitation

No LLM can provide a legal guarantee that generated code is free of patterns
derived from differently-licensed training data. This is an unsolved problem
in the field. The `License:` META field and SPDX validation in `pcdp-lint`
are necessary but not sufficient for license compliance.

### What the Paradigm Provides

The paradigm is better positioned than generic AI coding assistants:

- The `License:` META field declares intent upfront, validated by `pcdp-lint`
- The translator receives an explicit license constraint and must acknowledge
  it in the translation report
- The audit bundle contains the generated source, making SCA scanning
  straightforward
- The translation report documents any known license-relevant deviations

### Recommended CI Pipeline Addition

For commercial and regulated deployments, add Software Composition Analysis
as a required step in the audit bundle pipeline, after code generation and
before deployment sign-off:

```
Specification → pcdp-lint → AI translator → generated code
  → SCA scan → audit bundle → human review → deployment
```

**Recommended tools:**

| Tool | Strength | Relevant for |
|---|---|---|
| REUSE (FSFE) | SPDX header enforcement per file | All projects; pairs with spec META |
| FOSSology | Deep license scanning, snippet detection | Regulated/commercial deployments |
| Black Duck | Enterprise SCA, policy enforcement | Large organisations |
| Snyk | Developer-friendly, CI integration | Rapid iteration pipelines |

**REUSE** is particularly relevant: it enforces SPDX license headers at
the file level, which pairs directly with the `License:` field in spec META.
An OBS package for REUSE is available for openSUSE and SLES. A REUSE
compliance check can be added to the audit bundle CI step with no additional
infrastructure.

### License Compatibility Quick Reference

| Generated code license | Can use Apache-2.0 libraries | Can use GPL-2.0-only libraries |
|---|---|---|
| GPL-2.0-only | Yes | Yes |
| Apache-2.0 | Yes | **No — derived work conflict** |
| CC-BY-4.0 | N/A (not for code) | N/A |
| Proprietary | Yes (check terms) | **No** |

### Guidance for Translators

When the spec declares `License: Apache-2.0` or any non-copyleft license,
the translator must:
- Avoid generating code that is structurally identical to known GPL-licensed
  implementations
- Document in the translation report any standard algorithms used that have
  well-known GPL implementations (e.g. specific sorting algorithms, parsers)
- Leave the SCA verification to the human reviewer and the CI pipeline

When the spec declares `License: GPL-2.0-only`, the translator has more
freedom — GPL-compatible code may be used freely.

---

---

## A.16 Large Projects — Partitioning, Interfaces, and Composition

### The Problem

Everything in PCDP v0.3.x assumes a single component with a single specification
file. Real software systems are composed of many components with defined interfaces
between them. A payment system is not one spec — it is an account service, a
transfer service, a ledger, an audit log, and a notification system, each with
precisely defined contracts between them.

The paradigm must scale to this reality. The architect role in PCDP is to
decompose a system into components, define the interfaces between them, and
specify the build order. Individual domain experts then author the per-component
specifications.

### Core Concepts

**Component Interface**
A spec may declare that it exposes a public interface — a set of types,
function signatures, and invariants that other components may depend on.
The interface is a subset of the spec's TYPES and BEHAVIOR sections,
explicitly marked as exported.

Interface declaration in a spec:

```markdown
## INTERFACE
EXPORTS:
  TYPES:
    - Account
    - TransferResult
    - ErrorCode
  BEHAVIOR:
    - transfer
  INVARIANTS:
    - GLOBAL: ∀ a: Account. a.balance >= 0
```

**Spec Import**
A spec may declare that it depends on another component's interface.
Imports are resolved at translation time — the translator receives both
the spec and all imported interface definitions.

Import declaration in a spec META:

```markdown
Imports:
  - account-service: ./account-service.md#INTERFACE
  - audit-log: ./audit-log.md#INTERFACE
```

The `#INTERFACE` fragment means "import only the exported interface,
not the full implementation spec."

**Project Manifest**
A top-level file (`pcdp-project.md`) that declares all components in a
system, their dependencies, build order, and system-level invariants.
The manifest is the architect's primary artifact in a multi-component project.
See Appendix A.17 for the `project-manifest` deployment template.

**System-Level Invariants**
Invariants that span multiple components belong in the project manifest,
not in individual component specs:

```markdown
## SYSTEM-INVARIANTS
- GLOBAL: Σ(all account balances) is conserved across all services
- GLOBAL: every transfer_id is globally unique across all instances
- GLOBAL: audit log entry exists for every state-changing operation
```

### Architectural Workflow for Large Projects

```
1. Architect authors pcdp-project.md
   Declares: components, dependencies, build order, system invariants

2. Architect authors interface specs (*.interface.md)
   Declares: exported types, function signatures, invariants
   Frozen before implementation specs are written

3. Domain experts author component specs (*.md)
   Each spec imports the interfaces it depends on
   Implementation specified against the imported interface contract

4. AI translates each component spec independently
   Import resolution provides full type information at translation time
   Generated code respects interface contracts by construction

5. pcdp-lint validates the full project
   Checks: all imports resolve, interface contracts consistent,
   no circular dependencies, system invariants present
```

### Divide and Conquer

The decomposition principle: **a component spec should be translatable
independently of all other components.** This means:

- Interfaces must be fully specified before implementation specs are written
- A component spec that imports an interface gets all type information
  it needs from that import — no implicit knowledge of other components
- The translator generating component A does not need to know anything
  about component B beyond B's exported interface

This mirrors the classical principle of programming to interfaces, not
implementations — applied at the specification level.

### Component Granularity

A useful heuristic for decomposition: **a component spec should fit
comfortably in a single LLM context window.** If a spec requires scrolling
past thousands of lines to understand, it should be split.

Practical guidance:
- A component handles one coherent domain concept (accounts, transfers, sessions)
- A component exposes a small, stable interface (3-10 functions)
- A component has at most 20-30 EXAMPLES
- If a spec has more than 5 BEHAVIOR sections, consider splitting

### Interface Versioning

Interfaces are versioned independently of their implementing components.
An interface version change that breaks existing importers is a major version
bump. Backwards-compatible additions are minor version bumps.

```markdown
## META
Deployment:  backend-service
Interface-Version: 1.2.0    ← version of the exported interface
Version:     2.4.1          ← version of this implementation spec
...
```

Importers declare the minimum interface version they require:

```markdown
Imports:
  - account-service: ./account-service.md#INTERFACE@>=1.2.0
```

### What This Means for pcdp-lint

`pcdp-lint` v1 validates single specs in isolation. A future version
(v2 scope) will validate full projects:

- Resolve all imports and check they exist
- Verify imported interface versions satisfy declared requirements
- Check for circular dependencies
- Validate system invariants reference only exported types and behaviors
- Produce a project-level audit bundle covering all components

---

## A.17 The mcp-server-pcdp: MCP Server for PCDP

### Motivation

The filesystem-based template system (TEMPLATE_DIR, systemd-style preset
layering) works well for single-developer use. For teams and enterprises,
a network-accessible MCP server is more practical:

- Templates and presets managed centrally, not per-machine
- Org-specific template customisations deployed without touching developer machines
- Any MCP-capable host (mcphost, VS Code, Claude Desktop, web UI) can
  access templates and validation without filesystem configuration
- Versioned template registry: `cli-tool@0.3.8` and `cli-tool@latest` simultaneously
- Access control: teams can have private templates not in the public registry

### Architecture

`mcp-server-pcdp` is a single MCP server exposing three tool groups:

```
mcp-server-pcdp
├── template/*
│   ├── list          list available templates with versions
│   ├── get           retrieve a template by name and version
│   └── resolve       resolve effective settings after preset layering
│
├── lint/*
│   ├── validate      run pcdp-lint rules against a spec, return diagnostics
│   └── check-project validate a full project manifest and all imports
│
└── project/*
    ├── get-interface  retrieve exported interface from a spec
    └── resolve-imports resolve all imports for a spec
```

### The LLM as Wizard

The spec-authoring guidance (what was previously conceived as `pcdp-wizard`)
is not a server responsibility — it is the LLM's role. An LLM connected to
`mcp-server-pcdp` via any MCP-capable host can:

1. Call `template/list` to show available deployment templates
2. Call `template/get` to retrieve the selected template
3. Ask the user the right questions based on the template's required fields
4. Call `lint/validate` to check the spec as it is being written
5. Iterate until lint passes

No wizard tool needed. The MCP server provides the data and validation layer;
the LLM provides the conversational layer. This works in mcphost, VS Code
with Copilot, Claude Desktop, or any other MCP-capable environment without
any changes to the server.

### Preset Layering via MCP

The server implements the same systemd-style layering as the filesystem
approach, but resolved server-side:

```
vendor defaults      built into server binary
org presets          /etc/mcp-server-pcdp/presets/
team presets         per-request context header
user presets         per-request context header
```

A client requesting `template/resolve` passes its context (org, team, user)
and receives the fully-merged effective settings — no client-side merging needed.

### Implementation

`mcp-server-pcdp` is specified using PCDP (`mcp-server-pcdp.md`) with
`Deployment: mcp-server`. Default implementation language: Go.
It is generated from that spec. Self-hosting applies.

The `mcp-server` deployment template (to be written in v0.3.9) defines
the full constraint set for MCP server components: transport (stdio / HTTP+SSE),
tool registration format, error response conventions, and packaging.

## Changelog

| Version | Date | Changes |
|---------|------|---------|
| 0.3.8 | 2026-03-19 | Added A.16: Large Projects — Partitioning, Interfaces, and Composition. Added A.17: mcp-server-pcdp MCP architecture. Dropped pcdp-wizard as standalone CLI — wizard behaviour is the LLM's role; mcp-server-pcdp provides the data layer. Added project-manifest and mcp-server to deployment template roadmap. |
| 0.3.7 | 2026-03-18 | Anonymized all LLM/vendor names in A.14. Removed version numbers from all internal filename references. |
| 0.3.6 | 2026-03-18 | crypto-library → verified-library. python-tool added. library-c-abi CPS note (CMake 4.3). |
| 0.3.5 | 2026-03-18 | Renamed spec-lint → pcdp-lint. post-coding paths → pcdp. Filename convention. Curly brace placeholders. |
| 0.3.4 | 2026-03-18 | CC-BY-4.0 for specs/templates, GPL-2.0-only for tools. A.15 SCA. |
| 0.3.3 | 2026-03-17 | Expanded A.13/A.14. translation_report/ in audit bundle. DELIVERABLES expanded. |
| 0.3.2 | 2026-03-17 | A.13 prompt. A.14 empirical tests. DELIVERABLES in template. Unified versioning. |
| 0.3.1 | 2026-03-17 | Changelog moved to end. A.12 industry landscape. BEHAVIOR/INTERNAL. |
| 0.3.0 | 2026-03-17 | Deployment template system. Target: removed from META. |
| 0.2.3 | 2026-02-10 | Workflow diagram. Dual-path architecture. |
| 0.2.1 | 2026-02-10 | Initial public draft. |

