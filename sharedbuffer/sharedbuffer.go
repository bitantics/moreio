/*
sharedbuffer provides a buffer which supports concurrent access by multiple readers.

The buffer automatically flushes any data which has been read by all readers.
To create a reader, simply call sb.NewReader(), given a SharedBuffer sb. If a
consumer is done with the buffer, it must signal so by closing its reader. If
a reader is not closed, the buffer will not flush any data past the unused
reader's position!
*/
package sharedbuffer

import (
	"container/heap"
	"errors"
	"io"
	"sync"
)

var ErrClosedBuffer = errors.New("cannot write to closed buffer")

// SharedBuffer represents a concurrently shared buffer
type SharedBuffer struct {
	readers readers
	start   int
	buf     []byte
	closed  bool

	lock    sync.RWMutex
	newData chan struct{}
}

// New creates an initialized SharedBuffer
func New() *SharedBuffer {
	sb := SharedBuffer{
		readers: make(readers, 0),
		buf:     make([]byte, 0),
		closed:  false,
	}
	sb.newData = make(chan struct{})
	return &sb
}

// NewReader creates a registered reader for the buffer. This reader must be
// closed when it is done, lest you hate having free memory.
func (sb *SharedBuffer) NewReader() io.ReadCloser {
	return sb.NewReaderAt(int64(sb.start))
}

// NewReaderAt generates a registered reader which will block until the buffer
// fills to the given offset
func (sb *SharedBuffer) NewReaderAt(off int64) io.ReadCloser {
	sb.lock.Lock()
	defer sb.lock.Unlock()

	r := &reader{
		idx: len(sb.readers),
		sb:  sb,
		at:  int(off),
	}
	sb.readers = append(sb.readers, r)
	heap.Fix(&sb.readers, r.idx)

	return r
}

// Write puts data into the open buffer
func (sb *SharedBuffer) Write(p []byte) (n int, err error) {
	sb.lock.Lock()
	defer sb.lock.Unlock()

	if sb.closed {
		return 0, ErrClosedBuffer
	}

	sb.buf = append(sb.buf, p...)
	sb.signalNewData()

	return len(p), nil
}

// Close the buffer, preventing any further writes. Readers will return io.EOF
// after consuming the remainder.
func (sb *SharedBuffer) Close() error {
	sb.closed = true
	sb.signalNewData()

	return nil
}

// signalNewData unblocks waiting readers, allowing them to handle any new data
func (sb *SharedBuffer) signalNewData() {
	select {
	case sb.newData <- struct{}{}:
	default:
	}
}

// flush any collectively read data
func (sb *SharedBuffer) flush() int {
	slowestReader := sb.readers[0]
	stale := slowestReader.at - sb.start
	sb.buf, sb.start = sb.buf[stale:], slowestReader.at
	return stale
}
