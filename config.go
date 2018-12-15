package main

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/jacokoo/fff/ui"
	termbox "github.com/nsf/termbox-go"
	yaml "gopkg.in/yaml.v2"
)

var data = []byte(`
binding:
  # bindings for all mode
  all:
    "ctrl-q": ActionQuit					        # quit fff
  
  # bindings for normal mode
  normal:
    "s":                                  # Prefix, Sort File
      "n": ActionSortByName               # Sort By Name
      "m": ActionSortByMtime              # Sort By MTime
      "s": ActionSortBySize               # Sort By Size
    ".": ActionToggleHidden               # Toggle show hidden files
    "d": ActionToggleDetail               # Toggle show file details
    "j": ActionMoveDown                   # Move down
    "k": ActionMoveUp                     # Move up
    "l": ActionOpenFolderRight            # Open folder on right
    "h": ActionCloseFolderRight           # Go to parent folder
    ",": ActionShift                      # Shift column
    "K": ActionMoveToFirst                # Move to first item
    "J": ActionMoveToLast                 # Move to last item
    "ctrl-n": ActionMoveDown              # Move down
    "ctrl-p": ActionMoveUp                # Move up
    "enter": ActionOpenFolderRight        # Open folder on right
    "b":                                  # Prefix, Bookmark manage
      "b": ActionToggleBookmark           # Toggle show bookmark
      "n": ActionAddBookmark              # Bookmark current dir
      "d": ActionDeleteBookmark           # Delete bookmark
    "r": ActionRefresh                    # Refresh current dir
    "1": ActionChangeGroup0               # Change group to 1
    "2": ActionChangeGroup1               # Change group to 2
    "3": ActionChangeGroup2               # Change group to 3
    "4": ActionChangeGroup3               # Change group to 4
    "q": ActionQuit                       # quit fff
    "up": ActionMoveUp                    # Move up
    "down": ActionMoveDown                # Move down
    "right": ActionOpenFolderRight        # Open folder on right
    "left": ActionCloseFolderRight        # Go to parent folder
    "m": ActionToggleMark                 # Toggle mark
    "u": ActionClearMark                  # Clear all marks
    "g": ActionJumpCurrentDirOnce         # Jump over current dir and stop after one jump
    "G": ActionJumpCurrentDir             # Jump over current dir
    "i": ActionJumpAllOnce                # Jump over items that can jump and stop after one jump
	"I": ActionJumpAll                    # Jump over items that can jump
	"W": ActionJumpBookmark               # Jump over bookmarks
	"w": ActionJumpBookmarkOnce           # Jump over bookmarks and stop after one jump
    "f": ActionStartFilter                # Filter
    "F": ActionClearFilter                # Clear filter
    "+": ActionNewDir                     # Create new dir in current dir
    "N": ActionNewFile                    # Create new file in current dir
    "R": ActionRename                     # Rename current file

  # bindings for jump mode
  jump:
    "enter": ActionQuitJump
    "esc": ActionQuitJump

  input:
    "enter": ActionQuitInputMode
    "esc": ActionAbortInputMode
    "backspace": ActionInputDelete

color:
  normal: default
  keyword: cyan
  folder: cyan
  file: default
  marked: yellow
  statusbar: cyan
  statusbar-title: magenta
  tab: cyan
  jump: yellow
  filter: magenta
  indicator: green

`)

var colorMap = map[string]termbox.Attribute{
	"default": termbox.ColorDefault,
	"black":   termbox.ColorBlack,
	"blue":    termbox.ColorBlue,
	"cyan":    termbox.ColorCyan,
	"green":   termbox.ColorGreen,
	"magenta": termbox.ColorMagenta,
	"red":     termbox.ColorRed,
	"white":   termbox.ColorWhite,
	"yellow":  termbox.ColorYellow,
}

type config struct {
	normalKbds []*cmd
	jumpKbds   []*cmd
	inputKbds  []*cmd
	colors     map[string]*ui.Color
}

func (c *config) color(name string) *ui.Color {
	cc, has := c.colors[name]
	if has {
		return cc
	}
	return ui.ColorDefault
}

func readColors(ds interface{}, cfg *config) {
	dd, suc := ds.(map[interface{}]interface{})
	if !suc {
		return
	}
	for k, v := range dd {
		kk := fmt.Sprintf("%v", k)
		vv := fmt.Sprintf("%v", v)
		c, ex := colorMap[vv]
		if !ex {
			continue
		}
		if k == "tab" || k == "statusbar" || k == "jump" || k == "filter" {
			cfg.colors[kk] = &ui.Color{FG: c, BG: termbox.ColorDefault | termbox.AttrReverse}
		} else {
			cfg.colors[kk] = &ui.Color{FG: c, BG: termbox.ColorDefault}
		}
	}
}

func createCmd(key, action string) *cmd {
	return newCmd(key, action, nil)
}
func readChildren(ds map[interface{}]interface{}) []*cmd {
	cds := make([]*cmd, 0, len(ds))
	for k, v := range ds {
		kk, e1 := k.(string)
		vv, e2 := v.(string)
		if !e1 || !e2 {
			panic(fmt.Sprintf("key [%v] or action [%v] must be string", k, v))
		}
		cds = append(cds, newCmd(kk, vv, nil))
	}
	return cds
}

func readBinding(ds interface{}) []*cmd {
	dd, suc := ds.(map[interface{}]interface{})
	re := make([]*cmd, 0)
	if !suc {
		return re
	}

	for k, v := range dd {
		kk := fmt.Sprintf("%v", k)
		switch vv := v.(type) {
		case string:
			re = append(re, newCmd(kk, vv, nil))
		case map[interface{}]interface{}:
			cs := readChildren(vv)
			re = append(re, newCmd(kk, "", cs))

		}
	}
	return re
}

func readBindings(ds interface{}, cfg *config) {
	dd, suc := ds.(map[interface{}]interface{})
	if !suc {
		return
	}

	all := readBinding(dd["all"])
	cfg.normalKbds = append(all, cfg.normalKbds...)
	cfg.jumpKbds = append(all, cfg.jumpKbds...)
	cfg.inputKbds = append(all, cfg.inputKbds...)

	cfg.normalKbds = append(readBinding(dd["normal"]), cfg.normalKbds...)
	cfg.jumpKbds = append(readBinding(dd["jump"]), cfg.jumpKbds...)
	cfg.inputKbds = append(readBinding(dd["input"]), cfg.inputKbds...)
}

func readYaml(ds []byte, cfg *config) {
	var mp map[string]interface{}
	yaml.Unmarshal(ds, &mp)
	co, has := mp["color"]
	if has {
		readColors(co, cfg)
	}

	bd, has := mp["binding"]
	if has {
		readBindings(bd, cfg)
	}
}

func initConfig() *config {
	c := &config{colors: make(map[string]*ui.Color)}
	readYaml(data, c)

	f, err := ioutil.ReadFile(filepath.Join(configDir, "config.yml"))
	if err == nil {
		readYaml(f, c)
	}

	return c
}
