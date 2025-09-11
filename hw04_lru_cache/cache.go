package hw04lrucache

type Key string

type Cache interface {
	Set(key Key, value interface{}) bool
	Get(key Key) (interface{}, bool)
	Clear()
}

type lruCache struct {
	capacity int
	queue    List
	items    map[Key]*ListItem
}

type valueWithKey struct {
	value interface{}
	Key
}

func (l *lruCache) Set(key Key, value interface{}) bool {
	if v, ok := l.items[key]; ok {
		v.Value.(*valueWithKey).value = value
		l.queue.MoveToFront(v)
		return true
	}

	if l.queue.Len() == l.capacity {
		back := l.queue.Back()
		l.queue.Remove(back)
		key := back.Value.(*valueWithKey).Key
		delete(l.items, key)
	}

	item := l.queue.PushFront(&valueWithKey{
		value: value,
		Key:   key,
	})
	l.items[key] = item
	return false
}

func (l *lruCache) Get(key Key) (interface{}, bool) {
	if v, ok := l.items[key]; ok {
		l.queue.MoveToFront(v)
		return v.Value.(*valueWithKey).value, true
	}
	return nil, false
}

func (l *lruCache) Clear() {
	l.queue = NewList()
	l.items = make(map[Key]*ListItem, l.capacity)
}

func NewCache(capacity int) Cache {
	return &lruCache{
		capacity: capacity,
		queue:    NewList(),
		items:    make(map[Key]*ListItem, capacity),
	}
}
