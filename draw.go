package main

import (
	"fmt"
	"strings"

	"github.com/nsf/termbox-go"
)

const (
	chh  = '─'
	chv  = '│'
	chtl = '┌'
	chtr = '┐'
	chbl = '└'
	chbr = '┘'
	chlc = '├'
	chrc = '┤'
	chtc = '┬'
	chbc = '┴'
	chcc = '┼'
)

const (
	cdf = termbox.ColorDefault
	cbl = termbox.ColorBlue
	cgr = termbox.ColorGreen
	cre = termbox.ColorRed
	cbk = termbox.ColorBlack
)

func drawVLine(x, y int) {
	_, h := termbox.Size()
	termbox.SetCell(x, y, chtc, cdf, cdf)
	for i := y + 1; i < h; i++ {
		termbox.SetCell(x, i, chv, cdf, cdf)
	}
}

func drawLine(y int) {
	w, _ := termbox.Size()
	for i := 0; i < w; i++ {
		termbox.SetCell(i, y, chh, cdf, cdf)
	}
}

func drawString(x, y int, str string, fg, bg termbox.Attribute) int {
	i := 0
	for _, v := range str {
		termbox.SetCell(x+i, y, v, fg, bg)
		i++
	}
	return x + i
}

func (wo workspace) drawTabs(x, y int) (xx int) {
	xx = drawString(x, y, "TAB[ ", cgr, cdf)

	c := wo.currentGroup
	for i := 0; i < 4; i++ {
		fg, bg := cdf, cdf
		if i == c {
			fg, bg = cbk, cbl
		}
		xx = drawString(xx, y, fmt.Sprintf("%v", i+1), fg, bg)
		xx = drawString(xx, y, " ", cdf, cdf)
	}
	xx = drawString(xx, y, " ]", cgr, cdf)
	return
}

func (wo workspace) drawCurrent(x, y int) (xx int) {
	xx = drawString(x, y, "CURRENT[ ", cgr, cdf)
	xx = drawString(xx, y, wo.currentDir(), cdf, cdf)
	xx = drawString(xx, y, " ]", cgr, cdf)
	return
}

func (wo workspace) drawWd(x, y int) (xx int) {
	xx = drawString(x, y, "WD[ ", cgr, cdf)
	xx = drawString(xx, y, wd, cdf, cdf)
	xx = drawString(xx, y, " ]", cgr, cdf)
	return
}

func (wo workspace) drawTitle(x, y int) int {
	xx := wo.drawTabs(x, y+1)
	xx = wo.drawWd(xx+2, y+1)
	xx = wo.drawCurrent(xx+2, y+1)

	w, _ := termbox.Size()
	xx = drawString(x, y+2, strings.Repeat(string(chh), w), cdf, cdf)
	return y + 3
}
