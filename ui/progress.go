package ui

import (
	"fmt"
	"strings"

	"github.com/nsf/termbox-go"
)

// ProgressBar a progress bar
type ProgressBar struct {
	Width    int
	Progress int
	max      int
	bg       *Text
	*Drawable
}

// NewProgressBar create progress bar
func NewProgressBar(p *Point, width, progress int) *ProgressBar {
	bg := NewText(p, "")
	bg.Color = colorProgress().Reverse()
	return &ProgressBar{width, progress, 100, bg, NewDrawable(p)}
}

// Draw it
func (pb *ProgressBar) Draw() *Point {
	pw := pb.Width * pb.Progress / pb.max
	pb.bg.Data = strings.Repeat(" ", pw)
	Move(pb.bg, pb.Start)

	s := fmt.Sprintf("%d%%", pb.Progress*100/pb.max)
	start := pb.Width/2 - len(s)/2

	for i, v := range s {
		c := colorProgress()
		if start+i <= pw {
			c = c.Reverse()
		}
		termbox.SetCell(pb.Start.X+start+i, pb.Start.Y, v, c.FG, c.BG)
	}

	pb.End = pb.Start.RightN(pb.Width - 1)
	return pb.End
}
