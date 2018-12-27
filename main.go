package main

import (
	"os"
	"os/exec"
	"path/filepath"

	"github.com/jacokoo/fff/model"
	"github.com/jacokoo/fff/ui"

	termbox "github.com/nsf/termbox-go"
)

const (
	maxGroups   = 4
	columnWidth = 30
)

var (
	home      = os.Getenv("HOME")
	configDir = filepath.Join(home, ".fff")
	wd, _     = os.Getwd()
	wo        = model.NewWorkspace(maxGroups, wd, configDir)
	quit      = make(chan int)
	cfg       = initConfig()
	command   *exec.Cmd
	ac        = new(action)
	tm        = model.NewTaskManager()

	maxColumns int
	gui        *ui.UI
	clip       model.CopySource
)

func init() {
	ui.SetColors(cfg.colors)
}

func start(redraw bool) {
	if err := termbox.Init(); err != nil {
		panic(err)
	}

	w, _ := termbox.Size()
	maxColumns = w/columnWidth + 1

	if redraw {
		ui.Redraw()
	} else {
		gui = ui.Start(wo)
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
				ui.GuiQuit <- true
				termbox.Close()
				kbdQuit <- true
				quit <- quitIt
				return
			}
			kbd <- ev
		case termbox.EventResize:
			gui = ui.Recreate(wo)
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
			os.Chdir(wo.CurrentGroup().Path())
			command.Stdin = os.Stdin
			command.Stderr = os.Stderr
			command.Stdout = os.Stdout
			err := command.Run()
			command = nil
			go start(true)
			if err != nil {
				ui.MessageEvent.Send("Failed to execute command: " + err.Error())
			}
		}
	}
}
