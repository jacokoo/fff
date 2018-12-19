package main

import (
	"fmt"
	"os/exec"

	"github.com/jacokoo/fff/ui"

	"github.com/jacokoo/fff/model"

	termbox "github.com/nsf/termbox-go"
)

// Mode for key binding
type Mode uint8

// Mode
const (
	ModeNormal Mode = iota
	ModeJump
	ModeInput
	ModeDisabled
)

type cmd struct {
	useKey   bool
	key      termbox.Key
	ch       rune
	desc     string
	action   string
	prefix   bool
	children []*cmd
}

func (c *cmd) match(ev termbox.Event) bool {
	key, ch := ev.Key, ev.Ch
	if (c.useKey && c.key == key) || (!c.useKey && c.ch == ch) {
		return true
	}
	return false
}

func newCmd(key, action string, children []*cmd) *cmd {
	c := &cmd{false, 0, 0, "", action, false, children}
	kk, has := nameToKey[key]
	if has {
		c.useKey = true
		c.key = kk
	} else {
		c.ch = []rune(key)[0]
	}

	if children != nil {
		c.prefix = true
	}
	return c
}

var (
	keyToName = map[termbox.Key]string{
		termbox.KeyArrowDown:  "down",
		termbox.KeyArrowUp:    "up",
		termbox.KeyArrowRight: "right",
		termbox.KeyArrowLeft:  "left",
		termbox.KeyTab:        "tab",
		termbox.KeySpace:      "space",
		termbox.KeyEnter:      "enter",
		termbox.KeyEsc:        "esc",
		termbox.KeyBackspace2: "backspace",
		termbox.KeyHome:       "home",
		termbox.KeyEnd:        "end",
		termbox.KeyPgdn:       "pagedown",
		termbox.KeyPgup:       "pageup",
		termbox.KeyInsert:     "insert",
		termbox.KeyDelete:     "delete",

		termbox.KeyCtrlBackslash:  "ctrl-\\",
		termbox.KeyCtrlRsqBracket: "ctrl-]",
		termbox.KeyCtrlA:          "ctrl-a",
		termbox.KeyCtrlB:          "ctrl-b",
		termbox.KeyCtrlC:          "ctrl-c",
		termbox.KeyCtrlD:          "ctrl-d",
		termbox.KeyCtrlE:          "ctrl-e",
		termbox.KeyCtrlF:          "ctrl-f",
		termbox.KeyCtrlG:          "ctrl-g",
		termbox.KeyCtrlJ:          "ctrl-j",
		termbox.KeyCtrlK:          "ctrl-k",
		termbox.KeyCtrlL:          "ctrl-l",
		termbox.KeyCtrlN:          "ctrl-n",
		termbox.KeyCtrlO:          "ctrl-o",
		termbox.KeyCtrlP:          "ctrl-p",
		termbox.KeyCtrlQ:          "ctrl-q",
		termbox.KeyCtrlR:          "ctrl-r",
		termbox.KeyCtrlS:          "ctrl-s",
		termbox.KeyCtrlT:          "ctrl-t",
		termbox.KeyCtrlU:          "ctrl-u",
		termbox.KeyCtrlV:          "ctrl-v",
		termbox.KeyCtrlW:          "ctrl-w",
		termbox.KeyCtrlX:          "ctrl-x",
		termbox.KeyCtrlY:          "ctrl-y",
		termbox.KeyCtrlZ:          "ctrl-z",
	}

	nameToKey = func() map[string]termbox.Key {
		mp := make(map[string]termbox.Key, len(keyToName))
		for k, v := range keyToName {
			mp[v] = k
		}
		return mp
	}()

	actions = map[string]func(){
		"ActionQuit":               func() {},
		"ActionSortByName":         func() { ac.sort(model.OrderByName) },
		"ActionSortByMtime":        func() { ac.sort(model.OrderByMTime) },
		"ActionSortBySize":         func() { ac.sort(model.OrderBySize) },
		"ActionToggleHidden":       func() { ac.toggleHidden() },
		"ActionToggleDetail":       func() { ac.toggleDetails() },
		"ActionMoveDown":           func() { ac.move(1) },
		"ActionMoveUp":             func() { ac.move(-1) },
		"ActionMoveToFirst":        func() { ac.moveToFirst() },
		"ActionMoveToLast":         func() { ac.moveToLast() },
		"ActionOpenFolderRight":    func() { ac.openRight() },
		"ActionOpenFile":           func() { ac.openFile() },
		"ActionCloseFolderRight":   func() { ac.closeRight() },
		"ActionShift":              func() { ac.shift() },
		"ActionToggleBookmark":     func() { ac.toggleBookmark() },
		"ActionChangeGroup0":       func() { ac.changeGroup(0) },
		"ActionChangeGroup1":       func() { ac.changeGroup(1) },
		"ActionChangeGroup2":       func() { ac.changeGroup(2) },
		"ActionChangeGroup3":       func() { ac.changeGroup(3) },
		"ActionRefresh":            func() { ac.refresh() },
		"ActionQuitJump":           func() { quitJumpMode() },
		"ActionClearMark":          func() { ac.clearMark() },
		"ActionToggleMark":         func() { ac.toggleMark() },
		"ActionJumpCurrentDirOnce": func() { enterJumpMode(JumpModeCurrentDir, false) },
		"ActionJumpCurrentDir":     func() { enterJumpMode(JumpModeCurrentDir, true) },
		"ActionJumpBookmarkOnce":   func() { enterJumpMode(JumpModeBookmark, false) },
		"ActionJumpBookmark":       func() { enterJumpMode(JumpModeBookmark, true) },
		"ActionJumpAllOnce":        func() { enterJumpMode(JumpModeAll, false) },
		"ActionJumpAll":            func() { enterJumpMode(JumpModeAll, true) },
		"ActionStartFilter":        func() { enterInputMode(&columnInputer{wo.CurrentGroup().Current()}) },
		"ActionClearFilter":        func() { ac.clearFilter() },
		"ActionQuitInputMode":      func() { quitInputMode(false) },
		"ActionAbortInputMode":     func() { quitInputMode(true) },
		"ActionInputDelete":        func() { inputDelete() },
		"ActionNewFile":            func() { enterInputMode(newFileInputer) },
		"ActionNewDir":             func() { enterInputMode(newDirInputer) },
		"ActionRename":             func() { enterInputMode(renameInputer) },
		"ActionAddBookmark":        func() { enterInputMode(addBookmarkInputer) },
		"ActionDeleteBookmark":     func() { enterJumpMode(JumpModeDeleteBookmark, false) },

		"ActionDeleteFile": func() {
			s := ac.deletePrompt()
			if s == "" {
				return
			}
			deleteFileInputer.title = s
			enterInputMode(deleteFileInputer)
		},

		"ActionEdit": func() {
			file, err := wo.CurrentGroup().Current().CurrentFile()
			if err != nil || file.IsDir() {
				return
			}
			command = cfg.cmd(fmt.Sprintf("%s %s", cfg.editor, file.Path()))
		},
		"ActionView": func() {
			file, err := wo.CurrentGroup().Current().CurrentFile()
			if err != nil || file.IsDir() {
				return
			}
			command = cfg.cmd(fmt.Sprintf("%s %s", cfg.pager, file.Path()))
		},
		"ActionShell": func() {
			command = exec.Command(cfg.shell)
		},
	}

	mode    = ModeNormal
	jump    = make(chan rune)
	input   = make(chan rune)
	kbd     = make(chan termbox.Event)
	kbdQuit = make(chan bool)

	currentKbds        = cfg.normalKbds
	keyPrefixed        = false
	newFileInputer     = newNameInput("NEW FILE", func(name string) { ac.newFile(name) })
	newDirInputer      = newNameInput("NEW DIR", func(name string) { ac.newDir(name) })
	renameInputer      = newNameInput("RENAME", func(name string) { ac.rename(name) })
	addBookmarkInputer = newNameInput("BOOKMARK NAME", func(name string) { ac.addBookmark(name, wo.CurrentGroup().Path()) })

	deleteFileInputer = newNameInput("", func(name string) {
		if name == "y" {
			ac.deleteFiles()
		}
	})
)

