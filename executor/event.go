package executor

import (
	"sync"
)

// Disposer dispose something
type Disposer interface {
	Dispose()
}

type funcc struct {
	fn interface{}
}

type eventSource struct {
	listeners []*funcc
	lock      *sync.Mutex
}

type listenerDisposer struct {
	es *eventSource
	fn *funcc
}

func (ld *listenerDisposer) Dispose() {
	ld.es.off(ld.fn)
}

func (e *eventSource) on(fn interface{}) Disposer {
	e.lock.Lock()
	defer e.lock.Unlock()
	fc := &funcc{fn}
	e.listeners = append(e.listeners, fc)
	return &listenerDisposer{e, fc}
}

func (e *eventSource) off(fn *funcc) {
	e.lock.Lock()
	defer e.lock.Unlock()

	ls := make([]*funcc, 0)
	for _, v := range e.listeners {
		if v != fn {
			ls = append(ls, v)
		}
	}
	e.listeners = ls
}

func (e *eventSource) fire(fn func(fn interface{})) {
	ls := e.listeners[:]
	for _, v := range ls {
		fn(v.fn)
	}
}

func newEventSource() *eventSource {
	return &eventSource{nil, new(sync.Mutex)}
}
