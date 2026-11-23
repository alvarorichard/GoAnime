// Ported from the FMA path of log2 in ARM Optimized Routines
// (math/log2.c, math/log2_data.c):
//
//	Copyright (c) 2018-2025, Arm Limited.
//	SPDX-License-Identifier: MIT OR Apache-2.0 WITH LLVM-exception
//
// glibc builds its log2 from the same code. Its result affects q10/q11 output.

package encoder

import "math"

const (
	log2InvLn2Hi = 0x1.7154765200000p+0
	log2InvLn2Lo = 0x1.705fc2eefa200p-33
	log2A0       = -0x1.71547652b8339p-1
	log2A1       = 0x1.ec709dc3a04bep-2
	log2A2       = -0x1.7154764702ffbp-2
	log2A3       = 0x1.2776c50034c48p-2
	log2A4       = -0x1.ec7b328ea92bcp-3
	log2A5       = 0x1.a6225e117f92ep-3
	log2Off      = 0x3fe6000000000000
)

// log2Tab holds 1/c and log2(c) for the center c of each subinterval.
var log2Tab = [64]struct{ invc, logc float64 }{
	{0x1.724286bb1acf8p+0, -0x1.1095feecdb000p-1},
	{0x1.6e1f766d2cca1p+0, -0x1.08494bd76d000p-1},
	{0x1.6a13d0e30d48ap+0, -0x1.00143aee8f800p-1},
	{0x1.661ec32d06c85p+0, -0x1.efec5360b4000p-2},
	{0x1.623fa951198f8p+0, -0x1.dfdd91ab7e000p-2},
	{0x1.5e75ba4cf026cp+0, -0x1.cffae0cc79000p-2},
	{0x1.5ac055a214fb8p+0, -0x1.c043811fda000p-2},
	{0x1.571ed0f166e1ep+0, -0x1.b0b67323ae000p-2},
	{0x1.53909590bf835p+0, -0x1.a152f5a2db000p-2},
	{0x1.5014fed61adddp+0, -0x1.9217f5af86000p-2},
	{0x1.4cab88e487bd0p+0, -0x1.8304db0719000p-2},
	{0x1.49539b4334feep+0, -0x1.74189f9a9e000p-2},
	{0x1.460cbdfafd569p+0, -0x1.6552bb5199000p-2},
	{0x1.42d664ee4b953p+0, -0x1.56b23a29b1000p-2},
	{0x1.3fb01111dd8a6p+0, -0x1.483650f5fa000p-2},
	{0x1.3c995b70c5836p+0, -0x1.39de937f6a000p-2},
	{0x1.3991c4ab6fd4ap+0, -0x1.2baa1538d6000p-2},
	{0x1.3698e0ce099b5p+0, -0x1.1d98340ca4000p-2},
	{0x1.33ae48213e7b2p+0, -0x1.0fa853a40e000p-2},
	{0x1.30d191985bdb1p+0, -0x1.01d9c32e73000p-2},
	{0x1.2e025cab271d7p+0, -0x1.e857da2fa6000p-3},
	{0x1.2b404cf13cd82p+0, -0x1.cd3c8633d8000p-3},
	{0x1.288b02c7ccb50p+0, -0x1.b26034c14a000p-3},
	{0x1.25e2263944de5p+0, -0x1.97c1c2f4fe000p-3},
	{0x1.234563d8615b1p+0, -0x1.7d6023f800000p-3},
	{0x1.20b46e33eaf38p+0, -0x1.633a71a05e000p-3},
	{0x1.1e2eefdcda3ddp+0, -0x1.494f5e9570000p-3},
	{0x1.1bb4a580b3930p+0, -0x1.2f9e424e0a000p-3},
	{0x1.19453847f2200p+0, -0x1.162595afdc000p-3},
	{0x1.16e06c0d5d73cp+0, -0x1.f9c9a75bd8000p-4},
	{0x1.1485f47b7e4c2p+0, -0x1.c7b575bf9c000p-4},
	{0x1.12358ad0085d1p+0, -0x1.960c60ff48000p-4},
	{0x1.0fef00f532227p+0, -0x1.64ce247b60000p-4},
	{0x1.0db2077d03a8fp+0, -0x1.33f78b2014000p-4},
	{0x1.0b7e6d65980d9p+0, -0x1.0387d1a42c000p-4},
	{0x1.0953efe7b408dp+0, -0x1.a6f9208b50000p-5},
	{0x1.07325cac53b83p+0, -0x1.47a954f770000p-5},
	{0x1.05197e40d1b5cp+0, -0x1.d23a8c50c0000p-6},
	{0x1.03091c1208ea2p+0, -0x1.16a2629780000p-6},
	{0x1.0101025b37e21p+0, -0x1.720f8d8e80000p-8},
	{0x1.fc07ef9caa76bp-1, 0x1.6fe53b1500000p-7},
	{0x1.f4465d3f6f184p-1, 0x1.11ccce10f8000p-5},
	{0x1.ecc079f84107fp-1, 0x1.c4dfc8c8b8000p-5},
	{0x1.e573a99975ae8p-1, 0x1.3aa321e574000p-4},
	{0x1.de5d6f0bd3de6p-1, 0x1.918a0d08b8000p-4},
	{0x1.d77b681ff38b3p-1, 0x1.e72e9da044000p-4},
	{0x1.d0cb5724de943p-1, 0x1.1dcd2507f6000p-3},
	{0x1.ca4b2dc0e7563p-1, 0x1.476ab03dea000p-3},
	{0x1.c3f8ee8d6cb51p-1, 0x1.7074377e22000p-3},
	{0x1.bdd2b4f020c4cp-1, 0x1.98ede8ba94000p-3},
	{0x1.b7d6c006015cap-1, 0x1.c0db86ad2e000p-3},
	{0x1.b20366e2e338fp-1, 0x1.e840aafcee000p-3},
	{0x1.ac57026295039p-1, 0x1.0790ab4678000p-2},
	{0x1.a6d01bc2731ddp-1, 0x1.1ac056801c000p-2},
	{0x1.a16d3bc3ff18bp-1, 0x1.2db11d4fee000p-2},
	{0x1.9c2d14967feadp-1, 0x1.406464ec58000p-2},
	{0x1.970e4f47c9902p-1, 0x1.52dbe093af000p-2},
	{0x1.920fb3982bcf2p-1, 0x1.651902050d000p-2},
	{0x1.8d30187f759f1p-1, 0x1.771d2cdeaf000p-2},
	{0x1.886e5ebb9f66dp-1, 0x1.88e9c857d9000p-2},
	{0x1.83c97b658b994p-1, 0x1.9a80155e16000p-2},
	{0x1.7f405ffc61022p-1, 0x1.abe186ed3d000p-2},
	{0x1.7ad22181415cap-1, 0x1.bd0f2aea0e000p-2},
	{0x1.767dcf99eff8cp-1, 0x1.ce0a43dbf4000p-2},
}

