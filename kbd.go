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
)

type cmd struct {
	useKey   bool
	key      termbox.Key
	ch       rune
	action   func()
	prefix   bool
	children []*cmd
}

var (
	mode = ModeNormal
	jump = make(chan rune)
	kbd  = make(chan termbox.Event)

	kbds = []*cmd{
		{false, 0, 's', nil, true, []*cmd{
			{false, 0, 'n', ActionSortByName, false, nil},
			{false, 0, 'm', ActionSortByMtime, false, nil},
			{false, 0, 's', ActionSortBySize, false, nil},
		}},
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
		c.action()
		currentKbds = kbds
		return
	}

	currentKbds = c.children
}

// Action
var (
	ActionSortByName = func() {
		wo.sort(orderName)
	}
	ActionSortByMtime = func() {
		wo.sort(orderMTime)
	}
	ActionSortBySize = func() {
		wo.sort(orderSize)
	}
	ActionMoveDown = func() {
	}
	ActionMoveUp = func() {
	}
	ActionOpenFolderRight = func() {
	}
	ActionOpenFolderRoot = func() {
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

func kbdHandleJump(ch rune) {
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
			kbdHandleJump(ch)
		case ModeNormal:
			kbdHandleNormal(key, ch)
		}
	}
}

func kbdStart() {
	go handleKeyEvent()
}
