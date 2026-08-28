package clock

import "time"

type Clock interface {
	Now() time.Time
}

type FixedClock struct {
	value time.Time
}

func NewFixed(value time.Time) *FixedClock {
	return &FixedClock{value: value.UTC()}
}

func (c *FixedClock) Now() time.Time {
	if c == nil {
		return time.Unix(0, 0).UTC()
	}
	return c.value
}

func (c *FixedClock) Advance(delta time.Duration) {
	if c != nil {
		c.value = c.value.Add(delta)
	}
}

func (c *FixedClock) Set(value time.Time) {
	if c != nil {
		c.value = value.UTC()
	}
}

func Unix(value int64) time.Time {
	return time.Unix(value, 0).UTC()
}
