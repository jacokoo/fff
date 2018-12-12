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

var (
	keyMap = map[termbox.Key]string{
		termbox.KeyArrowDown:  "down",
		termbox.KeyArrowUp:    "up",
		termbox.KeyArrowRight: "right",
		termbox.KeyArrowLeft:  "left",
		termbox.KeyTab:        "tab",
		termbox.KeySpace:      "space",
		termbox.KeyEnter:      "enter",
		termbox.KeyEsc:        "esc",
		termbox.KeyBackspace:  "backspace",
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

	mode = ModeNormal
	jump = make(chan rune)
	kbd  = make(chan termbox.Event)

	kbds = []*cmd{
		{false, 0, 's', "Prefix, Sort File", "", true, []*cmd{
			{false, 0, 'n', "[n]Sort By Name", "ActionSortByName", false, nil},
			{false, 0, 'm', "[m]Sort By MTime", "ActionSortByMtime", false, nil},
			{false, 0, 's', "[s]Sort By Size", "ActionSortBySize", false, nil},
		}},
		{false, 0, '.', "Toggle show hidden files", "ActionToggleHidden", false, nil},
		{false, 0, 'j', "Move down", "ActionMoveDown", false, nil},
		{false, 0, 'k', "Move up", "ActionMoveUp", false, nil},
		{false, 0, 'l', "Open folder on right", "ActionOpenFolderRight", false, nil},
		{false, 0, 'h', "Go to parent folder", "ActionCloseFolderRight", false, nil},
		{false, 0, ',', "Shift column", "ActionShift", false, nil},
		{false, 0, '<', "Move to first item", "ActionMoveToFirst", false, nil},
		{false, 0, '>', "Move to last item", "ActionMoveToLast", false, nil},
		{true, termbox.KeyCtrlN, 0, "Move down", "ActionMoveDown", false, nil},
		{true, termbox.KeyCtrlP, 0, "Move up", "ActionMoveUp", false, nil},
		{true, termbox.KeyEnter, 0, "Open folder on right", "ActionOpenFolderRight", false, nil},
		{false, 0, 'b', "Prefix, Bookmark manage", "", true, []*cmd{
			{false, 0, 'b', "[b]Toggle show bookmark", "ActionToggleBookmark", false, nil},
		}},
		{false, 0, 'w', "Enter jump mode", "ActionEnterJump", false, nil},
		{false, 0, 'g', "Refresh current dir", "ActionRefresh", false, nil},
		{false, 0, '1', "Change group to 1", "ActionChangeGroup0", false, nil},
		{false, 0, '2', "Change group to 2", "ActionChangeGroup1", false, nil},
		{false, 0, '3', "Change group to 3", "ActionChangeGroup2", false, nil},
		{false, 0, '4', "Change group to 3", "ActionChangeGroup3", false, nil},
	}

	currentKbds = kbds
)

func doAction(key termbox.Key, ch rune) {
	if key == termbox.KeyEsc {
		currentKbds = kbds
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
		currentKbds = kbds
		return
	}

	if !c.prefix {
		ac, has := actions[c.action]
		if has {
			ac()
		}

		currentKbds = kbds
		return
	}

	currentKbds = c.children
}

// Action
var (
	actions = map[string]func(){
		"ActionSortByName":       func() { wo.sort(orderName) },
		"ActionSortByMtime":      func() { wo.sort(orderMTime) },
		"ActionSortBySize":       func() { wo.sort(orderSize) },
		"ActionToggleHidden":     func() { wo.toggleHidden() },
		"ActionMoveDown":         func() { wo.move(1) },
		"ActionMoveUp":           func() { wo.move(-1) },
		"ActionMoveToFirst":      func() { wo.moveToFirst() },
		"ActionMoveToLast":       func() { wo.moveToLast() },
		"ActionOpenFolderRight":  func() { wo.openRight() },
		"ActionCloseFolderRight": func() { wo.closeRight() },
		"ActionShift":            func() { wo.shift() },
		"ActionToggleBookmark":   func() { wo.toggleBookmark() },
		"ActionEnterJump":        func() { enterJumpMode() },
		"ActionChangeGroup0":     func() { wo.changeGroup(0) },
		"ActionChangeGroup1":     func() { wo.changeGroup(1) },
		"ActionChangeGroup2":     func() { wo.changeGroup(2) },
		"ActionChangeGroup3":     func() { wo.changeGroup(3) },
		"ActionRefresh":          func() { wo.refresh() },
	}
)

func isQuit(ev termbox.Event) bool {
	if ev.Key == termbox.KeyCtrlQ {
		return true
	}

	if mode == ModeNormal && ev.Ch == 'q' {
		return true
	}

	return false
}

func kbdHandleNormal(key termbox.Key, ch rune) {
	doAction(key, ch)
}

func kbdHandleJump(key termbox.Key, ch rune) {
	if key == termbox.KeyEsc || key == termbox.KeyEnter {
		quitJumpMode()
		return
	}
	if ch == 0 {
		return
	}
	jump <- ch
}

func kbdHandleInput(key termbox.Key, ch rune) {
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
