# Post-Coding Development
## Human Intent, Machine Implementation

**Status:** Draft
**Version:** 0.3.22
**Author:** Matthias G. Eckermann <pcd@mailbox.org>
**Date:** 2026-04-13

---

## Executive Summary

Informally known as **Piccadilly** — *the place where intent becomes implementation.*

The **Post-Coding Development** paradigm (PCD) fundamentally changes how software is created. Domain experts write specifications in structured Markdown describing what a system should do. AI translates those specifications into complete, packaged, tested implementations. Engineers never write implementation code. If the output is wrong, they fix the specification and regenerate — they never touch the code.

This is not AI-assisted coding. In vibe coding, humans write code and AI suggests; the code remains the primary artifact. In PCD, humans write specifications and AI generates all implementation code; the specification is the primary artifact and the source of truth.

PCD enables AI-augmented development in safety-critical and regulated domains — automotive, aviation, medical devices, finance, government — that currently prohibit AI code generation because AI-suggested code cannot be audited. The paradigm resolves this by making the specification the audit point: human-authored, human-reviewed, validated by tooling, and cryptographically linked to every generated artifact. The spec is what you certify. The code is what you verify.

The project is open source, technology-agnostic, and self-hosting: the toolchain that validates and serves specifications was itself produced from PCD specifications, with zero hand-written implementation code.

At AWS re:Invent 2025, Amazon CTO Werner Vogels dedicated a significant portion of his closing keynote to spec-driven development, coining the phrase *"Specifications are the new code"* and introducing the concept of *"verification debt"* — the dangerous gap that opens when AI generates code faster than humans can comprehend it. PCD was conceived and developed independently; the convergence confirms the direction.

### Value Proposition at a Glance

| Category | Traditional Coding | Vibe Coding | Post-Coding Development |
|---|---|---|---|
| Human writes | implementation code | code + prompts | specifications only |
| AI role | none | suggests / completes | translates spec → code |
| Primary artifact | source code | source code | specification |
| Target language | developer decides | developer decides | deployment template |
| If output is wrong | edit the code | edit the code | fix the spec, regenerate |
| Regulated domains | manual audit | prohibited | enabled by design |
| Provenance | implicit | implicit | cryptographically verifiable |

---

## 1. Introduction

AI has made code synthesis cheap, but unstructured generation is brittle and unsuitable for regulated environments. PCD bridges this gap through a fundamental shift: humans write what a system should do; AI writes how to do it. The separation is strict and non-negotiable — a spec author never touches implementation code, and the universal translation prompt never touches implementation details. Both are enforced by design, not convention.

The workflow has four steps. First, a domain expert authors a specification in structured Markdown describing behaviors, types, invariants, and examples. The specification format is constrained: required sections, formal notation for types, executable examples that serve as the acceptance test. An interview prompt exists for domain experts who prefer to be guided through the process rather than authoring the format directly — any capable LLM, including small models running locally, can conduct the interview and produce the specification from their answers. Second, `pcd-lint` validates the specification structure before any translation begins; no valid specification, no code. Third, the spec author selects a deployment template — `cli-tool`, `mcp-server`, `backend-service`, `cloud-native`, and others — which encodes all implementation decisions: target language, compiler flags, packaging format, delivery phases. Fourth, an AI translator reads the specification and the template and produces a complete implementation: source code, packaging, documentation, man pages, test infrastructure, and an audit log.

Templates enforce the separation of intent from implementation. A specification describes what. A template describes how to build it for a specific deployment context. Neither the spec author nor the universal translation prompt ever encounters a language name, a compiler flag, or a packaging detail. These are encoded in the template and resolved automatically. This is not a convenience — it is an architectural guarantee that specifications remain stable across language changes, platform migrations, and toolchain evolution. The spec written today for a Go binary is the same spec that produces a Rust binary tomorrow if the template changes.

For large specifications — dozens of behaviors, thousands of lines — the milestone mechanism partitions translation into sequential passes. The first milestone is always a scaffold pass: all files, all types, all function signatures, all stubs, compile gate only. Subsequent milestones fill in real implementations one scoped set of behaviors at a time. The scaffold holds without modification through all milestones, giving every subsequent translator a stable foundation. The `sitar` system information tool — 35 behaviors, 2900-line specification — was translated to both Go and Rust in single sessions each using this pattern.

Formal verification through meta-languages (Lean 4, F*, Dafny) is a supported path for highest-assurance requirements but is not the default or the primary case. The paradigm works — and has been empirically validated — on the direct path: specification → AI → code, with validation through executable examples, independent test generation, and the audit bundle.

---

## 2. Goals

PCD has three primary goals and several derived goals that follow from them.

The primary goals are: enable AI code generation in regulated and safety-critical domains that currently prohibit it; shift the engineering role from writing implementation code to authoring precise specifications; and produce a cryptographically auditable chain of custody from specification to deployed artifact.

The regulatory goal is the most consequential. ISO 26262, DO-178C, IEC 62304, and Common Criteria require auditable development processes. AI-suggested code is opaque and fails these requirements. PCD resolves this by making the human-authored, human-reviewed specification the audit point, not the generated code. The specification is certified; the code is verified against it. Every generated artifact — source files, binary, RPM, DEB, container image — embeds the SHA256 hash of the specification it was produced from, creating a verifiable link that does not depend on trusting the build pipeline.

The engineering role goal requires a genuine shift in what organizations value. The expensive, difficult, intellectually demanding work in PCD is writing precise specifications — capturing behaviors, invariants, and examples accurately enough that the generated code is correct. This is harder than writing code; it requires architectural thinking, not implementation thinking. The cheap part is running the translator. A complete translation run for a well-specified component costs one LLM session. This asymmetry has a direct implication: when in doubt whether a specification change warrants full regeneration or incremental update, the answer is always full regeneration. The spec is the investment; the translator run is cheap.

From these primary goals follow: removing target language as a specification concern; supporting incremental adoption in mixed codebases; providing open, replaceable tooling with no vendor lock-in to any AI provider; and enabling digital sovereignty through local LLM compatibility.

---

## 3. Tenets

**Specifications are the source of truth.** Generated code is a build artifact. If the code is wrong, the specification is fixed and the code is regenerated — the code is never edited directly. Version control the specification; treat the generated code as you would treat a compiled binary.

**Templates enforce the separation of intent from implementation.** A specification describes behavior, types, invariants, and examples. A template describes how to realize that specification in a specific deployment context. The separation is absolute: no language name, compiler flag, or packaging detail appears in a specification. This keeps specifications stable and reusable across implementation changes.

**The spec is what you certify. The code is what you verify.** Human review, regulatory audit, and formal sign-off target the specification. Automated verification — compile gate, example tests, independent test generation, static analysis, SCA — targets the generated code. The 4-eyes principle, where required by Common Criteria or equivalent standards, applies to the specification review, not the code review. This makes certification tractable for AI-generated code in regulated domains.

**Provenance is cryptographically verifiable.** Every generated artifact embeds the SHA256 hash of the specification that produced it. The claim "this artifact was produced from this specification" is not asserted — it is verifiable by anyone with access to the specification file and the artifact. This closes the audit trail without depending on build system logs or human attestation.

**AI translation is probabilistic; verification is structural.** No specification format can eliminate the inherent uncertainty of LLM-based code generation. The paradigm addresses this through multiple complementary mechanisms: pcd-lint validates specification structure before translation; executable examples serve as acceptance tests; independent test generation by a second agent surfaces specification ambiguities; the translation report documents every decision and deviation; and the audit bundle provides full traceability. The goal is not to make AI translation deterministic — it is to make its outputs verifiable.

**Specifications have a lifecycle.** A specification is not written once and frozen. Requirements change, behaviors are added, types evolve. When a specification changes, the impact must be assessed: does this change affect the scaffold, the type system, or the interfaces? If so, full regeneration is required. If the change is isolated to the steps of one or two behaviors, incremental update is viable. The scaffold boundary is the decision point. Implementation decisions from prior translation runs that are not captured in the specification are recorded in a decisions hints file alongside the specification — this preserves architectural memory across runs without contaminating the specification itself.

**Digital sovereignty is a design goal.** Any capable LLM can translate a PCD specification. The paradigm has been empirically validated on frontier cloud models, 120B open-weight models hosted at EU providers, and 30B local models running on commodity hardware. The specification format, the templates, and the prompts are all open source. No part of the workflow requires a specific vendor.

**Openness and self-hosting validate the paradigm.** The specification format, deployment templates, prompts, and all tooling are published under open licenses. The toolchain itself — `pcd-lint`, `mcp-server-pcd` — was produced from PCD specifications with zero hand-written implementation code. A paradigm that cannot describe its own tooling has not yet proved its expressiveness.

---

## 4. State of the Business

The software industry is converging on specification-driven development. Werner Vogels named it at re:Invent 2025. GitHub Copilot, Cursor, and similar tools have demonstrated mass-market demand for AI code generation, while simultaneously surfacing its primary limitation: AI-generated code is opaque, non-auditable, and prohibited in regulated domains. The gap between what AI can produce and what regulated industries can accept is large and growing.

PCD occupies a position that no current commercial tool addresses: AI code generation that is auditable by design, suitable for regulated domains, and vendor-neutral. The critical differentiator is not technical sophistication — it is the explicit placement of the human review point. Every AI coding assistant places the human in the loop at the code level, which is exactly where regulators require human comprehension. PCD places the human in the loop at the specification level, which is where human comprehension is both possible and valuable, and treats the code as a generated artifact subject to automated verification.

The empirical record supports the paradigm's claims. The same specification produced working implementations in Go, Rust, Java, and Python in independent translation runs, with language chosen at translation time from the deployment template, not from the specification. A 35-behavior, 2900-line specification for a Linux system information tool was translated to both Go and Rust in single sessions each. The `pcd-lint` validator and the `mcp-server-pcd` MCP server were both produced from their own PCD specifications. Across all runs, every AI model tested resolved the target language from the deployment template without being told explicitly — the core design claim validated empirically across eight runs, three continents, and multiple model families.

The copyright and IP question for AI-generated code remains legally unsettled globally. PCD's position is honest: we do not know who owns AI-generated code, and neither does anyone else. What PCD does is maximize human control and make provenance explicit. The specification is human-authored. The chain of custody from specification to artifact is cryptographically verifiable. The audit bundle contains the full translation record. Software Composition Analysis tooling is recommended in the CI pipeline. This is the most defensible position available given the current state of law, and it is substantially stronger than any alternative AI coding approach.

---

## 5. Lessons Learned

The most important lesson is that the specification is harder to write than the code. This surprises engineers who expect the AI to do the hard work. The AI does the mechanical work — syntax, packaging, boilerplate, test infrastructure. The human does the conceptual work — what are the invariants? what are the error cases? what does "correct" mean for this behavior? Writing a specification that is precise enough to produce correct code on the first translation run requires the same quality of thinking as designing a system, not just describing it. This is a feature, not a bug: the intellectual work that matters is now captured in a human-readable, version-controlled, auditable artifact.

