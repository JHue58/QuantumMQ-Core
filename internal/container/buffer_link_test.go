package container

import (
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
)

func TestLinkedList_Append_Next(t *testing.T) {
	link := NewLinkedList[int]()
	num := 3
	for i := 0; i < num; i++ {
		link.Append(i)
	}

	for i := 0; i < num; i++ {
		n, _ := link.Next()
		assert.Equal(t, i, n)
	}

	assert.Equal(t, 0, link.Len())

}

func TestLinkedList_RollBack(t *testing.T) {
	link := NewLinkedList[int]()
	num := 100
	for i := 0; i < num; i++ {
		link.Append(i)
	}

	next, err := link.Next()
	assert.NoError(t, err)
	assert.Equal(t, 0, next)
	assert.Equal(t, 99, link.Len())

	link.RollBack()
	assert.Equal(t, 100, link.Len())
	after, err := link.Next()
	assert.NoError(t, err)
	assert.Equal(t, 0, after)

	link.Release()
	link.RollBack()
	assert.Equal(t, 99, link.Len())

	for {
		_, err := link.Next()
		if err != nil {
			break
		}
	}

	assert.Equal(t, 0, link.Len())

	link.RollBack()
	link.Append(101)
	assert.Equal(t, 100, link.Len())

}

func TestLinkedList_Slice(t *testing.T) {
	link := NewLinkedList[int]()
	num := 100
	for i := 0; i < num; i++ {
		link.Append(i)
	}
	sl, err := link.Slice(101)
	assert.Error(t, err)
	sl, err = link.Slice(60)
	assert.NoError(t, err)
	assert.Equal(t, 60, sl.Len())
	assert.Equal(t, 40, link.Len())
	link.Append(101)
	next, err := link.Next()
	assert.NoError(t, err)
	assert.Equal(t, 60, next)
	next, err = sl.Next()
	assert.NoError(t, err)
	assert.Equal(t, 0, next)
	idx := 1

	length := sl.Len()
	for i := 0; i < length; i++ {
		next, err = sl.Next()
		assert.NoError(t, err)
		assert.Equal(t, idx, next)
		idx++
	}

	assert.Equal(t, 0, sl.Len())
	_, err = sl.Next()
	assert.Error(t, err)
	link.Release()
	sl.Release()

}

func TestLinkedList_Release(t *testing.T) {
	link := NewLinkedList[int]()
	num := 100
	for i := 0; i < num; i++ {
		link.Append(i)
	}
	link.Release()
	assert.Equal(t, num, link.Len())
	link.Next()
	link.Release()
	assert.Equal(t, num-1, link.Len())
	for link.Len() > 0 {
		link.Next()
	}
	assert.Equal(t, 0, link.Len())
	link.Release()
	assert.Equal(t, link.head, link.read)
	assert.Equal(t, link.head, link.write)

}

func TestLinkedList_RWRace(t *testing.T) {
	// use go test -race
	num := 100000
	wg := sync.WaitGroup{}
	wg.Add(num)
	link := NewLinkedList[int]()
	go func() {
		for i := 0; i < num; i++ {
			link.Append(i)
		}
	}()
	go func() {
		for i := 0; i < num; i++ {
			for {
				value, err := link.Next()
				if err != nil {
					continue
				}
				assert.Equal(t, i, value)
				break
			}
			wg.Done()
		}
	}()
	wg.Wait()
}
