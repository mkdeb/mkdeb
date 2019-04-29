package progress

import (
	"io"
)

// Reader is a progression reader.
type Reader struct {
	io.Reader
	callback func(uint64)
	length   uint64
}

// New creates a new progression reader instance.
func New(r io.Reader, callback func(uint64)) *Reader {
	return &Reader{
		Reader:   r,
		callback: callback,
	}
}

// Read satisfies the io.Reader interface.
func (r *Reader) Read(p []byte) (int, error) {
	n, err := r.Reader.Read(p)
	r.length += uint64(n)

	if err == nil || err == io.EOF {
		r.callback(r.length)
	}

	return n, err
}
