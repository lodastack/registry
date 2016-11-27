package node

import (
	"sync"
)

type job chan struct{}

type Fixed struct {
	sync.WaitGroup
	limit chan struct{}
	Err   chan error
}

func NewFixed(limit int) Fixed {
	return Fixed{limit: make(job, limit), Err: make(chan error, limit)}
}

func (t *Fixed) Take() {
	t.Add(1)
	t.limit <- struct{}{}
}

func (t *Fixed) Release() {
	t.Done()
	<-t.limit
}

func (t *Fixed) Error(err error) {
	t.Err <- err
}

func (t *Fixed) Close() {
	close(t.limit)
	close(t.Err)
}
