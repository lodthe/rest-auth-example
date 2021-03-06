package statstask

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

var ErrNotFound = errors.New("not found")

type Repository interface {
	Create(userID uuid.UUID) (*Task, error)
	Get(id uuid.UUID) (*Task, error)
	UpdateStatus(id uuid.UUID, oldStatus, newStatus Status) error
	SetResult(id uuid.UUID, result *Result) error
}

type Repo struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) *Repo {
	return &Repo{db: db}
}

func (r *Repo) Create(userID uuid.UUID) (*Task, error) {
	task := &Task{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UserID:    userID,
		Status:    StatusPending,
	}

	query := `INSERT INTO "tasks" (id, created_at, user_id, status) 
							VALUES (:id, :created_at, :user_id, :status)`
	_, err := r.db.NamedExec(query, task)
	if err != nil {
		return nil, errors.Wrap(err, "database error")
	}

	return task, nil
}

func (r *Repo) Get(id uuid.UUID) (*Task, error) {
	task := new(Task)
	err := r.db.Get(task, `SELECT * FROM "tasks" WHERE id = $1`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, errors.Wrap(err, "database error")
	}

	return task, nil
}

func (r *Repo) UpdateStatus(id uuid.UUID, oldStatus, newStatus Status) error {
	_, err := r.db.Exec(`UPDATE "tasks" SET status = $1 WHERE id = $2 AND status = $3`, newStatus, id, oldStatus)

	return err
}

func (r *Repo) SetResult(id uuid.UUID, result *Result) error {
	_, err := r.db.Exec(`UPDATE "tasks" SET result = $1 WHERE id = $2`, result, id)

	return err
}
