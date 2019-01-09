package model

import (
	"fmt"
	"os"
	"path/filepath"
)

// CopySource a copy source
type CopySource []FileItem

func getCopyItems(ci FileItem, target string) ([]string, []string) {
	if !ci.IsDir() {
		return []string{ci.Path()}, []string{filepath.Join(target, filepath.Base(ci.Path()))}
	}

	op, np := make([]string, 0), make([]string, 0)
	filepath.Walk(ci.Path(), func(path string, fi os.FileInfo, err error) error {
		ff := NewFile(path, fi)
		if ff.IsDir() {
			return nil
		}
		if ln, ok := ff.Link(); ok && ln.IsBroken() {
			return nil
		}

		rel, _ := filepath.Rel(filepath.Dir(ci.Path()), path)
		op = append(op, path)
		np = append(np, filepath.Join(target, rel))

		return nil
	})
	return op, np
}

// CopyTask create copy task
func (cs CopySource) CopyTask(op Operator, target string) (Task, bool) {
	ops, nps := make([]string, 0), make([]string, 0)
	for _, v := range cs {
		os, ns := getCopyItems(v, target)
		ops = append(ops, os...)
		nps = append(nps, ns...)
	}
	if len(ops) == 0 {
		return nil, false
	}

	tasks := make([]Task, len(ops))
	for i, v := range ops {
		oldpath := v
		newpath := nps[i]
		tasks[i] = NewTask(filepath.Base(v), func(progress chan<- int, quit <-chan bool, err chan<- error) {
			op.CopyFile(oldpath, newpath, progress, err, quit)
		})
	}

	return NewBatchTask("Copy", tasks), true
}

// MoveTo move copy souce items to target
func (cs CopySource) MoveTo(op Operator, target string) error {
	ops, nps := make([]string, 0), make([]string, 0)
	for _, v := range cs {
		os, ns := getCopyItems(v, target)
		ops = append(ops, os...)
		nps = append(nps, ns...)
	}
	if len(ops) == 0 {
		return fmt.Errorf("No file to move")
	}

	for i, v := range ops {
		err := op.Move(v, nps[i])
		if err != nil {
			return err
		}
	}
	return nil
}
