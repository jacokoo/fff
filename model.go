package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
)

const (
	orderName = iota
	orderMTime
	orderSize
)

type column struct {
	path       string
	files      []os.FileInfo
	order      int
	showHidden bool
	current    int
}

func (co *column) sort(order int) {

}

func newColumn(path string) column {
	fs, _ := ioutil.ReadDir(path)
	co := column{path, fs, orderName, false, 0}
	co.sort(orderName)
	return co
}

type group struct {
	path    string
	columns []column
}

func newGroup(path string) group {
	return group{path, []column{newColumn(path)}}
}

func (gr group) currentDir() string {
	co := gr.columns[len(gr.columns)-1]
	return co.path
}

func (gr group) currentSelect() string {
	co := gr.columns[len(gr.columns)-1]
	return filepath.Join(co.path, co.files[co.current].Name())
}

type workspace struct {
	bookmark     []string
	groups       []group
	currentGroup int
	showBookmark bool
}

func newWorkspace() workspace {
	gs := make([]group, maxGroups)
	gs[0] = newGroup(wd)
	bo := []string{"/User/guyong/ws", "/User/guyong/ws/go"}
	return workspace{bo, gs, 0, false}
}

func (wo workspace) currentDir() string {
	return wo.groups[wo.currentGroup].currentDir()
}
