package executor

var (
	_ = TaskState(new(taskState))
)

// TaskState state
type TaskState interface {
	IsCancelled() bool
	Progress(interface{})
}

// RunningTask with return value
type RunningTask interface {
	Name() string
	Future() Future
}

type taskState struct {
	fu *future
}

func newTaskState() *taskState {
	return &taskState{newFuture()}
}

func (t *taskState) IsCancelled() bool {
	t.fu.stateLock.RLock()
	defer t.fu.stateLock.RUnlock()
	return t.fu.state == futureCancelled
}

func (t *taskState) Progress(p interface{}) {
	t.fu.fireProgress(p)
}

// TaskFunc the func to do the task
type TaskFunc func(TaskState) (interface{}, error)

type task struct {
	name string
	fn   TaskFunc
}

func (t *task) Name() string {
	return t.name
}

func (t *task) Do(ts TaskState) (interface{}, error) {
	return t.fn(ts)
}
