package ui

const minWidth = 10

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
		is[i] = NewText(p.DownN(i), v)
	}

	cs := []*Color{colorFile(), colorFolder(), colorMarked()}
	return &List{selected, height, items, cs, is, colorHints, 0, 0, NewDrawable(p)}
}

// Draw it
func (l *List) Draw() *Point {
	var maxX = l.Start.X + minWidth
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
		p := v.MoveTo(l.Start.DownN(j))
		if p.X > maxX {
			maxX = p.X
		}
		j++
	}

	l.End.X = maxX
	l.End.Y = l.Start.Y + l.Height - 1
	return l.End
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
}

// SetData update items
func (l *List) SetData(items []string, hints []int, selected int) {
	l.Selected = selected
	l.Data = items
	l.colorHints = hints
	l.from = 0
	l.to = 0

	l.items = make([]*Text, len(items))
	for i, v := range items {
		l.items[i] = NewText(l.Start.DownN(i), v)
	}
}
