package main

import "github.com/jacokoo/fff/model"

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
	End(bool)
}

type nameInputer struct {
	title  string
	name   string
	action func(string)
}

func newNameInput(title string, action func(string)) *nameInputer {
	return &nameInputer{title, "", action}
}

func (n *nameInputer) Name() string {
	return n.title
}

func (n *nameInputer) Get() string {
	return n.name
}

func (n *nameInputer) Append(ch rune) {
	n.name += string(ch)
}

func (n *nameInputer) Delete() bool {
	if len(n.name) == 0 {
		return false
	}

	n.name = n.name[:len(n.name)-1]
	return true
}

func (n *nameInputer) End(abort bool) {
	if len(n.name) == 0 {
		return
	}
	if !abort {
		go n.action(n.name)
	}
	n.name = ""
}

type columnInputer struct {
	model.Column
}

func (co *columnInputer) Name() string {
	return "FILTER"
}

func (co *columnInputer) Get() string {
	return co.Filter()
}

func (co *columnInputer) Append(ch rune) {
	fi := co.Filter() + string(ch)
	co.SetFilter(fi)
	co.DoFilter()
	gui <- uiColumnContentChange
}

func (co *columnInputer) Delete() bool {
	f := co.Filter()
	if len(f) == 0 {
		return false
	}

	co.SetFilter(f[:len(f)-1])
	co.DoFilter()
	gui <- uiColumnContentChange
	return true
}

func (co *columnInputer) End(abort bool) {
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

func quitInputMode(abort bool) {
	inputer.End(abort)
	inputer = nil
	gui <- uiQuitInput
	inputQuit <- true
	changeMode(ModeNormal)
}

func inputDelete() {
	b := inputer.Delete()
	if !b {
		quitInputMode(true)
		return
	}

	gui <- uiInputChange
}