The target language must not be a specification concern. Early versions of this paradigm required spec authors to declare a target language. This was identified as an anti-pattern immediately: it pulls authors into implementation thinking and makes specifications brittle across language changes. Removing the language declaration from specifications and encoding it in deployment templates was the single most important design decision in the project's evolution.

Freeform prompts produce unreliable translations. The universal translation prompt is tightly constrained: it specifies delivery phases, compile gate requirements, stub contracts, and translation report format. The deployment template's EXECUTION section specifies the language-specific details. Neither document mentions the other's concerns. This separation took several iterations to get right, but it is what makes translations reproducible across different AI models and sessions.

The scaffold-first pattern is essential for large specifications. Without it, translators make different structural decisions across milestone runs, producing codebases with inconsistent package layouts and naming conventions. The scaffold milestone establishes the structure once; subsequent milestones fill it in. The scaffold held without modification through seven milestones in both Go and Rust translations of the `sitar` specification — exactly as designed.

LLM randomness creates consistency risk across multiple translation runs. Two runs from the same specification with the same model may make different architectural decisions. Over multiple incremental updates, a codebase can develop internal inconsistency. The decisions hints file was introduced to address this: implementation decisions not inferable from the specification are recorded alongside it, read by translators on subsequent guided regeneration or incremental update runs, and discarded when switching target languages. The spec stays clean; the implementation memory lives in a separate, disposable artifact.

The hard questions must be answered honestly. Copyright ownership of AI-generated code is legally unsettled. Memorisation and reproduction risk is non-zero with any LLM. Local model quality varies. None of these are reasons to avoid the paradigm; all of them are reasons to design it carefully. PCD's answers — explicit provenance, audit bundles, SCA tooling, spec hash embedding, human specification review as the certification point — are the most defensible available. Claiming certainty where none exists would undermine the credibility of the paradigm with exactly the regulated-domain audiences it is designed to serve.

---

## 6. Strategic Priorities

The toolchain is functional and self-hosting. `pcd-lint` validates specification structure against 17 rules. `mcp-server-pcd` serves templates, prompts, and hints to any MCP-capable LLM host, providing a complete translation environment in a single server connection. Nine deployment templates cover the primary deployment contexts. The empirical record spans multiple models, languages, and specification sizes. The foundation is solid.

The immediate priority is making the paradigm accessible. The specification format has a learning curve that the interview prompt partially addresses but does not eliminate. The `user-guide.md` covers the spec author workflow; the `technical-reference.md` covers the full toolchain. Making these documents complete, accurate, and usable for engineers who have not been part of the development process is the highest-leverage activity for adoption.

The specification lifecycle workflow needs tooling support. The `assess_change_impact` tool in `mcp-server-pcd` provides automated impact assessment — given a specification change, it recommends full regeneration or incremental update with reasoning. The `verify_spec_hash` tool checks whether generated artifacts are current with the specification. Together they close the loop on the lifecycle question: how do I know if my artifacts are still correct after a specification change? The answer is now automated and auditable.

Coding style and project integration deserve explicit guidance. When PCD-generated components are integrated into existing codebases with established conventions — OOP patterns, naming standards, framework idioms — the style hints file provides the mechanism. A `<project>.<language>.style.hints.md` alongside the specification encodes project-specific conventions; the translator reads it and produces conformant code. This is the correct answer to the objection that AI-generated code cannot adhere to coding philosophy: it can, given the right hints, and the hints mechanism is already in the framework.

The remaining templates — `gui-tool`, `python-tool`, `library-c-abi`, `verified-library`, `cloud-native`, `project-manifest` — need the spec hash embedding treatment applied to `cli-tool`, `mcp-server`, and `backend-service` this week. This is a mechanical update but a necessary one for consistency.

Publication timing is a strategic decision that remains open. The repository is private. The whitepaper, templates, and toolchain are complete enough for an informed audience to evaluate the paradigm. The decision to publish depends on organizational positioning considerations that are outside the scope of this document. What is clear is that the paradigm is ready to be evaluated — the self-hosting claim is real, the empirical record is solid, and the hard questions have honest answers.

---

# Appendix

## A.1 Constrained Specification Language

### The Translation Ambiguity Problem

**Critical weakness identified:** The gap between informal natural language specifications and formal code (Lean 4, F*, etc.) is where AI hallucination risk is highest. Free-form Markdown like "the transfer function should move money from one account to another and make sure accounts don't go negative" leaves massive room for misinterpretation.

**Solution:** Use **constrained specification language** with formal semantics embedded in Markdown structure. This reduces translation ambiguity by 80-90% compared to free-form natural language.

### Design Principles

**1. Required sections with machine validation:**
- Specifications must include: META, TYPES, BEHAVIOR, PRECONDITIONS, POSTCONDITIONS, INVARIANTS, EXAMPLES
- Optional but strongly recommended sections (v0.3.12+): INTERFACES, DEPENDENCIES
- Optional sections (v0.3.13+): TOOLCHAIN-CONSTRAINTS, DELIVERABLES (component-based)
- Every BEHAVIOR block must include a STEPS: ordered list (v0.3.12+)
- Every BEHAVIOR block may declare a Constraint: field — required | supported | forbidden (v0.3.13+)
- Any BEHAVIOR with error exits in STEPS must have at least one negative-path EXAMPLE (v0.3.13+, RULE-10)
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

STEPS:
  1. Validate preconditions; on failure return Err(appropriate ErrorCode).
  2. Begin atomic transaction.
  3. Debit from.balance by amount.
  4. Credit to.balance by amount.
  5. Write transfer_log entry with timestamp, from.id, to.id, amount, status.
  6. Commit transaction; on failure rollback and return Err(TRANSACTION_FAILED).

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
$ pcd-lint account_transfer.md

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
# pcd-lint.sh - Validate specification structure

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
  echo "       Use 'pcd-lint list-templates' to see available templates"
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

A `BEHAVIOR/INTERNAL` section describes implementation logic that is not directly user-facing — internal rules, algorithms, or sub-procedures invoked by a `BEHAVIOR`. Translators use `BEHAVIOR/INTERNAL` sections to generate private functions or methods. `pcd-lint` validates `BEHAVIOR/INTERNAL` sections with the same structural rules as `BEHAVIOR` sections.

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
2. pcd-lint validates structure and completeness
   pcd-lint resolves deployment template → target language
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

Every BEHAVIOR block must include PRECONDITIONS, STEPS, and POSTCONDITIONS.
STEPS are the algorithm — ordered, imperative, with explicit error exits.
A `MECHANISM:` annotation may follow any step where the implementation
pattern matters for correctness beyond what postconditions alone capture.

An optional `Constraint:` field classifies the behavior as `required`,
`supported`, or `forbidden`. If absent, the default is `required`.
A `forbidden` behavior must include a `reason:` line.

```markdown
## BEHAVIOR: transfer_funds
Constraint: required

INPUTS:
  from_account_id: AccountID
  to_account_id:   AccountID
  amount:          Amount
  transfer_id:     TransferID

PRECONDITIONS:
  - from_account exists and is active
  - to_account exists and is active
  - amount > 0
  - balance(from_account) >= amount
  - transfer_id not previously used

STEPS:
  1. Validate all preconditions; on failure return appropriate error immediately.
  2. Check transfer_id for idempotency; if already used return cached result.
  3. Begin SERIALIZABLE database transaction.
  4. Lock both accounts (consistent order to avoid deadlock).
  5. Debit from_account by amount.
  6. Credit to_account by amount.
  7. Insert transfer_record with transfer_id, timestamp, status=completed.
  8. Commit transaction.
     MECHANISM: use database SERIALIZABLE isolation, not application-level locking.
  9. Return TransferReceipt.

POSTCONDITIONS:
  - balance(from_account)_after = balance(from_account)_before - amount
  - balance(to_account)_after = balance(to_account)_before + amount
  - sum(all_balances)_after = sum(all_balances)_before
  - transfer_record created with transfer_id, timestamp, status=completed

ERRORS:
  - ERR_INSUFFICIENT_FUNDS if balance(from_account) < amount
  - ERR_ACCOUNT_NOT_FOUND if account does not exist
  - ERR_DUPLICATE_TRANSFER if transfer_id already used
  - ERR_ACCOUNT_FROZEN if account is not active
```

`pcd-lint` validates the `Constraint:` value (RULE-13) and records it in the
TRANSLATION_REPORT. Translators must not implement `forbidden` behaviors.
`supported` behaviors are implemented only if the resolved preset activates them.

**3a. Interfaces (optional, recommended for complex components)**

The INTERFACES section declares module boundary contracts — behavioural
interfaces that translators must implement, along with their required
test doubles. This section prevents translator discretion on abstraction
layer decisions and makes independent tests infrastructure-free.

```markdown
## INTERFACES

StorageAdapter {
  required-methods:
    GetAccount(ctx, id AccountID) → (Account, error)
    UpdateBalance(ctx, id AccountID, delta Amount) → error
    RecordTransfer(ctx, record TransferRecord) → error
    WithTransaction(ctx, fn func() error) → error
  implementations-required:
    production:  PostgresAdapter
    test-double: FakeAdapter {
      configurable fields: accounts map, recorded transfers
      WithTransaction: executes fn; on FakeAdapter.ForceTransactionErr → returns error
    }
}
```

**4. Invariants**

Invariants may be annotated with `[observable]` (verifiable by external
observation or the independent test suite) or `[implementation]` (verifiable
by code review or static analysis only). Mixing without annotation is
permitted but reduces audit utility.

```markdown
## INVARIANTS
- [observable]      sum(all_account_balances) is constant across all operations
- [observable]      all account balances >= 0 at all times
- [observable]      repeated calls with same transfer_id produce same result
- [implementation]  database transaction isolation is SERIALIZABLE, not READ COMMITTED
- [implementation]  account locks always acquired in consistent ascending-ID order
```

**4a. Dependencies (optional)**

The DEPENDENCIES section declares external library requirements and version
constraints. Translators are bound by these rules and must not fabricate
dependency versions.

```markdown
## DEPENDENCIES

github.com/lib/pq:
  minimum-version: v1.10.0
  rationale: PostgreSQL driver with context support

github.com/google/uuid:
  minimum-version: v1.3.0
  do-not-fabricate: true
```

**4b. Toolchain Constraints (optional)**

The `TOOLCHAIN-CONSTRAINTS` section declares spec-specific build and packaging
constraints beyond what the deployment template mandates. This covers OCI image
rules, generated file declarations, and forbidden toolchain patterns that are
necessary for a valid deliverable but are not expressible as generic template
rules. Entries use the `required | forbidden` vocabulary from the template table.

