package message

import "time"

type Message struct {
	id         int
	Body       string
	Expiration time.Duration
}

func (message *Message) GetId() int {
	return message.id
}

func (message *Message) SetId(id int) {
	message.id = id
}
