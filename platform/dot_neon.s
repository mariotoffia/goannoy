// +build neon

// https://blog.felixge.de/go-arm64-function-call-assembly/

#include "textflag.h"

// Function declaration
TEXT ·neonDotProductF32(SB), NOSPLIT, $0-32
  // Load vector pointers
  MOVD v1+0(FP), R4
  MOVD v2+8(FP), R5

  // Load vector length
  MOVD length+16(FP), R3

  // Prepare loop counter and result register
  MOVD $0, R6
  FMOVDZR F0

  // Loop
neon_dot_product_loop:
  // Load float32 values from v1 and v2
  FMOVWU (R4)(R6*4), F1
  FMOVWU (R5)(R6*4), F2

  // Multiply and accumulate
  FMADDS F1, F2, F0, F0

  // Increment loop counter and compare with the length
  ADD $1, R6, R6
  CMP R6, R3
  BLT neon_dot_product_loop

  // Store result and return
  FMOVWDU F0, res+24(FP)
  RET
