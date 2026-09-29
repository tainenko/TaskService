-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS tag
(
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    name VARCHAR(50) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,
    UNIQUE (user_id, name)
);

CREATE TABLE IF NOT EXISTS task_tag
(
    task_id INTEGER NOT NULL REFERENCES task (id) ON DELETE CASCADE,
    tag_id INTEGER NOT NULL REFERENCES tag (id) ON DELETE CASCADE,
    PRIMARY KEY (task_id, tag_id)
);
CREATE INDEX IF NOT EXISTS idx_task_tag_tag_id ON task_tag (tag_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS task_tag;
DROP TABLE IF EXISTS tag;
-- +goose StatementEnd
