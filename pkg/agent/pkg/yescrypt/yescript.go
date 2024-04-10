//go:build !without_cgo
// +build !without_cgo

package yescrypt

/*
#cgo LDFLAGS: -lcrypt
#include <stdlib.h>
#include <crypt.h>
*/
import "C"
import "unsafe"

// Verify verifies a yescrypt hash against a given key.
func Verify(key, hash string) bool {
	cdata := C.struct_crypt_data{}
	ckey := C.CString(key)
	chash := C.CString(hash)

	out := C.crypt_r(ckey, chash, &cdata)

	C.free(unsafe.Pointer(ckey))
	C.free(unsafe.Pointer(chash))

	return C.GoString(out) == hash
}
