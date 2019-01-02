package main

import (
	"unicode"

	"github.com/jacokoo/fff/ui"
)

// JumpMode describe the jump mode
type JumpMode struct {
	collect           func() []*ui.JumpItem
	continuousCollect func() []*ui.JumpItem
}

// Collect items for jump
func (mi *JumpMode) Collect() []*ui.JumpItem {
	return mi.collect()
}

// ContinuousCollect collect items for continuous jump
func (mi *JumpMode) ContinuousCollect() []*ui.JumpItem {
	if mi.continuousCollect == nil {
		return nil
	}
	return mi.continuousCollect()
}

// SupportContinuous if continuous jump supported
func (mi *JumpMode) SupportContinuous() bool {
	return mi.continuousCollect != nil
}

// JumpMode
var (
	jumpAll             *JumpMode
	cjumpAll            *JumpMode
	jumpBookmark        *JumpMode
	cjumpBookmark       *JumpMode
	jumpDeleteBookmark  *JumpMode
	cjumpDeleteBookmark *JumpMode
	jumpCurrentDir      *JumpMode
	cjumpCurrentDir     *JumpMode
)

var (
	jump      = make(chan rune)
	jumpQuit  = make(chan bool)
	jumpMode  *JumpMode
	bkMode    Mode
	jumpItems []*ui.JumpItem
)

func init() {
	jumpAll = &JumpMode{func() []*ui.JumpItem {
		its := append(collectBookmark(false), collectCurrentPath()...)
		return append(its, collectAllDir()...)
	}, nil}
	cjumpAll = &JumpMode{jumpAll.collect, collectCurrentDir}

	jumpBookmark = &JumpMode{func() []*ui.JumpItem {
		return collectBookmark(false)
	}, nil}
	cjumpBookmark = &JumpMode{jumpBookmark.collect, collectCurrentDir}

	jumpDeleteBookmark = &JumpMode{func() []*ui.JumpItem {
		return collectBookmark(true)
	}, nil}
	cjumpDeleteBookmark = &JumpMode{jumpDeleteBookmark.collect, jumpDeleteBookmark.collect}

	jumpCurrentDir = &JumpMode{collectCurrentDir, nil}
	cjumpCurrentDir = &JumpMode{collectCurrentDir, collectCurrentDir}
}

func collectAllDir() []*ui.JumpItem {
	items := make([]*ui.JumpItem, 0)
	gr := wo.CurrentGroup()
	ui.EachFileList(func(colIdx int, list *ui.List) {
		items = append(items, list.JumpItems(func(idx int) string {
			return gr.Columns()[colIdx].Files()[idx].Name()
		}, func(idx int) func() bool {
			return func() bool {
				return ac.jumpTo(colIdx, idx, jumpMode.SupportContinuous())
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
				return true
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
			return ac.jumpTo(len(wo.CurrentGroup().Columns())-1, idx, jumpMode.SupportContinuous())
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

func handleJumpResult(item *ui.JumpItem) {
	ui.GuiNeedAck = true
	co := item.Action()
	<-ui.GuiAck
	ui.GuiNeedAck = false

	if !co || !jumpMode.SupportContinuous() {
		quitJumpMode()
		return
	}

	items := jumpMode.ContinuousCollect()
	if len(items) == 0 {
		quitJumpMode()
		return
	}

	keyThem(items)

	changeMode(ModeJump)
	jumpItems = items
	ui.JumpRefreshEvent.Send(items)
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

func enterJumpMode(md *JumpMode) {
	bkMode = mode
	jumpMode = md
	jumpItems = md.Collect()
	keyThem(jumpItems)
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
	jumpMode = nil
	ui.JumpRefreshEvent.Send(jumpItems)
	changeMode(bkMode)
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
	used := make(map[rune]uint)
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

		its[k] = append(is, v)
		uk, ok := used[k]
		if !ok {
			uk = 0
		} else {
			uk++
		}
		used[k] = uk
		v.Key[1] = indexKey(uk)
	}
	for _, v := range its {
		if len(v) == 1 {
			v[0].Key = []rune{v[0].Key[0]}
		}
	}
}
