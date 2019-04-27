package archive

import (
	"archive/tar"
	"compress/bzip2"
	"compress/gzip"
	"io"
	"io/ioutil"

	"github.com/ulikunitz/xz"
)

// Reader is an archive reader.
type Reader struct {
	rc  io.ReadCloser
	tar *tar.Reader
}

// NewReader creates a new archive reader instance given an io.Reader and a compression format.
func NewReader(r io.Reader, compress int) (*Reader, error) {
	var (
		rc  io.ReadCloser
		err error
	)

	switch compress {
	case CompressNone:
		rc = ioutil.NopCloser(r)

	case CompressBzip2:
		rc = ioutil.NopCloser(bzip2.NewReader(r))

	case CompressGzip:
		rc, err = gzip.NewReader(r)
		if err != nil {
			return nil, err
		}

	case CompressXZ:
		r, err := xz.NewReader(r)
		if err != nil {
			return nil, err
		}
		rc = ioutil.NopCloser(r)

	default:
		return nil, ErrUnsupportedCompress
	}

	return &Reader{
		rc:  rc,
		tar: tar.NewReader(rc),
	}, nil
}

// Close closes the archive reader.
func (r *Reader) Close() error {
	// Check whether or not the reader satisfies the io.Closer interface, then close it if yes
	if r.rc != nil {
		r.rc.Close()
	}

	return nil
}

// Next advances to the next entry in the archive (see archive/tar Reader.Next for details).
func (r *Reader) Next() (*Header, error) {
	h, err := r.tar.Next()
	if err != nil {
		return nil, err
	}

	return &Header{
		Name:     h.Name,
		LinkName: h.Linkname,
		Size:     h.Size,
		Mode:     h.FileInfo().Mode(),
		User:     h.Uname,
		Group:    h.Gname,
		ModTime:  h.ModTime,
	}, nil
}

// Read reads from the current file in the archive (see archive/tar Reader.Read for details).
func (r *Reader) Read(b []byte) (int, error) {
	return r.tar.Read(b)
}
