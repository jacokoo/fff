package model

// Task a task
type Task interface {
	Name() string
	Count() int
	Progress() int
	Start(<-chan bool, chan<- error)
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
	progress  int
	listeners []chan<- int
	action    func(chan<- int, <-chan bool, chan<- error)
}

// NewTask create task
func NewTask(name string, action func(chan<- int, <-chan bool, chan<- error)) Task {
	return &DefaultTask{name, 0, nil, action}
}

// Name return task name
func (dt *DefaultTask) Name() string {
	return dt.name
}

// Count default to 100
func (dt *DefaultTask) Count() int {
	return 100
}

// Progress the progress
func (dt *DefaultTask) Progress() int {
	return dt.progress
}

func (dt *DefaultTask) close() {
	for _, v := range dt.listeners {
		close(v)
	}
	dt.listeners = nil
}

// Start start the task
func (dt *DefaultTask) Start(quit <-chan bool, err chan<- error) {
	defer dt.close()

	prog := make(chan int)
	qt := make(chan bool)

	go dt.action(prog, qt, err)
	for {
		select {
		case p, ok := <-prog:
			if !ok {
				return
			}
			dt.progress = p
			for _, v := range dt.listeners {
				v <- p
			}
		case <-quit:
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
	count int
	tasks []Task
	*DefaultTask
}

// NewBatchTask create batch task
func NewBatchTask(name string, tasks []Task) BatchTask {
	return &DefaultBatchTask{len(tasks), tasks, &DefaultTask{name, 0, nil, nil}}
}

// Count the count of tasks
func (bt *DefaultBatchTask) Count() int {
	return bt.count
}

// Progress the index of current task
func (bt *DefaultBatchTask) Progress() int {
	return bt.progress
}

// CurrentTask the current task
func (bt *DefaultBatchTask) CurrentTask() Task {
	return bt.tasks[bt.progress]
}

// Start task one by one
func (bt *DefaultBatchTask) Start(quit <-chan bool, err chan<- error) {
	defer close(err)
	defer bt.close()

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
				for _, v := range bt.listeners {
					v <- i
				}
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
