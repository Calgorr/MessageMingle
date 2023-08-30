package database

import "therealbroker/pkg/broker"

type Database interface {
	SaveMessage(msg *broker.Message, subject string) int
	FetchMessage(id int, subject string) (*broker.Message, error)
}
