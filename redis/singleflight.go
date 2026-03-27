package redis

import (
	"sync"
)

// SingleFlight 防止缓存击穿
type SingleFlight struct {
	mu    sync.Mutex
	calls map[string]*call
}

// call 表示一个正在进行中的调用
type call struct {
	wg  sync.WaitGroup
	val interface{}
	err error
	dups int
}

// NewSingleFlight 创建新的 SingleFlight 实例
func NewSingleFlight() *SingleFlight {
	return &SingleFlight{
		calls: make(map[string]*call),
	}
}

// Do 执行函数，确保同一个 key 同时只有一个在执行
func (sf *SingleFlight) Do(key string, fn func() (interface{}, error)) (interface{}, error) {
	sf.mu.Lock()
	if c, ok := sf.calls[key]; ok {
		c.dups++
		sf.mu.Unlock()
		c.wg.Wait()
		return c.val, c.err
	}

	c := &call{}
	c.wg.Add(1)
	sf.calls[key] = c
	sf.mu.Unlock()

	c.val, c.err = fn()
	c.wg.Done()

	sf.mu.Lock()
	delete(sf.calls, key)
	sf.mu.Unlock()

	return c.val, c.err
}

// DoChan 类似于 Do，但返回 channel
func (sf *SingleFlight) DoChan(key string, fn func() (interface{}, error)) <-chan result {
	ch := make(chan result, 1)
	go func() {
		val, err := sf.Do(key, fn)
		ch <- result{val: val, err: err}
	}()
	return ch
}

// Forget 忘记指定的 key
func (sf *SingleFlight) Forget(key string) {
	sf.mu.Lock()
	delete(sf.calls, key)
	sf.mu.Unlock()
}

// result 表示 DoChan 的结果
type result struct {
	val interface{}
	err error
}
