package hw04lrucache

type List interface {
	Len() int
	Front() *ListItem
	Back() *ListItem
	PushFront(v interface{}) *ListItem
	PushBack(v interface{}) *ListItem
	Remove(i *ListItem)
	MoveToFront(i *ListItem)
}

type ListItem struct {
	Value interface{}
	next  *ListItem
	prev  *ListItem
}

type list struct {
	length int
	front  *ListItem
	back   *ListItem
}

func (l *list) Len() int {
	return l.length
}

func (l *list) Front() *ListItem {
	return l.front
}

func (l *list) Back() *ListItem {
	return l.back
}

func (l *list) PushFront(v interface{}) *ListItem {
	item := &ListItem{
		Value: v,
		next:  l.front,
		prev:  nil,
	}
	if l.front != nil {
		l.front.prev = item
	} else {
		l.back = item
	}
	l.front = item
	l.length++
	return item
}

func (l *list) PushBack(v interface{}) *ListItem {
	item := &ListItem{
		Value: v,
		next:  nil,
		prev:  l.back,
	}
	if l.back != nil {
		l.back.next = item
	} else {
		l.front = item
	}
	l.back = item
	l.length++
	return item
}

func (l *list) Remove(i *ListItem) {
	if i.prev != nil {
		i.prev.next = i.next
	} else {
		l.front = i.next
	}

	if i.next != nil {
		i.next.prev = i.prev.next
	} else {
		l.back = i.prev
	}

	l.length--
}

func (l *list) MoveToFront(i *ListItem) {
	if l.length < 2 {
		return
	}

	if i.prev != nil {
		i.prev.next = i.next
	} else {
		return
	}

	if i.next != nil {
		i.next.prev = i.prev
	} else {
		l.back = i.prev
	}

	i.next = l.front
	i.prev = nil
	l.front.prev = i
	l.front = i
}

func NewList() List {
	return new(list)
}
