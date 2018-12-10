package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
)

const (
	orderName = iota
	orderMTime
	orderSize
)

type column struct {
	path       string
	files      []os.FileInfo
	order      int
	showHidden bool
	current    int
}

type files []os.FileInfo

func (c files) Len() int      { return len(c) }
func (c files) Swap(i, j int) { c[i], c[j] = c[j], c[i] }
func (c files) compare(i, j int) int {
	if c[i].IsDir() && !c[j].IsDir() {
		return -1
	}

	if !c[i].IsDir() && c[j].IsDir() {
		return 1
	}

	return 0
}

type byName struct{ files }
type byMTime struct{ files }
type bySize struct{ files }

func (c byName) Less(i, j int) bool {
	switch c.files.compare(i, j) {
	case -1:
		return true
	case 1:
		return false
	default:
		return c.files[i].Name() < c.files[j].Name()
	}
}

func (c byMTime) Less(i, j int) bool {
	switch c.files.compare(i, j) {
	case -1:
		return true
	case 1:
		return false
	default:
		return c.files[i].ModTime().After(c.files[j].ModTime())
	}
}

func (c bySize) Less(i, j int) bool {
	switch c.files.compare(i, j) {
	case -1:
		return true
	case 1:
		return false
	default:
		a, b := c.files[i], c.files[j]
		if a.Size() == b.Size() {
			return a.Name() < b.Name()
		}
		return a.Size() < b.Size()
	}
}

func (co *column) sort(order int) {
	switch order {
	case orderName:
		sort.Sort(byName{co.files})
	case orderMTime:
		sort.Sort(byMTime{co.files})
	case orderSize:
		sort.Sort(bySize{co.files})
	}
	co.order = order
}

func newColumn(path string) *column {
	fs, _ := ioutil.ReadDir(path)
	co := &column{path, fs, orderName, false, 0}
	co.sort(orderName)
	return co
}

type group struct {
	path    string
	columns []*column
}

func newGroup(path string) *group {
	return &group{path, []*column{newColumn(path)}}
}

func (gr group) currentDir() string {
	co := gr.columns[len(gr.columns)-1]
	return co.path
}

func (gr group) currentSelect() string {
	co := gr.columns[len(gr.columns)-1]
	return filepath.Join(co.path, co.files[co.current].Name())
}

type workspace struct {
	bookmark     []string
	groups       []*group
	group        int
	showBookmark bool
}

func newWorkspace() *workspace {
	gs := make([]*group, maxGroups)
	gs[0] = newGroup(wd)
	bo := []string{"/User/guyong/ws", "/User/guyong/ws/go"}
	return &workspace{bo, gs, 0, false}
}

func (w *workspace) currentGroup() *group {
	return w.groups[w.group]
}

func (w *workspace) currentDir() string {
	return wo.groups[wo.group].currentDir()
}

func (w *workspace) currentColumn() *column {
	cols := w.currentGroup().columns
	return cols[len(cols)-1]
}

func (w *workspace) sort(order int) {
	w.currentColumn().sort(order)
	gui <- uiChangeSort
}
