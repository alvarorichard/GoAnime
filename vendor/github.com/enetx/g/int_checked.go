package g

import "math"

// CheckedAdd adds two Ints, returning None if the addition overflows.
func (i Int) CheckedAdd(b Int) Option[Int] {
	sum := i + b
	if (b > 0 && sum < i) || (b < 0 && sum > i) {
		return None[Int]()
	}

	return Some(sum)
}

// CheckedSub subtracts b from the Int, returning None if the subtraction overflows.
func (i Int) CheckedSub(b Int) Option[Int] {
	diff := i - b
	if (b < 0 && diff < i) || (b > 0 && diff > i) {
		return None[Int]()
	}

	return Some(diff)
}

// CheckedMul multiplies two Ints, returning None if the multiplication overflows.
func (i Int) CheckedMul(b Int) Option[Int] {
	if i == 0 || b == 0 {
		return Some(Int(0))
	}

	if i == -1 {
		return b.CheckedNeg()
	}

	if b == -1 {
		return i.CheckedNeg()
	}

	c := i * b
	if c/b != i {
		return None[Int]()
	}

	return Some(c)
}

// CheckedDiv divides the Int by b, returning None if b is zero or the division overflows.
func (i Int) CheckedDiv(b Int) Option[Int] {
	if b == 0 || (i == math.MinInt && b == -1) {
		return None[Int]()
	}

	return Some(i / b)
}

// CheckedRem computes the remainder of the Int divided by b, returning None if b is zero
// or the operation overflows (i == MinInt and b == -1).
func (i Int) CheckedRem(b Int) Option[Int] {
	if b == 0 || (i == math.MinInt && b == -1) {
		return None[Int]()
	}

	return Some(i % b)
}

// CheckedNeg negates the Int, returning None if the negation overflows (i == MinInt).
func (i Int) CheckedNeg() Option[Int] {
	if i == math.MinInt {
		return None[Int]()
	}

	return Some(-i)
}

// CheckedAbs returns the absolute value of the Int, returning None if it overflows (i == MinInt).
func (i Int) CheckedAbs() Option[Int] {
	if i == math.MinInt {
		return None[Int]()
	}

	if i < 0 {
		return Some(-i)
	}

	return Some(i)
}

// CheckedPow raises the Int to the power of exp using exponentiation by squaring,
// returning None if exp is negative or the computation overflows. An exp of zero yields Some(1).
func (i Int) CheckedPow(exp Int) Option[Int] {
	if exp < 0 {
		return None[Int]()
	}

	result, base := Int(1), i

	for exp > 0 {
		if exp&1 == 1 {
			r := result.CheckedMul(base)
			if r.IsNone() {
				return None[Int]()
			}

			result = r.Some()
		}

		exp >>= 1

		if exp > 0 {
			b := base.CheckedMul(base)
			if b.IsNone() {
				return None[Int]()
			}

			base = b.Some()
		}
	}

	return Some(result)
}

// SaturatingAdd adds two Ints, clamping the result to MinInt or MaxInt on overflow.
func (i Int) SaturatingAdd(b Int) Int {
	sum := i + b
	if b > 0 && sum < i {
		return math.MaxInt
	}

	if b < 0 && sum > i {
		return math.MinInt
	}

	return sum
}

// SaturatingSub subtracts b from the Int, clamping the result to MinInt or MaxInt on overflow.
func (i Int) SaturatingSub(b Int) Int {
	diff := i - b
	if b < 0 && diff < i {
		return math.MaxInt
	}

	if b > 0 && diff > i {
		return math.MinInt
	}

	return diff
}

// SaturatingMul multiplies two Ints, clamping the result to MinInt or MaxInt on overflow.
func (i Int) SaturatingMul(b Int) Int {
	if c := i.CheckedMul(b); c.IsSome() {
		return c.Some()
	}

	if (i > 0) == (b > 0) {
		return math.MaxInt
	}

	return math.MinInt
}

// OverflowingAdd adds two Ints, returning the wrapped result and a flag indicating overflow.
func (i Int) OverflowingAdd(b Int) (Int, bool) {
	sum := i + b

	return sum, (b > 0 && sum < i) || (b < 0 && sum > i)
}

// OverflowingSub subtracts b from the Int, returning the wrapped result and a flag indicating overflow.
func (i Int) OverflowingSub(b Int) (Int, bool) {
	diff := i - b

	return diff, (b < 0 && diff < i) || (b > 0 && diff > i)
}

// OverflowingMul multiplies two Ints, returning the wrapped result and a flag indicating overflow.
func (i Int) OverflowingMul(b Int) (Int, bool) {
	return i * b, i.CheckedMul(b).IsNone()
}

// Clamp restricts the Int to the inclusive range [min, max].
// The caller must ensure min <= max: this method does not
// panic on an inverted range — the lower bound is checked first, so the result
// is unspecified when min > max.
func (i Int) Clamp(min, max Int) Int {
	if i < min {
		return min
	}

	if i > max {
		return max
	}

	return i
}
