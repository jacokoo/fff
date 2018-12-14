package main

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/jacokoo/fff/ui"
	termbox "github.com/nsf/termbox-go"
)

const (
	uiChangeGroup = iota
	uiAddColumn
	uiRemoveColumn
	uiChangeSelect
	uiColumnContentChange
	uiToggleDetail
	uiOpenRight
	uiOpenRightWithShift
	uiCloseRight
	uiToParent
	uiShift
	uiToggleBookmark
	uiErrorMessage
	uiChangeRoot
	uiJumpRefresh
	uiJumpTo
	uiMarkChange
	uiInputChange
)

const (
	columnWidth         int = 30
	expandedColumnWidth int = 80
)

func cwidth(col *column) int {
	if col.expanded {
		return expandedColumnWidth
	}
	return columnWidth
}

func ccwidth() int {
	return cwidth(wo.currentColumn())
}

func last() *FileList {
	return uiLists[len(uiLists)-1]
}

var (
	gui              = make(chan int)
	guiAck           = make(chan bool)
	uiNeedAck        = false
	uiTab            *ui.Tab
	uiCurrent        *ui.Path
	uiIndicator      *ui.Text
	uiIndicatorCover *ui.Text
	uiColumns        *ui.Columns
	uiLists          []*FileList
	uiStatusMessage  *ui.StatusBackup
	uiStatusFilter   *ui.StatusBackup
	uiBookmark       *bookmark
	uiJumpItems      []*ui.Text
	maxColumns       = 5
)

func colorIndicator() *ui.Color { return cfg.color("indicator") }
func colorJump() *ui.Color      { return cfg.color("jump") }
func colorFilter() *ui.Color    { return cfg.color("filter") }

func handleUIEvent(ev int) {
	switch ev {
	case uiErrorMessage:
		uiStatusMessage.Restore().Set(0, message)
	case uiChangeGroup:
		uiTab.SwitchTo(wo.group)
		uiInitColumns()
		updateCurrent()
	case uiColumnContentChange:
		last().update()
	case uiToggleDetail:
		uiColumns.Remove()
		list := last()
		if list.col.expanded {
			idx := uiColumns.Add(ccwidth())
			uiColumns.ClearAt(idx)
		}
		list.update()
		if !list.col.expanded {
			uiColumns.Add(ccwidth())
		}
		updateCurrent()
	case uiChangeSelect:
		last().updateSelect()
		updateFileInfo()
	case uiOpenRight:
		uiColumns.Remove()
		last().update()
		uiColumns.Add(cwidth(last().col))

		idx := uiColumns.Add(ccwidth())
		pp := uiColumns.StartAt(idx)

		ls := NewFileList(pp, wo.currentColumn(), uiColumns.Height-2)
		ls.Draw()
		uiLists = append(uiLists, ls)
		updateCurrent()
	case uiOpenRightWithShift:
		uiInitColumns()
		updateCurrent()
	case uiCloseRight:
		last().Clear()
		uiColumns.Remove()
		uiLists = uiLists[:len(uiLists)-1]
		updateCurrent()
	case uiToParent:
		uiLists[0].update()
		updateCurrent()
	case uiShift:
		uiLists = uiLists[1:]
		redrawColumns()
	case uiJumpTo:
		uiInitColumns()
		updateCurrent()
	case uiChangeRoot:
		uiInitColumns()
		updateCurrent()
	case uiToggleBookmark:
		redrawColumns()
	case uiJumpRefresh:
		refreshJumpItems()
	case uiMarkChange:
		li := last()
		li.update()
		li.updateSelect()
		updateCurrent()
		updateFileInfo()
	case uiInputChange:
		p := uiStatusFilter.Restore().Set(1, inputText)
		termbox.SetCursor(p.X+1, p.Y)
	}
}

