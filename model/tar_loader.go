package model

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
)

var (
	_ = FileItem(new(tarFile))
	_ = FileItem(new(tarDir))
	_ = FileOp(new(tarFile))
	_ = DirOp(new(tarDir))
)

func openTar(a archive) (io.Closer, *tar.Reader, error) {
	in, err := a.origin().(FileOp).Reader()
	if err != nil {
		return nil, nil, err
	}
	wrapper := a.config().(*tarWrapper)

	if wrapper.reader != nil {
		wr, err := wrapper.reader(in)
		if err != nil {
			return nil, nil, err
		}
		in = newReadCloser(wr, wr, in)
	}

	return in, tar.NewReader(in), nil
}

func writeTar(a archive) (io.Closer, *tar.Writer, error) {
	wrapper := a.config().(*tarWrapper)
	if wrapper.reader != nil || wrapper.writer != nil {
		return nil, nil, a.error("write to compressed tar is not supported")
	}

	in, err := a.origin().(FileOp).Reader()
	if err != nil {
		return nil, nil, err
	}
	ii, ok := in.(io.ReadSeeker)
	if !ok {
		in.Close()
		return nil, nil, a.error("can not append to tar, is not a seeker")
	}

	num := 1
	buf := make([]byte, 512)
	for {
		ii.Seek(int64(-512*num), io.SeekEnd)
		n, _ := ii.Read(buf)
		if n < 512 {
			num--
			break
		}
		count := 0
		for _, v := range buf[:n] {
			count += int(v)
		}
		if count > 0 {
			num--
			break
		}
		num++
	}
	in.Close()

	out, err := a.origin().(FileOp).Writer(os.O_WRONLY)
	if err != nil {
		return nil, nil, err
	}
	o, ok := out.(io.Seeker)
	if !ok {
		out.Close()
		return nil, nil, a.error("can not append to tar, is not a seeker")
	}

	_, err = o.Seek(int64(-512*num), io.SeekEnd)
	if err != nil {
		out.Close()
		return nil, nil, err
	}

	return out, tar.NewWriter(out), nil
}

type tarWrapper struct {
	reader func(io.Reader) (io.ReadCloser, error)
	writer func(io.Writer) (io.WriteCloser, error)
}

type tarFile struct {
	*archiveFileOp
}

func newTarFile(ai archiveItem) *tarFile {
	return &tarFile{&archiveFileOp{&archiveOp{ai}}}
}

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

	c.Close()
	return nil, tf.archive().error("file not found")
}

type tarDir struct {
	*archiveDirOp
}

func newTarDir(ai archiveItem) *tarDir {
	return &tarDir{&archiveDirOp{&archiveOp{ai}}}
}

func (td *tarDir) write(w *tar.Writer, root string, item FileItem) ([]Task, error) {
	name, _ := filepath.Rel(root, item.Path())
	name = filepath.Join(td.ipath(), name)
	for _, v := range td.archive().items() {
		if name == v.ipath() {
			return nil, nil
		}
	}

	if item.IsDir() {
		return td.writeDir(w, root, item)
	}

	return []Task{NewTask(item.Name(), func(progress chan<- int, quit <-chan bool, eh chan<- error) {
		defer close(progress)
		defer close(eh)

		r, err := item.(FileOp).Reader()
		if err != nil {
			eh <- err
			return
		}
		defer r.Close()

		h := &tar.Header{
			Name:    name,
			Mode:    int64(item.Mode()),
			ModTime: item.ModTime(),
			Size:    item.Size(),
		}

		err = w.WriteHeader(h)
		if err != nil {
			eh <- err
			return
		}

		quited := false
		go func() {
			<-quit
			quited = true
		}()

		buf := make([]byte, 4096)
		pg := 0
		si := float64(item.Size())
		var count int64

		for !quited {
			n, err := r.Read(buf)
			if n > 0 {
				_, err2 := w.Write(buf[:n])
				if err2 != nil {
					eh <- err2
					return
				}

				count += int64(n)
				pp := int(float64(count) / si * 100)
				if pp > pg {
					pg = pp
					progress <- pg
				}
			}
			if err == io.EOF {
				break
			}
			if err != nil {
				eh <- err
				break
			}
		}
	})}, nil
}

func (td *tarDir) writeDir(w *tar.Writer, root string, item FileItem) ([]Task, error) {
	its, err := item.(DirOp).Read()
	if err != nil {
		return nil, err
	}

	ts := make([]Task, 0)
	for _, v := range its {
		s, err := td.write(w, root, v)
		if err != nil {
			return ts, err
		}
		ts = append(ts, s...)
	}
	return ts, nil
}

func (td *tarDir) Write(items []FileItem) (Task, error) {
	c, w, err := writeTar(td.archive())
	if err != nil {
		return nil, err
	}

	ts := make([]Task, 0)
	for _, v := range items {
		s, err := td.write(w, filepath.Dir(v.Path()), v)
		if err != nil {
			return nil, err
		}
		ts = append(ts, s...)
	}

	bt := NewBatchTask("Copy", ts)
	bt.Attach(NewListener(nil, func() {
		w.Close()
		c.Close()
	}))

	return bt, nil
}

type tarLoader struct{}

func (*tarLoader) Name() string      { return "tar" }
func (*tarLoader) Seperator() string { return "/" }
func (*tarLoader) Support(item FileItem) bool {
	return !item.IsDir() && strings.HasSuffix(item.Name(), ".tar")
}

func newTarArchive(ld Loader, wrapper *tarWrapper, item FileItem) (archive, error) {
	ta := &defaultArchive{ld, item, nil, nil, wrapper}
	file, reader, err := openTar(ta)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	ta.ro = newTarDir(ta.createRoot())

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
		dai := ta.create(h.FileInfo(), name)

		if dai.IsDir() {
			items = append(items, newTarDir(dai))
		} else {
			items = append(items, newTarFile(dai))
		}
	}
	ta.its = items
	checkMissedDir(ta, func(it *defaultArchiveItem) archiveItem {
		return newTarDir(it)
	})
	return ta, nil
}

func (tl *tarLoader) Create(item FileItem) (FileItem, error) {
	ta, err := newTarArchive(tl, new(tarWrapper), item)
	if err != nil {
		return nil, err
	}
	return ta.root(), nil
}

type tgzLoader struct{}

func (*tgzLoader) Name() string      { return "tgz" }
func (*tgzLoader) Seperator() string { return "/" }
func (*tgzLoader) Support(item FileItem) bool {
	if item.IsDir() {
		return false
	}
	name := item.Name()
	return strings.HasSuffix(name, ".tgz") || strings.HasSuffix(name, ".tar.gz")
}

func (tl *tgzLoader) Create(item FileItem) (FileItem, error) {
	ta, err := newTarArchive(tl, &tarWrapper{func(reader io.Reader) (io.ReadCloser, error) {
		return gzip.NewReader(reader)
	}, func(writer io.Writer) (io.WriteCloser, error) {
		return gzip.NewWriter(writer), nil
	}}, item)
	if err != nil {
		return nil, err
	}
	return ta.root(), nil
}
