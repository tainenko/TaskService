-- +goose Up
-- +goose StatementBegin
ALTER TABLE task
    ADD COLUMN IF NOT EXISTS description TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS due_date TIMESTAMP WITH TIME ZONE,
    ADD COLUMN IF NOT EXISTS priority INTEGER NOT NULL DEFAULT 0;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE task
    DROP COLUMN IF EXISTS description,
    DROP COLUMN IF EXISTS due_date,
    DROP COLUMN IF EXISTS priority;
-- +goose StatementEnd
