package script

import (
	"context"
	"errors"
)

// Future：异步执行句柄
type Future struct {
	ch   chan *Result
	ctx  context.Context
	done bool
}

func (f *Future) Wait() *Result {
	if f.done {
		select {
		case r := <-f.ch:
			f.ch <- r // 放回供 Result/Err 再读（简单复用一次）
			return r
		default:
			return &Result{Success: false, Err: errors.New("future already consumed")}
		}
	}
	r := <-f.ch
	// 标记“可复读一次”：再塞回去
	f.ch <- r
	f.done = true
	return r
}
func (f *Future) Result() (any, error) {
	r := f.Wait()
	return r.Value, r.Err
}
func (f *Future) Err() error {
	return f.Wait().Err
}
func (f *Future) Context() context.Context { return f.ctx }
