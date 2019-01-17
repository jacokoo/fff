package model

import (
	"errors"
	"io"
	"os"
	"path"
	"strings"
	"time"
)

type archive interface {
	prefix() string
	origin() FileItem
	root() FileItem
	items() []archiveItem
	config() interface{}
}

type archiveItem interface {
	archive() archive
	ipath() string
	depth() int
	FileItem
}

type defaultArchive struct {
	pre string
	og  FileItem
	ro  archiveItem
	its []archiveItem
	cfg interface{}
}

func (da *defaultArchive) prefix() string       { return da.pre }
func (da *defaultArchive) origin() FileItem     { return da.og }
func (da *defaultArchive) root() FileItem       { return da.ro }
func (da *defaultArchive) items() []archiveItem { return da.its }
func (da *defaultArchive) config() interface{}  { return da.cfg }

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

func archiveChildren(parent archiveItem) []FileItem {
	fis, depth := make([]FileItem, 0), parent.depth()+1
	for _, v := range parent.archive().items() {
		dep, p := -1, ""
		switch tt := v.(type) {
		case archiveItem:
			dep, p = tt.depth(), tt.ipath()
		}
		if dep == depth && strings.HasPrefix(p, parent.ipath()) {
			fis = append(fis, v)
		}
	}
	return fis
}

func (da *defaultArchive) create(fi os.FileInfo, prefix, ipath string) *defaultArchiveItem {
	ipath = path.Clean(ipath)
	p := da.og.Path() + prefix + "/" + ipath
	ffi := &fileItem{p, nil, fi}
	return &defaultArchiveItem{da, ipath, len(strings.Split(ipath, "/")), ffi}
}

func (da *defaultArchive) createRoot(prefix string) *defaultArchiveItem {
	return &defaultArchiveItem{da, "", 0, &archiveRootItem{prefix, da.og}}
}

type missedDir struct {
	name    string
	mode    os.FileMode
	modTime time.Time
	size    int64
}

func (m *missedDir) Name() string       { return m.name }
func (m *missedDir) Size() int64        { return m.size }
func (m *missedDir) Mode() os.FileMode  { return m.mode }
func (m *missedDir) ModTime() time.Time { return m.modTime }
func (m *missedDir) IsDir() bool        { return true }
func (m *missedDir) Sys() interface{}   { return nil }

func (da *defaultArchive) createMissedDir(prefix, ipath string) *defaultArchiveItem {
	m := &missedDir{path.Base(ipath), 0755, time.Now(), 0}
	return da.create(m, prefix, ipath)
}

func checkMissedDir(ar *defaultArchive, toDir func(*defaultArchiveItem) archiveItem) {
	has := make(map[string]bool)
	for _, v := range ar.its {
		name := v.ipath()
		if v.IsDir() {
			has[name] = true
		} else {
			nn := path.Dir(name)
			if _, ok := has[nn]; !ok {
				has[nn] = false
			}
		}
	}
	for k, v := range has {
		if v {
			continue
		}
		for {
			md := ar.createMissedDir(ar.prefix(), k)
			ar.its = append(ar.its, toDir(md))
			has[k] = true
			if md.depth() == 1 {
				break
			}
			k = path.Dir(k)
		}
	}
}

type archiveRootItem struct {
	prefix string
	FileItem
}

func (zr *archiveRootItem) Path() string { return zr.FileItem.Path() + zr.prefix + "/" }

type readCloserN struct {
	io.Reader
	closers []io.Closer
}

func (rc *readCloserN) Close() error {
	var ers []string
	for _, v := range rc.closers {
		if err := v.Close(); err != nil {
			ers = append(ers, err.Error())
		}
	}
	return errors.New(strings.Join(ers, "\n"))
}

func newReadCloser(reader io.Reader, closers ...io.Closer) io.ReadCloser {
	return &readCloserN{reader, closers}
}
