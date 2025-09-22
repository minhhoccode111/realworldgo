-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS users (
  id           UUID          PRIMARY KEY   DEFAULT GEN_RANDOM_UUID(),
  email        VARCHAR(320)  NOT     NULL  UNIQUE CHECK (LENGTH(email) BETWEEN 5 AND 320),
  username     VARCHAR(50)   NOT     NULL  UNIQUE CHECK (LENGTH(TRIM(username)) BETWEEN 2 AND 50),
  password     VARCHAR(255)  NOT     NULL,
  bio          VARCHAR(255),
  image        VARCHAR(2048),
  created_at   TIMESTAMPTZ   DEFAULT NOW(),
  updated_at   TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS articles (
  id           UUID          PRIMARY KEY   DEFAULT GEN_RANDOM_UUID(),
  author_id    UUID          NOT     NULL,
  slug         VARCHAR(255)  NOT     NULL  UNIQUE CHECK (LENGTH(TRIM(slug)) BETWEEN 2 AND 255),
  title        VARCHAR(255)  NOT     NULL,
  body         TEXT          NOT     NULL,
  description  VARCHAR(255),
  created_at   TIMESTAMPTZ   DEFAULT NOW(),
  updated_at   TIMESTAMPTZ,
  deleted_at   TIMESTAMPTZ,
  CONSTRAINT   fk_articles_author_id  FOREIGN KEY (author_id)  REFERENCES users(id)
);

CREATE TABLE IF NOT EXISTS comments (
  id           UUID          PRIMARY KEY   DEFAULT GEN_RANDOM_UUID(),
  author_id    UUID          NOT     NULL,
  article_id   UUID          NOT     NULL,
  body         TEXT          NOT     NULL,
  created_at   TIMESTAMPTZ   DEFAULT NOW(),
  deleted_at   TIMESTAMPTZ,
  CONSTRAINT   fk_comments_author_id  FOREIGN KEY (author_id)  REFERENCES users(id),
  CONSTRAINT   fk_comments_article_id FOREIGN KEY (article_id) REFERENCES articles(id)
);

CREATE TABLE IF NOT EXISTS tags (
  id           UUID          PRIMARY KEY   DEFAULT GEN_RANDOM_UUID(),
  name         VARCHAR(50)   NOT     NULL  UNIQUE CHECK (LENGTH(TRIM(name)) BETWEEN 1 AND 50)
);

CREATE TABLE IF NOT EXISTS article_tags (
  tag_id       UUID          NOT     NULL  REFERENCES tags(id),
  article_id   UUID          NOT     NULL  REFERENCES articles(id),
  PRIMARY      KEY           (tag_id, article_id)
);

CREATE TABLE IF NOT EXISTS favorites (
  user_id      UUID          NOT     NULL  REFERENCES users(id),
  article_id   UUID          NOT     NULL  REFERENCES articles(id),
  PRIMARY      KEY           (user_id, article_id)
);

CREATE TABLE IF NOT EXISTS follows (
  follower_id  UUID          NOT     NULL  REFERENCES users(id),
  following_id UUID          NOT     NULL  REFERENCES users(id),
  PRIMARY      KEY           (follower_id, following_id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS follows;
DROP TABLE IF EXISTS favorites;
DROP TABLE IF EXISTS article_tags;
DROP TABLE IF EXISTS comments;
DROP TABLE IF EXISTS articles;
DROP TABLE IF EXISTS tags;
DROP TABLE IF EXISTS users;
-- +goose StatementEnd
