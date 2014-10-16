/*
rollingreader implements a concatenation of io.Readers, which may be included at any time.

To use a RollingReader, simply create one with New(). An arbitrary number of io.Readers
may be put in at initialization or any later time. This differs io.MultiReader, which
only supports including readers at initialization.

A RollingReader will only return EOF after it has been closed and all readers have been
consumed.
*/
package rollingreader

import (
	"errors"
	"io"
)

// RollingReader stores its included io.Readers
type RollingReader struct {
	readers   []io.Reader
	newReader chan struct{}
}

var ErrClosedReader = errors.New("closed reader")

// closer is a nonexistent reader which acts as a delimiter
var closer *RollingReader

// New creates a RollingReader given any number of starting readers.
// They are read from in order.
func New(readers ...io.Reader) *RollingReader {
	rr := &RollingReader{
		readers:   make([]io.Reader, 0),
		newReader: make(chan struct{}),
	}

	if len(readers) > 0 {
		rr.Add(io.MultiReader(readers...))
	}
	return rr
}

// Add a reader to be consumed last
func (rr *RollingReader) Add(r io.Reader) error {
	// Can't add to a closed RollingReader
	readersN := len(rr.readers)
	if readersN > 0 && rr.readers[readersN-1] == closer {
		return ErrClosedReader
	}

	rr.readers = append(rr.readers, r)

	// Broadcast (non-blocking) new reader
	select {
	case rr.newReader <- struct{}{}:
	default:
	}

	return nil
}

// Close the RollingReader, preventing any further reader additions
func (rr *RollingReader) Close() error {
	rr.Add(closer)
	return nil
}

// Read will return any available data from the concatenated readers.
// Blocks until any data is available or an error occurs.
func (rr *RollingReader) Read(p []byte) (n int, err error) {
	n, err = 0, nil

	if len(rr.readers) == 0 {
		<-rr.newReader
	}

	reader := rr.readers[0]
	if reader == closer {
		return 0, io.EOF
	}

	if n, err = reader.Read(p); err == io.EOF {
		rr.readers = rr.readers[1:]
		err = nil
	}

	return
}
