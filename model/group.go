package model

import (
	"errors"
	"path/filepath"
	"strings"
)

// CloseResult of Group.CloseDir
type CloseResult uint8

// Close types
const (
	CloseSuccess CloseResult = iota
	CloseToParent
	CloseNothing
)

// Group is a column group
type Group interface {
	Path() string
	Columns() []Column
	Current() Column
	Shift() bool
	OpenDir() error
	OpenRoot(root string) error
	CloseDir() CloseResult
	JumpTo(colIdx, fileIdx int) bool
	Refresh()
	Record()
	Restore()
	Operator
}

type old struct {
	from    string
	to      string
	current string
}

// LocalGroup local group
type LocalGroup struct {
	path    string
	old     *old
	columns []Column
	Operator
}

// Path current path
func (g *LocalGroup) Path() string {
	return g.path
}

// Columns get columns
func (g *LocalGroup) Columns() []Column {
	return g.columns
}

// Current column
func (g *LocalGroup) Current() Column {
	return g.columns[len(g.columns)-1]
}

// Shift column
func (g *LocalGroup) Shift() bool {
	if len(g.columns) == 1 {
		return false
	}
	g.columns = g.columns[1:]
	return true
}

// OpenDir selected dir
func (g *LocalGroup) OpenDir() error {
	co := g.Current()
	fi, err := co.CurrentFile()
	if err != nil {
		return err
	}
	if !fi.IsDir() {
		return errors.New("not a dir")
	}

	co.ClearMark()
	if co.IsShowDetail() {
		co.ToggleDetail()
	}

	pa := fi.Path()
	items, err := g.ReadDir(pa)
	if err != nil {
		return err
	}

	cc := NewLocalColumn(pa, items)
	g.path = pa
	g.columns = append(g.columns, cc)
	return nil
}

// OpenRoot open path in first column
func (g *LocalGroup) OpenRoot(root string) error {
	items, err := g.ReadDir(root)
	if err != nil {
		return err
	}

	g.columns = g.columns[:1]
	g.columns[0].Refresh(root, items)
	g.path = root
	return nil
}

// CloseDir close current dir
func (g *LocalGroup) CloseDir() CloseResult {
	if len(g.columns) > 1 {
		g.columns = g.columns[:len(g.columns)-1]
		g.path = g.Current().Path()
		return CloseSuccess
	}

	pa := filepath.Dir(g.path)
	if pa == g.path {
		return CloseNothing
	}

	g.OpenRoot(pa)
	return CloseToParent
}

// JumpTo a file
func (g *LocalGroup) JumpTo(colIdx, fileIdx int) bool {
	if colIdx >= len(g.columns) || fileIdx >= len(g.columns[colIdx].Files()) {
		return false
	}

	g.columns = g.columns[:colIdx+1]
	g.path = g.Current().Path()
	g.Current().Select(fileIdx)
	return true
}

// Refresh current dirs
func (g *LocalGroup) Refresh() {
	co := g.Current()
	fs, _ := g.ReadDir(g.path)
	co.Refresh(g.path, fs)
}

// Record record current path for jump back
func (g *LocalGroup) Record() {
	from := g.columns[0].Path()
	co := g.Current()
	to := co.Path()
	current := co.Files()[co.Current()].Name()
	g.old = &old{from, to, current}
}

// Restore to previous status
func (g *LocalGroup) Restore() {
	if g.old == nil {
		return
	}
	old := g.old
	g.Record()
	current := g.old
	defer func() { g.old = current }()

	if err := g.OpenRoot(old.from); err != nil {
		return
	}

	sep := string(filepath.Separator)
	ts1, ts2 := strings.Split(old.from, sep), strings.Split(old.to, sep)
	for i := len(ts1); i < len(ts2); i++ {
		if suc := g.Current().SelectByName(ts2[i]); !suc {
			return
		}
		if err := g.OpenDir(); err != nil {
			return
		}
	}
	g.Current().SelectByName(old.current)
}

// NewLocalGroup create local group
func NewLocalGroup(path string) (Group, error) {
	op := &LocalOperator{}
	fs, err := op.ReadDir(path)
	if err != nil {
		return nil, err
	}
	co := NewLocalColumn(path, fs)
	return &LocalGroup{path, nil, []Column{co}, op}, nil
}
