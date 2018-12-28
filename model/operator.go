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

// Operator file operators
type Operator interface {
	ReadDir(path string) ([]FileItem, error)
	Rename(parent, old, new string) error
	Move(path, newPath string) error
	NewFile(parent, name string) error
	NewDir(parent, name string) error
	DeleteFile(path string) error
	DeleteDir(path string) error
	CopyFile(path, newPath string, progress chan<- int, result chan<- error, quit <-chan bool)
	Open(path string) error
}

// LocalOperator use local fs
type LocalOperator struct {
}

// ReadDir read dir files
func (o *LocalOperator) ReadDir(path string) ([]FileItem, error) {
	fs, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, err
	}

	return NewFiles(path, fs), nil
}

// Rename file
func (o *LocalOperator) Rename(parent, file, name string) error {
	return os.Rename(filepath.Join(parent, file), filepath.Join(parent, name))
}

// Move file
func (o *LocalOperator) Move(path, newPath string) error {
	dir := filepath.Dir(newPath)
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return err
	}
	return os.Rename(path, newPath)
}

// NewFile create file
func (o *LocalOperator) NewFile(parent, name string) error {
	_, err := os.Create(filepath.Join(parent, name))
	return err
}

// NewDir create dir
func (o *LocalOperator) NewDir(parent, name string) error {
	return os.MkdirAll(filepath.Join(parent, name), 0755)
}

// DeleteFile delete file
func (o *LocalOperator) DeleteFile(path string) error {
	return os.Remove(path)
}

// DeleteDir delete dir
func (o *LocalOperator) DeleteDir(path string) error {
	return os.RemoveAll(path)
}

// CopyFile copy file
func (o *LocalOperator) CopyFile(path, newPath string, progress chan<- int, result chan<- error, quit <-chan bool) {
	defer close(result)
	defer close(progress)

	if path == newPath {
		return
	}

	fi, err := os.Lstat(path)
	if err != nil {
		result <- err
		return
	}
	si := float64(fi.Size())

	fo, err := os.Open(path)
	if err != nil {
		result <- err
		return
	}
	defer fo.Close()

	dir := filepath.Dir(newPath)
	err = os.MkdirAll(dir, 0755)
	if err != nil {
		result <- err
		return
	}

	fn, err := os.OpenFile(newPath, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		result <- err
		return
	}
	defer fn.Close()

	buf := make([]byte, 4096)
	var count int64
	pg := 0
	var quited = false

	go func() {
		<-quit
		quited = true
	}()

	for !quited {
		n, err := fo.Read(buf)
		if err == io.EOF {
			break
		}

		if err != nil {
			result <- err
			return
		}

		_, err = fn.Write(buf[:n])
		if err != nil {
			result <- err
			return
		}

		count += int64(n)
		pp := int(float64(count) / si * 100)
		if pp > pg {
			pg = pp
			progress <- pg
		}
	}
}

// Open file
func (o *LocalOperator) Open(path string) error {
	open := ""
	switch runtime.GOOS {
	case "darwin":
		open = "open"
	case "linux":
		open = "xdg-open"
	default:
		return fmt.Errorf("not supported")
	}

	cmd := exec.Command(open, path) // macos only
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	return cmd.Start()
}
