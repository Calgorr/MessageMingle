package database

import (
	"context"
	"therealbroker/pkg/broker"
)

type Database interface {
	SetMessageID(ctx context.Context, msg *broker.Message, subject string)
	SaveMessage(ctx context.Context, msg *broker.Message, subject string) int
	FetchMessage(ctx context.Context, id int, subject string) (*broker.Message, error)
}
