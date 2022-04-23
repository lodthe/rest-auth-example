package auth

import (
	"database/sql"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

var ErrNotFound = errors.New("not found")

type Repository interface {
	Create(token *Token) error
	Get(id uuid.UUID) (*Token, error)
}

type Repo struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) *Repo {
	return &Repo{db: db}
}

func (r *Repo) Create(token *Token) error {
	query := `INSERT INTO "tokens" (id, type, parent_id, issued_at, expires_at, user_id) 
							VALUES (:id, :type, :parent_id, :issued_at, :expires_at, :user_id)`
	_, err := r.db.NamedExec(query, token)
	if err != nil {
		return errors.Wrap(err, "database error")
	}

	return nil
}

func (r *Repo) Get(id uuid.UUID) (*Token, error) {
	task := new(Token)
	err := r.db.Get(task, `SELECT * FROM "tokens" WHERE id = $1`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, errors.Wrap(err, "database error")
	}

	return task, nil
}
