package ui

var (
	singleCorner = "┬"
	doubleCorner = "╥"
	cornerReset  = "─"
)

// ColumnItem a item
type ColumnItem struct {
	item     Drawer
	showLine bool
	line     *VLine
	corner   *Text
	*Drawable
}

func newColumnItem(height int, singleLine bool, item Drawer) *ColumnItem {
	var line *VLine
	var corner *Text
	if singleLine {
		line = NewVLine(ZeroPoint, height)
		corner = NewText(ZeroPoint, singleCorner)
	} else {
		line = NewVDLine(ZeroPoint, height)
		corner = NewText(ZeroPoint, doubleCorner)
	}

	return &ColumnItem{item, true, line, corner, NewDrawable(ZeroPoint)}
}

// Draw it
func (ci *ColumnItem) Draw() *Point {
	p := ci.item.MoveTo(ci.Start)
	if !ci.showLine {
		ci.End = p
		return p
	}

	p = p.Right()
	p.Y = ci.Start.Y
	ci.corner.MoveTo(p.Up())
	p = ci.line.MoveTo(p)
	ci.End = p
	return p
}

// MoveTo update location
func (ci *ColumnItem) MoveTo(p *Point) *Point {
	ci.Start = p
	return ci.Draw()
}

// Clear it
func (ci *ColumnItem) Clear() {
	ci.Rect.Clear()
	ss := ci.corner.Data
	ci.corner.Data = cornerReset
	ci.corner.Draw()
	ci.corner.Data = ss
}

// Column container
type Column struct {
	Width, Height int
	items         []*ColumnItem
	itemMap       map[Drawer]*ColumnItem
	line          *HLine
	*Drawable
}

// Draw it
func (c *Column) Draw() *Point {
	p := c.line.MoveTo(c.Start)
	c.End.X = p.X
	c.End.Y = c.Start.Y + c.Height - 1

	p = c.Start.Down()
	for _, v := range c.items {
		pp := v.MoveTo(p).Right()
		pp.Y = p.Y
		p = pp
	}
	return c.End
}

// MoveTo update loation
func (c *Column) MoveTo(p *Point) *Point {
	c.Start = p
	return c.Draw()
}

func (c *Column) add(item Drawer, singleLine bool) {
	var p *Point
	if len(c.items) == 0 {
		p = c.Start.Down()
	} else {
		p = c.items[len(c.items)-1].End.Right()
		p.Y = c.Start.Y + 1
	}
	col := newColumnItem(c.Height-1, singleLine, item)
	c.items = append(c.items, col)
	c.itemMap[item] = col

	col.MoveTo(p)
}

// Add column
func (c *Column) Add(item Drawer) {
	c.add(item, true)
}

// Add2 column with double line
func (c *Column) Add2(item Drawer) {
	c.add(item, false)
}

// Get column item
func (c *Column) Get(item Drawer) *ColumnItem {
	return c.itemMap[item]
}

// Remove the last column
func (c *Column) Remove() {
	l := len(c.items) - 1
	c.items[l].Clear()
	c.items = c.items[:l]
}

// RemoveAll the columns
func (c *Column) RemoveAll() {
	c.Clear()
	c.line.Draw()
	c.items = c.items[:0]
}