```markdown
## TOOLCHAIN-CONSTRAINTS

OCI:
  forbidden: HEALTHCHECK instruction in Containerfile
  required:  COPY --chmod=0755 for all installed binaries
  required:  multi-stage build; final stage FROM SLE-BCI or scratch

GENERATED-FILES:
  - zz_generated.deepcopy.go     // regenerate with controller-gen; do not hand-author
  - api/v1alpha1/groupversion_info.go
```

`pcd-lint` validates section structure as RULE-11. The translator reads this
section alongside the template DELIVERABLES and records compliance in the
TRANSLATION_REPORT.

**4c. Component-based Deliverables (optional)**

The `DELIVERABLES` section in a spec declares *logical component categories*,
not concrete filenames. Filenames are a template concern — the deployment
template's DELIVERABLES table maps logical components to language-specific
file names. This keeps specs language-agnostic.

```markdown
## DELIVERABLES

COMPONENT: api-types
  purpose: CRD type definitions for all custom resources
  required: true

COMPONENT: transport-layer
  purpose: Implements declared INTERFACES (production + test double + stub)
  required: true

COMPONENT: reconciler
  purpose: Implements all BEHAVIOR blocks marked Constraint: required
  required: true

COMPONENT: shared-helpers
  purpose: Shared constants, requeue durations, condition helpers, validators
  required: true
  note: Must be separate from reconciler to avoid circular imports
```

The deployment template maps each COMPONENT to concrete filenames for the
resolved language via a COMPONENT column in its DELIVERABLES table. Translators
may add additional files (e.g. `common.go`) provided they declare them in the
TRANSLATION_REPORT.

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

**Multi-pass EXAMPLES (v0.3.12+)**

For behaviours that span multiple invocations — reconcilers, retries, state
machines — EXAMPLES may contain multiple WHEN/THEN pairs. Each pair represents
one invocation. Intermediate states are normative and must be satisfied by the
implementation.

```markdown
EXAMPLE: graceful-stop-with-timeout
GIVEN:
  VM "testvm-01", spec.desiredState = Stopped, shutdownTimeout = "120s"
  Domain is Running on the remote host
  shutdownStartTime is unset

WHEN:  reconcile runs (pass 1)
THEN:
  domain.Shutdown() is called
  shutdownStartTime annotation is set to now()
  MECHANISM: do NOT block; requeue immediately
  result = RequeueAfter(10s)

WHEN:  reconcile runs (pass 2); elapsed < 120s; domain is Shutoff
THEN:
  status.phase = Stopped
  shutdownStartTime annotation is cleared
  result = RequeueAfter(60s)

WHEN:  reconcile runs (pass 2 alternate); elapsed >= 120s; domain is still Running
THEN:
  domain.Destroy() is called
  status.phase = Stopped
  status.conditions includes {type: ForcePowerOff, reason: ShutdownTimeout}
  result = RequeueAfter(60s)
```

Single-pass examples remain valid as a special case (one WHEN/THEN pair).
Multi-pass examples are identified by two or more WHEN/THEN pairs.

**Required negative-path EXAMPLES (v0.3.13+)**

Any BEHAVIOR block whose STEPS contain one or more error exits must include
at least one EXAMPLE whose THEN clause verifies the error outcome. This rule
applies specifically to BEHAVIORs involving mutual-exclusion validation,
timeouts, deletion/finalizer flows, retries, and unsupported conditions.
`pcd-lint` enforces this as RULE-10 (error, not warning).

```markdown
EXAMPLE: transfer_insufficient_funds
GIVEN:
  account_a = Account { id: 1, balance: 20 }
  account_b = Account { id: 2, balance: 50 }
WHEN:
  result = transfer(account_a, account_b, 30)
THEN:
  result = Err(INSUFFICIENT_FUNDS)
  account_a.balance = 20  // unchanged
  account_b.balance = 50  // unchanged
```

Happy-path EXAMPLES alone are not sufficient for any BEHAVIOR that declares
ERRORS or whose STEPS include `on failure →` exits.

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
│   ├── TRANSLATION_REPORT.md        # AI translator self-evaluation
│   └── translation-workflow.pikchr  # Pikchr workflow diagram (renders to SVG)
├── independent_tests/               # Optional: second-agent test cases
│   └── independent_test.go          # Tests generated by a second LLM
└── metadata.json                    # Traceability, hashes, versions
```

The `translation_report/` directory is a required artifact for any translated
component. In regulated domain contexts (Common Criteria, ISO 26262), the
translation report provides the human-reviewable record of what the AI
translator decided and why — the closest equivalent to a compiler log for
the AI translation step. It must be produced by the translator, written to
disk, and included in the audit bundle before submission for review.

The `translation-workflow.pikchr` diagram visualises the specific translation
run — inputs, decisions, outputs — as a machine-generated audit trail.
See Appendix A.18 for the diagram format and content requirements.

The `independent_tests/` directory is optional but strongly recommended for
safety-critical components. A second AI translator independently reads the
specification and generates test cases, without access to the primary
translator's implementation. These tests are run against the generated code;
failures indicate specification ambiguity or translation errors.
See Appendix A.18 for the second-agent test generation approach.

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
    "preset_source": "/etc/pcd/presets/org.toml"
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
- Install `pcd-lint` for specification validation.
- Select deployment templates applicable to the project domain.
- Configure org-level presets in `/etc/pcd/presets/`.
- No specification writing training required — domain experts use the
  interview prompt (`prompts/interview-prompt.md`) to produce specs via
  a guided AI conversation. The format is learned by the tool, not the human.

**Phase 1: AI-assisted specification alongside existing code**
- Domain expert runs the interview prompt with any capable LLM (including
  small models running locally — no cloud dependency required).
- LLM asks questions; expert answers in plain language; LLM produces the spec.
- Expert reviews the produced specification and corrects any misunderstandings.
- Run `pcd-lint` to validate structural correctness.
- Keep current manual implementation unchanged during this phase.
- Goal: Validate that domain experts can produce adequate specifications
  through the interview process, without learning the specification format.

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

**Reference implementation (`pcd-lint`): `GPL-2.0-only`**
The `pcd-lint` validator and any other reference tools are licensed under
GNU General Public License version 2 only. This follows the Linux kernel model:
companies that ship or modify `pcd-lint` must contribute their changes back.
This forces collaboration on the validation toolchain — the compliance layer
that must remain community-controlled and vendor-neutral for the paradigm to
be trustworthy in regulated markets. No single company can fork `pcd-lint`,
add proprietary validation rules, and use it to lock in customers.

The strategic rationale: the GPL-2.0-only reference implementation was a
significant factor in Linux's success — it forced competitors to work together
on the platform layer. The same dynamic applies here: interoperability on
the validator is more valuable than any single company's proprietary advantage.

**License compatibility note:**
- CC-BY-4.0 specifications may be implemented by GPL-2.0-only, Apache-2.0,
  proprietary, or any other licensed tools without conflict.
- GPL-2.0-only `pcd-lint` may validate Apache-2.0, CC-BY-4.0, and other
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
4. Build `pcd-lint` as reference implementation under the paradigm itself.
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
  ↓ [pcd-lint validates + deployment template resolves target language]
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
| `spack-package` | Python | — | Spack declarative DSL; `spack audit` compliance required; no EXECUTION phase | Declarative specification target. Output is `package.py`, not an executable. Spack 1.1.1 baseline. No man page. |

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

Valid values include any language with a functioning compiler/interpreter: COBOL, Fortran, PHP, Python, Perl, Ruby, Java, C, C++, Go, Rust, and others. The toolchain must be able to generate code compatible with the existing codebase. The pcd-lint tool will warn if no translator backend exists for the declared language.

### Preset Layering (systemd-style)

Presets follow a layered override model identical in principle to systemd's unit file loading. Later layers override earlier ones; the first match for any given setting wins in reverse order.

```
/usr/share/pcd/templates/        # shipped deployment template definitions (read-only)
/usr/share/pcd/presets/          # shipped vendor/community presets (read-only)
/usr/share/pcd/hints/            # shipped library hints files (read-only)
/etc/pcd/presets/                # system administrator overrides
/etc/pcd/hints/                  # system administrator hints
~/.config/pcd/presets/           # user-level overrides
~/.config/pcd/hints/             # user-level hints
<project-dir>/.pcd/presets/      # project-local overrides (committed to git)
<project-dir>/.pcd/hints/        # project-local hints (committed to git)
```

### Template Search Path

Deployment template files (`*.template.md`) are located at runtime using
a compile-time variable `TEMPLATE_DIR`. This ensures no runtime path
discovery and no environment variable magic — the path is baked in at
build time, consistent with supply chain security requirements.

```
TEMPLATE_DIR (compile-time default, read-only)     /usr/share/pcd/templates/
/etc/pcd/templates/                         system administrator additions
~/.config/pcd/templates/                    user additions
<project-dir>/.pcd/templates/               project-local additions
```

Later entries take precedence. A template file found in a later path
overrides one with the same name in an earlier path, allowing
organisations and projects to ship custom or overriding templates
without modifying the system-level installation.

The `TEMPLATE_DIR` default for Linux OBS packages is
`/usr/share/pcd/templates/`. Platform defaults for macOS
and Windows are deferred to v2.

For OBS packaging, `TEMPLATE_DIR` must be set at build time:
```
%build
make build TEMPLATE_DIR=/usr/share/pcd/templates/
```

Example preset file (`/etc/pcd/presets/suse.toml`):

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

Example project-local override (`.pcd/presets/project.toml`):

```toml
[templates.cli-tool]
default_language = "rust"
# This project uses Rust for the CLI layer; overrides org default of Go
```

### Library Hints Files

A `hints/` directory sits alongside `presets/` in the preset hierarchy.
Hints files contain library-specific implementation knowledge that belongs
neither in the spec (which must be language-agnostic) nor in the template
(which covers language and deployment conventions, not library internals).

**File naming convention:**
```
<template>.<language>.<library-name>.hints.md
```

Example:
```
/usr/share/pcd/hints/cloud-native.go.go-libvirt.hints.md
```

The translator reads all matching hints files after template resolution,
before generating code. Hints files are:
- **Advisory only** — cannot override spec invariants or template constraints
- **Version-tagged** in their META section
- **Maintainable independently** as libraries evolve, without touching specs or templates

Example hints file content:
```markdown
# cloud-native · Go · github.com/digitalocean/go-libvirt

## META
Version: 1.0.0
Library: github.com/digitalocean/go-libvirt
Language: Go
Template: cloud-native

## Version selection
This module has no tagged releases. Use a verified pseudo-version.
DO NOT fabricate commit hashes or timestamps.
Verification: git ls-remote https://github.com/digitalocean/go-libvirt.git HEAD

## API shapes
- DomainGetInfo returns 6 individual values: (state uint8, maxMem uint64,
  memory uint64, nrVirtCPU uint16, cpuTime uint64, err error)
  NOT a struct, despite DomainGetInfoRet existing as a type.
