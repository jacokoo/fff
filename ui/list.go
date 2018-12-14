package ui

// List a list of string
type List struct {
	Selected   int
	Height     int
	Data       []string
	colors     []*Color
	items      []*Text
	colorHints []int
	from, to   int
	*Drawable
}

// NewList create a list
func NewList(p *Point, selected, height int, items []string, colorHints []int) *List {
	is := make([]*Text, len(items))
	for i, v := range items {
		is[i] = NewText(p.BottomN(i), v)
	}

	cs := []*Color{colorFile(), colorFolder(), colorMarked()}
	return &List{selected, height, items, cs, is, colorHints, 0, 0, NewDrawable(p)}
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

	l.from = from
	l.to = to
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

// Clear it
func (l *List) Clear() {
	for _, v := range l.items {
		v.Clear()
	}
}

// MoveTo update location
func (l *List) MoveTo(p *Point) *Point {
	l.Start = p
	return l.Draw()
}

// Select change the selected item to item
func (l *List) Select(item int) {
	if len(l.items) == 0 {
		return
	}
	old := l.items[l.Selected]
	old.Color = l.colors[l.colorHints[l.Selected]]
	l.Selected = item
	l.Draw()
}

// SetData update items
func (l *List) SetData(items []string, hints []int, selected int) {
	l.Clear()
	l.Selected = selected
	l.Data = items
	l.colorHints = hints

	l.items = make([]*Text, len(items))
	for i, v := range items {
		l.items[i] = NewText(l.Start.BottomN(i), v)
	}
	l.Draw()
}

// ItemRange the items showed
func (l *List) ItemRange() (int, int) {
	return l.from, l.to
}

// ItemRects return the showed item rects
func (l *List) ItemRects() []*Rect {
	rs := make([]*Rect, 0)
	for i := l.from; i < l.to; i++ {
		rs = append(rs, l.items[i].Rect)
	}
	return rs
}
