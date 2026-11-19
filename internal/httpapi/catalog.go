package httpapi

import (
	"net/http"
)

func (s *Server) listMovies(w http.ResponseWriter, r *http.Request) {
	movies, err := s.catalog.ListMovies(r.Context())
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, movies)
}

func (s *Server) getMovie(w http.ResponseWriter, r *http.Request) {
	movie, err := s.catalog.MovieByID(r.Context(), r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, movie)
}

func (s *Server) listShowtimes(w http.ResponseWriter, r *http.Request) {
	showtimes, err := s.catalog.ListShowtimes(r.Context(), r.URL.Query().Get("movie_id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, showtimes)
}

func (s *Server) getShowtime(w http.ResponseWriter, r *http.Request) {
	showtime, err := s.catalog.ShowtimeByID(r.Context(), r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, showtime)
}

func (s *Server) listSeats(w http.ResponseWriter, r *http.Request) {
	showtimeID := r.PathValue("id")
	showtime, err := s.catalog.ShowtimeByID(r.Context(), showtimeID)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	userID := ""
	if user, ok := userFromContext(r.Context()); ok {
		userID = user.ID
	}

	seats, err := s.booking.ListSeats(r.Context(), showtimeID, userID)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"showtime": showtime,
		"seats":    seats,
	})
}
