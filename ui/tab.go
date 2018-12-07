package ui

// Tab is the ui tab sheet
type Tab struct {
	Current int
	keyed   *Keyed
	names   *DrawerList
}

// NewTab create tab
func NewTab(p *Point, name string, names []string) *Tab {
	ns := make([]Drawer, len(names))
	for i, v := range names {
		t := NewText(ZeroPoint, v)
		if i == 0 {
			t.Color = ColorSelected
		}
		ns[i] = t
	}

	pa := func(pp *Point) *Point {
		return pp.RightN(0)
	}

	dl := &DrawerList{NewDrawable(ZeroPoint), ns, pa}
	k := NewKeyed(p, name, dl)
	return &Tab{0, k, dl}
}

// Draw it
func (t *Tab) Draw() *Point {
	return t.keyed.Draw()
}

// Clear it
func (t *Tab) Clear() {
	t.keyed.Clear()
}

// Update it
func (t *Tab) Update(p *Point, data interface{}) *Point {
	pp := t.keyed.UpdateXY(p)
	return pp
}

// UpdateXY update location
func (t *Tab) UpdateXY(p *Point) *Point {
	return t.keyed.UpdateXY(p)
}

// UpdateData update select index
func (t *Tab) UpdateData(data interface{}) *Point {
	i := data.(int)
	if t.Current == i {
		return t.keyed.End
	}

	t1 := t.names.Drawers[t.Current].(*Text)
	t2 := t.names.Drawers[i].(*Text)

	t1.Clear()
	t2.Clear()

	t1.Color = ColorNormal
	t2.Color = ColorSelected
	t1.Draw()
	t2.Draw()

	return t.keyed.End
}
