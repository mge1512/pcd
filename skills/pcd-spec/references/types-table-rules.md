# PCD TYPES Table Rules (Shared)

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
Includes: ../../shared/spec/types-table-rules.md
```

and its TYPES, BEHAVIORs and INVARIANTS are merged into the host's
effective specification before translation (see doc/spec-composition.md).

It is the single source of truth for the **tabular TYPES notation**: the
grammar of `### TYPE:` tables inside a spec's `## TYPES` section, the
closed cell vocabularies, and the structural validation rules RULE-22
through RULE-25. The current consumer is:

- `libpcd` (Deployment: library) - the shared PCD engine. Front-end
  tools (`pcd-lint`, `pcd-emit-types`, `mcp-server-pcd`) consume these
  rules through libpcd's public interface (DEPENDENCIES binding), not
  through inclusion.

The emission of TYPE tables into JSON Schema is defined separately in
`types-emit-rules.md`. This file defines only what a well-formed TYPE
table *is*; that file defines what it *produces*.

Rule-level acceptance EXAMPLES are intentionally not kept in this shared
spec. They are maintained in the consuming host's EXAMPLES section
(libpcd.spec.md), following the precedent set by lint-rules.md.

---

## Purpose and Position

PCD specs describe data structures in the `## TYPES` section. The
free-form notation (`Name := { field: type, ... }`) is expressive but not
mechanically derivable. The tabular notation defined here is the
declarative subset: a closed grammar whose instances can be validated by
lint and lowered deterministically to JSON Schema and, through existing
per-language generators or the deterministic emitter, to type definitions
in any target language.

Design rule inherited from the framework: **the table is the normative,
human-certifiable artifact**. Everything derived from it (JSON Schema,
generated types, diagrams) is a regenerable build output. The table is
what a reviewer signs.

The notation deliberately keeps the column set small and closed. New
requirements do not become new columns; they become entries in the
Constraints vocabulary, separate INVARIANTS, or a Spec-Schema revision.

---

## TYPES

```
TypeTableKind := record | variant | enum
// The three table kinds. Declared by the Kind: line of a TYPE block.

TypeName := string where matches("^[A-Z][A-Za-z0-9]*$")
// PascalCase. Names TYPE tables and is used verbatim in $defs keys
// and $ref targets on emission.

FieldName := string where matches("^[a-z][a-z0-9_]*$")
// snake_case. Used verbatim as JSON property names on emission.

ScalarType := string | integer | number | boolean

TypeRef := ScalarType | TypeName | ScalarType "[]" | TypeName "[]"
// A field's Type cell. "[]" denotes an array of the element type.
// Nested arrays (T[][]) are not permitted in v0.1.

ReqValue := yes | no

ConstraintToken :=
    "len " Range          // string length bounds -> minLength/maxLength
  | Range                 // numeric bounds -> minimum/maximum
  | "pattern: " Regex     // regular expression -> pattern
  | "one-of(" Literals ")"// closed value set -> enum
// Range := A ".." B | A ".." | ".." B
//   A, B decimal literals; A <= B when both present. Integer literals
//   for integer-typed fields; integer or decimal for number-typed.
// Regex: JSON Schema "pattern" semantics (ECMA-262 subset). A literal
//   pipe inside a table cell must be written as \| (GFM escape); the
//   parser unescapes \| to | before use.
// Literals: comma-separated tokens without commas, parentheses, or
//   unescaped pipes. Emitted as JSON strings for string-typed fields,
//   as numeric literals for integer/number-typed fields.

ConstraintCell := "-" | ConstraintToken { "; " ConstraintToken }
// Multiple constraints are separated by semicolon-space.
// "-" means no constraints.

DefaultCell := "-" | literal
// "-" means no default. Otherwise a literal compatible with the field's
// Type: verbatim cell text for string, a numeric literal for
// integer/number, true|false for boolean, a declared Value for a field
// whose Type references an enum table. Fields of record, variant, or
// array type take no default in v0.1 ("-" only).

TagName := string where matches("^[a-z][a-z0-9_]*$")
// Discriminator property name for variant tables. Default: "kind".

VariantName := string where matches("^[A-Za-z][A-Za-z0-9_-]*$")
// Emitted verbatim as the discriminator's const value.

EnumValue := string where matches("^[A-Za-z0-9][A-Za-z0-9_.-]*$")
// Emitted verbatim as an enum member.

PayloadCell := "-" | PayloadField { ", " PayloadField }
// PayloadField := FieldName ": " TypeRef
// Payload fields carry no per-field constraints in v0.1. Where a
// payload needs constrained fields, declare a record table and
// reference it by TypeName.

LifecycleColumns := [ "Since" ] [ "Deprecated" ] [ "Replaced-by" ]
// Optional trailing columns, each independently optional, appearing in
// this relative order after the kind-specific required columns.
// Since:       SpecVersion value or "-"
// Deprecated:  yes | -
// Replaced-by: TypeName, FieldName, VariantName, or EnumValue, or "-"

TypesTableModel := {
  tables: List<TypeTable>       // document order preserved
}

TypeTable := {
  name:     TypeName,
  kind:     TypeTableKind,
  tag:      TagName,            // variant only; "kind" when no Tag: line
  line:     u32,                // 1-based line of the ### TYPE: heading
  rows:     List<TableRow>      // row order preserved
}

TableRow := {
  cells: Map<string, string>    // header column name -> raw cell text,
                                // \| already unescaped
}
```

