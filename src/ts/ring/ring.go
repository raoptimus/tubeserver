package ring

import (
	goring "container/ring"
	"sync"
)

type ring struct {
	ring *goring.Ring
	mu   sync.Mutex
}

func New(n int) *ring {
	return &ring{ring: goring.New(n)}
}

func (r *ring) List() []interface{} {
	r.mu.Lock()
	defer r.mu.Unlock()
	var items []interface{}
	r.ring.Do(func(x interface{}) {
		if x != nil {
			items = append(items, x)
		}
	})
	return items
}

func (r *ring) Push(x interface{}) {
	r.mu.Lock()
	r.ring.Value = x
	r.ring = r.ring.Next()
	r.mu.Unlock()
}
