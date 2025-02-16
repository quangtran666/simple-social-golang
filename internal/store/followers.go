package store

import (
	"context"
	"database/sql"
	"errors"
	"github.com/lib/pq"
)

type Follower struct {
	UserId     int64  `json:"user_id"`
	FollowerId int64  `json:"follower_id"`
	CreatedAt  string `json:"created_at"`
}

type FollowerStore struct {
	db *sql.DB
}

func (s *FollowerStore) Follow(ctx context.Context, followerID, userId int64) error {
	query := `
			INSERT INTO followers (user_id, follower_id)
			VALUES ($1, $2)
			`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	_, err := s.db.ExecContext(ctx, query, userId, followerID)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			return ErrConflict
		}
	}

	return nil
}

func (s *FollowerStore) UnFollow(ctx context.Context, followerID, userId int64) error {
	query := `
			DELETE FROM followers
			WHERE user_id = $1 AND follower_id = $2
			`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	_, err := s.db.ExecContext(ctx, query, userId, followerID)
	if err != nil {
		return err
	}

	return nil
}
