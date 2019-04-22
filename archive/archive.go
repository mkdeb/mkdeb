package archive

const (
	_ = iota
	// CompressNone is a null compression format.
	CompressNone
	// CompressBzip2 is a Bzip2 compression format.
	CompressBzip2
	// CompressGzip is a Gzip compression format.
	CompressGzip
	// CompressXZ is a XZ compression format.
	CompressXZ
)
