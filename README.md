# go-eventpublisherfunk
## easy to use for CQRS/ES patten, ealily add event publisher in you main logic.
一个简单的可以用来在主逻辑中增加 Event Publish 操作的library。

使用同步的方式处理 EventData

```go
package main

import (
	"fmt"

	pfunk "github.com/chunshengster/go-eventpublisherfunk"
)

func CallBackFunc(d pfunk.EventData) (string, error) {
	fmt.Println(d)
	return d.ID, nil
}

func main() {
	p := pfunk.NewPublisher()
	p.RegisterTopicHandleFunc("testtopic", CallBackFunc)
	id, err := p.PublishEvent(pfunk.EventData{
		ID:    "1",
		Data:  "publisher test1",
		Topic: "testtopic",
	})
	if err != nil {
		fmt.Println("publish event return error", ":", err)
	} else {
		fmt.Println("publish event id ", ":", id)
	}
}

```



使用异步的方式处理EventData

```go
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

```



结合ants（https://github.com/panjf2000/ants) 作为TopicHandle提高异步处理效率

```go
package main

import (
	"fmt"
	"strconv"
	"sync"

	pfunk "github.com/chunshengster/go-eventpublisherfunk"
	"github.com/panjf2000/ants"
)

// func main() {
// 	pfunk
// }
var AntsPoolWithFunc *ants.PoolWithFunc

func realCallback(i interface{}) {
	defer func(i interface{}) {
		if d, ok := i.(pfunk.EventData); ok {
			fmt.Println("realCallback got input", ":", d.ID, ":", d.Data, ":", d.Topic)
		} else {
			fmt.Println("realCallback got input error", ":", i)
		}
	}(i)
}
func antPoolInit() *ants.PoolWithFunc {
	apf, err := ants.NewPoolWithFunc(1000, realCallback)
	if err != nil {
		fmt.Println("antPoolInit err:", err)
	}
	return apf
}

func AsyncJobHandler(d pfunk.EventData) (string, error) {
	err := AntsPoolWithFunc.Invoke(d)
	if err != nil {
		return "", err
	}
	fmt.Println("AsyncJobHandler antsrunning", ":", AntsPoolWithFunc.Running())
	return d.ID, nil
}

func main() {
	AntsPoolWithFunc = antPoolInit()
	p := pfunk.NewAsyncPublisher()
	defer AntsPoolWithFunc.Release()
	defer p.CloseAsyncPublisher()
	err := p.RegisterTopicHandleFunc("antsTest", AsyncJobHandler)
	if err != nil {
		fmt.Println("RegisterTopicHandleFunc error", ":", err)
		return
	}
	wg := sync.WaitGroup{}
	for i := 0; i < 10000; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			re, err := p.PublishEventAsync(pfunk.EventData{
				ID:    strconv.Itoa(i),
				Data:  "antsTestData" + strconv.Itoa(i),
				Topic: "antsTest"})
			fmt.Println("p.PublishEventAsync returned ", ":", re, ":", err)
		}(i)
	}
	wg.Wait()
}

```

