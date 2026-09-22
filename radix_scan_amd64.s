//go:build !purego && !race

#include "textflag.h"

// The keys are non-negative, non-NaN float64s, so CMPPD has the same
// ordering as an unsigned comparison of their bit patterns.
TEXT ·radixScan(SB), NOSPLIT, $0-33
	MOVQ points_base+0(FP), SI
	MOVQ points_len+8(FP), CX
	XORQ AX, AX
	XORQ DX, DX
	TESTQ CX, CX
	JZ done
	MOVQ (SI), R8
	MOVQ R8, X0
	PUNPCKLQDQ X0, X0
	MOVOU X0, X1
	PXOR X2, X2
	PXOR X3, X3
	CMPQ CX, $8
	JB tail
loop:
	MOVOU (SI), X4
	MOVOU 16(SI), X5
	MOVOU 32(SI), X6
	MOVOU 48(SI), X7
	SHUFPD $1, X4, X1
	MOVOU X4, X8
	SHUFPD $1, X5, X8
	MOVOU X5, X9
	SHUFPD $1, X6, X9
	MOVOU X6, X10
	SHUFPD $1, X7, X10
	CMPPD X4, X1, $6
	CMPPD X5, X8, $6
	CMPPD X6, X9, $6
	CMPPD X7, X10, $6
	POR X1, X8
	POR X9, X10
	POR X8, X10
	POR X10, X3
	MOVOU X7, X1
	PXOR X0, X4
	PXOR X0, X5
	PXOR X0, X6
	PXOR X0, X7
	POR X4, X5
	POR X6, X7
	POR X5, X7
	POR X7, X2
	ADDQ $64, SI
	SUBQ $8, CX
	CMPQ CX, $8
	JAE loop
	MOVQ X2, AX
	PSRLDQ $8, X2
	MOVQ X2, R9
	ORQ R9, AX
	MOVMSKPD X3, DX
	PSRLDQ $8, X1
	MOVQ X1, R9
	JMP tailCheck
tail:
	MOVQ R8, R9
tailCheck:
	TESTQ CX, CX
	JZ done
tailLoop:
	MOVQ (SI), R10
	MOVQ R10, R11
	XORQ R8, R11
	ORQ R11, AX
	CMPQ R10, R9
	SETCS R11
	MOVBQZX R11, R11
	ORQ R11, DX
	MOVQ R10, R9
	ADDQ $8, SI
	DECQ CX
	JNZ tailLoop
done:
	MOVQ AX, varying+24(FP)
	TESTQ DX, DX
	SETEQ ordered+32(FP)
	RET
