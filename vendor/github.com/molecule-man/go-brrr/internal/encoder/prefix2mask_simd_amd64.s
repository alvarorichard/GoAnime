//go:build amd64 && !purego

#include "textflag.h"

// func prefix2Mask64(p *byte, c0, c1 byte) uint64
// Bit j is set when p[j] == c0 && p[j+1] == c1, for j in [0,64).
// Reads 65 bytes from p. SSE2 only, so the byte broadcast goes through
// PUNPCKLBW/PSHUFLW/PSHUFD rather than PSHUFB.
TEXT ·prefix2Mask64(SB), NOSPLIT|NOFRAME, $0-24
	MOVQ    p+0(FP), SI
	MOVBLZX c0+8(FP), AX
	MOVBLZX c1+9(FP), BX

	MOVD      AX, X0
	PUNPCKLBW X0, X0
	PSHUFLW   $0, X0, X0
	PSHUFD    $0, X0, X0

	MOVD      BX, X1
	PUNPCKLBW X1, X1
	PSHUFLW   $0, X1, X1
	PSHUFD    $0, X1, X1

	MOVOU   0(SI), X2
	MOVOU   1(SI), X3
	PCMPEQB X0, X2
	PCMPEQB X1, X3
	PAND    X3, X2
	PMOVMSKB X2, R9

	MOVOU   16(SI), X4
	MOVOU   17(SI), X5
	PCMPEQB X0, X4
	PCMPEQB X1, X5
	PAND    X5, X4
	PMOVMSKB X4, R10
	SHLQ    $16, R10
	ORQ     R10, R9

	MOVOU   32(SI), X6
	MOVOU   33(SI), X7
	PCMPEQB X0, X6
	PCMPEQB X1, X7
	PAND    X7, X6
	PMOVMSKB X6, R11
	SHLQ    $32, R11
	ORQ     R11, R9

	MOVOU   48(SI), X8
	MOVOU   49(SI), X9
	PCMPEQB X0, X8
	PCMPEQB X1, X9
	PAND    X9, X8
	PMOVMSKB X8, R12
	SHLQ    $48, R12
	ORQ     R12, R9

	MOVQ R9, ret+16(FP)
	RET
