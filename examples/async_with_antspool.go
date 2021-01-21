package examples

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

func AsyncWithAntspool() {
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
