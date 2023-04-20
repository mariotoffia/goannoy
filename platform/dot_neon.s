// +build neon

// https://blog.felixge.de/go-arm64-function-call-assembly/

#include "textflag.h"

// func DotProduct(v1, v2 *float32, length int) float32
TEXT ·neonDotProductF32(SB), NOSPLIT, $0-28
    MOVD R0, R0   // Load the first vector pointer (v1) into x0
    MOVD R1, R1   // Load the second vector pointer (v2) into x1
    MOVD R2, R2   // Load the vector length into w2

    FMOVS F0, ZR // Initialize the result (dot product) to 0

    // Neon loop
    CBZ R2, done
loop:
    MOVD.P 16(R0), V1 // Load 4 float32 values from v1
    MOVD.P 16(R1), V2 // Load 4 float32 values from v2

    FMULS V3.S[0], V1.S[0], V2.S[0] // Multiply v1 and v2 element-wise
    FMULS V3.S[1], V1.S[1], V2.S[1]
    FMULS V3.S[2], V1.S[2], V2.S[2]
    FMULS V3.S[3], V1.S[3], V2.S[3]

    FADDS F1, V3.S[0], V3.S[1]
    FADDS F1, F1, V3.S[2]
    FADDS F1, F1, V3.S[3]

    FADDS F0, F0, F1          // Add the sum to the result (dot product)

    SUBS R2, R2, $4          // Decrement the length by 4
    BGT loop                 // Continue loop if there are more elements

done:
    FMOVS F0, F0 // Return the result (dot product) in f0
    RET         // Return
