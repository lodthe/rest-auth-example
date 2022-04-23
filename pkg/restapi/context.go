package restapi

import (
	"context"
	"net/http"

	"github.com/lodthe/rest-auth-example/internal/muser"
	zlog "github.com/rs/zerolog/log"
)

const contextUserKey = "user"

func putUserIntoContext(ctx context.Context, user *muser.User) context.Context {
	return context.WithValue(ctx, contextUserKey, user)
}

func loadUserFromRequest(w http.ResponseWriter, r *http.Request) (*muser.User, bool) {
	user := r.Context().Value(contextUserKey)
	converted, ok := user.(*muser.User)
	if !ok {
		zlog.Error().Interface("context", r.Context()).Msg("failed to user from ctx")
		writeError(w, "internal error", http.StatusInternalServerError)
	}

	return converted, ok
}
