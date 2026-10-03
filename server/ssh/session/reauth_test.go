package session

import (
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/stretchr/testify/assert"
)

func TestNeedsReauth(t *testing.T) {
	now := time.Date(2026, 7, 22, 12, 0, 0, 0, time.UTC)
	clock.Freeze(t, now)

	ptr := func(v int) *int { return &v }
	at := func(d time.Duration) *time.Time {
		ts := now.Add(d)

		return &ts
	}

	cases := []struct {
		description  string
		lastReauthAt *time.Time
		period       *int
		expected     bool
	}{
		{"always when period is nil", at(-time.Minute), nil, true},
		{"always when period is zero", at(-time.Minute), ptr(0), true},
		{"challenges when never re-authed", nil, ptr(3600), true},
		{"fresh within the window", at(-30 * time.Minute), ptr(3600), false},
		{"stale at the window boundary", at(-time.Hour), ptr(3600), true},
		{"stale past the window", at(-2 * time.Hour), ptr(3600), true},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			assert.Equal(t, tc.expected, needsReauth(tc.lastReauthAt, tc.period))
		})
	}
}
