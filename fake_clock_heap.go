package clock

import (
	"container/heap"
	"sync"
	"time"
)

var _ core = (*FakeHeap)(nil)

// FakeHeap is an implementation of Clock intended for testing backed by a heap.
type FakeHeap struct {
	mu sync.Mutex

	cur     time.Time
	waiters waiterHeap
	Clock   Clock
}

// NewFakeHeap returns a Clock that can be controlled by the Fake.
func NewFakeHeap() *FakeHeap {
	f := &FakeHeap{
		cur: time.Date(2000, 1, 2, 3, 4, 5, 6, time.UTC),
	}
	f.Clock = Clock{f}
	return f
}

type fakeTicker2 struct {
	f *FakeHeap
	w *waiter
}

func (ft *fakeTicker2) Stop() {
	ft.f.removeWaiter(ft.w)
}

func (ft *fakeTicker2) Reset(d time.Duration) {
	ft.f.mu.Lock()
	defer ft.f.mu.Unlock()

	ft.f.removeWaiterLocked(ft.w)
	ft.w.when = ft.f.cur.Add(d)
	ft.w.period = d
	ft.f.addWaiterLocked(ft.w)
}

type fakeTimer2 struct {
	f *FakeHeap
	w *waiter
}

func (ft *fakeTimer2) Stop() bool {
	return ft.f.removeWaiter(ft.w)
}

func (ft *fakeTimer2) Reset(d time.Duration) bool {
	ft.f.mu.Lock()
	defer ft.f.mu.Unlock()

	removed := ft.f.removeWaiterLocked(ft.w)
	ft.w.when = ft.f.cur.Add(d)
	ft.w.period = d
	ft.f.addWaiterLocked(ft.w)
	return removed
}

// Ticker returns a new Ticker containing a channel that sends the current mock time.
// The ticker will drop tickets to make up for slow receivers.
func (f *FakeHeap) ticker(d time.Duration) *Ticker {
	if d <= 0 {
		panic("non-positive interval for NewTicker")
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	c := make(chan time.Time, 1) // buffer matches time.Ticker.
	w := &waiter{
		when:   f.cur.Add(d),
		period: d,
		c:      c,
	}
	f.addWaiterLocked(w)

	return &Ticker{
		C:    c,
		impl: &fakeTicker2{f, w},
	}
}

func (f *FakeHeap) timer(d time.Duration) *Timer {
	f.mu.Lock()
	defer f.mu.Unlock()

	c := make(chan time.Time, 1) // buffer matches time.Ticker.
	w := &waiter{
		when: f.cur.Add(d),
		c:    c,
	}
	f.addWaiterLocked(w)

	return &Timer{
		C:    c,
		impl: &fakeTimer2{f, w},
	}
}

func (f *FakeHeap) afterFunc(d time.Duration, fn func()) *Timer {
	f.mu.Lock()
	defer f.mu.Unlock()

	w := &waiter{
		when: f.cur.Add(d),
		fn:   fn,
	}
	f.addWaiterLocked(w)

	return &Timer{
		impl: &fakeTimer2{f, w},
	}
}

func (f *FakeHeap) now() time.Time {
	f.mu.Lock()
	defer f.mu.Unlock()

	return f.cur
}

func (f *FakeHeap) sleep(d time.Duration) {
	<-f.timer(d).C
}

// Add updates the time.
func (m *FakeHeap) Add(d time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	endTime := m.cur.Add(d)

	for len(m.waiters) > 0 {
		w := m.waiters[0]
		if w.when.After(endTime) {
			break
		}

		i := heap.Pop(&m.waiters)
		if i.(*waiter) != w {
			panic("wtf")
		}

		m.cur = w.when
		m.processWaiterLocked(w, endTime)
	}

	m.cur = endTime
}

func (m *FakeHeap) addWaiterLocked(w *waiter) {
	heap.Push(&m.waiters, w)
}

func (m *FakeHeap) removeWaiter(w *waiter) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.removeWaiterLocked(w)
}

func (m *FakeHeap) removeWaiterLocked(w *waiter) bool {
	// Already removed.
	if w.selfIdx == -1 {
		return false
	}

	heap.Remove(&m.waiters, w.selfIdx)
	return true
}

func (m *FakeHeap) processWaiterLocked(w *waiter, endTime time.Time) {
	m.mu.Unlock()
	if w.c != nil {
		select {
		case w.c <- w.when:
		default:
			// Tickers drop ticks with slow receivers.
		}
	}
	if w.fn != nil {
		go w.fn()
	}

	// best-effort run background stuff
	// time.Sleep(time.Millisecond)
	m.mu.Lock()

	if w.period > 0 {
		w.when = m.cur.Add(w.period)
		m.addWaiterLocked(w)
	}
}

type waiterHeap []*waiter

func (wh waiterHeap) Len() int {
	return len(wh)
}
func (wh waiterHeap) Less(i, j int) bool {
	return wh[i].when.Before(wh[j].when)
}
func (wh waiterHeap) Swap(i, j int) {
	wh[i], wh[j] = wh[j], wh[i]
	wh[i].selfIdx = i
	wh[j].selfIdx = j
}

func (wh *waiterHeap) Push(x any) {
	n := len(*wh)
	item := x.(*waiter)
	item.selfIdx = n
	*wh = append(*wh, item)
}

func (wh *waiterHeap) Pop() any {
	old := *wh
	n := len(old)
	item := old[n-1]
	old[n-1] = nil    // avoid memory leak
	item.selfIdx = -1 // for safety
	*wh = old[:n-1]
	return item
}
