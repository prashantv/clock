package clock

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestMultipleTickers(t *testing.T) {
	const numTickers = 10
	tickers := make([]*Ticker, numTickers)

	f := NewFake()
	for i := range tickers {
		tickMS := i + 1
		tickers[i] = f.Clock.Ticker(time.Duration(tickMS) * time.Millisecond)
	}

	for i := 1; i <= 10*numTickers; i++ {
		f.Add(time.Millisecond)

		for j, ticker := range tickers {
			tickMS := j + 1

			if i%tickMS == 0 {
				require.Len(t, ticker.C, 1)
				<-ticker.C
			} else {
				assert.Len(t, ticker.C, 0)
			}
		}
	}

	// Test dropped ticks
	for i := 1; i <= 10; i++ {
		f.Add(time.Millisecond * numTickers)

		for _, ticker := range tickers {
			require.Len(t, ticker.C, 1)
			<-ticker.C
		}
	}
}
