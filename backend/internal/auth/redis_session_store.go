package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const sessionKeyPrefix = "session:"

// RedisSessionStore implementa SessionStore sobre Redis. Encaja con la
// regla de arquitectura: las sesiones son datos transitorios (expiran
// solas vía TTL de Redis), PostgreSQL sigue siendo la fuente de verdad
// de la cuenta.
type RedisSessionStore struct {
	rdb *redis.Client
	ttl time.Duration
}

func NewRedisSessionStore(rdb *redis.Client, ttl time.Duration) *RedisSessionStore {
	return &RedisSessionStore{rdb: rdb, ttl: ttl}
}

var _ SessionStore = (*RedisSessionStore)(nil)

func (s *RedisSessionStore) Create(ctx context.Context, userID uuid.UUID) (string, error) {
	token, err := generateToken()
	if err != nil {
		return "", err
	}

	key := sessionKeyPrefix + hashToken(token)
	if err := s.rdb.Set(ctx, key, userID.String(), s.ttl).Err(); err != nil {
		return "", fmt.Errorf("auth: no se pudo crear la sesión: %w", err)
	}

	return token, nil
}

func (s *RedisSessionStore) Get(ctx context.Context, token string) (*Session, error) {
	key := sessionKeyPrefix + hashToken(token)

	val, err := s.rdb.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, ErrTokenInvalid
		}
		return nil, fmt.Errorf("auth: no se pudo leer la sesión: %w", err)
	}

	userID, err := uuid.Parse(val)
	if err != nil {
		return nil, fmt.Errorf("auth: sesión corrupta: %w", err)
	}

	ttl, err := s.rdb.TTL(ctx, key).Result()
	if err != nil {
		return nil, fmt.Errorf("auth: no se pudo leer el TTL de la sesión: %w", err)
	}

	now := time.Now()
	return &Session{
		ID:        token,
		UserID:    userID,
		ExpiresAt: now.Add(ttl),
	}, nil
}

func (s *RedisSessionStore) Delete(ctx context.Context, token string) error {
	key := sessionKeyPrefix + hashToken(token)
	if err := s.rdb.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("auth: no se pudo eliminar la sesión: %w", err)
	}
	return nil
}
