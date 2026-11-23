package calculator

import (
	"fmt"
	"math"
	"strings"
	"testing"
)

// Expression tests
func TestExpressions(t *testing.T) {
	// Every pair of operators follows the expected order of operations.
	t.Run("AllOperatorPairs", func(t *testing.T) {
		for _, tc := range []struct {
			first, second string
			want          float64
		}{
			{"+", "+", 17},
			{"+", "-", 13},
			{"+", "*", 18},
			{"+", "/", 13.5},
			{"-", "+", 11},
			{"-", "-", 7},
			{"-", "*", 6},
			{"-", "/", 10.5},
			{"*", "+", 38},
			{"*", "-", 34},
			{"*", "*", 72},
			{"*", "/", 18},
			{"/", "+", 6},
			{"/", "-", 2},
			{"/", "*", 8},
			{"/", "/", 2},
		} {
			t.Run(tc.first+"/"+tc.second, func(t *testing.T) {
				expr := fmt.Sprintf("12 %s 3 %s 2", tc.first, tc.second)
				got, p := Evaluate(expr)
				if p != nil || got != tc.want {
					t.Fatalf("%q = %v, %v; want %v", expr, got, p, tc.want)
				}
			})
		}
	})

	// Valid expressions produce the expected results, including decimal rounding.
	t.Run("NumericResults", func(t *testing.T) {
		for _, tc := range []struct {
			name, expr string
			want       float64
		}{
			{"precedence", "2 + 3 * 4", 14},
			{"left associative", "24 / 4 * 2", 12},
			{"subtraction", "8 - 3 * 4 + 2", -2},
			{"decimal rounding", "0.1 + 0.2", 0.3},
			{"decimal cancellation", "0.1 + 0.2 - 0.3", 0},
			{"zero numerator", "0 / 5", 0},
			{"negative numerator", "-9 / 4", -2.25},
			{"negative divisor", "8 / -2", -4},
			{"two negatives", "-3 * -2", 6},
			{"subtract a negative", "-5 - -3", -2},
			{"zero product", "0 * -8", 0},
			{"signed scientific notation", "-1.25E+2 * 5e-1", -62.5},
			{"zero with exponent", "-0.00e-400 + 2", 2},
			{"repeating decimal", "1 / 3", 0.333333333333333},
			{"round up at fifteen significant digits", "1.234567890123456 + 0", 1.23456789012346},
			{"round down at fifteen significant digits", "1.234567890123451 + 0", 1.23456789012345},
			{"carry across decimal magnitude", "9.999999999999999 + 0", 10},
			{"just above a rounding midpoint", "1.23456789012345501 + 0", 1.23456789012346},
			{"just below an integer midpoint", "999999999999999.49 + 0", 999999999999999},
			{"exact midpoint rounds to even", "1.000000000000005 + 0", 1},
			{"negative midpoint rounds to even", "-1.000000000000015 + 0", -1.00000000000002},
			{"round only at the end", "1 / 3 * 3", 1},
			{"exact large integer difference", "9007199254740993 - 9007199254740992", 1},
			{"subnormal result", "1e-320 * 1", 1e-320},
			{"small nonzero divisor", "1e-300 / 1e-300", 1},
			{"large representable result", "1e307 * 10", 1e308},
			{"alternating terms", "100 - 8 * 3 + 12 / 4 * 2", 82},
		} {
			t.Run(tc.name, func(t *testing.T) {
				got, p := Evaluate(tc.expr)
				if p != nil || got != tc.want {
					t.Fatalf("%q = %v, %v; want %v", tc.expr, got, p, tc.want)
				}
			})
		}
	})

	// Division by zero and out-of-range numbers return the expected errors.
	t.Run("NumericErrors", func(t *testing.T) {
		for _, tc := range []struct{ name, expr, code string }{
			{"zero divisor", "2 / -0", "division_by_zero"},
			{"overflowing input", "1e400 + 2", "out_of_range"},
			{"underflowing input", "1e-400 + 2", "out_of_range"},
			{"overflowing product", "1e300 * 1e300", "out_of_range"},
			{"zero divided by zero", "0 / 0", "division_by_zero"},
			{"zero divisor in a later term", "8 + 2 / 0", "division_by_zero"},
			{"signed exponent zero divisor", "1 / -0e8", "division_by_zero"},
			{"negative input out of range", "-1e400 + 0", "out_of_range"},
			{"sum exceeds range", "1e308 + 1e308", "out_of_range"},
			{"negative result exceeds range", "-1e308 - 1e308", "out_of_range"},
			{"overflow before later cancellation", "1e308 + 1e308 - 1e308", "out_of_range"},
			{"quotient overflows", "1e308 / 0.1", "out_of_range"},
			{"product underflows", "1e-200 * 1e-200", "out_of_range"},
			{"quotient underflows", "1e-300 / 1e100", "out_of_range"},
			{"final sum underflows", "1.1e-323 - 1e-323", "out_of_range"},
			{"rounded value exceeds range", "1.7976931348623157e308 + 0", "out_of_range"},
		} {
			t.Run(tc.name, func(t *testing.T) {
				_, p := Evaluate(tc.expr)
				if p == nil || p.Status != 422 || p.Code != tc.code {
					t.Fatalf("%q returned %v, want code %s", tc.expr, p, tc.code)
				}
			})
		}
	})
}

