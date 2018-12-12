package main

import (
	"fmt"

	"github.com/jacokoo/fff/ui"
	termbox "github.com/nsf/termbox-go"
)

// FileList is a list of file
type FileList struct {
	col      *column
	height   int
	list     *ui.List
	filter   *ui.Text
	indicate *ui.Text
	*ui.Drawable
}

// NewFileList create file list
func NewFileList(p *ui.Point, col *column, height int) *FileList {
	h := height - 1
	filter := ui.NewText(p.BottomN(h), "")
	filter.Color = &ui.Color{FG: termbox.ColorBlue, BG: termbox.ColorDefault | termbox.AttrReverse}
	ns, hs := fileNames(col)
	list := ui.NewList(p, col.current, h, ns, hs)
	return &FileList{col, h, list, filter, ui.NewText(p, ""), ui.NewDrawable(p)}
}

// Draw it
func (fl *FileList) Draw() {
	fl.list.Draw()
	fl.End.X = fl.Start.X + columnWidth
	fl.End.Y = fl.Start.Y + fl.height

	fl.updateFilter()
	fl.updateIndicate()
}

// MoveTo update location
func (fl *FileList) MoveTo(p *ui.Point) *ui.Point {
	fl.Start = p
	fl.End.X = fl.Start.X + columnWidth
	fl.End.Y = fl.Start.Y + fl.height
	fl.updateIndicate()
	fl.filter.MoveTo(p.BottomN(fl.height))
	fl.list.MoveTo(p)
	return fl.End
}

func (fl *FileList) updateFilter() {
	s := fl.col.filter
	if len(s) != 0 {
		s = "F: " + s
	}
	fl.filter.Clear()
	fl.filter.SetValue(s).Draw()
}

func (fl *FileList) updateIndicate() {
	data := fmt.Sprintf("[%d/%d]", fl.col.current+1, len(fl.col.files))
	fl.indicate.Clear()
	fl.indicate.SetValue(data).MoveTo(&ui.Point{X: fl.End.X - 2 - len(data), Y: fl.End.Y})
}

func (fl *FileList) update() {
	ns, hs := fileNames(fl.col)
	fl.list.SetData(ns, hs)
	fl.updateIndicate()
	fl.updateFilter()
}

func (fl *FileList) updateSelect() {
	fl.list.Select(fl.col.current)
	fl.updateIndicate()
}
