package httpapi

import (
	"net/http"
	"strings"
	"time"
	"unicode"

	"github.com/Nithin0620/bookCinema/internal/catalog"
)

func (s *Server) createMovie(w http.ResponseWriter, r *http.Request) {
	var request struct {
		ID          string `json:"id"`
		Title       string `json:"title"`
		Description string `json:"description"`
		DurationMin int    `json:"duration_min"`
		PosterURL   string `json:"poster_url"`
	}
	if !decodeJSON(w, r, &request) {
		return
	}
	if strings.TrimSpace(request.Title) == "" {
		writeError(w, http.StatusBadRequest, "title_required")
		return
	}
	if request.ID == "" {
		request.ID = slugify(request.Title)
	}
	if request.ID == "" {
		writeError(w, http.StatusBadRequest, "invalid_id")
		return
	}

	movie, err := s.catalog.CreateMovie(r.Context(), catalog.Movie{
		ID:          request.ID,
		Title:       strings.TrimSpace(request.Title),
		Description: strings.TrimSpace(request.Description),
		DurationMin: request.DurationMin,
		PosterURL:   strings.TrimSpace(request.PosterURL),
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, movie)
}

func (s *Server) createShowtime(w http.ResponseWriter, r *http.Request) {
	var request struct {
		MovieID            string `json:"movie_id"`
		StartsAt           string `json:"starts_at"`
		ScreenName         string `json:"screen_name"`
		Rows               int    `json:"rows"`
		SeatsPerRow        int    `json:"seats_per_row"`
		PremiumRows        int    `json:"premium_rows"`
		PremiumPriceCents  int    `json:"premium_price_cents"`
		StandardPriceCents int    `json:"standard_price_cents"`
	}
	if !decodeJSON(w, r, &request) {
		return
	}

	startsAt, err := time.Parse(time.RFC3339, request.StartsAt)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_starts_at")
		return
	}
	if request.MovieID == "" || strings.TrimSpace(request.ScreenName) == "" {
		writeError(w, http.StatusBadRequest, "movie_and_screen_required")
		return
	}
	if request.Rows < 1 || request.Rows > 26 || request.SeatsPerRow < 1 || request.SeatsPerRow > 30 {
		writeError(w, http.StatusBadRequest, "invalid_layout")
		return
	}
	if request.PremiumRows < 1 || request.PremiumRows > request.Rows {
		writeError(w, http.StatusBadRequest, "invalid_premium_rows")
		return
	}
	if request.PremiumPriceCents < 0 || request.StandardPriceCents < 0 {
		writeError(w, http.StatusBadRequest, "invalid_price")
		return
	}

	if _, err := s.catalog.MovieByID(r.Context(), request.MovieID); err != nil {
		writeServiceError(w, err)
		return
	}

	showtime, err := s.catalog.CreateShowtime(r.Context(), catalog.Showtime{
		MovieID:            request.MovieID,
		StartsAt:           startsAt,
		ScreenName:         strings.TrimSpace(request.ScreenName),
		Rows:               request.Rows,
		SeatsPerRow:        request.SeatsPerRow,
		PremiumRows:        request.PremiumRows,
		PremiumPriceCents:  request.PremiumPriceCents,
		StandardPriceCents: request.StandardPriceCents,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, showtime)
}

func slugify(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	var builder strings.Builder
	for _, char := range value {
		switch {
		case char >= 'a' && char <= 'z', char >= '0' && char <= '9', char == '-':
			builder.WriteRune(char)
		case unicode.IsSpace(char):
			builder.WriteByte('-')
		}
	}
	return strings.Trim(builder.String(), "-")
}
