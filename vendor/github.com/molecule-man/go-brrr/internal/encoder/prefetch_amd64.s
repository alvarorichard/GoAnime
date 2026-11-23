//go:build amd64 && !purego

#include "textflag.h"

// func prefetch2(a, b unsafe.Pointer)
TEXT ·prefetch2(SB), NOSPLIT|NOFRAME, $0-16
	MOVQ       a+0(FP), AX
	MOVQ       b+8(FP), BX
	PREFETCHT0 (AX)
	PREFETCHT0 (BX)
	RET