- ConnectGetMaxVcpus requires libvirt.OptString{"kvm"}, not a plain string.
- ConnectGetLibVersion returns uint64, not uint32.
```

The DEPENDENCIES section of a spec (Finding 6) may reference a hints file
by name, establishing a traceable link between the dependency declaration
and its authoritative API documentation.

### Type Bindings

Deployment templates may declare a `TYPE-BINDINGS` table that maps logical
spec types to concrete language types for the resolved language. This is the
correct location for platform-specific type mappings — the spec declares
`Duration`, `Condition`, or `Timestamp` as logical types; the template maps
them to the ecosystem's canonical types once the language is resolved.

This preserves the spec's language-agnosticism (the spec never names
`metav1.Duration`) while ensuring all translators use the same concrete type,
eliminating the interoperability divergence that arises from translator
discretion.

```markdown
## TYPE-BINDINGS

| Spec Type  | LANGUAGE=Go                                   | Notes                          |
|------------|-----------------------------------------------|--------------------------------|
| Duration   | metav1.Duration (k8s.io/apimachinery/meta/v1) | Serialises as "60s", "5m" etc. |
| Timestamp  | metav1.Time                                   |                                |
| Condition  | metav1.Condition                              | Use standard k8s condition type|
| List<T>    | []T                                           |                                |
```

The translator reads the TYPE-BINDINGS table after resolving LANGUAGE and
applies it mechanically to every TYPES occurrence. A binding row for a logical
type *overrides translator discretion* for that type when the matching LANGUAGE
is active. The spec author is never involved — TYPE-BINDINGS live in the
template, not the spec.

The `cloud-native.template.md` carries the Kubernetes ecosystem TYPE-BINDINGS.
Other templates define their own bindings appropriate to their ecosystem.

---

### pcd-lint as Reference Implementation

`pcd-lint` is the first component to be specified and generated under this paradigm. This serves two purposes:

1. **Empirical validation:** If the paradigm cannot describe its own tooling unambiguously, the template design is incomplete. Any gap discovered during `pcd-lint` specification authoring feeds back directly into template and schema improvements.

2. **Demonstrable credibility:** A specification system that can specify itself is a stronger argument than any hand-picked external example.

The `pcd-lint` specification uses:

```markdown
## META
Deployment: cli-tool
Verification: none
Safety-Level: QM
```

Target language resolves to Go (system default for `cli-tool`), unless overridden by org preset. The spec author — in this case, the project itself — does not need to decide or declare a target language.

The `pcd-lint` specification will be developed as the next working artifact, with the template schema defined above as its validation schema.

### Open Questions for v0.4.0

The following questions were deferred to keep v0.3.0 focused:

1. **Interface compatibility for enhance-existing:** When enhancing existing code, should the spec declare the interface type (extend-module, replace-function, add-endpoint, ffi-wrapper)? Deferred — start with `Language:` declaration only and add interface typing when the first real enhance-existing use case is attempted.

2. **Verification level in templates:** Should templates specify a default verification level (none / lean4 / fstar) or leave this always to the spec author? Current position: verification level stays with the spec author as an explicit architectural decision. Templates may recommend but not mandate.

3. **Template versioning:** How are breaking changes to deployment templates handled when a spec was authored against an older template version? Requires a template version pinning mechanism, likely in the audit bundle metadata.

---

## A.12 Related Work and Industry Landscape

### Overview

The Post-Coding Development combines several established ideas in a novel way. Each individual ingredient has precedent; the combination does not exist as a productised, accessible, regulated-domain-ready system.

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

**AWS Kiro and Spec-Driven Development (Industry, 2025)**
At AWS re:Invent 2025, Amazon CTO Werner Vogels dedicated his final keynote to spec-driven development as the answer to what he called *verification debt*—the gap between AI code generation speed and human comprehension speed. AWS engineer Clare Liguori demonstrated Kiro, an IDE built around writing specifications before AI generates code, reporting roughly a 2× improvement in shipping time over unstructured AI coding. Vogels explicitly positioned specifications against "vibe coding," citing Dijkstra's formal specifications and the Apollo Guidance System as historical precedents \[Vogels2025\].

Key differences from Post-Coding Development: Kiro is a proprietary, IDE-integrated, AWS-hosted product. It does not define a portable, lintable specification format, has no deployment template abstraction, no formal verification path, no supply chain or packaging conventions, and no pathway to regulated-domain certification. The paradigms are complementary rather than competing. The convergence of Werner Vogels' final keynote with the core thesis of this work—independently developed—is notable external validation of the problem framing.

**Structured-Prompt-Driven Development (SPDD), Thoughtworks (Industry, 2026)**
On 28 April 2026, Wei Zhang and Jessie Xia of Thoughtworks published *Structured-Prompt-Driven Development*, a method that treats prompts as version-controlled, reviewable, reusable engineering artifacts. SPDD organises prompts through a seven-part *REASONS Canvas* (Requirements, Entities, Approach, Structure, Operations, Norms, Safeguards) and ships an open-source CLI (`openspdd`) that orchestrates the workflow from analysis through code generation. The core invariant — *"when reality diverges, fix the prompt first, then update the code"* — is shared with PCD \[Zhang2026\].

Key differences from Post-Coding Development: SPDD permits hand-editing of generated code under an explicit *refactor-then-sync* path that flows changes from code back to the prompt; its Operations section specifies method signatures and parameter types, making the prompt language-bound rather than language-neutral; and it does not address formal verification, supply-chain provenance, or regulated-domain certification. Birgitta Böckeler categorises SPDD as a *spec-anchored* approach; PCD is *spec-only*. The two paradigms target different domains and are complementary: SPDD targets enterprise IT delivery where prompts and code evolve together; PCD targets regulated and long-lived components where the specification must remain the sole audit point. The convergence — SPDD, AWS Kiro, and PCD all arriving at structured-spec-before-code in the same window — is independent confirmation of the problem framing.

### A Note on the Name

The term **Post-Coding Development** is occasionally used informally in SDLC
contexts to mean "the activities that happen after the coding phase" — testing,
deployment, and maintenance. That usage is descriptive and unpublished; it has
not produced a named methodology, a body of literature, or an abbreviation.

This paradigm uses **Post-Coding Development** as a proper noun with a
deliberately inverted meaning: *coding is no longer the central human activity*.
The human role shifts from writing implementation code to authoring specifications.
Code becomes a generated output — a compiler artefact — rather than the primary
work product. The name marks that shift, not a phase that follows coding.

The informal alias **Piccadilly** — *the place where intent becomes
implementation* — further distinguishes the project identity and is used
throughout the tooling and documentation.

### Comparative Summary


| Approach | Human writes | AI layer | Formal verification | Regulated domains | Deployment templates |
|---|---|---|---|---|---|
| OpenAPI / AsyncAPI | Structured schema | Code-gen tools | No | No | No |
| Gherkin / BDD | Natural language tests | No | No | No | No |
| TLA+ / Alloy | Formal spec language | No | Yes (model checking) | Partial | No |
| F* / HACL* | F* code | No | Yes (dependent types) | Yes (crypto) | No |
| Dafny | Dafny code | No | Yes (SMT) | No | No |
| LLM code generation | Prompts + code | Yes | No | **Prohibited** | No |
| AWS Kiro | IDE-guided spec | Yes | No | No | No |
| Thoughtworks SPDD | Structured prompt (REASONS Canvas) | Yes | No | No | No |
| Program synthesis (research) | Formal spec language | Partial | Yes | No | No |
| **Post-Coding Development** | **Structured natural language** | **Yes** | **Optional, pluggable** | **Yes (primary target)** | **Yes** |

### What Is Genuinely Novel

The combination that does not exist elsewhere:

1. **Natural language as the primary artifact** — not a programming language, not a formal language, not a visual editor. Structured Markdown that domain experts can write without programming training.

2. **Deployment templates as a first-class concept** — target language is not a human decision; it is derived from deployment context. Specifications are technology-agnostic and stable across language ecosystem changes.

3. **Formal verification as optional and pluggable** — teams choose verification level based on risk and regulatory context. The same specification format works for a quick CLI tool (no verification) and a safety-critical automotive function (Lean 4 verified path).

4. **Regulated-domain certification as a design goal** — audit bundles, traceability matrices, and formal proofs are first-class outputs, not afterthoughts. The paradigm is designed to satisfy ISO 26262, DO-178C, IEC 62304, and Common Criteria requirements.

5. **Self-hosting from the first artifact** — the `pcd-lint` validator is developed using its own specification format. This is not a theoretical claim; it is an empirical test run from the beginning.

### Academic Framing

For readers from a formal methods or programming languages background: this paradigm sits at the intersection of *specification-driven development*, *correct-by-construction synthesis*, and *AI-mediated program translation*. The novelty is not in any individual technique but in the system design that makes these accessible to domain experts in regulated industries — and in the deployment template abstraction that decouples specification authoring from implementation technology choices.

---

## A.13 Recommended Translator Prompt

### Architecture: Two-Layer Prompt Design

The translator prompt system uses a two-layer architecture introduced in
v0.3.16:

```
prompts/prompt.md               ← Universal principles (language-agnostic)
                                  Valid for any deployment type.
                                  Delegates execution recipe to the template.

templates/<n>.template.md       ← Template-specific execution recipe
  ## EXECUTION                    Delivery phases in order.
    ### Input files               Resume logic for partial runs.
    ### Delivery phases           Compile gate with language-specific commands.
    ### Resume logic              Build gate (for templates with container builds).
    ### Compile gate
```

**Why this split?**

The previous design embedded language-specific instructions (Go commands,
file names, phase ordering) in the generic prompt. This created two problems:

1. A prompt for a Go CLI tool contained references to `go mod tidy` and
   `go build ./...` — wrong for a Python tool or a C library.
2. Adding a new deployment type required changing the shared prompt, risking
   regressions in existing translations.

With the two-layer design, the generic prompt contains only universal
principles. Everything language- or deployment-specific lives in the
template's `## EXECUTION` section. A new deployment type gets its own
execution recipe without touching the shared prompt.

**pcd-lint validates** that every deployment template (files with
`Deployment: template` in META) contains a `## EXECUTION` section with
the required subsections. See RULE-14.

---

### The Universal Prompt (`prompts/prompt.md`)

The universal prompt covers:

- Deriving the target language from the deployment template
- Delegating phase ordering and compile gate to the template's EXECUTION section
- Applying TYPE-BINDINGS, GENERATED-FILE-BINDINGS, STEPS, Constraint: fields
- Implementing INTERFACES declarations and test doubles
- Not fabricating dependency versions
- Delivery modes (filesystem, in-memory, inline)
- Translation report requirements including the confidence table

The prompt does not name any programming language, build tool, or
deployment-specific file. It is reproduced in full at `prompts/prompt.md`
in the repository.

---

### The EXECUTION Section in Templates

Every deployment template contains a `## EXECUTION` section with four
required subsections:

**`### Input files`** — lists what the translator receives in the working
directory. Clarifies which files to read before generating code (especially
important for hints files).

