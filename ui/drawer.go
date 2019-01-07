package ui

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

// DrawerList a batch of drawer
type DrawerList struct {
	*Drawable
	Drawers []Drawer
	padding func(*Point) *Point
}

// Draw it
func (d *DrawerList) Draw() *Point {
	p := d.Start
	for i, v := range d.Drawers {
		p = Move(v, p)
		if i != len(d.Drawers)-1 {
			p = d.padding(p)
		}
	}
	d.End = p
	return p
}
