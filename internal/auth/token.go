package auth

import (
	"time"

	"github.com/google/uuid"
)

type Type string

const (
	TypeRefreshToken Type = "REFRESH_TOKEN"
	TypeAccessToken  Type = "ACCESS_TOKEN"
)

type Token struct {
	ID uuid.UUID `db:"id"`

	Type Type `db:"type"`

	ParentID *uuid.UUID `db:"parent_id"`

	IssuedAt  time.Time  `db:"issued_at"`
	ExpiresAt *time.Time `db:"expires_at"`

	UserID uuid.UUID `db:"user_id"`
}

func (t *Token) IsExpired() bool {
	return t.ExpiresAt != nil && t.ExpiresAt.Before(time.Now())
}
