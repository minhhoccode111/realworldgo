-- +goose Up
-- +goose StatementBegin
ALTER TABLE comments
ADD COLUMN updated_at TIMESTAMPTZ DEFAULT NOW();
-- +goose StatementEnd

-- update existing comments
UPDATE comments
SET updated_at = created_at;

-- +goose Down
-- +goose StatementBegin
ALTER TABLE comments
DROP COLUMN updated_at;
-- +goose StatementEnd
