package task

import (
	"sync"
)

type Limiter struct {
	sync.Mutex
	capacity int
	running  int
}

func NewLimiter(capacity int) *Limiter {
	return &Limiter{
		capacity: capacity,
	}
}

func (limiter *Limiter) Tune(capacity int) {
	limiter.Lock()
	defer limiter.Unlock()

	limiter.capacity = capacity
}

func (limiter *Limiter) Add(delta int) bool {
	limiter.Lock()
	defer limiter.Unlock()

	target := limiter.running + delta
	if target > limiter.capacity {
		return false
	}
	limiter.running = target

	return true
}

func (limiter *Limiter) Done() {
	limiter.Lock()
	defer limiter.Unlock()

	limiter.running -= 1
}

func (limiter *Limiter) Available() bool {
	limiter.Lock()
	defer limiter.Unlock()

	return limiter.running < limiter.capacity
}

func (limiter *Limiter) DoneAndIsEmpty() bool {
	limiter.Lock()
	defer limiter.Unlock()

	limiter.running -= 1
	return limiter.running == 0
}
