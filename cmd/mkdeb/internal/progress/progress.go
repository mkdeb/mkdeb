package progress

import (
	"io"
)

type Reader struct {
	io.Reader
	callback func(uint64)
	length   uint64
}

func NewReader(r io.Reader, callback func(uint64)) *Reader {
	return &Reader{
		Reader:   r,
		callback: callback,
	}
}

func (r *Reader) Read(p []byte) (int, error) {
	n, err := r.Reader.Read(p)
	r.length += uint64(n)

	if err == nil || err == io.EOF {
		r.callback(r.length)
	}

	return n, err
}
