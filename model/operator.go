package model

import (
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
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
	CopyFile(path, newPath string, result chan<- error)
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
func (o *LocalOperator) CopyFile(path, newPath string, result chan<- error) {
	defer close(result)
	fo, err := os.Open(path)
	if err != nil {
		result <- err
		return
	}

	fn, err := os.OpenFile(newPath, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		result <- err
		return
	}

	_, err = io.Copy(fn, fo)
	if err != nil {
		result <- err
	}
}

// Open file
func (o *LocalOperator) Open(path string) error {
	cmd := exec.Command("open", path)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	return cmd.Start()
}
