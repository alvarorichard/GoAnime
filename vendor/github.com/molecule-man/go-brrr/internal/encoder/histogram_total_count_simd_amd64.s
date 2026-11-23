// Assembly for histogramTotalCount.go. Go assembly cannot live inside a .go file.

//go:build amd64 && !purego

#include "textflag.h"

// func histogramTotalCountAsm(h []uint32, n int) uint32
// Four independent accumulators so the sum is not serialised on one PADDD.
TEXT ·histogramTotalCountAsm(SB), NOSPLIT|NOFRAME, $0-36
	MOVQ h_base+0(FP), SI
	MOVQ n+24(FP), CX
	XORQ AX, AX
	PXOR X0, X0
	PXOR X2, X2
	PXOR X4, X4
	PXOR X6, X6
	MOVQ CX, DX
	SUBQ $16, DX

sum16:
	CMPQ  AX, DX
	JG    sum16fold
	MOVOU (SI)(AX*4), X1
	MOVOU 16(SI)(AX*4), X3
	MOVOU 32(SI)(AX*4), X5
	MOVOU 48(SI)(AX*4), X7
	PADDL X1, X0
	PADDL X3, X2
	PADDL X5, X4
	PADDL X7, X6
	ADDQ  $16, AX
	JMP   sum16

sum16fold:
	PADDL  X2, X0
	PADDL  X6, X4
	PADDL  X4, X0
	MOVQ   CX, DX
	SUBQ   $4, DX

sum16_4:
	CMPQ  AX, DX
	JG    sum16_done
	MOVOU (SI)(AX*4), X1
	PADDL X1, X0
	ADDQ  $4, AX
	JMP   sum16_4

sum16_done:
	PSHUFD $0x0E, X0, X1
	PADDL  X1, X0
	PSHUFD $0x01, X0, X1
	PADDL  X1, X0
	MOVL   X0, R8

sum16_scalar:
	CMPQ AX, CX
	JGE  sum16_ret
	MOVL (SI)(AX*4), R9
	ADDL R9, R8
	INCQ AX
	JMP  sum16_scalar

sum16_ret:
	MOVL R8, ret+32(FP)
	RET
