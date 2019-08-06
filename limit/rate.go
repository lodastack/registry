package limit

import (
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// Create a custom visitor struct which holds the rate limiter for each
// visitor and the last time that the visitor was seen.
type visitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

type RateLimiter struct {
	Visitors map[string]*visitor
	r        rate.Limit
	b        int
	mtx      sync.RWMutex

	timeout time.Duration
}

// NewRateLimiter returns a rate limiter
func NewRateLimiter(r rate.Limit, b int) *RateLimiter {
	rl := RateLimiter{
		Visitors: make(map[string]*visitor),
		r:        r,
		b:        b,
		timeout:  5 * time.Minute,
	}
	rl.Clean()
	return &rl
}

// Clean Run a background goroutine to remove old entries from the visitors map.
func (r *RateLimiter) Clean() {
	go r.cleanupVisitors()
}

func (r *RateLimiter) addVisitor(ip string) *rate.Limiter {
	limiter := rate.NewLimiter(r.r, r.b)
	r.mtx.Lock()
	// Include the current time when creating a new visitor.
	r.Visitors[ip] = &visitor{limiter, time.Now()}
	r.mtx.Unlock()
	return limiter
}

func (r *RateLimiter) GetVisitor(ip string) *rate.Limiter {
	r.mtx.RLock()
	v, exists := r.Visitors[ip]
	if !exists {
		r.mtx.RUnlock()
		return r.addVisitor(ip)
	}

	// Update the last seen time for the visitor.
	v.lastSeen = time.Now()
	r.mtx.RUnlock()
	return v.limiter
}

// Every minute check the map for visitors that haven't been seen for
// more than 3 minutes and delete the entries.
func (r *RateLimiter) cleanupVisitors() {
	for {
		time.Sleep(time.Minute)
		r.mtx.Lock()
		for ip, v := range r.Visitors {
			if time.Now().Sub(v.lastSeen) > r.timeout {
				delete(r.Visitors, ip)
			}
		}
		r.mtx.Unlock()
	}
}
