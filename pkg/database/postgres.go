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
	host     = "postgres"
	port     = 5432
	user     = "postgres"
	password = "postgres"
	dbname   = "broker"
)

func NewPostgresDatabase() Database {
	createDatabase()
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
	db, err := sql.Open("postgres", psqlInfo)
	db.SetMaxOpenConns(90)
	db.SetMaxIdleConns(45)
	if err != nil {
		panic(err)
	}
	err = db.Ping()
	if err != nil {
		panic(err)
	}
	postgresMigration(db)
	return &postgresDatabase{db: db}
}

func createDatabase() {
	db, err := sql.Open("postgres", fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s sslmode=disable",
		host, port, user, password))
	if err != nil {
		panic(err)
	}
	defer db.Close()

	_, err = db.Exec(fmt.Sprintf("CREATE DATABASE %s", dbname))
	if err != nil {
		panic(err)
	}
}

func postgresMigration(db *sql.DB) {
	_, err := db.Exec(`
        CREATE TABLE IF NOT EXISTS message_broker (
            id BIGSERIAL PRIMARY KEY,
            subject TEXT,
            body TEXT,
            expiration TIMESTAMPTZ
        )
    `)
	if err != nil {
		panic(err)
	}
}

func (p *postgresDatabase) SetMessageID(ctx context.Context, msg *broker.Message, subject string) {

}

func (p *postgresDatabase) SaveMessage(ctx context.Context, msg *broker.Message, subject string) int {
	_, globalSpan := otel.Tracer(exporter.DefaultServiceName).Start(ctx, "SaveMessagePostgres method")
	defer globalSpan.End()
	expirationDate := time.Now().Add(msg.Expiration)
	err := p.db.QueryRow(
		"INSERT INTO message_broker (subject, body, expiration) VALUES ($1, $2, $3) RETURNING ID",
		subject, msg.Body, expirationDate,
	).Scan(&msg.ID)
	if err != nil {
		panic(err)
	}
	return msg.ID
}

func (p *postgresDatabase) FetchMessage(ctx context.Context, id int, subject string) (*broker.Message, error) {
	_, globalSpan := otel.Tracer(exporter.DefaultServiceName).Start(ctx, "FetchMessagePostgres method")
	defer globalSpan.End()
	var body string
	var expiration time.Time
	err := p.db.QueryRow("SELECT body, expiration FROM message_broker WHERE id = $1 AND subject = $2", id, subject).Scan(&body, &expiration)
	if err != nil {
		if err == sql.ErrNoRows {
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
