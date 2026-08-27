package envs

import (
	"strings"
	"testing"
)

func TestAnnouncementFor(t *testing.T) {
	cases := []struct {
		edition       Edition
		wantTitle     string
		wantCommunity bool
	}{
		{Community, "Welcome to ShellHub Community!", true},
		{Enterprise, "Welcome to ShellHub Enterprise!", false},
		{Cloud, "Welcome to ShellHub Cloud!", false},
		{"", "Welcome to ShellHub Community!", true},
	}

	for _, tc := range cases {
		t.Run(string(tc.edition), func(t *testing.T) {
			got := AnnouncementFor(tc.edition)

			if !strings.Contains(got, tc.wantTitle) {
				t.Errorf("expected title %q in announcement, got:\n%s", tc.wantTitle, got)
			}

			hasCommunityLinks := strings.Contains(got, "github.com/shellhub-io/shellhub")
			if hasCommunityLinks != tc.wantCommunity {
				t.Errorf("community links: want %v, got %v", tc.wantCommunity, hasCommunityLinks)
			}
		})
	}
}
