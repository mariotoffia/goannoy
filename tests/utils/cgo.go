package utils

// #include <string.h>
import "C"
import "unsafe"

func Memcpy(dst, src unsafe.Pointer, size int64) {
	C.memcpy(dst, src, C.size_t(size))
}
