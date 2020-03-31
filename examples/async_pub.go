package main

import (
	"fmt"
	"strconv"
	"sync"

	pfunk "github.com/chunshengster/go-eventpublisherfunk"
)

func async_callbacks(d pfunk.EventData) (string, error) {
	fmt.Println("got from asyc publisher", ":", d.ID, ":", d.Data, ":", d.Topic)
	return d.ID, nil
}

func error_handle(e error) {
	fmt.Println(e)
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
	err = p.RegisterErrorHandleFunc(error_handle)
	if err != nil {
		fmt.Println("p.RegisterErrorHandleFunc returned error", ":", err)
		return
	}

	wg := &sync.WaitGroup{}
	for i := 0; i < 10000; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			id, err := p.PublishEventAsync(pfunk.EventData{
				ID:    strconv.Itoa(i),
				Data:  "async data " + strconv.Itoa(i),
				Topic: "asyncTopic",
			})
			fmt.Println("publish event async data", ":", id, "err", ":", err)
		}(i)
	}
	wg.Wait()
}
