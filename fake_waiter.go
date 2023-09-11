package clock

// Operation is Clock method called.
type Operation string

// Operations on the clock.
// TODO: Track sleeps separate from a timer.
// TODO: What to do about Clock operations that use the Core.
const (
	OpTicker Operation = "ticker"
	OpTimer  Operation = "timer"
)

// WaitForMatcher matches against a specific waiter.
type WaitForMatcher interface {
	Match(*Waiter) bool
}

func (o Operation) Match(w *Waiter) bool {
	return w.Op == o
}

func (f *Fake) WaitFor(m WaitForMatcher) *Waiter {
	f.mu.Lock()
	defer f.mu.Unlock()

	for {
		waiter, ok := f.matchWaiterLocked(m)
		if ok {
			return waiter
		}

		// TODO: Consider a timeout after which we panic.
		f.waitForCond.Wait()
	}
}

func (f *Fake) matchWaiterLocked(m WaitForMatcher) (*Waiter, bool) {
	for _, w := range f.waiters {
		if m.Match(w) {
			return w, true
		}
	}

	return nil, false
}

func (f *Fake) Next() (*Waiter, bool) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if len(f.waiters) == 0 {
		return nil, false
	}

	return f.waiters[0], true
}

// Waiters returns all waiters in order of when they're scheduled to run.
func (f *Fake) Waiters() []*Waiter {
	// TODO IMPLEMENT
	return nil
}