func changeMode(to Mode) {
	mode = to
	restoreKbds()
}

func restoreKbds() {
	if keyPrefixed {
		ui.MessageEvent.Send("")
	}
	keyPrefixed = false

	switch mode {
	case ModeNormal:
		currentKbds = cfg.normalKbds
	case ModeJump:
		currentKbds = cfg.jumpKbds
	case ModeInput:
		currentKbds = cfg.inputKbds
	default:
		currentKbds = nil
	}
}

func doAction(ev termbox.Event) {
	if ev.Key == termbox.KeyEsc && keyPrefixed {
		restoreKbds()
		return
	}

	var c *cmd
	for _, v := range currentKbds {
		if v.match(ev) {
			c = v
			break
		}
	}

	if c == nil {
		restoreKbds()
		return
	}

	if !c.prefix {
		ac, has := actions[c.action]
		if has {
			ac()
		}

		restoreKbds()
		return
	}

	currentKbds = c.children
	m := ""
	for i, v := range currentKbds {
		key := string(v.ch)
		if v.useKey {
			key = keyToName[v.key]
		}
		m = fmt.Sprintf("%s[%s]%s", m, key, v.desc)
		if i != len(currentKbds)-1 {
			m += "    "
		}
	}
	ui.MessageEvent.Send(m)
	keyPrefixed = true
}

func isQuit(ev termbox.Event) bool {
	for _, kb := range currentKbds {
		if kb.match(ev) && kb.action == "ActionQuit" {
			return true
		}
	}

	return false
}

func isShell(ev termbox.Event) bool {
	if mode != ModeNormal {
		return false
	}

	for _, kb := range currentKbds {
		if !kb.match(ev) {
			continue
		}
		if kb.action == "ActionShell" {
			return true
		}
		if kb.action == "ActionEdit" || kb.action == "ActionView" {
			file, err := wo.CurrentGroup().Current().CurrentFile()
			return err == nil && !file.IsDir()
		}
	}

	return false
}

func kbdHandleNormal(ev termbox.Event) {
	doAction(ev)
}

func kbdHandleJump(ev termbox.Event) {
	if ev.Ch != 0 {
		jump <- ev.Ch
		return
	}
	doAction(ev)
}

func kbdHandleInput(ev termbox.Event) {
	if ev.Ch != 0 {
		input <- ev.Ch
		return
	}

	if ev.Key == termbox.KeySpace {
		input <- ' '
		return
	}

	doAction(ev)
}

func handleKeyEvent() {
	for {
		select {
		case ev := <-kbd:
			switch mode {
			case ModeInput:
				kbdHandleInput(ev)
			case ModeJump:
				kbdHandleJump(ev)
			case ModeNormal:
				kbdHandleNormal(ev)
			}
		case <-kbdQuit:
			return
		}
	}
}

func kbdStart() {
	go handleKeyEvent()
}
