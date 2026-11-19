package httpapi

import (
	"log"
	"net/http"

	"github.com/Nithin0620/bookCinema/internal/ticket"
)

func (s *Server) myTickets(w http.ResponseWriter, r *http.Request) {
	user, _ := userFromContext(r.Context())
	tickets, err := s.tickets.ListByUser(r.Context(), user.ID)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, tickets)
}

func (s *Server) ticketQR(w http.ResponseWriter, r *http.Request) {
	user, _ := userFromContext(r.Context())
	detail, err := s.tickets.GetOwned(r.Context(), r.PathValue("id"), user.ID)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	png, err := ticket.QRCode(detail.Code)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	w.Header().Set("Content-Type", "image/png")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(png); err != nil {
		log.Printf("write ticket qr: %v", err)
	}
}
