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
	left := NewText(ZeroPoint, "")
	left.Color = colorKeyword()

	right := NewText(ZeroPoint, "")
	right.Color = colorKeyword()
	return &Keyed{key, item, NewDrawable(p), left, right}
}

// Draw it
func (k *Keyed) Draw() *Point {
	k.left.Data = fmt.Sprintf("%s[", k.Key)
	e := Move(k.left, k.Start)

	e = Move(k.item, e.Right())
	k.right.Data = "]"
	k.End = Move(k.right, e.Right())
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

// FloatText restore the cells after clear
type FloatText struct {
	width  int
	backed []termbox.Cell
	*Text
}

// NewFloatText create float text
func NewFloatText(p *Point, data string) *FloatText {
	return &FloatText{0, nil, NewText(p, data)}
}

// Draw it
func (ft *FloatText) Draw() *Point {
	w := 0
	for _, v := range ft.Text.Data {
		w += runewidth.RuneWidth(v)
	}
	ft.width = w

	cs := make([]termbox.Cell, 0)
	width, _ := termbox.Size()
	cells := termbox.CellBuffer()
	base := width * ft.Start.Y
	for i := 0; i < w; i++ {
		cs = append(cs, cells[base+ft.Start.X+i])
	}
	ft.backed = cs

	return ft.Text.Draw()
}

// Clear it
func (ft *FloatText) Clear() {
	for i := 0; i < ft.width; i++ {
		termbox.SetCell(ft.Start.X+i, ft.Start.Y, ft.backed[i].Ch, ft.backed[i].Fg, ft.backed[i].Bg)
	}
	ft.backed = nil
}
