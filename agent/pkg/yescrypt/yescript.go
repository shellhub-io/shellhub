// +build !without_cgo

package yescrypt

/*
#cgo LDFLAGS: -lcrypt
#include <stdlib.h>
#include <crypt.h>
*/
import "C"

import (
	"unsafe"
)

// Verify verifies a yescrypt hash against a given key
func Verify(key, hash string) bool {
	ckey := C.CString(key)
	chash := C.CString(hash)

	defer C.free(unsafe.Pointer(ckey))
	defer C.free(unsafe.Pointer(chash))

	out := C.crypt(ckey, chash)

	return C.GoString(out) == hash
}
