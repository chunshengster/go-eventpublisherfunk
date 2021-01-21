package examples

import (
	"fmt"

	pfunk "github.com/chunshengster/go-eventpublisherfunk"
)

func CallBackFunc(d pfunk.EventData) (string, error) {
	fmt.Println(d)
	return d.ID, nil
}

func EchoData() {
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
