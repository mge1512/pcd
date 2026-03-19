# Account Transfer

## META
Deployment:  backend-service
Version:     0.3.7
Spec-Schema: 0.3.7
Author:      Matthias G. Eckermann <pcdp@mailbox.org>
License:     CC-BY-4.0
Verification: lean4
Safety-Level: financial-integrity-critical

---

## TYPES

```
AccountId := u64 where id > 0

Balance := i64 where balance >= 0

Amount := i64 where amount > 0

Account := {
  id:      AccountId,
  balance: Balance
}

ErrorCode := INSUFFICIENT_FUNDS | SAME_ACCOUNT | INVALID_AMOUNT

TransferResult := Ok | Err(ErrorCode)
```

---

## BEHAVIOR: transfer

INPUTS:
```
from:   Account
to:     Account
amount: Amount
```

OUTPUTS:
```
result: TransferResult
```

PRECONDITIONS:
- from.balance >= amount
- from.id ≠ to.id
- amount > 0

POSTCONDITIONS:
- result = Ok ⟹ from.balance' = from.balance - amount
- result = Ok ⟹ to.balance' = to.balance + amount
- result = Ok ⟹ ∀ other: Account. other ∉ {from, to} ⟹ other.balance' = other.balance
- result = Err(_) ⟹ from.balance' = from.balance ∧ to.balance' = to.balance

SIDE-EFFECTS:
- Creates transfer_log entry with timestamp, from.id, to.id, amount, result

---

## PRECONDITIONS

- from and to are valid accounts with positive ids
- amount is a positive integer
- Database transaction context is established before calling transfer
- Optimistic lock on account.balance is held by caller

---

## POSTCONDITIONS

- Total system balance is conserved: Σ(all_balances)' = Σ(all_balances)
- All balances remain non-negative after any transfer
- Transfer log entry is created for every invocation regardless of result
- No partial state: either both balances are updated or neither is

---

## INVARIANTS

- GLOBAL: ∀ a: Account. a.balance >= 0
- GLOBAL: Σ(all_balances) is constant across all transfer operations
- GLOBAL: result = Ok ⟹ from.balance' = from.balance - amount
- GLOBAL: result = Err(_) ⟹ from.balance' = from.balance ∧ to.balance' = to.balance
- GLOBAL: transfer is idempotent when combined with a unique transfer_id

---

## EXAMPLES

EXAMPLE: successful_transfer
GIVEN:
  from = Account { id: 1, balance: 100 }
  to   = Account { id: 2, balance: 50 }
  amount = 30
WHEN:
  result = transfer(from, to, amount)
THEN:
  result = Ok
  from.balance = 70
  to.balance = 80
  Σ(balances) = 150  // conservation holds

EXAMPLE: insufficient_funds
GIVEN:
  from = Account { id: 1, balance: 20 }
  to   = Account { id: 2, balance: 50 }
  amount = 30
WHEN:
  result = transfer(from, to, amount)
THEN:
  result = Err(INSUFFICIENT_FUNDS)
  from.balance = 20  // unchanged
  to.balance = 50    // unchanged

EXAMPLE: same_account_rejection
GIVEN:
  from = Account { id: 1, balance: 100 }
  to   = Account { id: 1, balance: 100 }  // same account
  amount = 30
WHEN:
  result = transfer(from, to, amount)
THEN:
  result = Err(SAME_ACCOUNT)
  from.balance = 100  // unchanged

EXAMPLE: zero_amount_rejection
GIVEN:
  from = Account { id: 1, balance: 100 }
  to   = Account { id: 2, balance: 50 }
  amount = 0
WHEN:
  result = transfer(from, to, amount)
THEN:
  result = Err(INVALID_AMOUNT)
  from.balance = 100  // unchanged
  to.balance = 50     // unchanged

EXAMPLE: exact_balance_transfer
GIVEN:
  from = Account { id: 1, balance: 50 }
  to   = Account { id: 2, balance: 0 }
  amount = 50
WHEN:
  result = transfer(from, to, amount)
THEN:
  result = Ok
  from.balance = 0   // exactly drained
  to.balance = 50
  Σ(balances) = 50   // conservation holds

---

## DEPLOYMENT

Runtime: Backend REST API endpoint /api/v1/transfer
Database: PostgreSQL, requires SERIALIZABLE transaction isolation
Concurrency: Multiple instances, optimistic locking on account.balance
Monitoring: Prometheus metrics on transfer_success, transfer_failure
Logging: All attempts logged with user_id, from_id, to_id, amount, result
Error Handling: Return structured errors, never panic
Idempotency: Caller provides unique transfer_id; repeated calls with same
             transfer_id return original result without re-executing