// Parser tests
func TestParser(t *testing.T) {
	// Check grouping, precedence, contextual percentages, and numerical precision.
	t.Run("ExpressionValues", func(t *testing.T) {
		for _, tc := range []struct {
			input string
			want  float64
		}{
			{"(2 + 3)^2", 25}, {"√(9 + 7)", 4}, {"2^3^2", 512},
			{"-2^2", -4}, {"(-2)^2", 4}, {"2^(-3)", .125}, {"2^-2^2", .0625},
			{"10%", .1}, {"200 × 10%", 20}, {"200 ÷ 10%", 2000},
			{"200 + 10%", 220}, {"200 − 10%", 180}, {"200 + (10%)", 220},
			{"(100 + 50) + 10%", 165}, {"200 + (5 + 5)%", 220},
			{"200 + 10% × 2", 200.2}, {"200 + 10%^2", 200.01},
			{"100 + 15% + 15%", 132.25}, {"3 - 2 - 1", 0},
			{"(.1 + 0.2) - .3", 0}, {"1. / 3 * 3", 1},
			{"010 + .5E-1", 10.05}, {"1e+2 + 1e-1", 100.1},
			{"1.000000000000005", 1}, {"-1.000000000000015", -1.00000000000002},
			{"9007199254740993 - 9007199254740992", 1}, {"sqrt ( .81 ) - .9", 0},
			{"sqrt(2)", 1.4142135623731}, {"2^(1/3)", 1.25992104989487},
			{"sqrt(1e-320)", 1e-160}, {"(1e-320)^0.5", 1e-160}, {"0^0", 1}, {"0^.5", 0},
			{"(-1)^9007199254740993", -1}, {"(-1)^(-1e308)", 1},
			{"1.00000000000000000001^1e20", 2.71828182845905},
			{"1.00000000000000000001^-1e20", .367879441171442},
			{"(1.00000000000000000001^(1e20 + .5))", 2.71828182845905},
			{"(1.234567890123456789^2)-(1.234567890123456789*1.234567890123456789)", 0},
			{strings.Repeat("(", 32) + "1" + strings.Repeat(")", 32), 1},
			{strings.Repeat("sqrt(", 32) + "1" + strings.Repeat(")", 32), 1},
			{"1" + strings.Repeat("+1", 128), 129},
		} {
			t.Run(tc.input, func(t *testing.T) {
				got, p := Evaluate(tc.input)
				if p != nil || got != tc.want {
					t.Fatalf("got %v, %v; want %v", got, p, tc.want)
				}
			})
		}
	})

	// Reject invalid syntax and expressions beyond the parser limits.
	t.Run("MalformedExpression", func(t *testing.T) {
		for _, input := range []string{
			"", "  ", ".", "1..0", "nan", "1e", "2e+", "0x10", "3_000",
			"(1+2", "1+2)", "1+", "()", "sqrt()", "√9", "2(3)", "(2)(3)",
			"1 / 0 garbage", "1e400 +", "1 2", "4 // 2",
			strings.Repeat("(", 33) + "1" + strings.Repeat(")", 33),
			strings.Repeat("√(", 33) + "1" + strings.Repeat(")", 33),
			"1" + strings.Repeat("^1", 129), strings.Repeat("-", 129) + "0",
			"0" + strings.Repeat("%", 129),
		} {
			_, p := Evaluate(input)
			if p == nil || p.Status != 400 || p.Code != "invalid_expression" {
				t.Errorf("input %q returned %v", input, p)
			}
		}
	})

	// Report domain and range errors for syntactically valid expressions.
	t.Run("ExpressionNumericErrors", func(t *testing.T) {
		for _, tc := range []struct{ input, code string }{
			{"2 / (-0)", "division_by_zero"}, {"0^-1", "division_by_zero"},
			{"0^(-.5)", "division_by_zero"}, {"sqrt(-1)", "domain_error"},
			{"(-2)^.5", "domain_error"}, {"(-8)^(1/3)", "domain_error"},
			{"2^1e308", "out_of_range"}, {"2^-1e308", "out_of_range"},
			{"1e-400", "out_of_range"}, {"1e400", "out_of_range"},
			{"1e308 + 1e308 - 1e308", "out_of_range"},
			{"1e-200 * 1e-200", "out_of_range"}, {"1e-320 ^ 1.5", "out_of_range"},
		} {
			_, p := Evaluate(tc.input)
			if p == nil || p.Status != 422 || p.Code != tc.code {
				t.Errorf("input %q returned %v, want %s", tc.input, p, tc.code)
			}
		}
	})
}