**`### Resume logic`** — instructs the translator to check for existing
files before writing, enabling partial run recovery without rewriting
completed phases.

**`### Delivery phases`** — the ordered list of what to produce and when.
This replaces the phase numbering that previously lived in the prompt.
Each template defines its own phases appropriate to its deployment type.
TRANSLATION_REPORT.md is always the last phase.

**`### Compile gate`** — the language-specific build verification steps:
dependency resolution (`go mod tidy`), compilation (`go build ./...`),
and any build-tool-specific steps (container builds for cloud-native,
OCI image verification for mcp-server). If the translator cannot execute
shell commands, it must document this explicitly — silent omission is
not permitted.

---

### System Prompt (mcphost config)

Set once per invocation. Replace placeholders with actual filenames.

```
Your working directory is <target-dir>/
The following input files are present there:
  - <deployment-template>.template.md  — the deployment template
  - <spec-name>.md                     — the component specification
Do NOT output file contents to the terminal.
Do NOT write the input files to disk again.
Write ALL output files to disk using the filesystem write tool.
After writing each file, confirm the filename written.
```

For hints files, add a line per file:
```
  - <template>.<language>.<library>.hints.md  — library hints for <library>
```

**Note for small models:** replace the placeholders in the system prompt
with the actual filenames. Small models do not reliably resolve
angle-bracket placeholders. See `prompts/README-small-models.md`.

---

### Prompt Design Rationale

**Template owns the execution recipe.** A translator reading a single
template gets everything it needs: deliverables, phase order, compile gate,
resume logic. No cross-referencing required. A new template author writes
one self-contained file.

**Concrete filenames in the system prompt.** Smaller models do not
reliably resolve angle-bracket placeholders to actual files. The system
prompt must name files explicitly.

**Resume logic in the template.** Partial run recovery is deployment-specific
because different templates produce different file sets. The generic prompt
cannot know which files are expected for a cloud-native component vs a
CLI tool.

**Compile gate in the template.** `go mod tidy` is a Go instruction. A
Python tool uses `pip install` or `uv sync`. A C library uses `make`.
These belong in the template, not in a generic prompt.

**"Do not ask clarifying questions"** — forces ambiguities into the
translation report. This is the feedback mechanism for spec improvement.

---

## A.14 Empirical Testing: pcd-lint

In March 2026, the `pcd-lint` specification and `cli-tool` deployment template
were submitted to multiple LLMs across different environments as an empirical
test of the paradigm. This appendix documents all test runs and findings.

LLM identities are anonymized. The test methodology and findings are what
matter — not which commercial product produced them. Labels reflect
capability class and environment rather than vendor.

### Test Configuration

- **Specification:** `pcd-lint.md`
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
| LLM-G | Small, API-accessible | US cloud | API (direct) |

### Universal Finding: Language Resolution

**Every model tested resolved Go as the target language by reading the
deployment template, without being told explicitly.** All cited the
template's `LANGUAGE | Go | default` entry as the source of their decision.
This is the paradigm's core claim — and it held across all tested models,
environments, and prompt versions.

### Deliverable Completeness by Run

| | LLM-A | LLM-B | LLM-C browser | LLM-C mcphost v1 | LLM-C mcphost v2 | LLM-D | LLM-E best | LLM-F | LLM-G |
|---|---|---|---|---|---|---|---|---|
| main.go / all 14 rules | ~6/7 | 7/7 | 7/7 | 7/7 | 7/7 | scaffold | 7/7 | TBD | **14/14** |
| RPM spec | No | No | Yes | Yes | Yes | Yes | Yes | TBD | Yes |
| DEB complete | No | No | workaround | workaround | **Yes proper** | Yes | Yes | TBD | Yes |
| Containerfile | ? | ? | No | No | No | No | **Yes** | TBD | No |
| Makefile | ? | ? | Yes | Yes | Yes | ? | Yes | TBD | Yes |
| README with OBS | No | No | ? | ? | **Yes** | ? | ? | TBD | Yes |
| LICENSE | No | No | No | No | No | No | Yes | TBD | Yes |
| Report to disk | No | No | No | No | **Yes** | **Yes** | Partial | TBD | Yes |
| Template constraints table | No | No | No | No | **Yes** | No | No | TBD | Yes |
| Confidence calibration | Good 85% | Poor 100% | Good 94% | Poor 100% | Good 90-95% | **Excellent** | Poor 100% | TBD | Good 90% |

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

### LLM-G: Small model, API-only, complete run

On 2026-03-25, `pcd-lint.md` v0.3.13 and `cli-tool.template.md` v0.3.13 were
submitted to a small frontier model via the Anthropic API — using the new KIT
tool with filesystem access and the ability to run test builds.  The model
delivered all source and packaging artifacts inline, as clearly separated files
in the response.

**Configuration:**
- Model: small frontier model (anonymized as LLM-G)
- Interface: direct API call, Agentic tools, no MCP servers
- Prompt: two-layer prompt (generic `prompts/prompt.md` + template EXECUTION section)
- Delivery mode: direct to disk (filesystem access)

**Results:**
- All 14 validation rules implemented and passing
- 15/15 tests passing
- Full compile gate: `go mod tidy` + `go build ./...` PASS
- Static binary: 2.9 MB, no runtime dependencies
- RPM spec, Makefile, README with OBS install instructions
- LICENSE (reference), TRANSLATION_REPORT.md with template constraints table
- Confidence calibration: good (~90%) — most examples rated correctly
- Run time: fast 

**Key observations:**

*Correct rule numbering.* The implementation handles all rules including
RULE-11 (TOOLCHAIN-CONSTRAINTS) and RULE-12 (cross-section consistency),
though the deliverables summary skipped mentioning them explicitly —
an indexing gap in the report, not an implementation gap.

*CLI style correct.* All commands use bare-word and key=value syntax
(`pcd-lint strict=true myspec.md`, `pcd-lint list-templates`) with no
`--flags`. One slip in INDEX.md (`pcd-lint help`) did not affect the
implementation.

*SPDX compound expressions.* The `Apache-2.0 OR MIT` compound expression
was handled correctly — a detail that requires reading the SPDX spec, not
just matching a fixed list.

*Fenced code block isolation.* `EXAMPLE: fake` inside a fenced block does
not trigger RULE-07. The line-by-line state machine with fence-depth counter
correctly ignores structural markers inside code fences.

*Two-layer prompt validation.* This is the first complete run using the
v0.3.16 two-layer prompt architecture (generic prompt + template EXECUTION
section). The model followed the EXECUTION section phase ordering without
any language-specific instructions in the generic prompt. The architecture
worked as designed.

**Significance:** Demonstrates that a small frontier model, via direct API
with no MCP infrastructure, can produce a complete, tested, packaging-ready
PCD implementation. The barrier to entry for PCD translation is lower than
the earlier test series suggested — capable models do not require extended
context windows, filesystem MCP servers, or extended reasoning modes.

### Infrastructure Lessons


**mcphost version matters.**
The `MultiContent is deprecated` warnings indicate an older SDK version.
Tool call reliability and large payload handling improve with SDK updates.
Recommend tracking mcphost version in the audit bundle metadata.

**max_tokens must be set high.**
16384 minimum for a complete pcd-lint implementation including packaging.
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

### LLM-H: v0.3.21 regeneration runs (2026-04-07)

On 2026-04-07, both `pcd-lint` and `mcp-server-pcd` were regenerated from
their updated v0.3.21 specifications using Claude Sonnet 4.5 and Claude Haiku.
This was the first empirical test of the milestone mechanism, RULE-15/16/17,
and the `set_milestone_status` tool in a real translation run.

#### pcd-lint v0.3.21

Three approaches were tested in parallel: Haiku fresh, Sonnet incremental
(update existing code), and Sonnet fresh.

| Approach | Model | Result | Rules |
|---|---|---|---|
| Fresh translation | Haiku | ✓ compile + self-validate | 13/17 |
| Incremental update | Sonnet 4.5 | ✗ stalled (output length) | — |
| Fresh translation | Sonnet 4.5 | ✓ **chosen** | **17/17** |

**Haiku** succeeded on a fresh translation: the binary compiled, passed
pcd-lint self-validation, and validated a real deployment template. However,
RULE-10 (negative-path example detection), RULE-11 (TOOLCHAIN-CONSTRAINTS),
RULE-12 (cross-section consistency), and RULE-16 (MILESTONE BEHAVIOR names)
were either missing or stub-only. Haiku stayed within the provided input
files without prompting.

**Sonnet incremental** stalled with max_tokens set to 16384 — the existing
code plus all new rule implementations exceeded the output window. Raising
max_tokens to 128000 (the hard API ceiling for Sonnet 4.5/4.6) and disabling
compaction resolved the stall.

**Sonnet fresh** at 128000 tokens implemented all 17 rules. The key
architectural decision — extracting the rule engine into `internal/lint/lint.go`
to allow `mcp-server-pcd` to import it as a library — was made independently
by both Haiku and Sonnet, converging on the same package structure. Sonnet's
`internal/lint/lint.go` matches the DELIVERABLES declaration in the spec
and is importable by external packages; Haiku's `rules.go` in the main
package is not. Sonnet was chosen.

#### mcp-server-pcd v0.2.0

Tested in two rounds. Round 1: input contained only the spec, one hints file,
and one template. Round 2: full input set (all 9 templates, all 6 hints files,
all 3 prompts).

**Round 1 finding:** without the full asset set in the input directory, all
translators produced placeholder or partial embedded assets. Asset embedding
is driven by what is present at translation time — the spec cannot substitute
for missing input files.

**Round 2 results:**

| Approach | Model | Embedded assets | Notes |
|---|---|---|---|
| Fresh | Haiku | 4 of 18 | 13 session noise files (COMPLETION_SUMMARY.txt etc.) |
| Fresh | Sonnet 4.5 | **18 of 18** ✓ | **chosen** |
| Incremental | Sonnet 4.5 | 0 (placeholders) | Correct module path; logic updated correctly |

**Sonnet fresh** correctly embedded all 18 assets (9 templates, 6 hints
files, 3 prompts including the new `reverse-prompt.md`). The `set_milestone_status`
tool and `internal/milestone/` package were implemented correctly. The
`reverse` prompt resource (`pcd://prompts/reverse`) was registered.

**Haiku** produced a functional binary but embedded only 4 assets — the
files that happened to be in the input directory at translation time. It
also generated 13 housekeeping files (`COMPLETION_SUMMARY.txt`,
`DELIVERY_MANIFEST.txt`, `ENHANCEMENT_REPORT.md`, `START_HERE.md` etc.)
that are agentic session artefacts, not spec deliverables. These are
disqualifying for a public repository.

**Sonnet incremental** correctly updated the rule logic and preserved the
existing module path (`github.com/mge1512/mcp-server-pcd`) but produced
placeholder assets in both rounds. Fresh translation is required for
mcp-server-pcd due to the asset staging complexity.

