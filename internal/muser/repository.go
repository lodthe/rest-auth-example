package muser

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

var ErrNotFound = errors.New("not found")

type Repository interface {
	Create(user *User) error
	Get(id uuid.UUID) (*User, error)
	GetByUsername(username string) (*User, error)
}

type Repo struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) *Repo {
	return &Repo{db: db}
}

func (r *Repo) Create(user *User) error {
	user.CreatedAt = time.Now()

	query := `INSERT INTO "users" (id, created_at, username, avatar, sex, email) 
							VALUES (:id, :created_at, :username, :avatar, :sex, :email)`
	_, err := r.db.NamedExec(query, user)
	if err != nil {
		return errors.Wrap(err, "database error")
	}

	return nil
}

func (r *Repo) Get(id uuid.UUID) (*User, error) {
	task := new(User)
	err := r.db.Get(task, `SELECT * FROM "users" WHERE id = $1`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, errors.Wrap(err, "database error")
	}

	return task, nil
}

func (r *Repo) GetByUsername(username string) (*User, error) {
	task := new(User)
	err := r.db.Get(task, `SELECT * FROM "users" WHERE username = $1`, username)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, errors.Wrap(err, "database error")
	}

	return task, nil
}
