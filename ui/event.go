package ui

import (
	"fmt"

	"github.com/jacokoo/fff/model"
	termbox "github.com/nsf/termbox-go"
)

// EventType event type
type EventType uint8

const (
	// BatchEvent multiple message
	BatchEvent EventType = iota

	// MessageEvent Data: string
	MessageEvent

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

	// ClipChangedEvent clip changed Data CopySource
	ClipChangedEvent

	changeCurrent
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

// With create Event
func (et EventType) With(data interface{}) Event {
	return Event{et, data}
}

func (et EventType) dispatch(data interface{}) {
	dispatch(et, data)
}

// Batch send batch events
func Batch(ev ...Event) {
	mp := make(map[EventType]interface{}, len(ev))
	for _, v := range ev {
		mp[v.Type] = v.Data
	}
	BatchEvent.Send(mp)
}

func dispatch(e EventType, data interface{}) {
	h, has := handlers[e]
	if has {
		h(data)
	}
}

var handlers = make(map[EventType]func(interface{}))

func init() {
	for k, v := range map[EventType]func(interface{}){
		MessageEvent: func(data interface{}) {
			ui.StatusMessage.Restore().Set(0, data.(string))
		},

		changeCurrent: func(data interface{}) {
			ui.Clip.Clear()
			ui.Path.Clear()
			ui.Path.SetValue(data.(string))
			p := ui.Path.Draw()
			if ui.Clip.Data != "" {
				ui.Clip.MoveTo(p.Right())
			}
		},

		ChangeGroupEvent: func(data interface{}) {
			wo := data.(*model.Workspace)
			ui.Tab.SwitchTo(wo.Current)
			initFiles(wo.IsShowBookmark(), wo.CurrentGroup())
			changeCurrent.dispatch(wo.CurrentGroup().Path())
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
			changeCurrent.dispatch(g.Current().Path())
		},

		CloseRightEvent: func(data interface{}) {
			co := data.(model.Column)
			ui.Column.Remove()
			setFileInfo(co)
			ui.Column.resetIndicator()
			changeCurrent.dispatch(co.Path())
		},

		ToParentEvent: func(data interface{}) {
			co := ui.Column.Last()
			mco := data.(model.Column)
			co.Clear()
			co.item.(*FileList).setData(mco)
			co.Draw()
			changeCurrent.dispatch(mco.Path())
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
			changeCurrent.dispatch(g.Current().Path())
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

		ClipChangedEvent: func(data interface{}) {
			ui.Clip.Clear()
			ui.Clip.Data = ""
			if data != nil {
				ui.Clip.SetData(fmt.Sprintf("%d clips", len(data.(model.CopySource))))
				ui.Clip.Draw()
			}
		},

		BatchEvent: func(data interface{}) {
			for k, v := range data.(map[EventType]interface{}) {
				dispatch(k, v)
			}
		},
	} {
		handlers[k] = v
	}
}

func startEventLoop() {
	for {
		select {
		case ev := <-Gui:
			dispatch(ev.Type, ev.Data)
			termbox.Flush()
			if GuiNeedAck {
				GuiAck <- true
			}
		case <-GuiQuit:
			return
		}
	}
}
