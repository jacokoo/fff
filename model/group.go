package model

import (
	"fmt"
	"path/filepath"
	"strings"
)

var (
	_ = Group(new(LocalGroup))
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
	CloseDir() (CloseResult, error)
	JumpTo(colIdx, fileIdx int) bool
	Refresh() error
	Record()
	Restore()
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
		nfi, err := LoadFile(fi)
		if err != nil {
			return err
		}
		fi = nfi
	}

	co.ClearMark()
	if co.IsShowDetail() {
		co.ToggleDetail()
	}

	cc, err := NewLocalColumn(fi)
	if err != nil {
		return err
	}

	g.path = fi.Path()
	g.columns = append(g.columns, cc)
	return nil
}

// OpenRoot open path in first column
func (g *LocalGroup) OpenRoot(root string) error {
	item, err := Load(root)
	if err != nil {
		return err
	}

	if !item.IsDir() {
		return fmt.Errorf("path: %s is not a dir", root)
	}

	for i := len(g.columns) - 1; i > 0; i-- {
		err := g.columns[i].File().(DirOp).Close()
		if err != nil {
			return err
		}
	}

	err = g.columns[0].Refresh(item)
	if err != nil {
		return err
	}

	g.columns = g.columns[:1]
	g.path = root
	return nil
}

// CloseDir close current dir
func (g *LocalGroup) CloseDir() (CloseResult, error) {
	file := g.Current().File()
	op := file.(DirOp)
	if len(g.columns) > 1 {
		err := op.Close()
		if err != nil {
			return CloseNothing, err
		}
		g.columns = g.columns[:len(g.columns)-1]
		g.path = g.Current().Path()
		return CloseSuccess, nil
	}

	parent, err := op.Dir()
	if err != nil {
		return CloseNothing, nil
	}

	if parent == file {
		return CloseNothing, nil
	}

	return CloseToParent, g.Current().Refresh(parent)
}

// JumpTo a file
func (g *LocalGroup) JumpTo(colIdx, fileIdx int) bool {
	if colIdx >= len(g.columns) || fileIdx >= len(g.columns[colIdx].Files()) {
		return false
	}

	for i := len(g.columns) - 1; i > colIdx; i-- {
		g.columns[i].File().(DirOp).Close()
	}

	g.columns = g.columns[:colIdx+1]
	g.path = g.Current().Path()
	g.Current().Select(fileIdx)
	return true
}

// Refresh current dirs
func (g *LocalGroup) Refresh() error {
	return g.Current().Refresh(nil)
}

// Record record current path for jump back
func (g *LocalGroup) Record() {
	from := g.columns[0].Path()
	co := g.Current()
	to := co.Path()
	current := ""
	if fi, err := co.CurrentFile(); err == nil {
		current = fi.Name()
	}
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
	fi, err := Load(path)
	if err != nil {
		return nil, err
	}
	if !fi.IsDir() {
		return nil, fmt.Errorf("path: %s is not dir", path)
	}

	co, err := NewLocalColumn(fi)
	if err != nil {
		return nil, err
	}
	return &LocalGroup{path, nil, []Column{co}}, nil
}
