package ui

var (
	singleCorner = "┬"
	doubleCorner = "╥"
	cornerReset  = "─"
)

type column struct {
	Width, Height int
	*Drawable

	corner *Text
	vline  *VLine
}

func newColumn(p *Point, width, height int, corner string) *column {
	pp := p.RightN(width)
	ppp := pp.RightN(0)
	ppp.Y--

	var line *VLine
	if corner == singleCorner {
		line = NewVLine(pp, height)
	} else {
		line = NewVDLine(pp, height)
	}
	return &column{width, height, NewDrawable(p), NewText(ppp, corner), line}
}

// Draw It
func (c *column) Draw() *Point {
	c.corner.Draw()
	p := c.vline.Draw()
	p.Y--
	c.End = p
	return p
}

// MoveTo update location
func (c *column) MoveTo(p *Point) *Point {
	c.Start = p
	return c.Draw()
}

// Clear it
func (c *column) Clear() {
	c.Rect.Clear()
	c.corner.SetValue(cornerReset).Draw()
}

// InnerStart point for content draw
func (c *column) InnerStart() *Point {
	return &Point{c.Start.X + 1, c.Start.Y + 1}
}

// Columns represent dirs
type Columns struct {
	Width, Height int
	*Drawable

	line    *HLine
	columns []*column
}

// NewColumns create Columns
func NewColumns(p *Point, width, height int) *Columns {
	l := NewHLine(p, width)
	return &Columns{width, height, NewDrawable(p), l, make([]*column, 0)}
}

// Draw it
func (c *Columns) Draw() *Point {
	c.line.Draw()
	c.End.X = c.line.End.X
	c.End.Y = c.Start.Y + c.Height - 1
	for _, v := range c.columns {
		v.Draw()
	}
	return c.End
}

// MoveTo update location
func (c *Columns) MoveTo(p *Point) *Point {
	c.Start = p
	c.line.MoveTo(p)

	pp := p.Bottom()
	for _, v := range c.columns {
		v.MoveTo(pp)
		pp = pp.RightN(v.Width)
	}

	return c.End
}

func (c *Columns) add(width int, corner string) int {
	p := c.Start.Bottom()
	if len(c.columns) > 0 {
		p.X = c.columns[len(c.columns)-1].End.X + 1
	}

	co := newColumn(p, width, c.Height-1, corner)
	c.columns = append(c.columns, co)
	co.Draw()
	return len(c.columns) - 1
}

// Add a new column
func (c *Columns) Add(width int) int {
	return c.add(width, singleCorner)
}

// Add2 a new column with double line border
func (c *Columns) Add2(width int) int {
	return c.add(width, doubleCorner)
}

// StartAt returns the content start point
func (c *Columns) StartAt(index int) *Point {
	p := c.columns[index].Start.RightN(0)
	return p
}

// Remove the last column
func (c *Columns) Remove() {
	l := len(c.columns) - 1
	c.columns[l].Clear()
	c.columns = c.columns[:l]
}

// ClearAt clear column at idx
func (c *Columns) ClearAt(idx int) {
	c.columns[idx].Clear()
}

// RemoveAll the columns
func (c *Columns) RemoveAll() {
	c.Clear()
	c.line.Draw()
	c.columns = c.columns[:0]
}
