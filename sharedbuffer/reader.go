/*
SharedBuffer reader. What to know when creating such a reader:

 1. Timeless access to all present and future buffer data
 2. Close() must be called when it's done
 3. Will return a EOF after the buffer is closed and all data has been read
*/
package sharedbuffer

import (
	"container/heap"
	"errors"
	"io"
)

// reader represents a consumer of a SharedBuffer
type reader struct {
	at  int
	idx int
	sb  *SharedBuffer
}

var ErrClosedReader = errors.New("closed reader")

// Read some data from the buffer. Will block until data is available or an error occurs
func (r *reader) Read(p []byte) (n int, err error) {
	if r.idx < 0 || r.sb == nil {
		return 0, ErrClosedReader
	}

	r.sb.lock.Lock()
	defer r.sb.lock.Unlock()
	n, err = 0, nil

	// Block until available data or error
	for !r.availableData() && !r.sb.closed {
		r.sb.lock.Unlock()
		<-r.sb.newData
		r.sb.lock.Lock()
	}
	if !r.availableData() && r.sb.closed {
		return 0, io.EOF
	}

	// Copy data and move the reader's position in the buffer
	readStart := r.at - r.sb.start
	r.at += copy(p, r.sb.buf[readStart:])

	// Tell SharedBuffer to resort its readers
	heap.Fix(&r.sb.readers, r.idx)
	r.sb.flush()

	return
}

// Close the reader. Tells the SharedBuffer to forget about its data guarantees.
func (r *reader) Close() error {
	r.sb.lock.Lock()
	defer r.sb.lock.Unlock()

	// SharedBuffer can now forget about tracking this reader
	heap.Remove(&r.sb.readers, r.idx)

	r.at = 0
	r.idx = -1
	r.sb = nil

	return nil
}

// availableData returns true if the buffer has new data after the reader's
// current position
func (r reader) availableData() bool {
	return r.at < r.sb.start+len(r.sb.buf)
}

/*
readers is a heap enabled collection of reader instances.

A SharedBuffer effectively stores what data its slowest reader
has yet to consume. As the slowest reader can change after any
read, it is good to have an efficient lookup for the slowest
reader.

A heap of readers reduces a naive O(n) lookup time to O(log n),
where n is the number of active readers.
*/
type readers []*reader

func (rs readers) Len() int           { return len(rs) }
func (rs readers) Less(i, j int) bool { return rs[i].at < rs[j].at }

func (rs readers) Swap(i, j int) {
	rs[i], rs[j] = rs[j], rs[i]
	rs[i].idx, rs[j].idx = rs[j].idx, rs[i].idx
}

func (rs *readers) Push(x interface{}) {
	r := x.(*reader)
	*rs = append(*rs, r)
	r.idx = len(*rs) - 1
}

func (rs *readers) Pop() interface{} {
	h := *rs
	l := len(h)
	p := h[l-1]
	*rs = h[:l-1]

	p.idx = -1
	return p
}
