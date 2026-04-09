# calc-interest

Simple interest calculator — CLI tool

**Version:** 0.1.0  
**License:** Apache-2.0  
**Spec-Schema:** 0.3.21

---

## Overview

`calc-interest` reads principal, annual rate, and number of periods from
standard input (one value per line), computes simple interest and total
repayment amount, and writes the results to standard output.

**Formula:**

```
interest = principal × rate × periods
total    = principal + interest
```

No compounding. No network access. No file I/O beyond stdin/stdout/stderr.

---

## Installation

### openSUSE / SLES

```sh
zypper install calc-interest
```

### Debian / Ubuntu

```sh
apt install calc-interest
```

### Fedora / RHEL

```sh
dnf install calc-interest
```

Packages are distributed via [build.opensuse.org](https://build.opensuse.org).  
Curl-based installation is **not** supported (supply chain security requirement).

---

## Usage

```sh
echo -e "PRINCIPAL\nRATE\nPERIODS" | calc-interest
```

### Input format

Three values on separate lines via stdin:

| Line | Field     | Type    | Constraint                  | Example  |
|------|-----------|---------|-----------------------------|----------|
| 1    | principal | decimal | > 0, ≤ 9 999 999.99        | 10000.00 |
| 2    | rate      | decimal | > 0, ≤ 999.9999 (fraction) | 0.0350   |
| 3    | periods   | integer | ≥ 1, ≤ 999                 | 12       |

Rate is a **decimal fraction** (e.g. `0.0350` = 3.50%).

### Output format

```
INTEREST: <value>
TOTAL:    <value>
```

Both values are formatted to two decimal places.

---

## Example

```sh
echo -e "10000.00\n0.0350\n12" | calc-interest
```

```
INTEREST: 4200.00
TOTAL:    14200.00
```

---

## Exit codes

| Code | Meaning                                                   |
|------|-----------------------------------------------------------|
| 0    | Success                                                   |
| 1    | Read failure or arithmetic overflow                       |
| 2    | Invalid input (non-positive principal/rate, periods < 1)  |

---

## Building from source

Requirements: JDK 17+, Maven 3.8+, pandoc (for man page)

```sh
# Compile and package
make build

# Run tests
make test

# Generate man page
make man

# Install to /usr/local
make install PREFIX=/usr/local
```

---

## Running tests

```sh
mvn test
```

All tests in `src/test/java/org/example/calcinterest/CalcInterestTest.java`
run without any live external service.

---

## Flags

This tool takes no command-line flags. All input is provided via stdin.

---

## License

Apache License, Version 2.0  
<https://www.apache.org/licenses/LICENSE-2.0>
