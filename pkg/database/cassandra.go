package database

import (
	"context"
	"therealbroker/internal/exporter"
	"therealbroker/pkg/broker"
	"time"

	"github.com/gocql/gocql"
	"go.opentelemetry.io/otel"
)

type cassandraDatabase struct {
	session *gocql.Session
}

const (
	contactPoints = "127.0.0.1"
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

func (c *cassandraDatabase) SaveMessage(msg *broker.Message, subject string) int {
	_, globalSpan := otel.Tracer(exporter.DefaultServiceName).Start(context.Background(), "SaveMessageCassandra method")
	defer globalSpan.End()
	expirationDate := time.Now().Add(msg.Expiration)
	query := c.session.Query(
		"INSERT INTO message_broker (id, subject, body, expiration, expirationduration) VALUES (?, ?, ?, ?, ?)",
		gocql.TimeUUID(), subject, msg.Body, expirationDate, msg.Expiration,
	)
	if err := query.Exec(); err != nil {
		panic(err)
	}
	return 0
}

func (c *cassandraDatabase) FetchMessage(id int, subject string) (*broker.Message, error) {
	_, globalSpan := otel.Tracer(exporter.DefaultServiceName).Start(context.Background(), "FetchCassandra method")
	defer globalSpan.End()
	var body string
	var expiration time.Time
	var expirationDuration time.Duration
	query := c.session.Query(
		"SELECT body, expiration, expirationduration FROM message_broker WHERE id = ? AND subject = ?",
		id, subject,
	)
	if err := query.Scan(&body, &expiration, &expirationDuration); err != nil {
		if err == gocql.ErrNotFound {
			return nil, broker.ErrInvalidID
		}
		return nil, err
	}

	msg := &broker.Message{
		ID:         id,
		Body:       body,
		Expiration: expirationDuration,
		IsExpired:  time.Now().After(expiration),
	}

	if msg.IsExpired {
		return &broker.Message{}, broker.ErrExpiredID
	}

	return msg, nil
}
