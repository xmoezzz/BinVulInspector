package task

import (
	"sync"
)

type Limiters struct {
	sync.Map
}

func (c *Limiters) AddLimiter(key string, capacity int) {
	c.Store(key, NewLimiter(capacity))
}

func (c *Limiters) Add(lang string) bool {
	if v, ok := c.Load(lang); ok {
		return v.(*Limiter).Add(1)
	}
	return true
}

func (c *Limiters) Done(lang string) {
	if v, ok := c.Load(lang); ok {
		v.(*Limiter).Done()
	}
}

func (c *Limiters) Tune(lang string, capacity int) {
	if v, ok := c.Load(lang); ok {
		v.(*Limiter).Tune(capacity)
	}
}
