package repository

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type AuthRepository interface {
	CreateAuth(userid uint, accessUuid, refreshUuid string, atExpires, rtExpires int64) error
	FetchAuth(uuid string) (string, error)
	DeleteAuth(uuid string) error
}

type authRepository struct {
	redis *redis.Client
}

func NewAuthRepository(redis *redis.Client) AuthRepository {
	return &authRepository{redis}
}

func (r *authRepository) CreateAuth(userid uint, accessUuid, refreshUuid string, atExpires, rtExpires int64) error {
	at := time.Unix(atExpires, 0)
	rt := time.Unix(rtExpires, 0)
	now := time.Now()

	// userIdStr := string(rune(userid)) // Warning: check casting, usually fmt.Sprint(userid)

	err := r.redis.Set(context.Background(), accessUuid, userid, at.Sub(now)).Err()
	if err != nil {
		return err
	}
	err = r.redis.Set(context.Background(), refreshUuid, userid, rt.Sub(now)).Err()
	if err != nil {
		return err
	}
	return nil
}

func (r *authRepository) FetchAuth(uuid string) (string, error) {
	return r.redis.Get(context.Background(), uuid).Result()
}

func (r *authRepository) DeleteAuth(uuid string) error {
	return r.redis.Del(context.Background(), uuid).Err()
}
