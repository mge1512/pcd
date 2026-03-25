# PCDP Specification Interview — Usage Guide

`interview-prompt.md` is a prompt that instructs any LLM to produce a complete
PCDP specification from a conversation with a domain expert. Two options are
supported:

- **Option 1 — Full interview:** The model asks questions from scratch.
  Good for new components with no prior documentation.

- **Option 2 — Gap-fill:** The expert provides existing material (email,
  meeting notes, design doc, ticket). The model extracts what it can, flags
  contradictions, confirms the extraction, then asks only for what is missing.
  Good for components that already have some documentation.

The expert does not need to know the PCDP format, any programming language,
or formal notation.

## How to use it

### With mcphost or a local model

```bash
# Set interview-prompt.md as the system prompt in your mcphost config:
# config.yaml:
#   systemPrompt: "@prompts/interview-prompt.md"

mcphost
```

The model will ask the first question immediately.

### With any chat interface (browser, API, Claude Desktop, etc.)

Paste the entire contents of `interview-prompt.md` as the system prompt,
or as the first message prefixed with "Your instructions:".

### With a small local model (Ollama, llama.cpp, etc.)

Small models work well with this prompt. Recommended minimum:
- 7B for simple components (Option 1)
- 7B for Option 2 on short source material
- 13B+ for Option 2 on long or complex source material

```bash
ollama run llama3.2 "$(cat prompts/interview-prompt.md)"
```

For Option 2 with a document: paste or pipe the source material into the
conversation after the model asks its opening question.

## What the interview produces

At the end, the model writes a complete PCDP specification in Markdown.
Copy it into a `.md` file, then validate:

```bash
pcdp-lint mycomponent.md
```

Then translate to code using the standard translation prompt:

```bash
# see prompts/prompt.md or prompts/README-small-models.md
```

## Full workflow

```
Option 1: AI interviews → human reviews → pcdp-lint → AI translates → code
Option 2: human provides material → AI extracts + gaps → human reviews → pcdp-lint → AI translates → code
```

## Phase coverage

| Phase | Option 1 | Option 2 | Output section(s) |
|---|---|---|---|
| Mode selection | asked first | asked first | — |
| Extraction | — | automatic from material | all sections |
| Contradiction resolution | — | before proceeding | any section |
| 1. Component identity | full questions | gaps only | META |
| 2. Data model | full questions | gaps only | TYPES |
| 3. External systems | full questions | gaps only | INTERFACES |
| 4. Operations and steps | full questions | gaps only | BEHAVIOR + STEPS |
| 5. Rules and constraints | full questions | gaps only | PRECONDITIONS, POSTCONDITIONS, INVARIANTS |
| 6. Concrete examples | full questions | gaps only | EXAMPLES |
| 7. External libraries | full questions | gaps only | DEPENDENCIES |
| 8. Assembly | writes spec | writes spec | full specification |
| 9. Self-check | validates | validates | — |

## Contradiction handling

When the source material (Option 2) contains conflicting values, the model
stops immediately, states the contradiction and its two sources, and asks
the expert to resolve it before continuing. It never guesses, never picks
the more conservative value, and never continues with an unresolved contradiction.

## Tips for domain experts

- Answer in plain language. The model translates into formal notation.
- For Option 2: paste or attach your material after the model asks its
  opening question. Any format works — email text, bullet points, prose.
- Phase summaries and the extraction review are checkpoints. Correct
  misunderstandings there, before they propagate into the spec.
- The worked examples in the prompt show what the conversation looks like.

## Tips for model selection

Phase 8 (assembly) writes the full specification in one block — the most
demanding step. If using a very small model (3B or less), consider switching
to a larger model for Phase 8 only by copying the conversation transcript
and asking the larger model to produce the spec from it.

Phase 9 (self-check) is more reliable with larger models. For critical
specifications, run `pcdp-lint` regardless of the model's self-assessment.

For Option 2 with long source material (multiple documents, long design
docs), a model with a larger context window is preferable.

## Relationship to the translation prompt

This prompt produces a specification. The translation prompt takes that
specification and produces code. They are separate steps intentionally —
different models can be used for each.

```
interview-prompt.md  ──►  specification (.md)
                                  │
                          pcdp-lint (validate)
                                  │
                     prompts/prompt.md  ──►  code + audit bundle
```
