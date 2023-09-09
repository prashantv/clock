package clock

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTicker(t *testing.T) {
	f := NewFake()

	ticker := f.Clock.Ticker(time.Second)

	for i := 0; i < 10; i++ {
		f.Add(999 * time.Millisecond)
		assert.Len(t, ticker.C, 0)

		f.Add(time.Millisecond)

		assert.Len(t, ticker.C, 1)
		<-ticker.C
	}

	// Lost ticks
	for i := 0; i < 5; i++ {
		f.Add(10 * time.Second)
		assert.Len(t, ticker.C, 1)
		<-ticker.C
	}

	// Reset
	ticker.Reset(2 * time.Second)
	for i := 0; i < 10; i++ {
		f.Add(time.Second + 999*time.Millisecond)
		assert.Len(t, ticker.C, 0)

		f.Add(time.Millisecond)

		assert.Len(t, ticker.C, 1)
		<-ticker.C
	}

	// Test double-stopping
	for i := 0; i < 5; i++ {
		ticker.Stop()

		f.Add(10 * time.Second)
		assert.Len(t, ticker.C, 0)
	}

	// Double-stop

	// Reset re-enables the timer.
	ticker.Reset(2 * time.Second)
	for i := 0; i < 10; i++ {
		f.Add(time.Second + 999*time.Millisecond)
		assert.Len(t, ticker.C, 0)

		f.Add(time.Millisecond)

		assert.Len(t, ticker.C, 1)
		<-ticker.C
	}
}
