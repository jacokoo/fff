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
	uiChangeWd
	uiColumnContentChange
	uiOpenRight
	uiCloseRight
	uiToParent
	uiShift
	uiToggleBookmark
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
)

func handleUIEvent(ev int) {
	switch ev {
	case uiChangeGroup:
		uiTab.SwitchTo(2)
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
	case uiToggleBookmark:
		redrawColumns()
	}

	termbox.Flush()
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
	list := ui.NewList(p, 0, h, ns, hs)
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

type bookmark struct {
	width, height int
	title         *ui.Text
	line          *ui.HLine
	list          *ui.List
	*ui.Drawable
}

func bookmarkNames() ([]string, []int) {
	ns := make([]string, 0)
	hs := make([]int, 0)

	for k := range wo.bookmark {
		ns = append(ns, fmt.Sprintf("  %s", k))
		hs = append(hs, 1)
	}
	return ns, hs
}

func newBookmark(p *ui.Point, width, height int) *bookmark {
	t := ui.NewText(p, "BOOKMARKS[bb]")
	line := ui.NewHLine(p, width)

	ns, hs := bookmarkNames()
	list := ui.NewList(p, -1, height-4, ns, hs)
	return &bookmark{width, height, t, line, list, ui.NewDrawable(p)}
}

func (b *bookmark) Draw() *ui.Point {
	b.title.MoveTo(b.Start.Bottom().MoveRightN((b.width - 13) / 2))
	b.line.MoveTo(b.Start.BottomN(3))
	b.list.Start = b.Start.BottomN(4)
	b.list.Draw()
	b.End.X = b.Start.X + b.width
	b.End.Y = b.Start.Y + b.height
	return b.End
}

func (b *bookmark) MoveTo(p *ui.Point) *ui.Point {
	b.Start = p
	return b.Draw()
}

func (b *bookmark) update() {
	ns, hs := bookmarkNames()
	b.list.SetData(ns, hs)
}
