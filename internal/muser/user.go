package muser

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID `db:"id"`
	CreatedAt time.Time `db:"created_at"`

	Username string  `db:"username"`
	Avatar   *string `db:"avatar"`
	Sex      string  `db:"sex"`
	Email    string  `db:"email"`
}

func New() *User {
	return &User{
		ID: uuid.New(),
	}
}
