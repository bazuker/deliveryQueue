# DeliveryQueue

DeliveryQueue is a queue with embedded rate limiter that guarantees that delivery function will not be triggered more than specified number of times per second.

## Example
```Go
const (
	MaxMessagesPerSecond = 31
	MessagesToDeliver = 120
)

func main() {
	wg := &sync.WaitGroup{}

	dq, _ := queue.NewDeliveryQueue(MaxMessagesPerSecond, func(item interface{}) {
		log.Println(item)
		wg.Done()
	})

	go dq.Poll()

	for i := 1; i <= MessagesToDeliver; i++ {
		dq.Add(i)
	}

	wg.Add(MessagesToDeliver)

	log.Println("All items are submitted")

	start := time.Now()
	wg.Wait()
	log.Printf("Delivered %d messages in %f seconds", MessagesToDeliver, time.Now().Sub(start).Seconds())
}
```