package main

import (
	"fmt"
	"os"

	"github.com/jacokoo/fff/model"
)

type workspace struct {
	groups       []model.Group
	group        int
	showBookmark bool
}

func newWorkspace() *workspace {
	gs := make([]model.Group, maxGroups)
	g, err := model.NewLocalGroup(wd)
	if err != nil {
		panic(err)
	}
	gs[0] = g

	return &workspace{gs, 0, true}
}

func (w *workspace) currentGroup() model.Group {
	return w.groups[w.group]
}

func (w *workspace) currentDir() string {
	return w.currentGroup().Path()
}

func (w *workspace) currentColumn() model.Column {
	return w.currentGroup().Current()
}

func (w *workspace) sort(order model.Order) {
	w.currentColumn().Sort(order)
	gui <- uiColumnContentChange
}

func (w *workspace) toggleHidden() {
	co := w.currentColumn()
	co.ToggleHidden()
	co.Update()
	gui <- uiColumnContentChange
}

func (w *workspace) toggleDetails() {
	w.currentColumn().ToggleDetail()
	gui <- uiToggleDetail
}

func (w *workspace) move(n int) {
	if w.currentColumn().Move(n) {
		gui <- uiChangeSelect
	}
}

func (w *workspace) moveToFirst() {
	if w.currentColumn().SelectFirst() {
		gui <- uiChangeSelect
	}
}

func (w *workspace) moveToLast() {
	if w.currentColumn().SelectLast() {
		gui <- uiChangeSelect
	}
}

func (w *workspace) openRight() {
	gu := w.currentGroup()
	err := gu.OpenDir()
	if err != nil {
		message = err.Error()
		gui <- uiErrorMessage
		return
	}

	if len(gu.Columns()) >= maxColumns {
		gu.Shift()
		gui <- uiOpenRightWithShift
	} else {
		gui <- uiOpenRight
	}
}

func (w *workspace) closeRight() {
	gu := w.currentGroup()
	switch re := gu.CloseDir(); re {
	case model.CloseNothing:
		return
	case model.CloseSuccess:
		gui <- uiCloseRight
	case model.CloseToParent:
		gui <- uiToParent
	}
}

func (w *workspace) shift() {
	w.currentGroup().Shift()
	gui <- uiShift
}

func (w *workspace) toggleBookmark() {
	w.showBookmark = !w.showBookmark
	gui <- uiToggleBookmark
}

func (w *workspace) changeGroup(idx int) {
	w.group = idx
	if w.groups[idx] == nil {
		g, _ := model.NewLocalGroup(wd)
		w.groups[idx] = g
	}
	gui <- uiChangeGroup
}

func (w *workspace) openRoot(path string) {
	err := w.currentGroup().OpenRoot(path)
	if err != nil {
		message = "Can not read dir " + path
		gui <- uiErrorMessage
		return
	}
	gui <- uiChangeRoot
}

func (w *workspace) jumpTo(colIdx, fileIdx int, openIt bool) bool {
	gu := w.currentGroup()
	suc := gu.JumpTo(colIdx, fileIdx)
	if !suc {
		return false
	}
	co := gu.Current()

	fi, err := co.CurrentFile()
	if err != nil || !openIt || !fi.IsDir() {
		gui <- uiJumpTo
		return false
	}

	co.Update()
	if co.IsShowDetail() {
		co.ToggleDetail()
	}

	gu.OpenDir()
	if len(gu.Columns()) >= maxColumns {
		gu.Shift()
	}
	gui <- uiJumpTo
	return true
}

func (w *workspace) refresh() {
	w.currentGroup().Refresh()
	gui <- uiColumnContentChange
}

func (w *workspace) toggleMark() {
	co := w.currentColumn()
	co.ToggleMark()
	co.Move(1)
	gui <- uiMarkChange
}

func (w *workspace) clearMark() {
	w.currentColumn().ClearMark()
	gui <- uiMarkChange
}

func (w *workspace) clearFilter() {
	co := w.currentColumn()
	co.SetFilter("")
	co.Update()
	gui <- uiColumnContentChange
}

func (w *workspace) newFile(name string) {
	g := w.currentGroup()
	if err := g.NewFile(g.Path(), name); err != nil {
		message = err.Error()
		gui <- uiErrorMessage
		return
	}
	g.Refresh()
	g.Current().SelectByName(name)
	gui <- uiColumnContentChange
}

func (w *workspace) newDir(name string) {
	g := w.currentGroup()
	if err := g.NewDir(g.Path(), name); err != nil {
		message = err.Error()
		gui <- uiErrorMessage
		return
	}
	g.Refresh()
	g.Current().SelectByName(name)
	gui <- uiColumnContentChange
}

func (w *workspace) rename(name string) {
	g := w.currentGroup()
	co := g.Current()
	fi, err := co.CurrentFile()
	if err != nil {
		message = "no file selected"
		gui <- uiErrorMessage
		return
	}

	if err := g.Rename(g.Path(), fi.Name(), name); err != nil {
		message = fmt.Sprintf("Can not rename %s to %s, %s", fi.Name(), name, err.Error())
		gui <- uiErrorMessage
		return
	}
	g.Refresh()
	g.Current().SelectByName(name)
	gui <- uiColumnContentChange
}

func selectString(dirs, files int) string {
	m := "Selected"
	u := "s"
	if files != 0 {
		if files == 1 {
			u = ""
		}
		m = fmt.Sprintf("%s %d file%s", m, files, u)
	}

	if files != 0 && dirs != 0 {
		m += " and "
	}

	u = "s"
	if dirs != 0 {
		m = fmt.Sprintf("%s %d dir%s", m, dirs, u)
	}

	return m
}

func (w *workspace) deletePrompt() string {
	files := w.currentColumn().Marked()
	if len(files) == 0 {
		return ""
	}

	fc, dc := 0, 0
	for _, v := range files {
		if v.IsDir() {
			dc++
		} else {
			fc++
		}
	}

	m := selectString(dc, fc)

	u := "them"
	if fc+dc == 1 {
		u = "it"
	}
	m = fmt.Sprintf("%s. Are you sure to delete %s? (y/n)", m, u)
	return m
}

func (w *workspace) deleteFiles() {
	g := w.currentGroup()
	co := g.Current()
	if len(co.Files()) == 0 {
		return
	}

	selected, er := co.CurrentFile()
	files := co.Marked()
	fc, dc := 0, 0
	for _, v := range files {
		if v.IsDir() {
			err := os.RemoveAll(v.Path())
			if err == nil {
				dc++
			}
			continue
		}

		err := os.Remove(v.Path())
		if err == nil {
			fc++
		}
	}

	m := selectString(dc, fc)
	m += " Deleted"
	message = m
	gui <- uiErrorMessage

	g.Refresh()
	if er == nil {
		co.SelectByName(selected.Name())
	} else {
		co.Select(0)
	}
	gui <- uiColumnContentChange
}
