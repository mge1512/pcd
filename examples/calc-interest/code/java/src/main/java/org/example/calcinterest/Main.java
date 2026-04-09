package org.example.calcinterest;

import java.io.BufferedReader;
import java.io.InputStreamReader;
import java.math.BigDecimal;
import java.math.RoundingMode;

/**
 * calc-interest — Simple interest calculator.
 *
 * <p>Reads principal, annual rate, and number of periods from standard input
 * (one value per line), computes simple interest and total repayment amount,
 * then writes results to standard output.
 *
 * <p>Exit codes:
 *   0 — success
 *   1 — read failure or arithmetic overflow
 *   2 — invalid input value
 *
 * <p>Specification: calc-interest v0.1.0 (Spec-Schema 0.3.21)
 * License: Apache-2.0
 */
public final class Main {

    // -----------------------------------------------------------------------
    // Domain constraints (from spec TYPES section)
    // -----------------------------------------------------------------------

    /** Principal: decimal, value > 0 and value <= 9999999.99 */
    private static final BigDecimal PRINCIPAL_MAX = new BigDecimal("9999999.99");

    /** Rate: decimal, value > 0 and value <= 999.9999 */
    private static final BigDecimal RATE_MAX = new BigDecimal("999.9999");

    /** Periods: integer, value >= 1 and value <= 999 */
    private static final int PERIODS_MIN = 1;
    private static final int PERIODS_MAX = 999;

    /** Interest / Total upper bound */
    private static final BigDecimal RESULT_MAX = new BigDecimal("9999999.99");

    // -----------------------------------------------------------------------
    // Exit codes (from spec ERRORS section and template EXIT-CODE-* rows)
    // -----------------------------------------------------------------------
    private static final int EXIT_OK              = 0;
    private static final int EXIT_READ_OVERFLOW   = 1;
    private static final int EXIT_INVALID_INPUT   = 2;

    // -----------------------------------------------------------------------
    // Private constructor — utility class, not instantiated
    // -----------------------------------------------------------------------
    private Main() {}

    // -----------------------------------------------------------------------
    // Entry point
    // -----------------------------------------------------------------------

    /**
     * Main entry point.
     *
     * <p>BEHAVIOR: calculate-simple-interest — all STEPS implemented in order.
     *
     * @param args command-line arguments (not used; tool reads from stdin)
     */
    public static void main(String[] args) {
        // Install clean-exit signal handlers for SIGTERM and SIGINT
        // (template SIGNAL-HANDLING: required for both SIGTERM and SIGINT)
        Runtime.getRuntime().addShutdownHook(new Thread(() -> {
            // Shutdown hook fires on SIGTERM / SIGINT.
            // No partial output cleanup needed here because we write
            // to stdout only after all computation succeeds (step 9-10).
        }));

        BufferedReader reader = new BufferedReader(new InputStreamReader(System.in));

        // STEP 1: Read principal from stdin; on failure → exit 1, write error to stderr.
        BigDecimal principal = readDecimal(reader, "principal");

        // STEP 2: Read rate from stdin; on failure → exit 1, write error to stderr.
        BigDecimal rate = readDecimal(reader, "rate");

        // STEP 3: Read periods from stdin; on failure → exit 1, write error to stderr.
        int periods = readInteger(reader, "periods");

        // STEP 4: Validate principal > 0; on failure → exit 2, write "invalid principal".
        if (principal.compareTo(BigDecimal.ZERO) <= 0 || principal.compareTo(PRINCIPAL_MAX) > 0) {
            System.err.println("invalid principal");
            System.exit(EXIT_INVALID_INPUT);
        }

        // STEP 5: Validate rate > 0; on failure → exit 2, write "invalid rate".
        if (rate.compareTo(BigDecimal.ZERO) <= 0 || rate.compareTo(RATE_MAX) > 0) {
            System.err.println("invalid rate");
            System.exit(EXIT_INVALID_INPUT);
        }

        // STEP 6: Validate periods >= 1; on failure → exit 2, write "invalid periods".
        if (periods < PERIODS_MIN || periods > PERIODS_MAX) {
            System.err.println("invalid periods");
            System.exit(EXIT_INVALID_INPUT);
        }

        // STEP 7: Compute interest = principal * rate * periods;
        //         on overflow → exit 1, write error to stderr.
        BigDecimal periodsDecimal = BigDecimal.valueOf(periods);
        BigDecimal interest;
        try {
            interest = principal.multiply(rate).multiply(periodsDecimal)
                                .setScale(2, RoundingMode.HALF_UP);
        } catch (ArithmeticException e) {
            System.err.println("arithmetic error computing interest: " + e.getMessage());
            System.exit(EXIT_READ_OVERFLOW);
            return; // unreachable, satisfies compiler
        }
        if (interest.compareTo(RESULT_MAX) > 0) {
            System.err.println("arithmetic overflow: interest exceeds maximum value");
            System.exit(EXIT_READ_OVERFLOW);
        }

        // STEP 8: Compute total = principal + interest;
        //         on overflow → exit 1, write error to stderr.
        BigDecimal total;
        try {
            total = principal.add(interest).setScale(2, RoundingMode.HALF_UP);
        } catch (ArithmeticException e) {
            System.err.println("arithmetic error computing total: " + e.getMessage());
            System.exit(EXIT_READ_OVERFLOW);
            return; // unreachable, satisfies compiler
        }
        if (total.compareTo(RESULT_MAX) > 0) {
            System.err.println("arithmetic overflow: total exceeds maximum value");
            System.exit(EXIT_READ_OVERFLOW);
        }

        // STEP 9: Write "INTEREST: {interest}" to stdout with 2 decimal places.
        System.out.printf("INTEREST: %.2f%n", interest);

        // STEP 10: Write "TOTAL:    {total}" to stdout with 2 decimal places.
        System.out.printf("TOTAL:    %.2f%n", total);

        // STEP 11: Exit with code 0.
        System.exit(EXIT_OK);
    }

