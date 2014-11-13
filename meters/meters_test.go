package meters

import (
	"bytes"
	randbytes "crypto/rand"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestMeters(t *testing.T) {
	TEST_SIZE := 13337
	in, out := make([]byte, TEST_SIZE), make([]byte, TEST_SIZE)
	randbytes.Read(in)

	Convey("Given a buffer with attached read and write meters", t, func() {
		var buf bytes.Buffer
		rm, wm := NewReadMeter(&buf), NewWriteMeter(&buf)

		Convey("Write some data to the buffer and then read it out", func() {
			n, _ := wm.Write(in)
			rm.Read(out)

			Convey("Then both meters' reading should equal the input data length", func() {
				So(wm.Reading(), ShouldEqual, n)
				So(wm.Reading(), ShouldEqual, rm.Reading())
			})
		})
	})
}
