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

func reportSizes(received <-chan os.Signal) {
	for range received {
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

func reportInterrupt(received <-chan os.Signal) {
	<-received

	if _, err := fmt.Fprintln(os.Stdout, "GOT-INT"); err != nil {
		os.Exit(1)
	}

	os.Exit(3)
}

func main() {
	reporters := map[string]struct {
		signal syscall.Signal
		report func(<-chan os.Signal)
	}{
		"WINCH": {syscall.SIGWINCH, reportSizes},
		"INT":   {syscall.SIGINT, reportInterrupt},
	}

	if len(os.Args) != 2 {
		_, _ = fmt.Fprintln(os.Stderr, "usage: signal-reporter WINCH|INT")
		os.Exit(2)
	}

	reporter, ok := reporters[os.Args[1]]
	if !ok {
		_, _ = fmt.Fprintln(os.Stderr, "unknown signal:", os.Args[1])
		os.Exit(2)
	}

	received := make(chan os.Signal, 1)
	signal.Notify(received, reporter.signal)

	if _, err := fmt.Fprintln(os.Stdout, "READY"); err != nil {
		os.Exit(1)
	}

	reporter.report(received)
}