#### Infrastructure findings (2026-04-07)

**max_tokens ceiling.** Sonnet 4.5/4.6 has a hard API limit of 128000 output
tokens. Requests with max_tokens > 128000 return HTTP 400. For complete
mcp-server-pcd translations, 128000 tokens is sufficient but tight.

**Filesystem restriction is essential for clean runs.** Sonnet in agentic
mode will explore the filesystem and read existing code if not constrained.
Restricting the working directory to `/tmp/` prevented contamination. Haiku
stayed within the provided input files without prompting.

**No root access constraint.** Sonnet attempted to install Go module
dependencies via system package manager on the first run. Adding an explicit
"no root access; use go mod vendor" constraint in the system prompt resolved
this. All subsequent runs used user-level GOPATH/GOCACHE correctly.

**URL references in system prompt are fetched.** Providing the canonical
GitHub URL to `pcd-lint/code/internal/lint/lint.go` caused Sonnet to fetch
and read the live file. This is the correct approach for sharing library
code between tools without violating Go's `internal/` package import rules.

**Module path is a systematic translator gap.** No translator can determine
the author's GitHub username from the spec alone. The `Author:` field
provides an email address but not a repository path. Both Sonnet runs
produced incorrect module paths (`pcd-tools/pcd-lint`, `mcp-server-pcd`)
that required a post-generation fix commit. This is a candidate for a new
META field in a future schema version.

**Commit transparency pattern.** Both tools were pushed with a three-commit
history: (1) original baseline code, (2) AI-generated translation with
model credited in the commit message, (3) module path correction. This
provides full provenance — human-written baseline, AI translation, and
the single manual correction are all visible and attributable.

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
in the field. The `License:` META field and SPDX validation in `pcd-lint`
are necessary but not sufficient for license compliance.

### What the Paradigm Provides

The paradigm is better positioned than generic AI coding assistants:

- The `License:` META field declares intent upfront, validated by `pcd-lint`
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
Specification → pcd-lint → AI translator → generated code
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

Everything in PCD v0.3.x assumes a single component with a single specification
file. Real software systems are composed of many components with defined interfaces
between them. A payment system is not one spec — it is an account service, a
transfer service, a ledger, an audit log, and a notification system, each with
precisely defined contracts between them.

The paradigm must scale to this reality. The architect role in PCD is to
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
A top-level file (`pcd-project.md`) that declares all components in a
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
1. Architect authors pcd-project.md
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

5. pcd-lint validates the full project
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

### What This Means for pcd-lint

`pcd-lint` v1 validates single specs in isolation. A future version
(v2 scope) will validate full projects:

- Resolve all imports and check they exist
- Verify imported interface versions satisfy declared requirements
- Check for circular dependencies
- Validate system invariants reference only exported types and behaviors
- Produce a project-level audit bundle covering all components

---

## A.17 The mcp-server-pcd: MCP Server for PCD

### Motivation

The filesystem-based template system works well for single-developer use.
`mcp-server-pcd` makes the full PCD toolchain accessible to any MCP-capable
LLM host — no local file copies of templates, prompts, or hints required.
The LLM connects to the server and has everything it needs in one session.

Benefits over filesystem-only access:

- Templates, prompts, and hints served centrally — no per-machine installation
- Any MCP-capable host (mcphost, Claude Desktop, VS Code, KIT, custom agents)
  connects with a single config entry
- Lint validation returns structured diagnostics the LLM can act on directly
- Self-describing: LLM hosts can browse the full resource space via MCP native
  resource protocol without prior knowledge of URI patterns

### Tools

Seven tools covering the full PCD workflow:

| Tool | Description |
|---|---|
| `list_templates` | List all installed deployment templates with version and default language |
| `get_template` | Retrieve full template content by name and version |
| `list_resources` | List all available resources (templates, prompts, hints) |
| `read_resource` | Read any resource by `pcd://` URI |
| `lint_content` | Validate a PCD spec given as a string; returns structured diagnostics |
| `lint_file` | Validate a spec file on disk; returns structured diagnostics |
| `get_schema_version` | Return the Spec-Schema version the server was built against |

### Resources

Resources are browseable natively by any MCP client — no tool call required:

| URI pattern | Content |
|---|---|
| `pcd://templates/{name}` | Full deployment template Markdown |
| `pcd://prompts/interview` | Interview prompt — guides AI-assisted spec authoring |
| `pcd://prompts/translator` | Universal translator prompt |
| `pcd://hints/{template}.{lang}.{lib}` | Library-specific translator hints |

Prompt content is compiled into the binary as string constants — no runtime
filesystem dependency and no separate install path for prompts.

### The LLM as Wizard

The server provides the data and validation layer; the LLM provides the
conversational layer. A connected LLM can run the complete PCD workflow
in a single session without any manual file handling:

1. Read `pcd://prompts/interview` → conduct spec authoring interview
2. Call `lint_content` on the produced spec → receive structured diagnostics
3. Fix any errors → call `lint_content` again until clean
4. Read `pcd://templates/{name}` → translate spec to code
5. Iterate

No wizard tool. No file copying. No manual prompt management.

### Transports

Both transports are implemented in the same binary, selected by bare-word argument:

```bash
mcp-server-pcd stdio   # for mcphost, Claude Desktop, VS Code
mcp-server-pcd http    # for web-based hosts; default listen: 127.0.0.1:8080
```

**mcphost config (stdio):**
```yaml
mcpServers:
  pcd:
    command: mcp-server-pcd
    args: [stdio]
```

**mcphost config (HTTP, if running as service):**
```yaml
mcpServers:
  pcd:
    url: http://127.0.0.1:8080/mcp
```

### Implementation and self-hosting

`mcp-server-pcd` is specified in PCD format at
`tools/mcp-server-pcd/spec/mcp-server-pcd.md` with `Deployment: mcp-server`.
The implementation is generated from that spec. Self-hosting applies.

The `mcp-server` deployment template (`templates/mcp-server.template.md`) defines
the full constraint set: transport requirements (stdio + streamable-HTTP both
required), framework selection (mcp-go v0.46.0 default), container base image
(`registry.suse.com/bci/golang:latest` — never unqualified names), packaging,
and the six-phase EXECUTION section governing the translation run.

The server was translated by Claude Haiku via direct API call (no MCP
infrastructure, no filesystem server) and passed its compile gate and test
suite. See A.14 for the LLM-G empirical run details.

## A.18 Improving Translation Confidence

### The Problem

AI translation is probabilistic. Structured, readable specifications reduce
ambiguity but cannot eliminate the inherent uncertainty of LLM-based code
generation. This appendix describes two complementary techniques — adopted
in v0.3.10 — that increase confidence in the correctness of generated code
without requiring full formal verification.

Two further approaches (Subplot-based acceptance testing and dual-LLM
behavioural comparison) are under research and are documented at the end
of this appendix as planned future work.

---

### Idea 2 (Implemented): Second-Agent Independent Test Generation

#### Rationale

The EXAMPLES section of a PCD specification is written by the spec author.
It tests what the author thought of. A second AI model, independently reading
the same specification, will generate tests for cases the author did not
consider — boundary conditions, error paths, invariant violations, and
edge cases at the margins of the type constraints.

This is not the same as the primary translator generating its own tests.
The second agent has no access to the primary translation. It reads only
the specification and produces test cases from its own independent reading.
Divergence between what the second agent expects and what the primary
translation produces surfaces specification ambiguity or translation errors.

#### Workflow

```
Spec → Primary LLM → generated code + TRANSLATION_REPORT.md

Spec → Second LLM → INDEPENDENT_TESTS.go (or equivalent)
                    (no access to primary translation)

INDEPENDENT_TESTS.go run against generated code
  → All pass: confidence increases significantly
  → Failures: spec ambiguity or translation error — investigate
```

#### What the Second Agent Produces

The second agent receives only the specification file and the following
instruction:

```
You are a test engineer, not a developer. You have not seen and must
not assume any implementation of the component described in this spec.

Read the specification and generate a test file in {target language}
that tests the component's behaviour from the outside — as a black box.
Focus on:
- Boundary conditions at the edges of type constraints
- Error paths and all declared error codes
- Invariant violations that should be rejected
- Cases not covered by the EXAMPLES section
- Concurrent or repeated invocations where applicable

Do not test internal implementation details. Test only the declared
BEHAVIOR, PRECONDITIONS, POSTCONDITIONS, and INVARIANTS.
```

The output is `INDEPENDENT_TESTS.{ext}` placed in `independent_tests/`
in the audit bundle. For Go: standard `_test.go` format, runnable with
`go test ./...`.

#### Confidence Interpretation

| Result | Interpretation |
|---|---|
| All independent tests pass | High confidence in translation correctness |
| Failures on EXAMPLES cases | Likely translation error — regenerate |
| Failures on boundary cases only | Possible spec underspecification — clarify and regenerate |
| Many failures broadly | Significant translation error or severe ambiguity |

The second-agent test results are documented in TRANSLATION_REPORT.md
under a required section: **Independent Test Results**.

#### TRANSLATION_REPORT Confidence Table (v0.3.13+)

The per-example confidence table in TRANSLATION_REPORT.md must include
a `Verification-method` column and an `Unverified-claims` column.
Confidence backed by a named, runnable test carries weight; confidence
with no test must be explicitly disclosed as unverified.

```markdown
## Confidence per EXAMPLE

| EXAMPLE                | Confidence | Verification method                  | Unverified claims                        |
|------------------------|------------|--------------------------------------|------------------------------------------|
| successful_transfer    | 90%        | FakeAdapter unit test (TestTransfer) | concurrent contention behaviour          |
| insufficient_funds     | 95%        | FakeAdapter unit test (TestErrFunds) | none                                     |
| same_account_rejection | 85%        | No test yet                          | entire rejection path unverified         |
```

A claim is `verified` only if it references a specific test function in
`independent_tests/` that passes without a live external service. Unverified
claims must be listed explicitly, not silently omitted.

#### Formal Rules for Independent Tests (v0.3.13+)

The following rules govern the `independent_tests/` deliverable and are
enforced by the deployment template, not individually by specs:

1. **One test function per EXAMPLE.** Each EXAMPLE in the spec must have a
   corresponding `Test<CamelCasedExampleName>` function in the independent
   tests file.

2. **Tests use only declared test doubles.** Independent tests must import
   only the interfaces and test doubles declared in the spec's INTERFACES
   section. They must not import the production implementation or call live
   external services.

3. **Tests must pass without infrastructure.** `go test ./independent_tests/`
   must succeed with no live cluster, no live daemon, and no network access.

4. **Multi-pass examples generate multi-step tests.** For an EXAMPLE with
   multiple WHEN/THEN pairs, the test function simulates each invocation in
   sequence using the test double's state machine.

#### Integration with Templates

