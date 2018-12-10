package main

import (
	"fmt"
	"strings"

	"github.com/jacokoo/fff/ui"
	termbox "github.com/nsf/termbox-go"
)

const (
	uiChangeGroup = iota
	uiAddColumn
	uiRemoveColumn
	uiChangeSelect
	uiChangeWd
	uiChangeSort
)

const (
	columnWidth int = 30
)

var (
	gui       = make(chan int)
	uiTab     *ui.Tab
	uiWd      *ui.Label
	uiCurrent *ui.Label
	uiColumns *ui.Columns
	uiLists   []*ui.List
)

func handleUIEvent(ev int) {
	switch ev {
	case uiChangeGroup:
		uiTab.SwitchTo(2)
	case uiChangeSort:
		ns, hs := fileNames(wo.currentColumn())
		uiLists[len(uiLists)-1].SetData(ns, hs)
	}

	termbox.Flush()
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
		re := columnWidth - len(si) - 4

		if len(na) >= re {
			na = na[0:len(na)-5] + "..."
		}
		re -= len(na)

		names[i] = fmt.Sprintf("  %s%s%s  ", na, strings.Repeat(" ", re), si)
		hints[i] = 0
		if v.IsDir() {
			hints[i] = 1
		}
	}
	return names, hints
}

func uiInitColumns() {
	uiColumns.RemoveAll()
	uiLists = uiLists[0:]

	for _, v := range wo.currentGroup().columns {
		ii := uiColumns.Add(columnWidth)
		p := uiColumns.StartAt(ii)

		ns, hs := fileNames(v)
		list := ui.NewList(p, 0, ns, hs)
		list.Draw()
		uiLists = append(uiLists, list)
	}
}

func uiInit() {
	groups := len(wo.groups)
	names := make([]string, groups)
	for i := 0; i < groups; i++ {
		names[i] = fmt.Sprintf(" %d ", i)
	}

	uiTab = ui.NewTab(&ui.Point{X: 0, Y: 1}, "TAB", names)
	p := uiTab.Draw()

	uiWd = ui.NewLabel(p.RightN(2), "WD", replaceHome(wd))
	p = uiWd.Draw()

	uiCurrent = ui.NewLabel(p.RightN(2), "CURRENT", replaceHome(wo.currentDir()))
	p = uiCurrent.Draw()

	w, h := termbox.Size()
	p = p.BottomN(2)
	p.X = 0

	uiColumns = ui.NewColumns(p, w, h-3)
	uiColumns.Draw()

	uiInitColumns()
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

// FileList is a list of file
type FileList struct {
	col      *column
	ui       *ui.List
	filter   *ui.Text
	indicate *ui.Text
}

// NewFileList create file list
func NewFileList(p *ui.Point, col column) *FileList {
	return nil
}
