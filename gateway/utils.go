package main

import (
	"golang.org/x/sys/unix"
	"log"
	"os"
	"path"
	"strconv"
	"strings"
)

// rlimitMaxNumFiles returns the maximum number of open files allowed by the system.
func rlimitMaxNumFiles() int {
	var rLimit unix.Rlimit
	if err := unix.Getrlimit(unix.RLIMIT_NOFILE, &rLimit); err != nil {
		log.Fatal(err)
	}
	return int(rLimit.Max)
}

// getSysctl retrieves the value of a given sysctl parameter.
func getSysctl(sysctl string) int {
	data, err := os.ReadFile(
		path.Join("/proc/sys/", strings.ReplaceAll(sysctl, ".", "/")),
	)
	if err != nil {
		log.Println(err)
		return -1
	}

	value, err := strconv.Atoi(strings.Trim(string(data), " \n"))
	if err != nil {
		log.Println(err)
		return -1
	}

	return value
}
