package ui

// Drawer is the base interface for ui components
type Drawer interface {
	Draw() *Point
	Clear()
	MoveTo(p *Point) *Point
}

// Drawable contains base properties for draw
type Drawable struct {
	*Rect
	*Color
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
		p = v.MoveTo(p)
		if i != len(d.Drawers)-1 {
			p = d.padding(p)
		}
	}
	d.End = p
	return p
}

// MoveTo update location
func (d DrawerList) MoveTo(p *Point) *Point {
	d.Start = p
	return d.Draw()
}
