package main

import (
	"fmt"
	"os"

	"github.com/jacokoo/fff/ui"

	termbox "github.com/nsf/termbox-go"
)

const (
	maxGroups = 4
)

var (
	wo      = newWorkspace()
	home    = os.Getenv("HOME")
	wd, _   = os.Getwd()
	cfg     = initConfig()
	quit    = make(chan int)
	message string
)

func main() {
	ui.SetColors(cfg.colors)
	if err := termbox.Init(); err != nil {
		panic(err)
	}
	defer termbox.Close()

	w, _ := termbox.Size()
	maxColumns = w / columnWidth

	if len(home) == 0 {
		home = "/root"
	}
	uiStart()
	kbdStart()

loop:
	for {
		switch ev := termbox.PollEvent(); ev.Type {
		case termbox.EventKey:
			if isQuit(ev) {
				fmt.Println("hello")
				break loop
			}
			kbd <- ev
		case termbox.EventResize:
			termbox.Flush()
		}
	}
}
