package store

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

var (
	ErrNotFound          = errors.New("resources not found")
	ErrConflict          = errors.New("resources already exists")
	QueryTimeoutDuration = 5 * time.Second
)

type Storage struct {
	Posts interface {
		GetByID(ctx context.Context, id int64) (*Post, error)
		Create(ctx context.Context, post *Post) error
		Update(ctx context.Context, post *Post) error
		Delete(ctx context.Context, id int64) error
		GetUserFeed(ctx context.Context, userID int64, fq PaginatedFeedQuery) ([]*PostWithMetadata, error)
	}
	Users interface {
		GetByID(ctx context.Context, id int64) (*User, error)
		GetByEmail(ctx context.Context, email string) (*User, error)
		Create(ctx context.Context, tx *sql.Tx, user *User) error
		CreateAndInvite(ctx context.Context, user *User, token string, exp time.Duration) error
		Activate(ctx context.Context, token string) error
		Delete(ctx context.Context, id int64) error
	}
	Comments interface {
		Create(ctx context.Context, comment *Comment) error
		GetByPostID(ctx context.Context, postID int64) ([]Comment, error)
	}
	Followers interface {
		Follow(ctx context.Context, followerID, userID int64) error
		UnFollow(ctx context.Context, followerID, userID int64) error
	}
	Roles interface {
		GetByName(ctx context.Context, roleName string) (*Role, error)
	}
}

func NewStorage(db *sql.DB) Storage {
	return Storage{
		Posts:     &PostStore{db: db},
		Users:     &UserStore{db: db},
		Comments:  &CommentStore{db: db},
		Followers: &FollowerStore{db: db},
		Roles:     &RoleStore{db: db},
	}
}

func withTx(db *sql.DB, ctx context.Context, fn func(*sql.Tx) error) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	if err := fn(tx); err != nil {
		_ = tx.Rollback()
		return err
	}

	return tx.Commit()
}
