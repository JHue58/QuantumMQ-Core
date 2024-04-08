package container

import (
	errors "QuantumMQ-Core/pkg/errors"
	"fmt"
)

// MsgQueue 消息队列（一个循环队列）
type MsgQueue[T any] struct {
	Items []T    // 队列元素
	Head  uint64 // 队首指针
	Tail  uint64 //队尾指针
	Cap   uint64 // 队列容量
}

// NewMsgQueue 创建新的空消息队列
func NewMsgQueue[T any](cap uint64) *MsgQueue[T] {
	return &MsgQueue[T]{
		Items: make([]T, cap+1),
		Head:  0,
		Tail:  0,
		Cap:   cap + 1,
	}
}

// ExpandMsgQueue 扩容消息队列
func ExpandMsgQueue[T any](queue *MsgQueue[T], cap uint64) {
	queue.Items = append(queue.Items, make([]T, cap)...)
	queue.Cap = queue.Cap + cap
	fmt.Println("队列发生扩容")
}

// IsEmpty 队列是否为空
func (queue *MsgQueue[T]) IsEmpty() bool {
	return queue.Head == queue.Tail
}

// IsFull 队列是否已满
func (queue *MsgQueue[T]) IsFull() bool {
	return (queue.Tail+1)%queue.Cap == queue.Head // 判断尾指针的下一个是否是头指针
}

// Push 将消息块添加到该队列末尾
func (queue *MsgQueue[T]) Push(msg T) (uint64, *errors.QError) {
	// return异常
	if queue.IsFull() {
		return 0, errors.New("队列已满")
	}
	queue.Items[queue.Tail] = msg
	tail := queue.Tail
	queue.Tail = (queue.Tail + 1) % queue.Cap // 尾指针后移
	return tail, nil
}

// Pop 将该队列首元素弹出并返回
func (queue *MsgQueue[T]) Pop() (value T, err *errors.QError) {
	if queue.IsEmpty() {
		return value, errors.New("队列为空")
	}
	msg := queue.Items[queue.Head]
	queue.Head = (queue.Head + 1) % queue.Cap // 头指针后移
	return msg, nil
}

// Peek 获取队首元素但不出队
func (queue *MsgQueue[T]) Peek() (value T, err *errors.QError) {
	if queue.IsEmpty() {
		return value, errors.New("队列为空")
	}
	return queue.Items[queue.Head], nil
}

// Back 获取该队列尾元素但不出队
func (queue *MsgQueue[T]) Back() (value T, err *errors.QError) {
	if queue.IsEmpty() {
		return value, errors.New("队列为空")
	}
	return queue.Items[(queue.Tail+queue.Cap-1)%queue.Cap], nil // 因为 tail 指向的那个少用的位置,所以不可以直接return items[tail]
}
