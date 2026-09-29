package ratelimit

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type clock struct{ t time.Time }

func (c *clock) now() time.Time { return c.t }

func newTest(burst int, perMin float64, ttl time.Duration) (*KeyedLimiter, *clock) {
	c := &clock{t: time.Unix(1_000_000, 0)}
	l := NewKeyedLimiter(PerMinute(perMin), burst, ttl)
	l.now = c.now
	return l, c
}

func TestAllow_BurstThenBlockThenRefill(t *testing.T) {
	l, c := newTest(3, 6, time.Hour) // 1 token per 10s

	for i := 0; i < 3; i++ {
		ok, _ := l.Allow("k")
		assert.True(t, ok, "burst request %d", i)
	}
	ok, wait := l.Allow("k")
	assert.False(t, ok)
	assert.InDelta(t, 10, wait.Seconds(), 0.1)

	c.t = c.t.Add(10 * time.Second)
	ok, _ = l.Allow("k")
	assert.True(t, ok, "one token refilled")
	ok, _ = l.Allow("k")
	assert.False(t, ok)
}

func TestAllow_KeysAreIndependent(t *testing.T) {
	l, _ := newTest(1, 1, time.Hour)
	ok, _ := l.Allow("a")
	assert.True(t, ok)
	ok, _ = l.Allow("a")
	assert.False(t, ok)
	ok, _ = l.Allow("b")
	assert.True(t, ok)
}

func TestFailAllowedReset(t *testing.T) {
	l, _ := newTest(2, 1, time.Hour)

	ok, _ := l.Allowed("k")
	assert.True(t, ok)
	ok, _ = l.Allowed("k")
	assert.True(t, ok, "Allowed must not consume")

	l.Fail("k")
	l.Fail("k")
	ok, wait := l.Allowed("k")
	assert.False(t, ok)
	assert.Greater(t, wait, time.Duration(0))

	l.Reset("k")
	ok, _ = l.Allowed("k")
	assert.True(t, ok)
}

func TestSweepEvictsIdleKeys(t *testing.T) {
	l, c := newTest(1, 1, time.Minute)
	for _, k := range []string{"a", "b", "c"} {
		l.Allow(k)
	}
	assert.Equal(t, 3, l.Len())

	c.t = c.t.Add(2 * time.Minute)
	l.Allow("d") // triggers sweep
	assert.Equal(t, 1, l.Len(), "idle keys evicted, new key kept")
}
