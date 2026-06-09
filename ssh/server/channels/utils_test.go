package channels

import (
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// emptyReader always returns (0, nil), the broken-reader case that makes io.Copy
// busy-loop without the guard.
type emptyReader struct{}

func (emptyReader) Read(_ []byte) (int, error) { return 0, nil }

// scriptedReader returns each step in order, then (0, io.EOF).
type scriptedReader struct {
	steps []readStep
	i     int
}

type readStep struct {
	data []byte
	err  error
}

func (r *scriptedReader) Read(p []byte) (int, error) {
	if r.i >= len(r.steps) {
		return 0, io.EOF
	}

	step := r.steps[r.i]
	r.i++

	n := copy(p, step.data)

	return n, step.err
}

func TestDeadReadGuard(t *testing.T) {
	t.Run("io.Copy terminates on a reader stuck at (0, nil)", func(t *testing.T) {
		_, err := io.Copy(io.Discard, &deadReadGuard{r: emptyReader{}})
		assert.ErrorIs(t, err, io.ErrNoProgress)
	})

	t.Run("returns ErrNoProgress after maxConsecutiveEmptyReads", func(t *testing.T) {
		g := &deadReadGuard{r: emptyReader{}}
		buf := make([]byte, 8)

		// The first maxConsecutiveEmptyReads-1 calls report (0, nil)...
		for range maxConsecutiveEmptyReads - 1 {
			n, err := g.Read(buf)
			require.Equal(t, 0, n)
			require.NoError(t, err)
		}

		// ...and the next one trips the guard.
		n, err := g.Read(buf)
		assert.Equal(t, 0, n)
		assert.ErrorIs(t, err, io.ErrNoProgress)
	})

	t.Run("real data resets the empty-read counter", func(t *testing.T) {
		// 99 empty reads, then a real read, then empties again: the data read in
		// the middle must reset the counter so the guard does not trip early.
		steps := make([]readStep, 0, maxConsecutiveEmptyReads*2)
		for range maxConsecutiveEmptyReads - 1 {
			steps = append(steps, readStep{})
		}
		steps = append(steps, readStep{data: []byte("hi")})
		for range maxConsecutiveEmptyReads - 1 {
			steps = append(steps, readStep{})
		}

		g := &deadReadGuard{r: &scriptedReader{steps: steps}}
		buf := make([]byte, 8)

		total := 0
		for i := range steps {
			n, err := g.Read(buf)
			require.NoErrorf(t, err, "unexpected error at step %d", i)
			total += n
		}

		// The single non-empty step ("hi") must have flowed through, and resetting
		// the counter on it kept the guard from tripping across the surrounding empties.
		assert.Equal(t, len("hi"), total)
	})

	t.Run("EOF passes straight through", func(t *testing.T) {
		g := &deadReadGuard{r: &scriptedReader{steps: []readStep{{err: io.EOF}}}}
		n, err := g.Read(make([]byte, 8))
		assert.Equal(t, 0, n)
		assert.ErrorIs(t, err, io.EOF)
	})

	t.Run("real errors pass straight through", func(t *testing.T) {
		g := &deadReadGuard{r: &scriptedReader{steps: []readStep{{err: io.ErrClosedPipe}}}}
		n, err := g.Read(make([]byte, 8))
		assert.Equal(t, 0, n)
		assert.ErrorIs(t, err, io.ErrClosedPipe)
	})
}
