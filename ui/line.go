package ui

import (
	"strings"

	"github.com/nsf/termbox-go"
)

// HLine horizontal line
type HLine struct {
	text *Text
}

// NewHLine create HLine
func NewHLine(p *Point, width int) *HLine {
	return &HLine{NewText(p, strings.Repeat("─", width))}
}

// Draw it
func (h *HLine) Draw() *Point {
	return h.text.Draw()
}

// Update it, ignore data
func (h *HLine) Update(p *Point, data interface{}) *Point {
	return h.text.UpdateXY(p)
}

// UpdateXY update location
func (h *HLine) UpdateXY(p *Point) *Point {
	return h.text.UpdateXY(p)
}

// UpdateData ignored
func (h *HLine) UpdateData(data interface{}) *Point {
	return h.text.End
}

// VLine vertical line
type VLine struct {
	height int
	*Drawable
}

// NewVLine create vline
func NewVLine(p *Point, height int) *VLine {
	return &VLine{height, NewDrawable(p)}
}

// Draw it
func (v *VLine) Draw() *Point {
	i := 0
	for ; i < v.height; i++ {
		termbox.SetCell(v.Start.X, v.Start.Y+i, '│', v.fg, v.bg)
	}
	v.End.Y = v.Start.Y + i
	return v.End
}

// Update it, ignore data
func (v *VLine) Update(p *Point, data interface{}) *Point {
	v.Start = p
	return v.Draw()
}

// UpdateXY update location
func (v *VLine) UpdateXY(p *Point) *Point {
	return v.Update(p, nil)
}

// UpdateData ignored
func (v *VLine) UpdateData(data interface{}) *Point {
	return v.End
}
