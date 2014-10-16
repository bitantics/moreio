package sharedbuffer

import (
	"crypto/rand"
	. "github.com/smartystreets/goconvey/convey"
	"io"
	"testing"
	"time"
)

var TEST_BUFFER_SIZE = 1024

func TestSharedBufferIntegrity(t *testing.T) {
	in := make([]byte, TEST_BUFFER_SIZE)
	rand.Read(in)

	Convey("Given a filled SharedBuffer", t, func() {
		sb := New()
		sb.Write(in)

		Convey("Given a single reader", func() {
			r := sb.NewReader()

			Convey("The same data should be read", func() {
				out := make([]byte, TEST_BUFFER_SIZE)
				r.Read(out)
				So(out, ShouldResemble, in)
			})

			Reset(func() {
				r.Close()
			})
		})

		Convey("Given multiple readers", func() {
			rs := make([]io.ReadCloser, 0)
			for i := 0; i < 5; i += 1 {
				rs = append(rs, sb.NewReader())
			}

			Convey("The same data should be read by all", func() {
				for _, r := range rs {
					out := make([]byte, TEST_BUFFER_SIZE)
					r.Read(out)
					So(out, ShouldResemble, in)
				}
			})
		})

		Reset(func() {
			sb.Close()
		})
	})
}

func TestSharedBufferFlushing(t *testing.T) {
	in := make([]byte, TEST_BUFFER_SIZE)
	rand.Read(in)

	Convey("Given a SharedBuffer and two readers", t, func() {
		sb := New()
		r1, r2 := sb.NewReader(), sb.NewReader()
		sb.Write(in)

		Convey("Buffer should not be flushed after one reader is done", func() {
			out := make([]byte, TEST_BUFFER_SIZE)
			r1.Read(out)
			So(len(sb.buf), ShouldEqual, TEST_BUFFER_SIZE)

			Convey("Buffer should be flushed after both readers are done", func() {
				r2.Read(out)
				So(sb.buf, ShouldBeEmpty)
			})
		})
	})
}

func TestSharedBufferBehavior(t *testing.T) {
	in := make([]byte, TEST_BUFFER_SIZE)
	rand.Read(in)

	Convey("Given a SharedBuffer and one reader", t, func() {
		sb := New()
		r1 := sb.NewReader()

		Convey("Reader should block if buffer is empty", func() {
			read := make(chan struct{})
			out := make([]byte, TEST_BUFFER_SIZE)

			go func() {
				r1.Read(out)
				read <- struct{}{}
			}()

			select {
			case <-read:
				So(false, ShouldBeTrue)
			default:
				So(true, ShouldBeTrue)
			}

			Convey("Data should go in with a single Write()", func() {
				n, _ := sb.Write(in)
				So(n, ShouldEqual, len(in))

				Convey("Reader should unblock within 1ms", func() {
					select {
					case <-read:
						So(true, ShouldBeTrue)
					case <-time.After(time.Millisecond):
						So(false, ShouldBeTrue)
					}

					Convey("Reader should return EOF after buffer is empty and closed", func() {
						So(sb.Close(), ShouldBeNil)
						_, err := r1.Read(out)
						So(err, ShouldEqual, io.EOF)
					})
				})
			})
		})
	})
}
