package store

import (
	"context"
	"database/sql"
)

type Role struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Level       int    `json:"level"`
	Description string `json:"description"`
}

type RoleStore struct {
	db *sql.DB
}

func (r *RoleStore) GetByName(ctx context.Context, roleName string) (*Role, error) {
	query := `
			SELECT id, name, level, description
			FROM roles
			WHERE name = $1	
			`

	role := &Role{}
	err := r.db.QueryRowContext(ctx, query, roleName).Scan(
		&role.ID,
		&role.Name,
		&role.Level,
		&role.Description)
	if err != nil {
		return nil, err
	}

	return role, nil
}
