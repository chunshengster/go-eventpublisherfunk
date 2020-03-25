package eventpublisherfunk

import (
	"fmt"
	"sync"
)

const (
	DefaultBusCapacity          = 128
	DefaultDataChannelCapacity  = 1024
	DefaultErrorChannelCapacity = 1024
)

type Topic string
type EventData struct {
	ID   string
	Data interface{}
	Topic
}
type HandleFunc func(EventData) (eventId string, err error)
type ErrorHandleFunc func(error)

type HandleFuncs []HandleFunc
type DataChannel chan EventData
type ErrorChannel chan error

type Publisher interface {
	PublishEvent(EventData) (string, error)
	PublishEventAsync(EventData) (string, error)
	RegisterTopicHandleFunc(string, HandleFunc) error
	RegisterErrorHandleFunc(ErrorHandleFunc) error
}

type Bus struct {
	handles    *sync.Map
	errhandles ErrorHandleFunc
}

type publisher struct {
	Bus     *Bus
	DataCh  DataChannel
	ErrorCh ErrorChannel
	async   bool
}

func NewPublisher() Publisher {
	p := new(publisher)
	p.Bus.handles = new(sync.Map)
	return p
}

func NewAsyncPublisher() Publisher {
	p := &publisher{
		Bus: &Bus{
			handles: new(sync.Map),
		},
		DataCh:  make(DataChannel, DefaultDataChannelCapacity),
		ErrorCh: make(ErrorChannel, DefaultErrorChannelCapacity),
		async:   true,
	}
	return p
}

func (p *publisher) RegisterTopicHandleFunc(topic string, f HandleFunc) error {
	p.Bus.handles.Store(topic, f)
	return nil
}

func (p *publisher) RegisterErrorHandleFunc(f ErrorHandleFunc) error {
	p.Bus.errhandles = f
	return nil
}

func (p *publisher) PublishEvent(d EventData) (string, error) {
	return p.publish(d)
}

func (p *publisher) PublishEventAsync(e EventData) (string, error) {
	return "", fmt.Errorf("DefaultError,need redefine")
}

func (p *publisher) publish(d EventData) (string, error) {
	t := d.Topic
	if t != "" {
		if f, ok := p.Bus.handles.Load(t); ok {
			return f.(HandleFunc)(d)
		}
	}
	return "", fmt.Errorf("DefaultError,need redefine")
}

func (p *publisher) publisherasyncWorker() error {
	go func() {
		wg := &sync.WaitGroup{}
		wg.Add(2)
		go func(wg *sync.WaitGroup) {
			defer wg.Done()
			for d := range p.DataCh {
				_, err := p.publish(d)
				if err != nil {
					p.ErrorCh <- err
				}
			}
		}(wg)

		go func(wg *sync.WaitGroup) {
			defer wg.Done()
			for err := range p.ErrorCh {
				if p.Bus.errhandles != nil {
					p.Bus.errhandles(err)
				}
			}
		}(wg)

		wg.Wait()
	}()
	return nil
}
