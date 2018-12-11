package ui

import (
	"github.com/nsf/termbox-go"
)

// List a list of string
type List struct {
	Selected   int
	Height     int
	Data       []string
	colors     []*Color
	items      []*Text
	colorHints []int
	*Drawable
}

// NewList create a list
func NewList(p *Point, selected, height int, items []string, colorHints []int) *List {
	is := make([]*Text, len(items))
	for i, v := range items {
		is[i] = NewText(p.BottomN(i), v)
	}

	cs := []*Color{
		{termbox.ColorDefault, termbox.ColorDefault},
		{termbox.ColorCyan, termbox.ColorDefault},
	}
	return &List{selected, height, items, cs, is, colorHints, NewDrawable(p)}
}

// Draw it
func (l *List) Draw() *Point {
	var maxX = 0
	from, to := 0, l.Height
	if to > len(l.items) {
		to = len(l.items)
	} else {
		delta := l.Selected - l.Height/2
		if delta > 0 {
			to += delta
			from += delta
		}

		if to > len(l.items) {
			delta = to - len(l.items)
			to -= delta
			from -= delta
		}
	}

	j := 0
	for i := from; i < to; i++ {
		v := l.items[i]
		v.Color = l.colors[l.colorHints[i]]
		if i == l.Selected {
			v.Color = v.Color.Reverse()
		}
		p := v.MoveTo(l.Start.BottomN(j))
		if p.X > maxX {
			maxX = p.X
		}
		j++
	}

	l.End.X = maxX - 1
	l.End.Y = l.Start.Y + l.Height
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

	l.End.X = maxX - 1
	l.End.Y = l.Start.Y + l.Height
	return l.End
}

// Select change the selected item to item
func (l *List) Select(item int) {
	old := l.items[l.Selected]
	old.Color = l.colors[l.colorHints[l.Selected]]
	l.Selected = item
	l.Draw()
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