---

## Grammar: the TYPE block

A TYPE block appears inside the `## TYPES` section of a spec (host or
merged), at column 0, outside fenced code blocks. Structural detection
follows the same fence-depth and column-0 conventions as all other PCD
markers (see BEHAVIOR/INTERNAL: code-fence-tracking in lint-rules.md).

Block form (shown fenced here so it is not parsed as real structure):

```markdown
### TYPE: <TypeName>
Kind: record | variant | enum
Tag: <TagName>            (optional; variant tables only)

| ...header per Kind... |
|---|
| ...one row per field / variant / value... |
```

Rules of the block:

- The heading is exactly `### TYPE: ` followed by a TypeName.
- The `Kind:` line is the first non-blank line after the heading.
- The optional `Tag:` line, if present, immediately follows the
  `Kind:` line and is permitted only when Kind is variant.
- The table is the first GFM table following the Kind/Tag lines. Blank
  lines and prose paragraphs may separate them; prose may also follow
  the table. Prose is documentation, never structure.
- Cell text is trimmed of leading and trailing whitespace. The GFM
  escape `\|` denotes a literal pipe inside a cell.

Required header per Kind (exact column names, exact order), optionally
followed by LifecycleColumns:

```
record:   | Field | Type | Req | Constraints | Default | Description |
variant:  | Variant | Payload | Description |
enum:     | Value | Description |
```

---

## BEHAVIOR: parse-types-tables
Constraint: required

Extracts all TYPE blocks from a merged spec and produces the
TypesTableModel plus any structural diagnostics. Parsing and validation
run together; RULE-22 through RULE-25 (below) define the diagnostics.

INPUTS:
```
spec: merged spec text with per-line origin information
```

OUTPUTS:
```
model:       TypesTableModel      // empty when no TYPE blocks present
diagnostics: List<Diagnostic>     // RULE-22 .. RULE-25 findings
```

PRECONDITIONS:
- spec is the merged spec text (Includes already resolved)

STEPS:
1. Scan the `## TYPES` section line by line with the fence-depth
   counter active; lines inside fences are skipped.
2. On each line matching `^### TYPE: ` at column 0: open a new TYPE
   block; record name and line number.
3. Within an open block, read the `Kind:` line, the optional `Tag:`
   line, and the first following GFM table (header row, divider row,
   data rows). A table row is a column-0 line beginning with `|`.
