package ui

import termbox "github.com/nsf/termbox-go"

var (
	// ZeroPoint x and y all are 0
	ZeroPoint = &Point{0, 0}
)

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

// Left returns the left point of current point
func (p *Point) Left() *Point {
	return &Point{p.X - 1, p.Y}
}

// LeftN returns the nth left point of current point
func (p *Point) LeftN(n int) *Point {
	return &Point{p.X - n, p.Y}
}

// Down returns the bottom point of current point
func (p *Point) Down() *Point {
	return &Point{p.X, p.Y + 1}
}

// DownN returns the nth bottom point of current point
func (p *Point) DownN(n int) *Point {
	return &Point{p.X, p.Y + n}
}

// Up returns the top point of current point
func (p *Point) Up() *Point {
	return &Point{p.X, p.Y - 1}
}

// UpN returns the nth top point of current point
func (p *Point) UpN(n int) *Point {
	return &Point{p.X, p.Y - n}
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

// MoveLeft moves the current point to the left by 1
// and reterns it self
func (p *Point) MoveLeft() *Point {
	p.X--
	return p
}

// MoveLeftN moves the current point to the left by N
// and reterns it self
func (p *Point) MoveLeftN(n int) *Point {
	p.X -= n
	return p
}

// MoveDown moves the current point to the bottom by 1
// and reterns it self
func (p *Point) MoveDown() *Point {
	p.Y++
	return p
}

// MoveDownN moves the current point to the bottom by n
// and reterns it self
func (p *Point) MoveDownN(n int) *Point {
	p.Y += n
	return p
}

// MoveUp moves the current point to the above by 1
// and reterns it self
func (p *Point) MoveUp() *Point {
	p.Y--
	return p
}

// MoveUpN moves the current point to the above by n
// and reterns it self
func (p *Point) MoveUpN(n int) *Point {
	p.Y += n
	return p
}

// To creates a rect with another point
func (p *Point) To(p2 *Point) *Rect {
	return &Rect{p, p2}
}

// ToRect creates a empty rect
func (p *Point) ToRect() *Rect {
	return &Rect{p, &Point{p.X, p.Y}}
}

// Rect represent a rectangle position
type Rect struct {
	Start, End *Point
}

// Clear the content of the Rect
func (r *Rect) Clear() {
	for i := r.Start.X; i <= r.End.X; i++ {
		for j := r.Start.Y; j <= r.End.Y; j++ {
			termbox.SetCell(i, j, ' ', termbox.ColorDefault, termbox.ColorDefault)
		}
	}
}
