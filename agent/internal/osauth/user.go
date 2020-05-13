package osauth

/*
#cgo LDFLAGS: -lcrypt
#define _GNU_SOURCE 1
#include <stdlib.h>
#include <stdio.h>
#include <string.h>
#include <pwd.h>
*/
import "C"

import (
	"strconv"
	"unsafe"
)

type User struct {
	UID      string
	GID      string
	Username string
	Name     string
	HomeDir  string
	Shell    string
}

func LookupUser(username string) *User {
	cuser := C.CString(username)
	defer C.free(unsafe.Pointer(cuser))

	cfilename := C.CString(passwdFilename)
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
			return &User{
				UID:      strconv.FormatUint(uint64(pwd.pw_uid), 10),
				GID:      strconv.FormatUint(uint64(pwd.pw_gid), 10),
				Username: C.GoString(pwd.pw_name),
				Name:     C.GoString(pwd.pw_gecos),
				HomeDir:  C.GoString(pwd.pw_dir),
				Shell:    C.GoString(pwd.pw_shell),
			}
		}
	}

	return nil
}
