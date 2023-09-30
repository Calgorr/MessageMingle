package database

import (
	"context"
	"sync"
	"therealbroker/internal/exporter"
	"therealbroker/pkg/broker"
	"time"

	"go.opentelemetry.io/otel"
)

type in_memory struct {
	messages map[string]map[int]*broker.Message
	sync.RWMutex
}

func NewInMemory() (Database, error) {
	return &in_memory{
		messages: make(map[string]map[int]*broker.Message),
	}, nil
}

func (i *in_memory) SetMessageID(ctx context.Context, msg *broker.Message, subject string) {
	_, globalSpan := otel.Tracer(exporter.DefaultServiceName).Start(ctx, "SetMessageIDInMemory method")
	defer globalSpan.End()
	if i.messages[subject] == nil {
		i.Lock()
		i.messages[subject] = make(map[int]*broker.Message)
		i.Unlock()
	}
	msg.ID = len(i.messages[subject]) + 1
}

func (i *in_memory) SaveMessage(ctx context.Context, msg *broker.Message, subject string) (int, error) {
	_, globalSpan := otel.Tracer(exporter.DefaultServiceName).Start(ctx, "SaveMessageInMemory method")
	defer globalSpan.End()
	go func() {
		if msg.Expiration == 0 {
			return
		}
		<-time.After(msg.Expiration)
		msg.IsExpired = true
	}()
	i.Lock()
	i.messages[subject][msg.ID] = msg
	i.Unlock()
	return msg.ID, nil
}

func (i *in_memory) FetchMessage(ctx context.Context, id int, subject string) (*broker.Message, error) {
	_, globalSpan := otel.Tracer(exporter.DefaultServiceName).Start(ctx, "FetchmessgeInMemory method")
	defer globalSpan.End()
	msg := i.messages[subject][id]
	if msg.IsExpired {
		return &broker.Message{}, broker.ErrExpiredID
	}
	return msg, nil
}
