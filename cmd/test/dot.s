// cSpell: disable
#include "textflag.h"

// func dot(x, y *float64, n int) float64
TEXT ·dot(SB), NOSPLIT, $0-56
	MOVD    x_base+0(FP), R0;		// x argument
	MOVD    R0, R3;
	MOVD    y_base+8(FP), R1;	// y argument
	MOVD    R1, R4;
	MOVD    n_base+16(FP), R2;	// n argument
	FMOVD		$(0.0), F0            // Set D0 to 0.0 (floating point value)

loop:
	// Implement the dot product here
	
done:
	SUB     R0, R3;                                              
	SUB     R1, R4;                                              
	MOVD    R3, ret+56(FP);					// Return the dot product (float64)                                 
	RET
