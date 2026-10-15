package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"unsafe"
)

type winsize struct {
	rows, cols, xpixel, ypixel uint16
}

func terminalSize(fd uintptr) (winsize, error) {
	var size winsize

	if _, _, errno := syscall.Syscall(syscall.SYS_IOCTL, fd, syscall.TIOCGWINSZ, uintptr(unsafe.Pointer(&size))); errno != 0 { //nolint:gosec // TIOCGWINSZ fills a winsize the kernel defines; the file builds alone in the agent image, without x/sys
		return size, errno
	}

	return size, nil
}

func main() {
	resized := make(chan os.Signal, 1)
	signal.Notify(resized, syscall.SIGWINCH)

	if _, err := fmt.Fprintln(os.Stdout, "READY"); err != nil {
		os.Exit(1)
	}

	for range resized {
		size, err := terminalSize(os.Stdin.Fd())
		if err != nil {
			_, _ = fmt.Fprintln(os.Stderr, "reading the terminal size:", err)
			os.Exit(1)
		}

		if _, err := fmt.Fprintf(os.Stdout, "SIZE %d %d\n", size.rows, size.cols); err != nil {
			os.Exit(1)
		}
	}
}
