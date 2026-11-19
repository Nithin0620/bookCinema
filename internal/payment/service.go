package payment

import (
	"context"
	"errors"
	"fmt"
	"math/rand/v2"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	goredis "github.com/redis/go-redis/v9"
)

var (
	ErrInvalidChallenge = errors.New("invalid payment challenge")
	ErrWrongAnswer      = errors.New("wrong puzzle answer")
)

var verifyScript = goredis.NewScript(`
local answer = redis.call('GET', KEYS[1])
if not answer then return 0 end
redis.call('DEL', KEYS[1])
redis.call('DEL', KEYS[2])
if tonumber(answer) == tonumber(ARGV[1]) then return 1 end
return 2
`)

type Challenge struct {
	ID        string    `json:"id"`
	Question  string    `json:"question"`
	ExpiresAt time.Time `json:"expires_at"`
}

type Service struct {
	rdb *goredis.Client
	ttl time.Duration
}

func NewService(rdb *goredis.Client, ttl time.Duration) *Service {
	return &Service{rdb: rdb, ttl: ttl}
}

func (s *Service) NewChallenge(ctx context.Context) (Challenge, error) {
	left := rand.IntN(9) + 1
	right := rand.IntN(9) + 1
	id := uuid.NewString()
	question := fmt.Sprintf("What is %d + %d?", left, right)

	pipe := s.rdb.Pipeline()
	pipe.Set(ctx, answerKey(id), strconv.Itoa(left+right), s.ttl)
	pipe.Set(ctx, questionKey(id), question, s.ttl)
	if _, err := pipe.Exec(ctx); err != nil {
		return Challenge{}, fmt.Errorf("store challenge: %w", err)
	}

	return Challenge{
		ID:        id,
		Question:  question,
		ExpiresAt: time.Now().Add(s.ttl),
	}, nil
}

func (s *Service) Verify(ctx context.Context, id, answer string) error {
	if id == "" {
		return ErrInvalidChallenge
	}

	result, err := verifyScript.Run(
		ctx,
		s.rdb,
		[]string{answerKey(id), questionKey(id)},
		strings.TrimSpace(answer),
	).Int()
	if err != nil {
		return fmt.Errorf("verify challenge: %w", err)
	}

	switch result {
	case 1:
		return nil
	case 2:
		return ErrWrongAnswer
	default:
		return ErrInvalidChallenge
	}
}

func answerKey(id string) string {
	return "puzzle:" + id
}

func questionKey(id string) string {
	return "puzzle:question:" + id
}
