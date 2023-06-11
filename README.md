# DeliveryQueue
[![Go Report Card](https://goreportcard.com/badge/github.com/bazuker/deliveryQueue)](https://goreportcard.com/report/github.com/bazuker/deliveryQueue)

DeliveryQueue is a queue with embedded rate limiter that guarantees that delivery function will not be triggered more than specified number of times per second.

```Bash
go get -u github.com/bazuker/deliveryQueue
```

## Example
```Go
const (
	MaxMessagesPerSecond = 30
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

	log.Println("All messages are submitted")

	start := time.Now()
	wg.Wait()
	log.Printf("Delivered %d messages in %f seconds", MessagesToDeliver, time.Now().Sub(start).Seconds())
}
```

```
2019/07/13 17:25:57 All messages are submitted
2019/07/13 17:25:57 1
2019/07/13 17:25:57 2
...
2019/07/13 17:26:01 119
2019/07/13 17:26:01 120
2019/07/13 17:26:01 Delivered 120 messages in 4.148081 seconds
```