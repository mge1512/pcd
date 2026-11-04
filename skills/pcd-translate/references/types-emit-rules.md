# PCD TYPES Emission Rules (Shared)

## META
Deployment:   none
Version:      0.1.0
Spec-Schema:  0.4.0
Author:       Matthias G. Eckermann <pcd@mailbox.org>
License:      CC-BY-4.0
Verification: none
Safety-Level: QM

This specification is a composition target (Deployment: none). It is not
translated into an implementation on its own; it is included by host specs
via the META field

```
Includes: ../../shared/spec/types-emit-rules.md
```

It is the single source of truth for **lowering TYPE tables to
JSON Schema**: the normative mapping from the tabular TYPES notation
(defined in `types-table-rules.md`) to a JSON Schema 2020-12 document,
and the canonical output form that makes the lowering deterministic.
The current consumer is:

- `libpcd` (Deployment: library) - the shared PCD engine. Front-end
  tools (`pcd-emit-types`, `mcp-server-pcd`) consume the emission
  through libpcd's public interface (DEPENDENCIES binding), not through
  inclusion.

Golden table-in/schema-out pairs are maintained in libpcd.spec.md's
EXAMPLES section, following the lint-rules.md precedent of keeping
acceptance examples with the consuming host. Those pairs are the
conformance anchor for this mapping: an implementation is correct when
it reproduces them byte for byte.

---

## Design constraints

1. **Determinism.** Same table, same bytes. The output form is fully
   specified: member order, indentation, line endings, escaping. Two
   independent implementations of this file must produce byte-identical
   output for the same input; CI byte-diffs them (differential
   self-test).
2. **Standard keywords carry structure.** The emitted document is plain
   JSON Schema 2020-12. Structural semantics use only standard keywords
   (oneOf, const, enum, required, additionalProperties, bounds,
   pattern, deprecated), so unmodified off-the-shelf generators can
   consume the output. PCD extensions are annotation-only `x-pcd-*`
   members that validators ignore by design. An `x-pcd-*` member never
   changes what validates.
3. **Provenance in the artifact, spec hash alone.** The emitted document
   embeds the merged spec hash as `x-pcd-spec-sha256` - the same policy
   as generated source artifacts. Tool name and version are not
   embedded: output bytes stay stable across tool releases when the
   mapping is unchanged. The mapping's own version is this file's
   Version, which participates in libpcd's merged spec hash.

---

## TYPES

```
CanonicalSchemaText := string
// A JSON Schema 2020-12 document in the canonical output form defined
// below. UTF-8, LF line endings, exactly one trailing newline.

EmitInputs := {
  model:      TypesTableModel,   // from parse-types-tables
  spec_title: string,            // first "#" heading of the host spec
  spec_hash:  string             // merged Spec-SHA256, lowercase hex
}
```

---

## Mapping (normative)

| Table construct | JSON Schema |
|---|---|
| the spec | one document; `$schema` 2020-12, `title` = spec title, `x-pcd-spec-sha256` = merged hash, `$defs` = one entry per TYPE table in document order |
| Kind: record | `"type": "object"`, `properties` in row order, `required` from Req=yes rows in row order (omitted when empty), `"additionalProperties": false` |
| Kind: variant | `oneOf` with one branch per row in row order; each branch a closed object whose first property is the tag with `"const": "<Variant>"`; payload fields follow in declared order; `required` = tag then payload field names; `"additionalProperties": false` |
| Kind: enum, plain form | `"type": "string"`, `"enum": [values in row order]` - used when every Description cell is `-` and no lifecycle cell differs from `-` |
| Kind: enum, annotated form | `"type": "string"`, `oneOf` of `{ "const": value, ... }` branches in row order - used otherwise |
| Type: string / integer / number / boolean | `"type": "<scalar>"` |
| Type: TypeName | `{ "$ref": "#/$defs/TypeName" }` |
| Type: T[] | `"type": "array"`, `"items": { scalar or $ref form of T }` |
| Req = yes | field name listed in `required` |
| len A..B | `"minLength": A`, `"maxLength": B` (open side omitted) |
| A..B (numeric) | `"minimum": A`, `"maximum": B` (open side omitted) |
| pattern: R | `"pattern": "R"` (\| unescaped to \|, then JSON-escaped) |
| one-of(a,b,c) | `"enum": [a, b, c]` - JSON strings for string fields, numeric literals for integer/number fields |
| Default (not `-`) | `"default": <literal>` typed per the field's Type |
| Description (not `-`) | `"description": "<cell text>"` |
| Since (not `-`) | `"x-pcd-since": "<value>"` |
| Deprecated = yes | `"deprecated": true` |
| Replaced-by (not `-`) | `"x-pcd-replaced-by": "<value>"` |

Placement of lifecycle members: on the property schema (record rows), on
the branch schema (variant rows), or on the const branch (enum rows in
annotated form). Whole-table lifecycle is not expressible in v0.1; the
unit of deprecation is the row.

---

## Canonical output form (normative)

Serialization:

- UTF-8; LF line endings; the document ends with exactly one newline;
  no trailing whitespace on any line.
- Two-space indentation. Every object member on its own line. One space
  after the colon.
