package model

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

// Default bookmark names
const (
	HomeName = "Home[~]"
	RootName = "Root[/]"
)

var (
	home = os.Getenv("HOME")
)

// Bookmark manager
type Bookmark struct {
	path     string
	items    map[string]string
	Names    []string
	MaxWidth int
}

func (b *Bookmark) write() {
	dir := filepath.Dir(b.path)
	if _, err := os.Stat(dir); err != nil {
		os.MkdirAll(dir, 0755)
	}
	s := ""
	for _, k := range b.Names {
		if k == HomeName || k == RootName {
			continue
		}

		s = fmt.Sprintf("%s%s=%s\n", s, k, b.items[k])
	}
	ioutil.WriteFile(b.path, []byte(s), 0644)
}

func (b *Bookmark) read() {
	bs, err := ioutil.ReadFile(b.path)
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
		b.items[name] = value
		if len(name) > b.MaxWidth {
			b.MaxWidth = len(name)
		}
		b.Names = append(b.Names, name)
	}
}

func (b *Bookmark) addBookmark(name, value string, write bool) error {
	_, has := b.items[name]
	if has {
		return fmt.Errorf("Bookmark %s already exists", name)
	}
	b.items[name] = value
	if len(name) > b.MaxWidth {
		b.MaxWidth = len(name)
	}
	b.Names = append(b.Names, name)
	if write {
		b.write()
	}
	return nil
}

// Add a new bookmark
func (b *Bookmark) Add(name, value string) error {
	return b.addBookmark(name, value, true)
}

// Delete a bookmark
func (b *Bookmark) Delete(name string) error {
	if name == HomeName || name == RootName {
		return fmt.Errorf("Can not delete %s, %s", HomeName, RootName)
	}

	_, has := b.items[name]
	if !has {
		return fmt.Errorf("Bookmark %s is not exists", name)
	}

	delete(b.items, name)
	bks := make([]string, 0)
	for _, v := range b.Names {
		if v == name {
			continue
		}
		bks = append(bks, v)
	}
	b.Names = bks
	b.write()
	return nil
}

// Get value by name
func (b *Bookmark) Get(name string) (string, bool) {
	v, has := b.items[name]
	if !has {
		return "", false
	}

	return v, true
}

// IsFixed is the name fixed
func (b *Bookmark) IsFixed(name string) bool {
	return name == HomeName || name == RootName
}

// NewBookmark create bookmark
func NewBookmark(path string) *Bookmark {
	b := &Bookmark{path, make(map[string]string), nil, 0}
	b.addBookmark(HomeName, home, false)
	b.addBookmark(RootName, "/", false)
	b.read()
	return b
}
