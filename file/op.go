package file

import (
	"io"
	"os"
	"path/filepath"

	"github.com/jacokoo/fff/executor"
)

var (
	_ = CommonOp(new(commonOp))
	_ = FileOp(new(fileOp))

	exec = executor.Fixed(1, true)
)

type commonOp struct {
	File
}

func (co *commonOp) Dir() executor.Future {
	return co.Path().Dir().Load()
}

func (co *commonOp) Rename(name string) executor.Future {
	p := co.Path().String()
	return exec.Submit("rename", func(executor.TaskState) (interface{}, error) {
		return nil, os.Rename(p, filepath.Join(filepath.Dir(p), name))
	})
}

func (co *commonOp) Delete() executor.Future {
	p := co.Path().String()
	return exec.Submit("delete", func(executor.TaskState) (interface{}, error) {
		if !co.IsDir() {
			return nil, os.Remove(p)
		}

		return nil, os.RemoveAll(p)
	})
}

type fileOp struct {
	*commonOp
}

func (fo *fileOp) Reader() (io.ReadCloser, error) {
	return os.Open(fo.Path().String())
}

func (fo *fileOp) Writer(flag int) (io.WriteCloser, error) {
	return os.OpenFile(fo.Path().String(), flag, 0644)
}

type dirOp struct {
	*commonOp
}

func (do *dirOp) Read() executor.Future {

}
func (do *dirOp) Write([]File) executor.Future {

}
func (do *dirOp) Move([]File) executor.Future {

}
func (do *dirOp) NewFile(string) executor.Future {

}
func (do *dirOp) NewDir(string) executor.Future {

}
func (do *dirOp) To(string) executor.Future {

}
