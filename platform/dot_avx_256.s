// +build avx256

// cSpell: disable

#include "textflag.h"

// void hsum256_ps_avx(float* result, __m256 v)
TEXT ·hsum256_ps_avx(SB), NOSPLIT, $0
    // Extract upper 128 bits and add them to the lower 128 bits
    VEXTRACTF128 $1, Y0, X0
    VADDPS X0, X1, X1

    // Move high 64 bits to low 64 bits and add them
    VMOVHLPS X1, X1, X0
    VADDPS X0, X1, X1

    // Shuffle the 32-bit float and add it
    VSHUFPS $0x55, X1, X1, X0
    VADDSS X0, X1, X1

    // Store the result
    VMOVSS X1, 0(DI)
    RET

// void avxDotProduct(float* result, const float* x, const float *y, int f)
TEXT ·avxDotProduct(SB), NOSPLIT, $0
    // Save registers
    PUSHQ BP
    MOVQ SP, BP

    // Load arguments
    MOVQ result+0(FP), DI
    MOVQ x+8(FP), SI
    MOVQ y+16(FP), DX
    MOVL f+24(FP), AX

    // Initialize dot product register
    VXORPS Y0, Y0, Y0

    // Main loop for dot product
DotProductLoop:
    CMPL AX, $7
    JLE DotProductRemaining

    // Load 8 values from x and y and multiply them
    VMOVUPS (SI), Y1
    VMOVUPS (DX), Y2
    VMULPS Y2, Y1, Y1

    // Add the result to the dot product register
    VADDPS Y1, Y0, Y0

    // Increment x, y, and decrement f by 8
    ADDQ $32, SI
    ADDQ $32, DX
    SUBL $8, AX

    JMP DotProductLoop

DotProductRemaining:
    // Calculate the horizontal sum of the dot product register
    MOVQ DI, CX
    CALL ·hsum256_ps_avx(SB)

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
