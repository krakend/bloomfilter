package rotate

import (
	"compress/gzip"
	"io"
)

// Compressor type with new writer and reader function interface
type Compressor interface {
	NewWriter(io.Writer) io.WriteCloser
	NewReader(io.Reader) (io.Reader, error)
}

var compressor Compressor = new(Gzip)

// SetCompressor is a compressor setter
func SetCompressor(c Compressor) {
	compressor = c
}

// Gzip type implementing the Compressor interface
type Gzip int

// NewWriter implementation fo Gzip
func (g *Gzip) NewWriter(w io.Writer) io.WriteCloser {
	return gzip.NewWriter(w)
}

// NewReader implementation fo Gzip
func (g *Gzip) NewReader(r io.Reader) (io.Reader, error) {
	return gzip.NewReader(r)
}
