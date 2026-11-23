package calculator

import (
	"math"
	"math/big"
	"strconv"
	"strings"
)

// The expression lexer has already validated the decimal syntax.
func parseNumber(text string) (*big.Rat, *Problem) {
	approx, err := strconv.ParseFloat(text, 64)
	if err != nil {
		return nil, RangeProblem
	}
	if approx == 0 {
		// 0e-400 is zero; 1e-400 is a nonzero number that underflowed.
		coefficient, _, _ := strings.Cut(strings.ToLower(text), "e")
		if strings.ContainsAny(coefficient, "123456789") {
			return nil, RangeProblem
		}
		return new(big.Rat), nil
	}
	value, ok := new(big.Rat).SetString(text)
	if !ok {
		return nil, RangeProblem
	}
	return value, nil
}

func inRange(value *big.Rat) bool {
	n, _ := value.Float64()
	return !math.IsInf(n, 0) && (n != 0 || value.Sign() == 0)
}

// Round the exact fraction to 15 significant decimal digits, with ties to even.
func roundedResult(value *big.Rat) (float64, *Problem) {
	if value.Sign() == 0 {
		return 0, nil
	}
	if !inRange(value) {
		return 0, RangeProblem
	}

	// Shift the decimal point until the integer part has 15 digits.
	numerator := new(big.Int).Abs(value.Num())
	denominator := new(big.Int).Set(value.Denom())
	const digits = 15
	shift := digits - 1 + len(denominator.String()) - len(numerator.String())
	ten := big.NewInt(10)
	factor := new(big.Int).Exp(ten, big.NewInt(int64(max(shift, -shift))), nil)
	if shift >= 0 {
		numerator.Mul(numerator, factor)
	} else {
		denominator.Mul(denominator, factor)
	}
	// Counts alone can be off by one digit (for example, 1/3 versus 3/1).
	if numerator.Cmp(new(big.Int).Mul(denominator, big.NewInt(1e14))) < 0 {
		numerator.Mul(numerator, ten)
		shift++
	}

	remainder := new(big.Int)
	rounded, _ := new(big.Int).QuoRem(numerator, denominator, remainder)
	half := remainder.Lsh(remainder, 1).Cmp(denominator)
	if half > 0 || (half == 0 && rounded.Bit(0) == 1) {
		rounded.Add(rounded, big.NewInt(1))
	}
	if value.Sign() < 0 {
		rounded.Neg(rounded)
	}

	result, err := strconv.ParseFloat(rounded.String()+"e"+strconv.Itoa(-shift), 64)
	if err != nil {
		return 0, RangeProblem
	}
	return result, nil
}
