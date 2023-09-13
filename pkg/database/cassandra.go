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

func NewCassandraDatabase() Database {
	cluster := gocql.NewCluster(contactPoints)
	cluster.Keyspace = keyspace
	session, err := cluster.CreateSession()
	if err != nil {
		panic(err)
	}
	return &cassandraDatabase{session: session}
}

func (c *cassandraDatabase) SetMessageID(ctx context.Context, msg *broker.Message, subject string) {
	_, globalSpan := otel.Tracer(exporter.DefaultServiceName).Start(ctx, "SetMessageIDCassandra method")
	defer globalSpan.End()
	msg.ID = snowflake.GenerateSnowflake(ctx)
}

func (c *cassandraDatabase) SaveMessage(ctx context.Context, msg *broker.Message, subject string) int {
	_, globalSpan := otel.Tracer(exporter.DefaultServiceName).Start(ctx, "SaveMessageCassandra method")
	defer globalSpan.End()
	expirationDate := time.Now().Add(msg.Expiration)
	query := c.session.Query(
		"INSERT INTO message_broker (id, subject, body, expiration) VALUES (?, ?, ?, ?)",
		msg.ID, subject, msg.Body, expirationDate,
	)
	if err := query.Exec(); err != nil {
		panic(err)
	}
	return msg.ID
}

func (c *cassandraDatabase) FetchMessage(ctx context.Context, id int, subject string) (*broker.Message, error) {
	_, globalSpan := otel.Tracer(exporter.DefaultServiceName).Start(ctx, "FetchCassandra method")
	defer globalSpan.End()
	var body string
	var expiration time.Time
	query := c.session.Query(
		"SELECT body, expiration FROM message_broker WHERE id = ? ALLOW FILTERING",
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
