package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Nithin0620/bookCinema/internal/auth"
	"github.com/Nithin0620/bookCinema/internal/booking"
	"github.com/Nithin0620/bookCinema/internal/catalog"
	"github.com/Nithin0620/bookCinema/internal/config"
	"github.com/Nithin0620/bookCinema/internal/httpapi"
	"github.com/Nithin0620/bookCinema/internal/notify"
	"github.com/Nithin0620/bookCinema/internal/payment"
	"github.com/Nithin0620/bookCinema/internal/platform/postgres"
	redisplatform "github.com/Nithin0620/bookCinema/internal/platform/redis"
	"github.com/Nithin0620/bookCinema/internal/ticket"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	db, err := postgres.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	rdb, err := redisplatform.New(ctx, cfg.RedisAddr)
	if err != nil {
		log.Fatal(err)
	}
	defer rdb.Close()

	authService := auth.NewService(auth.NewPostgresStore(db), cfg.SessionTTL, cfg.AdminEmails)
	catalogStore := catalog.NewPostgresStore(db)
	bookingService := booking.NewService(
		booking.NewRedisHoldStore(rdb, cfg.HoldTTL),
		booking.NewPostgresBookingStore(db),
		catalogStore,
		cfg.HoldTTL,
	)
	paymentService := payment.NewService(rdb, 5*time.Minute)
	ticketStore := ticket.NewPostgresStore(db)
	notificationStore := notify.NewPostgresStore(db)

	api := httpapi.New(
		authService,
		catalogStore,
		bookingService,
		paymentService,
		ticketStore,
		notificationStore,
		cfg,
	)
	httpServer := &http.Server{
		Addr:              cfg.Addr,
		Handler:           api.Routes(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	serverErrors := make(chan error, 1)
	go func() {
		log.Printf("bookCinema listening on %s", cfg.Addr)
		serverErrors <- httpServer.ListenAndServe()
	}()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		if !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
	case <-signals:
		log.Printf("shutting down")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("shutdown: %v", err)
	}
}
