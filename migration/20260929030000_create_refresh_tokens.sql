-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS refresh_token
(
    id BIGSERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    -- All tokens descended from one login share a family so reuse can revoke the chain.
    family_id VARCHAR(32) NOT NULL,
    -- SHA-256 of the token; the token itself is never stored.
    token_hash CHAR(64) NOT NULL UNIQUE,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    revoked_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_refresh_token_family_id ON refresh_token (family_id);
CREATE INDEX IF NOT EXISTS idx_refresh_token_user_id ON refresh_token (user_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS refresh_token;
-- +goose StatementEnd