func refreshJumpItems() {
	if uiJumpItems != nil {
		for _, v := range uiJumpItems {
			v.Clear()
		}
		uiJumpItems = nil
	}

	for _, v := range jumpItems {
		if len(v.key) == 0 {
			continue
		}
		ji := ui.NewText(v.point, string(v.key))
		ji.Color = colorJump()
		ji.Draw()
		uiJumpItems = append(uiJumpItems, ji)
	}
}

func redrawColumns() {
	uiColumns.RemoveAll()
	if wo.showBookmark {
		p := uiColumns.StartAt(uiColumns.Add2(uiBookmark.width))
		uiBookmark.MoveTo(p)
	}
	for _, v := range uiLists {
		p := uiColumns.StartAt(uiColumns.Add(cwidth(v.col)))
		v.MoveTo(p)
	}
	updateCurrent()
}

func pathItems(path string) []string {
	ts := strings.Split(path, string(filepath.Separator))
	if ts[0] == "" {
		ts[0] = "/"
	}
	if ts[len(ts)-1] == "" {
		ts = ts[:len(ts)-1]
	}
	return ts
}

func updateCurrent() {
	uiCurrent.SetValue(pathItems(wo.currentDir()))
	uiIndicator.Clear()
	uiIndicatorCover.MoveTo(uiIndicator.Start)

	p := last().Start.RightN(ccwidth()/2 - 1)
	p.Y--
	uiIndicator.MoveTo(p)
}

func updateFileInfo() {
	co := wo.currentColumn()
	fi := co.files[co.current]
	m := fmt.Sprintf("%s  %s  %s", fi.ModTime().Format("2006-01-02 15:04:05"), fi.Mode().String(), fi.Name())
	uiStatusMessage.Restore().Set(0, m)
}

func uiInitColumns() {
	uiLists = nil
	uiColumns.RemoveAll()

	if wo.showBookmark {
		p := uiColumns.StartAt(uiColumns.Add2(uiBookmark.width))
		uiBookmark.MoveTo(p)
	}

	for _, v := range wo.currentGroup().columns {
		ii := uiColumns.Add(cwidth(v))
		p := uiColumns.StartAt(ii)
		list := NewFileList(p, v, uiColumns.Height-2)
		list.Draw()
		uiLists = append(uiLists, list)
	}
}

func uiInit() {
	groups := len(wo.groups)
	names := make([]string, groups)
	for i := 0; i < groups; i++ {
		names[i] = fmt.Sprintf(" %d ", i+1)
	}

	uiTab = ui.NewTab(&ui.Point{X: 0, Y: 1}, "", names)
	p := uiTab.Draw()

	uiCurrent = ui.NewPath(p.RightN(2), "", pathItems(wo.currentDir()))
	p = uiCurrent.Draw()

	w, h := termbox.Size()
	p = p.BottomN(2)
	p.X = 0

	uiBookmark = newBookmark(p, h-4)
	uiColumns = ui.NewColumns(p, w, h-4)
	uiColumns.Draw()

	uiInitColumns()
	i := 0
	if wo.showBookmark {
		i = 1
	}
	p = uiColumns.StartAt(i)
	uiIndicator = ui.NewText(&ui.Point{X: p.X + ccwidth()/2 - 1, Y: p.Y - 1}, " ▼ ")
	uiIndicator.Color = colorIndicator()
	uiIndicator.Draw()
	uiIndicatorCover = ui.NewText(uiIndicator.Start, "────")

	ss := ui.NewStatus()
	ss.Draw()

	ss.Add(0)
	uiStatusMessage = ss.Backup()
	si := ss.Add(2)
	si.Color = colorFilter()
	si.SetValue("FILTER ")
	ss.Add(0)
	uiStatusFilter = ss.Backup()
	updateFileInfo()
}

func startEventLoop() {
	for {
		handleUIEvent(<-gui)
		termbox.Flush()
		if uiNeedAck {
			guiAck <- true
		}
	}
}

func uiStart() {
	uiInit()
	termbox.Flush()
	go startEventLoop()
}
