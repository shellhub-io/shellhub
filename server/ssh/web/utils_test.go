package web

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDimensionsBounds(t *testing.T) {
	cases := []struct {
		description  string
		dim          Dimensions
		expectedRows int
		expectedCols int
	}{
		{
			description:  "keeps an ordinary terminal size",
			dim:          Dimensions{Rows: 24, Cols: 80},
			expectedRows: 24,
			expectedCols: 80,
		},
		{
			description:  "keeps the largest size a winsize can hold",
			dim:          Dimensions{Rows: math.MaxUint16, Cols: math.MaxUint16},
			expectedRows: math.MaxUint16,
			expectedCols: math.MaxUint16,
		},
		{
			description:  "bounds a size the kernel could not represent",
			dim:          Dimensions{Rows: math.MaxUint32, Cols: math.MaxUint32},
			expectedRows: math.MaxUint16,
			expectedCols: math.MaxUint16,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			assert.Equal(t, tc.expectedRows, tc.dim.rows())
			assert.Equal(t, tc.expectedCols, tc.dim.cols())
		})
	}
}
