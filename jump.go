package main

import (
	"github.com/jacokoo/fff/ui"
)

type jumpItem struct {
	key    rune
	action func()
	point  *ui.Point
}

func collectList(list *ui.List, fn func(int, *ui.Point)) {
	from, to := list.ItemRange()
	rs := list.ItemRects()
	for i := from; i < to; i++ {
		p := rs[i].Start.RightN(0)
		p.X--
		fn(i, p)
	}
}

func oneCharKey(items []*jumpItem) {

}

func twoCharKey(items []*jumpItem) {

}

func keyThem(items []*jumpItem) {
	count := len(items)
	if count <= 52 { // a-zA-Z
		oneCharKey(items)
		return
	}
}

func collectJumps() []*jumpItem {
	items := make([]*jumpItem, 0)
	for i, v := range uiTab.TabRects() {
		idx := i
		ji := &jumpItem{rune(49 + i), func() {
			wo.changeGroup(idx)
		}, v.Start.Right()}
		items = append(items, ji)
	}

	s := len(items)
	if wo.showBookmark {
		collectList(uiBookmark.list, func(idx int, p *ui.Point) {
			key := uiBookmark.keys[idx]
			items = append(items, &jumpItem{0, func() {
				wo.openRoot(wo.bookmark[key])
			}, p})
		})
	}

	for i, v := range uiLists {
		colIdx := i
		collectList(v.list, func(idx int, p *ui.Point) {
			items = append(items, &jumpItem{0, func() {
				wo.jumpTo(colIdx, idx)
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

func enterJumpMode() {
	mode = ModeJump
}
