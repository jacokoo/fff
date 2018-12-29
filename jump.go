package main

import (
	"unicode"

	"github.com/jacokoo/fff/ui"
)

func keyIndex(key rune) uint {
	if key >= 'a' && key <= 'z' {
		return uint(int(key) - int('a'))
	}

	if key >= 'A' && key <= 'Z' {
		return uint(26 + int(key) - int('A'))
	}

	return 52
}

func indexKey(idx uint) rune {
	if idx >= 0 && idx <= 25 {
		return rune('a' + idx)
	}

	if idx > 25 && idx <= 51 {
		return rune('A' + idx - 26)
	}

	return ' '
}

func keyThem(items []*ui.JumpItem) {
	used := make(map[rune]uint64)
	its := make(map[rune][]*ui.JumpItem)
	for _, v := range items {
		k := unicode.ToLower(v.Key[0])
		if (k < 'a' || k > 'z') && (k < '0' || k > '9') {
			k = '-'
		}
		v.Key[0] = k

		is, ok := its[k]
		if !ok {
			is = make([]*ui.JumpItem, 0)
		}

		idx := keyIndex(v.Key[1])
		if idx == 52 {
			v.Key[1] = '-'
		} else {
			us, ok := used[v.Key[0]]
			if !ok {
				us = 1 << 52
			}
			used[v.Key[0]] = us | 1<<idx
		}

		position := -1
		for i, vv := range is {
			if keyIndex(vv.Key[1]) > keyIndex(v.Key[1]) {
				position = i
				break
			}
		}

		if position == -1 {
			is = append(is, v)
		} else {
			is = append(is, nil)
			copy(is[position+1:], is[position:])
			is[position] = v
		}

		its[k] = is
	}
	flatIt(used, its)
}

func next(current *uint, used *uint64) rune {
	for *current < 52 && (*used&(1<<*current)) != 0 {
		*current++
	}
	*used = *used | (1 << *current)
	return indexKey(*current)
}

func flatIt(usedKeys map[rune]uint64, items map[rune][]*ui.JumpItem) {
	for k, v := range items {
		var current uint
		count, used := 0, usedKeys[k]
		vc := len(v)
		if vc == 1 {
			v[0].Key = []rune{v[0].Key[0]}
			continue
		}

		for i := 1; i < vc; i++ {
			if v[i].Key[1] == v[i-1].Key[1] {
				count++
				continue
			}
			if count == 0 {
				continue
			}
			for j := count; j > 0; j-- {
				v[i-j].Key[1] = next(&current, &used)
			}
			count = 0
		}
		if v[vc-1].Key[1] == '-' {
			count++
		}
		if count > 0 {
			for j := count; j > 0; j-- {
				v[vc-j].Key[1] = next(&current, &used)
			}
		}
	}
}

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
	gr := wo.CurrentGroup()
	ui.EachFileList(func(colIdx int, list *ui.List) {
		items = append(items, list.JumpItems(func(idx int) string {
			return gr.Columns()[colIdx].Files()[idx].Name()
		}, func(idx int) func() bool {
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
	return ui.BookmarkList().JumpItems(func(idx int) string {
		return bk.Names[idx]
	}, func(idx int) func() bool {
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
	co := wo.CurrentGroup().Current()
	return ui.CurrentFileList().JumpItems(func(idx int) string {
		return co.Files()[idx].Name()
	}, func(idx int) func() bool {
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
		jumpItems = collectCurrentDir()
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
