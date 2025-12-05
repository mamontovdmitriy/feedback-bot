package entity

import (
	"time"
)

type MessageUpdate struct {
	Id        int       `db:"id"`
	Message   string    `db:"message"`
	Processed bool      `db:"is_processed"`
	CreatedAt time.Time `db:"created_at"`
}
