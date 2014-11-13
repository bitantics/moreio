package meters

import (
	"io"
)

// ReadMeter wraps an io.Reader to count the total bytes read
type ReadMeter struct {
	reader  io.Reader
	reading int
}

// NewReadMeter simply wraps the given reader and returns a new ReadMeter
func NewReadMeter(r io.Reader) *ReadMeter {
	return &ReadMeter{reader: r}
}

// Read from the underlying io.Reader
func (rm *ReadMeter) Read(p []byte) (n int, err error) {
	n, err = rm.reader.Read(p)
	rm.reading += n
	return
}

// Reading is the count of total bytes read thus far
func (rm *ReadMeter) Reading() int {
	return rm.reading
}

// WriteMeter wraps an io.Writer to count the total bytes written
type WriteMeter struct {
	writer  io.Writer
	reading int
}

// NewWriteMeter simply wraps the given writer and returns a new WriteMeter
func NewWriteMeter(w io.Writer) *WriteMeter {
	return &WriteMeter{writer: w}
}

// Write to the underlying io.Writer
func (wm *WriteMeter) Write(p []byte) (n int, err error) {
	n, err = wm.writer.Write(p)
	wm.reading += n
	return
}

// Reading is the count of total bytes written thus far
func (wm *WriteMeter) Reading() int {
	return wm.reading
}
