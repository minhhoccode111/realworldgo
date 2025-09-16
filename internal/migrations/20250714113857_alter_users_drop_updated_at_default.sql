-- +goose Up
-- +goose StatementBegin
ALTER TABLE users ALTER COLUMN updated_at DROP DEFAULT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE users ALTER COLUMN updated_at SET DEFAULT NOW();
-- +goose StatementEnd
