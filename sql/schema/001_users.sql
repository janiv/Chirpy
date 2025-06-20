-- +goose Up
CREATE TABLE users (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    email TEXT UNIQUE NOT NULL,
    hashed_password TEXT DEFAULT 'unset' NOT NULL,
    is_chirpy_red BOOLEAN DEFAULT false
);

-- +goose Down
DROP TABLE users;