package ui

// Tab is the ui tab sheet
type Tab struct {
	Current int
	names   *FlowLayout
	*Keyed
}

// NewTab create tab
func NewTab(p *Point, name string, names []string) *Tab {
	dl := NewFlowLayout(ZeroPoint, nil)
	for i, v := range names {
		t := NewText(ZeroPoint, v)
		if i == 0 {
			t.Color = colorTab()
		}
		dl.Append(t)
	}

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

	t1.Color = colorNormal()
	t2.Color = colorTab()
	Redraw(t1)
	Redraw(t2)

	return t.Keyed.End
}
