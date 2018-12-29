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
	*Drawable
}

// NewBox create single line box
func NewBox(p *Point, item Drawer) *Box {
	return &Box{true, item, NewDrawable(p)}
}

// NewDBox create double line box
func NewDBox(p *Point, item Drawer) *Box {
	return &Box{false, item, NewDrawable(p)}
}

// Draw it
func (b *Box) draw() *Point {
	p := b.Start.Right().MoveDown()
	p = b.item.MoveTo(p)
	br := p.RightN(3).MoveDown()
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
	b.Clear()
	return b.draw()
}

// MoveTo update location
func (b *Box) MoveTo(p *Point) *Point {
	b.Start = p
	return b.Draw()
}
