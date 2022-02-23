package models

import "time"

type Message struct {
	Id        int
	TopicID   int
	Body      string
	CreatedAT time.Time
	ExpiredAt time.Time
	DeletedAt time.Time
}
