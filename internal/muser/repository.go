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
	FindByUsernames(usernames []string) ([]User, error)
	Update(user *User) error
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

func (r *Repo) Update(user *User) error {
	_, err := r.db.NamedExec(`UPDATE users 
									SET username = :username, avatar = :avatar, sex = :sex, email = :email 
									WHERE id = :id`, user)

	return err
}

func (r *Repo) FindByUsernames(usernames []string) ([]User, error) {
	query, args, err := sqlx.In(`SELECT * FROM users WHERE username in (?) ORDER BY id`, usernames)
	if err != nil {
		return nil, errors.Wrap(err, "failed to construct a query")
	}

	query = r.db.Rebind(query)
	rows, err := r.db.Queryx(query, args...)
	if err != nil {
		return nil, errors.Wrap(err, "database error")
	}

	users := make([]User, 0, len(usernames))
	var user User
	for rows.Next() {
		err := rows.StructScan(&user)
		if err != nil {
			return nil, errors.Wrap(err, "scan failed")
		}

		users = append(users, user)
	}

	return users, nil
}
