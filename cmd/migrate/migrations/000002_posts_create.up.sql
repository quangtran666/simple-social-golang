CREATE TABLE IF NOT EXISTS posts (
    id bigserial PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    user_id bigint NOT NULL,
    content TEXT NOT NULL,
    created_at timestamp(0) with time zone NOT NULL DEFAULT now(),
    FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
);