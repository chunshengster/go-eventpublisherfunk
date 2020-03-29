package eventpublisherfunk

import (
	"context"
	"fmt"
	"sync"
)

const (
	DefaultBusCapacity          = 128
	DefaultDataChannelCapacity  = 1024
	DefaultErrorChannelCapacity = 1024
)

// type Topic string
type EventData struct {
	ID   string
	Data interface{}
	Topic string
}
type HandleFunc func(EventData) (eventId string, err error)
type ErrorHandleFunc func(error)

// type HandleFuncs []HandleFunc
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
	Bus     *Bus
	DataCh  DataChannel
	ErrorCh ErrorChannel
	async   bool
	ctx     context.Context
	cancel  context.CancelFunc
	wg      sync.WaitGroup
	// antsPool ants.PoolWithFunc
}

func NewPublisher() Publisher {
	p := new(publisher)
	p.Bus = &Bus{
		handles:    new(sync.Map),
		errhandles: nil,
	}
	return p
}

func NewAsyncPublisher() Publisher {
	ctx, cancel := context.WithCancel(context.Background())
	p := &publisher{
		Bus: &Bus{
			handles: new(sync.Map),
		},
		DataCh:  make(DataChannel, DefaultDataChannelCapacity),
		ErrorCh: make(ErrorChannel, DefaultErrorChannelCapacity),
		async:   true,
		ctx:     ctx,
		cancel:  cancel,
		wg:      sync.WaitGroup{},
	}
	p.publisherasyncWorker()
	return p
}

func (p *publisher) WithAnts() Publisher {
	return p
}

func (p *publisher) CloseAsyncPublisher() error {
	select {
	case <-p.ctx.Done():
	default:
		p.cancel()
		p.wg.Wait()
		close(p.DataCh)
		close(p.ErrorCh)
	}
	return nil
}

func (p *publisher) DumpBus() *Bus {
	return p.Bus
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

func (p *publisher) PublishEventAsync(d EventData) (string, error) {
	if p.async != true {
		return "", fmt.Errorf("DefaultError,need async")
	}
	select { //
	case <-p.ctx.Done():
		return "", fmt.Errorf("DefaultError")
	default:
		p.wg.Add(1)
		defer p.wg.Done()
		p.DataCh <- d
		return d.ID, nil
	}
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
		go func(wg *sync.WaitGroup) {
			wg.Add(1)
			defer wg.Done()
			for d := range p.DataCh {
				_, err := p.publish(d)
				if err != nil {
					p.ErrorCh <- err
				}
			}
		}(&p.wg)

		go func(wg *sync.WaitGroup) {
			wg.Add(1)
			defer wg.Done()
			for err := range p.ErrorCh {
				if p.Bus.errhandles != nil {
					p.Bus.errhandles(err)
				}
			}
		}(&p.wg)
	}()
	return nil
}
