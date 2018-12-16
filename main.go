package main

import (
	"os"
	"os/exec"
	"path/filepath"

	"github.com/jacokoo/fff/ui"

	termbox "github.com/nsf/termbox-go"
)

const (
	maxGroups = 4
)

var (
	wo        = newWorkspace()
	home      = os.Getenv("HOME")
	configDir = filepath.Join(home, ".fff")
	wd, _     = os.Getwd()
	quit      = make(chan int)
	message   string
	cfg       = initConfig()
	command   *exec.Cmd
)

func init() {
	initBookmark()
	ui.SetColors(cfg.colors)
}

func start(redraw bool) {
	if err := termbox.Init(); err != nil {
		panic(err)
	}

	w, _ := termbox.Size()
	maxColumns = w/columnWidth + 1

	if redraw {
		uiRedraw()
	} else {
		uiStart()
	}
	kbdStart()

	for {
		switch ev := termbox.PollEvent(); ev.Type {
		case termbox.EventKey:
			quitIt := 0
			if isShell(ev) {
				kbd <- ev
				quitIt = 2
			}

			if quitIt == 0 && isQuit(ev) {
				quitIt = 1
			}

			if quitIt != 0 {
				guiQuit <- true
				termbox.Close()
				kbdQuit <- true
				quit <- quitIt
				return
			}
			kbd <- ev
		case termbox.EventResize:
			termbox.Flush()
		}
	}
}

func main() {
	go start(false)
	for {
		switch ev := <-quit; ev {
		case 1:
			return
		case 2:
			if command == nil {
				go start(true)
				break
			}
			os.Chdir(wo.currentDir())
			command.Stdin = os.Stdin
			command.Stderr = os.Stderr
			command.Stdout = os.Stdout
			err := command.Run()
			if err != nil {
				message = err.Error()
			}
			command = nil
			go start(true)
		}
	}
}