4. Split rows on unescaped `|`; trim cells; unescape `\|` to `|`.
5. Apply RULE-22 (block and header structure).
6. Apply RULE-23 (cell vocabulary).
7. Apply RULE-24 (references and uniqueness).
8. Apply RULE-25 (lifecycle discipline).
   MECHANISM: rules are independent; a failure in one rule does not
   prevent subsequent rules from running. All diagnostics are collected.
9. Return the model and the collected diagnostics. A block that fails
   RULE-22 is excluded from the model; blocks failing only RULE-23,
   RULE-24, or RULE-25 remain in the model so downstream consumers can
   report against them.

POSTCONDITIONS:
- model.tables preserves document order
- every diagnostic carries the producing rule identifier (RULE-22 ..
  RULE-25) and a 1-based line number
- the input spec is not modified

---

## BEHAVIOR: types-table-validation-rules
Constraint: required

Defines the ordered rule set applied to TYPE blocks. These rules apply
only when at least one `### TYPE:` heading is present in the merged
spec's `## TYPES` section; specs without TYPE tables are unaffected.

STEPS:
1. Apply RULE-22 (TYPE block and table structure).
2. Apply RULE-23 (cell vocabulary).
3. Apply RULE-24 (references resolve; names unique).
4. Apply RULE-25 (lifecycle discipline).
   MECHANISM: all rules run; no short-circuit on first error.

### RULE-22: TYPE block and table structure

Severity: Error.

```
for each line matching "^### TYPE: " in ## TYPES:
  let N = text after "### TYPE: "
  if N does not match TypeName pattern:
    emit Error, section="TYPES",
      message="TYPE name '{N}' is not PascalCase ([A-Z][A-Za-z0-9]*)."

  if first non-blank line after the heading does not match "^Kind: ":
    emit Error, section="TYPES",
      message="TYPE '{N}' is missing the Kind: line.
               Kind: must be the first non-blank line after the heading."
  else:
    let K = value of Kind:
    if K not in [ "record", "variant", "enum" ]:
      emit Error, section="TYPES",
        message="TYPE '{N}' has invalid Kind: value '{K}'.
                 Valid values: record, variant, enum."

  if a "Tag:" line is present AND K != "variant":
    emit Error, section="TYPES",
      message="TYPE '{N}' declares Tag: but Kind is '{K}'.
               Tag: is permitted only for variant tables."

  if no GFM table follows before the next column-0 heading:
    emit Error, section="TYPES",
      message="TYPE '{N}' has no table."
  else:
    let H = header row column names
    if H does not begin with the exact required columns for K:
      emit Error, section="TYPES",
        message="TYPE '{N}' header does not match the {K} column
                 contract. Expected: {expected}. Found: {found}."
    if trailing columns of H are not a prefix-ordered subset of
       [ "Since", "Deprecated", "Replaced-by" ]:
      emit Error, section="TYPES",
        message="TYPE '{N}' has unknown or misordered trailing
                 column '{col}'. Optional columns, in order:
                 Since, Deprecated, Replaced-by."
    if the divider row is absent:
      emit Error, section="TYPES",
        message="TYPE '{N}' table is missing the divider row."
    for each data row R:
      if cell count of R != cell count of H:
        emit Error, section="TYPES",
          message="TYPE '{N}' row {i} has {found} cells;
                   header declares {expected}."
    if the table has zero data rows:
      emit Error, section="TYPES",
        message="TYPE '{N}' table has no rows."
```

### RULE-23: Cell vocabulary

Severity: Error.

