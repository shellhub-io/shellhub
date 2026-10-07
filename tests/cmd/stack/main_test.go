package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultStackName(t *testing.T) {
	cases := []struct {
		name  string
		roots []string
		same  bool
	}{
		{
			name:  "the same checkout gets the same name",
			roots: []string{"/home/dev/src/shellhub-io/shellhub", "/home/dev/src/shellhub-io/shellhub"},
			same:  true,
		},
		{
			name:  "two worktrees get different names",
			roots: []string{"/home/dev/.t3/worktrees/a/shellhub", "/home/dev/.t3/worktrees/b/shellhub"},
			same:  false,
		},
		{
			name:  "a path that would not be a valid name still gives one",
			roots: []string{"/Users/Dev User/Shell Hub", "/"},
			same:  false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			first := defaultStackName(tc.roots[0])
			second := defaultStackName(tc.roots[1])

			assert.Regexp(t, validName, first)
			assert.Regexp(t, validName, second)
			assert.Equal(t, tc.same, first == second)
		})
	}
}
