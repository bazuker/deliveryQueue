package queue

import (
	"math"
	"sync/atomic"
	"testing"
	"time"
)

// This tests only ensures workability of the algorithm.
// Incompatible with race testing
func TestDeliveryQueue(t *testing.T) {
	const MaxMessagesPerSecond = 31
	const MessagesToDeliver = 120

	var done int32 = 0
	atomic.StoreInt32(&done, 0)

	dq, _ := NewDeliveryQueue(MaxMessagesPerSecond, func(item interface{}) {
		atomic.AddInt32(&done, 1)
	})

	go dq.Poll()

	for i := 1; i <= MessagesToDeliver; i++ {
		dq.Add(i)
	}

	maxSecondsToRun := time.Second*time.Duration(math.Round(MessagesToDeliver/MaxMessagesPerSecond)+2)

	afterFuncTimer := time.AfterFunc(maxSecondsToRun, func() {
		delivered := atomic.LoadInt32(&done)
		if delivered < MessagesToDeliver {
			t.Errorf("did not deliver all messages in time. Delivered %d out of %d", delivered, MessagesToDeliver)
		}
	})
	defer afterFuncTimer.Stop()

	time.Sleep(maxSecondsToRun)
}
