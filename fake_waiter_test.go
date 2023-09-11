package clock

import (
	"testing"
	"time"
)

func TestFakeWaiter(t *testing.T) {
	f := NewFake()

	done := make(chan struct{})
	go func() {
		// time.Sleep(time.Millisecond)

		t := f.Clock.Ticker(200 * time.Millisecond)
		<-t.C
		<-t.C

		close(done)
	}()

	// w := f.WaitFor(OpTicker)
	// assert.Equal(t, 200*time.Millisecond, w.Delay)

	time.Sleep(time.Millisecond)

	f.Add(time.Second)
	<-done
}
