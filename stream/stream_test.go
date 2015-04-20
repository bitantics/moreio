package stream

import (
	"bytes"
	"crypto/rand"
	. "github.com/smartystreets/goconvey/convey"
	"io"
	"testing"
)

const CHUNK_SIZE = 1000

func TestStream(t *testing.T) {
	type test struct {
		name  string
		input []byte
	}

	tsts := []test{
		test{
			"smaller than a chunk",
			randomBytes(CHUNK_SIZE / 3),
		},
		test{
			"exactly a few chunks",
			randomBytes(CHUNK_SIZE * 8),
		},
		test{
			"little bit more than a few chunks",
			randomBytes(CHUNK_SIZE*8 + CHUNK_SIZE/2),
		},
	}

	for _, tst := range tsts {
		testStream(t, tst.input, tst.name)
	}
}

func testStream(t *testing.T, in []byte, testName string) {
	Convey("Given an input "+testName, t, func() {
		inr := bytes.NewReader(in)
		var s bytes.Buffer

		Convey("When it is encoded into a stream", func() {
			e := NewEncoder(&s, CHUNK_SIZE)
			n, err := io.Copy(e, inr)
			So(err, ShouldBeNil)
			So(n, ShouldEqual, len(in))

			err = e.Close()
			So(err, ShouldBeNil)

			Convey("Then it should be decoded and read out intact", func() {
				var out bytes.Buffer
				d := NewDecoder(&s, CHUNK_SIZE)
				_, err := io.Copy(&out, d)

				So(out.Bytes(), ShouldResemble, in)
				So(err, ShouldBeNil)
			})
		})
	})
}

func randomBytes(n int) []byte {
	buf := make([]byte, n)
	rand.Read(buf)
	return buf
}
