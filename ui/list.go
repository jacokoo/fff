package ui

import (
	"github.com/nsf/termbox-go"
)

// List a list of string
type List struct {
	Selected   int
	Data       []string
	colors     []*Color
	items      []*Text
	colorHints []int
	*Drawable
}

// NewList create a list
func NewList(p *Point, selected int, items []string, colorHints []int) *List {
	is := make([]*Text, len(items))
	for i, v := range items {
		is[i] = NewText(p.BottomN(i), v)
	}

	cs := []*Color{
		{termbox.ColorDefault, termbox.ColorDefault},
		{termbox.ColorCyan, termbox.ColorDefault},
	}
	return &List{selected, items, cs, is, colorHints, NewDrawable(p)}
}

// Draw it
func (l *List) Draw() *Point {
	var maxX = 0
	for i, v := range l.items {
		v.Color = l.colors[l.colorHints[i]]
		if i == l.Selected {
			v.Color = &Color{v.Color.FG, v.Color.BG | termbox.AttrReverse}
		}
		p := v.Draw()
		if p.X > maxX {
			maxX = p.X
		}
	}

	l.End.X = maxX
	l.End.Y = l.Start.Y + len(l.items)
	return l.End
}

// MoveTo update location
func (l *List) MoveTo(p *Point) *Point {
	l.Start = p

	var maxX = 0
	for i, v := range l.items {
		pp := v.MoveTo(p.BottomN(i))
		if pp.X > maxX {
			maxX = pp.X
		}
	}

	l.End.X = maxX
	l.End.Y = l.Start.Y + len(l.items)
	return l.End
}

// Select change the selected item to item
func (l *List) Select(item int) {
	old, new := l.items[l.Selected], l.items[item]
	old.Color = ColorNormal
	new.Color = ColorSelected
	old.Draw()
	new.Draw()
}

// SetData update items
func (l *List) SetData(items []string, hints []int) {
	l.Clear()
	l.Selected = 0
	l.Data = items
	l.colorHints = hints

	l.items = make([]*Text, len(items))
	for i, v := range items {
		l.items[i] = NewText(l.Start.BottomN(i), v)
	}
	l.Draw()
}
