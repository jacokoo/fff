package model

import (
	"errors"
	"io"
	"os"
	"path"
	"strings"
)

type archive interface {
	origin() FileItem
	root() FileItem
	items() []archiveItem
}

type archiveItem interface {
	archive() archive
	ipath() string
	depth() int
	FileItem
}

type defaultArchive struct {
	og  FileItem
	ro  archiveItem
	its []archiveItem
}

func (da *defaultArchive) origin() FileItem     { return da.og }
func (da *defaultArchive) root() FileItem       { return da.ro }
func (da *defaultArchive) items() []archiveItem { return da.its }

type defaultArchiveItem struct {
	ar archive
	p  string
	d  int
	FileItem
}

type archiveFile struct {
	*defaultArchiveItem
}

type archiveDir struct {
	*defaultArchiveItem
}

func (da *defaultArchiveItem) archive() archive { return da.ar }
func (da *defaultArchiveItem) ipath() string    { return da.p }
func (da *defaultArchiveItem) depth() int       { return da.d }

func archiveTo(from archiveItem, to string) (FileItem, error) {
	p := path.Clean(to)
	p = path.Join(from.ipath(), p)
	for _, v := range from.archive().items() {
		switch tt := v.(type) {
		case *zipfile:
			if tt.ipath() == p {
				return v, nil
			}
		case *zipdir:
			if tt.ipath() == p {
				return v, nil
			}
		}
	}
	return nil, errors.New("not found")
}

func (da *defaultArchive) create(fi os.FileInfo, prefix, ipath string) *defaultArchiveItem {
	ipath = path.Clean(ipath)
	p := da.og.Path() + prefix + "/" + ipath
	ffi := &fileItem{p, nil, fi}
	return &defaultArchiveItem{da, ipath, len(strings.Split(ipath, "/")), ffi}
}

type archiveRootItem struct {
	prefix string
	FileItem
}

func (zr *archiveRootItem) Path() string { return zr.FileItem.Path() + zr.prefix + "/" }

func (da *defaultArchive) createRoot(prefix string) *defaultArchiveItem {
	return &defaultArchiveItem{da, "", 0, &archiveRootItem{prefix, da.og}}
}

type closer2 struct {
	io.ReadCloser
	other io.Closer
}

func (zf *closer2) Close() error {
	err1 := zf.ReadCloser.Close()
	err2 := zf.other.Close()
	if err1 != nil {
		return err1
	}
	if err2 != nil {
		return err2
	}
	return nil
}
