package goc

import "unsafe"

func unsafeCast[B, A any](a A) B {
	return *(*B)(unsafe.Pointer(&a))
}

func unsafeSliceCast[B, A any](a []A) B {
	return *(*B)(unsafe.Pointer(unsafe.SliceData(a)))
}
