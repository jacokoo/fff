package ui

import (
	termbox "github.com/nsf/termbox-go"
)

// Drawer is the base interface for ui components
type Drawer interface {
	Draw() *Point
	Clear()
	MoveTo(p *Point) *Point
}

// Point represent a point in screen
type Point struct {
	X, Y int
}

// Equals is used to compare two point
func (p *Point) Equals(o *Point) bool {
	return p.X == o.X && p.Y == o.Y
}

// Right returns the right point of current point
func (p *Point) Right() *Point {
	return &Point{p.X + 1, p.Y}
}

// RightN returns the nth right point of current point
func (p *Point) RightN(n int) *Point {
	return &Point{p.X + n, p.Y}
}

// Bottom returns the bottom point of current point
func (p *Point) Bottom() *Point {
	return &Point{p.X, p.Y + 1}
}

// BottomN returns the nth bottom point of current point
func (p *Point) BottomN(n int) *Point {
	return &Point{p.X, p.Y + n}
}

// MoveRight moves the current point to the right by 1
// and reterns it self
func (p *Point) MoveRight() *Point {
	p.X++
	return p
}

// MoveRightN moves the current point to the right by N
// and reterns it self
func (p *Point) MoveRightN(n int) *Point {
	p.X += n
	return p
}

// MoveBottom moves the current point to the bottom by 1
// and reterns it self
func (p *Point) MoveBottom() *Point {
	p.Y++
	return p
}

// MoveBottomN moves the current point to the bottom by n
// and reterns it self
func (p *Point) MoveBottomN(n int) *Point {
	p.Y += n
	return p
}

// Rect represent a rectangle position
type Rect struct {
	Start, End *Point
}

// To creates a rect with another point
func (p *Point) To(p2 *Point) *Rect {
	return &Rect{p, p2}
}

// ToRect creates a empty rect
func (p Point) ToRect() *Rect {
	return &Rect{&p, &Point{p.X, p.Y}}
}

// Clear the content of the Rect
func (r Rect) Clear() {
	for i := r.Start.X; i <= r.End.X; i++ {
		for j := r.Start.Y; j <= r.End.Y; j++ {
			termbox.SetCell(i, j, ' ', termbox.ColorDefault, termbox.ColorDefault)
		}
	}
}

// Color represent the forecolor and background color of an point
type Color struct {
	FG, BG termbox.Attribute
}

var (
	// ZeroPoint x = 0, y = 0
	ZeroPoint = &Point{0, 0}

	// ColorNormal ts the default color
	ColorNormal = &Color{termbox.ColorDefault, termbox.ColorDefault}

	// ColorKeyword is the keyword color
	ColorKeyword = &Color{termbox.ColorCyan, termbox.ColorDefault}

	// ColorSelected is the color for selected item
	ColorSelected = &Color{termbox.ColorWhite, termbox.ColorCyan}
)

// Drawable contains base properties for draw
type Drawable struct {
	*Rect
	*Color
}

// NewDrawable create drawable
func NewDrawable(p *Point) *Drawable {
	return &Drawable{p.ToRect(), ColorNormal}
}

// DrawerList a batch of drawer
type DrawerList struct {
	*Drawable
	Drawers []Drawer
	padding func(*Point) *Point
}

// Draw it
func (d *DrawerList) Draw() *Point {
	p := d.Start
	for _, v := range d.Drawers {
		p = d.padding(v.MoveTo(p))
	}
	d.End = p
	return p
}

// MoveTo update location
func (d DrawerList) MoveTo(p *Point) *Point {
	d.Start = p
	return d.Draw()
}
