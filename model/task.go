package model

// Task a task
type Task interface {
	Name() string
	Count() int
	Start(<-chan bool)
	Attach(chan<- int)
	Detach(chan<- int)
}

// BatchTask mutiple tasks
type BatchTask interface {
	CurrentTask() Task
	Task
}

// DefaultTask default task
type DefaultTask struct {
	name      string
	listeners []chan<- int
	action    func(chan<- int, <-chan bool)
}

// NewTask create task
func NewTask(name string, action func(chan<- int, <-chan bool)) Task {
	return &DefaultTask{name, nil, action}
}

// Name return task name
func (dt *DefaultTask) Name() string {
	return dt.name
}

// Count default to 100
func (dt *DefaultTask) Count() int {
	return 100
}

// Start start the task
func (dt *DefaultTask) Start(quit <-chan bool) {
	prog := make(chan int)
	qt := make(chan bool)

	go dt.action(prog, qt)
	for {
		select {
		case p, ok := <-prog:
			if !ok {
				dt.listeners = nil
				return
			}
			for _, v := range dt.listeners {
				v <- p
			}
		case <-quit:
			dt.listeners = nil
			qt <- true
			return
		}
	}
}

// Attach attach notifier
func (dt *DefaultTask) Attach(listener chan<- int) {
	dt.listeners = append(dt.listeners, listener)
}

// Detach attach notifier
func (dt *DefaultTask) Detach(listener chan<- int) {
	ls := make([]chan<- int, 0)
	for _, v := range dt.listeners {
		if v != listener {
			continue
		}
		ls = append(ls, v)
	}
	dt.listeners = ls
}

// DefaultBatchTask default batch task
type DefaultBatchTask struct {
	count   int
	current int
	tasks   []Task
	*DefaultTask
}

// Count the count of tasks
func (bt *DefaultBatchTask) Count() int {
	return bt.count
}

// Progress the index of current task
func (bt *DefaultBatchTask) Progress() int {
	return bt.current
}

// CurrentTask the current task
func (bt *DefaultBatchTask) CurrentTask() Task {
	return bt.tasks[bt.current]
}

// Start task one by one
func (bt *DefaultBatchTask) Start(quit <-chan bool) {
	for i, t := range bt.tasks {
		qt := make(chan bool)
		prog := make(chan int)
		bt.current = i

		t.Attach(prog)
		go t.Start(qt)
	progress:
		for {
			select {
			case _, ok := <-prog:
				if !ok {
					break progress
				}
				for _, v := range bt.listeners {
					v <- i
				}
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
}

// TaskAbort execute to abort task
type TaskAbort func()

func (tm *TaskManager) waitIt(task Task, ch chan int) {
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
			task.Detach(ch)
		}
	}
}

// Submit a task to execute
func (tm *TaskManager) Submit(task Task) TaskAbort {
	tm.Tasks = append(tm.Tasks, task)
	quit := make(chan bool)
	complete := make(chan int)
	task.Attach(complete)

	go tm.waitIt(task, complete)
	go task.Start(quit)

	return func() {
		quit <- true
	}
}
