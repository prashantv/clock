package clock

import (
	"sort"
	"sync"
	"time"
)

var _ core = (*Fake)(nil)

// Fake is an implementation of Clock intended for testing.
type Fake struct {
	mu sync.Mutex

	cur     time.Time
	waiters waiters
	Clock   Clock
}

type waiter struct {
	selfIdx int // used to remove itself.

	when   time.Time
	period time.Duration

	// Return indicates if the buffer was writtne to.
	c  chan time.Time
	fn func()
}

// NewFake returns a Clock that can be controlled by the Fake.
func NewFake() *Fake {
	f := &Fake{
		cur: time.Date(2000, 1, 2, 3, 4, 5, 6, time.UTC),
	}
	f.Clock = Clock{f}
	return f
}

type fakeTicker struct {
	f *Fake
	w *waiter
}

func (ft *fakeTicker) Stop() {
	ft.f.removeWaiter(ft.w)
}

func (ft *fakeTicker) Reset(d time.Duration) {
	ft.f.mu.Lock()
	defer ft.f.mu.Unlock()

	ft.f.removeWaiterLocked(ft.w)
	ft.w.when = ft.f.cur.Add(d)
	ft.w.period = d
	ft.f.addWaiterLocked(ft.w)
}

type fakeTimer struct {
	f *Fake
	w *waiter
}

func (ft *fakeTimer) Stop() bool {
	return ft.f.removeWaiter(ft.w)
}

func (ft *fakeTimer) Reset(d time.Duration) bool {
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
func (f *Fake) ticker(d time.Duration) *Ticker {
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
		impl: &fakeTicker{f, w},
	}
}

func (f *Fake) timer(d time.Duration) *Timer {
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
		impl: &fakeTimer{f, w},
	}
}

func (f *Fake) afterFunc(d time.Duration, fn func()) *Timer {
	f.mu.Lock()
	defer f.mu.Unlock()

	w := &waiter{
		when: f.cur.Add(d),
		fn:   fn,
	}
	f.addWaiterLocked(w)

	return &Timer{
		impl: &fakeTimer{f, w},
	}
}

func (f *Fake) now() time.Time {
	f.mu.Lock()
	defer f.mu.Unlock()

	return f.cur
}

func (f *Fake) sleep(d time.Duration) {
	<-f.timer(d).C
}

// Add updates the time.
func (m *Fake) Add(d time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	endTime := m.cur.Add(d)

	for len(m.waiters) > 0 {
		w := m.waiters[0]
		if w.when.After(endTime) {
			break
		}

		m.waiters[0] = nil
		m.waiters = m.waiters[1:]

		m.cur = w.when
		m.processWaiterLocked(w, endTime)
	}

	m.cur = endTime
}

func (m *Fake) addWaiterLocked(w *waiter) {
	m.waiters = append(m.waiters, w)
	w.selfIdx = len(m.waiters) - 1
	sort.Sort(m.waiters)
}

func (m *Fake) removeWaiter(w *waiter) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.removeWaiterLocked(w)
}

func (m *Fake) removeWaiterLocked(w *waiter) bool {
	// Already removed.
	if w.selfIdx == -1 {
		return false
	}

	for i := w.selfIdx; i+1 < len(m.waiters); i++ {
		m.waiters[i] = m.waiters[i+1]
		m.waiters[i].selfIdx = i
	}
	m.waiters = m.waiters[:len(m.waiters)-1]

	w.selfIdx = -1
	return true
}

func (m *Fake) processWaiterLocked(w *waiter, endTime time.Time) {
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
	time.Sleep(time.Millisecond)
	m.mu.Lock()

	if w.period > 0 {
		w.when = m.cur.Add(w.period)
		m.addWaiterLocked(w)
	}
}

type waiters []*waiter

func (ws waiters) Len() int {
	return len(ws)
}
func (ws waiters) Less(i, j int) bool {
	return ws[i].when.Before(ws[j].when)
}
func (ws waiters) Swap(i, j int) {
	ws[i], ws[j] = ws[j], ws[i]
	ws[i].selfIdx = i
	ws[j].selfIdx = j
}
