package mocks

import (
	"errors"
	"io"
)

// result captures a predefined read/write result.
type result struct {
	n   int   // number of bytes to return on Read/Write
	err error // error to return on Read/Write
}

// ReadCloser is a mock implementation of io.ReadCloser that behaves based on a predefined
// pipeline of read results. It can also mimic reading actual content when provided.
type ReadCloser struct {
	pipeline []result     // a list of predefined read results
	content  []byte       // actual content to read if provided
	close    func() error // custom close function to simulate closing behavior
	closed   bool         // indicates if the Reader has been closed
}

// NewReader initializes and returns a new Reader without any predefined behavior.
func NewReader() *ReadCloser {
	return &ReadCloser{
		pipeline: []result{},
		content:  nil,
		closed:   false,
		close: func() error {
			return nil
		},
	}
}

// WithContent sets the provided content for the Reader to return on its next Read.
// It will override any predefined pipeline.
func (r *ReadCloser) WithContent(content []byte) *ReadCloser {
	r.content = content
	r.pipeline = []result{}

	return r
}

// OnRead appends a predefined read result to the pipeline. The result will be used
// in the order added when the Reader's Read method is called.
func (r *ReadCloser) OnRead(n int, err error) *ReadCloser {
	r.pipeline = append(r.pipeline, result{n: n, err: err})

	return r
}

// OnClose sets a custom error return for the Close method.
func (r *ReadCloser) OnClose(err error) *ReadCloser {
	r.close = func() error {
		r.closed = true

		return err
	}

	return r
}

// EOF appends an EOF result to the pipeline, signaling the end of the data stream.
func (r *ReadCloser) EOF() *ReadCloser {
	r.pipeline = append(r.pipeline, result{n: 0, err: io.EOF})

	return r
}

// Read reads data based on the predefined pipeline. If actual content is provided,
// it will read from that content instead of the pipeline.
func (r *ReadCloser) Read(p []byte) (int, error) {
	if r.closed {
		return 0, errors.New("reader is closed")
	}

	if r.content != nil {
		copy(p, r.content)
		n := len(r.content)
		r.content = nil

		return n, nil
	}

	if len(r.pipeline) == 0 {
		return 0, nil
	}

	resp := r.pipeline[0]
	r.pipeline = r.pipeline[1:]

	return resp.n, resp.err
}

// Close closes the Reader, invoking any predefined closing behavior.
func (r *ReadCloser) Close() error {
	return r.close()
}

// WriteCloser is a mock implementation of io.WriteCloser that behaves based on a predefined
// pipeline of write results.
type WriteCloser struct {
	pipeline []result     // a list of predefined write results
	content  []byte       // actual content that was written
	close    func() error // custom close function to simulate closing behavior
	closed   bool         // indicates if the WriteCloser has been closed
}

// NewWriteCloser initializes and returns a new Writer without any predefined behavior.
// It can also mimic writing actual content when provided.
func NewWriteCloser() *WriteCloser {
	return &WriteCloser{
		pipeline: []result{},
		content:  nil,
		closed:   false,
		close: func() error {
			return nil
		},
	}
}

// WithContent sets the provided content for the Writer.
func (wc *WriteCloser) WithContent(content []byte) *WriteCloser {
	wc.content = content
	wc.pipeline = []result{}

	return wc
}

// OnWrite appends a predefined write result to the pipeline. The result will be used
// in the order added when the Writer's Write method is called.
func (wc *WriteCloser) OnWrite(n int, err error) *WriteCloser {
	wc.pipeline = append(wc.pipeline, result{n: n, err: err})

	return wc
}

// OnClose sets a custom error return for the Close method.
func (wc *WriteCloser) OnClose(err error) *WriteCloser {
	wc.close = func() error {
		wc.closed = true

		return err
	}

	return wc
}

// Write writes data based on the predefined pipeline. It also stores the actual written content.
func (wc *WriteCloser) Write(p []byte) (int, error) {
	if wc.closed {
		return 0, errors.New("writecloser is closed")
	}

	if wc.content != nil {
		length := len(wc.content)
		wc.content = append(wc.content, p[:length]...)

		return length, nil
	}

	if len(wc.pipeline) == 0 {
		return 0, nil
	}

	resp := wc.pipeline[0]
	wc.pipeline = wc.pipeline[1:]

	return resp.n, resp.err
}

// Close closes the WriteCloser, invoking any predefined closing behavior.
func (wc *WriteCloser) Close() error {
	return wc.close()
}
