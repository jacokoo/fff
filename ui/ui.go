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
	Tab        *Tab
	Path       *Path
	Clip       *Clip
	headerLeft *FlowLayout

	tasks       *Text
	helpMark    *Text
	headerRight *RightAlignFlowLayout

	Column   *Column
	Bookmark *Bookmark
	bkColumn *ColumnItem

	Status        *Status
	StatusMessage *StatusBackup
	StatusInput   *StatusBackup

	jumpItems []*FloatText
	help      *List
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
		co := ui.Column.Add(fl)
		co.showLine = !v.IsShowDetail()
	}

	ui.Column.Draw()
}

func createUI(wo *model.Workspace) {
	w, h := termbox.Size()
	ui.headerLeft = NewFlowLayout(ZeroPoint.Down(), nil)

	names := make([]string, len(wo.Groups))
	for i := range wo.Groups {
		names[i] = fmt.Sprintf(" %d ", i+1)
	}
	ui.Tab = NewTab(ZeroPoint, "", names)
	ui.Path = NewPath(ZeroPoint, "", wo.CurrentGroup().Path())
	ui.Clip = NewClip(ZeroPoint, h)
	ui.headerLeft.Append(ui.Tab, ui.Path, ui.Clip)
	ui.headerLeft.Draw()

	ui.Column = NewColumn(ZeroPoint.DownN(3), w, h-4)
	ui.Bookmark = NewBookmark(ZeroPoint, h-4, wo.Bookmark.Names)
	initFiles(wo.IsShowBookmark(), wo.CurrentGroup())

	ui.headerRight = NewRightAlignFlowLayout(ZeroPoint.Down().RightN(w), nil)
	ui.helpMark = NewText(ZeroPoint, "[?]")
	ui.tasks = NewText(ZeroPoint, "")
	ui.headerRight.Append(ui.tasks, ui.helpMark)
	ui.headerRight.Draw()

	ui.Status = NewStatus(&Point{0, h - 1}, w)
	ui.Status.Draw()

	ui.Status.Add(0)
	ui.StatusMessage = ui.Status.Backup()
	si := ui.Status.Add(2)
	si.Color = colorStatusBarTitle()
	ui.Status.Add(0)
	ui.StatusInput = ui.Status.Backup()

	setFileInfo(wo.CurrentGroup().Current())

	ui.help = NewHelp(h)
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

func redraw() {
	ui.Tab.Draw()
	ui.Path.Draw()
	ui.Clip.Draw()
	ui.helpMark.Draw()
	ui.tasks.Draw()
	ui.Column.Draw()
	ui.StatusMessage.Restore().Set(0, "")
}

// Redraw redraw ui
func (*UI) Redraw() {
	redraw()
	termbox.Flush()
	go startEventLoop()
}

// Recreate UI after resize
func Recreate(wo *model.Workspace) *UI {
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
	createUI(wo)
	termbox.Flush()
	return ui
}
