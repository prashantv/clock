package clock

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClockNow(t *testing.T) {
	t.Parallel()

	c := New()
	now := c.Now()

	assert.False(t, now.IsZero(),
		"expected non-zero time, got: %v", now)
}
