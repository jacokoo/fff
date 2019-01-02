package ui

import (
	"fmt"
)

const (
	bookmarkTitle = "BOOKMARKS"
)

// Bookmark ui
type Bookmark struct {
	Width, Height int
	title         *Text
	line          *HLine
	List          *List
	*Drawable
}

func bookmarkNames(keys []string) ([]string, []int) {
	ns := make([]string, 0)
	hs := make([]int, 0)

	for _, k := range keys {
		ns = append(ns, fmt.Sprintf("  %s", k))
		hs = append(hs, 1)
	}
	return ns, hs
}

// NewBookmark create bookmark
func NewBookmark(p *Point, height int, names []string) *Bookmark {
	t := NewText(p, bookmarkTitle)
	line := NewHLine(p, 0)

	list := NewList(p, -1, height-5, nil, nil)
	bk := &Bookmark{0, height, t, line, list, NewDrawable(p)}
	bk.SetData(names)
	return bk
}

// Draw it
func (b *Bookmark) Draw() *Point {
	b.title.MoveTo(b.Start.Down().MoveRightN((b.Width - 9) / 2))
	b.line.MoveTo(b.Start.DownN(3))
	b.List.Start = b.Start.DownN(4)
	b.List.Draw()
	b.End.X = b.Start.X + b.Width
	b.End.Y = b.Start.Y + b.Height
	return b.End
}

// MoveTo update location
func (b *Bookmark) MoveTo(p *Point) *Point {
	b.Start = p
	return b.Draw()
}

// SetData reset bookmark data
func (b *Bookmark) SetData(names []string) {
	w := len(bookmarkTitle)
	for _, v := range names {
		if l := len(v); l > w {
			w = l
		}
	}
	b.Width = w + 4
	b.line.ChangeWidth(b.Width + 1)
	ns, hs := bookmarkNames(names)
	b.List.SetData(ns, hs, -1)
}
