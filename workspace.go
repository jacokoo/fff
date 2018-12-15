package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

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

func (w *workspace) toggleDetails() {
	co := w.currentColumn()
	co.expanded = !co.expanded
	gui <- uiToggleDetail
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
	co.expanded = false

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

func (w *workspace) jumpTo(colIdx, fileIdx int, openIt bool) bool {
	gu := w.currentGroup()
	gu.columns = gu.columns[0 : colIdx+1]
	co := gu.columns[len(gu.columns)-1]
	co.current = fileIdx

	fi := co.files[fileIdx]
	if !openIt || !fi.IsDir() {
		gui <- uiJumpTo
		return false
	}
	co.unmarkAll()
	co.expanded = false

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
	w.currentColumn().refresh()
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

func (w *workspace) clearFilter() {
	co := w.currentColumn()
	co.filter = ""
	co.update()
	gui <- uiColumnContentChange
}

func (w *workspace) newFile(name string) {
	co := w.currentColumn()
	pa := filepath.Join(co.path, name)
	if _, err := os.Create(pa); err != nil {
		message = "Can not create file " + pa
		gui <- uiErrorMessage
		return
	}
	co.refreshWithName(name)
	gui <- uiColumnContentChange
}

func (w *workspace) newDir(name string) {
	co := w.currentColumn()
	pa := filepath.Join(co.path, name)
	if err := os.MkdirAll(pa, 0755); err != nil {
		message = "Can not create dir " + pa
		gui <- uiErrorMessage
		return
	}
	co.refreshWithName(name)
	gui <- uiColumnContentChange
}

func (w *workspace) rename(name string) {
	co := w.currentColumn()
	if len(co.files) == 0 {
		return
	}

	old := filepath.Join(co.path, co.files[co.current].Name())
	new := filepath.Join(co.path, name)

	if err := os.Rename(old, new); err != nil {
		message = fmt.Sprintf("Can not rename %s to %s", co.files[co.current].Name(), name)
		gui <- uiErrorMessage
		return
	}
	co.refreshWithName(name)
	gui <- uiColumnContentChange
}
