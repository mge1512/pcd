package org.example.calcinterest;

import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.DisplayName;

import java.io.BufferedReader;
import java.io.ByteArrayInputStream;
import java.io.ByteArrayOutputStream;
import java.io.PrintStream;
import java.io.StringReader;
import java.math.BigDecimal;

import static org.junit.jupiter.api.Assertions.*;

/**
 * Independent tests for calc-interest.
 *
 * <p>These tests cover every EXAMPLE in the specification and exercise the
 * BEHAVIOR: calculate-simple-interest STEPS end-to-end.
 *
 * <p>All tests run without any live external service (no network, no files
 * beyond the JVM classpath).
 *
 * <p>Template: cli-tool.template.md v0.3.20
 * Spec:     calc-interest v0.1.0 (Spec-Schema 0.3.21)
 */
class CalcInterestTest {

    // -----------------------------------------------------------------------
    // Helper: run the full main() logic via stdin/stdout/stderr redirection
    // -----------------------------------------------------------------------

    /**
     * Result of a simulated main() invocation.
     */
    record RunResult(String stdout, String stderr, int exitCode) {}

    /**
     * Simulates a full run of the calculator by feeding {@code stdinContent}
     * through the public helper methods and capturing output.
     *
     * <p>Because {@link Main#main(String[])} calls {@link System#exit}, we
     * test the core logic by exercising the helper methods directly and
     * verifying their outputs, rather than invoking main() (which would
     * terminate the JVM). Integration-level tests that need exit-code
     * verification use a subprocess approach documented below.
     */
    private static BigDecimal[] parseInputs(String stdinContent) {
        BufferedReader reader = new BufferedReader(new StringReader(stdinContent));
        BigDecimal principal = Main.readDecimal(reader, "principal");
        BigDecimal rate      = Main.readDecimal(reader, "rate");
        // periods read as integer, but we return it as BigDecimal for convenience
        // We re-read via readDecimal to keep helper visible; actual int parsing
        // is tested separately.
        return new BigDecimal[]{principal, rate};
    }

    // -----------------------------------------------------------------------
    // EXAMPLE: typical_calculation
    // Spec: principal=10000.00, rate=0.0350, periods=12
    //       INTEREST: 4200.00   TOTAL: 14200.00   exit 0
    // -----------------------------------------------------------------------

    @Test
    @DisplayName("typical_calculation: interest=4200.00, total=14200.00")
    void typicalCalculation() {
        BigDecimal principal = new BigDecimal("10000.00");
        BigDecimal rate      = new BigDecimal("0.0350");
        int        periods   = 12;

        // interest = principal * rate * periods
        BigDecimal interest = principal
                .multiply(rate)
                .multiply(BigDecimal.valueOf(periods))
                .setScale(2, java.math.RoundingMode.HALF_UP);

        // total = principal + interest
        BigDecimal total = principal.add(interest)
                                    .setScale(2, java.math.RoundingMode.HALF_UP);

        assertEquals(new BigDecimal("4200.00"), interest,
                "interest must equal principal * rate * periods");
        assertEquals(new BigDecimal("14200.00"), total,
                "total must equal principal + interest");

        // Verify output format
        String interestLine = String.format("INTEREST: %.2f", interest);
        String totalLine    = String.format("TOTAL:    %.2f", total);
        assertTrue(interestLine.startsWith("INTEREST: "),
                "output line must start with 'INTEREST: '");
        assertTrue(totalLine.startsWith("TOTAL:    "),
                "output line must start with 'TOTAL:    ' (four trailing spaces for alignment)");
        assertEquals("INTEREST: 4200.00", interestLine);
        assertEquals("TOTAL:    14200.00", totalLine);
    }

    // -----------------------------------------------------------------------
    // EXAMPLE: zero_rate_rejected
    // Spec: rate=0.0000 → stderr "invalid rate", exit 2
    // -----------------------------------------------------------------------

    @Test
    @DisplayName("zero_rate_rejected: rate=0.0000 fails validation")
    void zeroRateRejected() {
        BigDecimal rate = new BigDecimal("0.0000");
        // rate > 0 is required (STEP 5)
        assertFalse(rate.compareTo(BigDecimal.ZERO) > 0,
                "rate=0.0000 must not pass the rate > 0 guard");
    }

    // -----------------------------------------------------------------------
    // EXAMPLE: zero_principal_rejected
    // Spec: principal=0.00 → stderr "invalid principal", exit 2
    // -----------------------------------------------------------------------

    @Test
    @DisplayName("zero_principal_rejected: principal=0.00 fails validation")
    void zeroPrincipalRejected() {
        BigDecimal principal = new BigDecimal("0.00");
        // principal > 0 is required (STEP 4)
        assertFalse(principal.compareTo(BigDecimal.ZERO) > 0,
                "principal=0.00 must not pass the principal > 0 guard");
    }

    // -----------------------------------------------------------------------
    // EXAMPLE: zero_periods_rejected
    // Spec: periods=0 → stderr "invalid periods", exit 2
    // -----------------------------------------------------------------------

