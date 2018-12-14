package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const (
	orderName = iota
	orderMTime
	orderSize
)

type column struct {
	path       string
	filter     string
	origin     []os.FileInfo
	files      []os.FileInfo
	markes     []int
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
	co := &column{path, "", fs, fs, nil, orderName, false, 0}
	co.update()
	return co
}

// show/hide hidden files, do filter, clear markes
func (co *column) update() {
	fs := make([]os.FileInfo, 0)
	for _, v := range co.origin {
		if !strings.Contains(v.Name(), co.filter) {
			continue
		}

		if !co.showHidden && strings.HasPrefix(v.Name(), ".") {
			continue
		}

		fs = append(fs, v)
	}
	co.files = fs
	co.current = 0

	co.sort(co.order)
	co.unmarkAll()
}

func (co *column) marked(idx int) bool {
	for _, i := range co.markes {
		if i == idx {
			return true
		}
	}
	return false
}

func (co *column) toggleMark() {
	ii := -1
	for idx, i := range co.markes {
		if i == co.current {
			ii = idx
			break
		}
	}
	if ii == -1 {
		co.markes = append(co.markes, co.current)
		return
	}

	co.markes = append(co.markes[:ii], co.markes[ii+1:]...)
}

func (co *column) unmarkAll() {
	co.markes = nil
}

func (co *column) move(n int) {
	if len(co.files) == 0 {
		return
	}
	i := co.current + n
	if i < 0 {
		i = len(co.files) - 1
	}

	if i >= len(co.files) {
		i = 0
	}

	co.current = i
}

type group struct {
	path    string
	columns []*column
}

func newGroup(path string) *group {
	return &group{path, []*column{newColumn(path)}}
}

func (gr *group) currentDir() string {
	co := gr.columns[len(gr.columns)-1]
	return co.path
}

func (gr *group) currentSelect() string {
	co := gr.columns[len(gr.columns)-1]
	return filepath.Join(co.path, co.files[co.current].Name())
}

func (gr *group) shift() {
	if len(gr.columns) == 1 {
		return
	}
	gr.columns = gr.columns[1:]
}

type workspace struct {
	bookmark     map[string]string
	groups       []*group
	group        int
	showBookmark bool
}

func newWorkspace() *workspace {
	gs := make([]*group, maxGroups)
	gs[0] = newGroup(wd)
	bo := map[string]string{
		"ws": "/Users/guyong/ws",
		"go": "/Users/guyong/ws/go",
	}
	return &workspace{bo, gs, 0, true}
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
	gui <- uiColumnContentChange
}

func (w *workspace) toggleHidden() {
	co := w.currentColumn()
	co.showHidden = !co.showHidden
	co.update()
	gui <- uiColumnContentChange
}

func (w *workspace) move(n int) {
	co := w.currentColumn()
	if len(co.files) == 0 {
		return
	}
	co.move(n)
	gui <- uiChangeSelect
}

func (w *workspace) moveToFirst() {
	co := w.currentColumn()
	w.move(-co.current)
}

func (w *workspace) moveToLast() {
	co := w.currentColumn()
	w.move(len(co.files) - co.current - 1)
}

func (w *workspace) openRight() {
	gu := w.currentGroup()
	co := gu.columns[len(gu.columns)-1]
	if len(co.files) == 0 {
		return
	}
	fi := co.files[co.current]

	if !fi.IsDir() {
		return
	}
	co.unmarkAll()

	pa := filepath.Join(co.path, fi.Name())
	nc := newColumn(pa)
	gu.path = pa
	gu.columns = append(gu.columns, nc)
	if len(gu.columns) >= maxColumns {
		gu.shift()
		gui <- uiOpenRightWithShift
	} else {
		gui <- uiOpenRight
	}
}

func (w *workspace) closeRight() {
	gu := w.currentGroup()
	if len(gu.columns) == 1 {
		dir := filepath.Dir(gu.path)
		if dir == gu.path {
			return
		}
		gu.path = dir
		co := gu.columns[0]
		co.path = dir
		co.origin, _ = ioutil.ReadDir(dir)
		co.update()

		gui <- uiToParent
		return
	}
	gu.columns = gu.columns[:len(gu.columns)-1]
	gu.path = gu.columns[len(gu.columns)-1].path
	gui <- uiCloseRight
}

func (w *workspace) shift() {
	w.currentGroup().shift()
	gui <- uiShift
}

func (w *workspace) toggleBookmark() {
	w.showBookmark = !w.showBookmark
	gui <- uiToggleBookmark
}

func (w *workspace) changeGroup(idx int) {
	w.group = idx
	if w.groups[idx] == nil {
		w.groups[idx] = newGroup(wd)
	}
	gui <- uiChangeGroup
}

func (w *workspace) openRoot(path string) {
	fs, err := ioutil.ReadDir(path)
	if err != nil {
		message = "Can not read dir " + path
		gui <- uiErrorMessage
		return
	}

	gu := w.currentGroup()
	gu.path = path

	gu.columns = gu.columns[:1]
	co := gu.columns[0]
	co.origin = fs
	co.path = path
	co.update()

	gui <- uiChangeRoot
}

func (w *workspace) jumpTo(colIdx, fileIdx int) bool {
	gu := w.currentGroup()
	gu.columns = gu.columns[0 : colIdx+1]
	co := gu.columns[len(gu.columns)-1]
	co.current = fileIdx

	fi := co.files[fileIdx]
	if !fi.IsDir() {
		gui <- uiJumpTo
		return false
	}

	pa := filepath.Join(co.path, fi.Name())
	nc := newColumn(pa)
	gu.path = pa
	gu.columns = append(gu.columns, nc)
	if len(gu.columns) >= maxColumns {
		gu.columns = gu.columns[1:]
	}
	gui <- uiJumpTo
	return true
}

func (w *workspace) refresh() {
	co := w.currentColumn()
	co.origin, _ = ioutil.ReadDir(co.path)
	co.update()
	gui <- uiColumnContentChange
}

func (w *workspace) toggleMark() {
	co := w.currentColumn()
	co.toggleMark()
	co.move(1)

	gui <- uiMarkChange
}

func (w *workspace) clearMark() {
	co := w.currentColumn()
	co.unmarkAll()

	gui <- uiMarkChange
}
