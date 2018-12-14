package main

import (
	"fmt"

	"github.com/jacokoo/fff/ui"
)

type bookmark struct {
	width, height int
	title         *ui.Text
	line          *ui.HLine
	list          *ui.List
	keys          []string
	*ui.Drawable
}

func bookmarkNames(keys []string) ([]string, []int) {
	ns := make([]string, 0)
	hs := make([]int, 0)

	for _, k := range keys {
		ns = append(ns, fmt.Sprintf(" %s", k))
		hs = append(hs, 1)
	}
	return ns, hs
}

func newBookmark(p *ui.Point, height int) *bookmark {
	t := ui.NewText(p, "BOOKMARKS")
	max := 9
	keys := make([]string, 0)
	for k := range wo.bookmark {
		keys = append(keys, k)
		if len(k) > max {
			max = len(k)
		}
	}
	max += 2
	line := ui.NewHLine(p, max)

	ns, hs := bookmarkNames(keys)
	list := ui.NewList(p, -1, height-4, ns, hs)
	return &bookmark{max, height, t, line, list, keys, ui.NewDrawable(p)}
}

func (b *bookmark) Draw() *ui.Point {
	b.title.MoveTo(b.Start.Bottom().MoveRightN((b.width - 9) / 2))
	b.line.MoveTo(b.Start.BottomN(3))
	b.list.Start = b.Start.BottomN(4)
	b.list.Draw()
	b.End.X = b.Start.X + b.width
	b.End.Y = b.Start.Y + b.height
	return b.End
}

func (b *bookmark) MoveTo(p *ui.Point) *ui.Point {
	b.Start = p
	return b.Draw()
}

func (b *bookmark) update() {
	ns, hs := bookmarkNames(b.keys)
	b.list.SetData(ns, hs, -1)
}
