package model

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

var (
	_ = FileOp(new(file))
	_ = DirOp(new(dir))

	shell  string
	pager  string
	editor string
)

// SetDefault set default
func SetDefault(sh, pg, ed string) {
	shell = sh
	pager = pg
	editor = ed
}

func newCmd(args string) *exec.Cmd {
	cm := exec.Command(shell, "-c", args)
	cm.Stdin = os.Stdin
	cm.Stdout = os.Stdout
	cm.Stderr = os.Stderr
	return cm
}

// Op common operators for both file and dir
type Op interface {
	Dir() (FileItem, error)
	Rename(string) error
	Delete() error
	Open() error
}

// FileOp file operators
type FileOp interface {
	Reader() (io.ReadCloser, error)
	Writer(int) (io.WriteCloser, error)
	View() error
	Edit() error
	Op
}

// DirOp dir operators
type DirOp interface {
	Read() ([]FileItem, error)
	Write([]FileItem) (Task, error)
	Move([]FileItem) error
	NewFile(string) error
	NewDir(string) error
	To(string) (FileItem, error)
	Shell() error
	Close() error
	Op
}

type defaultOp struct {
	FileItem
}

func (do *defaultOp) Dir() (FileItem, error) {
	return Load(filepath.Dir(do.Path()))
}

func (do *defaultOp) Rename(name string) error {
	path := do.Path()
	return os.Rename(path, filepath.Join(filepath.Dir(path), name))
}

func (do *defaultOp) Delete() error {
	if !do.IsDir() {
		return os.Remove(do.Path())
	}

	return os.RemoveAll(do.Path())
}

func (do *defaultOp) Open() error {
	open := ""
	switch runtime.GOOS {
	case "darwin":
		open = "open"
	case "linux", "freebsd":
		open = "xdg-open"
	default:
		return fmt.Errorf("not supported")
	}

	return newCmd(fmt.Sprintf("%s %s", open, do.Path())).Start()
}

type defaultFileOp struct {
	*defaultOp
}

func (df *defaultFileOp) Reader() (io.ReadCloser, error) {
	return os.Open(df.Path())
}

func (df *defaultFileOp) Writer(flag int) (io.WriteCloser, error) {
	return os.OpenFile(df.Path(), flag, 0644)
}

func (df *defaultFileOp) Edit() error {
	return newCmd(fmt.Sprintf("%s %s", editor, df.Path())).Run()
}

func (df *defaultFileOp) View() error {
	return newCmd(fmt.Sprintf("%s %s", pager, df.Path())).Run()
}

type defaultDirOp struct {
	*defaultOp
}

type file struct {
	*fileItem
	*defaultFileOp
}

type dir struct {
	*fileItem
	*defaultDirOp
}

func (dd *dir) To(sub string) (FileItem, error) {
	if filepath.Dir(sub) == sub {
		return dd, nil
	}
	return dd.defaultDirOp.To(sub)
}

func (dd *defaultDirOp) To(sub string) (FileItem, error) {
	p := filepath.Join(dd.Path(), sub)
	stat, err := os.Stat(p)
	if err != nil {
		return nil, err
	}
	v := newFile(filepath.Dir(p), stat)
	if v.IsDir() {
		return &dir{v, &defaultDirOp{&defaultOp{v}}}, nil
	}
	return &file{v, &defaultFileOp{&defaultOp{v}}}, nil
}

func (dd *defaultDirOp) Read() ([]FileItem, error) {
	fis, err := ioutil.ReadDir(dd.Path())
	if err != nil {
		return nil, err
	}

	items := newFiles(dd.Path(), fis)
	rs := make([]FileItem, len(items))

	for i, v := range items {
		if v.IsDir() {
			rs[i] = &dir{v, &defaultDirOp{&defaultOp{v}}}
		} else {
			rs[i] = &file{v, &defaultFileOp{&defaultOp{v}}}
		}
	}
	return rs, nil
}

func (dd *defaultDirOp) NewFile(name string) error {
	_, err := os.Create(filepath.Join(dd.Path(), name))
	return err
}

func (dd *defaultDirOp) NewDir(name string) error {
	return os.MkdirAll(filepath.Join(dd.Path(), name), 0755)
}

func (dd *defaultDirOp) Move(items []FileItem) error {
	parent := filepath.Dir(dd.Path())
	for _, v := range items {
		err := os.Rename(v.Path(), filepath.Join(parent, v.Name()))
		if err != nil {
			return err
		}
	}
	return nil
}

func (dd *defaultDirOp) write(root string, item FileItem) ([]Task, error) {
	if item.IsDir() {
		return dd.writeDir(root, item)
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

		rel, err := filepath.Rel(root, item.Path())
		if err != nil {
			eh <- err
			return
		}
		path := filepath.Join(dd.Path(), rel)
		err = os.MkdirAll(filepath.Dir(path), 0755)
		if err != nil {
			eh <- err
			return
		}

		w, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0644)
		if err != nil {
			eh <- err
			return
		}

		buf := make([]byte, 4096)
		pg := 0
		si := float64(item.Size())

		var count int64
		var quited = false

		go func() {
			<-quit
			quited = true
		}()

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
				return
			}
		}
	})}, nil
}

func (dd *defaultDirOp) writeDir(root string, item FileItem) ([]Task, error) {
	its, err := item.(DirOp).Read()
	if err != nil {
		return nil, err
	}

	re := make([]Task, 0)
	for _, v := range its {
		ts, err := dd.write(root, v)
		if err != nil {
			return re, err
		}
		re = append(re, ts...)
	}
	return re, nil
}

func (dd *defaultDirOp) Write(items []FileItem) (Task, error) {
	re := make([]Task, 0)
	for _, v := range items {
		ts, err := dd.write(filepath.Dir(v.Path()), v)
		if err != nil {
			return nil, err
		}
		re = append(re, ts...)
	}
	return NewBatchTask("Copy", re), nil
}

func (dd *defaultDirOp) Close() error {
	return nil
}

func (dd *defaultDirOp) Shell() error {
	os.Chdir(dd.Path())
	cm := exec.Command(shell)
	cm.Stderr = os.Stderr
	cm.Stdin = os.Stdin
	cm.Stdout = os.Stdout
	return cm.Run()
}
