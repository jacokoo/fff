package file

import (
	"github.com/jacokoo/fff/executor"
)

type pathItem struct {
	loader Loader
	path   string
}

func (pi *pathItem) dir() *pathItem {
	np := pi.loader.PathDir(pi.path)
	if np == pi.path {
		return pi
	}
	return &pathItem{pi.loader, np}
}

type path struct {
	items []*pathItem
}

func (p *path) Load() executor.Future {
	fs := make([]func(interface{}) executor.Future, 0, len(p.items)*2)
	for _, v := range p.items {
		vv := v
		fs = append(fs, func(data interface{}) executor.Future {
			return vv.loader.Create(data.(File))
		})
		fs = append(fs, func(data interface{}) executor.Future {
			file := data.(DirOp)
			return file.To(vv.path)
		})
	}
	return executor.Combine(executor.OkFuture(File(nil)), fs...)
}

func (p *path) String() string {
	re := p.items[0].path
	for i := 1; i < len(p.items); i++ {
		re += p.items[i].loader.Name() + "://" + p.items[i].path
	}
	return re
}

func (p *path) Dir() Path {
	idx := len(p.items) - 1
	item := p.items[idx]
	p2 := item.dir()

	if p2 == item {
		if idx == 0 {
			return p
		}
		is := make([]*pathItem, len(p.items)-1)
		copy(is, p.items)
		return &path{is}
	}

	return &path{append(p.items[:idx], p2)}
}
