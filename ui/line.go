package ui

import (
	"strings"

	termbox "github.com/nsf/termbox-go"
)

// HLine horizontal line
type HLine struct {
	*Text
}

// NewHLine create HLine
func NewHLine(p *Point, width int) *HLine {
	return &HLine{NewText(p, strings.Repeat("─", width))}
}

// ChangeWidth change the line width
func (h *HLine) ChangeWidth(width int) {
	h.Text.SetValue(strings.Repeat("─", width))
}

// VLine vertical line
type VLine struct {
	token  rune
	height int
	*Drawable
}

// NewVLine create vline
func NewVLine(p *Point, height int) *VLine {
	return &VLine{'│', height, NewDrawable(p)}
}

// NewVDLine create double vertical line
func NewVDLine(p *Point, height int) *VLine {
	return &VLine{'║', height, NewDrawable(p)}
}

// Draw it
func (v *VLine) Draw() *Point {
	i := 0
	for ; i < v.height; i++ {
		termbox.SetCell(v.Start.X, v.Start.Y+i, v.token, v.FG, v.BG)
	}
	// v.End.X
	v.End.Y = v.Start.Y + i
	return v.End
}

// MoveTo update location
func (v *VLine) MoveTo(p *Point) *Point {
	v.Start = p
	return v.Draw()
}
