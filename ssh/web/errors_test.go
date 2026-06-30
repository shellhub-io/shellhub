package web

import (
	"errors"
	"testing"

	"github.com/shellhub-io/shellhub/ssh/pkg/banner"
)

func TestErrors(t *testing.T) {
	t.Run("ErrAccessDenied", func(t *testing.T) {
		if ErrAccessDenied == nil {
			t.Fatal("ErrAccessDenied must not be nil")
		}

		expected := "access to the device has been denied"
		if ErrAccessDenied.Error() != expected {
			t.Errorf("ErrAccessDenied message: got %q, want %q", ErrAccessDenied.Error(), expected)
		}
	})

	t.Run("ErrInvalidSSHID", func(t *testing.T) {
		if ErrInvalidSSHID == nil {
			t.Fatal("ErrInvalidSSHID must not be nil")
		}

		expected := "invalid sshid format"
		if ErrInvalidSSHID.Error() != expected {
			t.Errorf("ErrInvalidSSHID message: got %q, want %q", ErrInvalidSSHID.Error(), expected)
		}
	})
}

// TestMapBannerErrorSentinels verifies the load-bearing contract: mapBannerError
// must return the exact sentinel that the web UI and its errorMap key on. A
// mismatch between a sentinel's string and the UI's error-map key would silently
// show the wrong user-facing message without breaking any compile-time check.
func TestMapBannerErrorSentinels(t *testing.T) {
	cases := []struct {
		description string
		kind        banner.Kind
		want        error
	}{
		{
			description: "KindConnectionFailed maps to ErrConnect",
			kind:        banner.KindConnectionFailed,
			want:        ErrConnect,
		},
		{
			description: "KindAccessDenied maps to ErrAccessDenied",
			kind:        banner.KindAccessDenied,
			want:        ErrAccessDenied,
		},
		{
			description: "KindInvalidSSHID maps to ErrInvalidSSHID",
			kind:        banner.KindInvalidSSHID,
			want:        ErrInvalidSSHID,
		},
		{
			description: "KindNone (empty message) maps to ErrConnect",
			kind:        banner.KindNone,
			want:        ErrConnect,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			e := NewBannerError(banner.Message(tc.kind))
			got := mapBannerError(e)

			if !errors.Is(got, tc.want) {
				t.Errorf("mapBannerError for %v: got %v, want %v", tc.kind, got, tc.want)
			}
		})
	}
}
