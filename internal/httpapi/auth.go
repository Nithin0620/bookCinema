package httpapi

import (
	"net/http"
)

func (s *Server) register(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if !decodeJSON(w, r, &request) {
		return
	}

	user, token, err := s.auth.Register(r.Context(), request.Email, request.Password)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	s.setSessionCookie(w, token)
	writeJSON(w, http.StatusCreated, user)
}

func (s *Server) login(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if !decodeJSON(w, r, &request) {
		return
	}

	user, token, err := s.auth.Login(r.Context(), request.Email, request.Password)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	s.setSessionCookie(w, token)
	writeJSON(w, http.StatusOK, user)
}

func (s *Server) logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(sessionCookie)
	if err == nil {
		if err := s.auth.Logout(r.Context(), cookie.Value); err != nil {
			writeServiceError(w, err)
			return
		}
	}

	s.clearSessionCookie(w)
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) me(w http.ResponseWriter, r *http.Request) {
	user, ok := userFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "authentication_required")
		return
	}
	writeJSON(w, http.StatusOK, user)
}
