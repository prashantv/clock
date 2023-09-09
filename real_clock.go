package clock

import "time"

var _ core = realCore{}

type realCore struct{}

var _sharedRealClock = Clock{realCore{}}

func New() Clock {
	return _sharedRealClock
}

func (realCore) now() time.Time {
	return time.Now()
}

func (realCore) sleep(d time.Duration) {
	time.Sleep(d)
}

func (realCore) ticker(d time.Duration) *Ticker {
	ticker := time.NewTicker(d)
	return &Ticker{
		C:    ticker.C,
		impl: ticker,
	}
}

func (realCore) timer(d time.Duration) *Timer {
	timer := time.NewTimer(d)
	return &Timer{
		C:    timer.C,
		impl: timer,
	}
}

func (realCore) afterFunc(d time.Duration, f func()) *Timer {
	timer := time.AfterFunc(d, f)
	return &Timer{impl: timer}
}
