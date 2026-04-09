# calc-interest

Simple interest calculator — reads principal, rate, and periods from standard
input; writes computed interest and total to standard output.

## Installation

### openSUSE / SLES (zypper via OBS)

```sh
zypper addrepo https://download.opensuse.org/repositories/home:/unknown/openSUSE_Leap_15.6/ calc-interest
zypper refresh
zypper install calc-interest
```

### Debian / Ubuntu (apt via OBS)

```sh
echo "deb https://download.opensuse.org/repositories/home:/unknown/Debian_12/ ./" \
  | sudo tee /etc/apt/sources.list.d/calc-interest.list
sudo apt-get update
sudo apt-get install calc-interest
```

### Fedora / RHEL (dnf via OBS)

```sh
dnf config-manager --add-repo \
  https://download.opensuse.org/repositories/home:/unknown/Fedora_40/home:unknown.repo
dnf install calc-interest
```

> **Note:** OBS repository URLs above are illustrative. Replace with the actual
> published repository URL once the package is built on build.opensuse.org.

### Build from source

Requires Rust 1.70+ and Cargo.

```sh
cargo build --release
# Static binary at: target/release/calc-interest
```

## Usage

```
echo -e "PRINCIPAL\nRATE\nPERIODS" | calc-interest
```

**Input** (three lines on stdin):

| Line | Field | Description |
|------|-------|-------------|
| 1 | principal | Positive decimal, e.g. `10000.00` (max 9 999 999.99) |
| 2 | rate | Decimal fraction, e.g. `0.0350` for 3.50% (max 999.9999) |
| 3 | periods | Positive integer, e.g. `12` (max 999) |

**Output** (two lines on stdout):

```
INTEREST: <value>
TOTAL:    <value>
```

### Example

```sh
$ echo -e "10000.00\n0.0350\n12" | calc-interest
INTEREST: 4200.00
TOTAL:    14200.00
```

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | Read failure or arithmetic overflow |
| 2 | Invalid input value |

## Flags

None. This tool has no command-line flags or options.

## Notes

- Rate is a **decimal fraction** (0.035 = 3.5%), not a percentage integer.
- Formula: `interest = principal × rate × periods` (simple interest, no compounding).
- No network access. No file I/O beyond stdin/stdout/stderr.
- Idempotent: identical inputs always produce identical outputs.

## License

Apache-2.0 — see [LICENSE](LICENSE) and <https://www.apache.org/licenses/LICENSE-2.0>.
