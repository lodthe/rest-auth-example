package restapi

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/lodthe/rest-auth-example/internal/muser"
	"github.com/lodthe/rest-auth-example/internal/statstask"
	"github.com/lodthe/rest-auth-example/internal/taskqueue"
	"github.com/pkg/errors"
	zlog "github.com/rs/zerolog/log"
)

type statsHandler struct {
	userRepo muser.Repository
	taskRepo statstask.Repository
	producer *taskqueue.Producer
}

func newStatsHandler(taskRepo statstask.Repository, userRepo muser.Repository, producer *taskqueue.Producer) *statsHandler {
	return &statsHandler{
		taskRepo: taskRepo,
		userRepo: userRepo,
		producer: producer,
	}
}

func (h *statsHandler) handle(r chi.Router) {
	r.Get("/stats/{username}", h.createStatsTask)
	r.Get("/stats/tasks/{id}", h.getTaskStatus)
}

type CreateStatsTaskOutput struct {
	ID uuid.UUID `json:"id"`
}

func (h *statsHandler) createStatsTask(w http.ResponseWriter, r *http.Request) {
	username := chi.URLParam(r, "username")
	if username == "" {
		writeError(w, "missed username", http.StatusBadRequest)
		return
	}

	user, err := h.userRepo.GetByUsername(username)
	if errors.Is(err, muser.ErrNotFound) {
		writeError(w, "unknown user", http.StatusBadRequest)
		return
	}
	if err != nil {
		zlog.Error().Err(err).Str("username", username).Msg("failed to get user by username")
		writeError(w, "internal error", http.StatusInternalServerError)

		return
	}

	task, err := h.taskRepo.Create(user.ID)
	if err != nil {
		zlog.Error().Err(err).Str("user_id", user.ID.String()).Msg("failed to create a task")
		writeError(w, "internal error", http.StatusInternalServerError)

		return
	}

	err = h.producer.Produce(taskqueue.Task{ID: task.ID})
	if err != nil {
		zlog.Error().Err(err).Str("id", task.ID.String()).Msg("failed to publish")
		writeError(w, "internal error", http.StatusInternalServerError)

		return
	}

	writeResult(w, CreateStatsTaskOutput{
		ID: task.ID,
	})
}

type GetTaskStatusOutput struct {
	ID          uuid.UUID `json:"id"`
	Status      string    `json:"status"`
	DocumentURL string    `json:"document_url,omitempty"`
}

func (h *statsHandler) getTaskStatus(w http.ResponseWriter, r *http.Request) {
	taskID := chi.URLParam(r, "id")
	if taskID == "" {
		writeError(w, "missed task id", http.StatusBadRequest)
		return
	}

	convertedID, err := uuid.Parse(taskID)
	if err != nil {
		writeError(w, errors.Wrap(err, "UUID conversion failed").Error(), http.StatusBadRequest)
		return
	}

	task, err := h.taskRepo.Get(convertedID)
	if errors.Is(err, statstask.ErrNotFound) {
		writeError(w, "unknown task", http.StatusBadRequest)
		return
	}
	if err != nil {
		zlog.Error().Err(err).Str("task_id", taskID).Msg("failed to get stats task")
		writeError(w, "internal error", http.StatusInternalServerError)

		return
	}

	output := GetTaskStatusOutput{
		ID:     task.ID,
		Status: task.Status.String(),
	}
	if task.Result != nil {
		output.DocumentURL = task.Result.URL
	}

	writeResult(w, output)
}
