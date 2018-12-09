package main

import (
	"os"
	"path/filepath"
	"strings"

	termbox "github.com/nsf/termbox-go"
)

const (
	maxGroups = 4
)

var (
	wo      = newWorkspace()
	home    = os.Getenv("HOME")
	wd, _   = os.Getwd()
	cui     = make(chan int)
	cuiQuit = make(chan int)
)

func replaceHome(str string) string {
	if strings.HasPrefix(str, home) {
		return filepath.Join("~", str[len(home):])
	}
	return str
}

func main() {
	if err := termbox.Init(); err != nil {
		panic(err)
	}
	defer termbox.Close()

	if len(home) == 0 {
		home = "/root"
	}

	start()
	termbox.Flush()

loop:
	for {
		switch ev := termbox.PollEvent(); ev.Type {
		case termbox.EventKey:
			if ev.Ch == 'q' {
				cuiQuit <- 1
				break loop
			}
			if ev.Ch == '2' {
				cui <- uiChangeGroup
			}

			if ev.Ch == '3' {
				cui <- uiChangeWd
			}

			if ev.Ch == '4' {
				cui <- uiAddColumn
			}
		case termbox.EventResize:
			termbox.Clear(cdf, cdf)
			wo.drawTitle(0, 0)
			termbox.Flush()
		}
	}
}
