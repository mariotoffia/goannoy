package tests

import (
	"testing"
	"unsafe"

	"github.com/mariotoffia/goannoy/tests/utils"
)

func BenchmarkCopy(b *testing.B) {
	src := make([]byte, 1024*1024*1024)
	dst := make([]byte, len(src))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		copy(dst, src)
	}
}

func BenchmarkCopyUnsafe(b *testing.B) {
	src := make([]byte, 1024*1024*1024)
	dst := make([]byte, len(src))

	srcPtr := unsafe.Pointer(&src[0])
	dstPtr := unsafe.Pointer(&dst[0])

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for i := 0; i < len(src); i++ {
			*(*byte)(unsafe.Pointer(uintptr(dstPtr) + uintptr(i))) = *(*byte)(unsafe.Pointer(uintptr(srcPtr) + uintptr(i)))
		}
	}
}

func BenchmarkCGOMemcpy(b *testing.B) {
	src := make([]byte, 1024*1024*1024)
	dst := make([]byte, len(src))

	srcPtr := unsafe.Pointer(&src[0])
	dstPtr := unsafe.Pointer(&dst[0])

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		utils.Memcpy(dstPtr, srcPtr, int64(len(src)))
	}
}
