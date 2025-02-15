package store

import (
	"context"
	"database/sql"
)

type Comment struct {
	ID        int64  `json:"id"`
	Content   string `json:"content"`
	PostID    int64  `json:"post_id"`
	UserID    int64  `json:"user_id"`
	CreatedAt string `json:"created_at"`
	User      User   `json:"user"`
}

type CommentStore struct {
	db *sql.DB
}

func (c *CommentStore) GetByPostID(ctx context.Context, postID int64) ([]Comment, error) {
	query := `
			SELECT c.id, c.content, c.post_id, c.user_id, c.created_at, u.username, u.id
			FROM comments c
			INNER JOIN users u ON c.user_id = u.id
			WHERE c.post_id = $1
			ORDER BY c.created_at DESC
			`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	rows, err := c.db.QueryContext(ctx, query, postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []Comment
	for rows.Next() {
		var comment Comment
		comment.User = User{}
		if err := rows.Scan(
			&comment.ID,
			&comment.Content,
			&comment.PostID,
			&comment.UserID,
			&comment.CreatedAt,
			&comment.User.Username,
			&comment.User.ID,
		); err != nil {
			return nil, err
		}
		comments = append(comments, comment)
	}

	return comments, nil
}

func (c *CommentStore) Create(ctx context.Context, comment *Comment) error {
	query := `
			INSERT INTO comments (content, post_id, user_id)
			VALUES ($1, $2, $3) 
			RETURNING id, created_at
			`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	if err := c.db.QueryRowContext(
		ctx,
		query,
		comment.Content,
		comment.PostID,
		comment.UserID,
	).Scan(
		&comment.ID,
		&comment.CreatedAt); err != nil {
		return err
	}

	return nil
}
