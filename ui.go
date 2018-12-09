package main

import (
	"github.com/jacokoo/fff/ui"
	termbox "github.com/nsf/termbox-go"
)

const (
	uiChangeGroup = iota
	uiAddColumn
	uiRemoveColumn
	uiChangeSelect
	uiChangeWd
)

type uiContainer struct {
	tab     *ui.Tab
	wd      *ui.Label
	current *ui.Label
	co      *ui.Columns
}

func handleUIEvent(all *uiContainer, ev int) {
	switch ev {
	case uiChangeGroup:
		all.tab.SwitchTo(2)
	case uiChangeWd:
		all.wd.Clear()
		all.current.Clear()
		e := all.wd.SetValue("Hello").Draw()
		all.current.MoveTo(e.RightN(2))
	case uiAddColumn:
		all.co.Clear()
	}

	termbox.Flush()
}

func create() *uiContainer {
	tab := ui.NewTab(&ui.Point{X: 0, Y: 1}, "TAB", []string{
		" 1 ", " 2 ", " 3 ", " 4 ",
	})
	p := tab.Draw()

	ww := ui.NewLabel(p.RightN(2), "WD", wd)
	p = ww.Draw()

	cu := ui.NewLabel(p.RightN(2), "CURRENT", wo.currentDir())
	p = cu.Draw()

	w, h := termbox.Size()
	p = p.BottomN(2)
	p.X = 0

	co := ui.NewColumns(p, w, h-3)
	co.Draw()
	co.Add(30)
	co.Add(40)

	cc := &ui.Color{FG: termbox.ColorBlack, BG: termbox.ColorWhite}

	tx := ui.NewText(co.StartAt(0), "  HelloHelloHelloHello        ")
	tx.Color = cc
	tx.Draw()

	ui.NewText(co.StartAt(1), "  Foo").Draw()

	return &uiContainer{tab, ww, cu, co}
}

func startEventLoop(all *uiContainer) {
loop:
	for {
		select {
		case ev := <-cui:
			handleUIEvent(all, ev)
		case <-cuiQuit:
			break loop
		}
	}
}

func start() {
	all := create()
	go startEventLoop(all)
}
