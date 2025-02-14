package store

import (
	"context"
	"database/sql"
)

type Storage struct {
	Posts interface {
		Create(ctx context.Context, post *Post) error
	}
	Users interface {
		Create(ctx context.Context, user *User) error
	}
}

func NewStorage(db *sql.DB) Storage {
	return Storage{
		Posts: &PostsStore{db: db},
		Users: &UsersStore{db: db},
	}
}