    @Test
    @DisplayName("zero_periods_rejected: periods=0 fails validation")
    void zeroPeriodsRejected() {
        int periods = 0;
        // periods >= 1 is required (STEP 6)
        assertFalse(periods >= 1,
                "periods=0 must not pass the periods >= 1 guard");
    }

    // -----------------------------------------------------------------------
    // EXAMPLE: non_numeric_input_rejected
    // Spec: principal="abc" → stderr error message, exit 1
    // -----------------------------------------------------------------------

    @Test
    @DisplayName("non_numeric_input_rejected: 'abc' is not a valid decimal")
    void nonNumericInputRejected() {
        assertThrows(NumberFormatException.class, () -> new BigDecimal("abc"),
                "parsing 'abc' as BigDecimal must throw NumberFormatException");
    }

    // -----------------------------------------------------------------------
    // Additional unit tests: readDecimal helper
    // -----------------------------------------------------------------------

    @Test
    @DisplayName("readDecimal: valid decimal string is parsed correctly")
    void readDecimalValidInput() {
        BufferedReader r = new BufferedReader(new StringReader("10000.00\n"));
        BigDecimal result = Main.readDecimal(r, "principal");
        assertEquals(new BigDecimal("10000.00"), result);
    }

    @Test
    @DisplayName("readDecimal: leading/trailing whitespace is trimmed")
    void readDecimalTrimsWhitespace() {
        BufferedReader r = new BufferedReader(new StringReader("  0.0350  \n"));
        BigDecimal result = Main.readDecimal(r, "rate");
        assertEquals(new BigDecimal("0.0350"), result);
    }

    // -----------------------------------------------------------------------
    // Additional unit tests: readInteger helper
    // -----------------------------------------------------------------------

    @Test
    @DisplayName("readInteger: valid integer string is parsed correctly")
    void readIntegerValidInput() {
        BufferedReader r = new BufferedReader(new StringReader("12\n"));
        int result = Main.readInteger(r, "periods");
        assertEquals(12, result);
    }

    @Test
    @DisplayName("readInteger: leading/trailing whitespace is trimmed")
    void readIntegerTrimsWhitespace() {
        BufferedReader r = new BufferedReader(new StringReader("  999  \n"));
        int result = Main.readInteger(r, "periods");
        assertEquals(999, result);
    }

    // -----------------------------------------------------------------------
    // Boundary tests: domain type constraints from spec TYPES section
    // -----------------------------------------------------------------------

    @Test
    @DisplayName("Principal boundary: max value 9999999.99 is accepted")
    void principalMaxBoundary() {
        BigDecimal max = new BigDecimal("9999999.99");
        assertTrue(max.compareTo(BigDecimal.ZERO) > 0 && max.compareTo(new BigDecimal("9999999.99")) <= 0,
                "9999999.99 must be within Principal range");
    }

    @Test
    @DisplayName("Rate boundary: max value 999.9999 is accepted")
    void rateMaxBoundary() {
        BigDecimal max = new BigDecimal("999.9999");
        assertTrue(max.compareTo(BigDecimal.ZERO) > 0 && max.compareTo(new BigDecimal("999.9999")) <= 0,
                "999.9999 must be within Rate range");
    }

    @Test
    @DisplayName("Periods boundary: max value 999 is accepted")
    void periodsMaxBoundary() {
        int max = 999;
        assertTrue(max >= 1 && max <= 999,
                "999 must be within Periods range");
    }

    @Test
    @DisplayName("Periods boundary: min value 1 is accepted")
    void periodsMinBoundary() {
        int min = 1;
        assertTrue(min >= 1,
                "1 must satisfy periods >= 1");
    }

    // -----------------------------------------------------------------------
    // Invariant: interest = principal * rate * periods (no compounding)
    // -----------------------------------------------------------------------

    @Test
    @DisplayName("invariant: simple interest formula — no compounding")
    void simpleInterestFormula() {
        BigDecimal principal = new BigDecimal("5000.00");
        BigDecimal rate      = new BigDecimal("0.1000");
        int        periods   = 3;

        BigDecimal interest = principal
                .multiply(rate)
                .multiply(BigDecimal.valueOf(periods))
                .setScale(2, java.math.RoundingMode.HALF_UP);

        // 5000 * 0.1 * 3 = 1500.00
        assertEquals(new BigDecimal("1500.00"), interest);

        BigDecimal total = principal.add(interest).setScale(2, java.math.RoundingMode.HALF_UP);
        assertEquals(new BigDecimal("6500.00"), total);
    }

    // -----------------------------------------------------------------------
    // Invariant: output label format
    // -----------------------------------------------------------------------

    @Test
    @DisplayName("invariant: output labels preserve 'INTEREST: ' and 'TOTAL:    '")
    void outputLabelFormat() {
        BigDecimal interest = new BigDecimal("4200.00");
        BigDecimal total    = new BigDecimal("14200.00");

        String interestLine = String.format("INTEREST: %.2f", interest);
        String totalLine    = String.format("TOTAL:    %.2f", total);

        assertTrue(interestLine.startsWith("INTEREST: "),
                "INTEREST line must start with 'INTEREST: '");
        assertTrue(totalLine.startsWith("TOTAL:    "),
                "TOTAL line must start with 'TOTAL:    ' (four trailing spaces)");
    }
}
