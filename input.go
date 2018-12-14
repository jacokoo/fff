package main

var (
	inputQuit = make(chan bool)
	inputer   Inputer
)

// Inputer for input
type Inputer interface {
	Name() string
	Get() string
	Append(ch rune)
	Delete() bool
	End()
}

func handleInputKey() {
	for {
		select {
		case ch := <-input:
			inputer.Append(ch)
			gui <- uiInputChange
		case <-inputQuit:
			return
		}
	}
}

func enterInputMode(in Inputer) {
	changeMode(ModeInput)
	inputer = in
	gui <- uiInputChange
	go handleInputKey()
}

func quitInputMode() {
	updateFileInfo()
	inputer.End()
	inputer = nil
	gui <- uiQuitInput
	inputQuit <- true
	changeMode(ModeNormal)
}

func inputDelete() {
	b := inputer.Delete()
	if !b {
		quitInputMode()
		return
	}

	gui <- uiInputChange
}
