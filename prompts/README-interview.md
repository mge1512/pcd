# PCDP Specification Interview — Usage Guide

`interview-prompt.md` is a prompt that instructs any LLM to interview a
domain expert and produce a complete PCDP specification. The expert does
not need to know the PCDP format, any programming language, or formal
notation.

## How to use it

### With mcphost or a local model

```bash
# Place the prompt as the system prompt in your mcphost config:
# config.yaml:
#   systemPrompt: "@prompts/interview-prompt.md"

mcphost
```

Then simply start talking. The model will ask the first question.

### With any chat interface (browser, API, Claude Desktop, etc.)

1. Paste the entire contents of `interview-prompt.md` as the first message,
   prefixed with "Your instructions:" or as a system prompt.
2. The model will begin the interview immediately.

### With a small local model (Ollama, llama.cpp, etc.)

Small models work well for this prompt. Recommended minimum:
- 7B parameter model for simple components
- 13B+ for complex components with many operations or interfaces

The prompt is designed for one question at a time — this keeps the context
window manageable even on constrained hardware.

```bash
# Example with Ollama:
ollama run llama3.2 "$(cat prompts/interview-prompt.md)"
```

## What the interview produces

At the end of the interview, the model writes a complete PCDP specification
in Markdown format. Copy it into a `.md` file, then validate it:

```bash
pcdp-lint mycomponent.md
```

Fix any reported issues (the model may have missed a required field or
section), then use the spec for translation:

```bash
# Translation uses prompts/prompt.md or a component-specific variant
# See prompts/README-small-models.md for small model translation guidance
```

## What the interview covers

| Phase | Output section(s) |
|---|---|
| 1. Component identity | META |
| 2. Data model | TYPES |
| 3. External systems | INTERFACES |
| 4. Operations and steps | BEHAVIOR + STEPS |
| 5. Rules and constraints | PRECONDITIONS, POSTCONDITIONS, INVARIANTS |
| 6. Concrete examples | EXAMPLES (including multi-pass) |
| 7. External libraries | DEPENDENCIES |
| 8. Assembly | Full specification |
| 9. Self-check | pcdp-lint pre-flight |

## Tips for domain experts

- Answer in plain language. The model translates your answers into formal notation.
- If a question is unclear, say so. The model will rephrase it.
- The phase summaries are checkpoints — correct any misunderstanding there,
  not later.
- The worked example in the prompt shows what "good answers" look like.

## Tips for model selection

For **Phase 8 (assembly)**, the model writes the full specification in one block.
This is the most demanding step. If using a very small model (3B or less),
consider switching to a larger model just for Phase 8 by copying the interview
transcript and asking it to produce the spec.

For **Phase 9 (self-check)**, the model verifies its own output against a
checklist. Larger models are more reliable here. For critical specifications,
run pcdp-lint regardless of the model's self-assessment.

## Relationship to the translation prompt

This prompt produces a specification. The translation prompt (`prompts/prompt.md`
or a component-specific variant) takes that specification and produces code.
They are separate steps:

```
interview-prompt.md → specification (.md)
                               ↓
                     pcdp-lint (validate)
                               ↓
              prompts/prompt.md → code + audit bundle
```
