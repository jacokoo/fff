package main

import (
	"fmt"
	"strings"

	"github.com/jacokoo/fff/model"
	"github.com/jacokoo/fff/ui"
)

type action struct {
}

func (w *action) sort(order model.Order) {
	co := wo.CurrentGroup().Current()
	co.Sort(order)
	ui.ColumnContentChangeEvent.Send(co)
}

func (w *action) toggleHidden() {
	co := wo.CurrentGroup().Current()
	co.ToggleHidden()
	co.Update()
	ui.ColumnContentChangeEvent.Send(co)
}

func (w *action) toggleDetails() {
	co := wo.CurrentGroup().Current()
	co.ToggleDetail()
	ui.ToggleDetailEvent.Send(co)
}

func (w *action) move(n int) {
	co := wo.CurrentGroup().Current()

	if co.Move(n) {
		ui.ChangeSelectEvent.Send(co)
	}
}

func (w *action) moveToFirst() {
	co := wo.CurrentGroup().Current()

	if co.SelectFirst() {
		ui.ChangeSelectEvent.Send(co)
	}
}

func (w *action) moveToLast() {
	co := wo.CurrentGroup().Current()

	if co.SelectLast() {
		ui.ChangeSelectEvent.Send(co)
	}
}

func (w *action) openRight() {
	gu := wo.CurrentGroup()
	err := gu.OpenDir()
	if err != nil {
		ui.MessageEvent.Send(err.Error())
		return
	}

	if len(gu.Columns()) >= maxColumns {
		gu.Shift()
	}

	ui.OpenRightEvent.Send(gu)
}

func (w *action) closeRight() {
	gu := wo.CurrentGroup()
	switch re := gu.CloseDir(); re {
	case model.CloseNothing:
		return
	case model.CloseSuccess:
		ui.CloseRightEvent.Send(gu.Current())
	case model.CloseToParent:
		ui.ToParentEvent.Send(gu.Current())
	}
}

func (w *action) shift() {
	gu := wo.CurrentGroup()
	if gu.Shift() {
		ui.ShiftEvent.Send(gu)
	}
}

func (w *action) toggleBookmark() {
	wo.ToggleBookmark()
	ui.ToggleBookmarkEvent.Send(wo.IsShowBookmark())
}

func (w *action) changeGroup(idx int) {
	wo.Current = idx
	if wo.Groups[idx] == nil {
		g, _ := model.NewLocalGroup(wd)
		wo.Groups[idx] = g
	}
	ui.ChangeGroupEvent.Send(wo)
}

func (w *action) openRoot(path string) {
	err := wo.CurrentGroup().OpenRoot(path)
	if err != nil {
		ui.MessageEvent.Send("Can not read dir " + path)
		return
	}
	ui.ChangeRootEvent.Send(wo.CurrentGroup())
}

func (w *action) jumpTo(colIdx, fileIdx int, openIt bool) bool {
	gu := wo.CurrentGroup()
	suc := gu.JumpTo(colIdx, fileIdx)
	if !suc {
		return false
	}
	co := gu.Current()

	fi, err := co.CurrentFile()
	if err != nil || !openIt || !fi.IsDir() {
		ui.JumpToEvent.Send(gu)
		return false
	}

	co.ClearMark()
	if co.IsShowDetail() {
		co.ToggleDetail()
	}

	gu.OpenDir()
	if len(gu.Columns()) >= maxColumns {
		gu.Shift()
	}
	ui.JumpToEvent.Send(gu)
	return true
}

func (w *action) refresh() {
	wo.CurrentGroup().Refresh()
	ui.ColumnContentChangeEvent.Send(wo.CurrentGroup().Current())
}

func (w *action) toggleMark() {
	co := wo.CurrentGroup().Current()
	co.ToggleMark()
	co.Move(1)
	ui.ColumnContentChangeEvent.Send(co)
}

func (w *action) clearMark() {
	co := wo.CurrentGroup().Current()
	co.ClearMark()
	ui.ColumnContentChangeEvent.Send(co)
}

func (w *action) clearFilter() {
	co := wo.CurrentGroup().Current()
	co.SetFilter("")
	co.Update()
	ui.ColumnContentChangeEvent.Send(co)
}

func (w *action) newFile(name string) {
	g := wo.CurrentGroup()
	if err := g.NewFile(g.Path(), name); err != nil {
		ui.MessageEvent.Send(err.Error())
		return
	}
	g.Refresh()
	g.Current().SelectByName(name)
	ui.ColumnContentChangeEvent.Send(g.Current())
}

func (w *action) newDir(name string) {
	g := wo.CurrentGroup()
	if err := g.NewDir(g.Path(), name); err != nil {
		ui.MessageEvent.Send(err.Error())
		return
	}
	g.Refresh()
	g.Current().SelectByName(name)
	ui.ColumnContentChangeEvent.Send(g.Current())
}

