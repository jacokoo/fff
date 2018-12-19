package ui

// Tab is the ui tab sheet
type Tab struct {
	Current int
	names   *DrawerList
	*Keyed
}

// NewTab create tab
func NewTab(p *Point, name string, names []string) *Tab {
	ns := make([]Drawer, len(names))
	for i, v := range names {
		t := NewText(ZeroPoint, v)
		if i == 0 {
			t.Color = colorTab()
		}
		ns[i] = t
	}

	pa := func(pp *Point) *Point {
		return pp.Right()
	}

	dl := &DrawerList{NewDrawable(ZeroPoint), ns, pa}
	k := NewKeyed(p, name, dl)
	return &Tab{0, dl, k}
}

// SwitchTo update select index
func (t *Tab) SwitchTo(selected int) *Point {
	if t.Current == selected {
		return t.Keyed.End
	}

	t1 := t.names.Drawers[t.Current].(*Text)
	t2 := t.names.Drawers[selected].(*Text)
	t.Current = selected

	t1.Clear()
	t2.Clear()

	t1.Color = colorNormal()
	t2.Color = colorTab()
	t1.Draw()
	t2.Draw()

	return t.Keyed.End
}
