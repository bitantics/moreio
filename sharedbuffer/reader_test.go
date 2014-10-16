package sharedbuffer

import (
	"container/heap"
	. "github.com/smartystreets/goconvey/convey"
	"math/rand"
	"testing"
)

func TestReadersHeap(t *testing.T) {
	Convey("Given a SharedBuffer readers heap", t, func() {
		h := make(readers, 0)

		Convey("Push a few readers onto the heap", func() {
			for i := 0; i < 10; i += 1 {
				heap.Push(&h, &reader{at: rand.Intn(100)})
			}

			Convey("They should pop out, first to last", func() {
				prev := -1
				for h.Len() > 0 {
					next := heap.Pop(&h).(*reader).at
					So(next, ShouldBeGreaterThanOrEqualTo, prev)
					prev = next
				}
			})
		})
	})
}
