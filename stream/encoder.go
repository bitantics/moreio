package stream

import (
	"bytes"
	"encoding/binary"
	"io"
)

type Encoder struct {
	s    io.Writer
	size int
	ch   bytes.Buffer
}

// NewEncoder given a stream to write encoded chunks
// to and the chunk size.
func NewEncoder(s io.Writer, size int) *Encoder {
	var ch bytes.Buffer
	return &Encoder{s, size, ch}
}

// Write data into the Encoder. Written data won't be
// flushed to the stream until enough is present to
// encode a chunk.
func (e *Encoder) Write(p []byte) (n int, err error) {
	if n, err = e.ch.Write(p); err != nil {
		return
	}

	for e.ch.Len() > e.size {
		if _, err = e.encode(e.size); err != nil {
			return
		}
	}
	return
}

// encode a chunk of specified length into the stream
func (e *Encoder) encode(sz int) (n int, err error) {
	szb := make([]byte, binary.Size(uint64(e.size)))
	binary.PutUvarint(szb, uint64(sz))
	e.s.Write(szb)

	m, err := io.CopyN(e.s, &e.ch, int64(sz))
	return int(m), err
}

// Close the Encoder. Flushes any unwritten data to an
// incomplete chunk. This marks the end of the stream.
func (e *Encoder) Close() error {
	_, err := e.encode(e.ch.Len())
	return err
}
