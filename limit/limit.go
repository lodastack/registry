package limit

import (
	"sync"
)

type job chan struct{}

type Limit struct {
	sync.WaitGroup
	limit chan struct{}
	Err   chan error
}

func NewLimit(limit int) Limit {
	return Limit{limit: make(job, limit), Err: make(chan error, limit)}
}

func (l *Limit) Take() {
	l.Add(1)
	l.limit <- struct{}{}
}

func (l *Limit) Release() {
	l.Done()
	<-l.limit
}

func (l *Limit) Error(err error) {
	l.Err <- err
}

func (l *Limit) Close() {
	close(l.limit)
	close(l.Err)
}
