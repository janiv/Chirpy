-- +goose Up
CREATE TABLE chirps (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    body TEXT UNIQUE NOT NULL,
    user_id uuid NOT NULL,
    CONSTRAINT fk_users FOREIGN KEY (user_id)
    REFERENCES users(id) ON DELETE CASCADE
);

-- +goose Down
DROP TABLE chirps;