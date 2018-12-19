package main

import (
	"github.com/jacokoo/fff/ui"
)

// JumpMode describe the jump mode
type JumpMode uint8

// Jump
const (
	JumpModeAll JumpMode = iota
	JumpModeBookmark
	JumpModeCurrentDir
	JumpModeDeleteBookmark
)

var (
	jump         = make(chan rune)
	jumpQuit     = make(chan bool)
	continueJump = false
	jumpItems    []*ui.JumpItem
)

func oneCharKey(items []*ui.JumpItem) {
	idx, length := 0, len(items)
	for ch := 'a'; ch < 'z' && idx < length; ch++ {
		items[idx].Key = []rune{ch}
		idx++
	}

	if idx < length {
		for ch := 'A'; ch < 'Z' && idx < length; ch++ {
			items[idx].Key = []rune{ch}
			idx++
		}
	}
}

func twoCharKey(items []*ui.JumpItem) {
	idx, length := 0, len(items)
	for ch := 'a'; ch < 'z'; ch++ {
		for ch2 := 'a'; ch2 < 'z'; ch2++ {
			if idx >= length {
				return
			}
			items[idx].Key = []rune{ch, ch2}
			idx++
		}
	}
}

func keyThem(items []*ui.JumpItem) {
	count := len(items)
	if count <= 52 { // a-zA-Z
		oneCharKey(items)
		return
	}
	twoCharKey(items)
}

func handleJumpResult(item *ui.JumpItem) {
	ui.GuiNeedAck = true
	co := item.Action()
	<-ui.GuiAck
	ui.GuiNeedAck = false

	if !co || !continueJump {
		quitJumpMode()
		return
	}

	items := collectCurrentDir()
	if len(items) == 0 {
		quitJumpMode()
		return
	}

	keyThem(items)

	jumpItems = items
	changeMode(ModeJump)
	ui.JumpRefreshEvent.Send(jumpItems)
}

func handleKeys() {
	for {
	sc:
		select {
		case ch := <-jump:
			changeMode(ModeDisabled)
			var got = false
			for _, it := range jumpItems {
				if len(it.Key) == 0 {
					continue
				}
				if it.Key[0] != ch {
					it.Key = nil
					continue
				}
				if len(it.Key) == 1 {
					go handleJumpResult(it)
					break sc
				}

				it.Key = it.Key[1:]
				got = true
			}
			if got {
				ui.JumpRefreshEvent.Send(jumpItems)
				changeMode(ModeJump)
			} else {
				go quitJumpMode()
			}
		case <-jumpQuit:
			return
		}
	}
}

func collectAllDir() []*ui.JumpItem {
	items := make([]*ui.JumpItem, 0)
	ui.EachFileList(func(colIdx int, list *ui.List) {
		items = append(items, list.JumpItems(func(idx int) func() bool {
			return func() bool {
				return ac.jumpTo(colIdx, idx, continueJump)
			}
		})...)
	})
	return items
}

func collectBookmark(forDelete bool) []*ui.JumpItem {
	if !wo.IsShowBookmark() {
		return nil
	}
	bk := wo.Bookmark
	return ui.BookmarkList().JumpItems(func(idx int) func() bool {
		key := bk.Names[idx]
		if forDelete && bk.IsFixed(key) {
			return nil
		}
		fn := func() bool {
			v, has := bk.Get(key)
			if !has {
				return false
			}
			ac.openRoot(v)
			return true
		}
		if forDelete {
			fn = func() bool {
				ac.deleteBookmark(key)
				return false
			}
		}
		return fn
	})
}

func collectCurrentDir() []*ui.JumpItem {
	return ui.CurrentFileList().JumpItems(func(idx int) func() bool {
		return func() bool {
			return ac.jumpTo(len(wo.CurrentGroup().Columns())-1, idx, continueJump)
		}
	})
}

func collectGroups() []*ui.JumpItem {
	return gui.Tab.JumpItems(func(idx int) func() bool {
		return func() bool {
			ac.changeGroup(idx)
			return false
		}
	})
}

func collectCurrentPath() []*ui.JumpItem {
	return gui.Path.JumpItems(func(path string) func() bool {
		if path == "/" {
			return nil
		}
		return func() bool {
			ac.openRoot(path)
			return true
		}
	})
}

func enterJumpMode(md JumpMode, cj bool) {
	switch md {
	case JumpModeBookmark:
		jumpItems = collectBookmark(false)
	case JumpModeDeleteBookmark:
		jumpItems = collectBookmark(true)
	case JumpModeCurrentDir:
		jumpItems = append(collectCurrentPath(), collectCurrentDir()...)
	case JumpModeAll:
		jumpItems = append(collectBookmark(false), collectCurrentPath()...)
		jumpItems = append(jumpItems, collectAllDir()...)
	}
	keyThem(jumpItems)
	continueJump = cj

	ui.JumpRefreshEvent.Send(jumpItems)
	go handleKeys()
	changeMode(ModeJump)
}

func quitJumpMode() {
	if mode != ModeJump && mode != ModeDisabled {
		return
	}
	jumpQuit <- true
	jumpItems = nil
	ui.JumpRefreshEvent.Send(jumpItems)
	changeMode(ModeNormal)
}
