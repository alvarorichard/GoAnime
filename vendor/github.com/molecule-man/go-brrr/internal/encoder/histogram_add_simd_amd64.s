//go:build amd64 && !purego

#include "textflag.h"

// func histogramAddAsm(dst, src []uint32, n int)
// Thirty-two uint32 per iteration. Deeper unrolling than the alphabet sizes
// usually justify, included to show where the returns stop.
TEXT ·histogramAddAsm(SB), NOSPLIT|NOFRAME, $0-56
	MOVQ dst_base+0(FP), DI
	MOVQ src_base+24(FP), SI
	MOVQ n+48(FP), CX
	XORQ AX, AX
	MOVQ CX, DX
	SUBQ $32, DX

loop32:
	CMPQ AX, DX
	JG   tail32
	MOVOU (DI)(AX*4), X0
	MOVOU 16(DI)(AX*4), X2
	MOVOU 32(DI)(AX*4), X4
	MOVOU 48(DI)(AX*4), X6
	MOVOU (SI)(AX*4), X1
	MOVOU 16(SI)(AX*4), X3
	MOVOU 32(SI)(AX*4), X5
	MOVOU 48(SI)(AX*4), X7
	PADDL X1, X0
	PADDL X3, X2
	PADDL X5, X4
	PADDL X7, X6
	MOVOU X0, (DI)(AX*4)
	MOVOU X2, 16(DI)(AX*4)
	MOVOU X4, 32(DI)(AX*4)
	MOVOU X6, 48(DI)(AX*4)
	MOVOU 64(DI)(AX*4), X8
	MOVOU 80(DI)(AX*4), X10
	MOVOU 96(DI)(AX*4), X12
	MOVOU 112(DI)(AX*4), X14
	MOVOU 64(SI)(AX*4), X9
	MOVOU 80(SI)(AX*4), X11
	MOVOU 96(SI)(AX*4), X13
	MOVOU 112(SI)(AX*4), X15
	PADDL X9, X8
	PADDL X11, X10
	PADDL X13, X12
	PADDL X15, X14
	MOVOU X8, 64(DI)(AX*4)
	MOVOU X10, 80(DI)(AX*4)
	MOVOU X12, 96(DI)(AX*4)
	MOVOU X14, 112(DI)(AX*4)
	ADDQ  $32, AX
	JMP   loop32

tail32:
	MOVQ CX, DX
	SUBQ $4, DX

tail32_4:
	CMPQ AX, DX
	JG   tail32_1
	MOVOU (DI)(AX*4), X0
	MOVOU (SI)(AX*4), X1
	PADDL X1, X0
	MOVOU X0, (DI)(AX*4)
	ADDQ  $4, AX
	JMP   tail32_4

tail32_1:
	CMPQ AX, CX
	JGE  done32
	MOVL (SI)(AX*4), R8
	ADDL R8, (DI)(AX*4)
	INCQ AX
	JMP  tail32_1

done32:
	RET
