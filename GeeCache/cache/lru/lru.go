package lru

import "container/list"

type Cache struct {
	maxBytes  int64 // maxBytes 最大缓存数据长度，0 表示不限制
	nBytes    int64 // nBytes 当前缓存数据长度
	m         map[string]*list.Element
	l         *list.List
	onEvicted func(key string, value Value) // onEvicted 数据被淘汰时的回调函数
}

// entry 双向链表节点的数据类型
type entry struct {
	key   string
	value Value
}

// Value 值接口
//go:generate mockery --name Value --case snake
type Value interface {
	// Len 返回value的长度
	Len() int
}

type OnEvictedFunc func(key string, value Value)

// New 创建并初始化 lru cache
func New(maxBytes int64, onEvicted OnEvictedFunc) *Cache {
	return &Cache{
		maxBytes:  maxBytes,
		nBytes:    0,
		m:         make(map[string]*list.Element),
		l:         list.New(),
		onEvicted: onEvicted,
	}
}

// Get 根据key获取缓存的value
func (c *Cache) Get(key string) (v Value, ok bool) {
	if element, ok := c.m[key]; ok {
		c.l.MoveToFront(element)
		kv := element.Value.(*entry)
		return kv.value, true
	}
	return
}

// Put 向 Cache 中添加一个k，v
func (c *Cache) Put(key string, value Value) {
	if ele, ok := c.m[key]; ok {
		c.l.MoveToFront(ele)
		kv := ele.Value.(*entry)
		c.nBytes += int64(value.Len() - kv.value.Len())
		kv.value = value
	} else {
		ele := c.l.PushFront(&entry{key, value})
		c.m[key] = ele
		c.nBytes += int64(len(key) + value.Len())
	}

	for c.maxBytes != 0 && c.nBytes > c.maxBytes {
		c.removeOldest()
	}
}

// Len 返回元素个数
func (c *Cache) Len() int {
	return c.l.Len()
}

// removeOldest 移除最旧元素
func (c *Cache) removeOldest() {
	ele := c.l.Back()
	if ele == nil {
		return
	}

	c.l.Remove(ele)
	kv := ele.Value.(*entry)
	delete(c.m, kv.key)
	c.nBytes -= int64(len(kv.key) + kv.value.Len())
	if c.onEvicted != nil {
		c.onEvicted(kv.key, kv.value)
	}
}
