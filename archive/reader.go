package archive

import (
	"archive/tar"
	"compress/bzip2"
	"compress/gzip"
	"io"
	"io/ioutil"

	"github.com/ulikunitz/xz"
)

// Reader represents an archive reader instance.
type Reader struct {
	compress io.ReadCloser
	tar      *tar.Reader
}

// NewReader creates a new archive reader instance given an io.Reader and a compression format.
func NewReader(r io.Reader, compress int) (*Reader, error) {
	var (
		cr  io.ReadCloser
		err error
	)

	switch compress {
	case CompressBzip2:
		cr = ioutil.NopCloser(bzip2.NewReader(r))

	case CompressGzip:
		cr, err = gzip.NewReader(r)
		if err != nil {
			return nil, err
		}

	case CompressXZ:
		r, err := xz.NewReader(r)
		if err != nil {
			return nil, err
		}
		cr = ioutil.NopCloser(r)

	default:
		return nil, ErrUnsupportedCompress
	}

	return &Reader{
		compress: cr,
		tar:      tar.NewReader(cr),
	}, nil
}

// Close closes the archive reader instance.
func (r *Reader) Close() error {
	// Check whether or not the reader satisfies the io.Closer interface, then close it if yes
	if r.compress != nil {
		r.compress.Close()
	}

	return nil
}

// Next advances to the next entry in the archive (see archive/tar Reader.Next for details).
func (r *Reader) Next() (*tar.Header, error) {
	return r.tar.Next()
}

// Read reads from the current file in the archive (see archive/tar Reader.Read for details).
func (r *Reader) Read(b []byte) (int, error) {
	return r.tar.Read(b)
}
