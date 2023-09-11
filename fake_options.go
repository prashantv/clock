package clock

import "time"

type fakeOptions struct {
	scheduleAwaken func(time.Time)
}

func (o *fakeOptions) setDefaults() {
	if o.scheduleAwaken == nil {
		o.scheduleAwaken = func(t time.Time) {
			time.Sleep(time.Millisecond)
		}
	}
}

// FakeOption allows customizing `Fake`.
type FakeOption interface {
	apply(*fakeOptions)
}

type fakeOptionRunBackground struct {
	f func(time.Time)
}

func (fo fakeOptionRunBackground) apply(opts *fakeOptions) {
	opts.scheduleAwaken = fo.f
}

// WithFakeRunBackground is used to specify a custom function to run when advancing the clock
// and a background worker is woken up.
func WithFakeRunBackground(f func(time.Time)) FakeOption {
	return fakeOptionRunBackground{f}
}
