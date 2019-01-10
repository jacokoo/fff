package ui

import (
	"fmt"

	"github.com/jacokoo/fff/model"
)

const (
	taskDetailWidth = 60
)

// TaskItem render a task
type TaskItem struct {
	name *Text
	pb   *ProgressBar
	*Drawable
}

// NewTaskItem create task item
func NewTaskItem(p *Point, name string, width int) *TaskItem {
	pb := NewProgressBar(p, width, 0)
	return &TaskItem{NewText(p, name), pb, NewDrawable(p)}
}

// Draw it
func (ti *TaskItem) Draw() *Point {
	Move(ti.name, ti.Start)
	ti.End = Move(ti.pb, ti.Start.Down())
	return ti.End
}

// SetData update the progress
func (ti *TaskItem) SetData(name string, progress int) {
	ti.name.Data = name
	ti.pb.Progress = progress
}

// BatchTaskItem render a batch task
type BatchTaskItem struct {
	max      int
	progress *Text
	task     *TaskItem
	*Drawable
}

// NewBatchTaskItem create batch task
func NewBatchTaskItem(p *Point, max, width int) *BatchTaskItem {
	return &BatchTaskItem{max, NewText(p, ""), NewTaskItem(p, "", width), NewDrawable(p)}
}

// Draw it
func (bt *BatchTaskItem) Draw() *Point {
	p := Move(bt.task, bt.Start)
	Move(bt.progress, p.Up().MoveLeftN(len(bt.progress.Data)-1))
	bt.End = p
	return bt.End
}

// SetData update state
func (bt *BatchTaskItem) SetData(name string, current int, subname string, subprogress int) {
	bt.progress.Data = fmt.Sprintf("[%d/%d]", current, bt.max)
	bt.task.SetData(fmt.Sprintf("%s / %s", name, subname), subprogress)
}

type pool struct {
	ts  []*TaskItem
	bts []*BatchTaskItem
}

func (p *pool) getTask() *TaskItem {
	if len(p.ts) > 0 {
		ti := p.ts[0]
		p.ts = p.ts[1:]
		return ti
	}

	return NewTaskItem(ZeroPoint, "", taskDetailWidth)
}

func (p *pool) getBatchTask() *BatchTaskItem {
	if len(p.bts) > 0 {
		ti := p.bts[0]
		p.bts = p.bts[1:]
		return ti
	}

	return NewBatchTaskItem(ZeroPoint, 0, taskDetailWidth)
}

func (p *pool) release(d Drawer) {
	switch dd := d.(type) {
	case *TaskItem:
		p.ts = append(p.ts, dd)
	case *BatchTaskItem:
		p.bts = append(p.bts, dd)
	}
}

// Task ui
type Task struct {
	showDetail bool
	pool       *pool
	items      []Drawer
	layout     *VerticalLayout
	popup      *Popup
	*Text
}

// NewTask create task
func NewTask(p *Point) *Task {
	vl := NewVerticalLayout(p, func(p *Point) *Point {
		return p.Down()
	})
	box := NewDBox(p, vl, 1)
	return &Task{false, new(pool), nil, vl, NewPopup(p, box), NewText(p, "")}
}

// Open popup
func (t *Task) Open() {
	t.showDetail = true
	p := t.End.Down()
	p.X -= taskDetailWidth + 4
	Move(t.popup, p)
}

// Close popup
func (t *Task) Close() {
	t.showDetail = false
	t.popup.Clear()
}

// SetData update state
func (t *Task) SetData(ts []model.Task) {
	for _, v := range t.items {
		t.pool.release(v)
	}

	ss := make([]Drawer, len(ts))
	n := ""
	for i, v := range ts {
		switch vv := v.(type) {
		case model.BatchTask:
			bb := t.pool.getBatchTask()
			bb.max = vv.Count()
			ct := vv.CurrentTask()
			bb.SetData(vv.Name(), vv.Current()+1, ct.Name(), ct.Current())
			ss[i] = bb
		case model.Task:
			bb := t.pool.getTask()
			bb.SetData(vv.Name(), vv.Current())
			ss[i] = bb
		}
		n = fmt.Sprintf("%s[%s %d/%d]", n, v.Name(), v.Current()+1, v.Count())
	}

	t.Data = n
	t.items = ss
	t.layout.Drawers = ss
}
