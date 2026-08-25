package metrics

import (
	"fmt"
	"sort"
	"strings"
	"sync"
)

type Counter struct {
	mu     sync.RWMutex
	values map[string]int64
}

func New() *Counter { return &Counter{values: make(map[string]int64)} }

func (c *Counter) Inc(name string) {
	c.Add(name, 1)
}

func (c *Counter) Add(name string, amount int64) {
	if c == nil || name == "" {
		return
	}
	c.mu.Lock()
	c.values[name] += amount
	c.mu.Unlock()
}

func (c *Counter) Value(name string) int64 {
	if c == nil {
		return 0
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.values[name]
}

func (c *Counter) Snapshot() map[string]int64 {
	result := make(map[string]int64)
	if c == nil {
		return result
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	for key, value := range c.values {
		result[key] = value
	}
	return result
}

func (c *Counter) Prometheus() string {
	snapshot := c.Snapshot()
	keys := make([]string, 0, len(snapshot))
	for key := range snapshot {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	var builder strings.Builder
	for _, key := range keys {
		builder.WriteString(fmt.Sprintf("share_gateway_%s %d\n", key, snapshot[key]))
	}
	return builder.String()
}
