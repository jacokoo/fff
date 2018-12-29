package ui

import (
	termbox "github.com/nsf/termbox-go"
)

var (
	// ColorDefault default color
	ColorDefault = &Color{termbox.ColorDefault, termbox.ColorDefault}
	colors       map[string]*Color
)

// Color represent the forecolor and background color of an point
type Color struct {
	FG, BG termbox.Attribute
}

// SetColors set colors to use
func SetColors(cs map[string]*Color) {
	colors = cs
}

func getColor(name string) *Color {
	cs, has := colors[name]
	if has {
		return cs
	}
	return ColorDefault
}

//Reverse the color
func (c *Color) Reverse() *Color {
	return &Color{c.FG, c.BG | termbox.AttrReverse}
}

func colorKeyword() *Color        { return getColor("keyword") }
func colorNormal() *Color         { return getColor("normal") }
func colorTab() *Color            { return getColor("tab") }
func colorFile() *Color           { return getColor("file") }
func colorFolder() *Color         { return getColor("folder") }
func colorStatus() *Color         { return getColor("statusbar") }
func colorMarked() *Color         { return getColor("marked") }
func colorIndicator() *Color      { return getColor("indicator") }
func colorJump() *Color           { return getColor("jump") }
func colorFilter() *Color         { return getColor("filter") }
func colorStatusBarTitle() *Color { return getColor("statusbar-title") }
func colorClip() *Color           { return getColor("clip") }
