// Package singlefligh 防止缓存击穿
package singleflight

import "sync"

type call struct {
	wg  sync.WaitGroup
	v   interface{}
	err error
}

type Group struct {
	mu sync.Mutex // protects m
	m  map[string]*call
}

// Do 针对相同的 key，无论 Do 被调用多少次，函数 fn 都只会被调用一次，等待 fn 调用结束了，返回返回值或错误。
func (g *Group) Do(key string, fn func() (interface{}, error)) (interface{}, error) {
	g.mu.Lock()
	if g.m == nil {
		g.m = make(map[string]*call)
	}

	if c, ok := g.m[key]; ok {
		g.mu.Unlock()
		c.wg.Wait()
		return c.v, c.err
	}

	c := new(call)
	c.wg.Add(1)
	g.m[key] = c
	g.mu.Unlock()

	c.v, c.err = fn()
	c.wg.Done()

	g.mu.Lock()
	delete(g.m, key)
	g.mu.Unlock()

	return c.v, c.err
}
