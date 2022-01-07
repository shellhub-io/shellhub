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
	ckey := C.CString(key)
	chash := C.CString(hash)

	out := C.crypt(ckey, chash)

	C.free(unsafe.Pointer(ckey))
	C.free(unsafe.Pointer(chash))

	return C.GoString(out) == hash
}
