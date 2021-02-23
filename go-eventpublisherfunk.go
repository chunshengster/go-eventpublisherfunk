package eventpublisherfunk

import (
	"context"
	"fmt"
	"sync"

	"go.uber.org/atomic"
)

const (
	// DefaultBusCapacity          = 128
	// DefaultDataChannelCapacity describe the cap of data channel
	DefaultDataChannelCapacity = 65535
	// DefaultErrorChannelCapacity describe the cap of error channel
	DefaultErrorChannelCapacity = 1024
)

// EventData describe the standard event data struct
type EventData struct {
	ID    string
	Data  interface{}
	Topic string
}

//HandleFunc describe the func that used for handle EventData with special Topic
type HandleFunc func(EventData) (eventId string, err error)

// ErrorhandlFunc describe the error handler if error occurred,
type ErrorHandleFunc func(error)

type DataChannel chan EventData
type ErrorChannel chan error

type Publisher interface {
	PublishEvent(EventData) (string, error)
	PublishEventAsync(EventData) (string, error)
	RegisterTopicHandleFunc(string, HandleFunc) error
	RegisterErrorHandleFunc(ErrorHandleFunc) error
	CloseAsyncPublisher() error
	DumpBus() *Bus
}

type Bus struct {
	handles    *sync.Map
	errhandles ErrorHandleFunc
}

type publisher struct {
	Bus      *Bus
	DataCh   DataChannel
	ErrorCh  ErrorChannel
	async    *atomic.Bool
	ctx      context.Context
	cancel   context.CancelFunc
	workerWG *sync.WaitGroup
	bufferWG *sync.WaitGroup
}

// NewPublisher creates a new Publisher instance,where could be used through Publish.PublishEvent()
// No need to Close()
func NewPublisher() Publisher {
	p := new(publisher)
	p.Bus = &Bus{
		handles:    new(sync.Map),
		errhandles: nil,
	}
	return p
}

// NewAsyncPublish creates a new Publisher instance with publishasyncWorker, which has many goroutines to  process EventData.
// Need close the instance via CloseAsyncPublisher(), see examples in examples directory
func NewAsyncPublisher() Publisher {
	ctx, cancel := context.WithCancel(context.Background())
	p := &publisher{
		Bus: &Bus{
			handles: new(sync.Map),
		},
		DataCh:   make(DataChannel, DefaultDataChannelCapacity),
		ErrorCh:  make(ErrorChannel, DefaultErrorChannelCapacity),
		async:    atomic.NewBool(true),
		ctx:      ctx,
		cancel:   cancel,
		workerWG: &sync.WaitGroup{},
		bufferWG: &sync.WaitGroup{},
	}
	p.publisherasyncWorker()
	return p
}

// CloseAsyncPublisher() close the instance of Publisher
func (p *publisher) CloseAsyncPublisher() error {
	select {
	case <-p.ctx.Done():
	default:
		p.async.Swap(false)
		p.cancel()

		p.workerWG.Wait()
	}
	return nil
}

// DumpBus dumps the current Bus struct
func (p *publisher) DumpBus() *Bus {
	return p.Bus
}

// RegisterTopicHandleFunc register consumer of topic, which may do many things sync or async.
// users should privide (f HandleFunc) first
func (p *publisher) RegisterTopicHandleFunc(topic string, f HandleFunc) error {
	p.Bus.handles.Store(topic, f)
	return nil
}

// RegisterErrorHandleFunc register consumer of errors which may renturn by HandleFunc
func (p *publisher) RegisterErrorHandleFunc(f ErrorHandleFunc) error {
	p.Bus.errhandles = f
	return nil
}

// PublishEvent publish a DataEvent into Publish instance, the DataEvent data may be process by HandleFunc
func (p *publisher) PublishEvent(d EventData) (string, error) {
	return p.publish(d)
}

// PublishEventAsync publish a DataEvent into Publish instance, the Publisher instance must be initialized through NewAsyncPublisher()
func (p *publisher) PublishEventAsync(d EventData) (string, error) {
	if !p.async.Load() {
		return "", fmt.Errorf("DefaultError,need async")
	}
	select { //
	case <-p.ctx.Done():
		return "", fmt.Errorf("DefaultError")
	default:
	}
	p.bufferWG.Add(1)
	defer p.bufferWG.Done()
	p.DataCh <- d
	return d.ID, nil
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
	p.workerWG.Add(1)
	go func(wg *sync.WaitGroup) {
		defer wg.Done()
		for d := range p.DataCh {
			_, err := p.publish(d)
			if err != nil {
				p.ErrorCh <- err
			}

		}
	}(p.workerWG)

	p.workerWG.Add(1)
	go func(wg *sync.WaitGroup) {
		defer wg.Done()
		for err := range p.ErrorCh {
			if p.Bus.errhandles != nil {
				p.Bus.errhandles(err)
			}
		}
	}(p.workerWG)

	p.workerWG.Add(1)
	go func(wg *sync.WaitGroup) {
		defer wg.Done()
		<-p.ctx.Done()
		{
			p.bufferWG.Wait()
			close(p.DataCh)
			close(p.ErrorCh)
		}
	}(p.workerWG)
	return nil
}
