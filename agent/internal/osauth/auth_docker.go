// +build docker

package osauth

/*
#cgo LDFLAGS: -lcrypt
#define _GNU_SOURCE 1
#include <stdlib.h>
#include <stdio.h>
#include <shadow.h>
#include <string.h>
#include <pwd.h>
#include <crypt.h>
*/
import "C"

import (
	"unsafe"
)

const passwdFilename = "/host/etc/passwd"

func AuthUser(user, passwd string) bool {
	cuser := C.CString(user)
	defer C.free(unsafe.Pointer(cuser))

	cpasswd := C.CString(passwd)
	defer C.free(unsafe.Pointer(cpasswd))

	cfilename := C.CString("/host/etc/shadow")
	defer C.free(unsafe.Pointer(cfilename))

	cmode := C.CString("r")
	defer C.free(unsafe.Pointer(cmode))

	f := C.fopen(cfilename, cmode)
	defer C.fclose(f)

	var pwd *C.struct_spwd
	for {
		if pwd = C.fgetspent(f); pwd == nil {
			return false
		}

		if C.strcmp(cuser, pwd.sp_namp) == 0 {
			break
		}
	}

	if pwd == nil {
		return false
	}

	crypted := C.crypt(cpasswd, pwd.sp_pwdp)

	if C.strcmp(crypted, pwd.sp_pwdp) != 0 {
		return false
	}

	return true
}
