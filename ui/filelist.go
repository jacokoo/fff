package ui

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/jacokoo/fff/model"
	runewidth "github.com/mattn/go-runewidth"
)

// FileList is a list of file
type FileList struct {
	list      *List
	filter    *Text
	countInfo *Text
	*Drawable
}

func newFileList(p *Point, height int) *FileList {
	h := height - 1
	filter := NewText(p.DownN(h), "")
	filter.Color = colorFilter()
	list := NewList(p, 0, h, nil, nil)
	return &FileList{list, filter, NewText(p, ""), NewDrawable(p)}
}

func (fl *FileList) setData(names []string, hints []int, current int, filter string) {
	fl.list.SetData(names, hints, current)
	fl.filter.Data = filter
	fl.countInfo.Data = fmt.Sprintf("[%d/%d]", current+1, len(fl.list.Data))
}

func (fl *FileList) setFilter(filter string) {
	if len(filter) != 0 {
		filter = "F: " + filter
	}
	fl.filter.Data = filter
}

func (fl *FileList) setCurrent(current int) {
	fl.list.Select(current)
	fl.countInfo.Data = fmt.Sprintf("[%d/%d]", current+1, len(fl.list.Data))
}

// Draw it
func (fl *FileList) Draw() *Point {
	p := fl.list.Draw().Down()
	fl.End = p

	p = p.RightN(0)
	p.X = fl.Start.X
	fl.filter.MoveTo(p)
	fl.countInfo.MoveTo(p.LeftN(2 + len(fl.countInfo.Data)))
	return fl.End
}

// MoveTo update location
func (fl *FileList) MoveTo(p *Point) *Point {
	fl.Start = p
	return fl.Draw()
}

const (
	columnWidth int = 30
)

func truncName(str string, count int) (string, int) {
	s, c := "", 0
	for _, v := range str {
		w := runewidth.RuneWidth(v)
		if c+w > count {
			return s + "..", c + 2
		}
		s += string(v)
		c += w
	}
	return s, c
}

func formatSize(size int64) string {
	unit := "B"
	b := float32(size)

	if b > 1024 {
		unit = "K"
		b = b / 1024
	} else {
		return fmt.Sprintf("%dB", size)
	}

	if b > 1024 {
		unit = "M"
		b = b / 1024
	}

	if b > 1024 {
		unit = "G"
		b = b / 1024
	}
	return fmt.Sprintf("%.2f%s", b, unit)
}

func expandedName(size string, maxSize int, fi model.FileItem) string {
	ti := fi.ModTime().Format("2006-01-02 15:04:05")
	md := fi.Mode().String()
	si := strings.Repeat(" ", maxSize-len(size)) + size
	return fmt.Sprintf("%s  %s  %s  %s", ti, md, si, fi.Name())
}

func normalName(size string, v model.FileItem) string {
	na := v.Name()
	if v.IsDir() {
		fs, _ := ioutil.ReadDir(filepath.Join(v.Path(), na))
		size = fmt.Sprintf("%d it.", len(fs))
	}

	re := columnWidth - len(size) - 4
	na, c := truncName(na, re-3)
	re -= c
	if re < 0 {
		re = 0
	}

	return fmt.Sprintf("%s%s%s  ", na, strings.Repeat(" ", re), size)
}

func fileNames(co model.Column) ([]string, []int) {
	le := len(co.Files())
	names := make([]string, le)
	hints := make([]int, le)
	sis := make([]string, le)

	maxSize := 0
	for i, v := range co.Files() {
		sis[i] = formatSize(v.Size())
		if len(sis[i]) > maxSize {
			maxSize = len(sis[i])
		}
	}

	for i, v := range co.Files() {
		var n string
		if co.IsShowDetail() {
			n = expandedName(sis[i], maxSize, v)
		} else {
			n = normalName(sis[i], v)
		}

		mark := " "
		if co.IsMarked(i) {
			mark = "*"
		}

		names[i] = fmt.Sprintf(" %s%s", mark, n)
		hints[i] = 0
		if v.IsDir() {
			hints[i] = 1
		}
		if co.IsMarked(i) {
			hints[i] = 2
		}
	}
	return names, hints
}
