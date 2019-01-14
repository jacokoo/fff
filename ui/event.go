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

	// TaskChangedEvent Data: TaskManager
	TaskChangedEvent

	// ToggleTaskDetailEvent Data: bool
	ToggleTaskDetailEvent

	// ShowHelpEvent Data: bool
	ShowHelpEvent

	// ToggleClipDetailEvent Data: bool
	ToggleClipDetailEvent

	changeCurrent
)

// Event object
type Event struct {
	Type EventType
	Data interface{}
	Wait bool
}

// Send event
func (et EventType) Send(data interface{}) {
	ev := Event{et, data, true}
	Gui <- ev

	// jump feature need to wait until ui render finished to get the item position
	// maybe it should change a way to implement jump
	<-GuiAck
}

// With create Event
func (et EventType) With(data interface{}) Event {
	return Event{et, data, false}
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
			ui.Path.SetValue(data.(string))
			Redraw(ui.headerLeft)
		},

		ChangeGroupEvent: func(data interface{}) {
			wo := data.(*model.Workspace)
			ui.Tab.SwitchTo(wo.Current)
			initFiles(wo.IsShowBookmark(), wo.CurrentGroup())
			changeCurrent.dispatch(wo.CurrentGroup().Path())
		},

		ColumnContentChangeEvent: func(data interface{}) {
			co := ui.Column.Last()
			co.item.(*FileList).setData(data.(model.Column))
			co.showLine = !data.(model.Column).IsShowDetail()
			Redraw(co)
			ui.Column.resetIndicator()
		},

		ToggleDetailEvent: func(data interface{}) {
			mco := data.(model.Column)
			co := ui.Column.Last()
			co.item.(*FileList).setData(mco)
			co.showLine = !mco.IsShowDetail()
			Redraw(co)
			ui.Column.resetIndicator()
		},

		ChangeSelectEvent: func(data interface{}) {
			mco := data.(model.Column)
			co := ui.Column.Last()
			co.item.(*FileList).setCurrent(mco.Current())
			Redraw(co)
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
			ui.Column.Last().showLine = !cos[len(cos)-2].IsShowDetail()

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
			co.item.(*FileList).setData(mco)
			Redraw(co)
			changeCurrent.dispatch(mco.Path())
		},

		ShiftEvent: func(data interface{}) {
			ui.Column.Shift(ui.isShowBookmark())
			Redraw(ui.Column)
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
			if data.(bool) {
				its := []*ColumnItem{ui.bkColumn}
				ui.Column.items = append(its, ui.Column.items...)
			} else {
				ui.Column.items = ui.Column.items[1:]
			}
			Redraw(ui.Column)
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
			bk := data.(*model.Bookmark)
			ui.Bookmark.SetData(bk.Names)
			Redraw(ui.Column)
		},

		JumpRefreshEvent: func(data interface{}) {
			if ui.jumpItems != nil {
				for _, v := range ui.jumpItems {
					v.Clear()
				}
				ui.jumpItems = nil
			}

			if data == nil {
				return
			}

			for _, v := range data.([]*JumpItem) {
				if len(v.Key) == 0 {
					continue
				}

				t := NewFloatText(v.Point, string(v.Key))
				t.Color = colorJump()
				t.Draw()
				ui.jumpItems = append(ui.jumpItems, t)
			}
		},

		ClipChangedEvent: func(data interface{}) {
			if data == nil || len(data.(model.CopySource)) == 0 {
				ui.Clip.SetData(nil)
				Redraw(ui.headerLeft)

				if ui.Clip.showDetail {
					ui.Clip.Close()
				}
				return
			}

			items := make([]string, 0)
			cs := data.(model.CopySource)
			for _, v := range cs {
				items = append(items, fmt.Sprintf("  %s  ", v.Path()))
			}
			ui.Clip.SetData(items)
			Redraw(ui.headerLeft)

			if ui.Clip.showDetail {
				ui.Clip.Close()
				ui.Clip.Open()
			}
		},

		ToggleClipDetailEvent: func(data interface{}) {
			if len(ui.Clip.items) == 0 {
				return
			}
			if data.(bool) {
				ui.Clip.Open()
			} else {
				ui.Clip.Close()
			}
		},

		BatchEvent: func(data interface{}) {
			for k, v := range data.(map[EventType]interface{}) {
				dispatch(k, v)
			}
		},

		TaskChangedEvent: func(data interface{}) {
			tm := data.(*model.TaskManager)
			ui.Task.SetData(tm.Tasks)
			if len(tm.Tasks) == 0 && ui.Task.showDetail {
				ui.Task.Close()
			}
			Redraw(ui.headerRight)
			if ui.Task.showDetail {
				ui.Task.Close()
				ui.Task.Open()

				if ui.jumpItems != nil {
					for _, v := range ui.jumpItems {
						v.Draw()
					}
				}
			}
		},

		ToggleTaskDetailEvent: func(data interface{}) {
			if len(ui.Task.items) == 0 {
				return
			}
			if data.(bool) {
				ui.Task.Open()
			} else {
				ui.Task.Close()
			}
		},

		ShowHelpEvent: func(data interface{}) {
			termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
			if data.(bool) {
				ui.help.Draw()
			} else {
				redraw()
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
			if ev.Wait {
				GuiAck <- true
			}
		case <-GuiQuit:
			return
		}
	}
}
