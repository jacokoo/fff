package model

import (
	"path/filepath"
)

// Workspace hold all state
type Workspace struct {
	Groups         []Group
	Clip           CopySource
	Current        int
	Bookmark       *Bookmark
	showBookmark   bool
	showClipDetail bool
}

// NewWorkspace create workspace
func NewWorkspace(maxGroups int, wd, configDir string) *Workspace {
	gs := make([]Group, maxGroups)
	g, err := NewLocalGroup(wd)
	if err != nil {
		panic(err)
	}
	gs[0] = g

	return &Workspace{gs, nil, 0, NewBookmark(filepath.Join(configDir, "bookmarks")), true, false}
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

// IsShowClipDetail if to show clip detail
func (w *Workspace) IsShowClipDetail() bool {
	return w.showClipDetail
}

// ToggleClipDetail toggle clip detail
func (w *Workspace) ToggleClipDetail() {
	w.showClipDetail = !w.showClipDetail
}
