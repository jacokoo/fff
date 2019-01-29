package executor

import (
	"testing"
	"time"
)

func TestExecutor(t *testing.T) {
	exec := Fixed(2, false)
	fu, _ := exec.Submit("hello", func(ts TaskState) (interface{}, error) {
		time.Sleep(time.Second * 2)
		return 1, nil
	})
	fu2, _ := exec.Submit("world", func(ts TaskState) (interface{}, error) {
		return 2, nil
	})

	re, _ := fu.Get()
	re2, _ := fu2.Get()
	if re != 1 || re2 != 2 {
		t.Fail()
	}
}
