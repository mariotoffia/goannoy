// +build avx256
// cSpell: disable

#include "textflag.h"

// void avxDotProductF32AVX512(float* result, const float* x, const float *y, int f)
TEXT ·avxDotProductF32AVX512(SB), NOSPLIT, $0
    // Save registers
    PUSHQ BP
    MOVQ SP, BP

    // Load arguments
    MOVQ result+0(FP), DI
    MOVQ x+8(FP), SI
    MOVQ y+16(FP), DX
    MOVL f+24(FP), AX

    // Initialize dot product register
    VXORPS Z0, Z0, Z0

    // Main loop for dot product
DotProductLoop:
    CMPL AX, $15
    JLE DotProductRemaining

    // Load 16 values from x and y and perform FMA
    VMOVUPS (SI), Z1
    VMOVUPS (DX), Z2
    VFMADD231PS Z2, Z1, Z0

    // Increment x, y, and decrement f by 16
    ADDQ $64, SI
    ADDQ $64, DX
    SUBL $16, AX

    JMP DotProductLoop

DotProductRemaining:
    // Calculate the horizontal sum of the dot product register
    VREDUCEPS $0xFF, Z0, Z0
    VMOVSS Z0, X0

    // Remaining values loop
RemainingLoop:
    TESTL AX, AX
    JZ Done

    // Load x and y values
    MOVSS (SI), X1
    MOVSS (DX), X2

    // Multiply and add to the result
    MULSS X2, X1
    ADDSS X1, X0

    // Increment x, y, and decrement f
    ADDQ $4, SI
    ADDQ $4, DX
    DECL AX

    JMP RemainingLoop

Done:
    // Store the final result
    MOVSS X0, (DI)

    // Restore registers and return
    MOVQ BP, SP
    POPQ BP
    RET
