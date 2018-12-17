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
}

// File is a file item
type File struct {
	path string
	os.FileInfo
}

// Path the absolute path
func (f *File) Path() string {
	return f.path
}

// NewFile create file item
// path is the parent dir of info
func NewFile(path string, info os.FileInfo) FileItem {
	return &File{filepath.Join(path, info.Name()), info}
}

// NewFiles create a file list
func NewFiles(path string, infos []os.FileInfo) []FileItem {
	fis := make([]FileItem, len(infos))
	for i, v := range infos {
		fis[i] = NewFile(path, v)
	}
	return fis
}
