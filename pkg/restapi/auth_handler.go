package restapi

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/lodthe/rest-auth-example/internal/auth"
	"github.com/lodthe/rest-auth-example/internal/muser"
	"github.com/pkg/errors"
	zlog "github.com/rs/zerolog/log"
)

type authHandler struct {
	auth     *auth.Service
	userRepo muser.Repository
}

func newAuthHandler(authService *auth.Service, userRepo muser.Repository) *authHandler {
	return &authHandler{
		auth:     authService,
		userRepo: userRepo,
	}
}

func (h *authHandler) handle(r chi.Router) {
	r.Post("/auth/register", h.register)
	r.Post("/auth/issue-access-token", h.issueAccessToken)
}

type RegisterInput struct {
	Username string  `json:"username"`
	Avatar   *string `json:"avatar"`
	Sex      string  `json:"sex"`
	Email    string  `json:"email"`
}

type RegisterOutput struct {
	RefreshToken string `json:"refresh_token"`
}

func (h *authHandler) register(w http.ResponseWriter, r *http.Request) {
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

	_, token, err := h.auth.IssueRefreshToken(user.ID)
	if err != nil {
		zlog.Error().Err(err).Interface("user", user).Msg("failed to issue a new refresh token")
		writeError(w, "internal error", http.StatusInternalServerError)

		return
	}

	writeResult(w, RegisterOutput{
		RefreshToken: token,
	})
}

type IssueAccessToken struct {
	AccessToken string `json:"access_token"`
}

func (h *authHandler) issueAccessToken(w http.ResponseWriter, r *http.Request) {
	header := r.Header.Get("Authorization")
	refreshToken, err := h.auth.FetchToken(strings.TrimPrefix(header, "Bearer "))
	if err != nil {
		if errors.Is(err, auth.ErrUnauthorized) || errors.Is(err, auth.ErrInvalidToken) {
			writeError(w, "invalid token", http.StatusBadRequest)
		} else {
			zlog.Error().Err(err).Msg("token fetch failed")
			writeError(w, "internal error", http.StatusInternalServerError)
		}

		return
	}
	if refreshToken.IsExpired() {
		writeError(w, "token has expired", http.StatusBadRequest)
		return
	}

	_, accessToken, err := h.auth.IssueAccessToken(refreshToken.ID, refreshToken.UserID)
	if err != nil {
		zlog.Error().Err(err).Interface("refresh_token", refreshToken).Msg("failed to issue an access token")
		writeError(w, "token cannot be issued", http.StatusInternalServerError)

		return
	}

	zlog.Info().Str("user_id", refreshToken.UserID.String()).Msg("a new access token has been issued")

	writeResult(w, IssueAccessToken{
		AccessToken: accessToken,
	})
}
