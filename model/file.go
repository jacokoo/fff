package model

import (
	"os"
	"path/filepath"
	"time"
)

// FileItem represent a file, with absolute path
type FileItem interface {
	Path() string
	Name() string
	Size() int64
	ModTime() time.Time
	Mode() os.FileMode
	IsDir() bool
	Link() (Link, bool)
}

// Link symbolic link
type Link interface {
	IsBroken() bool
	Target() string
	IsDir() bool
}

// FileLink symbolic link
type FileLink struct {
	broken bool
	target string
	isDir  bool
}

// IsBroken is the link broken
func (fl *FileLink) IsBroken() bool {
	return fl.broken
}

// IsDir is the link link to a dir
func (fl *FileLink) IsDir() bool {
	return fl.isDir
}

// Target the target file path
func (fl *FileLink) Target() string {
	return fl.target
}

// File is a file item
type File struct {
	path string
	link *FileLink
	os.FileInfo
}

// Path the absolute path
func (f *File) Path() string {
	return f.path
}

// Link return the file link
func (f *File) Link() (Link, bool) {
	if f.link == nil {
		return nil, false
	}

	return f.link, true
}

// IsDir if the file links to a dir return true
func (f *File) IsDir() bool {
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
func NewFile(path string, info os.FileInfo) FileItem {
	var link *FileLink
	p := filepath.Join(path, info.Name())
	if info.Mode()&os.ModeSymlink != 0 {
		link = new(FileLink)
		st, err := os.Stat(p)
		link.broken = err != nil
		link.isDir = err == nil && st.IsDir()
		tr, _ := os.Readlink(p)
		link.target = tr
	}
	return &File{p, link, info}
}

// NewFiles create a file list
func NewFiles(path string, infos []os.FileInfo) []FileItem {
	fis := make([]FileItem, len(infos))
	for i, v := range infos {
		fis[i] = NewFile(path, v)
	}
	return fis
}
