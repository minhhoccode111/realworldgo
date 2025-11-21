# ![RealWorld Example App](logo.png)

> ### Golang codebase containing real world examples (CRUD, auth, advanced patterns, etc) that adheres to the [RealWorld](https://github.com/gothinkster/realworld) spec and API.

For more information on how to this works with other frontends/backends, head
over to the [RealWorld](https://github.com/gothinkster/realworld) repo.

## [Endpoints](https://docs.realworld.show/specifications/backend/endpoints/)

```http
       # Auth
POST   /users
POST   /users/login
GET    /user
PUT    /user

       # Article, Favorite, Comments
POST   /articles
GET    /articles
                ?tag={tag1,tag2}
                &author={username}
                &favorited={username}
                &limit={limit}
                &offset={offset}
GET    /articles/feed
GET    /articles/{slug}
PUT    /articles/{slug}
DELETE /articles/{slug}
POST   /articles/{slug}/favorite
DELETE /articles/{slug}/favorite
POST   /articles/{slug}/comments
GET    /articles/{slug}/comments
DELETE /articles/{slug}/comments/{commentId}

       # Profiles
GET    /profiles/{username}
POST   /profiles/{username}/follow
DELETE /profiles/{username}/follow

       # Tags
GET    /tags
```

## [API Response Format](https://docs.realworld.show/specifications/backend/api-response-format/)

## Database Design

```txt
- Users
       - id
idx    - email     - unique
idx    - username  - unique
       - password
       - image
       - bio
       - created_at
       - updated_at
- Articles
       - id
       - author_id
idx    - slug      - unique
       - body
       - title
       - description
       - created_at
       - updated_at
       - deleted_at
- Comments
       - id
       - article_id
       - author_id
       - body
       - created_at
       - deleted_at
- ArticleTags
       - article_id
       - tag_id
- Tags
       - id
idx    - name      - unique
- Favorites
       - user_id
       - article_id
- Follows
       - follower_id
       - following_id
```

- User vs. Article: One-to-Many
- Article vs. Comment: One-to-Many
- Article vs. Tag: Many-to-Many
- Follow User vs. User: Many-to-Many
- Favorite User vs. Article: Many-to-Many

## Concepts Learned

- A project file structure that I like :)
- Concurrency with `errgroup`
- Transaction, Rollback (with `defer`), Commit
- Batch Insert to improve performance
- `values := []any{}` must be `[]any` to be used as `values...` in
  `db.QueryContext`
- `QueryRowContext` returns a single row
- `QueryContext` returns multiple rows, which we have to `.Close()` manually
- `LEFT JOIN`
- `RETURNING id`
- `N+1 Problem` like get all author profiles after we get all the articles
- `WHERE id::TEXT = '...'` to prevent exception when `id` is not a uuid
- `SELECT EXISTS (SELECT 1 FROM ... WHERE ...);` to check for existence
- `CROSS JOIN` to combine all rows from one table to all rows in another table
- `STRING_AGG(t.name, ', ')` to aggregate tags into one column as a string
- `COALESCE(ARRAY_AGG(DISTINCT t.name) FILTER (WHERE t.name IS NOT NULL), '{}')`
  to aggregate tags into one column as an array
- `var tags pq.StringArray` represents a one-dimensional array of the PostgreSQL
  character types, then `[]string(tags)` to get array of strings
- `COUNT(DISTINCT id)` to count the number of distinct rows
- `COUNT(*) OVER()` to count all rows that match the `WHERE` before applying
  `LIMIT` and `OFFSET`
- `WHERE ('' = $1 OR username = $1)` skip if empty
- `ON CONFLICT (name) DO UPDATE SET name=EXCLUDED.name` if insert (or update)
  conflict update the old value to new value (same)
- `ON CONFLICT DO NOTHING`

## Todo

- [ ] Tag can be a slice of strings
- [ ] Allow update article's tags
- [x] Add tests with standard `testing` package
- [ ] Add cache with Redis
- [ ] Add notifications with SSE + RabbitMQ + Redis Pub/Sub Architecture

## MakeFile

Run build make command with tests

```bash
make all
```

Build the application

```bash
make build
```

Run the application

```bash
make run
```

Create DB container

```bash
make docker-run
```

Shutdown DB Container

```bash
make docker-down
```

DB Integrations Test

```bash
make itest
```

Live reload the application

```bash
make watch
```

Run the test suite

```bash
make test
```

Clean up binary from the last build

```bash
make clean
```

## Test Script

Run the [test script](./api-test/run-api-tests.sh)

```bash
./api-test/run-api-tests.sh
```

## Testing Strategy

- Unit tests rely solely on the Go standard `testing` package and live under the `internal` tree (config, utils, middleware). They cover configuration loading, HTTP helpers, validation logic, JWT generation, and security middleware flows.
- Run the full suite locally with `go test ./...` or `make test`. The commands set no external dependencies, so they execute quickly without Postgres.
- When adding new packages, co-locate `_test.go` files beside the implementation and follow the same table-driven style used in `internal/utils/validate_input_test.go` for clarity.
- Tests that depend on environment variables should use `t.Setenv` to stay hermetic. Middleware tests prefer small stubs (see `internal/middleware/auth_test.go`) instead of third-party mocks.

Will produce an [output](./api-test/out) like this:

```txt
┌─────────────────────────┬──────────────────┬─────────────────┐
│                         │         executed │          failed │
├─────────────────────────┼──────────────────┼─────────────────┤
│              iterations │                1 │               0 │
├─────────────────────────┼──────────────────┼─────────────────┤
│                requests │               32 │               0 │
├─────────────────────────┼──────────────────┼─────────────────┤
│            test-scripts │               48 │               0 │
├─────────────────────────┼──────────────────┼─────────────────┤
│      prerequest-scripts │               18 │               0 │
├─────────────────────────┼──────────────────┼─────────────────┤
│              assertions │              335 │               0 │
├─────────────────────────┴──────────────────┴─────────────────┤
│ total run duration: 16.7s                                    │
├──────────────────────────────────────────────────────────────┤
│ total data received: 32.01kB (approx)                        │
├──────────────────────────────────────────────────────────────┤
│ average response time: 9ms [min: 1ms, max: 61ms, s.d.: 17ms] │
└──────────────────────────────────────────────────────────────┘
```

## Contributing

Contributions are **welcome and highly appreciated**!\
This project follows the [RealWorld API
spec](https://github.com/gothinkster/realworld) — please make sure your changes
remain compliant.
