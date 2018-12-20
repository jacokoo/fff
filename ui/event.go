package ui

import (
	"fmt"

	"github.com/jacokoo/fff/model"
	termbox "github.com/nsf/termbox-go"
)

// EventType event type
type EventType uint8

const (
	// MessageEvent Data: string
	MessageEvent EventType = iota

	// ChangeGroupEvent Data: model.Workspace
	ChangeGroupEvent

	// ColumnContentChangeEvent Data: model.Column the current column
	ColumnContentChangeEvent

	// ToggleDetailEvent Data: model.Column the current column
	ToggleDetailEvent

	// ChangeSelectEvent Data: model.Column the current column
	ChangeSelectEvent

	// OpenRightEvent Data: model.Group current group
	OpenRightEvent

	// CloseRightEvent Data: model.Column, the new current colunn
	CloseRightEvent

	// ToParentEvent Data: model.Column, the current column
	ToParentEvent

	// ShiftEvent Data: model.Group
	ShiftEvent

	// JumpToEvent Data: model.Group
	JumpToEvent

	// ChangeRootEvent Data: model.Group
	ChangeRootEvent

	// ToggleBookmarkEvent Data: bool if to show bookmark
	ToggleBookmarkEvent

	// InputChangeEvent Data: [name string, value string]
	InputChangeEvent

	// QuitInputEvent Data: model.Column
	QuitInputEvent

	// BookmarkChangedEvent Data: model.Bookmark
	BookmarkChangedEvent

	// JumpRefreshEvent Data:
	JumpRefreshEvent
)

// Event object
type Event struct {
	Type EventType
	Data interface{}
}

// Send event
func (et EventType) Send(data interface{}) {
	ev := Event{et, data}
	Gui <- ev
}

var handlers = map[EventType]func(interface{}){
	MessageEvent: func(data interface{}) {
		ui.StatusMessage.Restore().Set(0, data.(string))
	},

	ChangeGroupEvent: func(data interface{}) {
		wo := data.(*model.Workspace)
		ui.Tab.SwitchTo(wo.Current)
		initFiles(wo.IsShowBookmark(), wo.CurrentGroup())
		ui.Path.SetValue(wo.CurrentGroup().Path())
	},

	ColumnContentChangeEvent: func(data interface{}) {
		co := ui.Column.Last()
		co.Clear()
		co.item.(*FileList).setData(data.(model.Column))
		co.Draw()
		ui.Column.resetIndicator()
	},

	ToggleDetailEvent: func(data interface{}) {
		mco := data.(model.Column)
		co := ui.Column.Last()
		co.Clear()
		co.item.(*FileList).setData(mco)
		co.showLine = !mco.IsShowDetail()
		co.Draw()
		ui.Column.resetIndicator()
	},

	ChangeSelectEvent: func(data interface{}) {
		mco := data.(model.Column)
		co := ui.Column.Last()
		co.Clear()
		co.item.(*FileList).setCurrent(mco.Current())
		co.Draw()
		setFileInfo(mco)
	},

	OpenRightEvent: func(data interface{}) {
		g := data.(model.Group)
		cos := g.Columns()
		ui.Column.Clear()

		if len(cos) == ui.fileCount() {
			ui.Column.Shift(ui.isShowBookmark())
		}

		last := ui.Column.Last().item.(*FileList)
		last.setData(cos[len(cos)-2])

		fl := newFileList(ZeroPoint, ui.Column.Height-1)
		fl.setData(g.Current())
		ui.Column.Add(fl)

		ui.Column.Draw()
		ui.Path.SetValue(g.Current().Path())
	},

	CloseRightEvent: func(data interface{}) {
		co := data.(model.Column)
		ui.Column.Remove()
		setFileInfo(co)
		ui.Column.resetIndicator()
		ui.Path.SetValue(co.Path())
	},

	ToParentEvent: func(data interface{}) {
		co := ui.Column.Last()
		mco := data.(model.Column)
		co.Clear()
		co.item.(*FileList).setData(mco)
		co.Draw()
		ui.Path.SetValue(mco.Path())
	},

	ShiftEvent: func(data interface{}) {
		ui.Column.Clear()
		ui.Column.Shift(ui.isShowBookmark())
		ui.Column.Draw()
	},

	JumpToEvent: func(data interface{}) {
		initFiles(ui.isShowBookmark(), data.(model.Group))
	},

	ChangeRootEvent: func(data interface{}) {
		g := data.(model.Group)
		initFiles(ui.isShowBookmark(), g)
		ui.Path.SetValue(g.Current().Path())
	},

	ToggleBookmarkEvent: func(data interface{}) {
		ui.Column.Clear()
		if data.(bool) {
			its := []*ColumnItem{ui.bkColumn}
			ui.Column.items = append(its, ui.Column.items...)
		} else {
			ui.Column.items = ui.Column.items[1:]
		}
		ui.Column.Draw()
	},

	InputChangeEvent: func(data interface{}) {
		ss := data.([]string)
		st := ui.StatusInput.Restore()
		st.Set(0, fmt.Sprintf(" %s ", ss[0]))
		p := st.Set(1, ss[1])
		x := p.X
		if len(ss[1]) != 0 {
			x++
		}
		termbox.SetCursor(x, p.Y)
	},

	QuitInputEvent: func(data interface{}) {
		termbox.SetCursor(-1, -1)
		setFileInfo(data.(model.Column))
	},

	BookmarkChangedEvent: func(data interface{}) {
		ui.Column.Clear()
		bk := data.(*model.Bookmark)
		ui.Bookmark.SetData(bk.Names)
		ui.Column.Draw()
	},

	JumpRefreshEvent: func(data interface{}) {
		if ui.jumpItems != nil {
			for _, v := range ui.jumpItems {
				v.Clear()
			}
			ui.jumpItems = nil
		}

		for _, v := range data.([]*JumpItem) {
			if len(v.Key) == 0 {
				continue
			}

			t := NewText(v.Point, string(v.Key))
			t.Color = colorJump()
			t.Draw()
			ui.jumpItems = append(ui.jumpItems, t)
		}
	},
}

func startEventLoop() {
	for {
		select {
		case ev := <-Gui:
			h, has := handlers[ev.Type]
			if !has {
				break
			}

			h(ev.Data)
			termbox.Flush()
			if GuiNeedAck {
				GuiAck <- true
			}
		case <-GuiQuit:
			return
		}
	}
}
