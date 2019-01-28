package model

import (
	"os"
	"path/filepath"
	"time"
)

var _ = FileItem(new(fileItem))

// FileItem represent a file, with absolute path
type FileItem interface {
	Path() string
	Name() string
	Size() int64
	ModTime() time.Time
	Mode() os.FileMode
	IsDir() bool
	Link() (Link, bool)
	Sys() interface{}
}

// Link symbolic link
type Link interface {
	IsBroken() bool
	Target() string
	IsDir() bool
}

type fileLink struct {
	broken bool
	target string
	isDir  bool
}

func (fl *fileLink) IsBroken() bool {
	return fl.broken
}

func (fl *fileLink) IsDir() bool {
	return fl.isDir
}

func (fl *fileLink) Target() string {
	return fl.target
}

type fileItem struct {
	path string
	link *fileLink
	os.FileInfo
}

func (f *fileItem) Path() string {
	return f.path
}

func (f *fileItem) Link() (Link, bool) {
	if f.link == nil {
		return nil, false
	}

	return f.link, true
}

func (f *fileItem) IsDir() bool {
	if f.FileInfo.IsDir() {
		return true
	}

	if f.link == nil {
		return false
	}

	return f.link.IsDir()
}

// NewFile create file item
// path is the parent dir of info
func newFile(path string, info os.FileInfo) *fileItem {
	var link *fileLink
	p := filepath.Join(path, info.Name())
	if info.Mode()&os.ModeSymlink != 0 {
		link = new(fileLink)
		st, err := os.Stat(p)
		link.broken = err != nil
		link.isDir = err == nil && st.IsDir()
		tr, _ := os.Readlink(p)
		link.target = tr
	}
	return &fileItem{p, link, info}
}

// NewFiles create a file list
func newFiles(path string, infos []os.FileInfo) []*fileItem {
	fis := make([]*fileItem, len(infos))
	for i, v := range infos {
		fis[i] = newFile(path, v)
	}
	return fis
}
