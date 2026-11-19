package httpapi

import (
	"context"
	"log"
	"net/http"
)

func (s *Server) myNotifications(w http.ResponseWriter, r *http.Request) {
	user, _ := userFromContext(r.Context())
	notifications, err := s.notifications.ListByUser(r.Context(), user.ID)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, notifications)
}

func (s *Server) markNotificationsRead(w http.ResponseWriter, r *http.Request) {
	user, _ := userFromContext(r.Context())
	if err := s.notifications.MarkAllRead(r.Context(), user.ID); err != nil {
		writeServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) recordNotification(ctx context.Context, userID, title, body string) {
	if err := s.notifications.Create(ctx, userID, title, body); err != nil {
		log.Printf("record notification: %v", err)
	}
}
