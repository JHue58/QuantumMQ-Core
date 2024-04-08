package container

import (
	"testing"
)

func TestQueue(t *testing.T) {
	queue := NewMsgQueue[[]byte](1)

	queue.Push([]byte("123"))
	ExpandMsgQueue(queue, 1)
	_, err := queue.Push([]byte("123"))
	if err != nil {
		panic(err)
	}

}
