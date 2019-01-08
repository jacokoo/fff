package ui

import (
	"github.com/nsf/termbox-go"
)

// Popup restore covered after clear
type Popup struct {
	Item          Drawer
	width, height int
	backed        [][]termbox.Cell
	*Drawable
}

// NewPopup create popup
func NewPopup(p *Point, item Drawer) *Popup {
	return &Popup{item, 0, 0, nil, NewDrawable(p)}
}

// Draw it
func (p *Popup) Draw() *Point {
	cells := termbox.CellBuffer()
	tw, _ := termbox.Size()
	w, h := measure(p.Item)
	p.width = w
	p.height = h

	bks := make([][]termbox.Cell, 0)
	for i := 0; i < h; i++ {
		base := (p.Start.Y+i)*tw + p.Start.X
		row := make([]termbox.Cell, w)
		for j := 0; j < w; j++ {
			row[j] = cells[base+j]
		}
		bks = append(bks, row)
	}
	p.backed = bks
	p.End = Move(p.Item, p.Start)
	return p.End
}

// Clear it
func (p *Popup) Clear() {
	for i := 0; i < p.height; i++ {
		for j := 0; j < p.width; j++ {
			c := p.backed[i][j]
			termbox.SetCell(p.Start.X+j, p.Start.Y+i, c.Ch, c.Fg, c.Bg)
		}
	}
}
