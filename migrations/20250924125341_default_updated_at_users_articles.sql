-- +goose Up
-- +goose StatementBegin
ALTER TABLE users
ALTER COLUMN updated_at SET DEFAULT NOW();
ALTER TABLE articles
ALTER COLUMN updated_at SET DEFAULT NOW();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE users
ALTER COLUMN updated_at DROP DEFAULT;
ALTER TABLE articles
ALTER COLUMN updated_at DROP DEFAULT;
-- +goose StatementEnd
