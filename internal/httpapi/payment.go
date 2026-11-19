package httpapi

import "net/http"

func (s *Server) newPuzzle(w http.ResponseWriter, r *http.Request) {
	challenge, err := s.payment.NewChallenge(r.Context())
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, challenge)
}
