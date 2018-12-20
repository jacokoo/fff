package ui

import (
	"fmt"

	"github.com/jacokoo/fff/model"
	termbox "github.com/nsf/termbox-go"
)

var (
	// Gui event chan
	Gui = make(chan Event)

	// GuiQuit chan
	GuiQuit = make(chan bool)

	// GuiAck gui render finish ack
	GuiAck = make(chan bool)

	// GuiNeedAck if need ack
	GuiNeedAck = false

	listeners = make(map[EventType]func(interface{}))
	ui        = new(UI)
)

// UI hold all ui items
type UI struct {
	Tab      *Tab
	Path     *Path
	Column   *Column
	Bookmark *Bookmark
	bkColumn *ColumnItem

	Status        *Status
	StatusMessage *StatusBackup
	StatusInput   *StatusBackup

	jumpItems []*Text
}

func (ui *UI) isShowBookmark() bool {
	return ui.bkColumn == ui.Column.items[0]
}

func (ui *UI) fileCount() int {
	if ui.isShowBookmark() {
		return len(ui.Column.items) - 1
	}

	return len(ui.Column.items)
}

func setFileInfo(co model.Column) {
	m := ""
	fi, err := co.CurrentFile()
	if err == nil {
		m = fmt.Sprintf("%s  %s  %s", fi.ModTime().Format("2006-01-02 15:04:05"), fi.Mode().String(), fi.Name())
	}
	ui.StatusMessage.Restore().Set(0, m)
}

func initFiles(showBookmark bool, g model.Group) {
	ui.Column.RemoveAll()

	if showBookmark {
		ui.bkColumn = ui.Column.Add2(ui.Bookmark)
	}

	for _, v := range g.Columns() {
		fl := newFileList(ZeroPoint, ui.Column.Height-1)
		fl.setData(v)
		ui.Column.Add(fl)
	}

	ui.Column.Draw()
}

func createUI(wo *model.Workspace) {
	p := ZeroPoint.Down()

	names := make([]string, len(wo.Groups))
	for i := range wo.Groups {
		names[i] = fmt.Sprintf(" %d ", i+1)
	}
	ui.Tab = NewTab(p, "", names)
	p = ui.Tab.Draw().Right()

	ui.Path = NewPath(p, "", wo.CurrentGroup().Path())
	ui.Path.Draw()

	p = ZeroPoint.DownN(3)
	w, h := termbox.Size()
	ui.Column = NewColumn(p, w, h-4)
	ui.Bookmark = NewBookmark(ZeroPoint, h-4, wo.Bookmark.Names)
	initFiles(wo.IsShowBookmark(), wo.CurrentGroup())

	ui.Status = NewStatus(&Point{0, h - 1}, w)
	ui.Status.Draw()

	ui.Status.Add(0)
	ui.StatusMessage = ui.Status.Backup()
	si := ui.Status.Add(2)
	si.Color = colorStatusBarTitle()
	ui.Status.Add(0)
	ui.StatusInput = ui.Status.Backup()

	setFileInfo(wo.CurrentGroup().Current())
}

// EachFileList walk through all file list
func EachFileList(fn func(int, *List)) {
	d := 0
	if ui.isShowBookmark() {
		d = 1
	}
	for i, v := range ui.Column.items {
		if v == ui.bkColumn {
			continue
		}
		fn(i-d, v.item.(*FileList).list)
	}
}

// CurrentFileList for jump current list
func CurrentFileList() *List {
	return ui.Column.Last().item.(*FileList).list
}

// BookmarkList for jump bookmark
func BookmarkList() *List {
	return ui.Bookmark.List
}

// Start ui
func Start(wo *model.Workspace) *UI {
	createUI(wo)
	termbox.Flush()
	go startEventLoop()
	return ui
}

// Redraw ui
func Redraw() {
	ui.Tab.Draw()
	ui.Path.Draw()
	ui.Column.Draw()
	ui.StatusMessage.Restore().Set(0, "")
	termbox.Flush()
	go startEventLoop()
}