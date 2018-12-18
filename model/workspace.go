package model

// Workspace hold all state
type Workspace struct {
	Groups       []Group
	Current      int
	showBookmark bool
	bookmark     *Bookmark
}

// CurrentGroup get the current group in use
func (w *Workspace) CurrentGroup() Group {
	return w.Groups[w.Current]
}

// SwitchTo change group
func (w *Workspace) SwitchTo(idx int) Group {
	if idx < 0 {
		idx = 0
	}
	if idx >= len(w.Groups) {
		idx = len(w.Groups) - 1
	}

	w.Current = idx
	return w.CurrentGroup()
}

// IsShowBookmark if to show bookmark
func (w *Workspace) IsShowBookmark() bool {
	return w.showBookmark
}

// ToggleBookmark toggle bookmark
func (w *Workspace) ToggleBookmark() {
	w.showBookmark = !w.showBookmark
}
