package ui

import (
	"strings"
)

const helpText = `Website: https://github.com/jacokoo/fff

Navigation:
      ↓, j    select next file                  ↑, k    select previous file
         J    select the last file                 K    select the first file
	  →, l    open selected dir                 ←, h    close current dir
		 f    filter files in current dir          F    clear filter
		ss    sort current dir by size            sm    sort current dir by modify time
		sn    sort current dir by name             g    refresh current dir
		 d    toggle show file details             .    toggle show hidden files
         ,    remove the first opened dir
		 w    jump over all items displayed once   W    jump over all items displayed
		 i    jump over the current dir once       I    jump over the current dir
1, 2, 3, 4    switch to corresponding context
		 ↵    open selected item use system default program
			  ensure input (during input), cancel jump (during jump)
	   esc    abort input (during input), cancel jump (during jump)

File:
         m    toggle mark file                     u    toggle mark all items
		 +    create new dir                       N    create new file
		 R    rename selected file                 D    delete selected/marked items
		 U    clear clips                          C    append selected/marked items to clip
		 P    paste all cliped items to current dir
		 M    move all cliped items to current dir

Bookmark:
	    bb    toggle show bookmark                bn    create bookmark
		bd    delete bookmark
		bw    jump over bookmark once             bW    jump over bookmark

Misc:
 q, ctrl-q    Quit fff                             v    open selected file via pager
		 !    start a shell in current dir         e    editor selected file
		 ?    for help

[Press any key to quit]`

// NewHelp create help
func NewHelp(height int) *List {
	ns := strings.Split(strings.Replace(helpText, "\t", "    ", -1), "\n")
	hs := make([]int, len(ns))
	for i := range ns {
		hs[i] = 0
		if i == 0 || i == 2 || i == 18 || i == 26 || i == 31 {
			hs[i] = 1
		} else if i == len(ns)-1 {
			hs[i] = 2
		}
	}

	return NewList(ZeroPoint, -1, height, ns, hs)
}
