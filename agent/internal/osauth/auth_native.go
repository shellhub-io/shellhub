// +build !docker

package osauth

/*
#cgo LDFLAGS: -lcrypt
#define _GNU_SOURCE 1
#include <stdlib.h>
#include <stdio.h>
#include <shadow.h>
#include <string.h>
#include <crypt.h>
*/
import "C"

import "unsafe"

const passwdFilename = "/etc/passwd"

func AuthUser(user, passwd string) bool {
	cuser := C.CString(user)
	defer C.free(unsafe.Pointer(cuser))

	cpasswd := C.CString(passwd)
	defer C.free(unsafe.Pointer(cpasswd))

	pwd := C.getspnam(cuser)
	if pwd == nil {
		return false
	}

	crypted := C.crypt(cpasswd, pwd.sp_pwdp)

	if C.strcmp(crypted, pwd.sp_pwdp) != 0 {
		return false
	}

	return true
}
