//go:build amd64 && !goannoy_scalar

#include "textflag.h"

// SSE2 is guaranteed by the amd64 baseline; no CPU feature probe is needed.
TEXT ·dotAsm(SB), NOSPLIT, $0-28
 MOVQ a+0(FP), AX
 MOVQ b+8(FP), BX
 MOVQ n+16(FP), CX
 XORPS X0,X0
loop4:
 CMPQ CX,$4
 JL reduce
 MOVUPS (AX),X1
 MOVUPS (BX),X2
 MULPS X2,X1
 ADDPS X1,X0
 ADDQ $16,AX
 ADDQ $16,BX
 SUBQ $4,CX
 JMP loop4
reduce:
 MOVAPS X0,X1
 SHUFPS $0x4e,X1,X1
 ADDPS X1,X0
 MOVAPS X0,X1
 SHUFPS $0xb1,X1,X1
 ADDSS X1,X0
tail:
 TESTQ CX,CX
 JE done
 MOVSS (AX),X1
 MULSS (BX),X1
 ADDSS X1,X0
 ADDQ $4,AX
 ADDQ $4,BX
 DECQ CX
 JMP tail
done:
 MOVSS X0,ret+24(FP)
 RET
