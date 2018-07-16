package atreugo

import (
	"unsafe"
)

func panicOnError(err error) {
	if err != nil {
		panic(err)
	}
}

// B2S convert bytes array to string without memory allocation (non safe)
func B2S(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

// index returns the first index of the target string `t`, or
// -1 if no match is found.
func indexOf(vs []string, t string) int {
	for i, v := range vs {
		if v == t {
			return i
		}
	}
	return -1
}

// include returns `true` if the target string t is in the
// slice.
func include(vs []string, t string) bool {
	return indexOf(vs, t) >= 0
}
