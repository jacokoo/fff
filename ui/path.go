package ui

import (
	"fmt"
	"path/filepath"
)

// Path represent seperated path
type Path struct {
	*Keyed
	items []*Text
}

func createPathItems(items []string) (*DrawerList, []*Text) {
	ds := make([]Drawer, 0, len(items)*2-1)
	its := make([]*Text, 0, len(items))

	t := NewText(ZeroPoint, items[0])
	t.Color = colorFolder()
	ds = append(ds, t)
	its = append(its, t)

	for i := 1; i < len(items); i++ {
		if i > 1 || (i == 1 && items[0] != "/") {
			ds = append(ds, NewText(ZeroPoint, fmt.Sprintf("%c", filepath.Separator)))
		}
		t = NewText(ZeroPoint, items[i])
		t.Color = colorFolder()
		ds = append(ds, t)
		its = append(its, t)
	}

	return &DrawerList{NewDrawable(ZeroPoint), ds, func(p *Point) *Point { return p.RightN(1) }}, its
}

// NewPath create path
func NewPath(p *Point, name string, items []string) *Path {
	dl, its := createPathItems(items)
	kd := NewKeyed(p, name, dl)
	return &Path{kd, its}
}

// SetValue update value
func (p *Path) SetValue(items []string) {
	p.Clear()
	dl, its := createPathItems(items)
	p.Keyed.item = dl
	p.items = its
	p.Draw()
}

// ItemRects rect for jump
func (p *Path) ItemRects() []*Rect {
	irs := make([]*Rect, len(p.items))

	for i, v := range p.items {
		irs[i] = v.Rect
	}
	return irs
}
