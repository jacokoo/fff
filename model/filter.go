package model

import "strings"

// Filterer content can be filtered
type Filterer interface {
	ToggleHidden()
	IsShowHidden() bool
	Filter() string
	SetFilter(filter string)
	DoFilter()
}

// BaseFilter base filter
type BaseFilter struct {
	filter     string
	showHidden bool
	*BaseFileList
}

// Filter get the filter
func (bf *BaseFilter) Filter() string {
	return bf.filter
}

// SetFilter set the filter
func (bf *BaseFilter) SetFilter(filter string) {
	bf.filter = filter
}

// IsShowHidden if to show hidden file
func (bf *BaseFilter) IsShowHidden() bool {
	return bf.showHidden
}

// ToggleHidden toggle show hidden file
func (bf *BaseFilter) ToggleHidden() {
	bf.showHidden = !bf.showHidden
}

// DoFilter apply filter
func (bf *BaseFilter) DoFilter() {
	fs := make([]FileItem, 0)
	for _, v := range bf.origins {
		if !strings.Contains(v.Name(), bf.filter) {
			continue
		}

		if !bf.showHidden && strings.HasPrefix(v.Name(), ".") {
			continue
		}

		fs = append(fs, v)
	}
	bf.files = fs
}
