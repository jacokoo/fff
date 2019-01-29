package executor

import (
	"sync"
	"testing"
	"time"
)

func TestFuture(t *testing.T) {
	f := newFuture()
	wg := new(sync.WaitGroup)
	wg.Add(10)
	for i := 0; i < 5; i++ {
		go func(idx int) {
			r, e := f.Get()
			t.Logf("get index %d, (%v, %v)", idx, r, e)
			if e != nil || r != 1 {
				t.Fail()
			}
			wg.Done()
		}(i)
	}

	for i := 0; i < 5; i++ {
		func(idx int) {
			f.OnComplete(func(data interface{}, err error) {
				t.Logf("complete index %d, (%v, %v)", idx, data, err)
				if err != nil || data != 1 {
					t.Fail()
				}
				wg.Done()
			})
		}(i)
	}

	time.Sleep(time.Second)
	f.set(1, nil)
	wg.Wait()

	if f.state != futureCompleted {
		t.Fatalf("state incorrect")
	}
}

func TestFutureCancel(t *testing.T) {
	f := newFuture()
	wg := new(sync.WaitGroup)
	wg.Add(10)
	for i := 0; i < 5; i++ {
		go func(idx int) {
			r, e := f.Get()
			t.Logf("get index %d, (%v, %v)", idx, r, e)
			if e == nil || r != nil {
				t.Fail()
			}
			wg.Done()
		}(i)
	}

	for i := 0; i < 5; i++ {
		func(idx int) {
			f.OnComplete(func(data interface{}, err error) {
				t.Logf("complete index %d, (%v, %v)", idx, data, err)
				if err == nil || data != nil {
					t.Fail()
				}
				wg.Done()
			})
		}(i)
	}

	time.Sleep(time.Second)
	f.Cancel()
	wg.Wait()

	if f.state != futureCancelled {
		t.Fatalf("state incorrect")
	}
}
