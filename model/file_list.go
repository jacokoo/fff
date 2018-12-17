package model

import "sort"

// Order type
type Order uint8

// Order types
const (
	OrderByName Order = iota
	OrderByMTime
	OrderBySize
)

// FileList sortable file list
type FileList interface {
	Files() []FileItem
	Sort(order Order)
	Order() Order
	ToggleDetail()
	IsShowDetail() bool
}

// BaseFileList a file list with back list
type BaseFileList struct {
	order      Order
	files      []FileItem
	origins    []FileItem
	showDetail bool
}

type items []FileItem

func (c items) Len() int      { return len(c) }
func (c items) Swap(i, j int) { c[i], c[j] = c[j], c[i] }
func (c items) compare(i, j int) int {
	if c[i].IsDir() && !c[j].IsDir() {
		return -1
	}

	if !c[i].IsDir() && c[j].IsDir() {
		return 1
	}

	return 0
}

// asc
type byName struct{ items }

// desc
type byMTime struct{ items }

// desc
type bySize struct{ items }

func (c byName) Less(i, j int) bool {
	switch c.items.compare(i, j) {
	case -1:
		return true
	case 1:
		return false
	default:
		return c.items[i].Name() < c.items[j].Name()
	}
}

func (c byMTime) Less(i, j int) bool {
	switch c.items.compare(i, j) {
	case -1:
		return true
	case 1:
		return false
	default:
		return c.items[i].ModTime().After(c.items[j].ModTime())
	}
}

func (c bySize) Less(i, j int) bool {
	switch c.items.compare(i, j) {
	case -1:
		return true
	case 1:
		return false
	default:
		a, b := c.items[i], c.items[j]
		if a.Size() == b.Size() {
			return a.Name() < b.Name()
		}
		return a.Size() > b.Size()
	}
}

// Sort files
func (fl *BaseFileList) Sort(order Order) {
	switch order {
	case OrderByName:
		sort.Sort(byName{fl.files})
	case OrderByMTime:
		sort.Sort(byMTime{fl.files})
	case OrderBySize:
		sort.Sort(bySize{fl.files})
	}
	fl.order = order
}

// Order the current order
func (fl *BaseFileList) Order() Order {
	return fl.order
}

// Files the files of file list
func (fl *BaseFileList) Files() []FileItem {
	return fl.files
}

// IsShowDetail is detail shown
func (fl *BaseFileList) IsShowDetail() bool {
	return fl.showDetail
}

// ToggleDetail toggle show detail
func (fl *BaseFileList) ToggleDetail() {
	fl.showDetail = !fl.showDetail
}
