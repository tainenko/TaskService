-- +goose Up
-- +goose StatementBegin
CREATE INDEX IF NOT EXISTS idx_task_name_status ON task (name, status);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_task_name_status;
-- +goose StatementEnd
