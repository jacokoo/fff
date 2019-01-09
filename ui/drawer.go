package ui

import (
	"github.com/nsf/termbox-go"
)

// Drawer is the base interface for ui components
type Drawer interface {
	setStart(*Point)
	Draw() *Point
	Clear()
}

type mover interface {
	moveTo(*Point) *Point
}

// Move a drawer
func Move(d Drawer, p *Point) *Point {
	if m, ok := d.(mover); ok {
		return m.moveTo(p)
	}
	d.setStart(p)
	return d.Draw()
}

// Redraw clear it and draw again
func Redraw(d Drawer) *Point {
	d.Clear()
	return d.Draw()
}

func measure(d Drawer) (int, int) {
	w, _ := termbox.Size()
	p := &Point{w + 1, 0}
	pp := Move(d, p)
	d.Clear()
	return pp.X - p.X + 1, pp.Y - p.Y + 1
}

// Drawable contains base properties for draw
type Drawable struct {
	*Rect
	*Color
}

// SetStart set the start point
func (d *Drawable) setStart(p *Point) {
	d.Start = p
}

// Draw it
func (d *Drawable) Draw() *Point {
	return d.End
}

// NewDrawable create drawable
func NewDrawable(p *Point) *Drawable {
	return &Drawable{p.ToRect(), getColor("normal")}
}
