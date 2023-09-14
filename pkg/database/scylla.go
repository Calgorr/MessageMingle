package database

import (
	"context"
	"log"

	"sync"
	"therealbroker/internal/exporter"
	"therealbroker/pkg/broker"
	"therealbroker/pkg/snowflake"
	"time"

	"github.com/gocql/gocql"
	"go.opentelemetry.io/otel"
)

type scyllaDatabase struct {
	session *gocql.Session
	sync.RWMutex
}

const (
	insertQuery = "INSERT INTO broker.message_broker (id, subject, body, expiration) VALUES (?, ?, ?, ?)"
	selectQuery = "SELECT body, expiration FROM broker.message_broker WHERE id = ? ALLOW FILTERING"
)

func NewScyllaDatabase() Database {
	cluster := gocql.NewCluster(contactPoints)
	session, err := cluster.CreateSession()
	if err != nil {
		log.Fatal(err)
	}
	scyllaMigration(session)
	cluster.Keyspace = keyspace
	return &scyllaDatabase{session: session}
}

func scyllaMigration(session *gocql.Session) {
	if err := session.Query(
		"CREATE KEYSPACE IF NOT EXISTS broker WITH replication = { 'class': 'SimpleStrategy', 'replication_factor': 1 };",
	).Exec(); err != nil {
		panic(err)
	}
	if err := session.Query(
		"CREATE TABLE IF NOT EXISTS broker.message_broker ( id bigint PRIMARY KEY, body text, expiration timestamp, subject text )",
	).Exec(); err != nil {
		panic(err)
	}
}

func (s *scyllaDatabase) SetMessageID(ctx context.Context, msg *broker.Message, subject string) {
	_, globalSpan := otel.Tracer(exporter.DefaultServiceName).Start(ctx, "SetMessageIDCassandra method")
	defer globalSpan.End()
	msg.ID = snowflake.GenerateSnowflake(ctx)
}

func (c *scyllaDatabase) SaveMessage(ctx context.Context, msg *broker.Message, subject string) int {
	_, globalSpan := otel.Tracer(exporter.DefaultServiceName).Start(ctx, "SaveMessageCassandra method")
	defer globalSpan.End()
	expirationDate := time.Now().Add(msg.Expiration)
	query := c.session.Query(
		insertQuery,
		msg.ID, subject, msg.Body, expirationDate,
	)
	if err := query.Exec(); err != nil {
		panic(err)
	}
	return msg.ID
}

func (c *scyllaDatabase) FetchMessage(ctx context.Context, id int, subject string) (*broker.Message, error) {
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
