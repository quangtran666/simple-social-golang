package cache

import (
	"context"
	"github.com/quangtran666/simple-social-golang/internal/store"
	"github.com/redis/go-redis/v9"
)

type Storage struct {
	Users interface {
		Get(ctx context.Context, id int64) (*store.User, error)
		Set(ctx context.Context, user *store.User) error
	}
}

func NewRedisStorage(rbd *redis.Client) Storage {
	return Storage{
		Users: &UserStore{rbd: rbd},
	}
}
