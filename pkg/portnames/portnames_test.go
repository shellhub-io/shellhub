package portnames

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/shellhub-io/shellhub/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOverride(t *testing.T) {
	cases := []struct {
		description string
		content     *string
		expected    map[int]string
		err         error
		line        int
	}{
		{
			description: "parses each line into a port and its name",
			content:     new("22 ssh\n 80  web \n"),
			expected:    map[int]string{22: "ssh", 80: "web"},
		},
		{
			description: "reports the line whose port is not a number",
			content:     new("22 ssh\nhttp web\n"),
			err:         ErrLine,
			line:        2,
		},
		{
			description: "reports a line with no name",
			content:     new("22\n"),
			err:         ErrLine,
			line:        1,
		},
		{
			description: "fails to read a line longer than the scanner buffer",
			content:     new("22 " + strings.Repeat("a", 70*1024) + "\n"),
			err:         ErrRead,
		},
		{
			description: "fails to open a file that does not exist",
			err:         ErrRead,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "overrides")
			if tc.content != nil {
				require.NoError(t, os.WriteFile(path, []byte(*tc.content), 0o600))
			}

			overrides, err := Override(path)

			if tc.err == nil {
				require.NoError(t, err)
				assert.Equal(t, tc.expected, overrides)

				return
			}

			require.ErrorIs(t, err, tc.err)
			assert.Nil(t, overrides)

			if tc.line != 0 {
				var detailed errors.Error
				require.True(t, errors.As(err, &detailed))
				assert.Equal(t, tc.line, detailed.Data)
			}
		})
	}
}
