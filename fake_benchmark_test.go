package clock

import (
	"fmt"
	"testing"
	"time"
)

// Don't sleep after scheduling a waiter for benchmarks.
func noopRunBackground(time.Time) {}

func BenchmarkFakeNTimers(b *testing.B) {
	for _, numTimers := range []int{1, 10, 100, 1000} {
		b.Run(fmt.Sprintf("%v timers", numTimers), func(b *testing.B) {
			f := NewFake(WithFakeRunBackground(noopRunBackground))
			timers := make([]*Ticker, 100)
			for i := range timers {
				if i == 0 {
					continue
				}
				timers[i] = f.Clock.Ticker(time.Duration(i) * time.Millisecond)
			}

			for i := 1; i < b.N; i++ {
				f.Add(time.Millisecond)

				for j, t := range timers {
					if j == 0 {
						continue
					}

					if i%j == 0 {
						<-t.C
					}
				}
			}
		})
	}
}
