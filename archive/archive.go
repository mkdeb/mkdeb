package archive

const (
	_ = iota
	// CompressNone represents a null compression format.
	CompressNone
	// CompressBzip2 represents a Bzip2 compression format.
	CompressBzip2
	// CompressGzip represents a Gzip compression format.
	CompressGzip
	// CompressXZ represents a XZ compression format.
	CompressXZ
)
