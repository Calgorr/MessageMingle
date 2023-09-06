package database

import (
	"context"
	"database/sql"
	"fmt"
	"therealbroker/internal/exporter"
	"therealbroker/pkg/broker"
	"time"

	_ "github.com/lib/pq"
	"go.opentelemetry.io/otel"
)

type postgresDatabase struct {
	db *sql.DB
}

const (
	host     = "localhost"
	port     = 8080
	user     = "postgres"
	password = "postgres"
	dbname   = "broker"
)

func NewPostgresDatabase() Database {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}
	err = db.Ping()
	if err != nil {
		panic(err)
	}
	return &postgresDatabase{db: db}
}

func (p *postgresDatabase) SaveMessage(ctx context.Context, msg *broker.Message, subject string) int {
	_, globalSpan := otel.Tracer(exporter.DefaultServiceName).Start(ctx, "SaveMessagePostgres method")
	defer globalSpan.End()
	expirationDate := time.Now().Add(msg.Expiration)
	var id int
	err := p.db.QueryRow(
		"INSERT INTO message_broker (subject, body, expiration, expirationduration) VALUES ($1, $2, $3, $4) RETURNING id",
		subject, msg.Body, expirationDate, msg.Expiration,
	).Scan(&id)
	if err != nil {
		panic(err)
	}
	return id
}

func (p *postgresDatabase) FetchMessage(ctx context.Context, id int, subject string) (*broker.Message, error) {
	_, globalSpan := otel.Tracer(exporter.DefaultServiceName).Start(ctx, "FetchMessagePostgres method")
	defer globalSpan.End()
	var body string
	var expiration time.Time
	var expirationDuration time.Duration
	err := p.db.QueryRow("SELECT body, expiration, expirationduration FROM message_broker WHERE id = $1 AND subject = $2", id, subject).Scan(&body, &expiration, &expirationDuration)
	if err != nil {
		if err == sql.ErrNoRows {
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
