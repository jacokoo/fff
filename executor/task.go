package executor

var (
	_ = TaskState(new(taskState))
)

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
