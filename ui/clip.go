package ui

import (
	"fmt"
)

// Clip render clip
type Clip struct {
	showed     bool
	showDetail bool
	items      []string
	list       *List
	box        *Box
	text       *Text
	*Drawable
}

// NewClip create clip
func NewClip(p *Point, height int) *Clip {
	text := NewText(p, "")
	text.Color = colorClip()
	list := NewList(p, -1, height, make([]string, 0), make([]int, 0))
	dl := &DrawerList{NewDrawable(p), []Drawer{text, list}, func(pp *Point) *Point {
		return text.Start.DownN(2)
	}}
	box := NewDBox(p, dl)
	box.Color = colorClip()
	return &Clip{false, false, make([]string, 0), list, box, text, NewDrawable(p)}
}

// Draw it
func (c *Clip) Draw() *Point {
	c.showed = true
	if c.showDetail {
		c.list.SetData(c.items, make([]int, len(c.items)), -1)
		c.End = Move(c.box, c.Start)
		return c.End
	}
	c.End = Move(c.text, c.Start)
	return c.End
}

// Clear it
func (c *Clip) Clear() {
	if !c.showed {
		return
	}
	c.showed = false
	c.Drawable.Clear()
}

func (c *Clip) moveTo(p *Point) *Point {
	c.Start = p
	if c.showDetail {
		c.Start = p.Up()
	}
	return c.Draw()
}

// SetData set data
func (c *Clip) SetData(items []string) {
	c.items = items
	c.list.SetData(items, make([]int, len(items)), -1)
	c.text.Data = fmt.Sprintf("[%d clips]", len(c.items))
	c.list.Height = len(items)
}