    // -----------------------------------------------------------------------
    // Helper: read a decimal value from stdin
    // -----------------------------------------------------------------------

    /**
     * Reads one line from {@code reader} and parses it as a {@link BigDecimal}.
     *
     * <p>On any I/O error or parse failure, writes a message to stderr and
     * exits with code 1 (EXIT_READ_OVERFLOW), as specified in STEP 1-3.
     *
     * @param reader the buffered reader wrapping stdin
     * @param fieldName human-readable field name used in error messages
     * @return the parsed {@link BigDecimal}
     */
    static BigDecimal readDecimal(BufferedReader reader, String fieldName) {
        String line;
        try {
            line = reader.readLine();
        } catch (java.io.IOException e) {
            System.err.println("error reading " + fieldName + ": " + e.getMessage());
            System.exit(EXIT_READ_OVERFLOW);
            return null; // unreachable
        }
        if (line == null) {
            System.err.println("error reading " + fieldName + ": unexpected end of input");
            System.exit(EXIT_READ_OVERFLOW);
            return null; // unreachable
        }
        try {
            return new BigDecimal(line.trim());
        } catch (NumberFormatException e) {
            System.err.println("error reading " + fieldName + ": not a valid number: " + line.trim());
            System.exit(EXIT_READ_OVERFLOW);
            return null; // unreachable
        }
    }

    // -----------------------------------------------------------------------
    // Helper: read an integer value from stdin
    // -----------------------------------------------------------------------

    /**
     * Reads one line from {@code reader} and parses it as an {@code int}.
     *
     * <p>On any I/O error or parse failure, writes a message to stderr and
     * exits with code 1 (EXIT_READ_OVERFLOW), as specified in STEP 3.
     *
     * @param reader the buffered reader wrapping stdin
     * @param fieldName human-readable field name used in error messages
     * @return the parsed integer value
     */
    static int readInteger(BufferedReader reader, String fieldName) {
        String line;
        try {
            line = reader.readLine();
        } catch (java.io.IOException e) {
            System.err.println("error reading " + fieldName + ": " + e.getMessage());
            System.exit(EXIT_READ_OVERFLOW);
            return 0; // unreachable
        }
        if (line == null) {
            System.err.println("error reading " + fieldName + ": unexpected end of input");
            System.exit(EXIT_READ_OVERFLOW);
            return 0; // unreachable
        }
        try {
            return Integer.parseInt(line.trim());
        } catch (NumberFormatException e) {
            System.err.println("error reading " + fieldName + ": not a valid integer: " + line.trim());
            System.exit(EXIT_READ_OVERFLOW);
            return 0; // unreachable
        }
    }
}
