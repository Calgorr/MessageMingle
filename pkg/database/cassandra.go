package database

import (
	"context"
	"os"
	"sync"
	"therealbroker/internal/exporter"
	"therealbroker/pkg/broker"
	"therealbroker/pkg/snowflake"
	"time"

	"github.com/gocql/gocql"
	"go.opentelemetry.io/otel"
)

type cassandraDatabase struct {
	session *gocql.Session
	sync.RWMutex
}

var (
	contactPoints = os.Getenv("DATABASE_TYPE")
	keyspace      = "broker"
)

func NewCassandraDatabase() (Database, error) {
	cluster := gocql.NewCluster(contactPoints)
	cluster.Keyspace = keyspace
	session, err := cluster.CreateSession()
	if err != nil {
		return nil, err
	}
	return &cassandraDatabase{session: session}, nil
}

func (c *cassandraDatabase) SetMessageID(ctx context.Context, msg *broker.Message, subject string) {
	_, globalSpan := otel.Tracer(exporter.DefaultServiceName).Start(ctx, "SetMessageIDCassandra method")
	defer globalSpan.End()
	msg.ID = snowflake.GenerateSnowflake(ctx)
}

func (c *cassandraDatabase) SaveMessage(ctx context.Context, msg *broker.Message, subject string) (int, error) {
	_, globalSpan := otel.Tracer(exporter.DefaultServiceName).Start(ctx, "SaveMessageCassandra method")
	defer globalSpan.End()
	expirationDate := time.Now().Add(msg.Expiration)
	query := c.session.Query(
		insertQuery,
		msg.ID, subject, msg.Body, expirationDate,
	)
	if err := query.Exec(); err != nil {
		return 0, err
	}
	return msg.ID, nil
}

func (c *cassandraDatabase) FetchMessage(ctx context.Context, id int, subject string) (*broker.Message, error) {
	_, globalSpan := otel.Tracer(exporter.DefaultServiceName).Start(ctx, "FetchCassandra method")
	defer globalSpan.End()
	var body string
	var expiration time.Time
	query := c.session.Query(
		selectQuery,
		id,
	)
	if err := query.Scan(&body, &expiration); err != nil {
		if err == gocql.ErrNotFound {
			return nil, broker.ErrInvalidID
		}
		return nil, err
	}
	msg := &broker.Message{
		ID:         id,
		Body:       body,
		Expiration: 0,
		IsExpired:  time.Now().After(expiration),
	}

	if msg.IsExpired {
		return &broker.Message{}, broker.ErrExpiredID
	}

	return msg, nil
}
