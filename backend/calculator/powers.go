package calculator

import "math/big"

const (
	// Cap new numerator and denominator bits produced by exact powers across
	// a request. Subsequent ordinary arithmetic still operates on rationals.
	maxExactPowerBits = 64 << 10
	approximationBits = 192
)

type powerBudget int

func squareRoot(value *big.Rat) (*big.Rat, *Problem) {
	if value.Sign() < 0 {
		return nil, &Problem{422, "domain_error", "Square root requires a nonnegative value."}
	}
	n := new(big.Int).Sqrt(value.Num())
	d := new(big.Int).Sqrt(value.Denom())
	square := new(big.Int)
	if square.Mul(n, n).Cmp(value.Num()) == 0 && square.Mul(d, d).Cmp(value.Denom()) == 0 {
		return new(big.Rat).SetFrac(n, d), nil
	}
	root := new(big.Float).SetPrec(approximationBits).SetRat(value)
	result, _ := root.Sqrt(root).Rat(nil)
	return result, nil
}

func (budget *powerBudget) pow(base, exponent *big.Rat) (*big.Rat, *Problem) {
	if exponent.Sign() == 0 {
		return big.NewRat(1, 1), nil
	}
	if base.Sign() == 0 {
		if exponent.Sign() < 0 {
			return nil, ZeroDivisor
		}
		return base, nil
	}
	if base.Sign() < 0 && !exponent.IsInt() {
		return nil, &Problem{422, "domain_error", "A negative base requires an integer exponent."}
	}
	negative := base.Sign() < 0 && exponent.Num().Bit(0) == 1
	base.Abs(base)
	if exponent.Sign() < 0 {
		base.Inv(base)
		exponent.Neg(exponent)
	}

	switch {
	case base.Cmp(big.NewRat(1, 1)) == 0:
		// Also handles (-1) raised to an integer too large for float64 parity.
	case exponent.Cmp(big.NewRat(1, 1)) == 0:
		// No additional work or loss of precision.
	case exponent.IsInt() && budget.allow(base, exponent.Num()):
		n := new(big.Int).Exp(base.Num(), exponent.Num(), nil)
		d := new(big.Int).Exp(base.Denom(), exponent.Num(), nil)
		base.SetFrac(n, d)
	default:
		var err *Problem
		base, err = approximatePower(base, exponent)
		if err != nil {
			return nil, err
		}
	}
	if negative {
		base.Neg(base)
	}
	if !inRange(base) {
		return nil, RangeProblem
	}
	return base, nil
}

func (budget *powerBudget) allow(base *big.Rat, exponent *big.Int) bool {
	bits := base.Num().BitLen() + base.Denom().BitLen()
	if !exponent.IsInt64() || exponent.Int64() > int64(int(*budget)/bits) {
		return false
	}
	*budget -= powerBudget(bits * int(exponent.Int64()))
	return true
}

// For a positive base and exponent, use binary exponentiation. Extra precision
// proportional to the exponent's bit length preserves differences from one
// that a large exponent amplifies. Inputs and intermediate storage stay bounded.
func approximatePower(base, exponent *big.Rat) (*big.Rat, *Problem) {
	remainder := new(big.Int)
	integer, _ := new(big.Int).QuoRem(exponent.Num(), exponent.Denom(), remainder)
	precision := uint(approximationBits + integer.BitLen())
	factor := new(big.Float).SetPrec(precision).SetRat(base)
	result := new(big.Float).SetPrec(precision).SetInt64(1)
	for i := integer.BitLen() - 1; i >= 0; i-- {
		result.Mul(result, result)
		if integer.Bit(i) != 0 {
			result.Mul(result, factor)
		}
		// Remaining factors move the magnitude in the same direction. Stop
		// early, before squaring can overflow even a big.Float exponent.
		if e := result.MantExp(nil); e > 1100 || e < -1100 {
			return nil, RangeProblem
		}
	}

	// The bits after the binary point select b^(1/2), b^(1/4), etc.
	// Omitting the tail changes log2(result) by at most 1075 * 2^-192.
	if remainder.Sign() != 0 {
		fraction := remainder.Lsh(remainder, approximationBits)
		fraction.Quo(fraction, exponent.Denom())
		lastBit := int(fraction.TrailingZeroBits())
		for i := approximationBits - 1; fraction.Sign() != 0 && i >= lastBit; i-- {
			factor.Sqrt(factor)
			if fraction.Bit(i) != 0 {
				result.Mul(result, factor)
			}
		}
	}
	value, _ := result.Rat(nil)
	return value, nil
}
