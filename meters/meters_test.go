package meters

import (
	"bytes"
	randbytes "crypto/rand"
	"fmt"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func shouldBlock(actual interface{}, expected ...interface{}) string {
	ch, _ := actual.(<-chan struct{})
	select {
	case <-ch:
		return fmt.Sprintf("Expected '%v' to block", ch)
	default:
		return ""
	}
}

func shouldNotBlock(actual interface{}, expected ...interface{}) string {
	if shouldBlock(actual) == "" {
		return fmt.Sprintf("Expected '%v' to not block", actual)
	}
	return ""
}

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
			Convey("Then waiting for the first byte should not block", func() {
				So(wm.WaitForReading(1), shouldNotBlock)
				So(rm.WaitForReading(1), shouldNotBlock)
			})
			Convey("Then waiting for EOF should still block (buffers don't emit EOF)", func() {
				So(rm.WaitForEOF, shouldBlock)
			})
		})

		Convey("Then waiting for nonexistent data should block", func() {
			So(wm.WaitForReading(1), shouldBlock)
			So(rm.WaitForReading(1), shouldBlock)
			So(rm.WaitForEOF, shouldBlock)
		})
	})
}
