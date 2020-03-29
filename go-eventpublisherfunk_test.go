package eventpublisherfunk

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"sync"
	"testing"
)

func CallBackFunc(t *testing.T, d EventData) error {
	t.Log("got EventData d", ":", d)
	if d.Topic == "" {
		return errors.New("EventData no topic")
	}
	if d.ID == "" {
		return errors.New("EventData no ID")
	}
	if s, ok := d.Data.(string); !ok || len(s) <= 0 {
		return errors.New("EventData must be a string and len >= 0")
	}
	return nil
}

func TestNewPublisher(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		// TODO: Add test cases.
		{
			name: "new Publisher test",
			want: "*eventpublisherfunk.publisher",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewPublisher(); !reflect.DeepEqual(reflect.TypeOf(got).String(), tt.want) {
				t.Errorf("NewPublisher() = %v, want %v", reflect.TypeOf(got).String(), tt.want)
			} else {
				t.Log("got instance", ":", got)
			}
		})
	}
}

func TestNewAsyncPublisher(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		// TODO: Add test cases.
		{
			name: "new async publisher ",
			want: "*eventpublisherfunk.publisher",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewAsyncPublisher(); !reflect.DeepEqual(reflect.TypeOf(got).String(), tt.want) {
				t.Errorf("NewAsyncPublisher() = %v, want %v", reflect.TypeOf(got).String(), tt.want)
			} else {
				t.Logf("NewAsyncPublisher got %v", got)
			}
		})
	}
}

func Test_publisher_WithAnts(t *testing.T) {
	type fields struct {
		Bus     *Bus
		DataCh  DataChannel
		ErrorCh ErrorChannel
		async   bool
		ctx     context.Context
		cancel  context.CancelFunc
		wg      sync.WaitGroup
	}
	tests := []struct {
		name   string
		fields fields
		want   Publisher
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &publisher{
				Bus:     tt.fields.Bus,
				DataCh:  tt.fields.DataCh,
				ErrorCh: tt.fields.ErrorCh,
				async:   tt.fields.async,
				ctx:     tt.fields.ctx,
				cancel:  tt.fields.cancel,
				wg:      tt.fields.wg,
			}
			if got := p.WithAnts(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("publisher.WithAnts() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_publisher_CloseAsyncPublisher(t *testing.T) {
	type fields struct {
		Bus     *Bus
		DataCh  DataChannel
		ErrorCh ErrorChannel
		async   bool
		ctx     context.Context
		cancel  context.CancelFunc
		wg      sync.WaitGroup
	}
	tests := []struct {
		name   string
		fields fields
	}{
		// TODO: Add test cases.
		{
			name:   "Test closeAsyncPublisher",
			fields: fields{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewAsyncPublisher()
			if err := p.CloseAsyncPublisher(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

var handleFunc_Demo = func(d EventData) (string, error) {
	fmt.Println("called from handleFunc_Demo", ":", d.ID, ":", d.Data, ":", d.Topic)
	if d.ID != "" {
		return d.ID, nil
	} else {
		return "", errors.New("topic empty")
	}
}

func Test_publisher_RegisterTopicHandleFunc(t *testing.T) {
	type fields Publisher
	type args struct {
		topic string
		f     HandleFunc
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name:   "publisher register topic handleFunc",
			fields: NewPublisher(),
			args: args{
				topic: "test1",
				f:     handleFunc_Demo,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := tt.fields
			if err := p.RegisterTopicHandleFunc(tt.args.topic, tt.args.f); (err != nil) != tt.wantErr {
				t.Errorf("publisher.RegisterTopicHandleFunc() error = %v, wantErr %v", err, tt.wantErr)
			} else {
				bus := p.DumpBus()
				if f, ok := bus.handles.Load(tt.args.topic); ok != false {
					fmt.Println("return f", ":", reflect.ValueOf(f), ":", reflect.TypeOf(f))
					if reflect.TypeOf(f).String() != "eventpublisherfunk.HandleFunc" {
						t.Errorf("error stored in handler")
					}
				} else {
					t.Errorf("no handler stored for topic")
				}
			}
		})
	}
}

var errHandleFunc_Demo = func(e error) {
	fmt.Println(e)
}

func Test_publisher_RegisterErrorHandleFunc(t *testing.T) {
	type fields Publisher
	type args struct {
		f ErrorHandleFunc
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name:   "register error handler",
			fields: NewPublisher(),
			args: args{
				f: errHandleFunc_Demo,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := tt.fields
			if err := p.RegisterErrorHandleFunc(tt.args.f); (err != nil) != tt.wantErr {
				t.Errorf("publisher.RegisterErrorHandleFunc() error = %v, wantErr %v", err, tt.wantErr)
			} else {
				t.Log(reflect.TypeOf(p.DumpBus().errhandles).String())
			}
		})
	}
}

func Test_publisher_PublishEvent(t *testing.T) {
	type fields Publisher
	type args struct {
		d EventData
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name:   "publisher test 1",
			fields: NewPublisher(),
			args: args{d: EventData{
				ID:    "1",
				Data:  "publisher test 1",
				Topic: "test1",
			}},
			want:    "1",
			wantErr: false,
		},
		{
			name:   "publisher test 2",
			fields: NewPublisher(),
			args: args{d: EventData{
				ID:    "2",
				Data:  "publisher test 2",
				Topic: "test2",
			}},
			want:    "2",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := tt.fields
			t.Log(tt.args.d.Topic)
			p.RegisterTopicHandleFunc(string(tt.args.d.Topic), handleFunc_Demo)
			got, err := p.PublishEvent(tt.args.d)
			t.Log("got", ":", got, "err", ":", err)
			if (err != nil) != tt.wantErr {
				t.Errorf("publisher.PublishEvent() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("publisher.PublishEvent() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_publisher_PublishEventAsync(t *testing.T) {
	type fields Publisher
	type args struct {
		d EventData
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name:   "PublishAsync test 1",
			fields: NewAsyncPublisher(),
			args: args{d: EventData{
				ID:    "1",
				Data:  "PublishAsync test 1",
				Topic: "PublishAsyncTest1",
			}},
			want:    "1",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := tt.fields
			defer p.CloseAsyncPublisher()
			p.RegisterTopicHandleFunc(tt.args.d.Topic, handleFunc_Demo)
			got, err := p.PublishEventAsync(tt.args.d)
			if (err != nil) != tt.wantErr {
				t.Errorf("publisher.PublishEventAsync() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("publisher.PublishEventAsync() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCallBackFunc(t *testing.T) {
	type args struct {
		t *testing.T
		d EventData
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := CallBackFunc(tt.args.t, tt.args.d); (err != nil) != tt.wantErr {
				t.Errorf("CallBackFunc() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
