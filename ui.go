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
	uiChangeWd
	uiChangeSort
)

var (
	gui       = make(chan int)
	uiTab     *ui.Tab
	uiWd      *ui.Label
	uiCurrent *ui.Label
	uiColumns *ui.Columns
	uiList    *ui.List
)

func handleUIEvent(ev int) {
	switch ev {
	case uiChangeGroup:
		uiTab.SwitchTo(2)
	case uiChangeSort:
		ns, hs := fileNames()
		uiList.SetData(ns, hs)
	}

	termbox.Flush()
}

func fileNames() ([]string, []int) {
	col := wo.currentColumn()
	names := make([]string, len(col.files))
	hints := make([]int, len(col.files))
	for i, v := range col.files {
		names[i] = fmt.Sprintf("  %s    %d", v.Name(), v.Size())
		hints[i] = 0
		if v.IsDir() {
			hints[i] = 1
		}
	}
	return names, hints
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
	uiColumns.Add(30)

	p = uiColumns.StartAt(0)
	ns, hs := fileNames()
	uiList = ui.NewList(p, 0, ns, hs)
	uiList.Draw()
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
