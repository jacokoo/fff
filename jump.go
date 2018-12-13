package main

import (
	"github.com/jacokoo/fff/ui"
)

var (
	jumpQuit    = make(chan bool)
	jumpToReady = make(chan bool)
	jumpItems   []*jumpItem
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
		p := rs[i].Start.Right()
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

func collectJumps() []*jumpItem {
	items := make([]*jumpItem, 0)
	for i, v := range uiTab.TabRects() {
		idx := i
		ji := &jumpItem{[]rune{rune(49 + i)}, func() bool {
			wo.changeGroup(idx)
			return false
		}, v.Start.Bottom().MoveRight()}
		items = append(items, ji)
	}

	s := len(items)
	if wo.showBookmark {
		collectList(uiBookmark.list, func(idx int, p *ui.Point) {
			key := uiBookmark.keys[idx]
			items = append(items, &jumpItem{nil, func() bool {
				wo.openRoot(wo.bookmark[key])
				return true
			}, p})
		})
	}

	for i, v := range uiLists {
		colIdx := i
		collectList(v.list, func(idx int, p *ui.Point) {
			items = append(items, &jumpItem{nil, func() bool {
				return wo.jumpTo(colIdx, idx)
			}, p})
		})
	}

	e := len(items)
	if s == e {
		return items
	}

	keyThem(items[s:])

	return items
}

func handleJumpResult(item *jumpItem) {
	uiNeedAck = true
	co := item.action()
	<-guiAck
	uiNeedAck = false

	if !co {
		quitJumpMode()
		return
	}

	items := make([]*jumpItem, 0)
	collectList(uiLists[len(uiLists)-1].list, func(idx int, p *ui.Point) {
		items = append(items, &jumpItem{nil, func() bool {
			return wo.jumpTo(len(uiLists)-1, idx)
		}, p})
	})
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

func enterJumpMode() {
	jumpItems = collectJumps()
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
