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
	ModeHelp
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

	limit = func(m Mode, fn func()) func() {
		return func() {
			if m == mode {
				fn()
			}
		}
	}

	actions = map[string]func(){
		"ActionQuit":               func() {},
		"ActionSortByName":         limit(ModeNormal, func() { ac.sort(model.OrderByName) }),
		"ActionSortByMtime":        limit(ModeNormal, func() { ac.sort(model.OrderByMTime) }),
		"ActionSortBySize":         limit(ModeNormal, func() { ac.sort(model.OrderBySize) }),
		"ActionToggleHidden":       limit(ModeNormal, func() { ac.toggleHidden() }),
		"ActionToggleDetail":       limit(ModeNormal, func() { ac.toggleDetails() }),
		"ActionMoveDown":           limit(ModeNormal, func() { ac.move(1) }),
		"ActionMoveUp":             limit(ModeNormal, func() { ac.move(-1) }),
		"ActionMoveToFirst":        limit(ModeNormal, func() { ac.moveToFirst() }),
		"ActionMoveToLast":         limit(ModeNormal, func() { ac.moveToLast() }),
		"ActionOpenFolderRight":    limit(ModeNormal, func() { ac.openRight() }),
		"ActionOpenFile":           limit(ModeNormal, func() { ac.openFile() }),
		"ActionCloseFolderRight":   limit(ModeNormal, func() { ac.closeRight() }),
		"ActionShift":              limit(ModeNormal, func() { ac.shift() }),
		"ActionToggleBookmark":     limit(ModeNormal, func() { ac.toggleBookmark() }),
		"ActionChangeGroup0":       limit(ModeNormal, func() { ac.changeGroup(0) }),
		"ActionChangeGroup1":       limit(ModeNormal, func() { ac.changeGroup(1) }),
		"ActionChangeGroup2":       limit(ModeNormal, func() { ac.changeGroup(2) }),
		"ActionChangeGroup3":       limit(ModeNormal, func() { ac.changeGroup(3) }),
		"ActionRefresh":            limit(ModeJump, func() { ac.refresh() }),
		"ActionQuitJump":           limit(ModeNormal, func() { quitJumpMode() }),
		"ActionClearMark":          limit(ModeNormal, func() { ac.clearMark() }),
		"ActionToggleMark":         limit(ModeNormal, func() { ac.toggleMark() }),
		"ActionJumpCurrentDirOnce": limit(ModeNormal, func() { enterJumpMode(JumpModeCurrentDir, false) }),
		"ActionJumpCurrentDir":     limit(ModeNormal, func() { enterJumpMode(JumpModeCurrentDir, true) }),
		"ActionJumpBookmarkOnce":   limit(ModeNormal, func() { enterJumpMode(JumpModeBookmark, false) }),
		"ActionJumpBookmark":       limit(ModeNormal, func() { enterJumpMode(JumpModeBookmark, true) }),
		"ActionJumpAllOnce":        limit(ModeNormal, func() { enterJumpMode(JumpModeAll, false) }),
		"ActionJumpAll":            limit(ModeNormal, func() { enterJumpMode(JumpModeAll, true) }),
		"ActionStartFilter":        limit(ModeNormal, func() { enterInputMode(&columnInputer{wo.CurrentGroup().Current()}) }),
		"ActionClearFilter":        limit(ModeNormal, func() { ac.clearFilter() }),
		"ActionQuitInputMode":      limit(ModeInput, func() { quitInputMode(false) }),
		"ActionAbortInputMode":     limit(ModeInput, func() { quitInputMode(true) }),
		"ActionInputDelete":        limit(ModeInput, func() { inputDelete() }),
		"ActionNewFile":            limit(ModeNormal, func() { enterInputMode(newFileInputer) }),
		"ActionNewDir":             limit(ModeNormal, func() { enterInputMode(newDirInputer) }),
		"ActionRename":             limit(ModeNormal, func() { enterInputMode(renameInputer) }),
		"ActionAddBookmark":        limit(ModeNormal, func() { enterInputMode(addBookmarkInputer) }),
		"ActionDeleteBookmark":     limit(ModeNormal, func() { enterJumpMode(JumpModeDeleteBookmark, false) }),
		"ActionAppendClip":         limit(ModeNormal, func() { ac.clipFile() }),
		"ActionPaste":              limit(ModeNormal, func() { ac.copyFile() }),
		"ActionMoveFile":           limit(ModeNormal, func() { ac.moveFile() }),
		"ActionClearClip":          limit(ModeNormal, func() { ac.clearClip() }),
		"ActionShowHelp":           limit(ModeNormal, func() { ac.showHelp() }),

		"ActionDeleteFile": limit(ModeNormal, func() {
			s := ac.deletePrompt()
			if s == "" {
				return
			}
			deleteFileInputer.title = s
			enterInputMode(deleteFileInputer)
		}),

		"ActionEdit": limit(ModeNormal, func() {
			file, err := wo.CurrentGroup().Current().CurrentFile()
			if err != nil || file.IsDir() {
				return
			}
			command = cfg.cmd(fmt.Sprintf("%s %s", cfg.editor, file.Path()))
		}),
		"ActionView": limit(ModeNormal, func() {
			file, err := wo.CurrentGroup().Current().CurrentFile()
			if err != nil || file.IsDir() {
				return
			}
			command = cfg.cmd(fmt.Sprintf("%s %s", cfg.pager, file.Path()))
		}),
		"ActionShell": limit(ModeNormal, func() {
			command = exec.Command(cfg.shell)
		}),
	}

	mode    = ModeNormal
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
			case ModeHelp:
				ac.closeHelp()
				restoreKbds()
			}
		case <-kbdQuit:
			return
		}
	}
}

func kbdStart() {
	go handleKeyEvent()
}
