package restapi

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/lodthe/rest-auth-example/internal/muser"
	zlog "github.com/rs/zerolog/log"
)

type usersHandler struct {
	userRepo muser.Repository
}

func newUsersHandler(userRepo muser.Repository) *usersHandler {
	return &usersHandler{
		userRepo: userRepo,
	}
}

func (h *usersHandler) handle(r chi.Router) {
	r.Post("/users/register", h.register)
}

type RegisterInput struct {
	Username string  `json:"username"`
	Avatar   *string `json:"*avatar"`
	Sex      string  `json:"sex"`
	Email    string  `json:"email"`
}

type RegisterOutput struct {
	RefreshToken string `json:"refresh_token"`
}

func (h *usersHandler) register(w http.ResponseWriter, r *http.Request) {
	var input RegisterInput
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		writeError(w, err.Error(), http.StatusBadRequest)
		return
	}

	if input.Username == "" {
		writeError(w, "missed username", http.StatusBadRequest)
		return
	}
	if input.Sex == "" {
		writeError(w, "missed sex", http.StatusBadRequest)
		return
	}
	if input.Email == "" {
		writeError(w, "missed email", http.StatusBadRequest)
		return
	}

	user := muser.New()
	user.Username = input.Username
	user.Avatar = input.Avatar
	user.Sex = input.Sex
	user.Email = input.Email

	err = h.userRepo.Create(user)
	if err != nil {
		zlog.Error().Err(err).Interface("user", user).Msg("failed to create a user")
		writeError(w, err.Error(), http.StatusBadRequest)

		return
	}

	writeResult(w, RegisterOutput{
		RefreshToken: "NOT IMPLEMENTED",
	})
}
