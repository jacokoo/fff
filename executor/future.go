package executor

import (
	"errors"
	"sync"
)

var (
	_ = Future(new(future))

	// ErrorCancelled error to use when cancel task
	ErrorCancelled = errors.New("Task is already cancelled")
)

type futureState uint16

// Future state
const (
	futurePadding futureState = iota
	futureBlocked
	futureCancelled
	futureCompleted
)

// Future get value after complete
type Future interface {
	Get() (interface{}, error)
	Cancel() error
	OnComplete(func(interface{}, error)) Disposer
	OnProgress(func(interface{})) Disposer
}

type future struct {
	completeSource *eventSource
	progressSource *eventSource

	result     interface{}
	error      error
	blocker    chan bool
	blockCount int
	state      futureState
	stateLock  *sync.RWMutex
}

func newFuture() *future {
	return &future{
		completeSource: newEventSource(),
		progressSource: newEventSource(),
		blocker:        make(chan bool),
		stateLock:      new(sync.RWMutex),
	}
}

func (f *future) Get() (interface{}, error) {
	f.stateLock.Lock()

	if f.state == futureCancelled {
		f.stateLock.Unlock()
		return nil, ErrorCancelled
	}

	if f.state == futureCompleted {
		f.stateLock.Unlock()
		return f.result, f.error
	}

	f.blockCount++
	f.state = futureBlocked
	f.stateLock.Unlock()

	<-f.blocker

	f.stateLock.RLock()
	defer f.stateLock.RUnlock()
	if f.state == futureCancelled {
		return nil, ErrorCancelled
	}

	if f.state == futureCompleted {
		return f.result, f.error
	}

	return nil, errors.New("Unknown state")
}

func (f *future) set(result interface{}, err error) {
	f.stateLock.Lock()
	defer f.stateLock.Unlock()

	if f.state != futurePadding && f.state != futureBlocked {
		return
	}
	f.result = result
	f.error = err

	old := f.state
	f.state = futureCompleted

	if old == futureBlocked {
		for i := 0; i < f.blockCount; i++ {
			f.blocker <- true
		}
	}

	f.fireComplete()
}

func (f *future) Cancel() error {
	f.stateLock.Lock()
	defer f.stateLock.Unlock()

	if f.state == futureCompleted {
		return errors.New("Task is already completed")
	}

	if f.state == futureCancelled {
		return nil
	}

	f.error = ErrorCancelled
	old := f.state
	f.state = futureCancelled

	if old == futureBlocked {
		for i := 0; i < f.blockCount; i++ {
			f.blocker <- true
		}
	}

	f.fireComplete()

	return nil
}

// OnComplete set complete callback
func (f *future) OnComplete(fn func(interface{}, error)) Disposer {
	return f.completeSource.on(fn)
}

// OnProgress set progress callback
func (f *future) OnProgress(fn func(interface{})) Disposer {
	return f.progressSource.on(fn)
}

func (f *future) fireComplete() {
	go f.completeSource.fire(func(fn interface{}) {
		fnn, _ := fn.(func(interface{}, error))
		fnn(f.result, f.error)
	})
}

func (f *future) fireProgress(p interface{}) {
	go f.progressSource.fire(func(fn interface{}) {
		fnn, _ := fn.(func(interface{}))
		fnn(p)
	})
}
