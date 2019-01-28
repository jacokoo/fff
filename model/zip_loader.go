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

var (
	_ = FileItem(new(zipfile))
	_ = FileItem(new(zipdir))
	_ = FileOp(new(zipfile))
	_ = DirOp(new(zipdir))
)

func openZip(a archive) (io.Closer, *zip.Reader, error) {
	in, err := a.origin().(FileOp).Reader()
	if err != nil {
		return nil, nil, err
	}
	ra, ok := in.(io.ReaderAt)
	if !ok {
		return nil, nil, errors.New(zipstring + ": can not open zip file, ReaderAt is required")
	}
	r, err := zip.NewReader(ra, a.origin().Size())
	if err != nil {
		return nil, nil, err
	}
	return in, r, nil
}

type zipfile struct {
	*archiveFileOp
}

func newZipfile(ai archiveItem) *zipfile {
	return &zipfile{&archiveFileOp{&archiveOp{ai}}}
}

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
	*archiveDirOp
}

func newZipdir(ai archiveItem) *zipdir {
	return &zipdir{&archiveDirOp{&archiveOp{ai}}}
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

	ar.ro = newZipdir(ar.createRoot(zipstring))

	items := make([]archiveItem, 0)
	for _, v := range reader.File {
		p := path.Clean(v.Name)
		ii := ar.create(v.FileInfo(), zipstring, p)
		if ii.IsDir() {
			items = append(items, newZipdir(ii))
		} else {
			items = append(items, newZipfile(ii))
		}
	}
	ar.its = items
	checkMissedDir(ar, func(it *defaultArchiveItem) archiveItem { return newZipdir(it) })
	return ar.ro, nil
}
