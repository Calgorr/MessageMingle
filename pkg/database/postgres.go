package database

import (
	"database/sql"
	"fmt"
	"therealbroker/pkg/broker"
	"time"

	_ "github.com/lib/pq"
)

type postgresDatabase struct {
	db *sql.DB
}

const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "postgres"
	dbname   = "broker"
)

func NewPostgresDatabase() (Database, error) {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return nil, err
	}
	err = db.Ping()
	if err != nil {
		return nil, err
	}
	return &postgresDatabase{db: db}, nil
}

func (p *postgresDatabase) SaveMessage(msg *broker.Message, subject string) int {
	expirationDate := time.Now().Add(msg.Expiration)
	result, err := p.db.Exec(
		"INSERT INTO message_broker (subject, body, expiration, expirationduration) VALUES ($1, $2, $3, $4)",
		subject, msg.Body, expirationDate, msg.Expiration,
	)
	if err != nil {
		panic(err)
	}
	id, _ := result.LastInsertId()
	return int(id)
}

func (p *postgresDatabase) FetchMessage(id int, subject string) (*broker.Message, error) {
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
