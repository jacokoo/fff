package model

import (
	"archive/zip"
	"errors"
	"io"
	"path"
	"strings"
)

const (
	zipstring = "@zip://"
)

func openZip(a archive) (io.Closer, *zip.Reader, error) {
	in, err := a.origin().(FileOp).Reader()
	if err != nil {
		return nil, nil, err
	}
	ra, ok := in.(io.ReaderAt)
	if !ok {
		return nil, nil, errors.New("not supported: reader is not a ReaderAt")
	}
	r, err := zip.NewReader(ra, a.origin().Size())
	if err != nil {
		return nil, nil, err
	}
	return in, r, nil
}

type zipItem struct {
	archiveItem
}

// Delete may support by write all other files to a tmp file then rename it back to override current zip file
func (zf *zipItem) Delete() error       { return errors.New("zip: delete is not supported") }
func (zf *zipItem) Rename(string) error { return errors.New("zip: rename is not supported") }
func (zf *zipItem) Open() error         { return zf.archive().origin().(Op).Open() }

type zipfile struct {
	*zipItem
}

func (zf *zipfile) IsDir() bool { return false }
func (zf *zipfile) Reader() (io.ReadCloser, error) {
	file, reader, err := openZip(zf.archive())
	if err != nil {
		return nil, err
	}
	for _, v := range reader.File {
		if v.Name == zf.ipath() {
			rr, err := v.Open()
			if err != nil {
				file.Close()
				return nil, err
			}
			return newReadCloser(rr, rr, file), nil
		}
	}
	file.Close()
	return nil, errors.New("zip: file not found")
}

type zipdir struct {
	*zipItem
}

func (zd *zipdir) IsDir() bool                   { return true }
func (zd *zipdir) NewFile(string) error          { return errors.New("zip: new file is not supported") }
func (zd *zipdir) NewDir(string) error           { return errors.New("zip: new dir is not supported") }
func (zd *zipdir) Move([]FileItem) error         { return errors.New("zip: move is not supported") }
func (zd *zipdir) To(p string) (FileItem, error) { return archiveTo(zd, p) }
func (zd *zipdir) Read() ([]FileItem, error)     { return archiveChildren(zd), nil }

func (zd *zipdir) Write([]FileItem) ([]Task, error) {
	return nil, nil
}

type zipLoader struct {
}

func (zl *zipLoader) Name() string { return "zip" }
func (zl *zipLoader) Support(item FileItem) bool {
	return !item.IsDir() && strings.HasSuffix(item.Name(), ".zip")
}

func (zl *zipLoader) Create(item FileItem) (FileItem, error) {
	ar := &defaultArchive{zipstring, item, nil, nil, nil}
	file, reader, err := openZip(ar)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	ar.ro = &zipdir{&zipItem{ar.createRoot(zipstring)}}

	items := make([]archiveItem, 0)
	for _, v := range reader.File {
		p := path.Clean(v.Name)
		ii := &zipItem{ar.create(v.FileInfo(), zipstring, p)}
		if ii.IsDir() {
			items = append(items, &zipdir{ii})
		} else {
			items = append(items, &zipfile{ii})
		}
	}
	ar.its = items
	checkMissedDir(ar, func(it *defaultArchiveItem) archiveItem {
		return &zipdir{&zipItem{it}}
	})
	return ar.ro, nil
}
