package clock

import "time"

type core interface {
	now() time.Time
	sleep(d time.Duration)
	ticker(d time.Duration) *Ticker
	timer(d time.Duration) *Timer
	afterFunc(d time.Duration, f func()) *Timer
}

type Ticker struct {
	C <-chan time.Time

	impl interface {
		Reset(d time.Duration)
		Stop()
	}
}

func (t *Ticker) Reset(d time.Duration) {
	t.impl.Reset(d)
}

func (t *Ticker) Stop() {
	t.impl.Stop()
}

type Timer struct {
	C <-chan time.Time

	impl interface {
		Reset(d time.Duration) bool
		Stop() bool
	}
}

func (t *Timer) Reset(d time.Duration) bool {
	return t.impl.Reset(d)
}

func (t *Timer) Stop() bool {
	return t.impl.Stop()
}