```
for each TYPE table T that passed RULE-22:
  case T.kind of
    record:
      for each row R:
        if R.Field does not match FieldName: emit Error (pattern shown)
        if R.Type is not a valid TypeRef: emit Error
        if R.Req not in [ "yes", "no" ]: emit Error
        if R.Constraints != "-" and any token fails the
           ConstraintToken grammar: emit Error naming the token
        if a constraint token class does not fit R.Type
           (len on non-string; numeric Range on non-numeric;
            pattern on non-string): emit Error
        if R.Default != "-" and the literal is not compatible with
           R.Type per DefaultCell: emit Error
    variant:
      for each row R:
        if R.Variant does not match VariantName: emit Error
        if R.Payload != "-" and any entry fails
           "FieldName: TypeRef": emit Error
    enum:
      for each row R:
        if R.Value does not match EnumValue: emit Error
  if T.kind = "variant" and T has a Tag: line whose value does not
     match TagName: emit Error
```

Diagnostic message form:
`TYPE '{N}' row {i}: {column} value '{v}' is not valid: {reason}.`

### RULE-24: References resolve; names unique

Severity: Error.

```
let table_names = names of all TYPE tables in the merged spec

// Uniqueness across the merged spec
for each name occurring more than once in table_names:
  emit Error, section="TYPES",
    message="TYPE table '{N}' is defined more than once."
for each name in table_names also declared with := in prose TYPES:
  emit Error, section="TYPES",
    message="TYPE '{N}' is defined both as a table and as a prose
             TYPES declaration. One authoritative definition only."

// Uniqueness within a table
for each table T:
  record:  duplicate Field values -> Error
  variant: duplicate Variant values -> Error
  enum:    duplicate Value values -> Error

// References
for each TypeRef mentioning a TypeName (field Type cells and
    variant Payload entries):
  if the TypeName not in table_names:
    emit Error, section="TYPES",
      message="TYPE '{N}' references '{ref}', which is not defined
               by any TYPE table in the merged spec."
```

Prose TYPES declarations may continue to reference table-defined names;
that direction is unconstrained.

### RULE-25: Lifecycle discipline

Severity: Warning.

```
for each table row (or table) carrying lifecycle columns:
  if Since != "-" and Since does not match SpecVersion:
    emit Warning: "invalid Since value '{v}'"
  if Deprecated not in [ "yes", "-" ]:
    emit Warning: "invalid Deprecated value '{v}'"
  if Deprecated = "yes" and Replaced-by column absent or "-":
    emit Warning, section="TYPES",
      message="'{name}' is Deprecated without Replaced-by.
               Deprecated entries should point to their successor."
  if Replaced-by != "-" and the target does not resolve to a
     TYPE table name, or to a Field/Variant/Value within the same
     table's namespace:
    emit Warning: "Replaced-by target '{v}' does not resolve."
```

Identifier permanence is a policy, not a statically checkable property:
a name, once published, is never reused for a different meaning.
Removal of a table or row is a spec-Version major event; deprecation
with Replaced-by is the migration path. RULE-25 warns on the checkable
part; the policy itself is stated here normatively.

---

## PRECONDITIONS

- Rules in this file apply only when the merged spec contains at least
  one `### TYPE:` heading inside `## TYPES`
- Rule evaluation receives the merged spec text (Includes resolved)

---

## POSTCONDITIONS

- Rule evaluation is read-only: no file modification, no network
  calls, no environment-variable reads for behaviour control
- Every diagnostic carries rule identifier, severity, section, message,
  and a 1-based line number

---

## INVARIANTS

- [observable]      a spec without `### TYPE:` headings produces no
  RULE-22 through RULE-25 diagnostics
- [observable]      all four rules are evaluated; rule processing does
  not short-circuit on the first Error
- [observable]      cell vocabularies are closed: any token outside the
  grammars defined in TYPES is a diagnostic, never a silent pass-through
- [implementation]  parsing preserves document order of tables and rows;
  emission (types-emit-rules.md) depends on this order

---

## CHANGELOG

- 2026.07.07.01 - Initial version. Tabular TYPES notation: TYPE block
  grammar (record, variant, enum), closed cell vocabularies
  (TypeRef, ConstraintToken, DefaultCell, PayloadCell, lifecycle
  columns), parse-types-tables BEHAVIOR, and validation rules RULE-22
  through RULE-25. Consumed by libpcd via Includes.
