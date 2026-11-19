package httpapi

import (
	"fmt"
	"log"
	"net/http"
)

func (s *Server) holdSeat(w http.ResponseWriter, r *http.Request) {
	user, _ := userFromContext(r.Context())
	hold, err := s.booking.HoldSeat(
		r.Context(),
		r.PathValue("id"),
		r.PathValue("seatID"),
		user.ID,
	)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, hold)
}

func (s *Server) checkout(w http.ResponseWriter, r *http.Request) {
	var request struct {
		SessionIDs []string `json:"session_ids"`
		Payment    struct {
			ID     string `json:"id"`
			Answer string `json:"answer"`
		} `json:"payment"`
	}
	if !decodeJSON(w, r, &request) {
		return
	}

	user, _ := userFromContext(r.Context())
	if err := s.payment.Verify(r.Context(), request.Payment.ID, request.Payment.Answer); err != nil {
		writeServiceError(w, err)
		return
	}

	bookings, err := s.booking.Checkout(r.Context(), user.ID, request.SessionIDs)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	if _, err := s.tickets.CreateForBookings(r.Context(), bookings); err != nil {
		log.Printf("create tickets: %v", err)
	}
	s.recordNotification(
		r.Context(),
		user.ID,
		"Booking confirmed",
		fmt.Sprintf("%d seat(s) confirmed for showtime %s.", len(bookings), bookings[0].ShowtimeID),
	)

	writeJSON(w, http.StatusCreated, map[string]any{"bookings": bookings})
}

func (s *Server) releaseSession(w http.ResponseWriter, r *http.Request) {
	user, _ := userFromContext(r.Context())
	if err := s.booking.Release(r.Context(), r.PathValue("id"), user.ID); err != nil {
		writeServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) myBookings(w http.ResponseWriter, r *http.Request) {
	user, _ := userFromContext(r.Context())
	bookings, err := s.booking.ListUserBookings(r.Context(), user.ID)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, bookings)
}

func (s *Server) cancelBooking(w http.ResponseWriter, r *http.Request) {
	user, _ := userFromContext(r.Context())
	bookingID := r.PathValue("id")
	if err := s.booking.CancelBooking(r.Context(), bookingID, user.ID); err != nil {
		writeServiceError(w, err)
		return
	}
	s.recordNotification(r.Context(), user.ID, "Booking cancelled", "A confirmed booking was cancelled.")
	w.WriteHeader(http.StatusNoContent)
}
