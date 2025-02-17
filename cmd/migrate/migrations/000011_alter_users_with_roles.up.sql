ALTER TABLE IF EXISTS users
ADD COLUMN role_id int REFERENCES roles(id) ON DELETE CASCADE DEFAULT 1;

UPDATE users
SET role_id = (
    SELECT id
    FROM roles
    WHERE name = 'user'
);

ALTER TABLE users
ALTER COLUMN role_id DROP DEFAULT;

ALTER TABLE users
ALTER COLUMN role_id SET NOT NULL;