package ui

import (
	"fmt"
)

// Clip render clip
type Clip struct {
	showDetail bool
	items      []string
	list       *List
	popup      *Popup
	*Text
}

// NewClip create clip
func NewClip(p *Point, height int) *Clip {
	text := NewText(p, "")
	text.Color = colorClip()
	list := NewList(p, -1, height, make([]string, 0), make([]int, 0))
	box := NewDBox(p, list)
	box.Color = colorClip()
	popup := NewPopup(p, box)
	return &Clip{false, make([]string, 0), list, popup, text}
}

// Draw it
func (c *Clip) Draw() *Point {
	if len(c.items) == 0 {
		return c.End
	}
	return c.Text.Draw()
}

// SetData set data
func (c *Clip) SetData(items []string) {
	c.items = items
	c.list.SetData(items, make([]int, len(items)), -1)
	c.Text.Data = fmt.Sprintf("[%d clips]", len(c.items))
	c.list.Height = len(items)
}

// Open popup
func (c *Clip) Open() {
	c.showDetail = true
	Move(c.popup, c.Start.Down())
}

// Close popup
func (c *Clip) Close() {
	c.showDetail = false
	c.popup.Clear()
}
