// +build neon

// https://blog.felixge.de/go-arm64-function-call-assembly/

#include "textflag.h"

// void neonDotProductF32(float* result, const float* x, const float *y, int f)
TEXT ·neonDotProductF32(SB), NOSPLIT, $0-32
    // Save registers
    STP X29, X30, [SP, #-16]!
    MOV X29, SP

    // Load arguments
    MOV X8, result+0(FP)
    MOV X9, x+8(FP)
    MOV X10, y+16(FP)
    MOV W11, f+24(FP)

    // Initialize dot product register
    FMOV V0.4S, #0.0

    // Main loop for dot product
DotProductLoop:
    CMP W11, #3
    BLS DotProductRemaining

    // Load 4 values from x and y and perform FMA
    LD1 {V1.4S}, [X9], #16
    LD1 {V2.4S}, [X10], #16
    FMLA V0.4S, V1.4S, V2.4S

    // Decrement f by 4
    SUB W11, W11, #4

    B DotProductLoop

DotProductRemaining:
    // Calculate the horizontal sum of the dot product register
    FADD V0.4S, V0.4S, V0.S[1]
    FADD V0.4S, V0.4S, V0.S[2]
    FADD V0.4S, V0.4S, V0.S[3]

    // Remaining values loop
RemainingLoop:
    CBZ W11, Done

    // Load x and y values
    LD1 {V1.S}[0], [X9], #4
    LD1 {V2.S}[0], [X10], #4

    // Multiply and add to the result
    FMLA V0.S, V1.S, V2.S

    // Decrement f
    SUB W11, W11, #1

    B RemainingLoop

Done:
    // Store the final result
    STR S0, [X8]

    // Restore registers and return
    MOV SP, X29
    LDP X29, X30, [SP], #16
    RET
