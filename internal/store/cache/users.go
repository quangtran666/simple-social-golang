package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/quangtran666/simple-social-golang/internal/store"
	"github.com/redis/go-redis/v9"
	"time"
)

type UserStore struct {
	rbd *redis.Client
}

const UserExpTime = time.Minute

func (u *UserStore) Get(ctx context.Context, userId int64) (*store.User, error) {
	cacheKey := fmt.Sprintf("user-%v", userId)

	data, err := u.rbd.Get(ctx, cacheKey).Result()
	if errors.Is(err, redis.Nil) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	user := &store.User{}
	if data != "" {
		err := json.Unmarshal([]byte(data), user)
		if err != nil {
			return nil, err
		}
	}

	return user, nil
}

func (u *UserStore) Set(ctx context.Context, user *store.User) error {
	cacheKey := fmt.Sprintf("user-%v", user.ID)

	json, err := json.Marshal(user)
	if err != nil {
		return err
	}

	return u.rbd.Set(ctx, cacheKey, json, UserExpTime).Err()
}
