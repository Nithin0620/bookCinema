//go:build integration

package payment

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"testing"
	"time"

	redisplatform "github.com/Nithin0620/bookCinema/internal/platform/redis"
)

func TestIntegrationChallengeVerify(t *testing.T) {
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		t.Skip("set REDIS_ADDR to run integration tests")
	}

	ctx := context.Background()
	rdb, err := redisplatform.New(ctx, redisAddr)
	if err != nil {
		t.Fatalf("connect redis: %v", err)
	}
	defer rdb.Close()

	service := NewService(rdb, time.Minute)
	challenge, err := service.NewChallenge(ctx)
	if err != nil {
		t.Fatalf("new challenge: %v", err)
	}

	var left, right int
	if _, err := fmt.Sscanf(challenge.Question, "What is %d + %d?", &left, &right); err != nil {
		t.Fatalf("parse question %q: %v", challenge.Question, err)
	}

	if err := service.Verify(ctx, challenge.ID, strconv.Itoa(left+right+1)); !errors.Is(err, ErrWrongAnswer) {
		t.Fatalf("wrong answer: got %v want %v", err, ErrWrongAnswer)
	}
	if err := service.Verify(ctx, challenge.ID, strconv.Itoa(left+right)); !errors.Is(err, ErrInvalidChallenge) {
		t.Fatalf("reused challenge: got %v want %v", err, ErrInvalidChallenge)
	}

	second, err := service.NewChallenge(ctx)
	if err != nil {
		t.Fatalf("second challenge: %v", err)
	}
	if _, err := fmt.Sscanf(second.Question, "What is %d + %d?", &left, &right); err != nil {
		t.Fatalf("parse second question %q: %v", second.Question, err)
	}
	if err := service.Verify(ctx, second.ID, fmt.Sprintf(" %d ", left+right)); err != nil {
		t.Fatalf("correct answer: %v", err)
	}
}
