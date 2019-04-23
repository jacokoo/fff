package executor

import (
	"errors"
	"sync"
)

type item struct {
	task  *task
	state *taskState
}

func (i *item) Name() string {
	return i.task.name
}

func (i *item) Future() Future {
	return i.state.fu
}

type limittedExecutor struct {
	max           int
	errorWhenFull bool
	ch            chan *item
	quit          chan bool
	running       []*item
	ended         bool
	lock          *sync.RWMutex
}

func (e *limittedExecutor) worker(i int) {
	for {
		select {
		case item := <-e.ch:
			if !item.state.IsCancelled() {
				e.running[i] = item
				data, err := item.task.Do(item.state)
				item.state.fu.set(data, err)
				e.running[i] = nil
			}
		case <-e.quit:
			return
		}
	}
}

func (e *limittedExecutor) init() {
	for i := 0; i < e.max; i++ {
		go e.worker(i)
	}
}

// Fixed create a limited executor
func Fixed(max uint, errorWhenFull bool) Executor {
	e := &limittedExecutor{
		int(max), errorWhenFull,
		make(chan *item), make(chan bool), make([]*item, int(max)),
		false, new(sync.RWMutex),
	}
	e.init()
	return e
}

func (e *limittedExecutor) Submit(name string, fn TaskFunc) Future {
	e.lock.RLock()
	if e.ended {
		return ErrorFuture(errors.New("Executor is already closed"))
	}
	e.lock.RUnlock()

	ts := newTaskState()
	item := &item{&task{name, fn}, ts}
	if !e.errorWhenFull {
		e.ch <- item
		return ts.fu
	}

	select {
	case e.ch <- item:
		return ts.fu
	default:
		return ErrorFuture(errors.New("task is full"))
	}
}

func (e *limittedExecutor) Running() []Task {
	re := make([]Task, 0)
	for _, v := range e.running {
		if v != nil {
			re = append(re, v)
		}
	}
	return re
}

func (e *limittedExecutor) Close() error {
	e.lock.Lock()
	defer e.lock.Unlock()

	if e.ended {
		return errors.New("Executor is already closed")
	}

	e.ended = true
	for i := 0; i < e.max; i++ {
		e.quit <- true
	}

	for _, v := range e.running {
		if v != nil {
			_ = v.Future().Cancel()
		}
	}
	return nil
}
