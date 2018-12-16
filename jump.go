package main

import (
	"fmt"
	"path/filepath"
	"strings"

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
	jumpQuit     = make(chan bool)
	jumpToReady  = make(chan bool)
	jumpItems    []*jumpItem
	continueJump = false
)

type jumpItem struct {
	key    []rune
	action func() bool
	point  *ui.Point
}

func collectList(list *ui.List, fn func(int, *ui.Point)) {
	from, to := list.ItemRange()
	rs := list.ItemRects()
	for i := from; i < to; i++ {
		p := rs[i-from].Start.Right()
		p.X--
		fn(i, p)
	}
}

func oneCharKey(items []*jumpItem) {
	idx, length := 0, len(items)
	for ch := 'a'; ch < 'z' && idx < length; ch++ {
		items[idx].key = []rune{ch}
		idx++
	}

	if idx < length {
		for ch := 'A'; ch < 'Z' && idx < length; ch++ {
			items[idx].key = []rune{ch}
			idx++
		}
	}
}

func twoCharKey(items []*jumpItem) {
	idx, length := 0, len(items)
	for ch := 'a'; ch < 'z'; ch++ {
		for ch2 := 'a'; ch2 < 'z'; ch2++ {
			if idx >= length {
				return
			}
			items[idx].key = []rune{ch, ch2}
			idx++
		}
	}
}

func keyThem(items []*jumpItem) {
	count := len(items)
	if count <= 52 { // a-zA-Z
		oneCharKey(items)
		return
	}
	twoCharKey(items)
}

func handleJumpResult(item *jumpItem) {
	uiNeedAck = true
	co := item.action()
	<-guiAck
	uiNeedAck = false

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
	gui <- uiJumpRefresh
}

func handleKeys() {
	for {
	sc:
		select {
		case ch := <-jump:
			changeMode(ModeDisabled)
			var got = false
			for _, it := range jumpItems {
				if len(it.key) == 0 {
					continue
				}
				if it.key[0] != ch {
					it.key = nil
					continue
				}
				if len(it.key) == 1 {
					go handleJumpResult(it)
					break sc
				}

				it.key = it.key[1:]
				got = true
			}
			if got {
				gui <- uiJumpRefresh
				changeMode(ModeJump)
			} else {
				go quitJumpMode()
			}
		case <-jumpQuit:
			return
		}
	}
}

func collectAllDir() []*jumpItem {
	items := make([]*jumpItem, 0)
	for i, v := range uiLists {
		colIdx := i
		collectList(v.list, func(idx int, p *ui.Point) {
			items = append(items, &jumpItem{nil, func() bool {
				return wo.jumpTo(colIdx, idx, continueJump)
			}, p})
		})
	}
	return items
}

func collectBookmark(forDelete bool) []*jumpItem {
	items := make([]*jumpItem, 0)
	if !wo.showBookmark {
		return items
	}
	collectList(uiBookmark.list, func(idx int, p *ui.Point) {
		key := bookmarkKeys[idx]
		if forDelete && (key == homeName || key == rootName) {
			return
		}
		fn := func() bool {
			wo.openRoot(bookmarks[key])
			return true
		}
		if forDelete {
			fn = func() bool {
				deleteBookmark(key)
				return false
			}
		}
		items = append(items, &jumpItem{nil, fn, p})
	})
	return items
}

func collectCurrentDir() []*jumpItem {
	items := make([]*jumpItem, 0)
	collectList(uiLists[len(uiLists)-1].list, func(idx int, p *ui.Point) {
		items = append(items, &jumpItem{nil, func() bool {
			return wo.jumpTo(len(uiLists)-1, idx, continueJump)
		}, p})
	})
	return items
}

func collectGroups() []*jumpItem {
	items := make([]*jumpItem, 0)
	for i, v := range uiTab.TabRects() {
		idx := i
		ji := &jumpItem{[]rune{rune(49 + i)}, func() bool {
			wo.changeGroup(idx)
			return false
		}, v.Start.Bottom().MoveRight()}
		items = append(items, ji)
	}
	return items
}

func collectCurrentPath() []*jumpItem {
	items := make([]*jumpItem, 0)
	its := pathItems(wo.currentDir())
	p := ""
	for i, v := range uiCurrent.ItemRects() {
		p += fmt.Sprintf("%c%s", filepath.Separator, its[i])
		pp := v.Start.Bottom()
		if i == 0 {
			pp.X--
		}
		if strings.HasPrefix(p, "//") {
			p = p[1:]
		}

		to := p
		ji := &jumpItem{nil, func() bool {
			wo.openRoot(to)
			return true
		}, pp}
		items = append(items, ji)
	}
	return items
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
	if md != JumpModeDeleteBookmark {
		jumpItems = append(jumpItems, collectGroups()...)
	}
	continueJump = cj

	gui <- uiJumpRefresh
	go handleKeys()
	changeMode(ModeJump)
}

func quitJumpMode() {
	if mode != ModeJump && mode != ModeDisabled {
		return
	}
	jumpQuit <- true
	jumpItems = nil
	gui <- uiJumpRefresh
	changeMode(ModeNormal)
}
