package restapi

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
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
	r.Get("/users", h.getUsers)
	r.Get("/users/myself", h.getMyself)
	r.Put("/users/myself", h.updateMyself)
}

type UserOutput struct {
	ID       uuid.UUID `json:"id"`
	Username string    `json:"username"`
	Avatar   *string   `json:"avatar,omitempty"`
	Sex      string    `json:"sex"`
	Email    string    `json:"email"`
}

func (h *usersHandler) convertUser(u *muser.User) UserOutput {
	return UserOutput{
		ID:       u.ID,
		Username: u.Username,
		Avatar:   u.Avatar,
		Sex:      u.Sex,
		Email:    u.Email,
	}
}

type GetMyselfOutput = UserOutput

func (h *usersHandler) getMyself(w http.ResponseWriter, r *http.Request) {
	user, ok := loadUserFromRequest(w, r)
	if !ok {
		return
	}

	writeResult(w, GetMyselfOutput(h.convertUser(user)))
}

type UpdateMyselfInput struct {
	Username string  `json:"username"`
	Avatar   *string `json:"avatar"`
	Sex      string  `json:"sex"`
	Email    string  `json:"email"`
}

type UpdateMyselfOutput = UserOutput

func (h *usersHandler) updateMyself(w http.ResponseWriter, r *http.Request) {
	user, ok := loadUserFromRequest(w, r)
	if !ok {
		return
	}

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

	user.Username = input.Username
	user.Avatar = input.Avatar
	user.Sex = input.Sex
	user.Email = input.Email

	err = h.userRepo.Update(user)
	if err != nil {
		zlog.Error().Err(err).Interface("user", user).Msg("failed to update user")
		writeError(w, err.Error(), http.StatusBadRequest)

		return
	}

	writeResult(w, UpdateMyselfOutput(h.convertUser(user)))
}

type GetUsersOutput struct {
	Users []UserOutput `json:"users"`
}

func (h *usersHandler) getUsers(w http.ResponseWriter, r *http.Request) {
	usernames := r.URL.Query().Get("usernames")
	if usernames == "" {
		writeError(w, "missed usernames query parameter", http.StatusBadRequest)
		return
	}

	list := strings.Split(usernames, ",")
	users, err := h.userRepo.FindByUsernames(list)
	if err != nil {
		zlog.Error().Err(err).Interface("list", list).Msg("failed to get by usernames")
		writeError(w, err.Error(), http.StatusBadRequest)

		return
	}

	output := GetUsersOutput{
		Users: make([]UserOutput, 0, len(users)),
	}
	for _, u := range users {
		output.Users = append(output.Users, h.convertUser(&u))
	}

	writeResult(w, output)
}
