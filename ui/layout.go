package ui

// FlowLayout layout from left to right
type FlowLayout struct {
	Drawers      []Drawer
	LeftPadding  int
	RightPadding int
	padding      func(*Point) *Point
	*Drawable
}

// NewFlowLayout create flow layout
func NewFlowLayout(p *Point, padding func(*Point) *Point, items ...Drawer) *FlowLayout {
	return &FlowLayout{items, 0, 0, padding, NewDrawable(p)}
}

// Draw it
func (fl *FlowLayout) Draw() *Point {
	p := fl.Start.RightN(fl.LeftPadding)
	for i, v := range fl.Drawers {
		p = Move(v, p)
		if i != len(fl.Drawers)-1 {
			p = fl.padding(p)
		}
	}
	fl.End = p.RightN(fl.RightPadding)
	return p
}

// Append items to layout
func (fl *FlowLayout) Append(ds ...Drawer) {
	ns := fl.Drawers
	for _, d := range ds {
		exists := false
		for _, e := range ns {
			if e == d {
				exists = true
				break
			}
		}
		if !exists {
			ns = append(ns, d)
		}
	}
	fl.Drawers = ns
}

// Remove item from layout
func (fl *FlowLayout) Remove(d Drawer) {
	ds := make([]Drawer, 0)
	for _, v := range fl.Drawers {
		if v != d {
			ds = append(ds, v)
		}
	}
	fl.Drawers = ds
}

// DoLayout re-layout
func (fl *FlowLayout) DoLayout() {
	fl.Clear()
	fl.Draw()
}

// RightAlignFlowLayout right align
type RightAlignFlowLayout struct {
	layout *FlowLayout
	*Drawable
}

// NewRightAlignFlowLayout create right align flow layout
func NewRightAlignFlowLayout(p *Point, padding func(*Point) *Point, items ...Drawer) *RightAlignFlowLayout {
	return &RightAlignFlowLayout{NewFlowLayout(p, padding, items...), NewDrawable(p)}
}

// Draw it
func (ra *RightAlignFlowLayout) Draw() *Point {
	w, _ := measure(ra.layout)
	s := ra.Start.LeftN(w)
	return Move(ra.layout, s)
}

// Clear it
func (ra *RightAlignFlowLayout) Clear() {
	ra.layout.Clear()
}

// Append items to layout
func (ra *RightAlignFlowLayout) Append(d ...Drawer) {
	ra.layout.Append(d...)
}

// Remove item from layout
func (ra *RightAlignFlowLayout) Remove(d Drawer) {
	ra.layout.Remove(d)
}

// DoLayout re-layout
func (ra *RightAlignFlowLayout) DoLayout() {
	ra.Clear()
	ra.Draw()
}
