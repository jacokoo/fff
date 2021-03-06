package ui

var (
	singleCorner    = "┬"
	doubleCorner    = "╥"
	cornerReset     = string(chh)
	indicatorString = " ▼ "
	indicatorReset  = string([]rune{chh, chh, chh})
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
	p := Move(ci.item, ci.Start)
	if !ci.showLine {
		ci.End = p
		return p
	}

	p = p.Right()
	p.Y = ci.Start.Y
	Move(ci.corner, p.Up())
	p = Move(ci.line, p)
	ci.End = p
	return p
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
	line          *HLine
	indicator     *Text
	*Drawable
}

// NewColumn create column
func NewColumn(p *Point, width, height int) *Column {
	return &Column{width, height, nil, NewHLine(p, width), NewText(p, ""), NewDrawable(p)}
}

// Draw it
func (c *Column) Draw() *Point {
	p := Move(c.line, c.Start)
	c.End.X = p.X
	c.End.Y = c.Start.Y + c.Height - 1

	p = c.Start.Down()
	for _, v := range c.items {
		pp := Move(v, p).Right()
		pp.Y = p.Y
		p = pp
	}

	c.resetIndicator()
	return c.End
}

func (c *Column) resetIndicator() {
	c.indicator.Data = indicatorReset
	c.indicator.Color = colorNormal()
	c.indicator.Draw()

	c.indicator.Data = indicatorString
	c.indicator.Color = colorIndicator()
	last := c.Last()
	Move(c.indicator, &Point{last.Start.X + (last.End.X-last.Start.X)/2 - 1, c.Start.Y})
}

func (c *Column) add(item Drawer, singleLine bool) *ColumnItem {
	var p *Point
	if len(c.items) == 0 {
		p = c.Start.Down()
	} else {
		p = c.items[len(c.items)-1].End.Right()
		p.Y = c.Start.Y + 1
	}
	col := newColumnItem(c.Height-1, singleLine, item)
	c.items = append(c.items, col)
	return col
}

// Add column
func (c *Column) Add(item Drawer) *ColumnItem {
	return c.add(item, true)
}

// Add2 column with double line
func (c *Column) Add2(item Drawer) *ColumnItem {
	return c.add(item, false)
}

// Last the last column item
func (c *Column) Last() *ColumnItem {
	return c.items[len(c.items)-1]
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

// Shift a clumn
func (c *Column) Shift(keepFirst bool) {
	its := make([]*ColumnItem, 0)

	idx := 1
	if keepFirst {
		its = append(its, c.items[0])
		idx = 2
	}

	its = append(its, c.items[idx:]...)
	c.items = its
}
