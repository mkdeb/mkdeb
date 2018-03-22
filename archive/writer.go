package archive

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"io"

	"github.com/ulikunitz/xz"
)

// Writer represents an archive writer instance.
type Writer struct {
	compress io.WriteCloser
	tar      *tar.Writer
}

// NewWriter creates a new archive writer instance given an io.Writer and a compression format.
func NewWriter(w io.Writer, compress int) (*Writer, error) {
	var (
		cw  io.WriteCloser
		err error
	)

	switch compress {
	case CompressGzip:
		cw = gzip.NewWriter(w)

	case CompressXZ:
		cw, err = xz.NewWriter(w)

	default:
		err = ErrUnsupportedCompress
	}

	if err != nil {
		return nil, err
	}

	return &Writer{
		compress: cw,
		tar:      tar.NewWriter(cw),
	}, nil
}

// Close closes the archive writer instance.
func (w *Writer) Close() error {
	if err := w.tar.Close(); err != nil {
		return err
	}

	return w.compress.Close()
}

// Write writes a file in the archive.
func (w *Writer) Write(b []byte) (int, error) {
	return w.tar.Write(b)
}

// WriteHeader writes a given file header to the archive.
func (w *Writer) WriteHeader(h *Header) error {
	return w.tar.WriteHeader(h.TarHeader())
}

// WriterBuffer represents an archive writer buffer instance.
type WriterBuffer struct {
	*Writer

	buffer *bytes.Buffer
}

// NewWriterBuffer creates a new archive writer buffer instance given a compression format.
func NewWriterBuffer(compress int) (*WriterBuffer, error) {
	buf := bytes.NewBuffer(nil)

	w, err := NewWriter(buf, compress)
	if err != nil {
		return nil, err
	}

	return &WriterBuffer{w, buf}, nil
}

// Bytes returns the unread content of the archive writer buffer.
func (w *WriterBuffer) Bytes() []byte {
	return w.buffer.Bytes()
}
