package database

import "therealbroker/pkg/broker"

type in_memory struct {
	messages map[string]map[int]broker.Message
}

func NewInMemory() Database {
	return &in_memory{
		messages: make(map[string]map[int]broker.Message),
	}
}

func (i *in_memory) SaveMessage(msg broker.Message, subject string) int {
	if i.messages[subject] == nil {
		i.messages[subject] = make(map[int]broker.Message)
	}
	i.messages[subject][len(i.messages[subject])+1] = msg
	msg.ID = len(i.messages[subject])
	return msg.ID
}

func (i *in_memory) FetchMessage(id int, subject string) (broker.Message, error) {
	msg := i.messages[subject][id]
	return msg, nil
}

func (i *in_memory) DeleteMessage(id int, subject string) error {
	delete(i.messages[subject], id)
	return nil
}