package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	termbox "github.com/nsf/termbox-go"
)

const (
	maxGroups = 4
)

var (
	wo    = newWorkspace()
	home  = os.Getenv("HOME")
	wd, _ = os.Getwd()
	quit  = make(chan int)
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
