package clock

import (
	"context"
	"time"
)

// Clock is an interface to time-related functionality, which can be backed by a fake for testing.
type Clock struct {
	core core
}

// Now returns the current time.
func (c Clock) Now() time.Time {
	return c.core.now()
}

// Sleep pauses the current goroutine for at least the duration d.
// A negative or zero duration causes Sleep to return immediately.
func (c Clock) Sleep(d time.Duration) {
	c.core.sleep(d)
}

// Tick is a convenience wrapper for Ticker providing access to the ticking channel only.
// While Tick is useful for clients that have no need to shut down the Ticker, be aware that without
// a way to shut it down the underlying Ticker cannot be recovered by the garbage collector; it "leaks".
// Unlike Ticker, Tick will return nil if d <= 0.
func (c Clock) Tick(d time.Duration) <-chan time.Time {
	if d <= 0 {
		return nil
	}

	return c.Ticker(d).C
}

// NewTicker returns a new Ticker containing a channel that will send the current time on the channel after each tick. The period of the ticks is specified by the duration argument. The ticker will adjust the time interval or drop ticks to make up for slow receivers. The duration d must be greater than zero; if not, NewTicker will panic. Stop the ticker to release associated resources.
func (c Clock) Ticker(d time.Duration) *Ticker {
	return c.core.ticker(d)
}

// After waits for the duration to elapse and then sends the current time on the returned channel.
// It is equivalent to `Timer(d).C`.
// The underlying Timer is not recovered by the garbage collector until the timer fires.
// If efficiency is a concern, use NewTimer instead and call Timer.Stop if the timer is no longer needed.
func (c Clock) After(d time.Duration) <-chan time.Time {
	return c.core.timer(d).C
}

// AfterFunc waits for the duration to elapse and then calls f in its own goroutine. It returns a Timer that can be used to cancel the call using its Stop method.
func (c Clock) AfterFunc(d time.Duration, f func()) *Timer {
	return c.core.afterFunc(d, f)
}

// Timer creates a new Timer that will send the current time on its channel after at least duration d.
func (c Clock) Timer(d time.Duration) *Timer {
	return c.core.timer(d)
}

// SleepContext pauses the current goroutine till either the provided context is done
// or until after the duration d, whichever is earlier.
// A negative or zero duration causes SleepContext to return immediately.
func (c Clock) SleepContext(ctx context.Context, d time.Duration) error {
	select {
	case <-c.After(d):
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
