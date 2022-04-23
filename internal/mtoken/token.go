package mtoken

import (
	"time"

	"github.com/google/uuid"
)

type Token struct {
	ID uuid.UUID `db:"id"`

	CreatedAt time.Time  `db:"created_at"`
	ExpiresAt *time.Time `db:"expires_at"`

	UserID uuid.UUID `db:"user_id"`
}
