-- +goose Up
-- +goose StatementBegin
CREATE INDEX IF NOT EXISTS idx_task_status ON task (status);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_task_status;
-- +goose StatementEnd