// log2Large returns log2(v) for v >= 256, bit-identical to x86_64 glibc log2
// on FMA hardware. It skips the special paths for x near 1, subnormals, inf
// and NaN.
func log2Large(v int) float64 {
	ix := math.Float64bits(float64(v))
	// v = 2^k * z, with z in [0x1.6p-1, 0x1.6p0) inside table interval i.
	tmp := ix - log2Off
	i := (tmp >> (52 - 6)) % uint64(len(log2Tab))
	k := int64(tmp) >> 52
	z := math.Float64frombits(ix - tmp&(0xfff<<52))
	invc, logc := log2Tab[i].invc, log2Tab[i].logc

	// Operation order and fusion follow the glibc __log2_fma machine code.
	// float64() stops the compiler from fusing other operations.
	r := math.FMA(z, invc, -1.0)
	t1 := float64(r * log2InvLn2Hi)
	t2 := math.FMA(r, log2InvLn2Lo, math.FMA(r, log2InvLn2Hi, -t1))

	t3 := float64(k) + logc
	hi := t3 + t1
	lo := t3 - hi + t1 + t2

	r2 := float64(r * r)
	r4 := float64(r2 * r2)
	p := math.FMA(math.FMA(r, log2A5, log2A4), r4,
		math.FMA(math.FMA(r, log2A3, log2A2), r2, math.FMA(r, log2A1, log2A0)))
	return math.FMA(r2, p, lo) + hi
}
