package main

import (
	"unicode"

	"github.com/jacokoo/fff/ui"
)

// the root is empty
// the first level is head key, won't change
// flat the second level
type keyTree struct {
	key      rune
	parent   *keyTree
	children []*keyTree
	item     *ui.JumpItem
}

func have1st(kt *keyTree, key rune) bool {
	for _, v := range kt.children {
		if v.key == key {
			return true
		}
	}
	return false
}

func have2nd(kt *keyTree, key rune) bool {
	for _, v := range kt.children {
		for _, vv := range v.children {
			if vv.key == key {
				return true
			}
		}
	}
	return false
}

func findKey(kt *keyTree, have func(*keyTree, rune) bool) rune {
	key := ' '
	for i := 'a'; i <= 'z'; i++ {
		if !have(kt, i) {
			key = i
			break
		}
	}

	if key != ' ' {
		return key
	}

	for i := 'A'; i <= 'Z'; i++ {
		if !have(kt, i) {
			key = i
			break
		}
	}
	return key
}

func flat1st(root *keyTree) {
	for _, v := range root.children {
		if v.key == '-' {
			v.key = findKey(root, have1st)
		}

		for _, vv := range v.children {
			flat3rd(vv)
		}
	}
}

func flat3rd(kt *keyTree) {
	if kt.key == '-' {
		kt.key = findKey(kt.parent, have2nd)
	}
	cc := len(kt.children)
	if len(kt.parent.children) == 1 && cc == 1 {
		kt.children[0].item.Key = []rune{kt.parent.key}
		return
	}

	if cc == 1 {
		kt.children[0].item.Key = []rune{kt.parent.key, kt.key}
		return
	}

	kt.children[0].key = kt.key
	kt.children[0].item.Key = []rune{kt.parent.key, kt.key}

	for i := 1; i < cc; i++ {
		kt.children[i].key = findKey(kt.parent, have2nd)
		kt.children[i].item.Key = []rune{kt.parent.key, kt.children[i].key}
	}
}

func (kt *keyTree) add(idx int, item *ui.JumpItem) {
	if idx >= len(item.Key) {
		kt.children = append(kt.children, &keyTree{'$', kt, nil, item})
		return
	}
	k := unicode.ToLower(item.Key[idx])
	if (k < 'a' || k > 'z') && (k < '0' || k > '9') {
		k = '-'
	}

	var p *keyTree
	for _, v := range kt.children {
		if v.key == k {
			p = v
			break
		}
	}

	if p == nil {
		p = &keyTree{k, kt, nil, nil}
		kt.children = append(kt.children, p)
	}

	p.add(idx+1, item)
}

// JumpMode describe the jump mode
type JumpMode uint8

// Jump
const (
	JumpModeAll JumpMode = iota
	JumpModeBookmark
	JumpModeCurrentDir
	JumpModeDeleteBookmark
)

var (
	jump         = make(chan rune)
	jumpQuit     = make(chan bool)
	continueJump = false
	jumpItems    []*ui.JumpItem
)

func keyThem(items []*ui.JumpItem) {
	tree := &keyTree{' ', nil, nil, nil}
	for _, v := range items {
		tree.add(0, v)
	}
	flat1st(tree)
}

func handleJumpResult(item *ui.JumpItem) {
	ui.GuiNeedAck = true
	co := item.Action()
	<-ui.GuiAck
	ui.GuiNeedAck = false

	if !co || !continueJump {
		quitJumpMode()
		return
	}

	items := collectCurrentDir()
	if len(items) == 0 {
		quitJumpMode()
		return
	}

	keyThem(items)

	jumpItems = items
	changeMode(ModeJump)
	ui.JumpRefreshEvent.Send(jumpItems)
}

func handleKeys() {
	for {
	sc:
		select {
		case ch := <-jump:
			changeMode(ModeDisabled)
			var got = false
			for _, it := range jumpItems {
				if len(it.Key) == 0 {
					continue
				}
				if it.Key[0] != ch {
					it.Key = nil
					continue
				}
				if len(it.Key) == 1 {
					go handleJumpResult(it)
					break sc
				}

				it.Key = it.Key[1:]
				got = true
			}
			if got {
				ui.JumpRefreshEvent.Send(jumpItems)
				changeMode(ModeJump)
			} else {
				go quitJumpMode()
			}
		case <-jumpQuit:
			return
		}
	}
}

func collectAllDir() []*ui.JumpItem {
	items := make([]*ui.JumpItem, 0)
	ui.EachFileList(func(colIdx int, list *ui.List) {
		items = append(items, list.JumpItems(func(idx int) func() bool {
			return func() bool {
				return ac.jumpTo(colIdx, idx, continueJump)
			}
		})...)
	})
	return items
}

func collectBookmark(forDelete bool) []*ui.JumpItem {
	if !wo.IsShowBookmark() {
		return nil
	}
	bk := wo.Bookmark
	return ui.BookmarkList().JumpItems(func(idx int) func() bool {
		key := bk.Names[idx]
		if forDelete && bk.IsFixed(key) {
			return nil
		}
		fn := func() bool {
			v, has := bk.Get(key)
			if !has {
				return false
			}
			ac.openRoot(v)
			return true
		}
		if forDelete {
			fn = func() bool {
				ac.deleteBookmark(key)
				return false
			}
		}
		return fn
	})
}

func collectCurrentDir() []*ui.JumpItem {
	return ui.CurrentFileList().JumpItems(func(idx int) func() bool {
		return func() bool {
			return ac.jumpTo(len(wo.CurrentGroup().Columns())-1, idx, continueJump)
		}
	})
}

func collectGroups() []*ui.JumpItem {
	return gui.Tab.JumpItems(func(idx int) func() bool {
		return func() bool {
			ac.changeGroup(idx)
			return false
		}
	})
}

func collectCurrentPath() []*ui.JumpItem {
	return gui.Path.JumpItems(func(path string) func() bool {
		if path == "/" {
			return nil
		}
		return func() bool {
			ac.openRoot(path)
			return true
		}
	})
}

func enterJumpMode(md JumpMode, cj bool) {
	switch md {
	case JumpModeBookmark:
		jumpItems = collectBookmark(false)
	case JumpModeDeleteBookmark:
		jumpItems = collectBookmark(true)
	case JumpModeCurrentDir:
		jumpItems = collectCurrentDir()
	case JumpModeAll:
		jumpItems = append(collectBookmark(false), collectCurrentPath()...)
		jumpItems = append(jumpItems, collectAllDir()...)
	}
	keyThem(jumpItems)
	continueJump = cj

	ui.JumpRefreshEvent.Send(jumpItems)
	go handleKeys()
	changeMode(ModeJump)
}

func quitJumpMode() {
	if mode != ModeJump && mode != ModeDisabled {
		return
	}
	jumpQuit <- true
	jumpItems = nil
	ui.JumpRefreshEvent.Send(jumpItems)
	changeMode(ModeNormal)
}
