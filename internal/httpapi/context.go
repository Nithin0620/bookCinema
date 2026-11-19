package httpapi

import (
	"context"
	"net/http"

	"github.com/Nithin0620/bookCinema/internal/auth"
)

type contextKey int

const userContextKey contextKey = iota

const sessionCookie = "bc_session"

func contextWithUser(ctx context.Context, user auth.User) context.Context {
	return context.WithValue(ctx, userContextKey, user)
}

func userFromContext(ctx context.Context) (auth.User, bool) {
	user, ok := ctx.Value(userContextKey).(auth.User)
	return user, ok
}

func (s *Server) requireUser(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := userFromContext(r.Context()); !ok {
			writeError(w, http.StatusUnauthorized, "authentication_required")
			return
		}
		next(w, r)
	}
}

func (s *Server) requireAdmin(next http.HandlerFunc) http.HandlerFunc {
	return s.requireUser(func(w http.ResponseWriter, r *http.Request) {
		user, _ := userFromContext(r.Context())
		if !user.IsAdmin {
			writeError(w, http.StatusForbidden, "admin_only")
			return
		}
		next(w, r)
	})
}
