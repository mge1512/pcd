# pcd-lint multiple authors

## META
Deployment:   cli-tool
Version:      0.1.0
Spec-Schema:  0.1.0
Author:       Jane Example <jane@example.org>
Author:       John Example <john@example.org>
License:      Apache-2.0
Verification: none
Safety-Level: QM

---

## TYPES

---

## BEHAVIOR: lint

---

## PRECONDITIONS

---

## POSTCONDITIONS

---

## INVARIANTS

---

## EXAMPLES

### EXAMPLE: multiple_authors
GIVEN:
  a valid spec with multiple authors
WHEN:
  pcd-lint is run
THEN:
  exit_code = 0
