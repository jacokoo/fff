package file

import (
	"os"
	"path/filepath"
)

var (
	_ = File(new(fileItem))
	_ = Link(new(fileLink))
)

func init() {
	r, _ := os.Stat("/")
	Root = newFile("", r)
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
	path Path
	link *fileLink
	os.FileInfo
}

func (f *fileItem) Path() Path {
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

func (f *fileItem) Group() string {
	return gid2name(f.FileInfo.Sys())
}

func (f *fileItem) Owner() string {
	return uid2name(f.FileInfo.Sys())
}

// path is the parent dir of info
func newFile(path string, info os.FileInfo) File {
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
	return &fileItem{parsePath(p), link, info}
}
