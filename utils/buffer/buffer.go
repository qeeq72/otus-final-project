package buffer

import (
	"sync"
	"time"
)

type IBufferGetSetter interface {
	Set(ts time.Time, value interface{}) bool
	Get(ts time.Time) (interface{}, bool)
	GetPeriod(ts time.Time, period int) ([]interface{}, bool)
	Clear()
}

type Buffer struct {
	capacity int
	queue    IList
	items    map[time.Time]*listItem
	mu       sync.Mutex
}

type bufferItem struct {
	key   time.Time
	value interface{}
}

func NewBuffer(capacity int) IBufferGetSetter {
	return &Buffer{
		capacity: capacity,
		queue:    newList(),
		items:    make(map[time.Time]*listItem, capacity),
	}
}

func (b *Buffer) Set(ts time.Time, value interface{}) bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	ts = ts.Truncate(time.Second)
	if _, exists := b.items[ts]; exists {
		b.items[ts].Value = bufferItem{key: ts, value: value}
		b.queue.MoveToFront(b.items[ts])
		return true
	}

	if b.queue.Len() == b.capacity {
		i := b.queue.Back()
		iValue, ok := i.Value.(bufferItem)
		if !ok {
			return false
		}
		delete(b.items, iValue.key)
		b.queue.Remove(b.queue.Back())
	}

	b.items[ts] = b.queue.PushFront(bufferItem{key: ts, value: value})
	return false
}

func (b *Buffer) Get(ts time.Time) (interface{}, bool) {
	b.mu.Lock()
	defer b.mu.Unlock()

	ts = ts.Truncate(time.Second)
	if i, exists := b.items[ts]; exists {
		iValue, ok := i.Value.(bufferItem)
		if !ok {
			return nil, false
		}
		return iValue.value, true
	}

	return nil, false
}

func (b *Buffer) GetPeriod(ts time.Time, period int) ([]interface{}, bool) {
	b.mu.Lock()
	defer b.mu.Unlock()

	ts = ts.Truncate(time.Second)
	var values []interface{}
	for i := 0; i < period; i++ {
		if i, exists := b.items[ts.Add(time.Duration(i)*time.Second)]; exists {

			iValue, ok := i.Value.(bufferItem)
			if !ok {
				return nil, false
			}

			values = append(values, iValue.value)
		}
	}

	return values, true
}

func (b *Buffer) Clear() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.queue = newList()
	for k := range b.items {
		delete(b.items, k)
	}
}
