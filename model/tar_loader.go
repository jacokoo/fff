package model

import (
	"archive/tar"
	"compress/gzip"
	"errors"
	"io"
	"path"
	"strings"
)

func openTar(a archive) (io.Closer, *tar.Reader, error) {
	in, err := a.origin().(FileOp).Reader()
	if err != nil {
		return nil, nil, err
	}
	wrapper := a.config().(func(io.Reader) (io.ReadCloser, error))

	if wrapper != nil {
		wr, err := wrapper(in)
		if err != nil {
			return nil, nil, err
		}
		in = newReadCloser(wr, wr, in)
	}

	return in, tar.NewReader(in), nil
}

type tarItem struct {
	archiveItem
}

// Delete may support by write all other files to a tmp file then rename it back to override current zip file
func (ti *tarItem) Delete() error       { return errors.New("tar: delete is not supported") }
func (ti *tarItem) Rename(string) error { return errors.New("tar: rename is not supported") }
func (ti *tarItem) Open() error         { return ti.archive().origin().(Op).Open() }

type tarFile struct {
	*tarItem
}

func (tf *tarFile) IsDir() bool { return false }
func (tf *tarFile) Reader() (io.ReadCloser, error) {
	c, re, err := openTar(tf.archive())
	if err != nil {
		return nil, err
	}

	for {
		h, err := re.Next()
		if err == io.EOF {
			c.Close()
			break
		}
		if err != nil {
			c.Close()
			return nil, err
		}

		if h.Name == tf.ipath() {
			return newReadCloser(re, c), nil
		}
	}

	return nil, errors.New("tar: file not found")
}

type tarDir struct {
	*tarItem
}

func (*tarDir) IsDir() bool                      { return true }
func (*tarDir) NewFile(string) error             { return errors.New("tar: new file is not supported") }
func (*tarDir) NewDir(string) error              { return errors.New("tar: new dir is not supported") }
func (*tarDir) Move([]FileItem) error            { return errors.New("tar: move is not supported") }
func (td *tarDir) To(p string) (FileItem, error) { return archiveTo(td, p) }
func (td *tarDir) Read() ([]FileItem, error)     { return archiveChildren(td), nil }

func (td *tarDir) Write([]FileItem) ([]Task, error) {
	return nil, nil
}

type tarLoader struct{}

func (*tarLoader) Name() string { return "tar" }
func (*tarLoader) Support(item FileItem) bool {
	return !item.IsDir() && strings.HasSuffix(item.Name(), ".tar")
}

func newTarArchive(prefix string, wrapper func(io.Reader) (io.ReadCloser, error), item FileItem) (archive, error) {
	ta := &defaultArchive{prefix, item, nil, nil, wrapper}
	file, reader, err := openTar(ta)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	ta.ro = &tarDir{&tarItem{ta.createRoot(prefix)}}

	items := make([]archiveItem, 0)
	for {
		h, err := reader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		name := path.Clean(h.Name)
		dai := ta.create(h.FileInfo(), prefix, name)
		ai := &tarItem{dai}

		if ai.IsDir() {
			items = append(items, &tarDir{ai})
		} else {
			items = append(items, &tarFile{ai})
		}
	}
	ta.its = items
	checkMissedDir(ta, func(it *defaultArchiveItem) archiveItem {
		return &tarDir{&tarItem{it}}
	})
	return ta, nil
}

func (*tarLoader) Create(item FileItem) (FileItem, error) {
	ta, err := newTarArchive("@tar://", nil, item)
	if err != nil {
		return nil, err
	}
	return ta.root(), nil
}

type tgzLoader struct{}

func (*tgzLoader) Name() string { return "tgz" }
func (*tgzLoader) Support(item FileItem) bool {
	if item.IsDir() {
		return false
	}
	name := item.Name()
	return strings.HasSuffix(name, ".tgz") || strings.HasSuffix(name, ".tar.gz")
}

func (*tgzLoader) Create(item FileItem) (FileItem, error) {
	ta, err := newTarArchive("@tgz://", func(reader io.Reader) (io.ReadCloser, error) {
		return gzip.NewReader(reader)
	}, item)
	if err != nil {
		return nil, err
	}
	return ta.root(), nil
}
