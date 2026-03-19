#include "textflag.h"

// func dotFloat32AcceleratedUnsafeARM64(a, b *float32, vectorLength int) float32
//
// NEON-vectorized float32 dot product.
// Main loop processes 16 floats/iteration using 4 vector accumulators
// to hide FMA pipeline latency (~4 cycles on most ARM64 cores).
TEXT ·dotFloat32AcceleratedUnsafeARM64(SB), NOSPLIT, $0-28
	MOVD	a+0(FP), R0
	MOVD	b+8(FP), R1
	MOVD	vectorLength+16(FP), R2

	// Zero 4 NEON accumulators (128-bit = 4×float32 each).
	// VEOR is recognized as a zeroing idiom — breaks dependencies.
	VEOR	V0.B16, V0.B16, V0.B16
	VEOR	V1.B16, V1.B16, V1.B16
	VEOR	V2.B16, V2.B16, V2.B16
	VEOR	V3.B16, V3.B16, V3.B16

	// Main loop: 16 floats per iteration (4×VLD1 of 4 floats each).
	// 4 independent VFMLA chains achieve full throughput.
loop16:
	CMP	$16, R2
	BLT	loop4_pre

	VLD1.P	16(R0), [V4.S4]
	VLD1.P	16(R1), [V5.S4]
	VFMLA	V5.S4, V4.S4, V0.S4

	VLD1.P	16(R0), [V6.S4]
	VLD1.P	16(R1), [V7.S4]
	VFMLA	V7.S4, V6.S4, V1.S4

	VLD1.P	16(R0), [V8.S4]
	VLD1.P	16(R1), [V9.S4]
	VFMLA	V9.S4, V8.S4, V2.S4

	VLD1.P	16(R0), [V10.S4]
	VLD1.P	16(R1), [V11.S4]
	VFMLA	V11.S4, V10.S4, V3.S4

	SUB	$16, R2
	B	loop16

	// Merge 4 accumulators into V0 before the 4-element tail.
loop4_pre:
	WORD	$0x4E21D400 // FADD V0.4S, V0.4S, V1.4S
	WORD	$0x4E23D442 // FADD V2.4S, V2.4S, V3.4S
	WORD	$0x4E22D400 // FADD V0.4S, V0.4S, V2.4S

	// Medium tail: 4 floats per iteration.
loop4:
	CMP	$4, R2
	BLT	reduce

	VLD1.P	16(R0), [V4.S4]
	VLD1.P	16(R1), [V5.S4]
	VFMLA	V5.S4, V4.S4, V0.S4

	SUB	$4, R2
	B	loop4

	// Horizontal reduction: 4 lanes → 1 scalar.
	// FADDP pairwise: {s0,s1,s2,s3} → {s0+s1,s2+s3,...} → {total,...}
reduce:
	WORD	$0x6E20D400 // FADDP V0.4S, V0.4S, V0.4S
	WORD	$0x6E20D400 // FADDP V0.4S, V0.4S, V0.4S

	// Scalar tail: 0–3 remaining elements.
	// F0 aliases V0.S[0], so scalar FMA accumulates into the same register.
tail:
	CBZ	R2, done
	FMOVS	(R0), F4
	FMOVS	(R1), F5
	FMADDS	F4, F0, F5, F0
	ADD	$4, R0
	ADD	$4, R1
	SUB	$1, R2
	B	tail

done:
	FMOVS	F0, ret+24(FP)
	RET
