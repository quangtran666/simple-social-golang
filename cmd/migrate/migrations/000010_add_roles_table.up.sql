CREATE TABLE IF NOT EXISTS roles (
    id bigserial PRIMARY KEY,
    name varchar(255) NOT NULL,
    level int NOT NULL DEFAULT 0,
    description text
);

INSERT INTO roles (name, level, description)
VALUES ('user', 1, 'A User can create posts and comments');

INSERT INTO roles (name, level, description)
VALUES ('moderator', 2, 'A moderator can update other users posts');

INSERT INTO roles (name, level, description)
VALUES ('admin', 3, 'A Admin can update and delete other users posts');