package broker

import (
	"context"
	"sync"
	"therealbroker/internal/exporter"
	"therealbroker/pkg/broker"
	"therealbroker/pkg/database"
	"time"

	"go.opentelemetry.io/otel"
)

var mainService *Module

type Module struct {
	isClosed      bool
	subscribers   map[string][]chan broker.Message
	ListenersLock sync.Mutex
	db            database.Database
}

func NewModule() broker.Broker {
	if mainService == nil {
		mainService := &Module{
			subscribers: make(map[string][]chan broker.Message),
			db:          database.NewInMemory(),
		}
		return mainService
	}
	mainService.isClosed = false
	return mainService
}

func (m *Module) Close() error {
	for _, v := range m.subscribers {
		for _, ch := range v {
			close(ch)
		}
	}
	m.isClosed = true
	return nil
}

func (m *Module) Publish(ctx context.Context, subject string, msg broker.Message) (int, error) {
	_, globalSpan := otel.Tracer(exporter.DefaultServiceName).Start(ctx, "Publish broker method")
	defer globalSpan.End()
	m.ListenersLock.Lock()
	defer m.ListenersLock.Unlock()
	if m.isClosed {
		return 0, broker.ErrUnavailable
	}
	var wg sync.WaitGroup
	for _, listener := range m.subscribers[subject] {
		wg.Add(1)
		go func(listener chan broker.Message) {
			defer wg.Done()
			listener <- msg
		}(listener)
	}
	wg.Wait()
	msg.ID = m.db.SaveMessage(&msg, subject)
	go func() {
		if msg.Expiration == 0 {
			return
		}
		<-time.After(msg.Expiration)
		msg.IsExpired = true
	}()
	return msg.ID, nil
}

func (m *Module) Subscribe(ctx context.Context, subject string) (<-chan broker.Message, error) {
	_, globalSpan := otel.Tracer(exporter.DefaultServiceName).Start(ctx, "Subscribe broker method")
	defer globalSpan.End()
	m.ListenersLock.Lock()
	defer m.ListenersLock.Unlock()
	if m.isClosed {
		return nil, broker.ErrUnavailable
	}
	select {
	case <-ctx.Done():
		return nil, broker.ErrCancelled
	default:
		ch := make(chan broker.Message, 100000)
		m.subscribers[subject] = append(m.subscribers[subject], ch)
		return ch, nil
	}

}

func (m *Module) Fetch(ctx context.Context, subject string, id int) (broker.Message, error) {
	_, globalSpan := otel.Tracer(exporter.DefaultServiceName).Start(ctx, "Fetch broker method")
	defer globalSpan.End()
	m.ListenersLock.Lock()
	defer m.ListenersLock.Unlock()
	if m.isClosed {
		return broker.Message{}, broker.ErrUnavailable
	}
	msg, err := m.db.FetchMessage(id, subject)
	if err != nil {
		return broker.Message{}, err
	}
	if msg.IsExpired {
		return broker.Message{}, broker.ErrExpiredID
	}
	return *msg, nil
}
