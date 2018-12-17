package model

import "errors"

// Selector content can be selected
type Selector interface {
	Current() int
	Select(idx int) bool
	SelectByName(name string) bool
	SelectFirst() bool
	SelectLast() bool
	Move(delta int) bool
	CurrentFile() (FileItem, error)
}

// BaseSelector base selector
type BaseSelector struct {
	current int
	*BaseFileList
}

// Current select index
func (co *BaseSelector) Current() int {
	return co.current
}

// Select by index
func (co *BaseSelector) Select(idx int) bool {
	if len(co.files) == 0 || idx == co.current {
		return false
	}

	if idx < 0 {
		idx = len(co.files) - 1
	}

	if idx >= len(co.files) {
		idx = 0
	}

	co.current = idx
	return true
}

// Move current select index by delta(can be negtive)
func (co *BaseSelector) Move(delta int) bool {
	return co.Select(co.current + delta)
}

// SelectByName select by name
func (co *BaseSelector) SelectByName(name string) bool {
	for i, v := range co.files {
		if v.Name() == name {
			co.current = i
			return true
		}
	}
	return false
}

// SelectFirst move select to first
func (co *BaseSelector) SelectFirst() bool {
	return co.Select(0)
}

// SelectLast move select to last
func (co *BaseSelector) SelectLast() bool {
	return co.Select(len(co.files) - 1)
}

// CurrentFile current selected file/dir
func (co *BaseSelector) CurrentFile() (FileItem, error) {
	if len(co.files) == 0 {
		return nil, errors.New("no file")
	}

	return co.files[co.current], nil
}
