package main

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
)

const (
	homeName = "Home[~]"
	rootName = "Root[/]"
)

var (
	bookmarks            = make(map[string]string)
	bookmarkKeys         = make([]string, 0)
	maxBookmarkNameWidth = len(homeName)
	bookmarkFileName     string
)

func readFromFile() {
	bs, err := ioutil.ReadFile(bookmarkFileName)
	if err != nil {
		return
	}

	for _, v := range strings.Split(string(bs), "\n") {
		ts := strings.Split(v, "=")
		if len(ts) != 2 {
			continue
		}
		name, value := strings.Trim(ts[0], " \t"), strings.Trim(ts[1], " \t")
		if strings.HasPrefix(value, "~/") {
			value = filepath.Join(home, value[2:])
		}
		bookmarks[name] = value
		if len(name) > maxBookmarkNameWidth {
			maxBookmarkNameWidth = len(name)
		}
		bookmarkKeys = append(bookmarkKeys, name)
	}
}

func writeToFile() {
	s := ""
	for _, k := range bookmarkKeys {
		if k == homeName || k == rootName {
			continue
		}

		s = fmt.Sprintf("%s%s=%s\n", s, k, bookmarks[k])
	}
	ioutil.WriteFile(bookmarkFileName, []byte(s), 0644)
}

func initBookmark() {
	bookmarkFileName = filepath.Join(configDir, "bookmarks")
	bookmarks[homeName] = home
	bookmarks[rootName] = "/"
	bookmarkKeys = append(bookmarkKeys, homeName)
	bookmarkKeys = append(bookmarkKeys, rootName)
	readFromFile()
}

func addBookmark(name, path string) {
	_, has := bookmarks[name]
	if has {
		message = fmt.Sprintf("Bookmark %s already exists", name)
		gui <- uiErrorMessage
		return
	}
	bookmarks[name] = path
	if len(name) > maxBookmarkNameWidth {
		maxBookmarkNameWidth = len(name)
	}
	bookmarkKeys = append(bookmarkKeys, name)
	writeToFile()
	gui <- uiBookmarkChanged
}
