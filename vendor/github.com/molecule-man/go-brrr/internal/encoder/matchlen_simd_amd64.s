// SSE2 byte-prefix match length for amd64.

//go:build amd64 && !purego

#include "textflag.h"

// func matchLenSIMD(dataPtr unsafe.Pointer, a, b uint, limit int) int
//
// Returns the number of bytes common to *(dataPtr+a) and *(dataPtr+b),
// up to limit. Both ranges must contain at least limit valid bytes.
//
// Pointer-based signature (rather than the larger []byte slice) cuts two
// stack stores at every caller in hash5.go's hot match-finding loops.
TEXT ·matchLenSIMD(SB), NOSPLIT|NOFRAME, $0-40
	MOVQ dataPtr+0(FP), AX
	MOVQ a+8(FP), DI
	MOVQ b+16(FP), SI
	MOVQ limit+24(FP), R8

	LEAQ (AX)(DI*1), DI       // DI = dataPtr+a
	LEAQ (AX)(SI*1), SI       // SI = dataPtr+b

	XORQ R9, R9               // i = 0
	MOVQ R8, R10
	SUBQ $16, R10             // R10 = limit - 16
	JL   tail8_check          // if limit < 16, skip 16-byte loop

simd_loop:
	MOVOU    (DI)(R9*1), X0
	MOVOU    (SI)(R9*1), X1
	PCMPEQB  X1, X0
	PMOVMSKB X0, R11
	XORQ     $0xFFFF, R11     // 1 bit per mismatching byte
	JNZ      simd_mismatch
	ADDQ     $16, R9
	CMPQ     R9, R10
	JLE      simd_loop

tail8_check:
	MOVQ R8, R10
	SUBQ $8, R10              // R10 = limit - 8
	CMPQ R9, R10
	JG   tail_byte            // if i > limit-8, skip 8-byte step

	MOVQ (DI)(R9*1), R11
	MOVQ (SI)(R9*1), R12
	XORQ R12, R11
	JNZ  tail8_mismatch
	ADDQ $8, R9

tail_byte:
	CMPQ R9, R8
	JGE  done
tail_byte_loop:
	MOVB (DI)(R9*1), R11
	MOVB (SI)(R9*1), R12
	CMPB R11, R12
	JNE  done
	INCQ R9
	CMPQ R9, R8
	JL   tail_byte_loop
	JMP  done

simd_mismatch:
	BSFQ R11, R11             // first mismatching byte (0..15)
	ADDQ R11, R9
	JMP  done

tail8_mismatch:
	BSFQ R11, R11             // first mismatching bit (0..63)
	SHRQ $3, R11              // /8 -> byte offset (0..7)
	ADDQ R11, R9

done:
	MOVQ R9, ret+32(FP)
	RET
