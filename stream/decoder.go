package stream

import (
	"encoding/binary"
	"io"
)

type Decoder struct {
	s      io.Reader
	size   int
	chLeft int
	last   bool
}

// NewDecoder given an encoded stream and chunk size
func NewDecoder(s io.Reader, size int) *Decoder {
	return &Decoder{
		s:      s,
		size:   size,
		chLeft: 0,
		last:   false,
	}
}

// Read and decode bytes from the encoded stream
func (d *Decoder) Read(p []byte) (n int, err error) {
	// Decode and read needed full chunks
	for d.chLeft < len(p)-n {
		m, err := d.s.Read(p[n : n+d.chLeft])
		n += m
		d.chLeft -= m
		if err != nil {
			return n, err
		}

		if d.chLeft == 0 && d.last {
			return n, io.EOF
		}

		if d.chLeft == 0 {
			if d.chLeft, err = d.decodeSize(); d.chLeft < d.size {
				d.last = true
			}
		}
		if err != nil {
			return n, err
		}
	}

	// Read portion of current decoded chunk
	m, err := d.s.Read(p[n:])
	n += m
	d.chLeft -= m
	return
}

// decodeSize of the next chunk by reading in an encoded unsigned integer. It
// may return an error if the underlying stream errors out.
func (d *Decoder) decodeSize() (size int, err error) {
	szb := make([]byte, binary.Size(uint64(d.size)))
	if _, err = d.s.Read(szb); err != nil {
		return -1, err
	}
	sz, _ := binary.Uvarint(szb)
	return int(sz), nil
}
