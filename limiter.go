package collyresponsible

import (
	"sync"
	"time"
)

type RequestLimiter struct {
	SleepDelay int
	sleepMin   int
	lock       *sync.RWMutex
}

func (r *RequestLimiter) Increase() {
	r.lock.Lock()
	defer r.lock.Unlock()
	r.SleepDelay++
}

func (r *RequestLimiter) Decrease() {
	r.lock.Lock()
	defer r.lock.Unlock()
	if r.SleepDelay > r.sleepMin {
		r.SleepDelay--
	}
}

func (r *RequestLimiter) Sleep() {
	r.lock.RLock()
	defer r.lock.RUnlock()
	time.Sleep(time.Duration(r.SleepDelay) * time.Second)
}

func NewLimiter(sleepDelay int) *RequestLimiter {
	return &RequestLimiter{
		SleepDelay: sleepDelay,
		sleepMin:   sleepDelay,
		lock:       &sync.RWMutex{},
	}
}
