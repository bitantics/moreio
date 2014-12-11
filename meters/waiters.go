package meters

import (
	"container/heap"
)

// waiter has a channel to unblock when its containing meter's reading reaches
// the specified value
type waiter struct {
	reading int64
	wait    chan struct{}
}

// waiters are heapified, enabling constant lookup of the waiter with the
// lowest reading. Meters need to know the waiters with the lowest readings
// everytime their values change. Only those waiters may be unblocked.
type waiters []waiter

// add a single waiter with the given reading to wait for
func (ws *waiters) add(reading int64) <-chan struct{} {
	// use a channel of size 1 so we definitely don't block while resuming
	// and can avoid using a goroutine
	ch := make(chan struct{}, 1)
	heap.Push(ws, waiter{reading, ch})
	return ch
}

// resume waiters with at most the given reading
func (ws *waiters) resume(reading int64) {
	if len(*ws) == 0 {
		return
	}

	// unblock resumed waiters, but don't modify the heap if we can help it
	for lw := (*ws)[0]; lw.reading <= reading; lw = (*ws)[0] {
		heap.Pop(ws)
		lw.wait <- struct{}{}
		close(lw.wait)

		if len(*ws) == 0 {
			break
		}
	}
}

// haltAll currently blocked waiters by closing their channels
func (ws *waiters) haltAll() {
	for _, w := range *ws {
		close(w.wait)
	}
	(*ws) = nil
}

// implement standard heap interface
func (ws waiters) Len() int           { return len(ws) }
func (ws waiters) Less(i, j int) bool { return ws[i].reading < ws[j].reading }

func (ws waiters) Swap(i, j int) {
	ws[i].reading, ws[j].reading = ws[j].reading, ws[i].reading
	ws[i].wait, ws[j].wait = ws[j].wait, ws[i].wait
}

func (ws *waiters) Push(x interface{}) {
	*ws = append(*ws, x.(waiter))
}

func (ws *waiters) Pop() interface{} {
	h, l := *ws, len(*ws)
	w := h[l-1]
	*ws = h[:l-1]
	return w
}
