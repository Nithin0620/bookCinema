package httpapi

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"

	"github.com/Nithin0620/bookCinema/internal/auth"
	"github.com/Nithin0620/bookCinema/internal/booking"
	"github.com/Nithin0620/bookCinema/internal/catalog"
	"github.com/Nithin0620/bookCinema/internal/payment"
	"github.com/Nithin0620/bookCinema/internal/ticket"
)

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Printf("write response: %v", err)
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) bool {
	decoder := json.NewDecoder(io.LimitReader(r.Body, 1<<20))
	if err := decoder.Decode(dst); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json")
		return false
	}
	return true
}

func writeServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, auth.ErrEmailTaken):
		writeError(w, http.StatusConflict, "email_taken")
	case errors.Is(err, auth.ErrInvalidEmail):
		writeError(w, http.StatusBadRequest, "invalid_email")
	case errors.Is(err, auth.ErrWeakPassword):
		writeError(w, http.StatusBadRequest, "weak_password")
	case errors.Is(err, auth.ErrInvalidCredentials):
		writeError(w, http.StatusUnauthorized, "invalid_credentials")
	case errors.Is(err, auth.ErrInvalidToken):
		writeError(w, http.StatusUnauthorized, "invalid_session")
	case errors.Is(err, booking.ErrSeatAlreadyBooked):
		writeError(w, http.StatusConflict, "seat_taken")
	case errors.Is(err, booking.ErrSeatNotFound):
		writeError(w, http.StatusNotFound, "seat_not_found")
	case errors.Is(err, booking.ErrShowtimeNotFound):
		writeError(w, http.StatusNotFound, "showtime_not_found")
	case errors.Is(err, booking.ErrInvalidSession):
		writeError(w, http.StatusNotFound, "session_not_found")
	case errors.Is(err, booking.ErrNotOwner):
		writeError(w, http.StatusForbidden, "not_owner")
	case errors.Is(err, booking.ErrEmptyCheckout):
		writeError(w, http.StatusBadRequest, "empty_checkout")
	case errors.Is(err, booking.ErrBookingNotFound):
		writeError(w, http.StatusNotFound, "booking_not_found")
	case errors.Is(err, catalog.ErrNotFound):
		writeError(w, http.StatusNotFound, "not_found")
	case errors.Is(err, catalog.ErrConflict):
		writeError(w, http.StatusConflict, "already_exists")
	case errors.Is(err, payment.ErrInvalidChallenge):
		writeError(w, http.StatusBadRequest, "invalid_challenge")
	case errors.Is(err, payment.ErrWrongAnswer):
		writeError(w, http.StatusBadRequest, "wrong_answer")
	case errors.Is(err, ticket.ErrNotFound):
		writeError(w, http.StatusNotFound, "ticket_not_found")
	default:
		log.Printf("request failed: %v", err)
		writeError(w, http.StatusInternalServerError, "internal_error")
	}
}
