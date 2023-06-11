package queue

import (
	"deliveryQueue/queue/buffer"
	"github.com/pkg/errors"
	"sync/atomic"
	"time"
)

const stop = 1

var (
	ErrInvalidMaxOps   = errors.New("maximum operations per second has to be greater than zero")
	ErrDeliveryFuncNil = errors.New("delivery function is nil")
)

type DeliveryFunc = func(item interface{})

// DeliveryQueue is a queue with embedded rate limiter
// that guarantees that deliveryFunc will not be executed more than
// specified number of times per second.
type DeliveryQueue struct {
	buffer       *buffer.Buffer
	interval     time.Duration
	lastDelivery time.Time
	deliver      DeliveryFunc
	stop         int32
}

func NewDeliveryQueue(maxOperationsPerSecond int, deliveryFunc DeliveryFunc) (*DeliveryQueue, error) {
	if maxOperationsPerSecond < 1 {
		return nil, ErrInvalidMaxOps
	}
	if deliveryFunc == nil {
		return nil, ErrDeliveryFuncNil
	}
	return &DeliveryQueue{
		buffer:       &buffer.Buffer{},
		interval:     time.Second / time.Duration(maxOperationsPerSecond),
		lastDelivery: time.Now(),
		deliver:      deliveryFunc,
	}, nil
}

// Add adds an item to the queue
func (q *DeliveryQueue) Add(item interface{}) {
	q.buffer.Push(item)
}

// Poll is processing items in the queue and delivers them.
// The function is blocking until StopPolling is called
func (q *DeliveryQueue) Poll() {
	for {
		if atomic.LoadInt32(&q.stop) == stop {
			return
		}
		item := q.buffer.Pull() // Blocks until the next item is available
		if time.Now().Sub(q.lastDelivery) < q.interval {
			time.Sleep(q.interval)
		}
		go q.deliver(item)
		q.lastDelivery = time.Now()
	}
}

// StopPolling stops execution of Poll and unblocks the goroutine
func (q *DeliveryQueue) StopPolling() {
	atomic.StoreInt32(&q.stop, stop)
}
