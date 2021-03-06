package model

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"strings"
	"time"
)

type archive interface {
	loader() Loader
	origin() FileItem
	root() FileItem
	items() []archiveItem
	config() interface{}
	error(string) error
}

type archiveItem interface {
	archive() archive
	ipath() string
	depth() int
	FileItem
}

type defaultArchive struct {
	ld  Loader
	og  FileItem
	ro  archiveItem
	its []archiveItem
	cfg interface{}
}

func (da *defaultArchive) loader() Loader         { return da.ld }
func (da *defaultArchive) origin() FileItem       { return da.og }
func (da *defaultArchive) root() FileItem         { return da.ro }
func (da *defaultArchive) items() []archiveItem   { return da.its }
func (da *defaultArchive) config() interface{}    { return da.cfg }
func (da *defaultArchive) error(msg string) error { return fmt.Errorf("%s: %s", da.ld.Name(), msg) }

type defaultArchiveItem struct {
	ar archive
	p  string
	d  int
	FileItem
}

func (da *defaultArchiveItem) archive() archive { return da.ar }
func (da *defaultArchiveItem) ipath() string    { return da.p }
func (da *defaultArchiveItem) depth() int       { return da.d }

func archiveTo(from archiveItem, to string) (FileItem, error) {
	p := path.Clean(to)
	p = path.Join(from.ipath(), p)
	for _, v := range from.archive().items() {
		switch tt := v.(type) {
		case archiveItem:
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

func (da *defaultArchive) create(fi os.FileInfo, ipath string) *defaultArchiveItem {
	ipath = path.Clean(ipath)
	p := da.og.Path() + LoaderString(da.ld) + "/" + ipath
	ffi := &fileItem{p, nil, fi}
	return &defaultArchiveItem{da, ipath, len(strings.Split(ipath, "/")), ffi}
}

func (da *defaultArchive) createRoot() *defaultArchiveItem {
	return &defaultArchiveItem{da, "", 0, &archiveRootItem{da, da.og}}
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

func (da *defaultArchive) createMissedDir(ipath string) *defaultArchiveItem {
	m := &missedDir{path.Base(ipath), 0755, time.Now(), 0}
	return da.create(m, ipath)
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
	for k := range has {
		v := has[k]
		if v {
			continue
		}
		for !has[k] {
			md := ar.createMissedDir(k)
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
	ar archive
	FileItem
}

func (zr *archiveRootItem) Path() string {
	return zr.FileItem.Path() + LoaderString(zr.ar.loader()) + "/"
}

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

type writeCloserN struct {
	io.Writer
	closers []io.Closer
}

func (rc *writeCloserN) Close() error {
	var ers []string
	for _, v := range rc.closers {
		if err := v.Close(); err != nil {
			ers = append(ers, err.Error())
		}
	}
	return errors.New(strings.Join(ers, "\n"))
}

func newWriteCloser(writer io.Writer, closers ...io.Closer) io.WriteCloser {
	return &writeCloserN{writer, closers}
}

type archiveOp struct {
	archiveItem
}

func (ti *archiveOp) Dir() (FileItem, error) {
	ip := ti.archiveItem.ipath()
	pp := path.Dir(ip)

	if ip == "" && pp == "." {
		return ti.archive().origin().(Op).Dir()
	}
	if ip == "" || pp == "." {
		return ti.archive().root(), nil
	}
	return ti.archive().root().(DirOp).To(pp)
}

func (ti *archiveOp) Open() error         { return ti.archive().origin().(Op).Open() }
func (ti *archiveOp) Delete() error       { return ti.archive().error("delete is not supported") }
func (ti *archiveOp) Rename(string) error { return ti.archive().error("rename is not supported") }

type archiveFileOp struct {
	*archiveOp
}

func (*archiveFileOp) IsDir() bool { return false }
func (ti *archiveFileOp) Writer(int) (io.WriteCloser, error) {
	return nil, ti.archive().error("writer is not supported")
}
func (ti *archiveFileOp) Edit() error { return ti.archive().error("edit is not supported") }
func (ti *archiveFileOp) View() error { return ti.archive().error("view is not supported") }

type archiveDirOp struct {
	*archiveOp
}

func (*archiveDirOp) IsDir() bool                      { return true }
func (td *archiveDirOp) To(p string) (FileItem, error) { return archiveTo(td, p) }
func (td *archiveDirOp) Read() ([]FileItem, error)     { return archiveChildren(td), nil }
func (td *archiveDirOp) NewFile(string) error          { return td.archive().error("new file is not supported") }
func (td *archiveDirOp) NewDir(string) error           { return td.archive().error("new dir is not supported") }
func (td *archiveDirOp) Move([]FileItem) error         { return td.archive().error("move is not supported") }

func (td *archiveDirOp) Write([]FileItem) (Task, error) {
	return nil, td.archive().error("write to dir is not supported")
}

func (td *archiveDirOp) Shell() error {
	op, err := td.archive().origin().(Op).Dir()
	if err != nil {
		return err
	}

	return op.(DirOp).Shell()
}
