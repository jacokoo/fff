package ui

const (
	chh   = '─'
	chv   = '│'
	chtl  = '┌'
	chtr  = '┐'
	chbl  = '└'
	chbr  = '┘'
	chdh  = '═'
	chdv  = '║'
	chdtl = '╔'
	chdtr = '╗'
	chdbl = '╚'
	chdbr = '╝'
)

// Box rect with border
type Box struct {
	singleLine bool
	item       Drawer
	padding    []int
	*Drawable
}

// NewBox create single line box
func NewBox(p *Point, item Drawer, padding ...int) *Box {
	return &Box{true, item, padding, NewDrawable(p)}
}

// NewDBox create double line box
func NewDBox(p *Point, item Drawer, padding ...int) *Box {
	return &Box{false, item, padding, NewDrawable(p)}
}

func (b *Box) ptop() int {
	if len(b.padding) == 0 {
		return 0
	}
	return b.padding[0]
}

func (b *Box) pright() int {
	l := len(b.padding)
	if l == 1 {
		return b.padding[0]
	}
	if l > 1 {
		return b.padding[1]
	}
	return 0
}
func (b *Box) pbottom() int {
	l := len(b.padding)
	if l == 1 {
		return b.padding[0]
	}
	if l > 2 {
		return b.padding[2]
	}
	return 0
}

func (b *Box) pleft() int {
	l := len(b.padding)
	if l == 1 {
		return b.padding[0]
	}
	if l == 3 {
		return b.padding[1]
	}
	if l == 4 {
		return b.padding[3]
	}
	return 0
}

// Draw it
func (b *Box) draw() *Point {
	p := b.Start.RightN(b.pleft() + 1).MoveDownN(b.ptop() + 1)
	p = Move(b.item, p)
	br := p.RightN(b.pright() + 1).MoveDownN(b.pbottom() + 1)
	bl := &Point{b.Start.X, br.Y}
	tl := b.Start
	tr := &Point{br.X, b.Start.Y}
	w := br.X - bl.X
	h := br.Y - tr.Y

	var lt, lb *HLine
	var lr, ll *VLine
	var ctl, ctr, cbl, cbr *Text

	if b.singleLine {
		lt = NewHLine(tl, w)
		lb = NewHLine(bl, w)
		lr = NewVLine(tl, h)
		ll = NewVLine(tr, h)
		ctl = NewText(tl, string(chtl))
		ctr = NewText(tr, string(chtr))
		cbl = NewText(bl, string(chbl))
		cbr = NewText(br, string(chbr))
	} else {
		lt = NewHDLine(tl, w)
		lb = NewHDLine(bl, w)
		lr = NewVDLine(tl, h)
		ll = NewVDLine(tr, h)
		ctl = NewText(tl, string(chdtl))
		ctr = NewText(tr, string(chdtr))
		cbl = NewText(bl, string(chdbl))
		cbr = NewText(br, string(chdbr))
	}

	lt.Color = b.Color
	lb.Color = b.Color
	lr.Color = b.Color
	ll.Color = b.Color
	ctl.Color = b.Color
	ctr.Color = b.Color
	cbl.Color = b.Color
	cbr.Color = b.Color
	lt.Draw()
	lb.Draw()
	lr.Draw()
	ll.Draw()
	ctl.Draw()
	ctr.Draw()
	cbl.Draw()
	cbr.Draw()

	b.End = br

	return br
}

// Draw it
func (b *Box) Draw() *Point {
	b.draw()
	b.Clear() // clear to clear the background of the box
	return b.draw()
}