- Arrays of scalars (`enum`, `required`) on a single line:
  `"required": ["mode", "image_path"]`. Arrays of objects (`oneOf`)
  multi-line, one element object per block.
- JSON string escaping is minimal: only `\"`, `\\`, and control
  characters (as `\n`, `\r`, `\t`, else `\u00XX`) are escaped. No
  escaping of forward slashes or non-ASCII characters.
- Numbers verbatim as written in the cell; the emitter never reformats
  numeric literals.

Member order:

```
document:        $schema, title, x-pcd-spec-sha256, $defs
record def:      type, properties, required, additionalProperties
property schema: description,
                 type | $ref | (type: array + items),
                 minLength, maxLength, pattern, minimum, maximum, enum,
                 default,
                 x-pcd-since, deprecated, x-pcd-replaced-by
variant def:     oneOf
variant branch:  description, type, properties, required,
                 additionalProperties,
                 x-pcd-since, deprecated, x-pcd-replaced-by
tag property:    const
payload property: type | $ref | (type: array + items)
enum def (plain):     type, enum
enum def (annotated): type, oneOf
enum branch:     const, description,
                 x-pcd-since, deprecated, x-pcd-replaced-by
```

Members whose source cell is `-` are omitted entirely. `$defs` keys are
the TypeName values verbatim; property keys are the FieldName values
verbatim; discriminator const values are the VariantName values
verbatim; enum members are the EnumValue values verbatim.

Illustration (non-normative; the normative pairs live in libpcd's
EXAMPLES). A single-field record

```markdown
### TYPE: Probe
Kind: record

| Field | Type   | Req | Constraints | Default | Description |
|-------|--------|-----|-------------|---------|-------------|
| name  | string | yes | len 1..16   | -       | Probe name  |
```

lowers to

```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "title": "example",
  "x-pcd-spec-sha256": "<merged hash>",
  "$defs": {
    "Probe": {
      "type": "object",
      "properties": {
        "name": {
          "description": "Probe name",
          "type": "string",
          "minLength": 1,
          "maxLength": 16
        }
      },
      "required": ["name"],
      "additionalProperties": false
    }
  }
}
```

---

## BEHAVIOR: emit-types-schema
Constraint: required

Lowers a validated TypesTableModel to the canonical JSON Schema text.
Emission is a total function over validated models: parse-time
diagnostics (RULE-22 through RULE-25 Errors) are the gate; a model that
passed the gate always emits.

INPUTS:
```
inputs: EmitInputs
```

OUTPUTS:
```
text: CanonicalSchemaText
```

PRECONDITIONS:
- inputs.model was produced by parse-types-tables and carries no
  RULE-22 through RULE-24 Error diagnostics
- inputs.spec_hash is the lowercase-hex merged Spec-SHA256 of the spec
  the model was parsed from

STEPS:
1. Open the document with `$schema`, `title` = inputs.spec_title,
   `x-pcd-spec-sha256` = inputs.spec_hash.
2. For each table in inputs.model.tables, in order, append one `$defs`
   entry per the Mapping section:
   a. record: object form; properties in row order; `required` from
      Req=yes rows; `additionalProperties: false`.
   b. variant: `oneOf` branches in row order; tag property first with
      `const` = Variant cell; payload fields per PayloadCell; branch
      `required` and `additionalProperties: false`.
   c. enum: plain form when all Description and lifecycle cells are
      `-`; annotated `oneOf`+`const` form otherwise.
3. Render per the canonical output form: member order, indentation,
   scalar-array inlining, minimal escaping, LF, one trailing newline.
4. Return the text.

POSTCONDITIONS:
- output is valid JSON and a valid JSON Schema 2020-12 document
- output is byte-identical across runs and across conforming
  implementations for identical inputs
- `x-pcd-spec-sha256` equals inputs.spec_hash
- removing every `x-pcd-*` member leaves validation behaviour unchanged
- the emitter performs no file, network, or environment access;
  reading the spec and writing the text are the caller's concern

---

## PRECONDITIONS

- Emission applies only to models produced by parse-types-tables
  (types-table-rules.md); the two files share the TypesTableModel
  contract

---

## POSTCONDITIONS

- Emission is read-only with respect to the spec and side-effect-free;
  it returns text and nothing else

---

## INVARIANTS

- [observable]      same table set, spec title, and spec hash produce
  byte-identical output - across runs, platforms, and conforming
  implementations
- [observable]      the emitted document uses only standard JSON Schema
  2020-12 keywords for structure; `x-pcd-*` members are annotations
  that never affect validation
- [observable]      the emitted document embeds the merged spec hash
  alone; no tool name, tool version, or timestamp appears in the output
- [implementation]  `$defs` order, property order, branch order, and
  enum member order follow document and row order of the source tables

---

## CHANGELOG

- 2026.07.07.01 - Initial version. Normative mapping from the tabular
  TYPES notation to JSON Schema 2020-12, canonical output form
  (member order, indentation, escaping, line discipline), and the
  emit-types-schema BEHAVIOR. Golden pairs live in libpcd.spec.md
  EXAMPLES. Consumed by libpcd via Includes.
