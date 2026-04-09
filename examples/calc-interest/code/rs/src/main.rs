// calc-interest — simple interest calculator
// Spec: calc-interest.spec.md v0.1.0
// License: Apache-2.0
//
// Reads principal, rate, and periods from stdin (one value per line),
// computes simple interest and total, then writes results to stdout.
//
// Exit codes:
//   0 — success
//   1 — read failure or arithmetic overflow
//   2 — invalid input value

use std::io::{self, BufRead, Write};
use std::process;

// ── Type constraints (from spec TYPES) ──────────────────────────────────────
//
// Principal : decimal  0 < v <= 9_999_999.99   (COBOL 9(7)V99)
// Rate      : decimal  0 < v <= 999.9999        (COBOL 9(3)V9999)
// Periods   : integer  1 <= v <= 999            (COBOL 9(3))
// Interest  : decimal  0 <= v <= 9_999_999.99
// Total     : decimal  0 <= v <= 9_999_999.99

const PRINCIPAL_MAX: f64 = 9_999_999.99;
const RATE_MAX: f64 = 999.9999;
const PERIODS_MAX: u32 = 999;
const MONETARY_MAX: f64 = 9_999_999.99;

fn main() {
    let stdin = io::stdin();
    let mut lines = stdin.lock().lines();

    // ── Step 1: Read principal ───────────────────────────────────────────────
    let principal_str = match lines.next() {
        Some(Ok(line)) => line,
        Some(Err(e)) => {
            eprintln!("error reading principal: {}", e);
            process::exit(1);
        }
        None => {
            eprintln!("error reading principal: unexpected end of input");
            process::exit(1);
        }
    };
    let principal: f64 = match principal_str.trim().parse() {
        Ok(v) => v,
        Err(e) => {
            eprintln!("error reading principal: {}", e);
            process::exit(1);
        }
    };

    // ── Step 2: Read rate ────────────────────────────────────────────────────
    let rate_str = match lines.next() {
        Some(Ok(line)) => line,
        Some(Err(e)) => {
            eprintln!("error reading rate: {}", e);
            process::exit(1);
        }
        None => {
            eprintln!("error reading rate: unexpected end of input");
            process::exit(1);
        }
    };
    let rate: f64 = match rate_str.trim().parse() {
        Ok(v) => v,
        Err(e) => {
            eprintln!("error reading rate: {}", e);
            process::exit(1);
        }
    };

    // ── Step 3: Read periods ─────────────────────────────────────────────────
    let periods_str = match lines.next() {
        Some(Ok(line)) => line,
        Some(Err(e)) => {
            eprintln!("error reading periods: {}", e);
            process::exit(1);
        }
        None => {
            eprintln!("error reading periods: unexpected end of input");
            process::exit(1);
        }
    };
    let periods: u32 = match periods_str.trim().parse() {
        Ok(v) => v,
        Err(e) => {
            eprintln!("error reading periods: {}", e);
            process::exit(1);
        }
    };

    // ── Step 4: Validate principal > 0 ──────────────────────────────────────
    if principal <= 0.0 || principal > PRINCIPAL_MAX {
        eprintln!("invalid principal");
        process::exit(2);
    }

    // ── Step 5: Validate rate > 0 ────────────────────────────────────────────
    if rate <= 0.0 || rate > RATE_MAX {
        eprintln!("invalid rate");
        process::exit(2);
    }

    // ── Step 6: Validate periods >= 1 ───────────────────────────────────────
    if periods < 1 || periods > PERIODS_MAX {
        eprintln!("invalid periods");
        process::exit(2);
    }

    // ── Step 7: Compute interest = principal * rate * periods ────────────────
    let interest = principal * rate * (periods as f64);
    if !interest.is_finite() || interest > MONETARY_MAX {
        eprintln!("error: arithmetic overflow computing interest");
        process::exit(1);
    }

    // ── Step 8: Compute total = principal + interest ─────────────────────────
    let total = principal + interest;
    if !total.is_finite() || total > MONETARY_MAX {
        eprintln!("error: arithmetic overflow computing total");
        process::exit(1);
    }

    // ── Step 9: Write "INTEREST: {interest}" to stdout (2 decimal places) ────
    // ── Step 10: Write "TOTAL:    {total}" to stdout (2 decimal places) ──────
    // Label "TOTAL:    " has 4 trailing spaces to match COBOL DISPLAY alignment.
    let stdout = io::stdout();
    let mut out = stdout.lock();
    writeln!(out, "INTEREST: {:.2}", interest).unwrap_or_else(|e| {
        eprintln!("error writing output: {}", e);
        process::exit(1);
    });
    writeln!(out, "TOTAL:    {:.2}", total).unwrap_or_else(|e| {
        eprintln!("error writing output: {}", e);
        process::exit(1);
    });

    // ── Step 11: Exit with code 0 ────────────────────────────────────────────
    process::exit(0);
}
