package ui

import (
	"fmt"
	"path/filepath"
	"strings"
)

// JumpItem present
type JumpItem struct {
	Key    []rune
	Action func() bool
	Point  *Point
}

func onlyTwo(rs []rune) []rune {
	if len(rs) < 3 {
		return rs
	}
	return rs[:2]
}

// JumpItems jump items of tab
func (t *Tab) JumpItems(fn func(int) func() bool) []*JumpItem {
	re := make([]*JumpItem, 0)
	ds := t.names.Drawers
	for i, v := range ds {
		t := v.(*Text)
		ac := fn(i)
		if ac == nil {
			continue
		}
		re = append(re, &JumpItem{
			[]rune(strings.Trim(t.Data, " ")),
			ac,
			t.Start.Down().MoveRight(),
		})
	}
	return re
}

// JumpItems jump items of path
func (p *Path) JumpItems(fn func(string) func() bool) []*JumpItem {
	re := make([]*JumpItem, 0)

	pa := ""
	for _, v := range p.items {
		pa = fmt.Sprintf("%s%c%s", pa, filepath.Separator, v.Data)
		if strings.HasPrefix(pa, "//") {
			pa = pa[1:]
		}
		ac := fn(pa)
		if ac == nil {
			continue
		}
		re = append(re, &JumpItem{onlyTwo([]rune(v.Data)), ac, v.Start.Down()})
	}
	return re
}

// JumpItems jump items of file list
func (fl *List) JumpItems(fn func(int) func() bool) []*JumpItem {
	re := make([]*JumpItem, 0)
	for i := fl.from; i < fl.to; i++ {
		it := fl.items[i]
		ac := fn(i)
		if ac == nil {
			continue
		}
		re = append(re, &JumpItem{onlyTwo([]rune(strings.Trim(it.Data, " "))), ac, it.Start.LeftN(0)})
	}
	return re
}
