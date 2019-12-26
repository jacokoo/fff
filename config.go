package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/jacokoo/fff/ui"
	termbox "github.com/nsf/termbox-go"
	yaml "gopkg.in/yaml.v2"
)

var data = []byte(`
binding:
  # bindings for all mode
  all:
    "ctrl-q": ActionQuit                  # quit fff

  # bindings for normal mode
  normal:
    "s":                                  # Prefix, Sort File
      "n": ActionSortByName               ; Sort By Name
      "m": ActionSortByMtime              ; Sort By MTime
      "s": ActionSortBySize               ; Sort By Size
    ".": ActionToggleHidden               # Toggle show hidden files
    "d": ActionToggleDetail               # Toggle show file details
    "j": ActionMoveDown                   # Move down
    "k": ActionMoveUp                     # Move up
    "l": ActionOpenFolderRight            # Open folder on right
    "h": ActionCloseFolderRight           # Go to parent folder
    "enter": ActionOpenFile               # Open file
    ",": ActionShift                      # Shift column
    "K": ActionMoveToFirst                # Move to first item
    "J": ActionMoveToLast                 # Move to last item
    "b":                                  # Prefix, Bookmark manage
      "b": ActionToggleBookmark           ; Toggle show bookmark
      "n": ActionAddBookmark              ; Bookmark current dir
      "d": ActionDeleteBookmarkOnce       ; Delete bookmark
      "D": ActionDeleteBookmark           ; Delete multiple bookmark
      "w": ActionJumpBookmarkOnce         ; Jump Once
      "W": ActionJumpBookmark             ; Jump
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
    "u": ActionToggleMarkAll              # Clear all marks
    "i": ActionJumpCurrentDirOnce         # Jump over current dir and stop after one jump
    "I": ActionJumpCurrentDir             # Jump over current dir
    "w": ActionJumpAllOnce                # Jump over items that can jump and stop after one jump
    "W": ActionJumpAll                    # Jump over items that can jump
    "f": ActionStartFilter                # Filter
    "F": ActionClearFilter                # Clear filter
    "g": ActionRefresh                    # Refresh current dir
    "+": ActionNewDir                     # Create new dir in current dir
    "N": ActionNewFile                    # Create new file in current dir
    "R": ActionRename                     # Rename current file
    "D": ActionDeleteFile                 # Delete marked files or current file
    "C": ActionAppendClip                 # Append file to clip
    "U": ActionClearClip                  # Clear clip
    "P": ActionPaste                      # Paste file
    "M": ActionMoveFile                   # Move file
    "!": ActionShell                      # Run shell
    "e": ActionEdit                       # Run editor
    "v": ActionView                       # Run pager
    "?": ActionShowHelp                   # Show help
    "-": ActionGoBack                     # Go back to previous dir
    "t":
      "c": ActionShowClipDetail           ; Show clip detail
      "t": ActionShowTaskDetail           ; Show task detail
      "d": ActionCloseTaskDetail          ; Close task detail
      "f": ActionFakeTask                 ; Fake task

  # bindings for jump mode
  jump:
    "enter": ActionQuitJump
    "esc": ActionQuitJump

  input:
    "enter": ActionQuitInputMode
    "esc": ActionAbortInputMode
    "backspace": ActionInputDelete

  clip:
    "w": ActionDeleteClipOnce             # Jump to delete clip once
    "W": ActionDeleteClip                 # Jump to delete clip

  task:
    "w": ActionCancelTaskOnce             # Jump to cancel task once
    "W": ActionCancelTask                 # Jump to cancel task

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
  clip: yellow

editor: vi
shell: sh
pager: less
single-column-mode: false
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
	normalKbds       []*cmd
	jumpKbds         []*cmd
	inputKbds        []*cmd
	clipKbds         []*cmd
	taskKbds         []*cmd
	colors           map[string]*ui.Color
	editor           string
	shell            string
	pager            string
	singleColumnMode bool
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
		switch k {
		case "tab", "statusbar", "jump", "filter":
			cfg.colors[kk] = &ui.Color{FG: c, BG: termbox.ColorDefault | termbox.AttrReverse}
		default:
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
		cmd := newCmd(kk, vv, nil)
		idx := strings.Index(vv, ";")
		if idx != -1 {
			cmd.action = strings.Trim(vv[:idx], " ")
			cmd.desc = strings.Trim(vv[idx+1:], " ")
		}
		cds = append(cds, cmd)
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
	cfg.clipKbds = append(all, cfg.clipKbds...)
	cfg.taskKbds = append(all, cfg.taskKbds...)

	cfg.normalKbds = append(readBinding(dd["normal"]), cfg.normalKbds...)
	cfg.jumpKbds = append(readBinding(dd["jump"]), cfg.jumpKbds...)
	cfg.inputKbds = append(readBinding(dd["input"]), cfg.inputKbds...)
	cfg.clipKbds = append(readBinding(dd["clip"]), cfg.clipKbds...)
	cfg.taskKbds = append(readBinding(dd["task"]), cfg.taskKbds...)
}

func readYaml(ds []byte, cfg *config) {
	var mp map[string]interface{}
	yaml.Unmarshal(ds, &mp)
	vv, has := mp["color"]
	if has {
		readColors(vv, cfg)
	}

	vv, has = mp["binding"]
	if has {
		readBindings(vv, cfg)
	}

	vv, has = mp["shell"]
	if has {
		cfg.shell = vv.(string)
	}

	vv, has = mp["editor"]
	if has {
		cfg.editor = vv.(string)
	}

	vv, has = mp["pager"]
	if has {
		cfg.pager = vv.(string)
	}

	vv, has = mp["single-column-mode"]
	if has && vv == true {
		cfg.singleColumnMode = true
	}
}

func (c *config) cmd(args string) *exec.Cmd {
	return exec.Command(c.shell, "-c", args)
}

func initConfig() *config {
	c := &config{colors: make(map[string]*ui.Color), shell: "", editor: "", pager: ""}
	readYaml(data, c)

	f, err := ioutil.ReadFile(filepath.Join(configDir, "config.yml"))
	if err == nil {
		readYaml(f, c)
	}

	shell := os.Getenv("SHELL")
	if shell != "" {
		c.shell = shell
	}

	return c
}
