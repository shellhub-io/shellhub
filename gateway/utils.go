package main

import (
	"log"
	"math"
	"os"
	"path"
	"strconv"
	"strings"

	"golang.org/x/sys/unix"
)

// rlimitMaxNumFiles returns the maximum number of open files allowed by the system.
func rlimitMaxNumFiles() int {
	var rLimit unix.Rlimit
	if err := unix.Getrlimit(unix.RLIMIT_NOFILE, &rLimit); err != nil {
		log.Fatal(err)
	}

	if rLimit.Max > math.MaxInt {
		return math.MaxInt
	}

	return int(rLimit.Max) //nolint:gosec
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

// halfString return the halfString of the string.
func halfString(s string) string {
	runes := []rune(s)
	n := len(runes) / 2

	return string(runes[:n]) + "..."
}
