# Design

## Interface

### Clock as struct vs interface

Most clock libraries make clock an interface, but this limits extensibility. Adding a new method to an interface is a breaking change.

Using a struct provides extensibility, while also allowing additional helper methods to be added without duplicating the implementation
across the real and mock implementation.

For example, `SleepContext` is a helper that sleeps with a context, and rather than duplicating the implementations between
the real and mock implementations, there's a single implementation backed by the `After` implementation.

### Exporting an interface

Though most clock libraries expose an interface, there's only 2 implementations: the real clock, and the fake clock. We don't expect
a need for additional implementations, so the underlying interface behind the Clock is not exported.


### Ticker and Timer as struct vs interface

Using an interface with a `C()` method requires wrapping the returned timer in an additional struct.
When this is converted to an interface, this adds an additional allocation.

This also allows new methods to be added to the Timer/Ticker without requiring a breaking change, to match Go.

## Fake clock

### TBD: Separate package

Is it useful to have the fake in a subpackage vs top-level package?

 * Top-level is simpler, vs separate package requiring some plumbing to hide the interface and clock constructor
 * Separate package avoids mock implementation from being compiled into prod code
 * Separate package is cleaner if we want the mock to take additional dependencies (e.g., on `testing.TB`), without bringing 
   in a `testing` package dependency into prod code.


### (To implement) Detect races

A common issue with tests using fake clocks is that they forget to wait for a ticker/timer before advancing the fake time.
We should be able to take advantage of the race detector to detect these as the ticker/timer will usually happen in
a background goroutine while the test goroutine advances the clock.