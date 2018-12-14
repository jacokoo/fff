package main

import (
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
		"ActionSortByName":         func() { wo.sort(orderName) },
		"ActionSortByMtime":        func() { wo.sort(orderMTime) },
		"ActionSortBySize":         func() { wo.sort(orderSize) },
		"ActionToggleHidden":       func() { wo.toggleHidden() },
		"ActionToggleDetail":       func() { wo.toggleDetails() },
		"ActionMoveDown":           func() { wo.move(1) },
		"ActionMoveUp":             func() { wo.move(-1) },
		"ActionMoveToFirst":        func() { wo.moveToFirst() },
		"ActionMoveToLast":         func() { wo.moveToLast() },
		"ActionOpenFolderRight":    func() { wo.openRight() },
		"ActionCloseFolderRight":   func() { wo.closeRight() },
		"ActionShift":              func() { wo.shift() },
		"ActionToggleBookmark":     func() { wo.toggleBookmark() },
		"ActionChangeGroup0":       func() { wo.changeGroup(0) },
		"ActionChangeGroup1":       func() { wo.changeGroup(1) },
		"ActionChangeGroup2":       func() { wo.changeGroup(2) },
		"ActionChangeGroup3":       func() { wo.changeGroup(3) },
		"ActionRefresh":            func() { wo.refresh() },
		"ActionQuitJump":           func() { quitJumpMode() },
		"ActionClearMark":          func() { wo.clearMark() },
		"ActionToggleMark":         func() { wo.toggleMark() },
		"ActionJumpCurrentDirOnce": func() { enterJumpMode(JumpModeCurrentDir, false) },
		"ActionJumpCurrentDir":     func() { enterJumpMode(JumpModeCurrentDir, true) },
		"ActionJumpBookmarkOnce":   func() { enterJumpMode(JumpModeBookmark, false) },
		"ActionJumpBookmark":       func() { enterJumpMode(JumpModeBookmark, true) },
		"ActionJumpAllOnce":        func() { enterJumpMode(JumpModeAll, false) },
		"ActionJumpAll":            func() { enterJumpMode(JumpModeAll, true) },
		"ActionEnterInputMode":     func() { enterInputMode() },
		"ActionQuitInputMode":      func() { quitInputMode() },
		"ActionInputDelete":        func() { inputDelete() },
	}

	mode  = ModeNormal
	jump  = make(chan rune)
	input = make(chan rune)
	kbd   = make(chan termbox.Event)

	currentKbds = cfg.normalKbds
	keyPrefixed = false
)

func changeMode(to Mode) {
	mode = to
	restoreKbds()
}

func restoreKbds() {
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

func doAction(key termbox.Key, ch rune) {
	if key == termbox.KeyEsc && keyPrefixed {
		restoreKbds()
		return
	}

	var c *cmd
	for _, v := range currentKbds {
		if ch == 0 && v.useKey && v.key == key {
			c = v
			break
		} else if !v.useKey && v.ch == ch {
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
	keyPrefixed = true
}

func isQuit(ev termbox.Event) bool {
	for _, kb := range currentKbds {
		if kb.action == "ActionQuit" {
			key, ch := ev.Key, ev.Ch
			if kb.useKey && kb.key == key {
				return true
			}

			if !kb.useKey && kb.ch == ch {
				return true
			}
		}
	}

	return false
}

func kbdHandleNormal(key termbox.Key, ch rune) {
	doAction(key, ch)
}

func kbdHandleJump(key termbox.Key, ch rune) {
	if ch != 0 {
		jump <- ch
		return
	}
	doAction(key, ch)
}

func kbdHandleInput(key termbox.Key, ch rune) {
	if ch != 0 {
		input <- ch
		return
	}
	doAction(key, ch)
}

func handleKeyEvent() {
	for {
		ev := <-kbd
		ch, key := ev.Ch, ev.Key
		switch mode {
		case ModeInput:
			kbdHandleInput(key, ch)
		case ModeJump:
			kbdHandleJump(key, ch)
		case ModeNormal:
			kbdHandleNormal(key, ch)
		}
	}
}

func kbdStart() {
	go handleKeyEvent()
}
