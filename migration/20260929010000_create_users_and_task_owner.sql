-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS users
(
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email ON users (email);

-- Nullable so existing tasks survive; tasks without an owner are not visible through the API.
ALTER TABLE task ADD COLUMN IF NOT EXISTS user_id INTEGER REFERENCES users (id) ON DELETE CASCADE;
CREATE INDEX IF NOT EXISTS idx_task_user_id ON task (user_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_task_user_id;
ALTER TABLE task DROP COLUMN IF EXISTS user_id;
DROP TABLE IF EXISTS users;
-- +goose StatementEnd
