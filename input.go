package main

import (
	"github.com/nsf/termbox-go"
)

var (
	inputQuit = make(chan bool)
	inputText = ""
)

func handleInputKey() {
	for {
		select {
		case ch := <-input:
			inputText += string(ch)
			gui <- uiInputChange
		case <-inputQuit:
			return
		}
	}
}

func enterInputMode() {
	changeMode(ModeInput)
	inputText = ""
	gui <- uiInputChange
	go handleInputKey()
}

func quitInputMode() {
	changeMode(ModeNormal)
	updateFileInfo()
	termbox.SetCursor(-1, -1)
	termbox.Flush()
	inputQuit <- true
}

func inputDelete() {
	if len(inputText) == 0 {
		quitInputMode()
		return
	}
	inputText = inputText[:len(inputText)-1]
	gui <- uiInputChange
}
