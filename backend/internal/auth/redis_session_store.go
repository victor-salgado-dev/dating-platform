package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const (
	sessionKeyPrefix = "session:"
	blockedKeyPrefix = "blocked_user:"
	activeKeyPrefix  = "active:"

	// touchInterval es cada cuánto, como mucho, se escribe last_active_at
	// en Postgres para un mismo usuario.
	touchInterval = 2 * time.Minute
)

// resolveScript hace en UN viaje a Redis lo que antes eran GET + TTL más una
// consulta a Postgres por petición:
//
//  1. lee la sesión (token hasheado -> user id);
//  2. comprueba si la cuenta está bloqueada (suspendida o eliminada);
//  3. si toca, marca "activo" con SET NX EX y avisa para refrescar
//     last_active_at (solo una vez por touchInterval).
//
// Devuelve {codigo, user_id[, motivo]} con codigo:
// 0 sesión inexistente, 1 ok, 2 cuenta bloqueada, 3 ok y toca refrescar.
var resolveScript = redis.NewScript(`
local uid = redis.call('GET', KEYS[1])
if not uid then return {0, ''} end
local blocked = redis.call('GET', ARGV[1] .. uid)
if blocked then return {2, uid, blocked} end
local touched = redis.call('SET', ARGV[2] .. uid, '1', 'NX', 'EX', ARGV[3])
if touched then return {3, uid} end
return {1, uid}
`)

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

var (
	_ SessionStore    = (*RedisSessionStore)(nil)
	_ SessionResolver = (*RedisSessionStore)(nil)
	_ AccountBlocker  = (*RedisSessionStore)(nil)
)

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

// Get lee la sesión y su TTL en un solo viaje (pipeline) en lugar de dos.
func (s *RedisSessionStore) Get(ctx context.Context, token string) (*Session, error) {
	key := sessionKeyPrefix + hashToken(token)

	pipe := s.rdb.Pipeline()
	getCmd := pipe.Get(ctx, key)
	ttlCmd := pipe.TTL(ctx, key)
	if _, err := pipe.Exec(ctx); err != nil && !errors.Is(err, redis.Nil) {
		return nil, fmt.Errorf("auth: no se pudo leer la sesión: %w", err)
	}

	val, err := getCmd.Result()
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

	ttl, err := ttlCmd.Result()
	if err != nil {
		return nil, fmt.Errorf("auth: no se pudo leer el TTL de la sesión: %w", err)
	}

	return &Session{
		ID:        token,
		UserID:    userID,
		ExpiresAt: time.Now().Add(ttl),
	}, nil
}

func (s *RedisSessionStore) Delete(ctx context.Context, token string) error {
	key := sessionKeyPrefix + hashToken(token)
	if err := s.rdb.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("auth: no se pudo eliminar la sesión: %w", err)
	}
	return nil
}

// Resolve implementa SessionResolver (ver resolveScript).
func (s *RedisSessionStore) Resolve(ctx context.Context, token string) (uuid.UUID, bool, error) {
	key := sessionKeyPrefix + hashToken(token)

	res, err := resolveScript.Run(ctx, s.rdb, []string{key},
		blockedKeyPrefix, activeKeyPrefix, int(touchInterval.Seconds())).Slice()
	if err != nil {
		return uuid.Nil, false, fmt.Errorf("auth: no se pudo resolver la sesión: %w", err)
	}
	if len(res) < 2 {
		return uuid.Nil, false, fmt.Errorf("auth: respuesta inesperada al resolver la sesión")
	}

	code, _ := res[0].(int64)
	uidStr, _ := res[1].(string)

	switch code {
	case 0:
		return uuid.Nil, false, ErrTokenInvalid
	case 2:
		// Una cuenta eliminada se comporta como antes: la sesión ya no vale.
		if len(res) >= 3 {
			if reason, _ := res[2].(string); reason == BlockDeleted {
				return uuid.Nil, false, ErrTokenInvalid
			}
		}
		return uuid.Nil, false, ErrAccountSuspended
	}

	userID, err := uuid.Parse(uidStr)
	if err != nil {
		return uuid.Nil, false, fmt.Errorf("auth: sesión corrupta: %w", err)
	}
	return userID, code == 3, nil
}

// BlockUser marca la cuenta como bloqueada durante tanto tiempo como duran
// las sesiones: cualquier sesión anterior queda inutilizada de inmediato y,
// pasado ese tiempo, todas habrían caducado igualmente.
func (s *RedisSessionStore) BlockUser(ctx context.Context, userID uuid.UUID, reason string) error {
	if err := s.rdb.Set(ctx, blockedKeyPrefix+userID.String(), reason, s.ttl).Err(); err != nil {
		return fmt.Errorf("auth: no se pudo bloquear la cuenta: %w", err)
	}
	return nil
}

func (s *RedisSessionStore) UnblockUser(ctx context.Context, userID uuid.UUID) error {
	if err := s.rdb.Del(ctx, blockedKeyPrefix+userID.String()).Err(); err != nil {
		return fmt.Errorf("auth: no se pudo desbloquear la cuenta: %w", err)
	}
	return nil
}
