package ui

import (
	termbox "github.com/nsf/termbox-go"
)

// StatusItem item in status bar
type StatusItem struct {
	padding int
	*Text
}

// Status bar
type Status struct {
	items []*StatusItem
	*Drawable
}

// StatusBackup backup status bar state
type StatusBackup struct {
	items  []*StatusItem
	status *Status
}

// NewStatus create status bar
func NewStatus() *Status {
	w, h := termbox.Size()
	d := NewDrawable(&Point{0, h - 1})
	d.Color = colorStatus()
	d.End.X = w

	return &Status{nil, d}
}

// Draw it
func (s *Status) Draw() *Point {
	p := s.Start.RightN(0)
	for _, v := range s.items {
		p = v.MoveTo(p)
		p = p.RightN(v.padding)
	}
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
func (s *Status) Set(idx int, str string) *Point {
	s.Clear()
	s.items[idx].Data = str
	s.Draw()
	return s.items[len(s.items)-1].End
}

// Add statusbar item
func (s *Status) Add(padding int) *StatusItem {
	si := &StatusItem{padding, NewText(ZeroPoint, "")}
	si.Color = s.Color
	s.items = append(s.items, si)
	return si
}

// Backup statusbar state
func (s *Status) Backup() *StatusBackup {
	ss := s.items
	s.items = nil
	return &StatusBackup{ss, s}
}

// Restore statusbar state
func (b *StatusBackup) Restore() *Status {
	b.status.items = b.items
	return b.status
}
