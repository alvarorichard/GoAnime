//go:build amd64 && !purego

#include "go_asm.h"
#include "textflag.h"

// func h6b5FindInBucket(h *h6b5, data unsafe.Pointer, keyTag, cur, curMasked,
//	mask, minPrev, maxLength, bestLen, bestScore uint,
//	next *h6b5Block, out *hasherSearchResult)
//
// No local spills: each store consumes a store-buffer entry.
// TZCNT acts as BSF on CPUs without BMI1. Both give the same result for nonzero input.
TEXT ·h6b5FindInBucket(SB), NOSPLIT|NOFRAME, $0-96
	MOVQ       next+80(FP), BX
	PREFETCHT0 (BX)
	PREFETCHT0 96(BX)

	MOVQ    h+0(FP), DI
	MOVQ    keyTag+16(FP), CX
	MOVBLZX CL, AX
	SHRQ    $8, CX
	MOVWLZX h6b5_num(DI)(CX*2), DX
	LEAQ    (CX)(CX*4), CX
	SHLQ    $5, CX
	LEAQ    h6b5_blocks(DI)(CX*1), DI

	// Tag mask of the block, rotated so bit 0 is the newest slot.
	MOVD      AX, X0
	PUNPCKLBW X0, X0
	PSHUFLW   $0, X0, X0
	PSHUFD    $0, X0, X0
	MOVOU     (DI), X1
	MOVOU     16(DI), X2
	PCMPEQB   X0, X1
	PCMPEQB   X0, X2
	PMOVMSKB  X1, AX
	PMOVMSKB  X2, BX
	SHLL      $16, BX
	ORL       BX, AX
	LEAL      1(DX), R15
	ANDL      $31, R15
	MOVL      R15, CX
	RORL      CX, AX

	// Mask empty slots before the bucket fills.
	MOVL  $0xFFFF, CX
	SUBL  DX, CX
	CMPL  CX, $32
	JCC   allstored
	MOVL  $1, DX
	SHLL  CX, DX
	DECL  DX
	ANDL  DX, AX

allstored:
	ADDQ  $32, DI
	MOVQ  data+8(FP), SI
	MOVQ  curMasked+32(FP), R9
	MOVQ  mask+40(FP), R10
	MOVQ  bestLen+64(FP), R13
	MOVQ  bestScore+72(FP), R12
	MOVL  (SI)(R9*1), R11
	LEAQ  (R9)(R13*1), CX
	CMPQ  CX, R10
	JHI   done
	MOVL  -3(SI)(CX*1), R14
	TESTL AX, AX
	JZ    done

loop:
	XORL   CX, CX
	TZCNTL AX, CX
	ADDL   R15, CX
	ANDL   $31, CX
	MOVL   (DI)(CX*4), BX
	CMPQ   BX, minPrev+48(FP)
	JCS    done
	ANDQ   R10, BX
	LEAQ   (BX)(R13*1), CX
	CMPL   -3(SI)(CX*1), R14
	JNE    next
	CMPQ   CX, R10
	JHI    next
	CMPL   (SI)(BX*1), R11
	JNE    next

	// Start at byte 4. BX counts the bytes left. R8-CX stays curMasked-prevMasked.
	LEAQ 4(SI)(BX*1), CX
	LEAQ 4(SI)(R9*1), R8
	MOVQ maxLength+56(FP), BX
	SUBQ $4, BX

len8:
	CMPQ BX, $8
	JLT  lentail
	MOVQ (CX), DX
	XORQ (R8), DX
	JNZ  lendiff
	ADDQ $8, CX
	ADDQ $8, R8
	SUBQ $8, BX
	JMP  len8

lendiff:
	TZCNTQ DX, DX
	SHRQ   $3, DX
	SUBQ   DX, BX
	JMP    lendone

lentail:
	CMPQ    BX, $0
	JLE     lendone
	MOVBLZX (CX), DX
	CMPB    DL, (R8)
	JNE     lendone
	INCQ    CX
	INCQ    R8
	DECQ    BX
	JMP     lentail

lendone:
	// The distance fits in the window, so the mask gives cur-prevRaw.
	MOVQ maxLength+56(FP), DX
	SUBQ BX, DX
	SUBQ CX, R8
	ANDQ R10, R8

	// score = 1920 + 135*len - 30*floor(log2(backward)); backward > 0.
	BSRQ   R8, CX
	IMUL3Q $30, CX, CX
	IMUL3Q $135, DX, BX
	ADDQ   $1920, BX
	SUBQ   CX, BX
	CMPQ   BX, R12
	JLS    next
	MOVQ   BX, R12
	MOVQ   DX, R13
	MOVQ   out+88(FP), CX
	MOVQ   DX, hasherSearchResult_len(CX)
	MOVQ   R8, hasherSearchResult_distance(CX)
	MOVQ   BX, hasherSearchResult_score(CX)
	LEAQ   (R9)(R13*1), CX
	CMPQ   CX, R10
	JHI    done
	MOVL   -3(SI)(CX*1), R14

next:
	LEAL -1(AX), CX
	ANDL CX, AX
	JNZ  loop

done:
	// Store cur after the scan so its distance cannot be zero.
	LEAL    31(R15), CX
	ANDL    $31, CX
	MOVQ    cur+24(FP), BX
	MOVL    BX, (DI)(CX*4)
	MOVQ    keyTag+16(FP), BX
	MOVB    BL, -32(DI)(CX*1)
	SHRQ    $8, BX
	MOVQ    h+0(FP), CX
	DECW    h6b5_num(CX)(BX*2)
	RET
