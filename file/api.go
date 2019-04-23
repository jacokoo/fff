package file

import (
	"io"
	"os"
	"time"

	. "github.com/jacokoo/fff/executor"
)

var (
	RequestCh  = make(chan *Request)
	ResponseCh = make(chan string)

	Root File
)

type Path interface {
	String() string
	Load() Future // File
	Dir() Path    // File
}

type Loader interface {
	Name() string
	Support(File) bool
	Create(File) Future // File
	PathDir(string) string
}

type Request struct {
	Title      string
	IsPassword bool
}

type File interface {
	Path() Path
	Name() string
	Size() int64
	ModTime() time.Time
	Mode() os.FileMode
	IsDir() bool
	Owner() string
	Group() string
	Link() (Link, bool)
	Sys() interface{}
}

// symbolic link
type Link interface {
	IsBroken() bool
	Target() string
	IsDir() bool
}

type CommonOp interface {
	Dir() Future          // File
	Rename(string) Future // nil
	Delete() Future       // nil
}

type FileOp interface {
	Reader() (io.ReadCloser, error)
	Writer(int) (io.WriteCloser, error)
	CommonOp
}

type DirOp interface {
	Read() Future          // []File
	Write([]File) Future   // Task
	Move([]File) Future    // nil
	NewFile(string) Future // nil
	NewDir(string) Future  // nil
	To(string) Future      // File
	CommonOp
}

// view via less
type Viewer interface {
	View() error
}

// open shell
type Sheller interface {
	Shell() error
}

// edit via vi
type Editor interface {
	Edit() error
}

// open by default app
type Opener interface {
	Open() error
}

func NewFile(path string, info os.FileInfo) File {
	return newFile(path, info)
}

func ParsePath(path string) Path {
	return parsePath(path)
}

func LoadFile(file File) Future {
	return loadFile(file)
}

func RegisterLoader(l Loader) {
	registerLoader(l)
}
