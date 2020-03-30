package main

import (
	"fmt"
	"strconv"

	pfunk "github.com/chunshengster/go-eventpublisherfunk"
)

func async_callbacks(d pfunk.EventData) (string, error) {
	fmt.Println("got from asyc publisher", ":", d.ID, ":", d.Data, ":", d.Topic)
	return d.ID, nil
}

func main() {
	// mywg := sync.WaitGroup{}
	p := pfunk.NewAsyncPublisher()
	defer p.CloseAsyncPublisher()
	err := p.RegisterTopicHandleFunc("asyncTopic", async_callbacks)
	if err != nil {
		fmt.Println("p.RegisterTopicHandleFunc returned error", ":", err)
		return
	}

	for i := 0; i < 1000; i++ {
		id, err := p.PublishEventAsync(pfunk.EventData{
			ID:    strconv.Itoa(i),
			Data:  "async data " + strconv.Itoa(i),
			Topic: "asyncTopic",
		})
		fmt.Println("publish event async data", ":", id, "err", ":", err)
	}
}
