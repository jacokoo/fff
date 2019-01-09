package ui

type drawerContainer struct {
	Drawers []Drawer
}

// Append items to layout
func (fl *drawerContainer) Append(ds ...Drawer) {
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
func (fl *drawerContainer) Remove(d Drawer) {
	ds := make([]Drawer, 0)
	for _, v := range fl.Drawers {
		if v != d {
			ds = append(ds, v)
		}
	}
	fl.Drawers = ds
}

// FlowLayout layout from left to right
type FlowLayout struct {
	LeftPadding  int
	RightPadding int
	padding      func(*Point) *Point
	*Drawable
	*drawerContainer
}

// NewFlowLayout create flow layout
func NewFlowLayout(p *Point, padding func(*Point) *Point, items ...Drawer) *FlowLayout {
	return &FlowLayout{0, 0, padding, NewDrawable(p), &drawerContainer{items}}
}

// Draw it
func (fl *FlowLayout) Draw() *Point {
	p := fl.Start.RightN(fl.LeftPadding)
	for i, v := range fl.Drawers {
		p = Move(v, p).Right()
		if i != len(fl.Drawers)-1 && fl.padding != nil {
			p = fl.padding(p)
		}
	}
	if len(fl.Drawers) > 0 {
		p.MoveLeft()
	}
	fl.End = p.RightN(fl.RightPadding)
	return p
}

// RightAlignFlowLayout right align
type RightAlignFlowLayout struct {
	*FlowLayout
	Start *Point
}

// NewRightAlignFlowLayout create right align flow layout
func NewRightAlignFlowLayout(p *Point, padding func(*Point) *Point, items ...Drawer) *RightAlignFlowLayout {
	return &RightAlignFlowLayout{NewFlowLayout(p, padding, items...), p}
}

// Draw it
func (ra *RightAlignFlowLayout) Draw() *Point {
	w, _ := measure(ra.FlowLayout)
	s := ra.Start.LeftN(w)
	return Move(ra.FlowLayout, s)
}

// VerticalLayout vertical flow layout
type VerticalLayout struct {
	TopPadding, BottomPadding int
	padding                   func(*Point) *Point
	*Drawable
	*drawerContainer
}

// NewVerticalLayout create flow layout
func NewVerticalLayout(p *Point, padding func(*Point) *Point, items ...Drawer) *VerticalLayout {
	return &VerticalLayout{0, 0, padding, NewDrawable(p), &drawerContainer{items}}
}

// Draw it
func (vl VerticalLayout) Draw() *Point {
	p := vl.Start.DownN(vl.TopPadding)
	for i, v := range vl.Drawers {
		pp := Move(v, p).Down()
		pp.X = p.X
		if i != len(vl.Drawers)-1 && vl.padding != nil {
			pp = vl.padding(pp)
		}
		p = pp
	}
	if len(vl.Drawers) > 0 {
		p.MoveUp()
	}
	vl.End = p.DownN(vl.BottomPadding)
	return vl.End
}
