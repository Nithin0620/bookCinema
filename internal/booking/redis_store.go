package booking

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

const holdKeyPrefix = "hold:"

var (
	createHoldScript = goredis.NewScript(`
local ok = redis.call('SET', KEYS[1], ARGV[1], 'NX', 'EX', tonumber(ARGV[2]))
if not ok then return 0 end
redis.call('SET', KEYS[2], KEYS[1], 'EX', tonumber(ARGV[2]))
return 1
`)

	getHoldScript = goredis.NewScript(`
local seatKey = redis.call('GET', KEYS[1])
if not seatKey then return nil end
local payload = redis.call('GET', seatKey)
if not payload then
  redis.call('DEL', KEYS[1])
  return nil
end
return payload
`)

	deleteHoldScript = goredis.NewScript(`
local seatKey = redis.call('GET', KEYS[1])
if seatKey then redis.call('DEL', seatKey) end
return redis.call('DEL', KEYS[1])
`)
)

type RedisHoldStore struct {
	rdb     *goredis.Client
	holdTTL time.Duration
}

func NewRedisHoldStore(rdb *goredis.Client, holdTTL time.Duration) *RedisHoldStore {
	return &RedisHoldStore{rdb: rdb, holdTTL: holdTTL}
}

func (s *RedisHoldStore) Create(ctx context.Context, hold Hold) (Hold, error) {
	payload, err := json.Marshal(hold)
	if err != nil {
		return Hold{}, fmt.Errorf("marshal hold: %w", err)
	}

	seatKey := holdKey(hold.ShowtimeID, hold.SeatID)
	sessionKey := sessionKey(hold.ID)
	ttl := int(s.holdTTL.Seconds())

	created, err := createHoldScript.Run(ctx, s.rdb, []string{seatKey, sessionKey}, string(payload), ttl).Int()
	if err != nil {
		return Hold{}, fmt.Errorf("create hold: %w", err)
	}
	if created == 0 {
		return Hold{}, ErrSeatAlreadyBooked
	}

	hold.ExpiresAt = time.Now().Add(s.holdTTL)
	return hold, nil
}

func (s *RedisHoldStore) GetOwned(ctx context.Context, sessionID, userID string) (Hold, error) {
	payload, err := getHoldScript.Run(ctx, s.rdb, []string{sessionKey(sessionID)}).Text()
	if errors.Is(err, goredis.Nil) {
		return Hold{}, ErrInvalidSession
	}
	if err != nil {
		return Hold{}, fmt.Errorf("get hold: %w", err)
	}

	hold, err := parseHold([]byte(payload))
	if err != nil {
		return Hold{}, ErrInvalidSession
	}
	if hold.UserID != userID {
		return Hold{}, ErrNotOwner
	}
	return hold, nil
}

func (s *RedisHoldStore) DeleteOwned(ctx context.Context, sessionID, userID string) error {
	if _, err := s.GetOwned(ctx, sessionID, userID); err != nil {
		return err
	}
	if err := deleteHoldScript.Run(ctx, s.rdb, []string{sessionKey(sessionID)}).Err(); err != nil {
		return fmt.Errorf("delete hold: %w", err)
	}
	return nil
}

func (s *RedisHoldStore) DeleteManyOwned(ctx context.Context, sessionIDs []string, userID string) error {
	for _, id := range sessionIDs {
		if err := s.DeleteOwned(ctx, id, userID); err != nil {
			return err
		}
	}
	return nil
}

func (s *RedisHoldStore) ListByShowtime(ctx context.Context, showtimeID string) ([]Hold, error) {
	pattern := fmt.Sprintf("%s%s:*", holdKeyPrefix, showtimeID)
	iter := s.rdb.Scan(ctx, 0, pattern, 100).Iterator()

	holds := []Hold{}
	for iter.Next(ctx) {
		payload, err := s.rdb.Get(ctx, iter.Val()).Bytes()
		if err != nil {
			continue
		}
		hold, err := parseHold(payload)
		if err != nil {
			continue
		}
		holds = append(holds, hold)
	}
	if err := iter.Err(); err != nil {
		return nil, fmt.Errorf("scan holds: %w", err)
	}
	return holds, nil
}

func holdKey(showtimeID, seatID string) string {
	return fmt.Sprintf("%s%s:%s", holdKeyPrefix, showtimeID, seatID)
}

func sessionKey(sessionID string) string {
	return "session:" + sessionID
}

func parseHold(payload []byte) (Hold, error) {
	var hold Hold
	if err := json.Unmarshal(payload, &hold); err != nil {
		return Hold{}, err
	}
	return hold, nil
}

var _ HoldStore = (*RedisHoldStore)(nil)
