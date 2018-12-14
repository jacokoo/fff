package main

import (
	"io/ioutil"
	"os"
	"sort"
	"strings"
)

const (
	orderName = iota
	orderMTime
	orderSize
)

type column struct {
	path       string
	filter     string
	origin     []os.FileInfo
	files      []os.FileInfo
	markes     []int
	order      int
	showHidden bool
	current    int
	expanded   bool
}

type files []os.FileInfo

func (c files) Len() int      { return len(c) }
func (c files) Swap(i, j int) { c[i], c[j] = c[j], c[i] }
func (c files) compare(i, j int) int {
	if c[i].IsDir() && !c[j].IsDir() {
		return -1
	}

	if !c[i].IsDir() && c[j].IsDir() {
		return 1
	}

	return 0
}

type byName struct{ files }
type byMTime struct{ files }
type bySize struct{ files }

func (c byName) Less(i, j int) bool {
	switch c.files.compare(i, j) {
	case -1:
		return true
	case 1:
		return false
	default:
		return c.files[i].Name() < c.files[j].Name()
	}
}

func (c byMTime) Less(i, j int) bool {
	switch c.files.compare(i, j) {
	case -1:
		return true
	case 1:
		return false
	default:
		return c.files[i].ModTime().After(c.files[j].ModTime())
	}
}

func (c bySize) Less(i, j int) bool {
	switch c.files.compare(i, j) {
	case -1:
		return true
	case 1:
		return false
	default:
		a, b := c.files[i], c.files[j]
		if a.Size() == b.Size() {
			return a.Name() < b.Name()
		}
		return a.Size() < b.Size()
	}
}

func (co *column) sort(order int) {
	switch order {
	case orderName:
		sort.Sort(byName{co.files})
	case orderMTime:
		sort.Sort(byMTime{co.files})
	case orderSize:
		sort.Sort(bySize{co.files})
	}
	co.order = order
}

func newColumn(path string) *column {
	fs, _ := ioutil.ReadDir(path)
	co := &column{path, "", fs, fs, nil, orderName, false, 0, false}
	co.update()
	return co
}

// show/hide hidden files, do filter, clear markes
func (co *column) update() {
	fs := make([]os.FileInfo, 0)
	for _, v := range co.origin {
		if !strings.Contains(v.Name(), co.filter) {
			continue
		}

		if !co.showHidden && strings.HasPrefix(v.Name(), ".") {
			continue
		}

		fs = append(fs, v)
	}
	co.files = fs
	co.current = 0

	co.sort(co.order)
	co.unmarkAll()
}

func (co *column) marked(idx int) bool {
	for _, i := range co.markes {
		if i == idx {
			return true
		}
	}
	return false
}

func (co *column) toggleMark() {
	ii := -1
	for idx, i := range co.markes {
		if i == co.current {
			ii = idx
			break
		}
	}
	if ii == -1 {
		co.markes = append(co.markes, co.current)
		return
	}

	co.markes = append(co.markes[:ii], co.markes[ii+1:]...)
}

func (co *column) unmarkAll() {
	co.markes = nil
}

func (co *column) move(n int) {
	if len(co.files) == 0 {
		return
	}
	i := co.current + n
	if i < 0 {
		i = len(co.files) - 1
	}

	if i >= len(co.files) {
		i = 0
	}

	co.current = i
}

// Name return the label
func (co *column) Name() string {
	return "FILTER"
}

// Get return the current value
func (co *column) Get() string {
	return co.filter
}

// Append the input value
func (co *column) Append(ch rune) {
	co.filter += string(ch)
	co.update()
	gui <- uiColumnContentChange
}

// Delete the last char
func (co *column) Delete() bool {
	f := co.filter
	if len(f) == 0 {
		return false
	}

	co.filter = co.filter[:len(co.filter)-1]
	co.update()
	gui <- uiColumnContentChange
	return true
}

// End end input
func (co *column) End() {
}
