/*
meters provide the ability to track usage of io.Writers and io.Readers.

Meters wrap the Writer/Reader by implementing the same interface. Each
meter is also an io.Closer, even if the wrapped struct isn't. meter.Close()
will halt all waiters and call the wrapped Close() method if it exists.
*/
package meters

import (
	"io"
)

// ReadMeter wraps an io.Reader to count the total bytes read
type ReadMeter struct {
	reader  io.Reader
	reading int64
	waiters waiters

	eofWaiter chan struct{}
	eof       bool
}

// NewReadMeter simply wraps the given reader and returns a new ReadMeter
func NewReadMeter(r io.Reader) *ReadMeter {
	return &ReadMeter{
		reader:    r,
		eofWaiter: make(chan struct{}, 1),
	}
}

// Read from the underlying io.Reader
func (rm *ReadMeter) Read(p []byte) (n int, err error) {
	n, err = rm.reader.Read(p)
	rm.reading += int64(n)
	rm.waiters.resume(rm.reading)

	if err == io.EOF {
		if !rm.eof {
			rm.eofWaiter <- struct{}{}
			rm.eof = true
		}
		rm.waiters.haltAll()
	}
	return
}

// Close the underlying Reader if possible. This also unblocks any
// remaining waiters.
func (rm *ReadMeter) Close() error {
	var err error
	if closer, ok := rm.reader.(io.Closer); ok {
		err = closer.Close()
	}
	rm.waiters.haltAll()
	close(rm.eofWaiter)
	return err
}

// Reading is the count of total bytes read thus far
func (rm *ReadMeter) Reading() int64 {
	return rm.reading
}

// WaitForReading will block until the specified total bytes are read
func (rm *ReadMeter) WaitForReading(reading int64) <-chan struct{} {
	w := rm.waiters.add(reading)
	rm.waiters.resume(rm.reading)
	return w
}

// WaitForEOF, blocking until the reader encounters the io.EOF error
func (rm *ReadMeter) WaitForEOF() <-chan struct{} {
	if !rm.eof {
		return rm.eofWaiter
	}

	w := make(chan struct{}, 1)
	w <- struct{}{}
	return w
}

// WriteMeter wraps an io.Writer to count the total bytes written
type WriteMeter struct {
	writer  io.Writer
	reading int64
	waiters waiters
}

// NewWriteMeter simply wraps the given writer and returns a new WriteMeter
func NewWriteMeter(w io.Writer) *WriteMeter {
	return &WriteMeter{writer: w}
}

// Write to the underlying io.Writer
func (wm *WriteMeter) Write(p []byte) (n int, err error) {
	n, err = wm.writer.Write(p)
	wm.reading += int64(n)
	wm.waiters.resume(wm.reading)
	return
}

// Close the underlying Writer if possible. This also unblocks any
// remaining waiters.
func (wm *WriteMeter) Close() error {
	var err error
	if closer, ok := wm.writer.(io.Closer); ok {
		err = closer.Close()
	}
	wm.waiters.haltAll()
	return err
}

// Reading is the count of total bytes written thus far
func (wm *WriteMeter) Reading() int64 {
	return wm.reading
}

// WaitForReading will block until the specified total bytes are written
func (wm *WriteMeter) WaitForReading(reading int64) <-chan struct{} {
	w := wm.waiters.add(reading)
	wm.waiters.resume(wm.reading)
	return w
}
