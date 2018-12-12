package ui

import (
	"fmt"

	termbox "github.com/nsf/termbox-go"
)

// Text represent a drawable text
type Text struct {
	Data string
	*Drawable
}

// NewText create a Text
func NewText(p *Point, data string) *Text {
	return &Text{data, NewDrawable(p)}
}

// Draw the text
func (t *Text) Draw() *Point {
	i := 0
	for _, v := range t.Data {
		termbox.SetCell(t.Start.X+i, t.Start.Y, v, t.FG, t.BG)
		i++
	}
	t.End.X = t.Start.X + i - 1
	t.End.Y = t.Start.Y
	return t.End
}

// MoveTo update location
func (t *Text) MoveTo(p *Point) *Point {
	t.Start = p
	return t.Draw()
}

// SetValue set the value
func (t *Text) SetValue(str string) *Text {
	t.Data = str
	return t
}

// Keyed is a container with key
type Keyed struct {
	Key  string
	item Drawer
	*Drawable

	start *Text
	end   *Text
}

// NewKeyed create Keyed
func NewKeyed(p *Point, key string, item Drawer) *Keyed {
	return &Keyed{key, item, NewDrawable(p), NewText(ZeroPoint, ""), NewText(ZeroPoint, "")}
}

// Draw it
func (k *Keyed) Draw() *Point {
	k.start.Data = fmt.Sprintf("%s[", k.Key)
	k.start.Color = ColorKeyword
	e := k.start.MoveTo(k.Start)

	e = k.item.MoveTo(e.Right())
	k.end.Data = "]"
	k.end.Color = ColorKeyword
	k.End = k.end.MoveTo(e.Right())
	return k.End
}

// MoveTo update location
func (k *Keyed) MoveTo(p *Point) *Point {
	k.Start = p
	e := k.start.MoveTo(p)
	e = k.item.MoveTo(e.Right())
	k.End = k.end.MoveTo(e.Right())
	return k.End
}

// Label represent a label
type Label struct {
	*Keyed
	text *Text
}

// NewLabel create Label
func NewLabel(p *Point, key string, value string) *Label {
	s := NewText(ZeroPoint, value)
	return &Label{NewKeyed(p, key, s), s}
}

// SetValue set the value of label
func (l *Label) SetValue(str string) *Label {
	l.text.Data = str
	return l
}
