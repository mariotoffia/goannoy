// +build neon

// https://blog.felixge.de/go-arm64-function-call-assembly/

#include "textflag.h"

// func ·neonDotProductF32(v1, v2 *float32, vectorLength uint32) float32
TEXT ·neonDotProductF32(SB),$0-20
    MOV  v1.ptr(R0), X4       // X4 holds the pointer to v1
    MOV  v2.ptr(R1), X5       // X5 holds the pointer to v2
    MOV  vectorLength(R2), W3  // W3 holds the length of vectors

    MOVZ $0x0, W0              // W0 holds the index i
    FMOV ZR, S2                // S2 holds the dot product result

loop:
    CMP W0, W3
    BGE done

    LDR S0, (X4)(W0*4)         // Load v1[i] into S0
    LDR S1, (X5)(W0*4)         // Load v2[i] into S1
    FMUL S0, S1, S0            // Multiply v1[i] * v2[i]
    FADD S2, S0, S2            // Add the result to the dot product: dotProduct += v1[i] * v2[i]

    ADD W0, W0, $1             // Increment the index i
    B   loop                   // Jump back to the loop

done:
    FMOV F0, S2                // Return the dot product result in F0
    RET
