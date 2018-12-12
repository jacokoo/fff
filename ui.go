package main

import (
	"fmt"
	"io/ioutil"
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
	uiOpenRight
	uiCloseRight
	uiToParent
	uiShift
	uiToggleBookmark
	uiErrorMessage
	uiChangeRoot
	uiJumpRefresh
)

const (
	columnWidth int = 30
)

var (
	gui              = make(chan int)
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

func handleUIEvent(ev int) {
	switch ev {
	case uiErrorMessage:
		uiStatus.Set(message)
	case uiChangeGroup:
		uiTab.SwitchTo(wo.group)
		uiLists = make([]*FileList, 0)
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
	case uiChangeRoot:
		fallthrough
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
	color := &ui.Color{FG: termbox.ColorRed, BG: termbox.ColorDefault | termbox.AttrReverse}

	for _, v := range jumpItems {
		if len(v.key) == 0 {
			continue
		}
		ji := ui.NewText(v.point, string(v.key))
		ji.Color = color
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

func formatSize(size int64) string {
	unit := "B"
	b := float32(size)

	if b > 1024 {
		unit = "K"
		b = b / 1024
	} else {
		return fmt.Sprintf("%dB", size)
	}

	if b > 1024 {
		unit = "M"
		b = b / 1024
	}

	if b > 1024 {
		unit = "G"
		b = b / 1024
	}
	return fmt.Sprintf("%.2f%s", b, unit)
}

func fileNames(col *column) ([]string, []int) {
	names := make([]string, len(col.files))
	hints := make([]int, len(col.files))
	for i, v := range col.files {
		na := v.Name()
		si := formatSize(v.Size())
		if v.IsDir() {
			fs, _ := ioutil.ReadDir(filepath.Join(col.path, na))
			si = fmt.Sprintf("%d it.", len(fs))
		}

		re := columnWidth - len(si) - 4

		if len(na) >= re {
			na = na[0:re-3] + "..."
		}
		re -= len(na)
		if re < 0 {
			re = 0
		}

		names[i] = fmt.Sprintf("  %s%s%s  ", na, strings.Repeat(" ", re), si)
		hints[i] = 0
		if v.IsDir() {
			hints[i] = 1
		}
	}
	return names, hints
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
	uiIndicator.Color = &ui.Color{FG: termbox.ColorGreen, BG: termbox.ColorDefault}
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
	}
}

func uiStart() {
	uiInit()
	termbox.Flush()
	go startEventLoop()
}
