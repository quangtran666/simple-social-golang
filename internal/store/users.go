package store

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"golang.org/x/crypto/bcrypt"
	"time"
)

var (
	ErrDuplicateEmail    = errors.New("email already exists")
	ErrDuplicateUsername = errors.New("username already exists")
)

type User struct {
	ID        int64    `json:"id"`
	Username  string   `json:"username"`
	Email     string   `json:"email"`
	Password  password `json:"-"`
	CreatedAt string   `json:"created_at"`
	IsActive  bool     `json:"is_active"`
	RoleID    int64    `json:"role_id"`
	Role      Role     `json:"role"`
}

type password struct {
	text *string
	hash []byte
}

func (p *password) Set(text string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(text), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	p.text = &text
	p.hash = hash

	return nil
}

type UserStore struct {
	db *sql.DB
}

func (s *UserStore) Create(ctx context.Context, tx *sql.Tx, user *User) error {
	query := `
			INSERT INTO users (username, email, password, role_id)
			VALUES ($1, $2, $3, (SELECT id FROM roles WHERE name = $4)) RETURNING id, created_at
			`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	if err := tx.QueryRowContext(
		ctx,
		query,
		user.Username,
		user.Email,
		user.Password.hash,
		user.Role.Name,
	).Scan(
		&user.ID,
		&user.CreatedAt); err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "users_email_key"`:
			return ErrDuplicateEmail
		case err.Error() == `pq: duplicate key value violates unique constraint "users_username_key"`:
			return ErrDuplicateUsername
		default:
			return err
		}
	}

	return nil
}

func (u *UserStore) GetByID(ctx context.Context, id int64) (*User, error) {
	query := `
			SELECT users.id, username, email, created_at, r.*
			FROM users
			INNER JOIN roles r ON users.role_id = r.id
			WHERE users.id = $1 AND is_active = true
			`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	user := &User{}
	if err := u.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.CreatedAt,
		&user.Role.ID,
		&user.Role.Name,
		&user.Role.Level,
		&user.Role.Description,
	); err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrNotFound
		default:
			return nil, err
		}
	}

	return user, nil
}

func (u *UserStore) CreateAndInvite(ctx context.Context, user *User, token string, invitationExp time.Duration) error {
	return withTx(u.db, ctx, func(tx *sql.Tx) error {
		if err := u.Create(ctx, tx, user); err != nil {
			return err
		}

		if err := u.createUserInviation(ctx, tx, token, invitationExp, user.ID); err != nil {
			return err
		}

		return nil
	})
}

func (u *UserStore) createUserInviation(ctx context.Context, tx *sql.Tx, token string, exp time.Duration, userId int64) error {
	query := `
			INSERT INTO user_invitations (token, expiry, user_id) VALUES ($1, $2, $3)
			`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	_, err := tx.ExecContext(ctx, query, token, time.Now().Add(exp), userId)
	if err != nil {
		return err
	}

	return nil
}

func (u *UserStore) Activate(ctx context.Context, token string) error {
	return withTx(u.db, ctx, func(tx *sql.Tx) error {
		user, err := u.getUserFromInvitation(ctx, tx, token)
		if err != nil {
			return err
		}

		user.IsActive = true
		if err := u.updateUser(ctx, tx, user); err != nil {
			return err
		}

		if err := u.deleteUserInvitation(ctx, tx, user.ID); err != nil {
			return err
		}

		return nil
	})
}

func (s *UserStore) getUserFromInvitation(ctx context.Context, tx *sql.Tx, token string) (*User, error) {
	query := `
			SELECT u.id, u.username, u.email, u.created_at, u.is_active
			FROM users u
			INNER JOIN user_invitations ui ON u.id = ui.user_id
			WHERE ui.token = $1 AND ui.expiry > $2
			`

	hash := sha256.Sum256([]byte(token))
	hashToken := hex.EncodeToString(hash[:])

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	user := &User{}
	if err := tx.QueryRowContext(ctx, query, hashToken, time.Now()).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.CreatedAt,
		&user.IsActive,
	); err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrNotFound
		default:
			return nil, err
		}
	}

	return user, nil
}

func (s *UserStore) updateUser(ctx context.Context, tx *sql.Tx, user *User) error {
	query := `
			UPDATE users
			SET username = $1, email = $2, is_active = $3
			WHERE id = $4
			`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	_, err := tx.ExecContext(ctx, query, user.Username, user.Email, user.IsActive, user.ID)
	if err != nil {
		return err
	}

	return nil
}

func (s *UserStore) deleteUserInvitation(ctx context.Context, tx *sql.Tx, userId int64) error {
	query := `
			DELETE FROM user_invitations
			WHERE user_id = $1
			`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	_, err := tx.ExecContext(ctx, query, userId)
	if err != nil {
		return err
	}

	return nil
}

func (u *UserStore) Delete(ctx context.Context, id int64) error {
	return withTx(u.db, ctx, func(tx *sql.Tx) error {
		if err := u.delete(ctx, tx, id); err != nil {
			return err
		}

		if err := u.deleteUserInvitation(ctx, tx, id); err != nil {
			return err
		}

		return nil
	})
}

func (s *UserStore) delete(ctx context.Context, tx *sql.Tx, id int64) error {
	query := `
			DELETE FROM users
			WHERE id = $1
			`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	_, err := tx.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	return nil
}

func (u *UserStore) GetByEmail(ctx context.Context, email string) (*User, error) {
	query := `
			SELECT id, username, email, created_at
			FROM users
			WHERE email = $1 AND is_active = true
			`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	user := &User{}
	if err := u.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.CreatedAt,
	); err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrNotFound
		default:
			return nil, err
		}
	}

	return user, nil
}
