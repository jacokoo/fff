package main

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/jacokoo/fff/model"

	"github.com/jacokoo/fff/ui"
	runewidth "github.com/mattn/go-runewidth"
)

// FileList is a list of file
type FileList struct {
	col      model.Column
	height   int
	list     *ui.List
	filter   *ui.Text
	indicate *ui.Text
	*ui.Drawable
}

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

func expandedName(size string, maxSize int, path string, fi model.FileItem) string {
	ti := fi.ModTime().Format("2006-01-02 15:04:05")
	md := fi.Mode().String()
	si := strings.Repeat(" ", maxSize-len(size)) + size
	return fmt.Sprintf("%s  %s  %s  %s", ti, md, si, fi.Name())
}

func normalName(size string, path string, v model.FileItem) string {
	na := v.Name()
	if v.IsDir() {
		fs, _ := ioutil.ReadDir(filepath.Join(path, na))
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

func fileNames(col model.Column) ([]string, []int) {
	le := len(col.Files())
	names := make([]string, le)
	hints := make([]int, le)
	sis := make([]string, le)

	maxSize := 0
	for i, v := range col.Files() {
		sis[i] = formatSize(v.Size())
		if len(sis[i]) > maxSize {
			maxSize = len(sis[i])
		}
	}

	for i, v := range col.Files() {
		var n string
		if col.IsShowDetail() {
			n = expandedName(sis[i], maxSize, col.Path(), v)
		} else {
			n = normalName(sis[i], col.Path(), v)
		}

		mark := " "
		if col.IsMarked(i) {
			mark = "*"
		}

		names[i] = fmt.Sprintf(" %s%s", mark, n)
		hints[i] = 0
		if v.IsDir() {
			hints[i] = 1
		}
		if col.IsMarked(i) {
			hints[i] = 2
		}
	}
	return names, hints
}

// NewFileList create file list
func NewFileList(p *ui.Point, col model.Column, height int) *FileList {
	h := height - 1
	filter := ui.NewText(p.BottomN(h), "")
	filter.Color = colorFilter()
	ns, hs := fileNames(col)
	list := ui.NewList(p, col.Current(), h, ns, hs)
	return &FileList{col, h, list, filter, ui.NewText(p, ""), ui.NewDrawable(p)}
}

// Draw it
func (fl *FileList) Draw() {
	fl.list.Draw()
	fl.End.X = fl.Start.X + columnWidth
	fl.End.Y = fl.Start.Y + fl.height

	fl.updateFilter()
	fl.updateIndicate()
}

// Clear it
func (fl *FileList) Clear() {
	fl.list.Clear()
	fl.filter.Clear()
	fl.indicate.Clear()
}

// MoveTo update location
func (fl *FileList) MoveTo(p *ui.Point) *ui.Point {
	fl.Start = p
	fl.End.X = fl.Start.X + cwidth(fl.col)
	fl.End.Y = fl.Start.Y + fl.height
	fl.updateIndicate()
	fl.filter.MoveTo(p.BottomN(fl.height))
	fl.list.MoveTo(p)
	return fl.End
}

func (fl *FileList) updateFilter() {
	s := fl.col.Filter()
	if len(s) != 0 {
		s = "F: " + s
	}
	fl.filter.Clear()
	fl.filter.SetValue(s).Draw()
}

func (fl *FileList) updateIndicate() {
	data := fmt.Sprintf("[%d/%d]", fl.col.Current()+1, len(fl.col.Files()))
	fl.indicate.Clear()
	fl.indicate.SetValue(data).MoveTo(&ui.Point{X: (fl.Start.X + cwidth(fl.col)) - 2 - len(data), Y: fl.End.Y})
}

func (fl *FileList) update() {
	ns, hs := fileNames(fl.col)
	fl.list.SetData(ns, hs, fl.col.Current())
	fl.updateIndicate()
	fl.updateFilter()
	fl.End.X = fl.Start.X + cwidth(fl.col)
}

func (fl *FileList) updateSelect() {
	fl.list.Clear()
	fl.list.Select(fl.col.Current())
	fl.updateIndicate()
}
