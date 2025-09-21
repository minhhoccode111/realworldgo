-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS users (
  id         UUID         PRIMARY KEY   DEFAULT gen_random_uuid(),
  email      VARCHAR(255) NOT     NULL  UNIQUE,
  password   VARCHAR(255) NOT     NULL,
  username   VARCHAR(100) NOT     NULL,
  created_at TIMESTAMPTZ  DEFAULT now(),
  updated_at TIMESTAMPTZ  DEFAULT now()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email ON users (email);
CREATE INDEX IF NOT EXISTS idx_users_username ON users (username);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS users;
DROP TYPE IF EXISTS user_role;
-- +goose StatementEnd
