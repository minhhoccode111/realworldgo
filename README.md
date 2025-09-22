# ![RealWorld Example App](logo.png)

> ### Golang codebase containing real world examples (CRUD, auth, advanced patterns, etc) that adheres to the [RealWorld](https://github.com/gothinkster/realworld) spec and API.

### [Demo](https://demo.realworld.build/)&nbsp;&nbsp;&nbsp;&nbsp;[RealWorld](https://github.com/gothinkster/realworld)

This codebase was created to demonstrate a fully fledged fullstack application built with **Golang** including CRUD operations, authentication, routing, pagination, and more.

We've gone to great lengths to adhere to the **Golang** community styleguides & best practices.

For more information on how to this works with other frontends/backends, head over to the [RealWorld](https://github.com/gothinkster/realworld) repo.

## How it works

> I try to use Clean Architecture as much as I can.

```txt
internal/
├── app/                 # Application layer (dependency injection)
├── handlers/            # HTTP handlers (thin layer)
├── services/            # Business logic layer
├── repositories/        # Data access layer
├── models/              # Domain models
├── middleware/          # HTTP middleware
├── utils/               # Utility functions
└── config/              # Configuration
```

## Getting started

> docker compose up

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

## Database (NoSQL features are not allowed, even though PostgreSQL has them)

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

## Concepts

## Todo

- [ ] Add tests

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
