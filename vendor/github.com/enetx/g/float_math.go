package g

import "math"

// IsNaN reports whether the Float is an IEEE 754 "not-a-number" value.
func (f Float) IsNaN() bool { return math.IsNaN(f.Std()) }

// IsInf reports whether the Float is an infinity, either positive or negative.
func (f Float) IsInf() bool { return math.IsInf(f.Std(), 0) }

// IsFinite reports whether the Float is neither NaN nor an infinity.
func (f Float) IsFinite() bool { return !f.IsNaN() && !f.IsInf() }

// IsNormal reports whether the Float is a normal IEEE 754 number:
// neither zero, subnormal, infinite, nor NaN.
func (f Float) IsNormal() bool {
	exp := f.Bits() >> 52 & 0x7ff

	return exp != 0 && exp != 0x7ff
}

// Signum returns a Float representing the sign of the Float:
// 1 if the sign bit is clear (including +0), -1 if the sign bit is set (including -0),
// and NaN if the Float is NaN.
func (f Float) Signum() Float {
	if f.IsNaN() {
		return f
	}

	return Float(math.Copysign(1, f.Std()))
}

// IsSignPositive reports whether the Float has a positive sign bit.
// This includes +0.0 and positive infinity.
// Note: NaN carries a sign bit too, so a NaN with a clear sign bit (e.g. math.NaN())
// is reported as sign-positive; use IsNaN to detect NaN itself.
func (f Float) IsSignPositive() bool { return !math.Signbit(f.Std()) }

// IsSignNegative reports whether the Float has a negative sign bit.
// This includes -0.0 and negative infinity.
// Note: NaN carries a sign bit too, so a NaN with a set sign bit (e.g.
// math.Copysign(math.NaN(), -1)) is reported as sign-negative; use IsNaN to
// detect NaN itself.
func (f Float) IsSignNegative() bool { return math.Signbit(f.Std()) }

// Ceil returns the least integer value greater than or equal to the Float.
func (f Float) Ceil() Float { return Float(math.Ceil(f.Std())) }

// Floor returns the greatest integer value less than or equal to the Float.
func (f Float) Floor() Float { return Float(math.Floor(f.Std())) }

// Trunc returns the integer part of the Float, rounding toward zero.
func (f Float) Trunc() Float { return Float(math.Trunc(f.Std())) }

// Fract returns the fractional part of the Float (f - f.Trunc()).
// For NaN and ±Inf the result is NaN.
func (f Float) Fract() Float { return f - f.Trunc() }

// Clamp restricts the Float to the inclusive range [min, max].
// If the Float is NaN, NaN is returned.
// The caller must ensure min <= max and that neither bound is NaN: this method
// does not panic on an invalid range — a NaN bound never
// compares true, so the corresponding check is silently skipped, and with
// min > max the lower bound wins.
func (f Float) Clamp(min, max Float) Float {
	if f < min {
		return min
	}

	if f > max {
		return max
	}

	return f
}

// Recip returns the reciprocal (multiplicative inverse) of the Float, 1/f.
func (f Float) Recip() Float { return 1 / f }

// Copysign returns a Float with the magnitude of the Float and the sign of sign.
func (f Float) Copysign(sign Float) Float { return Float(math.Copysign(f.Std(), sign.Std())) }

// MulAdd returns f*b + c computed as a fused multiply-add with only one rounding.
func (f Float) MulAdd(b, c Float) Float { return Float(math.FMA(f.Std(), b.Std(), c.Std())) }

// Hypot returns Sqrt(f*f + b*b), avoiding unnecessary overflow and underflow.
func (f Float) Hypot(b Float) Float { return Float(math.Hypot(f.Std(), b.Std())) }

// Cbrt returns the cube root of the Float.
func (f Float) Cbrt() Float { return Float(math.Cbrt(f.Std())) }

// Exp returns e**f, the base-e exponential of the Float.
func (f Float) Exp() Float { return Float(math.Exp(f.Std())) }

// Exp2 returns 2**f, the base-2 exponential of the Float.
func (f Float) Exp2() Float { return Float(math.Exp2(f.Std())) }

// ExpM1 returns e**f - 1, which is more accurate than Exp().Sub(1) when the Float is near zero.
func (f Float) ExpM1() Float { return Float(math.Expm1(f.Std())) }

// Ln returns the natural logarithm of the Float.
func (f Float) Ln() Float { return Float(math.Log(f.Std())) }

// Ln1p returns the natural logarithm of 1 plus the Float,
// which is more accurate than Add(1).Ln() when the Float is near zero.
func (f Float) Ln1p() Float { return Float(math.Log1p(f.Std())) }

// Log2 returns the base-2 logarithm of the Float.
func (f Float) Log2() Float { return Float(math.Log2(f.Std())) }

// Log10 returns the base-10 logarithm of the Float.
func (f Float) Log10() Float { return Float(math.Log10(f.Std())) }

// Sin returns the sine of the Float (in radians).
func (f Float) Sin() Float { return Float(math.Sin(f.Std())) }

// Cos returns the cosine of the Float (in radians).
func (f Float) Cos() Float { return Float(math.Cos(f.Std())) }

// Tan returns the tangent of the Float (in radians).
func (f Float) Tan() Float { return Float(math.Tan(f.Std())) }

// Asin returns the arcsine of the Float, in radians.
func (f Float) Asin() Float { return Float(math.Asin(f.Std())) }

// Acos returns the arccosine of the Float, in radians.
func (f Float) Acos() Float { return Float(math.Acos(f.Std())) }

// Atan returns the arctangent of the Float, in radians.
func (f Float) Atan() Float { return Float(math.Atan(f.Std())) }

// Atan2 returns the arctangent of f/b (with f as y and b as x), in radians,
// using the signs of the two to determine the quadrant of the result.
func (f Float) Atan2(b Float) Float { return Float(math.Atan2(f.Std(), b.Std())) }

// Sinh returns the hyperbolic sine of the Float.
func (f Float) Sinh() Float { return Float(math.Sinh(f.Std())) }

// Cosh returns the hyperbolic cosine of the Float.
func (f Float) Cosh() Float { return Float(math.Cosh(f.Std())) }

// Tanh returns the hyperbolic tangent of the Float.
func (f Float) Tanh() Float { return Float(math.Tanh(f.Std())) }

// Asinh returns the inverse hyperbolic sine of the Float.
func (f Float) Asinh() Float { return Float(math.Asinh(f.Std())) }

// Acosh returns the inverse hyperbolic cosine of the Float.
func (f Float) Acosh() Float { return Float(math.Acosh(f.Std())) }

// Atanh returns the inverse hyperbolic tangent of the Float.
func (f Float) Atanh() Float { return Float(math.Atanh(f.Std())) }

// ToDegrees converts the Float from radians to degrees.
func (f Float) ToDegrees() Float { return f * (180 / math.Pi) }

// ToRadians converts the Float from degrees to radians.
func (f Float) ToRadians() Float { return f * (math.Pi / 180) }
