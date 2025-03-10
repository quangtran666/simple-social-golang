﻿package store

import (
	"context"
	"database/sql"
	"errors"
	"github.com/lib/pq"
)

type Post struct {
	ID        int64     `json:"id"`
	Content   string    `json:"content"`
	Title     string    `json:"title"`
	UserID    int64     `json:"user_id"`
	Tags      []string  `json:"tags"`
	CreatedAt string    `json:"created_at"`
	UpdatedAt string    `json:"updated_at"`
	Version   int       `json:"version"`
	Comments  []Comment `json:"comments"`
	User      User      `json:"user"`
}

type PostWithMetadata struct {
	Post
	CommentCount int `json:"comments_count"`
}

type PostStore struct {
	db *sql.DB
}

func (s *PostStore) GetUserFeed(ctx context.Context, userID int64, fq PaginatedFeedQuery) ([]*PostWithMetadata, error) {
	query := `
			SELECT p.id, p.content, p.title, p.user_id, p.tags, p.created_at, p.updated_at, p.version,
			       u.username
			       ,COUNT(c.id) as comment_count
			FROM posts p
			LEFT JOIN comments c ON p.id  = c.post_id
			LEFT JOIN users u ON p.user_id = u.id 
			JOIN followers f ON p.user_id = f.follower_id OR p.user_id = $1	
			WHERE 
			    f.follower_id = $1 AND
				(p.title ILIKE '%' || $4 || '%' OR p.content ILIKE '%' || $4 || '%') AND
				(p.tags @> $5 OR $5 = '{}')
			GROUP BY p.id, u.username
			ORDER BY p.created_at ` + fq.Sort + `
			LIMIT $2
			OFFSET $3
			`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	rows, err := s.db.QueryContext(ctx, query, userID, fq.Limit, fq.Offset, fq.Search, pq.Array(fq.Tags))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var feeds []*PostWithMetadata
	for rows.Next() {
		post := &PostWithMetadata{}
		if err := rows.Scan(
			&post.ID,
			&post.Content,
			&post.Title,
			&post.UserID,
			pq.Array(&post.Tags),
			&post.CreatedAt,
			&post.UpdatedAt,
			&post.Version,
			&post.User.Username,
			&post.CommentCount,
		); err != nil {
			return nil, err
		}

		feeds = append(feeds, post)
	}

	return feeds, nil
}

func (s *PostStore) Create(ctx context.Context, post *Post) error {
	query := `
				INSERT INTO posts (content, title, user_id, tags) 
				VALUES ($1, $2, $3, $4) RETURNING id, created_at, updated_at
				`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	if err := s.db.QueryRowContext(
		ctx,
		query,
		post.Content,
		post.Title,
		post.UserID,
		pq.Array(post.Tags),
	).Scan(
		&post.ID,
		&post.CreatedAt,
		&post.UpdatedAt); err != nil {
		return err
	}

	return nil
}

func (s *PostStore) GetByID(ctx context.Context, id int64) (*Post, error) {
	query := `
			SELECT id, content, title, user_id, tags, created_at, updated_at, version
			FROM posts
			WHERE id = $1
			`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	post := &Post{}
	if err := s.db.QueryRowContext(ctx, query, id).Scan(
		&post.ID,
		&post.Content,
		&post.Title,
		&post.UserID,
		pq.Array(&post.Tags),
		&post.CreatedAt,
		&post.UpdatedAt,
		&post.Version,
	); err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrNotFound
		default:
			return nil, err
		}
	}

	return post, nil
}

func (p *PostStore) Update(ctx context.Context, post *Post) error {
	query := `
			UPDATE posts
			SET content = $1, title = $2, updated_at = NOW(), version = version + 1
			WHERE id = $3 AND version = $4
			RETURNING version
			`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	err := p.db.QueryRowContext(
		ctx,
		query,
		post.Content,
		post.Title,
		post.ID,
		post.Version).Scan(&post.Version)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrNotFound
		default:
			return err
		}
	}

	return nil
}

func (p *PostStore) Delete(ctx context.Context, id int64) error {
	query := `
			DELETE FROM posts
			WHERE id = $1
			`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	res, err := p.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return ErrNotFound
	}

	return nil
}
