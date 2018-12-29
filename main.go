package main

import (
	"fmt"
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
	configDir = filepath.Join(home, ".config/fff")
	wd, _     = os.Getwd()
	quit      = make(chan int)
	cfg       = initConfig()
	ac        = new(action)
	tm        = model.NewTaskManager()

	wo         *model.Workspace
	command    *exec.Cmd
	maxColumns int
	gui        *ui.UI
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

func wdFromArgs() {
	if len(os.Args) < 2 {
		return
	}

	s := os.Args[1]
	if !filepath.IsAbs(s) {
		s = filepath.Join(wd, s)
	}

	fi, err := os.Stat(s)
	if err != nil {
		return
	}

	if !fi.IsDir() {
		s = filepath.Dir(s)
	}

	wd = s
}

func checkWd() {
	wdFromArgs()

	if wd == "" {
		wd = "/"
	}

	fi, err := os.Stat(wd)
	if err != nil || !fi.IsDir() {
		panic(err)
	}
}

func main() {
	if len(os.Args) > 1 {
		h := os.Args[1]
		if h == "-h" || h == "--help" {
			fmt.Println(usageString)
			return
		}
	}

	checkWd()
	wo = model.NewWorkspace(maxGroups, wd, configDir)

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

const usageString = `Usage: fff [PATH]

Website: https://github.com/jacokoo/fff`
