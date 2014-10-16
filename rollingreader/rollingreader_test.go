package rollingreader

import (
	"bytes"
	cryptoRand "crypto/rand"
	. "github.com/smartystreets/goconvey/convey"
	"io/ioutil"
	mathRand "math/rand"
	"testing"
	"time"
)

func randomBytes() []byte {
	buf := make([]byte, mathRand.Intn(100))
	cryptoRand.Read(buf)
	return buf
}

func TestRollingReader(t *testing.T) {
	Convey("Given a RollingReader initialized with one reader", t, func() {
		in := randomBytes()
		rr := New(bytes.NewReader(in))

		Convey("Multiple arbitrarily timed readers' data should pass through correctly", func() {
			// Add delayed readers in the background
			go func() {
				for i := 0; i < 10; i += 1 {
					b := randomBytes()
					in = append(in, b...)
					rr.Add(bytes.NewReader(b))

					delay := time.Duration(mathRand.Intn(50))
					time.Sleep(delay * time.Millisecond)
				}
				rr.Close()
			}()

			// Read until EOF while readers are being added
			out, _ := ioutil.ReadAll(rr)
			So(out, ShouldResemble, in)
		})
	})
}
