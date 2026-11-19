package httpapi

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/Nithin0620/bookCinema/internal/auth"
	"github.com/Nithin0620/bookCinema/internal/booking"
	"github.com/Nithin0620/bookCinema/internal/catalog"
	"github.com/Nithin0620/bookCinema/internal/config"
	"github.com/Nithin0620/bookCinema/internal/notify"
	"github.com/Nithin0620/bookCinema/internal/payment"
	"github.com/Nithin0620/bookCinema/internal/ticket"
)

type Server struct {
	auth          *auth.Service
	catalog       catalog.Store
	booking       *booking.Service
	payment       *payment.Service
	tickets       ticket.Store
	notifications notify.Store
	cfg           config.Config
}

func New(
	authService *auth.Service,
	catalogStore catalog.Store,
	bookingService *booking.Service,
	paymentService *payment.Service,
	ticketStore ticket.Store,
	notificationStore notify.Store,
	cfg config.Config,
) *Server {
	return &Server{
		auth:          authService,
		catalog:       catalogStore,
		booking:       bookingService,
		payment:       paymentService,
		tickets:       ticketStore,
		notifications: notificationStore,
		cfg:           cfg,
	}
}

func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/healthz", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	mux.HandleFunc("POST /api/auth/register", s.register)
	mux.HandleFunc("POST /api/auth/login", s.login)
	mux.HandleFunc("POST /api/auth/logout", s.logout)
	mux.HandleFunc("GET /api/auth/me", s.me)

	mux.HandleFunc("GET /api/movies", s.listMovies)
	mux.HandleFunc("GET /api/movies/{id}", s.getMovie)
	mux.HandleFunc("GET /api/showtimes", s.listShowtimes)
	mux.HandleFunc("GET /api/showtimes/{id}", s.getShowtime)
	mux.HandleFunc("GET /api/showtimes/{id}/seats", s.listSeats)

	mux.HandleFunc("POST /api/showtimes/{id}/seats/{seatID}/hold", s.requireUser(s.holdSeat))
	mux.HandleFunc("GET /api/payment/challenge", s.requireUser(s.newPuzzle))
	mux.HandleFunc("POST /api/checkout", s.requireUser(s.checkout))
	mux.HandleFunc("DELETE /api/sessions/{id}", s.requireUser(s.releaseSession))
	mux.HandleFunc("GET /api/me/bookings", s.requireUser(s.myBookings))
	mux.HandleFunc("DELETE /api/bookings/{id}", s.requireUser(s.cancelBooking))
	mux.HandleFunc("GET /api/me/tickets", s.requireUser(s.myTickets))
	mux.HandleFunc("GET /api/tickets/{id}/qr.png", s.requireUser(s.ticketQR))
	mux.HandleFunc("GET /api/me/notifications", s.requireUser(s.myNotifications))
	mux.HandleFunc("POST /api/me/notifications/read", s.requireUser(s.markNotificationsRead))

	mux.HandleFunc("POST /api/admin/movies", s.requireAdmin(s.createMovie))
	mux.HandleFunc("POST /api/admin/showtimes", s.requireAdmin(s.createShowtime))

	mux.HandleFunc("/api/", func(w http.ResponseWriter, _ *http.Request) {
		writeError(w, http.StatusNotFound, "not_found")
	})

	mux.Handle("/", s.static())

	return s.withMiddleware(mux)
}

func (s *Server) static() http.Handler {
	root := filepath.Clean(s.cfg.StaticDir)
	fileServer := http.FileServer(http.Dir(root))
	indexPath := filepath.Join(root, "index.html")

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		if _, err := os.Stat(indexPath); err != nil {
			writeError(w, http.StatusServiceUnavailable, "frontend_not_built")
			return
		}

		requestPath := strings.TrimPrefix(r.URL.Path, "/")
		if requestPath == "" || requestPath == "index.html" {
			if requestPath == "index.html" {
				indexRequest := r.Clone(r.Context())
				indexRequest.URL.Path = "/"
				http.ServeFile(w, indexRequest, indexPath)
				return
			}
			http.ServeFile(w, r, indexPath)
			return
		}

		fullPath := filepath.Join(root, filepath.Clean(requestPath))
		if !strings.HasPrefix(fullPath, root+string(os.PathSeparator)) {
			writeError(w, http.StatusNotFound, "not_found")
			return
		}
		if info, err := os.Stat(fullPath); err == nil && !info.IsDir() {
			fileServer.ServeHTTP(w, r)
			return
		}

		http.ServeFile(w, r, indexPath)
	})
}