The `independent_tests/` directory and `INDEPENDENT_TESTS.go` are added
as an `optional-recommended` deliverable in the cli-tool template
DELIVERABLES table. For `verified-library` and `mcp-server` templates,
it is `required` for components with Safety-Level above QM.

---

### Cross-Section Consistency Checking (v0.3.13+, partial)

`pcd-lint` previously validated structural completeness only: required
sections present, META fields valid, EXAMPLES with GIVEN/WHEN/THEN.
Identifier and type name drift across sections — a method named `Dial`
in INTERFACES but `transport.Connect` in BEHAVIOR prose — was undetected.

RULE-12 adds cross-section consistency checking. Initial scope for v0.3.13:

1. **Identifier consistency (warning):** Method names declared in
   INTERFACES must appear verbatim in any BEHAVIOR STEPS that reference
   them. Mismatches are warnings with the conflicting locations identified.

2. **Type name consistency (error):** Type names declared in TYPES must
   not be redefined or renamed in BEHAVIOR sections. Conflicts are errors.

3. **File name consistency (warning):** File names declared in the spec
   DELIVERABLES COMPONENT sections must match file names referenced in
   BEHAVIOR/INTERNAL sections when both are present.

State-machine terminology and endpoint semantic consistency are out of
scope for v0.3.13 — they require semantic analysis beyond pattern
matching and are deferred to v0.4.0.

---

### Idea 3 (Implemented): Translation Workflow Diagram

#### Rationale

A TRANSLATION_REPORT.md documents decisions in prose. A workflow diagram
makes the same information visual and immediately scannable — which inputs
were consumed, which decisions were made, which outputs were produced. For
a regulatory auditor reviewing an audit bundle, the diagram provides
orientation before reading the detailed report.

#### Format: Pikchr

