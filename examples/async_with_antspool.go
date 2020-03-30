package main

import (
	"fmt"
	"strconv"

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
	apf, err := ants.NewPoolWithFunc(100, realCallback)
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
	for i := 0; i < 100; i++ {
		re, err := p.PublishEventAsync(pfunk.EventData{
			ID:    strconv.Itoa(i),
			Data:  "antsTestData" + strconv.Itoa(i),
			Topic: "antsTest"})
		fmt.Println("p.PublishEventAsync returned ", ":", re, ":", err)
	}
}
