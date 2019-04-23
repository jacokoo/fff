package executor

// Disposer dispose something
type Disposer interface {
	Dispose()
}

// Executor task executor
type Executor interface {
	Submit(string, TaskFunc) Future
	Running() []Task
	Close() error
}

// TaskState state
type TaskState interface {
	IsCancelled() bool
	Progress(interface{})
}

// TaskFunc the func to do the task
type TaskFunc func(TaskState) (interface{}, error)

// Task with return value
type Task interface {
	Name() string
	Future() Future
}

// Future get value after complete
type Future interface {
	Get() (interface{}, error)
	Cancel() error
	OnComplete(func(interface{}, error)) Disposer
	OnProgress(func(interface{})) Disposer
}

func CompletedFuture(data interface{}, err error) Future {
	f := newFuture()
	f.set(data, nil)
	return f
}

func OkFuture(data interface{}) Future {
	return CompletedFuture(data, nil)
}

func ErrorFuture(err error) Future {
	return CompletedFuture(nil, err)
}

func Combine(f Future, fn ...func(interface{}) Future) Future {
	return newCombinedFuture(f, fn)
}
