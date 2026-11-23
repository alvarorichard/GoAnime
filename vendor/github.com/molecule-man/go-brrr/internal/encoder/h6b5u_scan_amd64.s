//go:build amd64 && !purego

#include "go_asm.h"
#include "textflag.h"

// func h6b5uFindInBucket(h *h6b5u, data unsafe.Pointer, key, cur, curMasked,
//	mask, minPrev, maxLength, bestLen, bestScore uint,
//	nextKey uint, out *hasherSearchResult)
//
// No local spills: each store consumes a store-buffer entry.
// TZCNT acts as BSF on CPUs without BMI1. Both give the same result for nonzero input.
TEXT ·h6b5uFindInBucket(SB), NOSPLIT|NOFRAME, $0-96
	// Prefetch the next bucket and data at slot (num[nextKey]-1)&31.
	MOVQ       h+0(FP), DI
	MOVQ       nextKey+80(FP), CX
	MOVWLZX    h6b5u_num(DI)(CX*2), DX
	SHLQ       $7, CX
	LEAQ       h6b5u_buckets(DI)(CX*1), BX
	PREFETCHT0 (BX)
	PREFETCHT0 64(BX)
	DECL       DX
	ANDL       $31, DX
	MOVL       (BX)(DX*4), DX
	ANDQ       mask+40(FP), DX
	MOVQ       data+8(FP), SI
	PREFETCHT0 (SI)(DX*1)

	// R15 = n, AX = lowest index to scan: n-32, or 0 for a bucket with
	// fewer stored entries.
	MOVQ    h+0(FP), DI
	MOVQ    key+16(FP), CX
	MOVWLZX h6b5u_num(DI)(CX*2), R15
	SHLQ    $7, CX
	LEAQ    h6b5u_buckets(DI)(CX*1), DI
	XORL    AX, AX
	LEAL    -32(R15), CX
	CMPL    R15, $32
	CMOVLHI CX, AX

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

	// Scan newest first: slots (n-1)&31 down to AX&31.
	MOVL R15, DX

loop:
	CMPL   DX, AX
	JLS    done
	DECL   DX
	MOVL   DX, CX
	ANDL   $31, CX
	MOVL   (DI)(CX*4), BX
	CMPQ   BX, minPrev+48(FP)
	JCS    done
	ANDQ   R10, BX
	LEAQ   (BX)(R13*1), CX
	CMPL   -3(SI)(CX*1), R14
	JNE    loop
	CMPQ   CX, R10
	JHI    loop
	CMPL   (SI)(BX*1), R11
	JNE    loop
	MOVL   DX, R15

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
	JLS    restore
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


restore:
	MOVL R15, DX
	JMP  loop

done:
	// Store cur after the scan so its distance cannot be zero.
	MOVQ    h+0(FP), CX
	MOVQ    key+16(FP), BX
	MOVWLZX h6b5u_num(CX)(BX*2), DX
	INCW    h6b5u_num(CX)(BX*2)
	ANDL    $31, DX
	MOVQ    cur+24(FP), R8
	SHLQ    $7, BX
	LEAQ    h6b5u_buckets(CX)(BX*1), CX
	MOVL    R8, (CX)(DX*4)
	RET
