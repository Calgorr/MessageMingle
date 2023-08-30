package broker

import (
	"context"
	"fmt"
	"therealbroker/pkg/broker"
)

type Module struct {
	subscribers map[string][]chan broker.Message
}

func NewModule() broker.Broker {
	return &Module{
		subscribers: make(map[string][]chan broker.Message),
	}
}

func (m *Module) Close() error {
	for _, v := range m.subscribers {
		for _, ch := range v {
			close(ch)
		}
	}
	fmt.Println(m.subscribers)
	m = nil
	return nil
}

func (m *Module) Publish(ctx context.Context, subject string, msg broker.Message) (int, error) {
	panic("implement me")
}

func (m *Module) Subscribe(ctx context.Context, subject string) (<-chan broker.Message, error) {
	panic("implement me")
}

func (m *Module) Fetch(ctx context.Context, subject string, id int) (broker.Message, error) {
	panic("implement me")
}