// Generated expressions must match Go arithmetic for multiplication and subtraction.
func FuzzMixedArithmetic(f *testing.F) {
	f.Add(int16(12), int16(-3), int16(2), uint8(0))
	f.Add(int16(0), int16(0), int16(0), uint8(1))
	f.Fuzz(func(t *testing.T, a, b, c int16, order uint8) {
		x, y, z := int64(a), int64(b), int64(c)
		format, want := "%d - %d * %d", x-y*z
		if order%2 == 1 {
			format, want = "%d * %d - %d", x*y-z
		}
		input := fmt.Sprintf(format, a, b, c)
		got, p := Evaluate(input)
		if p != nil || got != float64(want) {
			t.Fatalf("%q got %v, %v; want %v", input, got, p, want)
		}
	})
}

// Arbitrary expressions must return either a finite number or a documented error status.
func FuzzExpressionParser(f *testing.F) {
	for _, seed := range []string{"√(9−5) × .5", "2^-3^2", "200+(5+5)%", "1/0!", "0^(-2)", "1e-400"} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, input string) {
		got, p := Evaluate(input)
		if p != nil {
			if p.Status != 400 && p.Status != 422 {
				t.Fatalf("unexpected error %v", p)
			}
		} else if math.IsNaN(got) || math.IsInf(got, 0) {
			t.Fatalf("nonfinite result %v", got)
		}
	})
}

// A grouped percentage adjusts the full amount on the left.
func FuzzGroupedPercent(f *testing.F) {
	f.Add(int16(200), int16(5), int16(5))
	f.Fuzz(func(t *testing.T, a, b, c int16) {
		input := fmt.Sprintf("%d + (%d + %d)%%", a, b, c)
		want := float64(int64(a)*(100+int64(b)+int64(c))) / 100
		got, p := Evaluate(input)
		if p != nil || got != want {
			t.Fatalf("%q got %v, %v; want %v", input, got, p, want)
		}
	})
}
