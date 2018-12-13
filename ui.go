package main

import (
	"fmt"

	"github.com/jacokoo/fff/ui"
	termbox "github.com/nsf/termbox-go"
)

const (
	uiChangeGroup = iota
	uiAddColumn
	uiRemoveColumn
	uiChangeSelect
	uiColumnContentChange
	uiOpenRight
	uiCloseRight
	uiToParent
	uiShift
	uiToggleBookmark
	uiErrorMessage
	uiChangeRoot
	uiJumpRefresh
	uiJumpTo
)

const (
	columnWidth int = 30
)

var (
	gui              = make(chan int)
	guiAck           = make(chan bool)
	uiNeedAck        = false
	uiTab            *ui.Tab
	uiCurrent        *ui.Label
	uiIndicator      *ui.Text
	uiIndicatorCover *ui.Text
	uiColumns        *ui.Columns
	uiLists          []*FileList
	uiStatus         *ui.Status
	uiBookmark       *bookmark
	uiJumpItems      []*ui.Text
)

func colorIndicator() *ui.Color { return cfg.color("indicator") }
func colorJump() *ui.Color      { return cfg.color("jump") }
func colorFilter() *ui.Color    { return cfg.color("filter") }

func handleUIEvent(ev int) {
	switch ev {
	case uiErrorMessage:
		uiStatus.Set(message)
	case uiChangeGroup:
		uiTab.SwitchTo(wo.group)
		uiLists = nil
		uiInitColumns()
		updateCurrent()
	case uiColumnContentChange:
		uiLists[len(uiLists)-1].update()
	case uiChangeSelect:
		uiLists[len(uiLists)-1].updateSelect()
		updateFileInfo()
	case uiOpenRight:
		idx := uiColumns.Add(columnWidth)
		pp := uiColumns.StartAt(idx)

		ls := NewFileList(pp, wo.currentColumn(), uiColumns.Height-2)
		ls.Draw()
		uiLists = append(uiLists, ls)
		updateCurrent()
	case uiCloseRight:
		uiLists[len(uiLists)-1].Clear()
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
		uiLists = nil
		uiInitColumns()
		updateCurrent()
	case uiChangeRoot:
		uiLists = nil
		uiInitColumns()
		updateCurrent()
	case uiToggleBookmark:
		redrawColumns()
	case uiJumpRefresh:
		refreshJumpItems()
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
		p := uiColumns.StartAt(uiColumns.Add2(20))
		uiBookmark.MoveTo(p)
	}
	for _, v := range uiLists {
		p := uiColumns.StartAt(uiColumns.Add(columnWidth))
		v.MoveTo(p)
	}
	updateCurrent()
}

func updateCurrent() {
	uiCurrent.Clear()
	uiCurrent.SetValue(replaceHome(wo.currentDir())).Draw()
	uiIndicator.Clear()
	uiIndicatorCover.MoveTo(uiIndicator.Start)

	p := uiLists[len(uiLists)-1].Start.RightN(columnWidth/2 - 1)
	p.Y--
	uiIndicator.MoveTo(p)
}

func updateFileInfo() {
	co := wo.currentColumn()
	fi := co.files[co.current]
	uiStatus.Set(fmt.Sprintf("%s  %s  %s", fi.ModTime().Format("2006-01-02 15:04:05"), fi.Mode().String(), fi.Name()))
}

func uiInitColumns() {
	uiColumns.RemoveAll()

	if wo.showBookmark {
		p := uiColumns.StartAt(uiColumns.Add2(20))
		uiBookmark.MoveTo(p)
	}

	for _, v := range wo.currentGroup().columns {
		ii := uiColumns.Add(columnWidth)
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

	uiTab = ui.NewTab(&ui.Point{X: 0, Y: 1}, "TAB", names)
	p := uiTab.Draw()

	uiCurrent = ui.NewLabel(p.RightN(2), "CURRENT", replaceHome(wo.currentDir()))
	p = uiCurrent.Draw()

	w, h := termbox.Size()
	p = p.BottomN(2)
	p.X = 0

	uiBookmark = newBookmark(p, 20, h-4)

	uiColumns = ui.NewColumns(p, w, h-4)
	uiColumns.Draw()

	uiInitColumns()
	i := 0
	if wo.showBookmark {
		i = 1
	}
	p = uiColumns.StartAt(i)
	uiIndicator = ui.NewText(&ui.Point{X: p.X + columnWidth/2 - 1, Y: p.Y - 1}, " ▼ ")
	uiIndicator.Color = colorIndicator()
	uiIndicator.Draw()
	uiIndicatorCover = ui.NewText(uiIndicator.Start, "────")

	uiStatus = ui.NewStatus()
	uiStatus.Draw()
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
