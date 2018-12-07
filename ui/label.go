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

// Draw the text
func (t *Text) Draw() *Point {
	i := 0
	for _, v := range t.Data {
		termbox.SetCell(t.Start.X+i, t.Start.Y, v, t.fg, t.bg)
		i++
	}
	t.End.X = t.Start.X + i
	t.End.Y = t.Start.Y
	return t.End
}

// Update it
func (t *Text) Update(p *Point, data interface{}) *Point {
	t.Start = p
	t.Data = data.(string)
	return t.Draw()
}

// UpdateXY is used to update the location of the text
func (t *Text) UpdateXY(p *Point) *Point {
	return t.Update(p, t.Data)
}

// UpdateData is only update the text
func (t *Text) UpdateData(data interface{}) *Point {
	return t.Update(t.Start, data)
}

// NewText create a Text
func NewText(p *Point, data string) *Text {
	return &Text{data, NewDrawable(p)}
}

// Keyed is a container with key
type Keyed struct {
	Key  string
	Item Drawer
	*Drawable

	start *Text
	end   *Text
}

// NewKeyed create Keyed
func NewKeyed(p *Point, key string, item Drawer) *Keyed {
	return &Keyed{key, item, NewDrawable(p), nil, nil}
}

// Draw it
func (k *Keyed) Draw() *Point {
	k.start = NewText(k.Start, fmt.Sprintf("%s[", k.Key))
	k.start.Color = ColorKeyword
	e := k.start.Draw()

	e = k.Item.UpdateXY(e.RightN(0))
	k.end = NewText(e.RightN(0), "]")
	k.end.Color = ColorKeyword

	e = k.end.Draw()
	k.End = e
	return e
}

// Update it
func (k *Keyed) Update(p *Point, data interface{}) *Point {
	k.Clear()
	k.Start = p
	e := k.start.UpdateXY(p)
	e = k.Item.Update(e.RightN(0), data)
	e = k.end.UpdateXY(e.RightN(0))
	k.End = e
	return e
}

// UpdateXY update location
func (k *Keyed) UpdateXY(p *Point) *Point {
	k.Clear()
	k.Start = p
	e := k.start.UpdateXY(p)
	e = k.Item.UpdateXY(e.RightN(0))
	e = k.end.UpdateXY(e.RightN(0))
	k.End = e
	return e
}

// UpdateData update location
func (k *Keyed) UpdateData(data interface{}) *Point {
	k.Item.Clear()
	k.end.Clear()

	e := k.Item.UpdateData(data)
	e = k.end.UpdateXY(e.RightN(0))
	k.End = e
	return e
}

// Label represent a label
type Label struct {
	*Keyed
}

// NewLabel create Label
func NewLabel(p *Point, key string, value string) *Label {
	s := NewText(ZeroPoint, value)
	return &Label{NewKeyed(p, key, s)}
}
