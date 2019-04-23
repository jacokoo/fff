package executor

import (
	"errors"
	"sync"
)

var (
	_ = Future(new(future))

	errorCancelled = errors.New("Task is already cancelled")
)

type futureState uint16

// Future state
const (
	futurePadding futureState = iota
	futureBlocked
	futureCancelled
	futureCompleted
)

type future struct {
	completeSource *eventSource
	progressSource *eventSource

	result    interface{}
	error     error
	blocker   *sync.Cond
	state     futureState
	stateLock *sync.RWMutex
}

func newFuture() *future {
	return &future{
		completeSource: newEventSource(),
		progressSource: newEventSource(),
		blocker:        sync.NewCond(new(sync.Mutex)),
		stateLock:      new(sync.RWMutex),
	}
}

func (f *future) Get() (interface{}, error) {
	f.stateLock.Lock()

	if f.state == futureCancelled {
		f.stateLock.Unlock()
		return nil, errorCancelled
	}

	if f.state == futureCompleted {
		f.stateLock.Unlock()
		return f.result, f.error
	}

	f.state = futureBlocked
	f.stateLock.Unlock()

	f.blocker.L.Lock()
	f.blocker.Wait()
	f.blocker.L.Unlock()

	f.stateLock.RLock()
	defer f.stateLock.RUnlock()
	if f.state == futureCancelled {
		return nil, errorCancelled
	}

	if f.state == futureCompleted {
		return f.result, f.error
	}

	return nil, errors.New("Unknown state")
}

func (f *future) set(result interface{}, err error) {
	f.stateLock.Lock()

	if f.state != futurePadding && f.state != futureBlocked {
		f.stateLock.Unlock()
		return
	}
	f.result = result
	f.error = err

	old := f.state
	f.state = futureCompleted
	f.stateLock.Unlock()

	if old == futureBlocked {
		f.blocker.Broadcast()
	}

	f.fireComplete()
}

func (f *future) Cancel() error {
	f.stateLock.Lock()

	if f.state == futureCompleted {
		f.stateLock.Unlock()
		return errors.New("Task is already completed")
	}

	if f.state == futureCancelled {
		f.stateLock.Unlock()
		return nil
	}

	f.error = errorCancelled
	old := f.state
	f.state = futureCancelled
	f.stateLock.Unlock()

	if old == futureBlocked {
		f.blocker.Broadcast()
	}

	f.fireComplete()

	return nil
}

// OnComplete set complete callback
func (f *future) OnComplete(fn func(interface{}, error)) Disposer {
	if f.state == futureCancelled || f.state == futureCompleted {
		fn(f.result, f.error)
	}
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

type combinedFuture struct {
	current Future
	idx     int
	items   []func(interface{}) Future
	mx      *sync.Mutex
	*future
}

func newCombinedFuture(start Future, fs []func(interface{}) Future) Future {
	cf := &combinedFuture{start, 0, fs, new(sync.Mutex), newFuture()}
	cf.doNext()
	return cf
}

func (cf *combinedFuture) doNext() {
	cf.current.OnComplete(func(data interface{}, err error) {
		if err != nil {
			cf.set(nil, err)
			return
		}
		if cf.idx == len(cf.items) {
			cf.set(data, nil)
			return
		}
		cf.mx.Lock()
		cf.current = cf.items[cf.idx](data)
		cf.mx.Unlock()
		cf.idx++
		cf.doNext()
	})
}

func (cf *combinedFuture) Cancel() error {
	cf.mx.Lock()
	defer cf.mx.Unlock()
	return cf.current.Cancel()
}