func (w *action) rename(name string) {
	g := wo.CurrentGroup()
	co := g.Current()
	fi, err := co.CurrentFile()
	if err != nil {
		ui.MessageEvent.Send("no file selected")
		return
	}

	if err := g.Rename(g.Path(), fi.Name(), name); err != nil {
		ui.MessageEvent.Send(fmt.Sprintf("Can not rename %s to %s, %s", fi.Name(), name, err.Error()))
		return
	}
	g.Refresh()
	g.Current().SelectByName(name)
	ui.ColumnContentChangeEvent.Send(g.Current())
}

func selectString(dirs, files int, prefix bool) string {
	m := ""
	if prefix {
		m = "Selected"
	}

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

func (w *action) deletePrompt() string {
	files := wo.CurrentGroup().Current().Marked()
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

	m := selectString(dc, fc, true)

	u := "them"
	if fc+dc == 1 {
		u = "it"
	}
	m = fmt.Sprintf("%s. Are you sure to delete %s? (y/n)", m, u)
	return m
}

func (w *action) deleteFiles() {
	g := wo.CurrentGroup()
	co := g.Current()
	if len(co.Files()) == 0 {
		return
	}

	selected, er := co.CurrentFile()
	files := co.Marked()
	fc, dc := 0, 0
	for _, v := range files {
		if v.IsDir() {
			err := g.DeleteDir(v.Path())
			if err == nil {
				dc++
			}
			continue
		}

		err := g.DeleteFile(v.Path())
		if err == nil {
			fc++
		}
	}

	g.Refresh()
	if er == nil {
		co.SelectByName(selected.Name())
	} else {
		co.Select(0)
	}
	ui.Batch(
		ui.ColumnContentChangeEvent.With(co),
		ui.MessageEvent.With(selectString(dc, fc, false)+" Deleted"),
	)
}

func (w *action) addBookmark(name, value string) {
	err := wo.Bookmark.Add(name, value)
	if err != nil {
		ui.MessageEvent.Send(err.Error())
		return
	}
	ui.BookmarkChangedEvent.Send(wo.Bookmark)
}

func (w *action) deleteBookmark(name string) {
	err := wo.Bookmark.Delete(name)
	if err != nil {
		ui.MessageEvent.Send(err.Error())
		return
	}
	ui.BookmarkChangedEvent.Send(wo.Bookmark)
}

func (w *action) openFile() {
	g := wo.CurrentGroup()
	file, err := g.Current().CurrentFile()
	if err != nil {
		ui.MessageEvent.Send(err.Error())
		return
	}

	err = g.Open(file.Path())
	if err != nil {
		ui.MessageEvent.Send(err.Error())
	}
}

func (w *action) clipFile() {
	co := wo.CurrentGroup().Current()
	if clip == nil {
		clip = model.CopySource(co.Marked())
		ui.Batch(
			ui.MessageEvent.With("Marked/Selected files are clipped"),
			ui.ClipChangedEvent.With(clip),
		)
		return
	}

	count := 0
	for _, v := range co.Marked() {
		has := false
		for _, vv := range clip {
			if strings.HasPrefix(v.Path(), vv.Path()) {
				has = true
				break
			}
		}
		if !has {
			count++
			clip = append(clip, v)
		}
	}

	ui.Batch(
		ui.MessageEvent.With(fmt.Sprintf("%d items appended to clip", count)),
		ui.ClipChangedEvent.With(clip),
	)
}

func (w *action) clearClip() {
	clip = nil
	ui.ClipChangedEvent.Send(nil)
}

func (w *action) copyFile() {
	if clip == nil {
		ui.MessageEvent.Send("No clipped files")
		return
	}

	g := wo.CurrentGroup()
	task, ok := clip.Task(g, g.Current().Path())
	clip = nil
	if !ok {
		ui.MessageEvent.Send("No file to copy")
		return
	}

	msg := tm.Submit(task)
	go func() {
		for v := range msg {
			ui.MessageEvent.Send(v)
		}
		ui.MessageEvent.Send("Copy done")
	}()
	ui.ClipChangedEvent.Send(nil)
}

func (w *action) moveFile() {
	if clip == nil {
		ui.MessageEvent.Send("No clipped files")
		return
	}

	g := wo.CurrentGroup()
	err := clip.MoveTo(g, g.Current().Path())
	if err != nil {
		ui.MessageEvent.Send(err.Error())
		return
	}

	for _, v := range clip {
		if v.IsDir() {
			g.DeleteDir(v.Path())
		}
	}
	clip = nil

	wo.CurrentGroup().Refresh()
	ui.Batch(
		ui.MessageEvent.With("Move done"),
		ui.ColumnContentChangeEvent.With(g.Current()),
		ui.ClipChangedEvent.With(nil),
	)
}
