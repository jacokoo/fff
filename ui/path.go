package ui

import (
	"fmt"
	"path/filepath"
	"strings"
)

// Path represent seperated path
type Path struct {
	*Keyed
	items     []*Text
	PathItems []string
}

func pathItems(path string) []string {
	ts := strings.Split(path, string(filepath.Separator))
	if ts[0] == "" {
		ts[0] = "/"
	}
	if ts[len(ts)-1] == "" {
		ts = ts[:len(ts)-1]
	}
	return ts
}

func createPathItems(items []string) (*FlowLayout, []*Text) {
	ds := NewFlowLayout(ZeroPoint, nil)
	its := make([]*Text, 0, len(items))

	t := NewText(ZeroPoint, items[0])
	t.Color = colorFolder()
	ds.Append(t)
	its = append(its, t)

	for i := 1; i < len(items); i++ {
		if i > 1 || (i == 1 && items[0] != "/") {
			ds.Append(NewText(ZeroPoint, fmt.Sprintf("%c", filepath.Separator)))
		}
		t = NewText(ZeroPoint, items[i])
		t.Color = colorFolder()
		ds.Append(t)
		its = append(its, t)
	}

	return ds, its
}

// NewPath create path
func NewPath(p *Point, name string, path string) *Path {
	ps := pathItems(path)
	dl, its := createPathItems(ps)
	kd := NewKeyed(p, name, dl)
	return &Path{kd, its, ps}
}

// SetValue update value
func (p *Path) SetValue(path string) {
	ps := pathItems(path)
	dl, its := createPathItems(ps)
	p.Keyed.item = dl
	p.items = its
	p.PathItems = ps
}

// ItemRects rect for jump
func (p *Path) ItemRects() []*Rect {
	irs := make([]*Rect, len(p.items))

	for i, v := range p.items {
		irs[i] = v.Rect
	}
	return irs
}
