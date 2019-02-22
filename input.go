package main

import (
	"strings"

	"github.com/jacokoo/fff/model"
	"github.com/jacokoo/fff/ui"
	"github.com/mattn/go-runewidth"
)

var (
	input     = make(chan rune)
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

func deleteLastChar(str string) (string, int) {
	chs := []rune(str)
	ch := chs[len(chs)-1]
	width := runewidth.RuneWidth(ch)
	return string(chs[:len(chs)-1]), width
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

	nn, _ := deleteLastChar(n.name)
	n.name = nn
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
	co.Update()
	ui.ColumnContentChangeEvent.Send(co.Column)
}

func (co *columnInputer) Delete() bool {
	f := co.Filter()
	if len(f) == 0 {
		return false
	}

	nn, _ := deleteLastChar(f)
	co.SetFilter(nn)
	co.Update()
	ui.ColumnContentChangeEvent.Send(co.Column)
	return true
}

func (co *columnInputer) End(abort bool) {
}

func handleInputKey() {
	for {
		select {
		case ch := <-input:
			inputer.Append(ch)
			ui.InputChangeEvent.Send([]string{inputer.Name(), inputer.Get()})
		case <-inputQuit:
			return
		}
	}
}

func enterInputMode(in Inputer) {
	changeMode(ModeInput)
	inputer = in
	ui.InputChangeEvent.Send([]string{inputer.Name(), inputer.Get()})
	go handleInputKey()
}

func quitInputMode(abort bool) {
	inputer.End(abort)
	inputer = nil
	ui.QuitInputEvent.Send(wo.CurrentGroup().Current())
	inputQuit <- true
	changeMode(ModeNormal)
}

func inputDelete() {
	b := inputer.Delete()
	if !b {
		quitInputMode(true)
		return
	}

	ui.InputChangeEvent.Send([]string{inputer.Name(), inputer.Get()})
}

type requestHandler struct {
	isPassword bool
	*nameInputer
}

func (rh *requestHandler) Get() string {
	if rh.isPassword {
		return strings.Repeat("*", len(rh.name))
	}
	return rh.name
}

func handleUserRequest() {
	for {
		req := <-model.RequestCh
		rh := &requestHandler{req.IsPassword, &nameInputer{req.Title, "", func(end string) {
			model.ResponseCh <- end
		}}}
		enterInputMode(rh)
	}
}
