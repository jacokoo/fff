package main

import (
	"fmt"
	"strings"

	"github.com/jacokoo/fff/ui"
	"github.com/nsf/termbox-go"
)

const (
	uiChangeGroup = iota
	uiAddColumn
	uiRemoveColumn
	uiChangeSelect
	uiChangeWd
)

type drawable interface {
	draw() (int, int)
	clear()
	updateXY(x, y int) (int, int)
	updateData(data interface{}) (int, int)
	update(x, y int, data interface{}) (int, int)
}

type rect struct {
	x, y, lastX, lastY int
}

type uiRect struct {
	*rect
	data   string
	fg, bg termbox.Attribute
}

func (r *uiRect) clear() {
	for i := 0; i < len(r.data); i++ {
		termbox.SetCell(r.x+i, r.y, ' ', termbox.ColorDefault, termbox.ColorDefault)
	}
}

func (r *uiRect) draw() (int, int) {
	i := 0
	for _, v := range r.data {
		termbox.SetCell(r.x+i, r.y, v, r.fg, r.bg)
		i++
	}
	r.lastX, r.lastY = r.x+i, r.y
	return r.lastX, r.lastY
}

func (r *uiRect) redraw() (int, int) {
	r.clear()
	return r.draw()
}

func (r *uiRect) update(x, y int, data interface{}) (int, int) {
	if x == r.x && y == r.y && data == r.data {
		return r.lastX, r.lastY
	}
	r.clear()
	r.x = x
	r.y = y
	r.data = data.(string)
	return r.draw()
}

func (r *uiRect) updateXY(x, y int) (int, int) {
	return r.update(x, y, r.data)
}

func (r *uiRect) updateData(data interface{}) (int, int) {
	return r.update(r.x, r.y, data)
}

func (r *uiRect) toNormal() *uiRect {
	r.fg = termbox.ColorDefault
	r.bg = termbox.ColorDefault
	return r
}

func (r *uiRect) toKeyword() *uiRect {
	r.fg = termbox.ColorGreen
	r.bg = termbox.ColorDefault
	return r
}

func (r *uiRect) toSelected() *uiRect {
	r.fg = termbox.ColorBlack
	r.bg = termbox.ColorBlue
	return r
}

func newUIRect(x, y int, data string) *uiRect {
	return &uiRect{&rect{x, y, 0, 0}, data, termbox.ColorDefault, termbox.ColorDefault}
}

type vline struct {
	*uiRect
	height int
}

func (v *vline) draw() (int, int) {
	for i := 0; i < v.height; i++ {
		j := 0
		for _, vv := range v.data {
			termbox.SetCell(v.x+j, v.y+i, vv, v.fg, v.bg)
			j++
		}
	}
	return v.x + len(v.data), v.y + v.height
}

func (v *vline) update(x, y int, height interface{}) (int, int) {
	v.clear()
	v.x = x
	v.y = y
	v.height = height.(int)
	return v.draw()
}

type uiTab struct {
	*rect
	count, current int
	items          []drawable
	numbers        []uiRect
}

func newUITab(x, y int) *uiTab {
	count := len(wo.groups)
	items := make([]drawable, 0, count+2)
	numbers := make([]uiRect, count)

	return &uiTab{&rect{x, y, 0, 0}, count, wo.currentGroup, items, numbers}
}

func (t *uiTab) draw() (int, int) {
	s := newUIRect(t.x, t.y, "TABS[").toKeyword()
	t.items = append(t.items, s)
	xx, yy := s.draw()

	for i := 0; i < t.count; i++ {
		s := newUIRect(xx+1, yy, fmt.Sprintf("%d", i+1))
		if i == t.current {
			s.toSelected()
		} else {
			s.toNormal()
		}
		t.numbers[i] = *s
		t.items = append(t.items, s)

		xx, yy = s.draw()
	}

	s = newUIRect(xx+1, yy, "]").toKeyword()
	t.items = append(t.items, s)
	t.lastX, t.lastY = s.draw()
	return t.lastX, t.lastY
}

func (t *uiTab) clear() {
	for _, v := range t.items {
		v.clear()
	}
}

func (t *uiTab) update(x, y int, data interface{}) (int, int) {
	if x == t.x && y == t.y && data == t.current {
		return t.lastX, t.lastY
	}

	if x == t.x && y == t.y {
		old := t.numbers[t.current].toNormal()
		ne := t.numbers[data.(int)].toSelected()

		old.redraw()
		ne.redraw()

		t.current = data.(int)
		return t.lastX, t.lastY
	}

	t.clear()
	return t.draw()
}

func (t *uiTab) updateXY(x, y int) (int, int) {
	return t.update(x, y, t.current)
}

func (t *uiTab) updateData(data interface{}) (int, int) {
	return t.update(t.x, t.y, data)
}

type uiKeyValue struct {
	*rect
	data  string
	items []drawable
	value *uiRect
}

func newUIKeyValue(x, y int, data string) *uiKeyValue {
	return &uiKeyValue{&rect{x, y, 0, 0}, data, make([]drawable, 0, 3), new(uiRect)}
}

func (w *uiKeyValue) draw() (int, int) {
	s := newUIRect(w.x, w.y, "WD[").toKeyword()
	w.items = append(w.items, s)
	xx, yy := s.draw()

	s = newUIRect(xx+1, yy, w.data)
	w.items = append(w.items, s)
	xx, yy = s.draw()
	w.value = s

	s = newUIRect(xx+1, yy, "]").toKeyword()
	w.items = append(w.items, s)
	w.lastX, w.lastY = s.draw()

	return w.lastX, w.lastY
}

func (w *uiKeyValue) clear() {
	for _, v := range w.items {
		v.clear()
	}
}

func (w *uiKeyValue) updateXY(x, y int) (int, int) {
	return w.update(x, y, w.data)
}

func (w *uiKeyValue) updateData(data interface{}) (int, int) {
	return w.update(w.x, w.y, data)
}

func (w *uiKeyValue) update(x, y int, data interface{}) (int, int) {
	w.clear()
	w.x, w.y = x, y
	return w.draw()
}

type uiContainer struct {
	tab     *ui.Tab
	wd      *ui.Label
	current *ui.Label
	// line    *uiRect
}

func handleUIEvent(all *uiContainer, ev int) {
	switch ev {
	case uiChangeGroup:
		all.tab.UpdateData(2)
	case uiChangeWd:
		e := all.wd.UpdateData("Hello")
		all.current.UpdateXY(e.RightN(2))
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

	w, _ := termbox.Size()

	line := newUIRect(0, p.Bottom().Y, strings.Repeat("─", w))
	line.draw()

	return &uiContainer{tab, ww, cu}
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
