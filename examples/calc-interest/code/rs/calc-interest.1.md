% CALC-INTEREST(1) calc-interest 0.1.0
% Unknown
% April 2026

# NAME

calc-interest — simple interest calculator

# SYNOPSIS

**calc-interest** < *input-file*

echo -e "PRINCIPAL\nRATE\nPERIODS" | **calc-interest**

# DESCRIPTION

**calc-interest** reads three numeric values from standard input (one per
line) and computes simple (flat) interest and total repayment amount.

The three input values must be provided on separate lines:

1. **principal** — the loan or investment amount (positive decimal, max 9 999 999.99)
2. **rate** — the interest rate as a decimal fraction (e.g. 0.0350 for 3.50%; positive decimal, max 999.9999)
3. **periods** — the number of time periods (positive integer 1–999)

The tool writes exactly two lines to standard output:

```
INTEREST: <value>
TOTAL:    <value>
```

Both values are formatted to two decimal places.

The formula used is **simple (flat) interest**, not compound interest:

    interest = principal × rate × periods
    total    = principal + interest

# OPTIONS

This tool takes no command-line options or flags. All input is read from
standard input.

# EXIT CODES

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | Read failure or arithmetic overflow |
| 2 | Invalid input value (non-positive principal, non-positive rate, periods < 1) |

# EXAMPLES

Calculate interest on a 10 000.00 loan at 3.50% for 12 periods:

```
$ echo -e "10000.00\n0.0350\n12" | calc-interest
INTEREST: 4200.00
TOTAL:    14200.00
```

# DIAGNOSTICS

Error messages are written to standard error. Standard output is empty or
partial on any error.

# NOTES

- No network access is performed.
- No files are read or written beyond stdin/stdout/stderr.
- The tool is idempotent: identical inputs always produce identical outputs.
- Rate is a decimal fraction, not a percentage integer (use 0.035, not 3.5).

# SEE ALSO

No related commands.

# AUTHOR

Unknown

# COPYRIGHT

Copyright (c) Unknown. Licensed under the Apache License, Version 2.0.
See <https://www.apache.org/licenses/LICENSE-2.0>.