Pikchr (https://pikchr.org/) is chosen over Mermaid, D2, and Graphviz DOT
for the following reasons:

- **Supply chain integrity:** Pikchr is distributed as a single self-contained
  C89 source file (`pikchr.c`) with no external dependencies beyond the
  standard C library. It is released under a zero-clause BSD license. The
  binary can be built from source and packaged (e.g. via OBS) — consistent 
  with the project's supply chain security requirements.
- **Layout control:** Pikchr gives the author precise spatial control over
  diagram layout. For technical workflow diagrams where spatial relationships
  carry meaning, this matters more than Mermaid's convenient auto-layout.
- **Pandoc integration:** A pandoc Lua filter renders Pikchr fenced blocks
  to SVG in the whitepaper and documentation pipeline without additional
  tooling.
- **Go wrapper available:** A Go library wraps the Pikchr C code, meaning
  `pcd-tools` could generate Pikchr diagrams natively without shelling out.

Mermaid is retained in `README.md` where GitHub native rendering matters
more than layout precision.

#### Diagram Content Requirements

The `translation-workflow.pikchr` file in `translation_report/` must show:

```
Inputs consumed:
  - Specification file (name, version)
  - Deployment template (name, version)
  - Preset files applied (if any)

Decisions made:
  - Target language resolved (template default or preset override)
  - Framework selected (for mcp-server: mcp-go vs go-sdk)
  - Verification path chosen (direct or verified)
  - Any deviations from template defaults

Outputs produced:
  - Source files (list)
  - Packaging artifacts (RPM, DEB, OCI — as applicable)
  - TRANSLATION_REPORT.md
  - INDEPENDENT_TESTS file (if second-agent run)

Validation results:
  - Compilation: pass/fail
  - EXAMPLES: N/N passed
  - Independent tests: N/N passed (if run)
```

#### Example Diagram Structure

```pikchr
# Translation workflow for: pcd-lint
# Generated: 2026-03-23

box "pcd-lint.md" "v0.3.10" fit
arrow
box "cli-tool.template.md" "v0.3.10" fit with .n at last arrow.e
arrow
diamond "Language?" fit
arrow "Go (default)" right
box "main.go" "go.mod" "Makefile" fit
arrow down from last box.s
box "pcd-lint.spec" "debian/" "Containerfile" fit
arrow down
box "TRANSLATION_REPORT.md" "translation-workflow.pikchr" fit color lightblue
```

The translator generates this diagram after all other files are written,
as the last step before TRANSLATION_REPORT.md. It documents the actual
run, not a generic template.

#### Tooling

`pikchr` should added to the `pcd-tools` package alongside `pcd-lint`.
Both are built from source, statically linked, and distributed as a single
package. 

---

### Idea 1 (Research): Subplot-Based Acceptance Testing

Subplot (https://gitlab.com/subplot/subplot) is a tool for writing
acceptance criteria in Markdown with embedded GIVEN/WHEN/THEN scenarios,
producing automated test suites from those documents. Its scenario format
is nearly identical to the EXAMPLES section of a PCD specification.

**Potential application:** Run Subplot against the EXAMPLES section of a
PCD spec to generate and execute an automated acceptance test suite against
the generated binary. This would partially replace the need for the
second-agent approach for EXAMPLES validation.

**Current status:** Subplot is a relatively small project. Before committing
to it as infrastructure, a prototype is needed: run the `pcd-lint.md`
EXAMPLES through Subplot against the generated `pcd-lint` binary and
evaluate the integration effort, maintained-ness, and Go support.

**Parked for:** v0.3.11 or later, pending prototype evaluation.

---

### Idea 4 (Research): Dual-LLM Behavioural Comparison at Scale

Dual-LLM translation is already described in Appendix A.10 for single
components. The open research question is how to compare two independent
translations of a larger multi-component system, where structural differences
between implementations make direct code comparison meaningless.

**Proposed approach for larger projects:**

Comparison must target behavioural equivalence, not structural identity:

1. **Interface contract compliance:** Run the EXAMPLES from the spec against
   both implementations. Pass/fail comparison, not structural diff.

2. **Property-based testing:** Generate random valid inputs derived from the
   TYPES section and verify both implementations produce identical outputs or
   equivalent error codes.

3. **Divergence as spec signal:** Where the two implementations disagree,
   flag the input as a specification ambiguity requiring clarification.

4. **Formal equivalence (verified path):** For components using the
   meta-language verified path, prove behavioural equivalence in the
   meta-language. This is what A.10 describes.

**`pcd-compare` tool:** A tool that takes two implementation directories
and a specification, runs behavioural equivalence testing, and produces a
comparison report. This is itself a candidate for specification under PCD
— `pcd-compare.md` with `Deployment: cli-tool`.

**Parked for:** v0.3.11 or later, pending specification of `pcd-compare`.

---

## References

- \[Vogels2025\] Werner Vogels, *"The Renaissance Developer"*, AWS re:Invent 2025 Special Closing Keynote, Las Vegas, December 2025. https://www.youtube.com/watch?v=3Y1G9najGiI

- \[Zhang2026\] Wei Zhang and Jessie Xia, *"Structured-Prompt-Driven Development (SPDD)"*, martinfowler.com, 28 April 2026. https://martinfowler.com/articles/structured-prompt-driven/

---

## Changelog

## A.19 Operational Guidance: Regeneration Strategy

### The Question

When a specification changes, should the translator start from scratch or
update the existing generated code? This is the most common operational
question in day-to-day PCD use, and the answer is not always obvious.

### The Short Answer

> **If the change affects TYPES, INTERFACES, INVARIANTS, or the scaffold
> milestone — regenerate from scratch.**
>
> **If the change is limited to STEPS or EXAMPLES of one or two isolated
> BEHAVIORs — update incrementally.**
>
> **If in doubt — regenerate. The spec is the investment; the translator
> run is cheap.**

### Why This Matters: LLM Randomness and Consistency Drift

LLM translation is probabilistic. Two runs from the same spec with the same
model may make different structural decisions — naming conventions, error
handling patterns, package layout. Over multiple incremental runs,
particularly with different models or model versions, a codebase can develop
internal inconsistency that no individual run introduced. This is the primary
risk of incremental update: not that any single change is wrong, but that
the accumulation of independently-reasonable decisions produces a codebase
that no longer has a coherent architecture.

Full regeneration from a clean spec eliminates this risk entirely. The
output is guaranteed internally consistent because it was produced in a
single pass from a single source of truth.

### Decision Framework

The change should be evaluated against four dimensions:

**1. Structural impact**

Changes to TYPES, INTERFACES, or INVARIANTS have high structural impact —
they touch the foundation that all BEHAVIORs build on. A changed type used
in five BEHAVIORs requires every one of those BEHAVIORs to be reconsidered.
A changed INTERFACE method signature propagates through every call site.
These changes warrant full regeneration regardless of how small they appear
in the spec diff.

Changes limited to STEPS or EXAMPLES of a single BEHAVIOR are isolated by
construction — the BEHAVIOR boundary is the unit of containment. If the
blast radius is genuinely one BEHAVIOR, incremental update is viable.

**2. Scaffold boundary**

The scaffold milestone (Scaffold: true) is a hard boundary. Any change
that touches the scaffold — new files, restructured packages, new types,
changed interface shapes — requires full regeneration. All subsequent
milestones were built on the scaffold's decisions; changing those
decisions invalidates them.

Conversely, if the scaffold is untouched and the change fits within one
unreleased milestone's scope, incremental update within that milestone
is appropriate. This is the intended use of the milestone mechanism.

**3. Released milestone constraint**

Released milestones are frozen — they are audit evidence. A change that
requires touching a released milestone's scope must go through full
regeneration. The new translation produces a new released milestone,
with a new audit bundle, superseding the old one.

**4. Codebase provenance**

If the existing code was produced by a single translation run from a clean
spec, incremental update carries lower risk. If the existing code has been
through multiple incremental runs, or was produced by different models,
the consistency drift risk is higher and full regeneration becomes more
attractive even for isolated changes.

### The Blast Radius Analysis

Before recommending incremental update, the analyst must estimate the blast
radius of the change: how many files, BEHAVIORs, and functions are affected?

- **1–2 BEHAVIORs, no shared types changed**: low blast radius → incremental viable
- **3–5 BEHAVIORs, or a shared type changed**: medium blast radius → judgement call
- **5+ BEHAVIORs, or INTERFACES changed**: high blast radius → full regeneration

### Automated Impact Assessment

The change-impact analysis is a natural candidate for automation. Given the
old spec, the new spec (or the diff between them), and optionally the
existing generated code, an LLM can:

- Identify which TYPES, INTERFACES, INVARIANTS, and BEHAVIORs are affected
- Estimate the blast radius across the codebase
- Check whether any released milestones are in scope
- Check whether the scaffold is affected
- Produce a structured recommendation with reasoning

This analysis is available as `prompts/change-impact.md` — a prompt that
takes a spec change description and returns a structured assessment. The
`assess_change_impact` tool in `mcp-server-pcd` (v0.3.0+) automates this
for pipeline use: the agent assesses the change before invoking the
translator, choosing the appropriate strategy without human intervention.

The tool can also catch a non-obvious class of problem: a change that
*looks* small in the spec diff but has a large blast radius because the
changed type or method is used throughout the codebase. Human reviewers
eyeballing a diff may miss this; an LLM reading both the spec and the
code will not.

### The Decisions Hints File

A clean process model requires a clean separation between what belongs in
the spec and what belongs in the generated implementation. This creates a
gap: when full regeneration is recommended, there may be architectural
decisions in the existing implementation — package layout, routing patterns,
error conventions — that are worth preserving but that are not captured in
the spec. Writing them into the spec would couple the spec to a specific
implementation. Ignoring them risks losing good prior decisions.

The solution is a fourth layer of hints file: `<specname>.<language>.decisions.hints.md`.

This file:

- Lives **next to the spec** (not in `hints/`) — it is spec-scoped, not
  template-scoped
- Is **language-specific** and disposable — when switching target languages,
  the old decisions hints file is discarded and a new one is produced by the
  first translation in the new language
- Is **generated by the translator** as a required deliverable alongside
  `TRANSLATION_REPORT.md`, capturing any architectural decisions made during
  the run that are not inferable from the spec alone
- Is **read by the translator** at the start of a guided regeneration or
  incremental update run; ignored on clean full regenerations from scratch
- Is **not a spec artifact** — it does not affect pcd-lint validation and
  is not part of the human-reviewed specification

When `assess_change_impact` recommends full regeneration and produces a
"decisions hints to preserve" list, that list is written into
`<specname>.<language>.decisions.hints.md` before the regeneration run.
The translator reads it as normative constraints — the same way it treats
any other hints file.

The three-state process model is therefore:

| State | Translator reads | Decisions hints |
|---|---|---|
| Clean full regeneration | Spec + template only | Ignored |
| Guided regeneration | Spec + template + decisions hints | Read as constraints |
| Incremental update | Spec diff + existing code + decisions hints | Read as constraints |

This model keeps the spec language-agnostic and permanent, the decisions
hints language-specific and disposable, and the process clean.

### Cost Considerations

Full regeneration sounds more expensive, but the economics are more
favourable than they appear:

- A complete translation run for a well-specified component costs one LLM
  session at approximately 128K output tokens — minutes, not hours
- The spec investment (the human work) is fully reused across every
  regeneration run and every target language
- An incremental update that silently introduces inconsistency may cost
  far more to diagnose and fix than a clean regeneration would have

The expensive part of PCD is writing the spec. The cheap part is running
the translator. This asymmetry should inform the default choice.

### Summary Table

| Change type | Blast radius | Scaffold affected | Recommendation |
|---|---|---|---|
| TYPES modified | Any | No | Full regeneration |
| INTERFACES modified | Any | No | Full regeneration |
| INVARIANTS modified | Any | No | Full regeneration |
| Scaffold milestone modified | Any | Yes | Full regeneration |
| BEHAVIOR added | Low | No | New milestone; incremental |
| BEHAVIOR STEPS modified | Low (1–2) | No | Incremental |
| BEHAVIOR STEPS modified | Medium (3–5) | No | Judgement call |
| EXAMPLES added/fixed | Any | No | Incremental |
| Released milestone in scope | Any | Any | Full regeneration |
| Multiple prior incremental runs | Any | No | Prefer full regeneration |

### Decisions hints file quick reference

| Situation | Action |
|---|---|
| Full regeneration recommended | Write "decisions hints to preserve" → `<specname>.<lang>.decisions.hints.md`, then regenerate |
| Incremental update recommended | Read existing `<specname>.<lang>.decisions.hints.md` before starting |
| Switching target language | Discard `<specname>.<oldlang>.decisions.hints.md`; new file created by first translation run |
| Clean full regeneration (no prior decisions) | No decisions hints file needed; translator starts clean |

---

| Version | Date | Changes |
|---------|------|---------|
| 0.3.22 | 2026-04-13/15 | Spec hash embedding (A.20, RULE-18, all templates, prompt.md); doc restructure (user-guide, technical-reference, whitepaper 6-pager); mcp-server-pcd v0.3.1 spec (verify_spec_hash); `spack-package` deployment template added (10th template; first declarative specification target type; derived from Spack 1.1.1 source; all 18 `spack audit` checks mapped to INVARIANTS) |
| 0.3.21 | 2026-04-04/07 | MILESTONE mechanism; Scaffold-first pattern; stub contract; RULE-15/16/17 added to pcd-lint; mcp-server-pcd set_milestone_status tool; doc restructure (user-guide, reverse-prompt, interview-prompt Phase 9) |
| 0.3.20 | 2026-04-01 | Man pages as required deliverable (all templates except cloud-native and project-manifest); pandoc as build-time dependency; all 9 templates pcd-lint clean |
| 0.3.19 | 2026-03-31 | LICENSE deliverable fix; BCI builder image fix; naming patches |
| 0.3.19 | 2026-03-27 | Added informal name "Piccadilly" and tagline "Meet me at the Piccadilly — the place where intent becomes implementation." README subtitle and whitepaper Executive Summary updated. Presentation header subtitle added. |
| 0.3.18 | 2026-03-26 | A.17 rewritten: mcp-server-pcd now fully specified, implemented, and tested. Updated to reflect actual tool set (7 tools), native MCP resource registration, embedded prompts, dual transports, registry.suse.com container base. README updated: Self-Hosting expanded to Tooling section covering both pcd-lint and mcp-server-pcd; repository layout updated. |
| 0.3.17 | 2026-03-25 | A.14 updated: LLM-G run added (small frontier model, direct API, no MCP, 15/15 tests pass, full compile gate, two-layer prompt v0.3.16 validated end-to-end). Models and deliverables tables updated. |
| 0.3.16 | 2026-03-25 | Two-layer prompt architecture: prompts/prompt.md is now fully language-agnostic; all delivery phases, resume logic, and compile gate instructions moved to template ## EXECUTION sections. EXECUTION section added to cli-tool, cloud-native, and mcp-server templates. RULE-14 added to pcd-lint: deployment templates must have ## EXECUTION section. A.13 rewritten to document the two-layer design. |
| 0.3.15 | 2026-03-25 | Added AI-interview workflow for spec authoring. Introduction principle 1 updated: domain experts no longer need to learn the spec format; prompts/interview-prompt.md guides any LLM to conduct a structured interview and produce a complete spec. A.4 Phase 0 and Phase 1 updated to reflect interview-based approach. |
| 0.3.14 | 2026-03-25 | cloud-native template v0.3.14: INDEPENDENT_TESTS Go naming note; deploy/operator.yaml dedup; HEALTHCHECK contradiction resolved; CRD scope declared in spec; go.sum as generated file. Hints files shipped: go-libvirt and golang-crypto-ssh. Phase 7 compile gate in translator prompt. |
| 0.3.13 | 2026-03-24 | Resolved all 8 items deferred from v0.3.12, plus 4 new findings from codex lessons. F1: TYPE-BINDINGS table added to deployment templates — maps logical spec types to language-specific types; spec stays language-agnostic. F7: component-based DELIVERABLES section added to spec schema — logical components only; filename mapping stays in templates. F8: TRANSLATION_REPORT confidence table now requires Verification-method and Unverified-claims columns. F10: four formal independent test rules added — one test per EXAMPLE, tests use only declared test doubles, infrastructure-free, multi-pass examples generate multi-step tests. F11 (codex): negative-path EXAMPLES now required for any BEHAVIOR with error exits in STEPS (RULE-10). F12 (codex): TOOLCHAIN-CONSTRAINTS optional spec section added for spec-specific OCI/build/generated-file constraints (RULE-11). F13 (codex): cross-section consistency checking added to pcd-lint as RULE-12 (identifier, type name, file name consistency; partial implementation). F14 (codex): Constraint: field added to BEHAVIOR headers — required \| supported \| forbidden with default=required (RULE-13). All templates bumped to v0.3.12. CONTRIBUTING.md updated. |
| 0.3.12 | 2026-03-24 | Applied 9 of 10 findings from remote-kvm-operator exercise. STEPS: now required in every BEHAVIOR block (F2). MECHANISM: inline annotation added (F2). Multi-pass WHEN/THEN EXAMPLES format added (F3). INTERFACES optional spec section added with test-double declarations (F4). hints/ directory added to preset hierarchy with library hints file format and naming convention (F5). DEPENDENCIES optional spec section added with do-not-fabricate flag (F6). [observable]/[implementation] INVARIANTS annotation — Option B adopted (F9). Deferred to v0.3.13: TYPE-BINDINGS in templates (F1), component-based DELIVERABLES (F7), TRANSLATION_REPORT confidence table update (F8), A.18 independent test formal rules (F10). |
| 0.3.11 | 2026-03-24 | Added cloud-native.template.md (complete). Added prompts/README-small-models.md. Added tools/pcd-lint/spec/prompt.md (component-specific hardcoded-filename prompt for small models). |
| 0.3.10 | 2026-03-23 | Added A.18: Improving Translation Confidence — second-agent independent test generation (implemented) and Pikchr workflow diagram (implemented); Subplot and dual-LLM comparison parked for research. Updated audit bundle structure with independent_tests/ and translation-workflow.pikchr. Updated A.13 prompt to request workflow diagram. |
| 0.3.9 | 2026-03-23 | Corrected logical fallacy regarding verifiability claims. Clarified that AI translation is probabilistic and that specification structure alone cannot guarantee correctness. Verifiability is achieved through multiple complementary mechanisms. |
| 0.3.8 | 2026-03-19 | Added A.16: Large Projects. Added A.17: mcp-server-pcd. project-manifest stub. pcd-wizard dropped. MCP naming convention. |
| 0.3.7 | 2026-03-18 | Anonymized all LLM/vendor names in A.14. Removed version numbers from all internal filename references. |
| 0.3.6 | 2026-03-18 | crypto-library → verified-library. python-tool added. library-c-abi CPS note. |
| 0.3.5 | 2026-03-18 | Renamed spec-lint → pcd-lint. post-coding paths → pcd. Filename convention. Curly brace placeholders. |
| 0.3.4 | 2026-03-18 | CC-BY-4.0 for specs/templates, GPL-2.0-only for tools. A.15 SCA. |
| 0.3.3 | 2026-03-17 | Expanded A.13/A.14. translation_report/ in audit bundle. DELIVERABLES expanded. |
| 0.3.2 | 2026-03-17 | A.13 prompt. A.14 empirical tests. DELIVERABLES in template. Unified versioning. |
| 0.3.1 | 2026-03-17 | Changelog moved to end. A.12 industry landscape. BEHAVIOR/INTERNAL. |
| 0.3.0 | 2026-03-17 | Deployment template system. Target: removed from META. |
| 0.2.3 | 2026-02-10 | Workflow diagram. Dual-path architecture. |
| 0.2.1 | 2026-02-10 | Change the collaborating AI |
| 0.1.0 | 2026-01-10 | First draft tracked in GIT  |
| 0.0.0 | 2026-01-05 | Start thinking about "Optimizing Programming Languages for AI and Human Review" |



