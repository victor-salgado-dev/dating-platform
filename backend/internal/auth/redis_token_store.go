package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// TokenPurpose distingue el espacio de nombres de cada tipo de token de
// un solo uso, para que un token de verificación de email nunca pueda
// colarse como token de reset de contraseña, aunque coincidiera el hash
// (no debería, pero la separación de namespace lo hace imposible).
type TokenPurpose string

const (
	PurposeEmailVerification TokenPurpose = "email_verify"
	PurposePasswordReset     TokenPurpose = "password_reset"
)

// TokenStore persiste tokens opacos de un solo uso, asociados a un
// usuario, con expiración. Vive en Redis por la misma razón que las
// sesiones: son datos transitorios, no fuente de verdad.
type TokenStore interface {
	// Create genera un token nuevo para userID bajo el propósito indicado.
	// Devuelve el token en claro (el que se envía por email).
	Create(ctx context.Context, purpose TokenPurpose, userID uuid.UUID, ttl time.Duration) (token string, err error)

	// Consume valida el token y lo invalida atómicamente (un solo uso).
	// Devuelve ErrTokenInvalid si no existe, ya se usó, o ha caducado.
	Consume(ctx context.Context, purpose TokenPurpose, token string) (uuid.UUID, error)
}

type RedisTokenStore struct {
	rdb *redis.Client
}

func NewRedisTokenStore(rdb *redis.Client) *RedisTokenStore {
	return &RedisTokenStore{rdb: rdb}
}

var _ TokenStore = (*RedisTokenStore)(nil)

func (s *RedisTokenStore) Create(ctx context.Context, purpose TokenPurpose, userID uuid.UUID, ttl time.Duration) (string, error) {
	token, err := generateToken()
	if err != nil {
		return "", err
	}

	key := tokenKey(purpose, token)
	if err := s.rdb.Set(ctx, key, userID.String(), ttl).Err(); err != nil {
		return "", fmt.Errorf("auth: no se pudo crear el token: %w", err)
	}

	return token, nil
}

func (s *RedisTokenStore) Consume(ctx context.Context, purpose TokenPurpose, token string) (uuid.UUID, error) {
	key := tokenKey(purpose, token)

	// GetDel es atómico: evita condiciones de carrera en las que el mismo
	// token se pudiera consumir dos veces en paralelo.
	val, err := s.rdb.GetDel(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return uuid.Nil, ErrTokenInvalid
		}
		return uuid.Nil, fmt.Errorf("auth: no se pudo leer el token: %w", err)
	}

	userID, err := uuid.Parse(val)
	if err != nil {
		return uuid.Nil, fmt.Errorf("auth: token corrupto: %w", err)
	}

	return userID, nil
}

func tokenKey(purpose TokenPurpose, token string) string {
	return fmt.Sprintf("token:%s:%s", purpose, hashToken(token))
}
