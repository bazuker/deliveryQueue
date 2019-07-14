package buffer

import (
	"sync"
	"sync/atomic"
)

const stop int32 = 1

type entry struct {
	v    interface{}
	next *entry
}

// Buffer is a FIFO queue.  The queue can grow to infinite size
// and pushing an item will never fail.
// This value must never be copied once created (in other words, make it a
// pointer value if shared across func/method boundaries).
type Buffer struct {
	ptr  *entry
	last *entry
	mu   sync.Mutex

	ch   chan interface{}
	once sync.Once
	stop int32
}

// Push pushes an item onto the Buffer.
func (b *Buffer) Push(item interface{}) {
	q := entry{v: item}

	b.mu.Lock()
	if b.ptr == nil {
		b.ptr = &q
		b.last = &q
	} else {
		b.last.next = &q
		b.last = &q
	}
	b.mu.Unlock()
}

// Pop pops an item from the Buffer. If an item cannot be returned, it returns ok == false.
// Note: Do not use Pop() and Next() together, use one or the other.
// Note: It is safe to use Pop() and Pull() together.
func (b *Buffer) Pop() (val interface{}, ok bool) {
	b.mu.Lock()
	if b.ptr != nil {
		v := b.ptr.v
		b.ptr = b.ptr.next
		b.mu.Unlock()
		return v, true
	}
	b.mu.Unlock()
	return nil, false
}

// Pull will block until it can pop an item from the buffer.
// Note: Do not use Pull() and Next() together, use one or the other.
// Note: It is safe to use Pop() and Pull() together.
func (b *Buffer) Pull() interface{} {
	sleeper := Sleeper{}
	for {
		v, ok := b.Pop()
		if ok {
			return v
		}
		sleeper.Sleep()
	}
}

// Next pulls an item from the Buffer until Close() is called.  This is nice for
// for/range loops, but is slightly slower than Pull() and requires a goroutine
// per Next() call.
// Note: Do not use Pop/Pull() and Next() together, use one or the other.
func (b *Buffer) Next() chan interface{} {
	b.mu.Lock()
	if b.ch == nil {
		b.ch = make(chan interface{}, 100)
	}
	b.mu.Unlock()

	go func() {
		for {
			// See if we have been told to stop by someone calling Close().
			if atomic.LoadInt32(&b.stop) == stop {
				b.once.Do(func() {
					close(b.ch)
				})
				return
			}

			b.ch <- b.Pull()
		}
	}()

	return b.ch
}

// Close closes the output channel used in Next() calls.  This is only needed
// if you are using .Next() and not Pop() or Pull().
func (b *Buffer) Close() {
	atomic.StoreInt32(&b.stop, stop)
}
