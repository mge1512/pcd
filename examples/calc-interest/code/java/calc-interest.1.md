% calc-interest(1) Version 0.1.0 | Simple Interest Calculator
% Unknown
% April 2026

# NAME

calc-interest — compute simple interest from principal, rate, and periods

# SYNOPSIS

```
echo -e "PRINCIPAL\nRATE\nPERIODS" | calc-interest
```

# DESCRIPTION

**calc-interest** reads three numeric values from standard input (one per
line), computes simple interest using the formula:

```
interest = principal × rate × periods
total    = principal + interest
```

and writes the results to standard output.

This is a *simple interest* calculation — there is no compounding.

# INPUT FORMAT

Three values are read from **stdin**, each on its own line:

1. **principal** — positive decimal, e.g. `10000.00` (max 9999999.99)
2. **rate** — positive decimal fraction, e.g. `0.0350` for 3.50% (max 999.9999)
3. **periods** — positive integer count of time periods, e.g. `12` (max 999)

# OUTPUT FORMAT

On success, exactly two lines are written to **stdout**:

```
INTEREST: <value>
TOTAL:    <value>
```

Both values are formatted to two decimal places.

# EXIT STATUS

| Code | Meaning |
|------|---------|
| 0    | Success |
| 1    | Read failure or arithmetic overflow |
| 2    | Invalid input value (non-positive principal or rate, or periods < 1) |

# EXAMPLES

Compute interest on a 10 000 principal at 3.50% for 12 periods:

```sh
echo -e "10000.00\n0.0350\n12" | calc-interest
```

Output:

```
INTEREST: 4200.00
TOTAL:    14200.00
```

# INSTALLATION

## openSUSE / SLES (zypper)

```sh
zypper install calc-interest
```

## Debian / Ubuntu (apt)

```sh
apt install calc-interest
```

## Fedora / RHEL (dnf)

```sh
dnf install calc-interest
```

Packages are distributed via [build.opensuse.org](https://build.opensuse.org)
(OBS). Curl-based installation is not supported.

# NOTES

- No network access is required or performed.
- No files are read or written beyond stdin/stdout/stderr.
- The tool is idempotent: identical inputs always produce identical outputs.
- Rate is expressed as a decimal fraction (e.g. `0.0350` = 3.50%), not as a percentage.

# SEE ALSO

**bc**(1), **awk**(1)

# LICENSE

Apache-2.0 — <https://www.apache.org/licenses/LICENSE-2.0>
