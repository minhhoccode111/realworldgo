-- +goose Up
-- +goose StatementBegin
DO $$ BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'user_role') THEN
    CREATE TYPE user_role AS ENUM ('admin', 'user');
  END IF;
END $$;

CREATE TABLE IF NOT EXISTS users (
  id         UUID         PRIMARY KEY   DEFAULT gen_random_uuid(),
  role       user_role    NOT     NULL,
  email      VARCHAR(255) NOT     NULL  UNIQUE,
  password   VARCHAR(255) NOT     NULL,
  is_active  BOOLEAN      DEFAULT FALSE,
  last_name  VARCHAR(100) NOT     NULL,
  first_name VARCHAR(100) NOT     NULL,
  created_at TIMESTAMPTZ  DEFAULT now(),
  updated_at TIMESTAMPTZ  DEFAULT now()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email ON users (email);
CREATE INDEX IF NOT EXISTS idx_users_fullname ON users (first_name, last_name);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS users;
DROP TYPE IF EXISTS user_role;
-- +goose StatementEnd
