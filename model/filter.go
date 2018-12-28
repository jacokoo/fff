package model

import (
	"strconv"
	"strings"
	"time"
)

const (
	kb = 1024
	mb = kb * kb
	gb = mb * kb
)

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

func matchByType(filter string, file FileItem) bool {
	if len(filter) == 0 {
		return true
	}

	if len(filter) > 1 {
		return false
	}

	if filter == "f" && !file.IsDir() {
		return true
	}

	if filter == "d" && file.IsDir() {
		return true
	}

	return false
}

func matchByTime(filter string, file FileItem) bool {
	if len(filter) == 0 {
		return true
	}

	var base time.Duration
	var dd time.Duration
	switch filter[len(filter)-1] {
	case 'h':
		base = time.Hour
	case 'd':
		base = time.Hour * 24
	case 'm':
		base = time.Hour * 24 * 30
	}

	if base > 0 {
		ii, err := strconv.Atoi(filter[:len(filter)-1])
		if err == nil {
			dd = base * time.Duration(ii)
		}
	} else {
		ii, err := strconv.Atoi(filter)
		if err == nil {
			dd = time.Hour * time.Duration(ii)
		}
	}

	return file.ModTime().After(time.Now().Add(dd * -1))
}

func matchBySize(filter string, gt bool, file FileItem) bool {
	if len(filter) == 0 {
		return true
	}

	var size float64
	base := 0
	switch filter[len(filter)-1] {
	case 'k':
		base = kb
	case 'm':
		base = mb
	case 'g':
		base = gb
	}
	if base > 0 {
		size, _ = strconv.ParseFloat(filter[:len(filter)-1], 32)
		size = size * float64(base)
	} else {
		size, _ = strconv.ParseFloat(filter, 32)
	}

	if gt {
		return file.Size() > int64(size)
	}

	return file.Size() < int64(size)
}

func (bf *BaseFilter) matchItem(filter string, file FileItem) bool {
	if len(filter) == 0 {
		return true
	}

	switch filter[0] {
	case '>':
		return matchBySize(filter[1:], true, file)
	case '<':
		return matchBySize(filter[1:], false, file)
	case '+':
		return matchByTime(filter[1:], file)
	case ':':
		return matchByType(filter[1:], file)
	default:
		if !strings.Contains(file.Name(), filter) {
			return false
		}
	}

	return true
}

func (bf *BaseFilter) match(file FileItem) bool {
	if len(bf.filter) == 0 {
		return true
	}

	fs := strings.Split(bf.filter, " ")
	for _, v := range fs {
		if !bf.matchItem(v, file) {
			return false
		}
	}
	return true
}

// DoFilter apply filter
func (bf *BaseFilter) DoFilter() {
	fs := make([]FileItem, 0)
	for _, v := range bf.origins {
		if !bf.match(v) {
			continue
		}

		if !bf.showHidden && strings.HasPrefix(v.Name(), ".") {
			continue
		}

		fs = append(fs, v)
	}
	bf.files = fs
}
