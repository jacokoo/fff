package ui

import (
	termbox "github.com/nsf/termbox-go"
)

// Status bar
type Status struct {
	text *Text
	*Drawable
}

// NewStatus create status bar
func NewStatus() *Status {
	w, h := termbox.Size()
	d := NewDrawable(&Point{0, h - 1})
	d.Color = colorStatus()
	d.End.X = w

	t := NewText(d.Start, "")
	t.Color = d.Color
	return &Status{t, d}
}

// Draw it
func (s *Status) Draw() *Point {
	s.text.Draw()
	return s.End
}

// Clear it
func (s *Status) Clear() {
	for i := s.Start.X; i < s.End.X; i++ {
		termbox.SetCell(i, s.Start.Y, ' ', s.Color.FG, s.Color.BG)
	}
}

// MoveTo update location
func (s *Status) MoveTo(p *Point) *Point {
	return s.End
}

// Set string to status bar
func (s *Status) Set(str string) {
	s.Clear()
	s.text.SetValue(str)
	s.text.Draw()
}
