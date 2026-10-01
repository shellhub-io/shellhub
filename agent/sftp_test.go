package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReportServeError(t *testing.T) {
	cases := []struct {
		name     string
		err      error
		expected string
	}{
		{name: "clean end reports nothing", err: nil, expected: ""},
		{name: "client EOF reports nothing", err: io.EOF, expected: ""},
		{name: "wrapped client EOF reports nothing", err: fmt.Errorf("serve: %w", io.EOF), expected: ""},
		{name: "real error is reported", err: errors.New("connection reset"), expected: "connection reset\n"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var stderr bytes.Buffer

			reportServeError(&stderr, tc.err)

			assert.Equal(t, tc.expected, stderr.String())
		})
	}
}
