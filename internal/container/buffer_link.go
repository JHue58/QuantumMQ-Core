package container

import (
	"fmt"
	"sync"
	"sync/atomic"
)

type Pined[T any] interface {
	PinNext() (value T, err error)
	PinRelease()
	PinRollBack()
	Len() int
}

// LinkedList 可并行读写的链表
type LinkedList[T any] struct {
	wLock sync.Mutex
	rLock sync.Mutex
	pLock sync.Mutex

	length int64
	head   *linkNode
	read   *linkNode
	write  *linkNode

	malloc int
}

func (l *LinkedList[T]) lock() {
	l.rLock.Lock()
}

func (l *LinkedList[T]) unlock() {
	l.rLock.Unlock()
}

func (l *LinkedList[T]) Pin() Pined[T] {
	l.rLock.Lock()
	l.pLock.Lock()
	return l
}

func NewLinkedList[T any]() *LinkedList[T] {
	l := &LinkedList[T]{}
	node := nodePool.Get().(*linkNode)
	l.head = node
	l.read = node
	l.write = node
	return l
}

// Append 添加数据
func (l *LinkedList[T]) Append(value T) {
	l.wLock.Lock()
	node := l.write
	node.value = value
	l.growth()
	l.recalLen(1)
	l.wLock.Unlock()
}

func (l *LinkedList[T]) growth() {
	if l.write.next == nil {
		l.write.next = nodePool.Get().(*linkNode)
		l.write = l.write.next
	}
}

// Next 读取数据
func (l *LinkedList[T]) Next() (value T, err error) {
	if l.Len() <= 0 {
		return value, fmt.Errorf("empty")
	}
	l.lock()
	l.recalLen(-1)
	l.malloc++

	read := l.read
	value = read.value.(T)

	l.read = read.next
	l.unlock()
	return
}

func (l *LinkedList[T]) pinCheck() {
	if l.pLock.TryLock() {
		l.pLock.Unlock()
		panic("use Pin Func before PIN")
	}
}

func (l *LinkedList[T]) PinNext() (value T, err error) {
	l.pinCheck()
	if l.Len() <= 0 {
		return value, fmt.Errorf("empty")
	}

	l.recalLen(-1)
	l.malloc++

	read := l.read
	value = read.value.(T)

	l.read = read.next

	return
}

// Slice 截取链表
func (l *LinkedList[T]) Slice(length int) (*LinkedList[T], error) {
	if l.Len() < length {
		return nil, fmt.Errorf("length is not enought")
	}
	l.lock()
	l.recalLen(-length)
	l.malloc += length
	sl := &LinkedList[T]{
		length: int64(length),
		head:   l.head.Refer(),
		read:   l.read.Refer(),
		write:  nil,
	}
	for i := 0; i < length; i++ {
		l.read = l.read.next.Refer()
	}
	for l.head != l.read {
		node := l.head
		l.head = l.head.next
		node.recycle()
	}

	l.unlock()
	return sl, nil

}

// PinRollBack 回滚所有的读取操作（Release之前）
func (l *LinkedList[T]) PinRollBack() {
	l.pinCheck()

	if l.read == l.head {
		l.pLock.Unlock()
		l.rLock.Unlock()
		return
	}
	l.read = l.head
	l.recalLen(l.malloc)
	l.malloc = 0
	l.pLock.Unlock()
	l.rLock.Unlock()
}

// RollBack 回滚所有的读取操作（Release之前）
func (l *LinkedList[T]) RollBack() {
	l.lock()

	if l.read == l.head {
		l.unlock()
		return
	}
	l.read = l.head
	l.recalLen(l.malloc)
	l.malloc = 0
	l.unlock()
}

// Release 释放节点
func (l *LinkedList[T]) Release() {
	l.lock()
	for l.head != l.read {
		node := l.head
		l.head = l.head.next
		node.recycle()
	}
	l.malloc = 0
	l.unlock()
}

// Len 链表长度
func (l *LinkedList[T]) Len() int {
	length := atomic.LoadInt64(&l.length)
	return int(length)
}

// PinRelease 释放节点
func (l *LinkedList[T]) PinRelease() {
	l.pinCheck()
	for l.head != l.read {
		node := l.head
		l.head = l.head.next
		node.recycle()
	}
	l.malloc = 0
	l.pLock.Unlock()
	l.rLock.Unlock()
}

func (l *LinkedList[T]) recalLen(delta int) (length int) {
	return int(atomic.AddInt64(&l.length, int64(delta)))
}

var nodePool = sync.Pool{New: func() any {
	return &linkNode{refer: 1}
}}

type linkNode struct {
	refer uint

	value any
	next  *linkNode
}

func (n *linkNode) Refer() *linkNode {
	n.refer++
	return n
}

func (n *linkNode) recycle() {
	if n.refer > 1 {
		n.refer--
		return
	}
	n.value = nil
	n.next = nil
	nodePool.Put(n)
}
