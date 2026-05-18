# pcd-lint multi-pass EXAMPLE

## META
Deployment:   cli-tool
Version:      0.1.0
Spec-Schema:  0.1.0
Author:       Jane Example <jane@example.org>
License:      Apache-2.0
Verification: none
Safety-Level: QM

---

## TYPES

---

## BEHAVIOR: reconcile
STEPS:
  1. on failure → exit 1

---

## PRECONDITIONS

---

## POSTCONDITIONS

---

## INVARIANTS

---

## EXAMPLES

### EXAMPLE: reconcile_graceful_stop
GIVEN:
  VM "testvm-01", spec.desiredState = Stopped
  Domain is Running
WHEN:  reconcile runs (pass 1)
THEN:
  domain.Shutdown() is called
  result = RequeueAfter(10s)
WHEN:  reconcile runs (pass 2); domain is Shutoff
THEN:
  status.phase = Stopped
  result = RequeueAfter(60s)
