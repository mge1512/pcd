# calc-interest

## META

Deployment:   cli-tool
Version:      0.1.0
Spec-Schema:  0.3.21
Author:       Unknown
License:      Apache-2.0
Verification: none
Safety-Level: QM

## TYPES

```
Principal := decimal where value > 0 and value <= 9999999.99
  // Maximum 7 integer digits, 2 decimal places (from COBOL PIC 9(7)V99)

Rate := decimal where value > 0 and value <= 999.9999
  // Maximum 3 integer digits, 4 decimal places (from COBOL PIC 9(3)V9999)
  // Expressed as a decimal fraction, e.g. 0.0350 for 3.50%

Periods := integer where value >= 1 and value <= 999
  // Number of time periods (e.g. months), maximum 3 digits

Interest := decimal where value >= 0 and value <= 9999999.99
  // Computed simple interest: Principal * Rate * Periods

Total := decimal where value >= 0 and value <= 9999999.99
  // Total amount: Principal + Interest

InterestResult := {
  interest: Interest,
  total:    Total
}
```

## BEHAVIOR: calculate-simple-interest
Constraint: required

Reads principal, annual rate, and number of periods from standard input,
computes simple interest and total repayment amount, then writes results
to standard output.

INPUTS:
```
principal: Principal   // read from stdin, line 1
rate:      Rate        // read from stdin, line 2; decimal fraction (e.g. 0.0350)
periods:   Periods     // read from stdin, line 3; integer count of periods
```

PRECONDITIONS:
- principal is a positive decimal value within the range of Principal
- rate is a positive decimal value within the range of Rate
- periods is a positive integer within the range of Periods
- All three values are provided on separate lines via stdin

STEPS:
1. Read principal from stdin; on failure → exit with code 1, write error to stderr.
2. Read rate from stdin; on failure → exit with code 1, write error to stderr.
3. Read periods from stdin; on failure → exit with code 1, write error to stderr.
4. Validate principal > 0; on failure → exit with code 2, write "invalid principal" to stderr.
5. Validate rate > 0; on failure → exit with code 2, write "invalid rate" to stderr.
6. Validate periods >= 1; on failure → exit with code 2, write "invalid periods" to stderr.
7. Compute interest = principal * rate * periods; on overflow → exit with code 1, write error to stderr.
8. Compute total = principal + interest; on overflow → exit with code 1, write error to stderr.
9. Write "INTEREST: {interest}" to stdout with 2 decimal places.
10. Write "TOTAL:    {total}" to stdout with 2 decimal places.
11. Exit with code 0.

POSTCONDITIONS:
- stdout contains exactly two lines: one beginning "INTEREST: " and one beginning "TOTAL:    "
- interest equals principal multiplied by rate multiplied by periods
- total equals principal plus interest
- stderr is empty on success

ERRORS:
- exit 1: read failure or arithmetic overflow
- exit 2: invalid input value (non-positive principal, non-positive rate, periods < 1)

## PRECONDITIONS

- The runtime environment provides stdin connected to a source of three newline-separated numeric values.
- No network access is required or permitted.
- No file system access beyond stdin/stdout/stderr is required.

## POSTCONDITIONS

- On success, stdout contains exactly the INTEREST and TOTAL lines.
- On any error, stderr contains a human-readable message and stdout is empty or partial.
- The tool is idempotent: identical inputs always produce identical outputs.

## INVARIANTS

- [observable]  interest = principal * rate * periods (simple interest formula, no compounding)
- [observable]  total = principal + interest
- [observable]  output lines preserve the label format "INTEREST: " and "TOTAL:    " (with trailing spaces for alignment)
- [implementation]  numeric precision follows the COBOL source: 2 decimal places for monetary values, 4 decimal places for rate
- [observable]  the tool reads exactly three values from stdin and produces exactly two lines on stdout

## EXAMPLES

EXAMPLE: typical_calculation
GIVEN:
  stdin contains:
    10000.00
    0.0350
    12
WHEN:
  calc-interest reads principal=10000.00, rate=0.0350, periods=12
THEN:
  stdout contains:
    INTEREST: 4200.00
    TOTAL:    14200.00
  exit code is 0

EXAMPLE: zero_rate_rejected
GIVEN:
  stdin contains:
    10000.00
    0.0000
    12
WHEN:
  calc-interest reads principal=10000.00, rate=0.0000, periods=12
THEN:
  stderr contains "invalid rate"
  exit code is 2

EXAMPLE: zero_principal_rejected
GIVEN:
  stdin contains:
    0.00
    0.0350
    12
WHEN:
  calc-interest reads principal=0.00, rate=0.0350, periods=12
THEN:
  stderr contains "invalid principal"
  exit code is 2

EXAMPLE: zero_periods_rejected
GIVEN:
  stdin contains:
    10000.00
    0.0350
    0
WHEN:
  calc-interest reads principal=10000.00, rate=0.0350, periods=0
THEN:
  stderr contains "invalid periods"
  exit code is 2

EXAMPLE: non_numeric_input_rejected
GIVEN:
  stdin contains:
    abc
    0.0350
    12
WHEN:
  calc-interest attempts to read principal
THEN:
  stderr contains an error message
  exit code is 1

## DEPENDENCIES

None. The tool uses only the standard library of the target language.

## DEPLOYMENT

Runtime: command-line tool executed in a shell or pipeline.
Reads three numeric values from stdin (one per line) and writes two
result lines to stdout. No configuration files, no network access,
no persistent state.

Reference implementation: COBOL source `calc-interest.cob`, compiled
with GnuCOBOL (`cobc -x`). The specification targets a modern language
reimplementation; the COBOL source is the authoritative behavioral reference.

Invocation example:
  echo -e "10000.00\n0.0350\n12" | ./calc-interest
