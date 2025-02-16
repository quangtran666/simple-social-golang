CREATE TABLE IF NOT EXISTS user_invitations (
    token bytea PRIMARY KEY,
    user_id bigint not null,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

ALTER TABLE users
ADD COLUMN is_active boolean NOT NULL DEFAULT false;