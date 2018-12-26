package ui

import (
	"fmt"

	runewidth "github.com/mattn/go-runewidth"
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
		i += runewidth.RuneWidth(v)
	}
	t.End = t.Start.RightN(i)
	if i > 0 {
		t.End.MoveLeft()
	}
	return t.End
}

// MoveTo update location
func (t *Text) MoveTo(p *Point) *Point {
	t.Start = p
	return t.Draw()
}

// Keyed is a container with key
type Keyed struct {
	Key  string
	item Drawer
	*Drawable

	left  *Text
	right *Text
}

// NewKeyed create Keyed
func NewKeyed(p *Point, key string, item Drawer) *Keyed {
	return &Keyed{key, item, NewDrawable(p), NewText(ZeroPoint, ""), NewText(ZeroPoint, "")}
}

// Draw it
func (k *Keyed) Draw() *Point {
	k.left.Data = fmt.Sprintf("%s[", k.Key)
	k.left.Color = colorKeyword()
	e := k.left.MoveTo(k.Start)

	e = k.item.MoveTo(e.Right())
	k.right.Data = "]"
	k.right.Color = colorKeyword()
	k.End = k.right.MoveTo(e.Right())
	return k.End
}

// MoveTo update location
func (k *Keyed) MoveTo(p *Point) *Point {
	k.Start = p
	e := k.left.MoveTo(p)
	e = k.item.MoveTo(e.Right())
	k.End = k.right.MoveTo(e.Right())
	return k.End
}

// Label key value
type Label struct {
	Data string
	text *Text
	*Keyed
}

// NewLabel create label
func NewLabel(p *Point, name, data string) *Label {
	text := NewText(ZeroPoint, data)
	return &Label{data, text, NewKeyed(p, name, text)}
}

// SetData set data
func (l *Label) SetData(data string) {
	l.Data = data
	l.text.Data = data
}
