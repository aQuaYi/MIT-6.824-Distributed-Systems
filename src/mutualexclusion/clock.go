package mutualexclusion

import (
	"math/rand"
	"sync"
)

type clock struct {
	time int
	rwmu *sync.RWMutex
}

func newClock() *clock {
	return &clock{
		time: rand.Intn(100),
	}
}

func (c *clock) getTime() int {
	c.rwmu.RLock()
	defer c.rwmu.RUnlock()
	return c.time
}

func (c *clock) update(t int) {
	c.rwmu.Lock()
	c.time = max(c.time, t+1)
	c.rwmu.Unlock()
}

func (c *clock) tick() {
	c.rwmu.Lock()
	c.time++
	c.rwmu.Unlock()
}
