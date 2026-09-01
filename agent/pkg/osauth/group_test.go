package osauth

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const groups = `root:x:0:root
wheel:x:10:root,user1
staff:x:50:user2
nogroup:x:65534:`

func TestParseGroupLine(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  Group
		ok    bool
	}{
		{
			name:  "simple group with members",
			input: "wheel:x:10:root,user1",
			want: Group{
				Name:     "wheel",
				Password: "x",
				GID:      10,
				Members:  []string{"root", "user1"},
			},
			ok: true,
		},
		{
			name:  "group without members",
			input: "nogroup:x:65534:",
			want: Group{
				Name:     "nogroup",
				Password: "x",
				GID:      65534,
				Members:  []string{},
			},
			ok: true,
		},
		{
			name:  "invalid parts",
			input: "badline:too:many:parts:here",
			ok:    false,
		},
		{
			name:  "bad gid",
			input: "g:x:badgid:member",
			ok:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseGroupLine(tt.input)
			if !tt.ok {
				assert.Error(t, err)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want.Name, got.Name)
			assert.Equal(t, tt.want.Password, got.Password)
			assert.Equal(t, tt.want.GID, got.GID)
			assert.Equal(t, tt.want.Members, got.Members)
		})
	}
}

func TestParseGroupReader(t *testing.T) {
	tests := []struct {
		name      string
		data      string
		wantCount int
		wantGID   uint32
		wantParts []string
	}{
		{
			name:      "default groups",
			data:      groups,
			wantCount: 4,
			wantGID:   10,
			wantParts: []string{"root", "user1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.data)
			m, err := parseGroupReader(reader)
			require.NoError(t, err)
			assert.Len(t, m, tt.wantCount)

			g, ok := m["wheel"]
			assert.True(t, ok)
			assert.Equal(t, tt.wantGID, g.GID)
			assert.Equal(t, tt.wantParts, g.Members)
		})
	}
}

func TestListGroupsFromFile(t *testing.T) {
	tests := []struct {
		name      string
		username  string
		wantFound []uint32
	}{
		{name: "user1 belongs to wheel", username: "user1", wantFound: []uint32{10}},
		{name: "user2 belongs to staff", username: "user2", wantFound: []uint32{50}},
		{name: "no groups for missing user", username: "unknown", wantFound: []uint32{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(groups)
			got, err := ListGroupsFromFile(tt.username, reader)
			require.NoError(t, err)

			for _, want := range tt.wantFound {
				assert.Contains(t, got, want)
			}
			if len(tt.wantFound) == 0 {
				assert.Empty(t, got)
			}
		})
	}
}

func TestParseUint32(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    uint32
		wantErr bool
	}{
		{name: "valid", input: "1000", want: 1000, wantErr: false},
		{name: "invalid", input: "notanumber", wantErr: true},
		{name: "empty", input: "", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseUint32(tt.input)
			if tt.wantErr {
				assert.Error(t, err)

				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
