package main

var data = []byte(`
bindings:
  all:
    ctrl-q: ActionQuit					# quit fff
  normal:
    s:                                  # Prefix, Sort File
      - n: ActionSortByName             # Sort By Name
      - m: ActionSortByMtime            # Sort By MTime
      - s: ActionSortBySize             # Sort By Size
    .: ActionToggleHidden               # Toggle show hidden files
    j: ActionMoveDown                   # Move down
    k: ActionMoveUp                     # Move up
    l: ActionOpenFolderRight            # Open folder on right
    h: ActionCloseFolderRight           # Go to parent folder
    ",": ActionShift                    # Shift column
    <: ActionMoveToFirst                # Move to first item
    ">": ActionMoveToLast               # Move to last item
    ctrl-n: ActionMoveDown              # Move down
    ctrl-p: ActionMoveUp                # Move up
    enter: ActionOpenFolderRight        # Open folder on right
    b:                                  # Prefix, Bookmark manage
      - b: ActionToggleBookmark         # Toggle show bookmark
    w: ActionEnterJump                  # Enter jump mode
    g: ActionRefresh                    # Refresh current dir
    1: ActionChangeGroup0               # Change group to 1
    2: ActionChangeGroup1               # Change group to 2
    3: ActionChangeGroup2               # Change group to 3
    4: ActionChangeGroup3               # Change group to 4

  jump:
    enter: "ActionQuitJump"
    esc: "ActionQuitJump"

color:
  normal: default
  keyword: cyan
  selected: cyan
  jump: red
`)
