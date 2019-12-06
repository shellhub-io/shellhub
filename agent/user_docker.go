// +build docker

package main

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
	"os/user"
	"unsafe"
)
import "strconv"

func Auth(user string, passwd string) bool {
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

func lookupUser(username string) *user.User {
	cuser := C.CString(username)
	defer C.free(unsafe.Pointer(cuser))

	cfilename := C.CString("/host/etc/passwd")
	defer C.free(unsafe.Pointer(cfilename))

	cmode := C.CString("r")
	defer C.free(unsafe.Pointer(cmode))

	f := C.fopen(cfilename, cmode)
	defer C.fclose(f)

	var pwd *C.struct_passwd
	for {
		if pwd = C.fgetpwent(f); pwd == nil {
			return nil
		}

		if C.strcmp(cuser, pwd.pw_name) == 0 {
			return &user.User{
				Uid:      strconv.FormatUint(uint64(pwd.pw_uid), 10),
				Gid:      strconv.FormatUint(uint64(pwd.pw_gid), 10),
				Username: C.GoString(pwd.pw_name),
				Name:     C.GoString(pwd.pw_gecos),
				HomeDir:  C.GoString(pwd.pw_dir),
			}
		}
	}

	return nil
}
