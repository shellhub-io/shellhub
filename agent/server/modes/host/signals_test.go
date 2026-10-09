package host

import (
	"os"
	"os/signal"
	"syscall"
	"testing"
	"time"

	gliderssh "github.com/gliderlabs/ssh"
)

func TestForwardSignals_NeverSignalsTheAgentsOwnGroup(t *testing.T) {
	cases := []struct {
		description string
		pgid        int
	}{
		{
			description: "a terminal without a foreground group",
			pgid:        0,
		},
		{
			description: "the agent's own group",
			pgid:        syscall.Getpgrp(),
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			caught := make(chan os.Signal, 1)
			signal.Notify(caught, syscall.SIGUSR1)
			t.Cleanup(func() { signal.Stop(caught) })

			sess := newFakeSession("signals", "root")
			sess.signals = make(chan chan<- gliderssh.Signal, 1)

			stop := forwardSignals(sess, func() int { return tc.pgid })
			t.Cleanup(stop)

			registered := <-sess.signals
			registered <- gliderssh.SIGUSR1

			select {
			case <-caught:
				t.Fatal("the agent signalled its own process group")
			case <-time.After(300 * time.Millisecond):
			}
		})
	}
}
