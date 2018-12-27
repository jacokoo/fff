package model

var (
	_ = Progresser(new(DefaultProgresser))
	_ = Task(new(DefaultTask))
)

// Progresser a progress notifier
type Progresser interface {
	Count() int
	Current() int
	Progress(int)
	Attach(chan<- int)
	Detach(chan<- int)
	End()
}

// DefaultProgresser a progress notifier
type DefaultProgresser struct {
	count     int
	progress  int
	listeners []chan<- int
}

// Count progress count
func (dp *DefaultProgresser) Count() int {
	return dp.count
}

// Current current progress
func (dp *DefaultProgresser) Current() int {
	return dp.progress
}

// Progress set the progress
func (dp *DefaultProgresser) Progress(c int) {
	dp.progress = c
	for _, v := range dp.listeners {
		v <- c
	}
}

// Attach attach notifier
func (dp *DefaultProgresser) Attach(listener chan<- int) {
	dp.listeners = append(dp.listeners, listener)
}

// Detach attach notifier
func (dp *DefaultProgresser) Detach(listener chan<- int) {
	ls := make([]chan<- int, 0)
	for _, v := range dp.listeners {
		if v != listener {
			continue
		}
		ls = append(ls, v)
	}
	dp.listeners = ls
}

// End close all listeners
func (dp *DefaultProgresser) End() {
	for _, v := range dp.listeners {
		close(v)
	}
	dp.listeners = nil
}

// Task a task
type Task interface {
	Name() string
	Start(<-chan bool, chan<- error)
	Progresser
}

// BatchTask mutiple tasks
type BatchTask interface {
	CurrentTask() Task
	Task
}

// DefaultTask default task
type DefaultTask struct {
	name   string
	action func(chan<- int, <-chan bool, chan<- error)
	*DefaultProgresser
}

// NewTask create task
func NewTask(name string, action func(chan<- int, <-chan bool, chan<- error)) Task {
	return &DefaultTask{name, action, &DefaultProgresser{100, 0, nil}}
}

// Name return task name
func (dt *DefaultTask) Name() string {
	return dt.name
}

// Start start the task
func (dt *DefaultTask) Start(quit <-chan bool, err chan<- error) {
	defer dt.End()

	prog := make(chan int)
	qt := make(chan bool)

	go dt.action(prog, qt, err)
	for {
		select {
		case p, ok := <-prog:
			if !ok {
				return
			}
			dt.Progress(p)
		case <-quit:
			qt <- true
			return
		}
	}
}

// DefaultBatchTask default batch task
type DefaultBatchTask struct {
	tasks []Task
	*DefaultTask
}

// NewBatchTask create batch task
func NewBatchTask(name string, tasks []Task) BatchTask {
	return &DefaultBatchTask{tasks, &DefaultTask{name, nil, &DefaultProgresser{len(tasks), 0, nil}}}
}

// CurrentTask the current task
func (bt *DefaultBatchTask) CurrentTask() Task {
	return bt.tasks[bt.progress]
}

// Start task one by one
func (bt *DefaultBatchTask) Start(quit <-chan bool, err chan<- error) {
	defer close(err)
	defer bt.End()

	for i, t := range bt.tasks {
		qt := make(chan bool)
		prog := make(chan int)
		err1 := make(chan error)
		bt.progress = i

		t.Attach(prog)
		go t.Start(qt, err1)
	progress:
		for {
			select {
			case _, ok := <-prog:
				if !ok {
					break progress
				}
				bt.Progress(i)
			case e, ok := <-err1:
				if !ok {
					break progress
				}
				err <- e
			case <-quit:
				qt <- true
				return
			}
		}
	}
}

// TaskManager manage tasks
type TaskManager struct {
	Tasks []Task
	quits map[Task]chan bool
}

// NewTaskManager create task manager
func NewTaskManager() *TaskManager {
	return &TaskManager{nil, make(map[Task]chan bool)}
}

func (tm *TaskManager) waitIt(task Task, ch chan int, message chan string) {
	defer close(message)

	for {
		_, ok := <-ch
		if !ok {
			ts := make([]Task, 0)
			for _, v := range tm.Tasks {
				if v != task {
					ts = append(ts, v)
				}
			}
			tm.Tasks = ts
			delete(tm.quits, task)
			task.Detach(ch)
			return
		}
	}
}

// Submit a task to execute
func (tm *TaskManager) Submit(task Task) <-chan string {
	tm.Tasks = append(tm.Tasks, task)
	quit := make(chan bool)
	complete := make(chan int)
	err := make(chan error)
	message := make(chan string)
	task.Attach(complete)

	go tm.waitIt(task, complete, message)
	go task.Start(quit, err)
	go func() {
		for v := range err {
			message <- v.Error()
		}
	}()

	tm.quits[task] = quit
	return message
}
